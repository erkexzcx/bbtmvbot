package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

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

type Post struct {
	ID       int
	Link     string
	Excluded bool
	Reason   string
	LastSeen string
}

func (d *Database) Users() []*User {
	return nil
}

func (d *Database) EnsureUserInDB(telegramID int64) {
	query := "INSERT OR IGNORE INTO users(telegram_id) VALUES(?)"
	_, err := d.db.Exec(query, telegramID)
	if err != nil {
		log.Fatalln(err)
	}
}

func (d *Database) LinkInDatabase(link string) bool {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) AS count FROM posts WHERE link=? LIMIT 1", link).Scan(&count)
	if err != nil {
		log.Fatalln(err)
	}
	return count > 0
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
