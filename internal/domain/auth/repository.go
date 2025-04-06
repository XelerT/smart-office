package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Defines the expected persistence operations for User data.
type Repository interface {
	// Inserts a new user into the database.
	// It should handle hashing internally or expect a hashed password.
	// Should return apperrors.ErrConflict if user already exists
	CreateUser(ctx context.Context, user *User) error

	// Finds a user based on login criteria
	// Returns apperrors.ErrNotFound if no user matches.
	FindUserForLogin(ctx context.Context, criteria LoginCriteria) (*User, error)

	// Finds a user by their email address.
	// Returns apperrors.ErrNotFound if not found.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// Finds a user by their unique name.
	// Returns apperrors.ErrNotFound if not found.
	FindByName(ctx context.Context, name string) (*User, error)

	// Updates the user's password hash and hashing algorithm type.
	UpdateUserPassword(ctx context.Context, userID primitive.ObjectID, newHashedPassword string, newAlgorithm HashAlgorithm) error

	// Maybe needed later?
	// FindByID(ctx context.Context, userID primitive.ObjectID) (*User, error)
	// UpdateUser(ctx context.Context, user *User) error
}
