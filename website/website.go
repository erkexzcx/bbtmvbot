package website

import (
	"bbtmvbot/database"
)

type Website interface {
	Retrieve(db *database.Database) []*Post
}

var Websites = map[string]Website{}

func Add(name string, w Website) {
	Websites[name] = w
}
