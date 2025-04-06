package mongo

import (
	"context"
	"errors"
	"fmt"

	// Import necessary types
	"smart-office/internal/domain/auth"
	apperrors "smart-office/internal/platform/errors"
	"smart-office/internal/platform/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const usersCollection = "users"

type MongoUserRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		db:         db,
		collection: db.Collection(usersCollection),
	}
}

// --- Implement Interface Methods ---

func (r *MongoUserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		// Check for MongoDB duplicate key error (code 11000)
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, we := range writeErr.WriteErrors {
				if we.Code == 11000 {
					logger.Log.Warnf("MongoRepo: CreateUser conflict (duplicate key) for user %s / %s", user.Name, user.Email)
					return fmt.Errorf("%w: user already exists", apperrors.ErrConflict)
				}
			}
		}
		logger.Log.Errorf("MongoRepo: Failed to insert user %s: %v", user.Name, err)
		return fmt.Errorf("database error creating user: %w", err)
	}
	logger.Log.Infof("MongoRepo: Successfully created user %s (%s)", user.Name, user.ID.Hex())
	return nil
}

func (r *MongoUserRepository) FindUserForLogin(ctx context.Context, criteria auth.LoginCriteria) (*auth.User, error) {
	var filter bson.M

	switch criteria.Type {
	case auth.LoginRegular:
		filter = bson.M{"name": criteria.Name}
	default:
		return nil, fmt.Errorf("unknown login criteria type")
	}

	var user auth.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: user not found for criteria %+v", apperrors.ErrNotFound, criteria)
		}
		logger.Log.Errorf("MongoRepo: Error finding user for login %+v: %v", criteria, err)
		return nil, fmt.Errorf("database error finding user: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	filter := bson.M{"email": email}
	var user auth.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: user not found with email %s", apperrors.ErrNotFound, email)
		}
		logger.Log.Errorf("MongoRepo: Error finding user by email %s: %v", email, err)
		return nil, fmt.Errorf("database error finding user by email: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) FindByName(ctx context.Context, name string) (*auth.User, error) {
	filter := bson.M{"name": name}
	var user auth.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: user not found with name %s", apperrors.ErrNotFound, name)
		}
		logger.Log.Errorf("MongoRepo: Error finding user by name %s: %v", name, err)
		return nil, fmt.Errorf("database error finding user by name: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) UpdateUserPassword(ctx context.Context, userID primitive.ObjectID, newHashedPassword string, newAlgorithm auth.HashAlgorithm) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"password": newHashedPassword,
			"hash":     newAlgorithm,
			// Consider updating UpdatedAt timestamp here as well
			// "updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Log.Errorf("MongoRepo: Failed to update password for user %s: %v", userID.Hex(), err)
		return fmt.Errorf("database error updating password: %w", err)
	}
	if result.MatchedCount == 0 {
		logger.Log.Warnf("MongoRepo: UpdateUserPassword - User %s not found.", userID.Hex())
		return fmt.Errorf("%w: user not found during password update", apperrors.ErrNotFound)
	}
	logger.Log.Infof("MongoRepo: Updated password for user %s", userID.Hex())
	return nil
}

var _ auth.Repository = (*MongoUserRepository)(nil)
