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
	collection := config.Database.Collection("usersTest")
	filter := bson.M{"nik": data.NIK}
	update := bson.M{"$set": bson.M{"status": data.Status, "description": data.Description, "updated_at": time.Now()}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err.Error())
	}
}

func CreateOrUpdate(ctx context.Context, data *UserData) {
	collection := config.Database.Collection("usersTest")
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"nik", data.NIK}}
	update := bson.D{{"$set", bson.D{{"name", data.Name}, {"role", data.Role}, {"directorate", data.Directorate}, {"status", data.Status}, {"description", data.Description}, {"created_at", data.CreatedAt}, {"updated_at", data.UpdatedAt}}}}
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println(err.Error())
	}
}

func ReadFromLocalDB(ctx context.Context, limit, skip int64) (resp []UserData, err error) {
	collection := config.Database.Collection("usersTest")

	filter := bson.D{{"status", 0}}
	optionFind := options.Find().SetLimit(limit).SetSkip(skip)

	cur, err := collection.Find(ctx, filter, optionFind)
	if err != nil {
		return
	}

	if err = cur.All(ctx, &resp); err != nil {
		return
	}

	return
}
