package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const CREATE_DB = `
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "users" (
	"telegram_id"	INTEGER NOT NULL UNIQUE,
	"enabled"	INTEGER NOT NULL DEFAULT 0,
	"price_from"	INTEGER NOT NULL DEFAULT 0,
	"price_to"	INTEGER NOT NULL DEFAULT 0,
	"rooms_from"	INTEGER NOT NULL DEFAULT 0,
	"rooms_to"	INTEGER NOT NULL DEFAULT 0,
	"year_from"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("telegram_id")
);
CREATE TABLE IF NOT EXISTS "posts" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	"link"	TEXT NOT NULL UNIQUE,
	"last_seen"	INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS "index_posts_link" ON "posts" (
	"link"
);
COMMIT;
`

type Database struct {
	db *sql.DB
}

func Open(path string) (*Database, error) {
	_, fileErr := os.Stat(path)
	d, err := sql.Open("sqlite3", "file:"+path+"?_mutex=full")
	if os.IsNotExist(fileErr) {
		_, err := d.Exec(CREATE_DB)
		if err != nil {
			panic(err)
		}
	}
	return &Database{d}, err
}

type User struct {
	TelegramID int64
	Enabled    bool
	PriceFrom  int
	PriceTo    int
	RoomsFrom  int
	RoomsTo    int
	YearFrom   int
}

func (d *Database) GetInterestedTelegramIDs(price, rooms, year int) []int64 {
	telegram_IDs := make([]int64, 0)
	query := "SELECT telegram_id FROM users WHERE enabled=1 AND ? >= price_from AND ? <= price_to AND ? >= rooms_from AND ? <= rooms_to AND ? >= year_from"
	rows, err := d.db.Query(query, price, price, rooms, rooms, year)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var telegramID int64
		if err = rows.Scan(&telegramID); err != nil {
			panic(err)
		}
		telegram_IDs = append(telegram_IDs, telegramID)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return telegram_IDs
}

func (d *Database) EnsureUserInDB(telegramID int64) {
	query := "INSERT OR IGNORE INTO users(telegram_id) VALUES(?)"
	_, err := d.db.Exec(query, telegramID)
	if err != nil {
		log.Fatalln(err)
	}
}

func (d *Database) InDatabase(link string) bool {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) AS count FROM posts WHERE link=? LIMIT 1", link).Scan(&count)
	if err != nil {
		log.Fatalln(err)
	}
	if count <= 0 {
		return false
	}
	query := "UPDATE posts SET last_seen=? WHERE link=?"
	_, err = d.db.Exec(query, time.Now().Unix(), link)
	if err != nil {
		panic(err)
	}
	return true
}

func (d *Database) AddPost(link string) int64 {
	query := "INSERT INTO posts(link, last_seen) VALUES(?, ?)"
	res, err := d.db.Exec(query, link, time.Now().Unix())
	if err != nil {
		panic(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return id
}

// Delete posts older than 30 days
func (d *Database) DeleteOldPosts() {
	query := "DELETE FROM posts WHERE last_seen < ?"
	_, err := d.db.Exec(query, time.Now().AddDate(0, 0, -30).Unix())
	if err != nil {
		panic(err)
	}
}

func (d *Database) GetUser(telegramID int64) *User {
	var u User
	query := "SELECT * FROM users WHERE telegram_id=? LIMIT 1"
	err := d.db.QueryRow(query, telegramID).Scan(&u.TelegramID, &u.Enabled, &u.PriceFrom, &u.PriceTo, &u.RoomsFrom, &u.RoomsTo, &u.YearFrom)
	if err != nil {
		panic(err)
	}
	return &u
}

func (d *Database) UpdateUser(user *User) {
	query := "UPDATE users SET enabled=1, price_from=?, price_to=?, rooms_from=?, rooms_to=?, year_from=? WHERE telegram_id=?"
	_, err := d.db.Exec(query, user.PriceFrom, user.PriceTo, user.RoomsFrom, user.RoomsTo, user.YearFrom, user.TelegramID)
	if err != nil {
		panic(err)
	}
}

func (d *Database) Enabled(telegramID int64) bool {
	var enabled int
	query := "SELECT enabled FROM users WHERE telegram_id=? LIMIT 1"
	err := d.db.QueryRow(query, telegramID).Scan(&enabled)
	if err != nil {
		panic(err)
	}
	return enabled == 1
}

func (d *Database) SetEnabled(telegramID int64, enabled bool) {
	enabledVal := 0
	if enabled {
		enabledVal = 1
	}
	query := "UPDATE users SET enabled=? WHERE telegram_id=?"
	_, err := d.db.Exec(query, enabledVal, telegramID)
	if err != nil {
		panic(err)
	}
}
