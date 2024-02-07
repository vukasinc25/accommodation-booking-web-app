package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AccoRepo struct {
	cli                         *mongo.Client
	logger                      *log.Logger
	reservation_service_address string
	tracer                      trace.Tracer
}

func New(ctx context.Context, logger *log.Logger, conn_reservation_service_address string, tracer trace.Tracer) (*AccoRepo, error) {

	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	return &AccoRepo{
		cli:                         client,
		logger:                      logger,
		reservation_service_address: conn_reservation_service_address,
		tracer:                      tracer,
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

func (ar *AccoRepo) GetAll(ctx context.Context) (Accommodations, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetAll")
	defer span.End()

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

func (ar *AccoRepo) GetAllByUsername(username string, ctx context.Context) (Accommodations, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetAllByUsername")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	//objID, _ := primitive.ObjectIDFromHex(id)
	accommoCursor, err := accommoCollection.Find(ctx, bson.D{{"username", username}})
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

func (ar *AccoRepo) GetAllById(id string, ctx context.Context) (Accommodations, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.Insert")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	// objID, _ := primitive.ObjectIDFromHex(id)
	accommoCursor, err := accommoCollection.Find(ctx, bson.D{{"_id", id}})
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

func (ar *AccoRepo) GetAllThatAreNotApproved() (Accommodations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodations Accommodations
	// objID, _ := primitive.ObjectIDFromHex(id)
	accommoCursor, err := accommoCollection.Find(ctx, bson.D{{"approved", "false"}})
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

func (ar *AccoRepo) GetAllRecommended(accoIds *ReqList) (*Accommodations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommoCollection := ar.getCollection()

	oids := make([]primitive.ObjectID, len(accoIds.List))
	for i := range accoIds.List {
		objID, err := primitive.ObjectIDFromHex(accoIds.List[i])
		if err == nil {
			oids = append(oids, objID)
		}
	}

	var accommodations Accommodations
	accoByLocationList, err := accommoCollection.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
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

func (ar *AccoRepo) Delete(username string, ctx context.Context) error {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.Delete")
	defer span.End()

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

func (ar *AccoRepo) UpdateAccommodation(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationCollection := ar.getCollection()
	// objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"approved": "true",
	}}
	result, err := accommodationCollection.UpdateOne(ctx, filter, update)
	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Println("Error ovde:", err)
		return err
	}

	return nil
}

func (ar *AccoRepo) GetById(id string, ctx context.Context) (*Accommodation, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetById")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollection()

	var accommodation Accommodation
	// objID, _ := primitive.ObjectIDFromHex(id)
	err := accommoCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&accommodation)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &accommodation, nil
}

func (ar *AccoRepo) DeleteById(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accomodationCollection := ar.getCollection()

	// objID, _ := primitive.ObjectIDFromHex(username)
	filter := bson.M{"_id": id}
	result, err := accomodationCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func (ar *AccoRepo) GetAllByLocation(location string, ctx context.Context) (*Accommodations, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetAllByLocation")
	defer span.End()

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

// Gets all acco_ids from the reservation service
func (ar *AccoRepo) GetAllAccoFromReservationServiceByDate(beginReservationDate string, endReservationDate string, ctx context.Context) (*http.Response, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetAllAccoFromReservationServiceByDate")
	defer span.End()

	url := ar.reservation_service_address + "/api/reservations/search_by_date/" + beginReservationDate + "/" + endReservationDate
	log.Println(url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return httpResp, nil
}

func (ar *AccoRepo) DeleteAccommodationGrade(userId string, id string, ctx context.Context) error {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.DeleteAccommodationGrade")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationGradeCollection := ar.getCollectionForAccommodationGrade()

	var accommodatioGrade AccommodationGrade
	err := accommodationGradeCollection.FindOne(ctx, bson.M{"_id": strings.TrimSpace(id)}).Decode(&accommodatioGrade)
	if err != nil {
		log.Println("Ove:", err)
		return err
	}

	log.Println("UserId:", accommodatioGrade.UserId)
	log.Println("UserId:", userId)
	if accommodatioGrade.UserId != strings.Trim(userId, `"`) {
		return errors.New("unauthorised")
	}

	filter := bson.M{"_id": id}
	result, err := accommodationGradeCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no accommodationGrade found for given id")
	}

	log.Printf("Documents deleted: %v\n", result.DeletedCount)

	err = ar.CreateAverageRating(accommodatioGrade.AccommodationId)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (ar *AccoRepo) GetAllByNoGuests(noGuestsString string, ctx context.Context) (*Accommodations, error) {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.GetAllByNoGuests")
	defer span.End()

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

func (ar *AccoRepo) CreateAverageRating(id string) error {
	//ctx, span := ar.tracer.Start(ctx, "AccoRepo.Insert")
	//defer span.End()

	log.Println("Usli u metodu")
	var averageRating float64
	grades, err := ar.GetAllAccommodationGrades(id)
	if err != nil {
		log.Println("Error in getAlAccommodationGrades method")
		return err
	}

	if grades == nil {
		log.Println("Accommodation grades for that id doesnt exists")
		return errors.New("accommodation greades for that id doesnt exists")
	}

	log.Println("Ovde", *grades)
	log.Println("Ovde", grades)

	for _, value := range *grades {
		log.Println("Grade:", value.Grade)
		averageRating += float64(value.Grade)
	}

	log.Println("Total Rating:", averageRating)

	if len(*grades) > 0 {
		averageRating /= float64(len(*grades))
	}
	log.Println("Average Rating:", averageRating)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationCollection := ar.getCollection()
	// objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"averageGrade": float64(averageRating),
	}}
	result, err := accommodationCollection.UpdateOne(ctx, filter, update)
	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Println("Error ovde:", err)
		return err
	}

	return nil
}

func (ar *AccoRepo) Insert(accommodation *Accommodation, ctx context.Context) error {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.Insert")
	defer span.End()

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accommodationCollection, err := ar.getCollection1()
	if err != nil {
		return errors.New("error in getting accommodation collection")
	}

	result, err := accommodationCollection.InsertOne(ctx, &accommodation)
	if err != nil {
		log.Println("Error when tryed to insert accommodation: ", err)
		return err
	}
	log.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ar *AccoRepo) SendRequestToReservationService(token string) (*http.Response, error) {
	url := ar.reservation_service_address + "/api/reservations/by_user"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := http.Client{Transport: tr}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return httpResp, nil
}

func (ar *AccoRepo) GetAllAccommodationGrades(id string) (*AccommodationGrades, error) {
	//ctx, span := ar.tracer.Start(ctx, "AccoRepo.Insert")
	//defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationGradeCollection := ar.getCollectionForAccommodationGrade()

	var accommodationGrades AccommodationGrades
	accommodationGradeList, err := accommodationGradeCollection.Find(ctx, bson.M{"accommodationId": id})
	log.Println(accommodationGradeList)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	if err = accommodationGradeList.All(ctx, &accommodationGrades); err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &accommodationGrades, nil
}

func (ar *AccoRepo) InsertAccommodationImg(id string, images []string, ctx context.Context) error {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.InsertAccommodationImg")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationCollection := ar.getCollection()
	// objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"images": images,
	}}
	result, err := accommodationCollection.UpdateOne(ctx, filter, update)
	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Println("Error ovde:", err)
		return err
	}

	return nil
}

func (ar *AccoRepo) GetAccommodationGradeByUserId(id string) (*AccommodationGrade, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommoCollection := ar.getCollectionForAccommodationGrade()

	var accommodationGrade AccommodationGrade
	// objID, _ := primitive.ObjectIDFromHex(id)
	err := accommoCollection.FindOne(ctx, bson.M{"userId": id}).Decode(&accommodationGrade)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	return &accommodationGrade, nil
}

func (ar *AccoRepo) getCollection1() (*mongo.Collection, error) {
	accommodationDatabase := ar.cli.Database("mongoDemo")
	accommodationCollection := accommodationDatabase.Collection("accommodations")
	name := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := accommodationCollection.Indexes().CreateOne(context.TODO(), name)
	if err != nil {
		log.Println("Error in creatingOne unique name index")
		return nil, err
	}
	return accommodationCollection, nil
}

func (ar *AccoRepo) UpdateAccommodationGradeByUserId(id string, grade int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accommodationCollection := ar.getCollectionForAccommodationGrade()
	// objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"userId": id}
	update := bson.M{"$set": bson.M{
		"grade": grade,
	}}
	result, err := accommodationCollection.UpdateOne(ctx, filter, update)
	log.Printf("Documents matched: %v\n", result.MatchedCount)
	log.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Println("Error ovde1:", err)
		return err
	}

	return nil
}

func (ar *AccoRepo) CreateAvailabilityPeriods(token string, reservation_by_acco *ReservationByAccommodation) (*http.Response, error) {
	url := ar.reservation_service_address + "/api/reservations/for_acco"

	reqBody, err := json.Marshal(reservation_by_acco)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := http.Client{Transport: tr}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return httpResp, nil
}

func (ar *AccoRepo) getCollection() *mongo.Collection {
	accommodationDatabase := ar.cli.Database("mongoDemo")
	accommodationCollection := accommodationDatabase.Collection("accommodations")
	return accommodationCollection
}

func (ar *AccoRepo) CreateGrade(accommodatioGrade *AccommodationGrade, token string, ctx context.Context) error {
	ctx, span := ar.tracer.Start(ctx, "AccoRepo.CreateGrade")
	defer span.End()

	log.Println("Usli u CreateGrade")

	grade, err := ar.GetAccommodationGradeByUserId(accommodatioGrade.UserId)
	if err != nil {
		ar.logger.Println("Error in GetAccommodationGradeByUserId func in CreateGrade", err)
		// return err
	}

	log.Println("Grade: ", grade)

	if grade == nil { // create accommodation grade
		response, err := ar.SendRequestToReservationService(token)
		if err != nil {
			log.Println("Error in SendRequestToReservationService method", err)
			return err
		}

		var userReservations ReservationsByUser
		if err := json.NewDecoder(response.Body).Decode(&userReservations); err != nil {
			log.Println("Cant decode userReservatins", err)
			return err
		}

		if userReservations == nil {
			log.Println("userReservation are empty")
			return errors.New("user with thid id dont have any reservations")
		}

		var bool = false
		for _, reservation := range userReservations {
			log.Println("Reservation:", reservation)
			log.Println("Reservation.AccoId:", reservation.AccoId)
			log.Println("Reservation.AccommodationId:", accommodatioGrade.AccommodationId)
			if strings.TrimSpace(reservation.AccoId) == strings.TrimSpace(accommodatioGrade.AccommodationId) {
				bool = true
				break
			}
		}

		if !bool {
			log.Println("check if user have reservations for accommodation")
			return errors.New("user dont have any reservations for this accommodation")
		}
		bool = false

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		accommodationCollection := ar.getCollectionForAccommodationGrade()

		result, err := accommodationCollection.InsertOne(ctx, &accommodatioGrade)
		if err != nil {
			ar.logger.Println(err)
			return err
		}
		ar.logger.Printf("Documents ID: %v\n", result.InsertedID)

		err = ar.CreateAverageRating(accommodatioGrade.AccommodationId)
		if err != nil {
			log.Println(err)
			return err
		}
	} else { // update accommodation grade
		err := ar.UpdateAccommodationGradeByUserId(accommodatioGrade.UserId, accommodatioGrade.Grade)
		if err != nil {
			ar.logger.Println("Error in UpdateAccommodationGradeByUserId func in CreateGrade")
			return err
		}
	}

	return nil
}

func (ar *AccoRepo) getCollectionForAccommodationGrade() *mongo.Collection {
	accommodationGradeDatabase := ar.cli.Database("mongoDemo")
	accommodationGradeCollection := accommodationGradeDatabase.Collection("accommodationsGrades")
	return accommodationGradeCollection
}
