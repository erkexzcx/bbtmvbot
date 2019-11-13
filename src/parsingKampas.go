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
		exists, _ := post{url: link}.postExistsInDB()
		if exists {
			continue
		}

		phone := ""
		heating := ""
		// Get content as Goquery Document:
		doc, err := downloadAsGoqueryDocument(link)
		if err == nil {
			attr, exists := doc.Find("div.sidebar span.hidden.hidden-phone > a.btn").First().Attr("href")
			if exists {
				phone = strings.ReplaceAll(strings.ReplaceAll(attr, "tel:", ""), " ", "")
			}

			heating = doc.Find("i.i-heating+span").Text()
		}

		p := post{
			url:         link,
			phone:       phone,
			description: strings.ReplaceAll(fmt.Sprintf("%v", mypost["description"]), "<br/>", "\n"),
			address:     fmt.Sprintf("%v", mypost["title"]),
			heating:     heating,
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
