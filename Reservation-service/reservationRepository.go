package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gocql/gocql"
)

type ReservationRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

func New(logger *log.Logger) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "reservation", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	cluster.Keyspace = "reservation"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &ReservationRepo{
		session: session,
		logger:  logger,
	}, nil
}

func (rs *ReservationRepo) CloseSession() {
	rs.session.Close()
}

// func (pr *ReservationRepo) Ping() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

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
					PRIMARY KEY ((acco_id, reservation_id), price)) 
					WITH CLUSTERING ORDER BY (price ASC, date DESC)`,
			"reservations_by_acco")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}

	err = rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(user_id UUID, reservation_id UUID, price int, date date, isDeleted boolean,
					PRIMARY KEY ((user_id, reservation_id), price)) 
					WITH CLUSTERING ORDER BY (price ASC, date DESC)`,
			"reservations_by_user")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}
	if err != nil {
		rs.logger.Println(err)
	}
}

// -------Reservation By Accommodation-------//
func (rs *ReservationRepo) GetReservationsByAcco(acco_id string) (ReservationsByAccommodation, error) {
	scanner := rs.session.Query(`SELECT acco_id, reservation_id, price, date, isDeleted,
								FROM reservations_by_acco WHERE acco_id = ? AND isDeleted = 0`,
		acco_id).Iter().Scanner()

	var reservations ReservationsByAccommodation
	for scanner.Next() {
		var res ReservationByAccommodation
		err := scanner.Scan(&res.AccoId, &res.ReservationId, &res.Price, &res.Date)
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

func (rs *ReservationRepo) Aaa(v http.ResponseWriter, req *http.Request) {
	log.Println("aezakmi")
}

func (rs *ReservationRepo) InsertReservationByAcco(resAcco *ReservationByAccommodation) error {
	reservationId, _ := gocql.RandomUUID()
	err := rs.session.Query(
		`INSERT INTO reservations_by_acco (acco_id, reservation_id, price, date) 
		VALUES (?, ?, ?, ?)`,
		resAcco.AccoId, reservationId, resAcco.Price, resAcco.Date).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

// -------Reservation By User-------//
func (rs *ReservationRepo) GetReservationsByUser(user_id string) (ReservationsByUser, error) {
	scanner := rs.session.Query(`SELECT user_id, reservation_id, price int, date date, isDeleted
								FROM reservations_by_user WHERE user_id = ? AND isDeleted = 0`,
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
}
