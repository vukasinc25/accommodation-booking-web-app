package main

import (
	"context"
	"fmt"
	"log"
<<<<<<< HEAD
	"net/http"
	"os"
=======
>>>>>>> 695427e4ea224977fd574165d928dfd42ba0902c
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ReservationRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

<<<<<<< HEAD
func New(logger *log.Logger) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")
	log.Println(db)
	log.Println("A sto ne radi")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	cluster.Timeout = time.Second * 55
	session, err := cluster.CreateSession()
=======
func New(ctx context.Context, logger *log.Logger) (*ReservationRepo, error) {
	// dburi := "mongodb+srv://mongo:mongo@cluster0.gdaah26.mongodb.net/?retryWrites=true&w=majority"

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
>>>>>>> 695427e4ea224977fd574165d928dfd42ba0902c
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ReservationRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (uh *ReservationRepo) Disconnect(ctx context.Context) error {
	err := uh.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pr *ReservationRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

<<<<<<< HEAD
// 	// Check connection -> if no error, connection is established
// 	err := pr.cli.Ping(ctx, readpref.Primary())
// 	if err != nil {
// 		pr.logger.Println(err)
// 	}

// 	// Print available databases
// 	databases, err := pr.cli.ListDatabaseNames(ctx, bson.M{})
// 	if err != nil {
// 		pr.logger.Println(err)
// 	}
// 	fmt.Println(databases)
// }

func (rs *ReservationRepo) CreateTables() {
	err := rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(acco_id UUID, reservation_id UUID, price int, date date, isDeleted boolean,
					PRIMARY KEY ((acco_id, reservation_id), date, price))
					WITH CLUSTERING ORDER BY (date DESC, price ASC)`,
			"reservations_by_acco")).Exec()
=======
	// Check connection -> if no error, connection is established
	err := pr.cli.Ping(ctx, readpref.Primary())
>>>>>>> 695427e4ea224977fd574165d928dfd42ba0902c
	if err != nil {
		pr.logger.Println(err)
	}

<<<<<<< HEAD
	err = rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(user_id UUID, reservation_id UUID, price int, date date, isDeleted boolean,
					PRIMARY KEY ((user_id, reservation_id), date, price))
					WITH CLUSTERING ORDER BY (date DESC, price ASC)`,
			"reservations_by_user")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}
}

// -------Reservation By Accommodation-------//
func (rs *ReservationRepo) GetReservationsByAcco(acco_id string) (ReservationsByAccommodation, error) {
	scanner := rs.session.Query(`SELECT acco_id, reservation_id, price, date, isDeleted
	 FROM reservations_by_acco WHERE acco_id = ? AND isDeleted = false ALLOW FILTERING;`,
		acco_id).Iter().Scanner()
=======
	// Print available databases
	databases, err := pr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		pr.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ur *ReservationRepo) Insert(patient *Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := ur.getCollection()
>>>>>>> 695427e4ea224977fd574165d928dfd42ba0902c

	result, err := patientsCollection.InsertOne(ctx, &patient)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (pr *ReservationRepo) GetAll() (Reservations, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reservationsCollection := pr.getCollection()
	pr.logger.Println("Collection: ", reservationsCollection)

	var reservations Reservations
	reservationsCursor, err := reservationsCollection.Find(ctx, bson.M{})
	if err != nil {
		pr.logger.Println("Cant find reservationCollection: ", err)
		return nil, err
	}
	if err = reservationsCursor.All(ctx, &reservations); err != nil {
		pr.logger.Println("Reservation Cursor.All: ", err)
		return nil, err
	}
	return reservations, nil
}

<<<<<<< HEAD
func (rs *ReservationRepo) Aaa(v http.ResponseWriter, req *http.Request) {
	log.Println("aezakmi")
}

func (rs *ReservationRepo) InsertReservationByAcco(resAcco *ReservationByAccommodation) error {
	reservationId, _ := gocql.RandomUUID()
	err := rs.session.Query(
		`INSERT INTO reservations_by_acco (acco_id, reservation_id, price, date) VALUES 
		(?, ?, ?, ?);`,
		resAcco.AccoId, reservationId, resAcco.Price, resAcco.Date).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

// -------Reservation By User-------//
func (rs *ReservationRepo) GetReservationsByUser(user_id string) (ReservationsByUser, error) {
	scanner := rs.session.Query(`SELECT user_id, reservation_id, price, date, isDeleted
	FROM reservations_by_acco WHERE user_id = ? AND isDeleted = false ALLOW FILTERING;`,
		user_id).Iter().Scanner()

	var reservations ReservationsByUser
	for scanner.Next() {
		var res ReservationByUser
		err := scanner.Scan(&res.UserId, &res.ReservationId, &res.Price, &res.Date, &res.IsDeleted)
		if err != nil {
			rs.logger.Println(err)
			return nil, err
		}
		reservations = append(reservations, &res)
	}
	if err := scanner.Err(); err != nil {
		rs.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

func (rs *ReservationRepo) InsertReservationByUser(resUser *ReservationByUser) error {
	reservationId, _ := gocql.RandomUUID()
	err := rs.session.Query(
		`INSERT INTO reservations_by_user (user_id, reservation_id, price, date) 
		VALUES (?, ?, ?, ?)`,
		resUser.UserId, reservationId, resUser.Price, resUser.Date).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

//--------------//

func (rs *ReservationRepo) UpdateReservationByAcco(accoId string, reservationId string, price string) error {
	// za Update je neophodno da pronadjemo vrednost po PRIMARNOM KLJUCU = PK + CK (ukljucuje sve kljuceve particije i klastera)
	// u ovom slucaju: PK = smerId, CK = student_id, indeks
	err := rs.session.Query(
		`UPDATE reservations_by_acco SET isDeleted = 1 where acoo_id = ? and reservation_id = ?`,
		accoId, reservationId).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

func (rs *ReservationRepo) GetDistinctIds(idColumnName string, tableName string) ([]string, error) {
	scanner := rs.session.Query(
		fmt.Sprintf(`SELECT DISTINCT %s FROM %s`, idColumnName, tableName)).
		Iter().Scanner()
	var ids []string
	for scanner.Next() {
		var id string
		err := scanner.Scan(&id)
		if err != nil {
			rs.logger.Println(err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := scanner.Err(); err != nil {
		rs.logger.Println(err)
		return nil, err
	}
	return ids, nil
=======
func (pr *ReservationRepo) getCollection() *mongo.Collection {
	patientDatabase := pr.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("reservations")
	return patientsCollection
>>>>>>> 695427e4ea224977fd574165d928dfd42ba0902c
}
