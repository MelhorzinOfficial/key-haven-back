package repository

import (
	"context"
	"errors"
	"key-haven-back/internal/domain/user"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrCredentialNotFound   = errors.New("credential not found")
	ErrCredentialNameExists = errors.New("credential with this name already exists in this vault")
)

type CredentialRepository interface {
	Create(ctx context.Context, credential *user.Credential) error
	FindByID(ctx context.Context, id string) (*user.Credential, error)
	FindByName(ctx context.Context, vaultID, name string) (*user.Credential, error)
	FindAllByVaultID(ctx context.Context, vaultID string) ([]*user.Credential, error)
	Update(ctx context.Context, credential *user.Credential) error
	Delete(ctx context.Context, id string) error
}

// MongoCredentialRepository implements CredentialRepository using MongoDB
type MongoCredentialRepository struct {
	repo *MongoRepository[user.Credential]
}

func NewCredentialRepository(database *mongo.Database) CredentialRepository {
	collection := database.Collection("credentials")
	repo := NewMongoRepository[user.Credential](collection)

	// Create a custom compound index directly on the collection
	// This works because we're using the collection directly, not through the repo
	_, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "vault_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Printf("Warning: Error creating compound index for credentials: %v", err)
	}

	return &MongoCredentialRepository{
		repo: repo,
	}
}

// Create adds a new credential to the database
func (r *MongoCredentialRepository) Create(ctx context.Context, credential *user.Credential) error {
	if credential.ID == "" {
		credential.ID = uuid.New().String()
	}

	// Check if credential with the same name already exists in this vault
	existingCred, err := r.FindByName(ctx, credential.VaultID, credential.Name)
	if err == nil && existingCred != nil {
		return ErrCredentialNameExists
	}

	return r.repo.Create(ctx, *credential)
}

// FindByID retrieves a credential by its ID
func (r *MongoCredentialRepository) FindByID(ctx context.Context, id string) (*user.Credential, error) {
	credential, err := r.repo.FindByID(ctx, id, "_id")
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}
	return credential, nil
}

// FindByName retrieves a credential by vault ID and credential name
func (r *MongoCredentialRepository) FindByName(ctx context.Context, vaultID, name string) (*user.Credential, error) {
	filter := bson.M{"vault_id": vaultID, "name": name}
	credential, err := r.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}
	return credential, nil
}

// FindAllByVaultID retrieves all credentials in a vault
func (r *MongoCredentialRepository) FindAllByVaultID(ctx context.Context, vaultID string) ([]*user.Credential, error) {
	filter := bson.M{"vault_id": vaultID}
	credentials, err := r.repo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]*user.Credential, len(credentials))
	for i, credential := range credentials {
		credential := credential // create a new variable to avoid issues with the closure
		result[i] = &credential
	}
	return result, nil
}

func (r *MongoCredentialRepository) Update(ctx context.Context, credential *user.Credential) error {
	// Check if the updated name conflicts with an existing credential in the same vault
	existingCred, err := r.FindByName(ctx, credential.VaultID, credential.Name)
	if err == nil && existingCred != nil && existingCred.ID != credential.ID {
		return ErrCredentialNameExists
	}

	update := bson.M{
		"$set": bson.M{
			"name":             credential.Name,
			"description":      credential.Description,
			"username":         credential.Username,
			"encrypt_password": credential.EncryptPassword,
			"url":              credential.URL,
			"vault_id":         credential.VaultID, // Allow moving to another vault
			"updated_at":       credential.UpdatedAt,
		},
	}

	return r.repo.Update(ctx, credential.ID, "_id", update)
}

func (r *MongoCredentialRepository) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id, "_id")
}
