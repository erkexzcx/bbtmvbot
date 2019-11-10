package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// DbUser is here to store data about the user from DB
type DbUser struct {
	id          int
	enabled     int
	priceFrom   int
	priceTo     int
	roomsFrom   int
	roomsTo     int
	yearFrom    int
	showWithFee int
}

type dbStats struct {
	postsCount        int
	usersCount        int
	enabledUsersCount int
	averagePriceFrom  int
	averagePriceTo    int
	averageRoomsFrom  int
	averageRoomsTo    int
	usersWithFee      int
}

var db *sql.DB

func databaseConnect() {
	// Open SQLite connection:
	var err error
	db, err = sql.Open("sqlite3", "file:./database.db?_mutex=full")
	if err != nil {
		fmt.Println(err)
	}
}
