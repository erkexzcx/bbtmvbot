package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func handleCommandStats(m *tb.Message) {
	s, err := getStats()
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}

	msg := fmt.Sprintf(statsTemplate, s.usersCount,
		s.enabledUsersCount, s.postsCount, s.averagePriceFrom,
		s.averagePriceTo, s.averageRoomsFrom, s.averageRoomsTo)

	sendTo(m.Sender, msg)
}

func handleCommandEnable(m *tb.Message) {
	query := "UPDATE users SET enabled=1 WHERE id=?"
	_, err := db.Exec(query, m.Sender.ID)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}

	ActiveSettingsText, err := getActiveSettingsText(m.Sender)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}
	sendTo(m.Sender, "Pranešimai įjungti! Naudokite komandą /disable kad juos išjungti.\n"+ActiveSettingsText)
}

func handleCommandDisable(m *tb.Message) {
	query := "UPDATE users SET enabled=0 WHERE id=?"
	_, err := db.Exec(query, m.Sender.ID)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}

	ActiveSettingsText, err := getActiveSettingsText(m.Sender)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}
	sendTo(m.Sender, "Pranešimai išjungti! Naudokite komandą /enable kad juos įjungti.\n\n"+ActiveSettingsText)
}

var validConfig = regexp.MustCompile(`^\/config (\d{1,5}) (\d{1,5}) (\d{1,2}) (\d{1,2}) (\d{4})$`)

func handleCommandConfig(m *tb.Message) {
	msg := strings.ToLower(strings.TrimSpace(m.Text))

	// Check if default:
	if msg == "/config" {
		sendTo(m.Sender, configText)
		return
	}

	// Check if input is valid (using regex)
	if !validConfig.MatchString(msg) {
		sendTo(m.Sender, configErrorText)
		return
	}

	// Extract variables from message (using regex)
	extracted := validConfig.FindStringSubmatch(msg)
	priceFrom, _ := strconv.Atoi(extracted[1])
	priceTo, _ := strconv.Atoi(extracted[2])
	roomsFrom, _ := strconv.Atoi(extracted[3])
	roomsTo, _ := strconv.Atoi(extracted[4])
	yearFrom, _ := strconv.Atoi(extracted[5])

	// Values check
	priceCorrect := priceFrom >= 0 || priceTo <= 100000 && priceTo >= priceFrom
	roomsCorrect := roomsFrom >= 0 || roomsTo <= 100 && roomsTo >= roomsFrom
	yearCorrect := yearFrom <= time.Now().Year()

	if !(priceCorrect && roomsCorrect && yearCorrect) {
		sendTo(m.Sender, configErrorText)
		return
	}

	// Update in DB
	query := "UPDATE users SET enabled=1, price_from=?, price_to=?, rooms_from=?, rooms_to=?, year_from=? WHERE id=?"
	_, err := db.Exec(query, priceFrom, priceTo, roomsFrom, roomsTo, yearFrom, m.Sender.ID)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}

	ActiveSettingsText, err := getActiveSettingsText(m.Sender)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}
	sendTo(m.Sender, "Nustatymai atnaujinti ir pranešimai įjungti!\n"+ActiveSettingsText)
}

func handleCommandHelp(m *tb.Message) {
	ActiveSettingsText, err := getActiveSettingsText(m.Sender)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}
	sendTo(m.Sender, helpText+"\n\n"+ActiveSettingsText)
}

func handleRequestInflux(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT 'portal' AS "type", 'alio.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%alio.lt%"
	UNION SELECT 'portal' AS "type", 'aruodas.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%aruodas.lt%"
	UNION SELECT 'portal' AS "type", 'domoplius.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%domoplius.lt%"
	UNION SELECT 'portal' AS "type", 'kampas.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%kampas.lt%"
	UNION SELECT 'portal' AS "type", 'nuomininkai.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%nuomininkai.lt%"
	UNION SELECT 'portal' AS "type", 'rinka.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%rinka.lt%"
	UNION SELECT 'portal' AS "type", 'skelbiu.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%skelbiu.lt%"
	UNION SELECT 'users' AS "type", 'visited' AS "key", COUNT(*) AS "value" FROM users
	UNION SELECT 'users' AS "type", 'enabled' AS "key", COUNT(*) AS "value" FROM users WHERE enabled = 1
	UNION SELECT 'user_preferences' AS "type", 'avg_price_from' AS "key", (SELECT CAST(AVG(price_from) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_price_to' AS "key", (SELECT CAST(AVG(price_to) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_rooms_from' AS "key", (SELECT CAST(AVG(rooms_from) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_rooms_to' AS "key", (SELECT CAST(AVG(rooms_to) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'posts' AS "type", 'total' AS "key", (SELECT COUNT(*) FROM posts) AS "value"
	UNION SELECT 'posts' AS "type", 'excluded' AS "key", (SELECT COUNT(*) FROM posts WHERE excluded=1) AS "value"
	UNION SELECT 'posts' AS "type", 'sent' AS "key", (SELECT COUNT(*) FROM posts WHERE excluded=0) AS "value"
	UNION SELECT 'posts' AS "type", 'no_price' AS "key", (SELECT COUNT(*) FROM posts WHERE reason="0eur price") AS "value"
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var typ, key, value string
		err = rows.Scan(&typ, &key, &value)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Fprintf(w, "bbtmvbot,type=%s,key=%s value=%s\n", typ, key, value)
	}
	if rows.Err() != nil {
		log.Println(err)
		return
	}
}
