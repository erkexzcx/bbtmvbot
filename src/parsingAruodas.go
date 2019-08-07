package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const urlAruodas = "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search"

func parseAruodas() {

	// Run 'parseAruodas' over and over again:
	defer func() {
		time.Sleep(3 * time.Minute)
		parseAruodas()
	}()

	// Get HTML:
	res, err := http.Get(urlAruodas)
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
	doc.Find("ul.search-result-list-v2 > li.result-item-v3").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Attr("data-id")
		if !exists {
			return
		}
		link := "https://m.aruodas.lt/" + strings.ReplaceAll(postUpstreamID, "loadObject", "") // https://m.aruodas.lt/4-919937/

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
		phone = postDoc.Find("a[data-id=\"subtitlePhone1\"][data-type=\"phone\"]").First().Text()

		// Extract description:
		descr = postDoc.Find("#advertInfoContainer > #collapsedTextBlock > #collapsedText").Text()

		// Extract address:
		addr = postDoc.Find(".show-advert-container > .advert-info-header > h1").Text()

		// Extract heating:
		el := postDoc.Find("dt:contains(\"Šildymas\")")
		if el.Length() != 0 {
			heating = el.Next().Text()
		}

		// Extract floor:
		el = postDoc.Find("dt:contains(\"Aukštas\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			floor, _ = strconv.Atoi(tmpStr)
		}

		// Extract floor total:
		el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			floorTotal, _ = strconv.Atoi(tmpStr)
		}

		// Extract area:
		el = postDoc.Find("dt:contains(\"Plotas\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			if strings.Contains(tmpStr, ",") {
				tmpStr = strings.Split(tmpStr, ",")[0]
			} else {
				tmpStr = strings.Split(tmpStr, " ")[0]
			}
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.ReplaceAll(tmpStr, " ", "")
			tmpStr = strings.ReplaceAll(tmpStr, "€", "")
			price, _ = strconv.Atoi(tmpStr)
		}

		// Extract rooms:
		el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			rooms, _ = strconv.Atoi(tmpStr)
		}

		// Extract year:
		el = postDoc.Find("dt:contains(\"Metai\")")
		if el.Length() != 0 {
			tmpStr = el.Next().Text()
			tmpStr = strings.TrimSpace(tmpStr)
			if strings.Contains(tmpStr, " ") {
				tmpStr = strings.Split(tmpStr, " ")[0]
			}
			year, _ = strconv.Atoi(tmpStr)
		}

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
