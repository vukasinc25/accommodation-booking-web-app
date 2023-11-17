package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type LocationRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

func New(ctx context.Context, logger *log.Logger) (*LocationRepo, error) {
	// dburi := "mongodb+srv://mongo:mongo@cluster0.gdaah26.mongodb.net/?retryWrites=true&w=majority"

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &LocationRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (uh *LocationRepo) Disconnect(ctx context.Context) error {
	err := uh.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pr *LocationRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := pr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		pr.logger.Println(err)
	}

	// Print available databases
	databases, err := pr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		pr.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ur *LocationRepo) Insert(patient *Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := ur.getCollection()

	result, err := patientsCollection.InsertOne(ctx, &patient)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (pr *LocationRepo) GetAll() (Locations, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locationsCollection := pr.getCollection()
	pr.logger.Println("Collection: ", locationsCollection)

	var locations Locations
	locationsCursor, err := locationsCollection.Find(ctx, bson.M{})
	if err != nil {
		pr.logger.Println("Cant find locationCollection: ", err)
		return nil, err
	}
	if err = locationsCursor.All(ctx, &locations); err != nil {
		pr.logger.Println("Location Cursor.All: ", err)
		return nil, err
	}
	return locations, nil
}

func (pr *LocationRepo) getCollection() *mongo.Collection {
	patientDatabase := pr.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("locations")
	return patientsCollection
}
