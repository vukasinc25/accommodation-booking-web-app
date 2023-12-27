package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
)

type ReservationRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

func New(logger *log.Logger) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")
	log.Println(db)
	log.Println("A sto ne radi")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	cluster.Timeout = time.Second * 55
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
	//RESERVATION BY ACCO
	err := rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
						(reservation_id UUID, acco_id text, host_id text, numberPeople int, priceByPeople int, priceByAcoommodation int,
						startDate date, endDate date, isDeleted boolean,
						PRIMARY KEY ((reservation_id, acco_id), startDate))
						WITH CLUSTERING ORDER BY (startDate DESC)`,
			"reservations_by_acco")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}

	//RESERVATION BY GUEST
	err = rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(user_id text, reservation_id text, acco_id text, price int, 
						begin_reservation_date date, numberOfPeople int, end_reservation_date date, isDeleted boolean,
					PRIMARY KEY ((user_id, reservation_id, acco_id, begin_reservation_date, end_reservation_date), price))
					WITH CLUSTERING ORDER BY (price ASC)`,
			"reservations_by_user")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}

	//FIND RESERVATION DATES FOR ACCO
	err = rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(id UUID, accommodation_id text, begin_reservation_date date, end_reservation_date date,
					PRIMARY KEY (accommodation_id, id))`,
			"reservations_dates_by_acco_id")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}

	//SEARCH - START AND END DATE
	err = rs.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
					(id UUID, accommodation_id text, begin_reservation_date date, end_reservation_date date, 
					PRIMARY KEY ((begin_reservation_date, end_reservation_date), id))`,
			"reservations_dates_by_date")).Exec()
	if err != nil {
		rs.logger.Println(err)
	}
}

// -------Reservation By Accommodation-------//
func (rs *ReservationRepo) GetReservationsByAcco(acco_id string) (ReservationsByAccommodation, error) {
	scanner := rs.session.Query(`SELECT * FROM reservations_by_acco WHERE acco_id = ? AND isDeleted = false ALLOW FILTERING;`,
		acco_id).Iter().Scanner() // lista
	var reservations ReservationsByAccommodation
	for scanner.Next() {
		var res ReservationByAccommodation
		err := scanner.Scan(&res.ReservationId, &res.AccoId, &res.StartDate, &res.EndDate, &res.HostId, &res.IsDeleted, &res.NumberPeople, &res.PriceByAccommodation, &res.PriceByPeople)
		if err != nil {
			rs.logger.Println("Cant 1", err)
			return nil, err
		}
		reservations = append(reservations, &res)
	}
	if err := scanner.Err(); err != nil {
		rs.logger.Println("Cant 2", err)
		return nil, err
	}
	log.Println(reservations)
	return reservations, nil
}

