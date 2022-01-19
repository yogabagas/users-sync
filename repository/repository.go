package repository

import (
	"context"
	"log"
	"my-github/users-sync/config"
)

func InsertLog(ctx context.Context, data LogData) {
	collection := config.Database.Collection("users")

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Println(err.Error())
	}
}
