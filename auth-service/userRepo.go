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

// UserRepo is a repository for MongoDB operations related to User.
type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// New creates a new UserRepo instance.
func New(ctx context.Context, logger *log.Logger) (*UserRepo, error) {
	dbURI := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect disconnects from the MongoDB client.
func (uh *UserRepo) Disconnect(ctx context.Context) error {
	err := uh.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Ping checks the MongoDB connection and prints available databases.
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

// Insert inserts a new user into the MongoDB collection.
func (uh *UserRepo) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := uh.getCollection()

	result, err := usersCollection.InsertOne(ctx, user)
	if err != nil {
		uh.logger.Println(err)
		return err
	}
	uh.logger.Printf("Document ID: %v\n", result.InsertedID)
	return nil
}

// GetAll retrieves all users from the MongoDB collection.
func (uh *UserRepo) GetAll() (Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := uh.getCollection()

	var users Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		uh.logger.Println("Cannot find user collection: ", err)
		return nil, err
	}
	if err = usersCursor.All(ctx, &users); err != nil {
		uh.logger.Println("User Cursor.All: ", err)
		return nil, err
	}
	return users, nil
}

// GetByUsername retrieves a user by username from the MongoDB collection.
func (uh *UserRepo) GetByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := uh.getCollection()

	var user User
	log.Println("Users Collection: ", usersCollection)
	log.Println("Username: ", username)
	err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		uh.logger.Println(err)
		return nil, err
	}
	return &user, nil
}

// getCollection returns the MongoDB collection.
func (uh *UserRepo) getCollection() *mongo.Collection {
	userDatabase := uh.cli.Database("mongoDemo")
	usersCollection := userDatabase.Collection("users")
	return usersCollection
}
