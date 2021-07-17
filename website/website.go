package website

import "bbtmvbot/database"

type Website interface {
	Retrieve(db *database.Database) ([]*Post, error)
}

var Websites = map[string]Website{}

func Add(name string, w Website) {
	Websites[name] = w
}

type Post struct {
	Link        string
	Phone       string
	Description string
	Address     string
	Heating     string
	Floor       int
	FloorTotal  int
	Area        int
	Price       int
	Rooms       int
	Year        int
}
