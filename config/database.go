package config

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var Database *mongo.Database

func InitDB() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	Database = client.Database("hr-appraisal")

	fmt.Println("database connected ...")
}
