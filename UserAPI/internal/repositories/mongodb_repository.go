package repositories

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewMongoDBRepository(uri string, databaseName string, collectionName string) (*MongoDBRepository, error) {
	clientOptions := options.Client().ApplyURI(uri)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to database client: %v", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	database := client.Database(databaseName)
	collection := database.Collection(collectionName)
	return &MongoDBRepository{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (r *MongoDBRepository) Close(ctx context.Context) error {
	if err := r.client.Disconnect(ctx); err != nil {
		log.Fatalf("Failed to disconnect from database client: %v", err)
		return err
	}
	return nil
}

type GenericResponse struct {
	Success	bool
	Data	[]map[string]interface{}
}

func (r *MongoDBRepository) Find(ctx context.Context, filter interface{}) (*GenericResponse, error) {
	opts := options.Find().SetProjection(bson.M{"_id": 0})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return &GenericResponse{
			Success: false,
			Data: nil,
		}, err
	}
	result := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			log.Println(err)
		}
		result = append(result, doc)
	}
	return &GenericResponse{
		Success: true,
		Data: result,
	}, nil
}

func (r *MongoDBRepository) Insert(ctx context.Context, document interface{}) (*GenericResponse, error) {
	_, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return &GenericResponse{
			Success: false,
			Data: nil,
		}, err
	}
	return &GenericResponse{
		Success: true,
		Data: nil,
	}, nil
}

func (r *MongoDBRepository) Update(ctx context.Context, filter interface{}, update interface{}) (*GenericResponse, error) {
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return &GenericResponse{
			Success: false,
			Data: nil,
		}, err
	}
	return &GenericResponse{
		Success: true,
		Data: nil,
	}, nil
}

func (r *MongoDBRepository) Delete(ctx context.Context, filter interface{}) (*GenericResponse, error) {
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return &GenericResponse{
			Success: false,
			Data: nil,
		}, err
	}
	return &GenericResponse{
		Success: true,
		Data: nil,
	}, nil
}
