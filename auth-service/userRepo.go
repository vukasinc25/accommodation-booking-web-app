package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

func New(ctx context.Context, logger *log.Logger) (*UserRepo, error) {
	//dburi := "mongodb+srv://mongo:mongo@cluster0.gdaah26.mongodb.net/?retryWrites=true&w=majority"

	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (uh *UserRepo) Disconnect(ctx context.Context) error {
	err := uh.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (uh *UserRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := uh.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		uh.logger.Println(err)
	}

	// Print available databases
	databases, err := uh.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		uh.logger.Println(err)
	}
	fmt.Println(databases)
}

func (uh *UserRepo) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := uh.getCollection()

	result, err := usersCollection.InsertOne(ctx, &user)
	if err != nil {
		uh.logger.Println(err)
		return err
	}
	uh.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (uh *UserRepo) GetAll() (Users, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := uh.getCollection()
	uh.logger.Println("Collection: ", usersCollection)

	var users Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		uh.logger.Println("Cant find userCollection: ", err)
		return nil, err
	}
	if err = usersCursor.All(ctx, &users); err != nil {
		uh.logger.Println("User Cursor.All: ", err)
		return nil, err
	}
	return users, nil
}

func (ur *UserRepo) GetByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var user User
	log.Println("UsersCollection: ", usersCollection)
	log.Println("UserName: ", username)
	// objUsername, _ := primitive.ObjectIDFromHex(username)
	err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &user, nil
}

func (uh *UserRepo) getCollection() *mongo.Collection {
	userDatabase := uh.cli.Database("mongoDemo")
	usersCollection := userDatabase.Collection("users")
	return usersCollection
}
