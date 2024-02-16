package bbtmvbot

import (
	"bbtmvbot/config"
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"path"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/playwright-community/playwright-go"
	telebot "gopkg.in/telebot.v3"
)

var (
	db *database.Database
	tb *telebot.Bot
)

func Start(c *config.Config) {
	// Open DB
	var err error
	db, err = database.Open(path.Join(c.DataDir, "database.db"))
	if err != nil {
		log.Fatalln(err)
	}

	// Init Telegram bot
	pref := telebot.Settings{
		Token:  c.TelegramApiKey,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	tb, err = telebot.NewBot(pref)
	if err != nil {
		log.Fatalln(err)
	}
	tb.Use(TelegramMiddlewareUserInDB)
	initTelegramHandlers()
	go tb.Start()

	// Init playwright
	launchOpts := playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String("/usr/bin/chromium"),
		Headless:       playwright.Bool(false),
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(launchOpts)
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(c.UserAgent),
	})
	// Make it available globally
	website.PlaywrightContext = context

	// Open and keep single blank page, so it's not closing
	_, err = context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	// Init cron
	location, _ := time.LoadLocation("Europe/Vilnius")
	s := gocron.NewScheduler(location)
	s.Every("10m").Do(refreshWebsites) // Retrieve new posts, send to users
	s.Every("24h").Do(cleanup)         // Cleanup (remove posts that are not seen in the last 30 days)

	// Start cron and block execution
	s.StartBlocking()
}

func refreshWebsites() {
	for title, site := range website.Websites {

		go func(title string, site website.Website) {
			posts := site.Retrieve(db)
			for _, post := range posts {
				go processPost(post)
			}
		}(title, site)

	}
}

func processPost(post *website.Post) {
	if post.IsExcludable() {
		db.AddPost(post.Link)
		return
	}

	insertedPostID := db.AddPost(post.Link)

	telegramIDs := db.GetInterestedTelegramIDs(post.Price, post.Rooms, post.Year)
	for _, telegramID := range telegramIDs {
		sendTelegram(telegramID, post.FormatTelegramMessage(insertedPostID))
	}
}

func cleanup() {
	db.DeleteOldPosts() // Older than 30 days
}
