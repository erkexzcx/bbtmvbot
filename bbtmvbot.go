package bbtmvbot

import (
	"bbtmvbot/config"
	"bbtmvbot/database"
	"bbtmvbot/logger"
	"bbtmvbot/website"
	"log"
	"path"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	telebot "gopkg.in/telebot.v3"
)

var (
	db *database.Database
	tb *telebot.Bot
)

func Start(c *config.Config) {
	// Init logger
	logFilePath := path.Join(c.DataDir, "bbtmvbot.log")
	logger.InitLogger(logFilePath, c.LogLevel)

	// Open DB
	var err error
	db, err = database.Open(path.Join(c.DataDir, "database.db"))
	if err != nil {
		log.Fatalln("Could not open database:", err)
	}

	// Init Telegram bot
	pref := telebot.Settings{
		Token:  c.TelegramApiKey,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	tb, err = telebot.NewBot(pref)
	if err != nil {
		log.Fatalln("Could not create Telegram bot:", err)
	}
	tb.Use(TelegramMiddlewareUserInDB)
	initTelegramHandlers()
	go tb.Start()

	// Init playwright
	launchOpts := playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String("/usr/bin/chromium"),
		Headless:       playwright.Bool(true),
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v\n", err)
	}
	browser, err := pw.Chromium.Launch(launchOpts)
	if err != nil {
		log.Fatalf("could not launch browser: %v\n", err)
	}
	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(c.UserAgent),
	})
	// Make it available globally
	website.PlaywrightContext = context

	// Open and keep single blank page, so it's not closing
	_, err = context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v\n", err)
	}

	// Start websites fetching
	go refreshWebsites()
	go cleanup()

	// Block current routine indefinitely
	select {}
}

func refreshWebsites() {
	postChan := make(chan *website.Post, len(website.Websites))
	wg := sync.WaitGroup{}

	// Accept incoming posts and process them
	go func() {
		for post := range postChan {
			processPost(post)
		}
	}()

	// Every 10 minutes fetch posts from all websites
	for {
		for _, site := range website.Websites {
			wg.Add(1)
			go func(site website.Website) {
				site.Retrieve(db, postChan)
				wg.Done()
			}(site)
		}
		wg.Wait()
		time.Sleep(10 * time.Minute)
	}
}

func processPost(post *website.Post) {
	// Process fields (trim, remove whitespaces)
	post.ProcessFields()

	// Set fee bool value
	post.DetectFee()

	// Detect critical issues with the post - these must exist
	var postErrors []string
	if len(post.Phone) == 0 {
		postErrors = append(postErrors, "empty phone")
	}
	if len(post.Description) == 0 {
		postErrors = append(postErrors, "empty description")
	}
	if len(post.Address) == 0 {
		postErrors = append(postErrors, "empty address")
	}
	if post.Price == 0 {
		postErrors = append(postErrors, "zero price")
	}
	if post.Rooms == 0 {
		postErrors = append(postErrors, "zero rooms")
	}
	if post.Year == 0 {
		postErrors = append(postErrors, "zero year")
	}

	// Detect less critical issues with the post - these should exist, but not necesarrily (e.g., not provided in post)
	var postWarnings []string
	if len(post.Address) == 0 {
		postWarnings = append(postWarnings, "empty address")
	}
	if len(post.Heating) == 0 {
		postWarnings = append(postWarnings, "empty heating")
	}
	if post.Floor == 0 {
		postWarnings = append(postWarnings, "empty floor")
	}
	if post.FloorTotal == 0 {
		postWarnings = append(postWarnings, "empty floorTotal")
	}
	if post.Area == 0 {
		postWarnings = append(postWarnings, "empty area")
	}

	// Print post to logger
	logger.Logger.Infow(
		"Post processed",
		"website", post.Website,
		"link", post.Link,
		"phone", post.Phone,
		"description_length", len(post.Description),
		"address", post.Address,
		"heating", post.Heating,
		"floor", post.Floor,
		"floor_total", post.FloorTotal,
		"area", post.Area,
		"price", post.Price,
		"rooms", post.Rooms,
		"year", post.Year,
		"fee", post.Fee,
		"post_errors", postErrors,
		"post_warnings", postWarnings,
	)

	// Always add to database, so it's not being opened more than once
	insertedPostID := db.AddPost(post.Link)

	// Do not send to Telegram if has fee or has below 3 not properly fetched fields
	if post.Fee || len(postErrors) > 0 {
		return
	}

	// Send to Telegram
	telegramIDs := db.GetInterestedTelegramIDs(post.Price, post.Rooms, post.Year)
	for _, telegramID := range telegramIDs {
		sendTelegram(telegramID, post.FormatTelegramMessage(insertedPostID))
	}
}

func cleanup() {
	db.DeleteOldPosts() // Older than 30 days
	time.Sleep(24 * time.Hour)
	go cleanup()
}
