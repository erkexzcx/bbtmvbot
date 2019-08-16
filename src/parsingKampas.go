package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func parseKampas() {

	url := "https://www.kampas.lt/api/classifieds/search-new?query={%22municipality%22%3A%2258%22%2C%22settlement%22%3A19220%2C%22page%22%3A1%2C%22sort%22%3A%22new%22%2C%22section%22%3A%22bustas-nuomai%22%2C%22type%22%3A%22flat%22}"

	// Get HTML:
	content, err := downloadAsBytes(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Decode json
	var result map[string]interface{}
	json.Unmarshal(content, &result)

	// Iterate through posts
	posts := result["hits"].([]interface{})
	for _, c := range posts {
		mypost := c.(map[string]interface{})

		link := "https://www.kampas.lt/skelbimai/" + fmt.Sprintf("%v", mypost["id"])

		// Skip if post already in DB:
		exists, _ := databasePostExists(post{url: link})
		if exists {
			continue
		}

		p := post{
			url:         link,
			phone:       "", // Impossible - anti bot too good
			description: strings.ReplaceAll(fmt.Sprintf("%v", mypost["description"]), "<br/>", "\n"),
			address:     fmt.Sprintf("%v", mypost["title"]),
			heating:     "", // Impossible
			floor:       interfaceToNumber(mypost["objectfloor"]),
			floorTotal:  interfaceToNumber(mypost["totalfloors"]),
			area:        interfaceToNumber(mypost["objectarea"]),
			price:       interfaceToNumber(mypost["objectprice"]),
			rooms:       interfaceToNumber(mypost["totalrooms"]),
			year:        interfaceToNumber(mypost["yearbuilt"]),
		}

		go p.processPost()
	}

}

func interfaceToNumber(i interface{}) int {
	number, _ := strconv.Atoi(fmt.Sprintf("%v", i))
	return number
}
