// In internal/handlers/auth.go

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"smart-office/internal/domain/auth"
	"smart-office/internal/platform/logger"
	mw "smart-office/internal/server"

	"golang.org/x/crypto/bcrypt"
)

// Holds dependencies needed by the auth HTTP handlers.
type AuthHandler struct {
	AuthService   auth.Service
	Authenticator *mw.Authenticator
	UserRepo      auth.Repository
}

func NewAuthHandler(authService auth.Service, authenticator *mw.Authenticator, userRepo auth.Repository) *AuthHandler {
	if authService == nil {
		panic("AuthHandler: AuthService cannot be nil")
	}
	if authenticator == nil {
		panic("AuthHandler: Authenticator cannot be nil")
	}
	return &AuthHandler{
		AuthService:   authService,
		Authenticator: authenticator,
		UserRepo:      userRepo,
	}
}

// RegisterRoutes connects the auth API endpoints to their handler methods.
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/passwordReset", h.PasswordReset)

	mux.Handle("POST /api/auth/load", h.Authenticator.Middleware(http.HandlerFunc(h.Load)))
}

// --- Handler Method Stubs ---

// Handles new user registration attempts
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warnf("Register: Failed to decode request body: %v", err)
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var userData *auth.UserData
	var err error
	providerKey := strings.ToLower(req.Type)

	switch providerKey {
	case "regular":
		if req.Name == nil || req.Email == nil || req.Password == nil ||
			*req.Name == "" || *req.Email == "" || *req.Password == "" {
			writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing name, email, or password"})
			return
		}
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
			return
		}
		userData = &auth.UserData{
			Name:     *req.Name,
			Email:    *req.Email,
			Password: *req.Password,
		}
		logger.Log.Infof("Register: Attempting regular registration for user '%s', email '%s'", *req.Name, *req.Email)

	default:
		logger.Log.Warnf("Register: Unsupported registration type '%s'", req.Type)
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Unsupported registration type: %s", req.Type)})
		return
	}

	createdUser, err := h.AuthService.RegisterUser(ctx, userData)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			logger.Log.Warnf("Register: Conflict - user already exists for name %s", userData.Name)
			writeJSONResponse(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("User %s already exists", userData.Name)}) // 409
		} else {
			logger.Log.Errorf("Register: Failed to register user %s: %v", userData.Name, err)
			writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create user account"}) // 500
		}
		return
	}

	token, err := h.AuthService.GenerateToken(ctx, createdUser)
	if err != nil {
		logger.Log.Errorf("Register: User %s created, but failed to generate token: %v", createdUser.Name, err)
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "User created, but failed to generate session token"})
		return
	}

	response := RegisterResponse{
		ID:       createdUser.ID.Hex(),
		Name:     createdUser.Name,
		Token:    token,
		Balance:  "0.00",
		Reserved: "0.00",
	}

	logger.Log.Infof("Register: Successfully registered user %s (%s) (Type: %s)", createdUser.Name, createdUser.ID.Hex(), providerKey)
	writeJSONResponse(w, http.StatusCreated, response)
}

