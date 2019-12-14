package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseAruodas() {

	url := "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search"

	// Get content as Goquery Document:
	doc, err := getGoqueryDocument(url)
	if err != nil {
		log.Println(err)
		return
	}

	// For each post in page:
	doc.Find("ul.search-result-list-v2 > li.result-item-v3:not([style='display: none'])").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Attr("data-id")
		if !exists {
			return
		}
		link := "https://m.aruodas.lt/" + strings.ReplaceAll(postUpstreamID, "loadObject", "") // https://m.aruodas.lt/4-919937/

		// Skip if post already in DB:
		exists, err := postURLInDB(link)
		if err != nil {
			log.Println(err)
			return
		}
		if exists {
			return
		}

		// Get post's content as Goquery Document:
		postDoc, err := getGoqueryDocument(link)
		if err != nil {
			log.Println(err)
			return
		}

		// ------------------------------------------------------------
		p := post{url: link}
		var tmp string

		// Extract phone:
		p.phone = postDoc.Find("a[data-id=\"subtitlePhone1\"][data-type=\"phone\"]").First().Text()

		// Extract description:
		p.description = postDoc.Find("#advertInfoContainer > #collapsedTextBlock > #collapsedText").Text()

		// Extract address:
		p.address = postDoc.Find(".show-advert-container > .advert-info-header > h1").Text()

		// Extract heating:
		el := postDoc.Find("dt:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.heating = el.Next().Text()
		}

		// Extract floor:
		el = postDoc.Find("dt:contains(\"Aukštas\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.floorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		el = postDoc.Find("dt:contains(\"Plotas\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ",") {
				tmp = strings.Split(tmp, ",")[0]
			} else {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find("dt:contains(\"Metai\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, " ") {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.year, _ = strconv.Atoi(tmp)
		}

		go p.processPost()
	})

}
