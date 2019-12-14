package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	tb "gopkg.in/tucnak/telebot.v2"
)

type user struct {
	id        int
	enabled   int
	priceFrom int
	priceTo   int
	roomsFrom int
	roomsTo   int
	yearFrom  int
}

type stats struct {
	postsCount        int
	usersCount        int
	enabledUsersCount int
	averagePriceFrom  int
	averagePriceTo    int
	averageRoomsFrom  int
	averageRoomsTo    int
}

var db *sql.DB

var bot *tb.Bot

// We need to ensure that only one goroutine at a time can access `sendTo` function:
var telegramMux sync.Mutex
var startTime time.Time
var elapsedTime time.Duration

func main() {

	// Connect to DB
	databaseConnect()
	defer db.Close()

	// Start web server for InfluxDB data
	go func() {
		http.HandleFunc("/influx", handleRequestInflux)
		log.Fatal(http.ListenAndServe(":3999", nil))
	}()

	// Setup Telegrambot API
	poller := &tb.LongPoller{Timeout: 15 * time.Second}
	middlewarePoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {

		// We only care about messages
		// TODO: Does this IF statement even needed?
		if upd.Message == nil {
			return false
		}

		// Make sure user exists in database
		if !ensureUserInDB(upd.Message.Sender.ID) {
			sendTo(upd.Message.Sender, errorText)
		}

		// Always accept all updates from Telegram
		return true
	})
	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token: readAPIFromFile(), URL: "",
		Poller: middlewarePoller,
	})
	if err != nil {
		log.Println(err)
		return
	}

	bot.Handle("/help", handleCommandHelp)
	bot.Handle("/config", handleCommandConfig)
	bot.Handle("/enable", handleCommandEnable)
	bot.Handle("/disable", handleCommandDisable)
	bot.Handle("/stats", handleCommandStats)

	// Start parsers in separate goroutines:
	go func() {
		time.Sleep(5 * time.Second) // Wait few seconds so Telegram bot starts up
		for {
			go parseAruodas()
			go parseSkelbiu()
			go parseDomoplius()
			go parseAlio()
			go parseRinka()
			go parseKampas()
			go parseNuomininkai()
			time.Sleep(3 * time.Minute)
		}
	}()

	// Start bot:
	bot.Start()
}

func databaseConnect() {
	var err error
	db, err = sql.Open("sqlite3", "file:./database.db?_mutex=full")
	if err != nil {
		log.Println(err)
	}
}

func getActiveSettingsText(sender *tb.User) (string, error) {
	// Get user data from DB:
	u, err := getUser(sender.ID)
	if err != nil {
		return "", err
	}

	var status string
	if u.enabled == 1 {
		status = "Įjungti"
	} else {
		status = "Išjungti"
	}

	msg := fmt.Sprintf(activeSettings, status, u.priceFrom,
		u.priceTo, u.roomsFrom, u.roomsTo, u.yearFrom)
	return msg, nil
}

func sendTo(sender *tb.User, msg string) {
	go func() {
		telegramMux.Lock()

		startTime = time.Now()
		bot.Send(sender, msg, &tb.SendOptions{
			ParseMode:             "Markdown",
			DisableWebPagePreview: true,
		})
		elapsedTime = time.Since(startTime)

		// See https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
		if elapsedTime < 30*time.Millisecond {
			time.Sleep(30*time.Millisecond - elapsedTime)
		}

		telegramMux.Unlock()
	}()
}

func readAPIFromFile() string {
	apiBytes, err := ioutil.ReadFile("telegram.conf")
	if err != nil {
		log.Println("Unable to read API from file:", err)
		os.Exit(1) // exit with return code 1
	}
	return strings.TrimSpace(string(apiBytes))
}

func ensureUserInDB(userID int) bool {
	query := "INSERT OR IGNORE INTO users(id) VALUES(?)"
	_, err := db.Exec(query, userID)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func getUser(userID int) (user, error) {
	query := "SELECT * FROM users WHERE id=? LIMIT 1"
	var u user
	err := db.QueryRow(query, userID).Scan(&u.id, &u.enabled,
		&u.priceFrom, &u.priceTo, &u.roomsFrom,
		&u.roomsTo, &u.yearFrom)
	if err != nil {
		log.Println(err)
		return user{}, err
	}
	return u, nil
}

func getStats() (stats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM posts) AS posts_count,
			(SELECT COUNT(*) FROM users) AS users_count,
			(SELECT COUNT(*) FROM users WHERE enabled=1) AS users_enabled_count,
			(SELECT CAST(AVG(price_from) AS INT) FROM users WHERE enabled=1) AS avg_price_from,
			(SELECT CAST(AVG(price_to) AS INT) FROM users WHERE enabled=1) AS avg_price_to,
			(SELECT CAST(AVG(rooms_from) AS INT) FROM users WHERE enabled=1) AS avg_rooms_from,
			(SELECT CAST(AVG(rooms_to) AS INT) FROM users WHERE enabled=1) AS avg_rooms_to
		FROM users LIMIT 1`
	var s stats
	err := db.QueryRow(query).Scan(&s.postsCount, &s.usersCount,
		&s.enabledUsersCount, &s.averagePriceFrom, &s.averagePriceTo,
		&s.averageRoomsFrom, &s.averageRoomsTo)
	if err != nil {
		log.Println(err)
		return stats{}, err
	}
	return s, nil
}
