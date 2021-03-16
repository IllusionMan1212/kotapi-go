package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// package-wide variables for the collection and the context
var Kots *mongo.Collection
var Ctx = context.TODO()

func InitializeDB() {
	clientOptions := options.Client().ApplyURI(os.Getenv("DB_URI"))
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to mongo")

	Kots = client.Database("kotapi").Collection("kots")

	// creates a unique index on the _id field, which also auto generates it
	Kots.Indexes().CreateOne(Ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// creates a unique index on the id field
	Kots.Indexes().CreateOne(Ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// creates a unique index on the id field
	Kots.Indexes().CreateOne(Ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "url", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// creates a unique index on the id field
	Kots.Indexes().CreateOne(Ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "compressed_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}
