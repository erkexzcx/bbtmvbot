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

const helpText = `
*Galimos komandos:*
/help - Pagalba
/config - Konfiguruoti pranešimus
/enable - Įjungti pranešimus
/disable - Išjungti pranešimus
/stats - Boto statistika

*Aprašymas:*
Tai yra botas (arba tiesiog _programa_), kuris skenuoja įvairius populiariausius būtų nuomos portalus ir ieško būtų Vilniuje, kuriems (potencialiai) nėra taikomas tarpininkavimo ar kažkoks kitas absurdiškas mokestis. Jeigu kyla klausimų arba pasitaikė pranešimas, kuriame yra tarpininkavimo mokestis - chat grupė https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA
`

const errorText = `Įvyko duomenų bazės klaida! Praneškite apie tai chat grupėje https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA`

const configText = "Naudokite tokį formatą:\n\n```\n/config <kaina_nuo> <kaina_iki> <kambariai_nuo> <kambariai_iki> <metai_nuo>\n```\nPavyzdys:\n```\n/config 200 330 1 2 2000\n```"
const configErrorText = "Neteisinga įvestis! " + configText

var validConfig = regexp.MustCompile(`^\/config (\d{1,5}) (\d{1,5}) (\d{1,2}) (\d{1,2}) (\d{4})$`)

func main() {

	// Connect to DB
	databaseConnect()
	defer db.Close()

	// Setup Telegrambot API
	poller := &tb.LongPoller{Timeout: 15 * time.Second}
	middlewarePoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {

		// We only care about messages
		// TODO: Does this IF statement even needed?
		if upd.Message == nil {
			return false
		}

		// Make sure user is in our database
		_init(upd.Message.Sender)

		// Always accept all updates from Telegram
		return true
	})
	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token: readAPIFromFile(), URL: "",
		Poller: middlewarePoller,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Handle("/help", func(m *tb.Message) {
		sendHelpText(m.Sender)
		sendUserInfo(m.Sender)
	})
	bot.Handle("/config", func(m *tb.Message) {
		updateSettings(m.Sender, m.Text)
	})
	bot.Handle("/enable", func(m *tb.Message) {
		enableNotifications(m.Sender)
	})
	bot.Handle("/disable", func(m *tb.Message) {
		disableNotifications(m.Sender)
	})
	bot.Handle("/stats", func(m *tb.Message) {
		sendStats(m.Sender)
	})

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
			time.Sleep(3 * time.Minute) // Run those functions every 3 minutes
		}
	}()

	// Start bot:
	bot.Start()
}

func updateSettings(sender *tb.User, message string) {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Check if default:
	if msg == "/config" {
		bot.Send(sender, configText, tb.ModeMarkdown)
		return
	}

	// Check if input is valid (using regex)
	if !validConfig.MatchString(msg) {
		bot.Send(sender, configErrorText, tb.ModeMarkdown)
		return
	}

	// Extract variables from message (using regex)
	extracted := validConfig.FindStringSubmatch(msg)
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
		bot.Send(sender, configErrorText, tb.ModeMarkdown)
		return
	}

	// All good, so update in DB:
	if !databaseSetConfig(sender.ID, priceFrom, priceTo, roomsFrom, roomsTo, yearFrom) {
		bot.Send(sender, errorText, tb.ModeMarkdown)
		return
	}

	bot.Send(sender, "Nustatymai atnaujinti ir pranešimai įjungti!")
	sendUserInfo(sender)
}

func enableNotifications(sender *tb.User) {
	if databaseSetEnableForUser(sender.ID, 1) {
		bot.Send(sender, "Pranešimai įjungti! Naudokite komandą /disable kad juos išjungti.", tb.ModeMarkdown)
		sendUserInfo(sender)
	} else {
		bot.Send(sender, errorText)
	}
}

func disableNotifications(sender *tb.User) {
	if databaseSetEnableForUser(sender.ID, 0) {
		bot.Send(sender, "Pranešimai išjungti! Naudokite komandą /enable kad juos įjungti.", tb.ModeMarkdown)
		sendUserInfo(sender)
	} else {
		bot.Send(sender, errorText)
	}
}

func sendStats(sender *tb.User) {
	s := databaseGetStatistics()

	msg := fmt.Sprintf(`
Boto statistinė informacija:
» *Naudotojų kiekis:* %d (iš jų %d įjungę pranešimus)
» *Nuscreipinta skelbimų:* %d
» *Vidutiniai kainų nustatymai:* Nuo %d€ iki %d€
» *Vidutiniai kambarių sk. nustatymai:* Nuo %d iki %d`,
		s.usersCount, s.enabledUsersCount,
		s.postsCount,
		s.averagePriceFrom, s.averagePriceTo,
		s.averageRoomsFrom, s.averageRoomsTo)

	bot.Send(sender, msg, tb.ModeMarkdown)
}

// execute this function on every command/message from user
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

	msg := fmt.Sprintf(`
Jūsų aktyvūs nustatymai:
» *Pranešimai:* %s
» *Kaina:* Nuo %d€ iki %d€
» *Kambarių sk.:* Nuo %d iki %d
» *Metai nuo:* %d`,
		status,
		user.priceFrom, user.priceTo,
		user.roomsFrom, user.roomsTo,
		user.yearFrom)

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
