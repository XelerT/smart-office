package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	// files "smart-office/internal/domain/files"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type HashAlgorithm string

const (
	HashBcrypt HashAlgorithm = "bcrypt"
	HashMD5    HashAlgorithm = "md5"
)

// Payment data stored in the database. Pointer fields are used for optional fields.
type Payment struct {
	ID            string             `bson:"id" json:"id"`
	Sum           string             `bson:"sum" json:"sum"`
	Key           string             `bson:"key" json:"-"` // from JSON responses
	PsID          int64              `bson:"ps_id" json:"ps_id"`
	FopReceiptKey *string            `bson:"fop_receipt_key,omitempty" json:"fop_receipt_key,omitempty"`
	CardNumber    *string            `bson:"card_number,omitempty" json:"card_number,omitempty"`
	Date          primitive.DateTime `bson:"date" json:"date"`
}

// User represents the main user entity in the database.
type User struct {
	ID                primitive.ObjectID   `bson:"_id" json:"id"`
	CreatedAt         primitive.DateTime   `bson:"createdAt" json:"createdAt"`
	UpdatedAt         primitive.DateTime   `bson:"updatedAt" json:"updatedAt"` // Last access time
	Name              string               `bson:"name" json:"name"`
	Password          string               `bson:"password" json:"-"`
	Email             string               `bson:"email" json:"email"`
	Role              Role                 `bson:"role" json:"role"`
	Hash              HashAlgorithm        `bson:"hash" json:"hash"` // Password hash algorithm
	Balance           primitive.Decimal128 `bson:"balance" json:"balance"`
	Reserved          primitive.Decimal128 `bson:"reserved" json:"reserved"`
	FavouritePrinters []string             `bson:"favouritePrinters" json:"favouritePrinters"`
	// Files             []files.File       `bson:"files" json:"files"`
	Files          []primitive.M `bson:"files" json:"files"`
	Payments       []Payment     `bson:"payments,omitempty" json:"-"`
	DefaultPrinter *string       `bson:"defaultPrinter,omitempty" json:"defaultPrinter,omitempty"`
}

// Represents the payload in JWT
type TokenData struct {
	Name string `json:"name"`
	Role Role   `json:"role"`
	// UserID string `json:"user_id"` // Do we need it?
	// StandardClaims jwt.StandardClaims // Do we need it?
}

// --- DTOs (Data Transfer Objects) ---

// The payment information sent to the frontend
type UserPaymentData struct {
	ID            string    `json:"id"`
	Sum           string    `json:"sum"`
	PsID          int64     `json:"ps_id"`
	FopReceiptKey *string   `json:"fop_receipt_key,omitempty"`
	CardNumber    *string   `json:"card_number,omitempty"`
	Date          time.Time `json:"date"`
}
