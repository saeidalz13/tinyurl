package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MustConnectToDb(mongoUri string) (*mongo.Client, *mongo.Collection) {
	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoUri).SetServerAPIOptions(serverApi)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	err = client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		panic(err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")
	coll := client.Database("tinyurl").Collection("urls")

	ensureIndexExistenceOnShortenedUrl(coll)
	return client, coll
}

func DisconnectClient(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}

func ensureIndexExistenceOnShortenedUrl(coll *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"shortened_url": 1},      // Create an ascending index on shortened_url
		Options: options.Index().SetUnique(true), // Ensure that the shortened_url is unique
	}

	// List existing indexes
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		log.Fatal(err)
	}

	// Check if the index already exists
	for _, index := range indexes {
		if index["name"] == "shortened_url_1" {
			log.Println("Index shortened_url_1 already exists")
			return
		}
	}

	// Create the index if it doesn't exist
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Index created on shortened_url")
}
