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

func MustConnectToDb(mongoUri string) (*mongo.Client, *mongo.Collection) {
	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoUri).SetServerAPIOptions(serverApi)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Send a ping to confirm a successful connection
	err = client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	coll := client.Database("tinyurl").Collection("urls")

	return client, coll
}

func DisconnectClient(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
