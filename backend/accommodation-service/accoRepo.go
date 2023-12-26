package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AccoRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

func New(ctx context.Context, logger *log.Logger) (*AccoRepo, error) {

	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	return &AccoRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (ar *AccoRepo) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ar *AccoRepo) Ping() {
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

func (ar *AccoRepo) GetAll() (Accommodations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	accommoCursor, err := accommoCollection.Find(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = accommoCursor.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return accommodations, nil
}

func (ar *AccoRepo) GetAllById(id string) (Accommodations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	//objID, _ := primitive.ObjectIDFromHex(id)
	accommoCursor, err := accommoCollection.Find(ctx, bson.D{{"username", id}})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = accommoCursor.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return accommodations, nil
}

func (ar *AccoRepo) GetById(id string) (*Accommodation, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodation Accommodation
	objID, _ := primitive.ObjectIDFromHex(id)
	err := accommoCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&accommodation)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &accommodation, nil
}

func (ar *AccoRepo) GetAllByLocation(location string) (*Accommodations, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	accoByLocationList, err := accommoCollection.Find(ctx, bson.M{"city": location})
	log.Println(accoByLocationList)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = accoByLocationList.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &accommodations, nil
}

func (ar *AccoRepo) GetAllByNoGuests(noGuestsString string) (*Accommodations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	noGuests, err := strconv.Atoi(noGuestsString)
	if err != nil {
		log.Println("SJEBO SE ABUUU", err)
		return nil, err
	}

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	accoByGuestList, err := accommoCollection.Find(ctx, bson.M{"minGuests": bson.M{"$lte": noGuests}, "maxGuests": bson.M{"$gte": noGuests}})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = accoByGuestList.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &accommodations, nil
}

func (ar *AccoRepo) Insert(accommodation *Accommodation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationCollection := ar.getCollection()

	result, err := accommodationCollection.InsertOne(ctx, &accommodation)
	if err != nil {
		ar.logger.Println(err)
		return err
	}
	ar.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ar *AccoRepo) getCollection() *mongo.Collection {
	patientDatabase := ar.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("accommodations")
	return patientsCollection
}
