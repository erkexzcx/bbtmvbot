package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const urlKampas = "https://www.kampas.lt/api/classifieds/search-new?query={%22municipality%22%3A%2258%22%2C%22settlement%22%3A19220%2C%22page%22%3A1%2C%22sort%22%3A%22new%22%2C%22section%22%3A%22bustas-nuomai%22%2C%22type%22%3A%22flat%22}"

func parseKampas() {

	// Wait few seconds so Telegram bot starts up
	time.Sleep(5 * time.Second)

	// Run 'parseKampas' over and over again:
	defer func() {
		time.Sleep(3 * time.Minute)
		parseKampas()
	}()

	// Get HTML:
	res, err := http.Get(urlKampas)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return
	}
	content, err := ioutil.ReadAll(res.Body)
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

		go processPost(p)
	}

}

func interfaceToNumber(i interface{}) int {
	number, _ := strconv.Atoi(fmt.Sprintf("%v", i))
	return number
}
