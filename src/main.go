package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	tb "gopkg.in/tucnak/telebot.v2"
)

var bot *tb.Bot

const helpText = "*Galimos komandos:*\n" +
	"/help - Pagalba\n" +
	"/config - Konfiguruoti pranešimus\n" +
	"/enable - Įjungti pranešimus\n" +
	"/disable - Išjungti pranešimus\n" +
	"/stats - Boto statistika\n\n" +
	"*Informacija:*\n" +
	"Tai yra botas (arba tiesiog _programa_), kuris skenuoja įvairius populiariausius būtų nuomos portalus ir ieško būtų Vilniuje, kuriems (potencialiai) nėra taikomas tarpininkavimo ar kažkoks kitas absurdiškas mokestis. Jeigu kyla klausimų arba pasitaikė pranešimas, kuriame yra tarpininkavimo mokestis - chat grupė https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA"

const errorText = "Įvyko duomenų bazės klaida! Praneškite apie tai chat grupėje https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA"
const configText = "Naudokite tokį formatą:\n\n```\n/config <kaina_nuo> <kaina_iki> <kambariai_nuo> <kambariai_iki> <metai_nuo>\n```\nPavyzdys:\n```\n/config 200 330 1 2 2000\n```"
const configErrorText = "Neteisinga įvestis! " + configText

var validConfig = regexp.MustCompile(`^\/config (\d{1,5}) (\d{1,5}) (\d{1,2}) (\d{1,2}) (\d{4})$`)

func main() {

	// Connect to DB
	databaseConnect()
	defer db.Close()

	// Start Telegrambot API
	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token: readAPIFromFile(), URL: "",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Handle("/help", func(m *tb.Message) {
		_init(m.Sender)
		sendHelpText(m.Sender)
		sendUserInfo(m.Sender)
	})
	bot.Handle("/config", func(m *tb.Message) {
		_init(m.Sender)

		msg := strings.ToLower(strings.TrimSpace(m.Text))

		// Check if default:
		if msg == "/config" {
			bot.Send(m.Sender, configText, tb.ModeMarkdown)
			return
		}

		// Check if input is valid (using regex)
		if !validConfig.MatchString(msg) {
			bot.Send(m.Sender, configErrorText, tb.ModeMarkdown)
			return
		}

		// Extract variables from message (using regex)
		extracted := validConfig.FindStringSubmatch(m.Text)
		priceFrom, _ := strconv.Atoi(extracted[1])
		priceTo, _ := strconv.Atoi(extracted[2])
		roomsFrom, _ := strconv.Atoi(extracted[3])
		roomsTo, _ := strconv.Atoi(extracted[4])
		yearFrom, _ := strconv.Atoi(extracted[5])

		// Check values and logic:
		currentTime := time.Now()
		valuesCheck := priceFrom <= 0 || priceTo <= 0 || roomsFrom <= 0 || roomsTo <= 0 || yearFrom < 1800 || yearFrom > currentTime.Year()
		logicCheck := priceFrom > priceTo || roomsFrom > roomsTo
		if valuesCheck || logicCheck {
			bot.Send(m.Sender, configErrorText, tb.ModeMarkdown)
			return
		}

		// All good, so update in DB:
		if !databaseSetConfig(m.Sender.ID, priceFrom, priceTo, roomsFrom, roomsTo, yearFrom) {
			bot.Send(m.Sender, errorText, tb.ModeMarkdown)
			return
		}

		bot.Send(m.Sender, "Nustatymai atnaujinti ir pranešimai įjungti!")
		sendUserInfo(m.Sender)
	})
	bot.Handle("/enable", func(m *tb.Message) {
		_init(m.Sender)

		if databaseSetEnableForUser(m.Sender.ID, 1) {
			bot.Send(m.Sender, "Pranešimai įjungti! Naudokite komandą /disable kad juos išjungti.", tb.ModeMarkdown)
			sendUserInfo(m.Sender)
		} else {
			bot.Send(m.Sender, errorText)
		}
	})
	bot.Handle("/disable", func(m *tb.Message) {
		_init(m.Sender)

		if databaseSetEnableForUser(m.Sender.ID, 0) {
			bot.Send(m.Sender, "Pranešimai išjungti! Naudokite komandą /enable kad juos įjungti.", tb.ModeMarkdown)
			sendUserInfo(m.Sender)
		} else {
			bot.Send(m.Sender, errorText)
		}
	})
	bot.Handle("/stats", func(m *tb.Message) {
		_init(m.Sender)

		s := databaseGetStatistics()

		msg := "Šiek tiek info iš boto duombazės:\n"
		msg += "» *Naudotojų kiekis:* `" + strconv.Itoa(s.usersCount) + " (iš jų " + strconv.Itoa(s.enabledUsersCount) + " įsijungę pranešimus)`\n"
		msg += "» *Nuscreipinta skelbimų:* `" + strconv.Itoa(s.postsCount) + "`\n"
		msg += "» *Vidutiniai kainų nustatymai:* `Nuo " + strconv.Itoa(s.averagePriceFrom) + "€ iki " + strconv.Itoa(s.averagePriceTo) + "€`\n"
		msg += "» *Vidutiniai kambarių sk. nustatymai:* `Nuo " + strconv.Itoa(s.averageRoomsFrom) + " iki " + strconv.Itoa(s.averageRoomsTo) + "`"

		bot.Send(m.Sender, msg, tb.ModeMarkdown)
	})

	// Start parsers in separate goroutines:
	go parseAruodas()
	go parseSkelbiu()
	go parseDomoplius()
	go parseAlio()
	go parseRinka()
	go parseKampas()
	go parseNuomininkai()

	// Start bot:
	bot.Start()
}

// _init ensures that sender is in DB
func _init(sender *tb.User) {
	if !databaseAddNewUser(sender.ID) {
		bot.Send(sender, errorText)
	}
}

// sendHelpText sends help text to the user
func sendHelpText(sender *tb.User) {
	bot.Send(sender, helpText, tb.ModeMarkdown)
}

// sendUserInfo sends user info (from DB) to the user
func sendUserInfo(sender *tb.User) {

	// Get user data from DB:
	user := databaseGetUser(sender.ID)
	if user == nil {
		bot.Send(sender, errorText)
		return
	}

	status := "Įjungti"
	if user.enabled != 1 {
		status = "Išjungti"
	}

	msg := "Jūsų aktyvūs nustatymai:\n"
	msg += "» *Pranešimai:* `" + status + "`\n"
	msg += "» *Kaina nuo:* `" + strconv.Itoa(user.priceFrom) + "`\n"
	msg += "» *Kaina iki:* `" + strconv.Itoa(user.priceTo) + "`\n"
	msg += "» *Kambarių sk. nuo:* `" + strconv.Itoa(user.roomsFrom) + "`\n"
	msg += "» *Kambarių sk. iki:* `" + strconv.Itoa(user.roomsTo) + "`\n"
	msg += "» *Metai nuo:* `" + strconv.Itoa(user.yearFrom) + "`\n"
	bot.Send(sender, msg, tb.ModeMarkdown)

}

func sendTo(userID int, msg string) {
	bot.Send(&tb.User{ID: userID}, msg, &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	})
}

func readAPIFromFile() string {
	apiBytes, err := ioutil.ReadFile("telegram.conf")
	if err != nil {
		fmt.Println("Unable to read API from file. Ensure that 'telegram.conf' file exists.")
		os.Exit(1) // exit with return code 1
	}
	return strings.TrimSpace(string(apiBytes))
}