// Handles user login via regular credentials.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warnf("Login: Failed to decode request body: %v", err)
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var user *auth.User                 // the user found in the DB
	var findCriteria auth.LoginCriteria // Criteria for DB lookup
	var err error

	switch strings.ToLower(req.Type) { // Use ToLower for case-insensitivity
	case "regular":
		if req.Name == nil || req.Password == nil || *req.Name == "" || *req.Password == "" {
			writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing name or password"})
			return
		}
		findCriteria = auth.LoginCriteria{Type: auth.LoginRegular, Name: *req.Name}
		logger.Log.Infof("Login: Attempting regular login for user '%s'", *req.Name)

	default:
		logger.Log.Warnf("Login: Unsupported login type '%s'", req.Type)
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Unsupported login type: %s", req.Type)})
		return
	}

	user, err = h.UserRepo.FindUserForLogin(ctx, findCriteria)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Log.Warnf("Login: User not found for criteria %+v", findCriteria)
			if strings.ToLower(req.Type) == "regular" {
				writeJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"}) // 401
			} else {
				writeJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "No user linked to this " + req.Type + " account"}) // 401
			}
		} else {
			logger.Log.Errorf("Login: Database error finding user for criteria %+v: %v", findCriteria, err)
			writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"}) // 500
		}
		return
	}

	if strings.ToLower(req.Type) == "regular" {
		passwordFromRequest := *req.Password

		passwordMatch := false

		switch user.Hash {
		case auth.HashBcrypt:
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordFromRequest))
			if err == nil {
				passwordMatch = true
			} else if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				logger.Log.Errorf("Login: Error comparing bcrypt hash for user %s (%s): %v", user.Name, user.ID.Hex(), err)
				writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error during login process"})
				return
			}

		default:
			logger.Log.Errorf("Login: Unknown password hash type '%s' for user %s (%s)", user.Hash, user.Name, user.ID.Hex())
			writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "User account configuration error"})
			return
		}

		if !passwordMatch {
			logger.Log.Warnf("Login: Invalid password attempt for user %s", user.Name)
			writeJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"}) // 401
			return
		}
	}

	token, err := h.AuthService.GenerateToken(ctx, user) // JWT generation
	if err != nil {
		logger.Log.Errorf("Login: Failed to generate token for user %s (%s): %v", user.Name, user.ID.Hex(), err)
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error generating session"})
		return
	}

	userResponseDTO, err := h.buildUserResponseDTO(ctx, user)
	if err != nil {
		logger.Log.Errorf("Login: Failed to build response DTO for user %s (%s): %v", user.Name, user.ID.Hex(), err)
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error preparing response"})
		return
	}

	loginResponse := LoginResponse{
		UserResponse: *userResponseDTO, // Embed DTO fields
		Token:        token,
	}

	logger.Log.Infof("Login: Successful login for user %s (%s) (Type: %s)", user.Name, user.ID.Hex(), req.Type)
	writeJSONResponse(w, http.StatusOK, loginResponse)
}

// Builds the common UserResponse DTO
func (h *AuthHandler) buildUserResponseDTO(ctx context.Context, user *auth.User) (*UserResponse, error) {

	response := &UserResponse{
		ID:                user.ID.Hex(),
		Name:              user.Name,
		FavouritePrinters: user.FavouritePrinters,
		Balance:           user.Balance.String(),
		Reserved:          user.Reserved.String(),
		DefaultPrinter:    user.DefaultPrinter,
		Files:             []FileData{},
		Payments:          []UserPaymentData{},
	}

	if user.Payments != nil {
		response.Payments = make([]UserPaymentData, 0, len(user.Payments))
		for _, p := range user.Payments {
			paymentDate := time.UnixMilli(int64(p.Date)).UTC()
			response.Payments = append(response.Payments,
				UserPaymentData{
					ID:            p.ID,
					Sum:           p.Sum,
					PsID:          p.PsID,
					FopReceiptKey: p.FopReceiptKey,
					CardNumber:    p.CardNumber,
					Date:          paymentDate,
				})
		}
	}

	// Map Files
	// TODO: Implement file mapping logic here

	return response, nil
}

// Retrieves the authenticated user context and formats the response
func (h *AuthHandler) Load(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentUser, ok := mw.UserFromContext(ctx)
	if !ok {
		logger.Log.Error("Load: User not found in context (middleware error?)")
		http.Error(w, "Internal Server Error: User context missing", http.StatusInternalServerError)
		return
	}
	logger.Log.Infof("Load: Handling request for user %s (%s)", currentUser.Name, currentUser.ID.Hex())

	response := UserResponse{
		ID:                currentUser.ID.Hex(),
		Name:              currentUser.Name,
		FavouritePrinters: currentUser.FavouritePrinters,
		Balance:           currentUser.Balance.String(),
		Reserved:          currentUser.Reserved.String(),
		DefaultPrinter:    currentUser.DefaultPrinter,

		Files:    []FileData{},
		Payments: []UserPaymentData{},
	}

	if currentUser.Payments != nil {
		response.Payments = make([]UserPaymentData, 0, len(currentUser.Payments))
		for _, p := range currentUser.Payments {
			// Convert BSON DateTime (ms since epoch)
			paymentDate := time.UnixMilli(int64(p.Date)).UTC()
			response.Payments = append(response.Payments, UserPaymentData{
				ID:            p.ID,
				Sum:           p.Sum,
				PsID:          p.PsID,
				FopReceiptKey: p.FopReceiptKey,
				CardNumber:    p.CardNumber,
				Date:          paymentDate,
			})
		}
	}

	// here we need to map Files

	logger.Log.Infof("Load: Successfully prepared response for user %s", currentUser.Name)
	writeJSONResponse(w, http.StatusOK, response)
}
