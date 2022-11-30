package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Res struct {
	Col bool
}

type Reservation struct {
	Room   int
	During string
}

func main() {

	db, err := sqlx.Connect("postgres", "postgres://postgres:password@localhost/rsvp?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	var reservations []Reservation
	err = db.Select(&reservations, "select * from reservation;")
	if err != nil {
		log.Fatalln(err)
	}
	reservation := reservations[0]
	fmt.Println(reservation.During)
}
