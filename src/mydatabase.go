package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// DbUser is here to store data about the user from DB
type DbUser struct {
	id        int
	enabled   int
	priceFrom int
	priceTo   int
	roomsFrom int
	roomsTo   int
	yearFrom  int
}

type dbStats struct {
	postsCount        int
	usersCount        int
	enabledUsersCount int
	averagePriceFrom  int
	averagePriceTo    int
	averageRoomsFrom  int
	averageRoomsTo    int
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

// databaseSetEnableForUser - set column "enabled" value either 1 or 0
func databaseSetEnableForUser(userID, value int) bool {

	sql := fmt.Sprintf("UPDATE users SET enabled = %d WHERE id = %d", value, userID)
	_, err := db.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// databaseSetConfig - set config values for user
func databaseSetConfig(userID, priceFrom, priceTo, roomsFrom, roomsTo, yearFrom int) bool {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return false
	}
	sql := "UPDATE users SET enabled=1, price_from=?, price_to=?, rooms_from=?, rooms_to=?, year_from=? WHERE id=?"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(priceFrom, priceTo, roomsFrom, roomsTo, yearFrom, userID)
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// databaseAddNewUser - Adds new user to the table "users"
func databaseAddNewUser(userID int) bool {

	sql := fmt.Sprintf("INSERT OR IGNORE INTO users(id) VALUES(%d)", userID)
	_, err := db.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true

}

// databaseGetUser - Get DbUser from DB
func databaseGetUser(userID int) *DbUser {

	sql := fmt.Sprintf("SELECT * FROM users WHERE id = %d LIMIT 1", userID)

	var user DbUser
	err := db.QueryRow(sql).Scan(
		&user.id,
		&user.enabled,
		&user.priceFrom,
		&user.priceTo,
		&user.roomsFrom,
		&user.roomsTo,
		&user.yearFrom)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &user
}

// databaseAddPost - Add post to database
func databaseAddPost(p post) int64 {

	sql := fmt.Sprintf("INSERT INTO posts(url) values (\"%s\")", p.url)

	res, err := db.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return lastInsertedID

}

func databasePostExists(p post) (bool, error) {
	var count int // Will store count here

	sql := fmt.Sprintf("SELECT COUNT(*) AS count FROM posts WHERE url=\"%s\" LIMIT 1", p.url)
	err := db.QueryRow(sql).Scan(&count)

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if count != 1 {
		return false, nil
	}
	return true, nil

}

func databaseGetStatistics() (stats dbStats) {
	sql := `
SELECT
	(SELECT COUNT(*) FROM posts) AS posts_count,
	(SELECT COUNT(*) FROM users) AS users_count,
	(SELECT COUNT(*) FROM users WHERE enabled=1) AS users_enabled_count,
	(SELECT CAST(AVG(price_from) AS INT) FROM users WHERE enabled=1) AS avg_price_from,
	(SELECT CAST(AVG(price_to) AS INT) FROM users WHERE enabled=1) AS avg_price_to,
	(SELECT CAST(AVG(rooms_from) AS INT) FROM users WHERE enabled=1) AS avg_rooms_from,
	(SELECT CAST(AVG(rooms_to) AS INT) FROM users WHERE enabled=1) AS avg_rooms_to
FROM users LIMIT 1
`

	db.QueryRow(sql).Scan(&stats.postsCount, &stats.usersCount,
		&stats.enabledUsersCount, &stats.averagePriceFrom,
		&stats.averagePriceTo, &stats.averageRoomsFrom,
		&stats.averageRoomsTo)

	return
}

// Exception - this function also send data!!!
func databaseGetUsersAndSendThem(p post, postID int64) {

	sql := fmt.Sprintf(`
	SELECT id FROM users WHERE
	enabled=1 AND
	((price_from <= %d AND price_to >= %d) OR %d = 0) AND
	((rooms_from <= %d AND rooms_to >= %d) OR %d = 0) AND
	(year_from <= %d OR %d = 0)`,
		p.price, p.price, p.price,
		p.rooms, p.rooms, p.rooms,
		p.year, p.year)

	rows, err := db.Query(sql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		err = rows.Scan(&userID)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Send to user:
		sendTo(&tb.User{ID: userID}, p.compileMessage(postID))
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return
	}

}
