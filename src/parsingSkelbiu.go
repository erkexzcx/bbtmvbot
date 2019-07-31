package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const urlSkelbiu = "https://www.skelbiu.lt/skelbimai/?cities=465&category_id=322&cities=465&district=0&cost_min=&cost_max=&status=0&space_min=&space_max=&rooms_min=&rooms_max=&building=0&year_min=&year_max=&floor_min=&floor_max=&floor_type=0&user_type=0&type=1&orderBy=1&import=2&keywords="

func parseSkelbiu() {

	// Wait few seconds so Telegram bot starts up
	time.Sleep(5 * time.Second)

	// Run 'parseSkelbiu' over and over again:
	defer func() {
		time.Sleep(3 * time.Minute)
		parseSkelbiu()
	}()

	// Get HTML:
	res, err := http.Get(urlSkelbiu)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For each post in page:
	doc.Find("#itemsList > ul > li.simpleAds").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Find("a.adsImage[data-item-id]").Attr("data-item-id")
		if !exists {
			return
		}
		link := "https://skelbiu.lt/skelbimai/" + postUpstreamID + ".html" // https://skelbiu.lt/42588321.html

		// Skip if post already in DB:
		exists, err := databasePostExists(post{url: link})
		if exists {
			return
		}

		// Download that URL:
		postRes, err := http.Get(link)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer postRes.Body.Close()
		if postRes.StatusCode != 200 {
			fmt.Printf("status code error: %d %s", postRes.StatusCode, postRes.Status)
			return
		}

		// Load the HTML document of post
		postDoc, err := goquery.NewDocumentFromReader(postRes.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Store variables here.
		var phone, descr, addr, heating, tmpStr string
		var floor, floorTotal, area, price, rooms, year int

		// Extract phone:
		phone = postDoc.Find("div.phone-button > div.primary").Text()
		phone = strings.ReplaceAll(phone, " ", "")

		// Extract description:
		descr = postDoc.Find("div[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := postDoc.Find(".detail > .title:contains(\"Mikrorajonas:\")").Next().Text()
		addrStreet := postDoc.Find(".detail > .title:contains(\"Gatvė:\")").Next().Text()
		addrHouseNum := postDoc.Find(".detail > .title:contains(\"Namo numeris:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addrHouseNum = strings.TrimSpace(addrHouseNum)
		addr = compileAddressWithStreet(addrState, addrStreet, addrHouseNum)

		// Extract heating:
		heating = postDoc.Find(".detail > .title:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmpStr = postDoc.Find(".detail > .title:contains(\"Aukštas:\")").Next().Text()
		floorTotal, _ = strconv.Atoi(tmpStr)

		// Extract floor total:
		tmpStr = postDoc.Find(".detail > .title:contains(\"Aukštų skaičius:\")").Next().Text()
		floor, _ = strconv.Atoi(tmpStr)

		// Extract area:
		tmpStr = postDoc.Find(".detail > .title:contains(\"Plotas, m²:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			if strings.Contains(tmpStr, ",") {
				tmpStr = strings.Split(tmpStr, ",")[0]
			} else {
				tmpStr = strings.Split(tmpStr, " ")[0]
			}
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		tmpStr = postDoc.Find("p.price:contains(\" €\")").Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.ReplaceAll(tmpStr, " ", "")
			tmpStr = strings.ReplaceAll(tmpStr, "€", "")
			price, _ = strconv.Atoi(tmpStr)
		}

		// Extract rooms:
		tmpStr = postDoc.Find(".detail > .title:contains(\"Kamb. sk.:\")").Next().Text()
		rooms, _ = strconv.Atoi(tmpStr)

		// Extract year:
		tmpStr = postDoc.Find(".detail > .title:contains(\"Metai:\")").Next().Text()
		year, _ = strconv.Atoi(tmpStr)

		p := post{
			url:         link,
			phone:       strings.TrimSpace(phone),
			description: strings.TrimSpace(descr),
			address:     strings.TrimSpace(addr),
			heating:     strings.TrimSpace(heating),
			floor:       floor,
			floorTotal:  floorTotal,
			area:        area,
			price:       price,
			rooms:       rooms,
			year:        year,
		}

		go processPost(p)
	})

}
