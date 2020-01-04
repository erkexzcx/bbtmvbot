package main

import (
	"fmt"
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
	sendTo(m.Sender, "Pranešimai įjungti! Naudokite komandą /disable kad juos išjungti.\n\n"+ActiveSettingsText)
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
	sendTo(m.Sender, "Nustatymai atnaujinti ir pranešimai įjungti!\n\n"+ActiveSettingsText)
}

func handleCommandHelp(m *tb.Message) {
	ActiveSettingsText, err := getActiveSettingsText(m.Sender)
	if err != nil {
		sendTo(m.Sender, errorText)
		return
	}
	sendTo(m.Sender, helpText+"\n\n"+ActiveSettingsText)
}
