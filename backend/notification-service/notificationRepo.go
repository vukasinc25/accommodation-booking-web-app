package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type NotificationRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	tracer trace.Tracer
}

func New(ctx context.Context, logger *log.Logger, tracer trace.Tracer) (*NotificationRepo, error) {

	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}
	return &NotificationRepo{
		cli: client, logger: logger, tracer: tracer,
	}, nil
}

func (ar *NotificationRepo) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ar *NotificationRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ar.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ar.logger.Println(err)
	}

	// Print available databases
	databases, err := ar.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ar *NotificationRepo) GetAll() (Notifications, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notificationCollection := ar.getCollection()

	var accommodations Notifications
	notificationCursor, err := notificationCollection.Find(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = notificationCursor.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return accommodations, nil
}

func (ar *NotificationRepo) GetAllByHostId(id string) (Notifications, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notificationCollection := ar.getCollection()

	var accommodations Notifications
	//objID, _ := primitive.ObjectIDFromHex(id)
	notificationCursor, err := notificationCollection.Find(ctx, bson.D{{"hostId", id}})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = notificationCursor.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return accommodations, nil
}

func (ar *NotificationRepo) Delete(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := ar.getCollection()

	// objID, _ := primitive.ObjectIDFromHex(username)
	filter := bson.M{"username": username}
	result, err := patientsCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func (ar *NotificationRepo) GetById(id string) (*Notification, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notificationCollection := ar.getCollection()

	var notification Notification
	objID, _ := primitive.ObjectIDFromHex(id)
	err := notificationCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&notification)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &notification, nil
}

func (ar *NotificationRepo) Insert(ctx context.Context, notification *Notification) error {
	ctx, span := ar.tracer.Start(ctx, "NotificationRepo.Insert")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationCollection := ar.getCollection()
	//notification.Date = time.Now()
	result, err := accommodationCollection.InsertOne(ctx, &notification)
	if err != nil {
		ar.logger.Println(err)
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ar *NotificationRepo) getCollection() *mongo.Collection {
	patientDatabase := ar.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("notifications")
	return patientsCollection
}
