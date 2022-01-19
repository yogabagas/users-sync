package repository

import (
	"context"
	"log"
	"my-github/users-sync/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertLog(ctx context.Context, data LogData) {
	collection := config.Database.Collection("users")

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Println(err.Error())
	}
}

func CreateOrUpdate(ctx context.Context, data *UserData) {
	collection := config.Database.Collection("users")
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"nik", data.NIK}}
	update := bson.D{{"$set", bson.D{{"name", data.Name}, {"role", data.Role}, {"directorate", data.Directorate}, {"status", data.Status}, {"description", data.Description}, {"created_at", data.CreatedAt}, {"updated_at", data.UpdatedAt}}}}
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println(err.Error())
	}
}
