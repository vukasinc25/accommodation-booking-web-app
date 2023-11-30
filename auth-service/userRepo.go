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
	usersCollection, err := uh.getCollection()
	if err != nil {
		log.Println("Duplicate key error: ", err)
		return err
	}
	result, err := usersCollection.InsertOne(ctx, &user)
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

	usersCollection, err := uh.getCollection()
	if err != nil {
		log.Println("Duplicate key error: ", err)
		return nil, err
	}

	uh.logger.Println("Collection: ", usersCollection)

	var users Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Cant find userCollection: ", err)
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

	usersCollection, err := uh.getCollection()
	if err != nil {
		log.Println("Error getting collection: ", err)
		return nil, err
	}
	var user User
	log.Println("Querying for user with username: ", username)
	// objUsername, _ := primitive.ObjectIDFromHex(username)
	err = usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		log.Println("Error decoding user document: ", err)
		return nil, err
	}
	return &user, nil
}

func (uh *UserRepo) CreateVerificationEmail(verificationEmil VerifyEmail) error { // MORA  DA PRIMA POINTER NA VERIFYEMAIL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	verificationCopy := verificationEmil

	verification := uh.getEmailCollection()
	result, err := verification.InsertOne(ctx, &verificationCopy)
	if err != nil {
		uh.logger.Println(err)
		return err
	}
	uh.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (uh *UserRepo) GetVerificationEmailByCode(code string) (*VerifyEmail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verifycationEmail VerifyEmail
	verifycationEmailCollection := uh.getEmailCollection()
	err := verifycationEmailCollection.FindOne(ctx, bson.M{"secretCode": code}).Decode(&verifycationEmail)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &verifycationEmail, nil
}

func (uh *UserRepo) IsVerificationEmailActive(code string) (bool, error) {
	verificationEmail, err := uh.GetVerificationEmailByCode(code)
	if err != nil {
		return false, err
	}

	currentTime := time.Now()
	return currentTime.Before(verificationEmail.ExpiredAt), nil
}

func (uh *UserRepo) GetVerificationEmailByUsername(username string) (*VerifyEmail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verifycationEmail VerifyEmail
	verifycationEmailCollection := uh.getEmailCollection()
	err := verifycationEmailCollection.FindOne(ctx, bson.M{"username": username}).Decode(&verifycationEmail)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &verifycationEmail, nil
}

func (uh *UserRepo) UpdateUsersVerificationEmail(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	userCollection, err := uh.getCollection()
	if err != nil {
		log.Println("Cant get User collection in UpdateUserVerificationEmail method")
		return err
	}

	// objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{
		"isEmailVerified": true,
	}}
	result, err := userCollection.UpdateOne(ctx, filter, update)
	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (uh *UserRepo) UpdateVerificationEmail(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	emailVerificationCollection := uh.getEmailCollection()

	filter := bson.M{"secretCode": code}
	update := bson.M{"$set": bson.M{
		"isUsed": true,
	}}

	log.Printf("Updating verification email with code: %s\n", code)
	result, err := emailVerificationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating verification email: %v\n", err)
		return err
	}

	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	return nil
}

func (uh *UserRepo) getCollection() (*mongo.Collection, error) {
	userDatabase := uh.cli.Database("mongoDemo")
	usersCollection := userDatabase.Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := usersCollection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Println("Error in creatingOne unique username index")
		return nil, err
	}
	return usersCollection, nil
}

func (uh *UserRepo) getEmailCollection() *mongo.Collection {
	userDatabase := uh.cli.Database("mongoDemo")
	usersCollection := userDatabase.Collection("verificatonEmails")
	return usersCollection
}
