package handlers // Or wherever you place handlers

import (
	"time"
	// "smart-office/internal/domain/print"
)

// LoginRequest represents the expected JSON body for POST /api/auth/login
type LoginRequest struct {
	Type     string  `json:"type"`
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
}

// UserPaymentData represents payment info for the API response.
type UserPaymentData struct {
	ID            string    `json:"id"`
	Sum           string    `json:"sum"`
	PsID          int64     `json:"ps_id"`
	FopReceiptKey *string   `json:"fop_receipt_key,omitempty"`
	CardNumber    *string   `json:"card_number,omitempty"`
	Date          time.Time `json:"date"`
}

// FileData represents file info for the API response.
type FileData struct {
	Name       string    `json:"name"`
	SystemName string    `json:"system_name"`
	Date       time.Time `json:"date"`
	Status     string    `json:"status"`
	// PrintSettings print.PrintSettings `json:"print_settings"`
	// Add Fill?
}

// UserResponse is the structure returned by the /load endpoint
type UserResponse struct {
	ID                string            `json:"_id"`
	Name              string            `json:"name"`
	Files             []FileData        `json:"files"`
	FavouritePrinters []string          `json:"favouritePrinters"`
	Balance           string            `json:"balance"`
	Reserved          string            `json:"reserved"`
	Payments          []UserPaymentData `json:"payments"`
	DefaultPrinter    *string           `json:"defaultPrinter,omitempty"`
}

type RegisterRequest struct {
	Type     string  `json:"type"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type RegisterResponse struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Token    string `json:"token"`
	Balance  string `json:"balance"`
	Reserved string `json:"reserved"`
}

type LoginResponse struct {
	UserResponse
	Token string `json:"token"`
}
