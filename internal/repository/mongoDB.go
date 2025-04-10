package repository

import (
	"context"
	"errors"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrDuplicateKey     = errors.New("duplicate key error")
)

type MongoRepository[T any] struct {
	collection *mongo.Collection
}

func NewMongoRepository[T any](collection *mongo.Collection) *MongoRepository[T] {
	return &MongoRepository[T]{
		collection: collection,
	}
}

func (r *MongoRepository[T]) CreateIndex(field string, unique bool) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: field, Value: 1}},
		Options: options.Index().SetUnique(unique),
	}

	_, err := r.collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Printf("Error creating index on field %s: %v", field, err)
		return err
	}
	return nil
}

func (r *MongoRepository[T]) Create(ctx context.Context, entity T) error {
	_, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		return err
	}
	return nil
}

func (r *MongoRepository[T]) FindByID(ctx context.Context, id string, idField string) (*T, error) {
	var entity T
	filter := bson.M{idField: id}

	err := r.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return &entity, nil
}

func (r *MongoRepository[T]) FindOne(ctx context.Context, filter bson.M) (*T, error) {
	var entity T

	err := r.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return &entity, nil
}

func (r *MongoRepository[T]) Find(ctx context.Context, filter bson.M) ([]T, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Println("Erro ao fechar o cursor:", err)
		}
	}()

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoRepository[T]) Update(ctx context.Context, id string, idField string, update bson.M) error {
	filter := bson.M{idField: id}

	// Set updated_at timestamp if the entity has this field
	var entity T
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Struct {
		_, hasUpdatedAt := entityType.FieldByName("UpdatedAt")
		if hasUpdatedAt && update["$set"] != nil {
			if updateSet, ok := update["$set"].(bson.M); ok {
				updateSet["updated_at"] = time.Now()
			}
		}
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (r *MongoRepository[T]) Delete(ctx context.Context, id string, idField string) error {
	filter := bson.M{idField: id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}
