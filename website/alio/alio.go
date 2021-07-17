package alio

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
)

type Alio struct{}

const LINK = "https://www.alio.lt/paieska/?category_id=1393&city_id=228626&search_block=1&search[eq][adresas_1]=228626&order=ad_id"

func (obj *Alio) Retrieve(db *database.Database) ([]*website.Post, error) {
	return nil, nil
}

func init() {
	website.Add("alio", &Alio{})
}