func (rs *ReservationRepo) InsertReservationByAcco(resAcco *ReservationByAccommodation) error {
	reservationId, _ := gocql.RandomUUID()
	err := rs.session.Query(
		`INSERT INTO reservations_by_acco (reservation_id, acco_id, host_id, numberPeople, priceByPeople, priceByAcoommodation,
			startDate, endDate, isDeleted) VALUES 
		(?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		reservationId, resAcco.AccoId, resAcco.HostId, resAcco.NumberPeople, resAcco.PriceByPeople, resAcco.PriceByAccommodation,
		resAcco.StartDate, resAcco.EndDate, false).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

// RESERVATION DATE FOR ACCO
func (rs *ReservationRepo) GetReservationsDatesByAccomodationId(acco_id string) (ReservationDatesByAccomodationId, error) {
	scanner := rs.session.Query(`SELECT begin_reservation_date, end_reservation_date
    FROM reservations_by_user
    WHERE acco_id = ? and isDeleted = false ALLOW FILTERING`,
		acco_id).Iter().Scanner()

	var dates ReservationDatesByAccomodationId
	for scanner.Next() {
		var res ReservationDate
		err := scanner.Scan(&res.BeginAccomodationDate, &res.EndAccomodationDate)
		if err != nil {
			rs.logger.Println(err)
			return nil, err
		}
		dates = append(dates, &res)
	}
	if err := scanner.Err(); err != nil {
		rs.logger.Println(err)
		return nil, err
	}
	return dates, nil
}

func (rs *ReservationRepo) InsertReservationDateForAccomodation(resDate *ReservationDateByDate) error { // -----------------------
	log.Println("Usli u Insert")

	overlap, err := rs.CheckOverlap(resDate.AccoId, resDate.BeginAccomodationDate, resDate.EndAccomodationDate)
	if err != nil {
		return err
	}

	if overlap {
		return errors.New("overlap detected: Cannot insert overlapping date range")
	}

	id, _ := gocql.RandomUUID()

	err = rs.session.Query(
		`INSERT INTO reservations_dates_by_accomodation_id (id, accommodation_id, begin_reservation_date, end_reservation_date) 
		VALUES (?, ?, ?, ?);`,
		id, resDate.AccoId, resDate.BeginAccomodationDate, resDate.EndAccomodationDate).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	log.Println("Insert prosao")
	return nil
}

// SEARCH - RESERVATION DATES BY START AND END DATE
func (rs *ReservationRepo) GetReservationsDatesByDate(begin_reservation_date string, end_reservation_date string) (ReservationDatesByDate, error) {
	scanner := rs.session.Query(`SELECT accommodation_id, begin_reservation_date, end_reservation_date
    FROM reservations_dates_by_date
    WHERE begin_reservation_date = ? AND end_reservation_date = ?`,
		begin_reservation_date, end_reservation_date).Iter().Scanner()

	var dates ReservationDatesByDate
	for scanner.Next() {
		var res ReservationDateByDate
		err := scanner.Scan(&res.AccoId, &res.BeginAccomodationDate, &res.EndAccomodationDate)
		if err != nil {
			rs.logger.Println(err)
			return nil, err
		}
		dates = append(dates, &res)
	}
	if err := scanner.Err(); err != nil {
		rs.logger.Println(err)
		return nil, err
	}
	return dates, nil
}

func (rs *ReservationRepo) InsertReservationDateByDate(resDate *ReservationDateByDate) error {
	reservationId, _ := gocql.RandomUUID()
	err := rs.session.Query(
		`INSERT INTO reservations_dates_by_date (id, accommodation_id, begin_reservation_date, end_reservation_date) 
		VALUES (?, ?, ?, ?);`,
		reservationId, resDate.AccoId, resDate.BeginAccomodationDate, resDate.EndAccomodationDate).Exec()
	if err != nil {
		rs.logger.Println(err)
		return err
	}
	return nil
}

// -------Reservation By User-------//
func (rs *ReservationRepo) GetReservationsByUser(user_id string) (ReservationsByUser, error) {
	scanner := rs.session.Query(`SELECT reservation_id, acco_id, price, 
	begin_reservation_date, numberOfPeople, end_reservation_date
	FROM reservations_by_user WHERE user_id = ? AND isDeleted = false ALLOW FILTERING;`,
		user_id).Iter().Scanner()

	var reservations ReservationsByUser
	for scanner.Next() {
		var res UserReservations
		err := scanner.Scan(&res.ReservationId, &res.AccoId, &res.Price,
			&res.StartDate, &res.NumberOfPeople, &res.EndDate)
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
	log.Println("Usli u metodu")
	overlap, err := rs.CheckOverlap(resUser.AccoId, resUser.StartDate, resUser.EndDate)
	if err != nil {
		return err
	}

	if overlap {
		return errors.New("Dates are already reserved for that accommodation")
	}

	err = rs.session.Query(
		`INSERT INTO reservations_by_user (user_id, reservation_id, acco_id, price, 
			begin_reservation_date, numberOfPeople, end_reservation_date, isDeleted) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		resUser.UserId, resUser.ReservationId, resUser.AccoId, 100,
		resUser.StartDate, 2, resUser.EndDate, false).Exec()
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

func (rs *ReservationRepo) CheckTable(user_id string, reservation_id string, acco_id string, start_date time.Time, end_date time.Time) (bool, error) {
	log.Println("Start: ", start_date.Format("2006-01-02"))
	log.Println("End: ", end_date.Format("2006-01-02"))

	var count int
	err := rs.session.Query(
		`SELECT COUNT(*) FROM reservation.reservations_by_user
		WHERE user_id = ? 
		AND reservation_id = ?
		AND acco_id = ?
		AND begin_reservation_date = ?
		AND end_reservation_date = ? ALLOW FILTERING`,
		user_id, reservation_id, acco_id, start_date, end_date).Scan(&count)

	if err != nil {
		rs.logger.Println(err)
		return false, err
	}

	log.Println("Count: ", count)

	return count > 0, nil
}

func (rs *ReservationRepo) CheckOverlap(accommodationID string, beginDate, endDate time.Time) (bool, error) {
	var count int
	err := rs.session.Query(
		`SELECT COUNT(*) FROM reservations_by_user
         WHERE acco_id = ? 
         AND begin_reservation_date <= ? AND end_reservation_date >= ? AND isDeleted = false ALLOW FILTERING`,
		accommodationID, endDate, beginDate).Scan(&count)

	if err != nil {
		rs.logger.Println(err)
		return false, err
	}

	log.Println("Count: ", count)

	return count > 0, nil
}

func (rs *ReservationRepo) UpdateReservationByUser(reservationByUser *ReservationByUser) error {
	overlap, err := rs.CheckTable(reservationByUser.UserId, reservationByUser.ReservationId, reservationByUser.AccoId, reservationByUser.StartDate, reservationByUser.EndDate)
	if err != nil {
		return err
	}

	log.Println("Overlap:", overlap)

	if !overlap {
		return errors.New("Cant find reservation")
	}

	passed, err := isDatePassed(reservationByUser.StartDate)
	if err != nil {
		log.Println("Error:", err)
		return err
	}

	if passed {
		return errors.New("Reservation cant be canceled")
	}

	// treba da se postavi validacija da li je datum prosao ako jeste onda ne moze da otkaze rezervaciju

	err = rs.session.Query(
		`UPDATE reservations_by_user SET isDeleted = true where user_id = ? and reservation_id = ? and acco_id = ? and begin_reservation_date = ? and end_reservation_date = ? and price = ?`,
		reservationByUser.UserId, reservationByUser.ReservationId, reservationByUser.AccoId, reservationByUser.StartDate, reservationByUser.EndDate, reservationByUser.Price).Exec()

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

func isDatePassed(dateStr time.Time) (bool, error) {
	currentDate := time.Now()
	return dateStr.Before(currentDate), nil
}
