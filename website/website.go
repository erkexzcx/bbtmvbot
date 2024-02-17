package website

import (
	"bbtmvbot/database"

	"github.com/playwright-community/playwright-go"
)

type Website interface {
	Retrieve(db *database.Database, c chan *Post)
	GetDomain() string
}

var Websites = map[string]Website{}

func Add(w Website) {
	Websites[w.GetDomain()] = w
}

var PlaywrightContext playwright.BrowserContext
