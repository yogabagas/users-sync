package repository

import (
	"context"
	"log"
	"my-github/users-sync/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateStatus(ctx context.Context, data LogData) {
	collection := config.Database.Collection("users")
	filter := bson.M{"nik": data.NIK}
	update := bson.M{"$set": bson.M{"status": data.Status, "description": data.Description, "updatedAt": time.Now()}}
	_, err := collection.UpdateOne(ctx, filter, update)
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