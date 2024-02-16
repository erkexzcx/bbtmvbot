package website

import (
	"bbtmvbot/database"

	"github.com/playwright-community/playwright-go"
)

type Website interface {
	Retrieve(db *database.Database) []*Post
}

var Websites = map[string]Website{}

func Add(name string, w Website) {
	Websites[name] = w
}

var PlaywrightContext playwright.BrowserContext
