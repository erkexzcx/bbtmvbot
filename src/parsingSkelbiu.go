package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseSkelbiu() {

	url := "https://www.skelbiu.lt/skelbimai/?cities=465&category_id=322&cities=465&district=0&cost_min=&cost_max=&status=0&space_min=&space_max=&rooms_min=&rooms_max=&building=0&year_min=&year_max=&floor_min=&floor_max=&floor_type=0&user_type=0&type=1&orderBy=1&import=2&keywords="

	// Get content as Goquery Document:
	doc, err := getGoqueryDocument(url)
	if err != nil {
		log.Println(err)
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
		tmp = postDoc.Find("div.phone-button > div.primary").Text()
		p.phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.description = postDoc.Find("div[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := postDoc.Find(".detail > .title:contains(\"Mikrorajonas:\")").Next().Text()
		addrStreet := postDoc.Find(".detail > .title:contains(\"Gatvė:\")").Next().Text()
		addrHouseNum := postDoc.Find(".detail > .title:contains(\"Namo numeris:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addrHouseNum = strings.TrimSpace(addrHouseNum)
		p.address = compileAddressWithStreet(addrState, addrStreet, addrHouseNum)

		// Extract heating:
		p.heating = postDoc.Find(".detail > .title:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmp = postDoc.Find(".detail > .title:contains(\"Aukštas:\")").Next().Text()
		p.floor, _ = strconv.Atoi(tmp)

		// Extract floor total:
		tmp = postDoc.Find(".detail > .title:contains(\"Aukštų skaičius:\")").Next().Text()
		p.floorTotal, _ = strconv.Atoi(tmp)

		// Extract area:
		tmp = postDoc.Find(".detail > .title:contains(\"Plotas, m²:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ",") {
				tmp = strings.Split(tmp, ",")[0]
			} else {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = postDoc.Find("p.price:contains(\" €\")").Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		tmp = postDoc.Find(".detail > .title:contains(\"Kamb. sk.:\")").Next().Text()
		p.rooms, _ = strconv.Atoi(tmp)

		// Extract year:
		tmp = postDoc.Find(".detail > .title:contains(\"Metai:\")").Next().Text()
		p.year, _ = strconv.Atoi(tmp)

		go p.processPost()
	})

}
