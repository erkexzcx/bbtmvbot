package bbtmvbot

import (
	"bbtmvbot/database"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	telebot "gopkg.in/telebot.v3"
)

func TelegramMiddlewareUserInDB(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		db.EnsureUserInDB(c.Chat().ID) // This ensures that user is always in DB
		return next(c)                 // continue execution chain
	}
}

func initTelegramHandlers() {
	tb.Handle("/start", handleCommandInfo)
	tb.Handle("/info", handleCommandInfo)
	tb.Handle("/enable", handleCommandEnable)
	tb.Handle("/disable", handleCommandDisable)
	tb.Handle("/config", handleCommandConfig)
}

func handleCommandInfo(c telebot.Context) error {
	return sendTelegram(c.Chat().ID, "BBTMV - 'Butų Be Tarpininkavimo Mokesčio Vilniuje' is a project intended to help find flats for a rent in Vilnius, Lithuania. All you have to do is to set config using /config command and wait until bot sends you notifications.\n\n**Fun fact** - if you are couple and looking for a flat, then create group chat and add this bot into that group - enable settings and bot will send notifications to the same chat. :)")
}

func handleCommandEnable(c telebot.Context) error {
	user := db.GetUser(c.Chat().ID)
	if user.PriceFrom == 0 && user.PriceTo == 0 && user.RoomsFrom == 0 && user.RoomsTo == 0 && user.YearFrom == 0 {
		return sendTelegram(c.Chat().ID, "You must first use /config command before using /enable or /disable commands!")
	}
	if user.Enabled {
		sendTelegram(c.Chat().ID, "Notifications are already enabled!")
		return nil
	}
	db.SetEnabled(c.Chat().ID, true)
	return sendTelegram(c.Chat().ID, "Notifications enabled!")
}

func handleCommandDisable(c telebot.Context) error {
	user := db.GetUser(c.Chat().ID)
	if user.PriceFrom == 0 && user.PriceTo == 0 && user.RoomsFrom == 0 && user.RoomsTo == 0 && user.YearFrom == 0 {
		return sendTelegram(c.Chat().ID, "You must first use `/config` command before using `/enable` or `/disable` commands!")
	}
	if !user.Enabled {
		return sendTelegram(c.Chat().ID, "Notifications are already disabled!")
	}
	db.SetEnabled(c.Chat().ID, false)
	return sendTelegram(c.Chat().ID, "Notifications disabled!")
}

var reConfigCommand = regexp.MustCompile(`^\/config (\d{1,5}) (\d{1,5}) (\d{1,2}) (\d{1,2}) (\d{4})$`)

const configText = "Use this format:\n\n```\n/config <price_from> <price_to> <rooms_from> <rooms_to> <year_from>\n```\nExample:\n```\n/config 200 330 1 2 2000\n```"

const configErrorText = "Wrong input! " + configText

func handleCommandConfig(c telebot.Context) error {
	msg := strings.ToLower(strings.TrimSpace(c.Message().Text))

	// Remove @<botname> from command if exists
	msg = strings.Split(msg, "@")[0]

	// Check if default
	if msg == "/config" {
		return sendTelegram(c.Chat().ID, configText+"\n\n"+activeSettings(c.Chat().ID))
	}

	if !reConfigCommand.MatchString(msg) {
		return sendTelegram(c.Chat().ID, configErrorText)
	}

	// Extract variables from message (using regex)
	match := reConfigCommand.FindStringSubmatch(msg)
	priceFrom, _ := strconv.Atoi(match[1])
	priceTo, _ := strconv.Atoi(match[2])
	roomsFrom, _ := strconv.Atoi(match[3])
	roomsTo, _ := strconv.Atoi(match[4])
	yearFrom, _ := strconv.Atoi(match[5])

	// Values check
	priceCorrect := priceFrom >= 0 || priceTo <= 100000 && priceTo >= priceFrom
	roomsCorrect := roomsFrom >= 0 || roomsTo <= 100 && roomsTo >= roomsFrom
	yearCorrect := yearFrom <= time.Now().Year()

	if !(priceCorrect && roomsCorrect && yearCorrect) {
		return sendTelegram(c.Chat().ID, configErrorText)
	}

	user := &database.User{
		TelegramID: c.Chat().ID,
		Enabled:    true,
		PriceFrom:  priceFrom,
		PriceTo:    priceTo,
		RoomsFrom:  roomsFrom,
		RoomsTo:    roomsTo,
		YearFrom:   yearFrom,
	}
	db.UpdateUser(user)
	return sendTelegram(c.Chat().ID, "Config updated!\n\n"+activeSettings(c.Chat().ID))
}

const userSettingsTemplate = `*Your active settings:*
» *Notifications:* %s
» *Price:* %d-%d€
» *Rooms:* %d-%d
» *Year from:* %d`

func activeSettings(telegramID int64) string {
	u := db.GetUser(telegramID)

	status := "Disabled"
	if u.Enabled {
		status = "Enabled"
	}

	msg := fmt.Sprintf(
		userSettingsTemplate,
		status,
		u.PriceFrom,
		u.PriceTo,
		u.RoomsFrom,
		u.RoomsTo,
		u.YearFrom,
	)
	return msg
}

var telegramMux sync.Mutex
var elapsedTime time.Duration

func sendTelegram(chatID int64, msg string) error {
	telegramMux.Lock()
	defer telegramMux.Unlock()

	startTime := time.Now()
	_, err := tb.Send(&telebot.Chat{ID: chatID}, msg, &telebot.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	})
	if err != nil {
		return err
	}
	elapsedTime = time.Since(startTime)

	// See https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
	time.Sleep(30*time.Millisecond - elapsedTime)
	return nil
}
