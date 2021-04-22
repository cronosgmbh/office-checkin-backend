package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectToDB() *mongo.Database {
	cs := fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
		cfg.MongoDB.Username,
		cfg.MongoDB.Password,
		cfg.MongoDB.Host,
		cfg.MongoDB.Database)
	client, err := mongo.NewClient(options.Client().ApplyURI(cs))
	if err != nil {
		logrus.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
	return client.Database(cfg.MongoDB.Database)
}
