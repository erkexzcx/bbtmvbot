package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
Tai yra botas (scriptas), kuris skenuoja įvairius populiariausius būtų nuomos portalus ir ieško būtų Vilniuje, kuriems (potencialiai) nėra taikomas tarpininkavimo mokestis. Jeigu kyla klausimų arba pasitaikė pranešimas, kuriame yra tarpininkavimo mokestis - chat grupė https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA
`

const errorText = `Įvyko duomenų bazės klaida! Praneškite apie tai chat grupėje https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA`

const configText = "Naudokite tokį formatą:\n\n```\n/config <kaina_nuo> <kaina_iki> <kambariai_nuo> <kambariai_iki> <metai_nuo>\n```\nPavyzdys:\n```\n/config 200 330 1 2 2000\n```"
const configErrorText = "Neteisinga įvestis! " + configText

var validConfig = regexp.MustCompile(`^\/config (\d{1,5}) (\d{1,5}) (\d{1,2}) (\d{1,2}) (\d{4})$`)

// We need to ensure that only one goroutine at a time can access `sendTo` function:
var telegramMux sync.Mutex
var startTime time.Time
var elapsedTime time.Duration

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
		sendTo(sender, configText)
		return
	}

	// Check if input is valid (using regex)
	if !validConfig.MatchString(msg) {
		sendTo(sender, configErrorText)
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
		sendTo(sender, configErrorText)
		return
	}

	// All good, so update in DB:
	if !databaseSetConfig(sender.ID, priceFrom, priceTo, roomsFrom, roomsTo, yearFrom) {
		sendTo(sender, errorText)
		return
	}

	sendTo(sender, "Nustatymai atnaujinti ir pranešimai įjungti!")
	sendUserInfo(sender)
}

func enableNotifications(sender *tb.User) {
	if databaseSetEnableForUser(sender.ID, 1) {
		sendTo(sender, "Pranešimai įjungti! Naudokite komandą /disable kad juos išjungti.")
		sendUserInfo(sender)
	} else {
		sendTo(sender, errorText)
	}
}

func disableNotifications(sender *tb.User) {
	if databaseSetEnableForUser(sender.ID, 0) {
		sendTo(sender, "Pranešimai išjungti! Naudokite komandą /enable kad juos įjungti.")
		sendUserInfo(sender)
	} else {
		sendTo(sender, errorText)
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

	sendTo(sender, msg)
}

// execute this function on every command/message from user
func _init(sender *tb.User) {
	if !databaseAddNewUser(sender.ID) {
		sendTo(sender, errorText)
	}
}

// sendHelpText sends help text to the user
func sendHelpText(sender *tb.User) {
	sendTo(sender, helpText)
}

// sendUserInfo sends user info (from DB) to the user
func sendUserInfo(sender *tb.User) {

	// Get user data from DB:
	user := databaseGetUser(sender.ID)
	if user == nil {
		sendTo(sender, errorText)
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

	sendTo(sender, msg)

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
		fmt.Println("Unable to read API from file. Ensure that 'telegram.conf' file exists.")
		os.Exit(1) // exit with return code 1
	}
	return strings.TrimSpace(string(apiBytes))
}
