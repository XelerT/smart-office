package auth

import (
	"net/http"

	"smart-office/internal/config"
)

// --- Placeholder Payload Structs ---

// Represents normalized user info obtained from various sources.
type UserData struct {
	Name      string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// --- Service Definition ---

type authService struct {
	cfg        *config.Config
	httpClient *http.Client
	// userRepo Repository
}

// LoginRequest defines the structure for the login request body.
type LoginRequest struct {
	Type     string  `json:"type"`
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
}

// Defines how to find the user during login
type LoginCriteria struct {
	Type LoginType
	Name string
}

type LoginType int

const (
	LoginRegular     LoginType = iota
	LoginOAuthGoogle           // For future
	LoginOAuthYandex
	LoginOAuthVK
)

// Defines the expected JSON body
type PasswordResetRequestPayload struct {
	Email string `json:"email"`
}
