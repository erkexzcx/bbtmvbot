package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseSkelbiu() {
	// Download page
	doc, err := fetchDocument(parseLinkSkelbiu)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("#itemsList > ul > li.simpleAds:not(.passivatedItem)").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, exists := s.Find("a.adsImage[data-item-id]").Attr("data-item-id")
		if !exists {
			return
		}
		p.Link = "https://skelbiu.lt/skelbimai/" + upstreamID + ".html" // https://skelbiu.lt/42588321.html

		// Skip if already in database:
		if p.InDatabase() {
			return
		}

		// Get post's content as Goquery Document:
		postDoc, err := fetchDocument(p.Link)
		if err != nil {
			log.Println(err)
			return
		}

		// ------------------------------------------------------------

		var tmp string

		// Extract phone:
		tmp = postDoc.Find("div.phone-button > div.primary").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("div[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := postDoc.Find(".detail > .title:contains(\"Mikrorajonas:\")").Next().Text()
		addrStreet := postDoc.Find(".detail > .title:contains(\"Gatvė:\")").Next().Text()
		addrHouseNum := postDoc.Find(".detail > .title:contains(\"Namo numeris:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addrHouseNum = strings.TrimSpace(addrHouseNum)
		p.Address = compileAddressWithStreet(addrState, addrStreet, addrHouseNum)

		// Extract heating:
		p.Heating = postDoc.Find(".detail > .title:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmp = postDoc.Find(".detail > .title:contains(\"Aukštas:\")").Next().Text()
		p.Floor, _ = strconv.Atoi(tmp)

		// Extract floor total:
		tmp = postDoc.Find(".detail > .title:contains(\"Aukštų skaičius:\")").Next().Text()
		p.FloorTotal, _ = strconv.Atoi(tmp)

		// Extract area:
		tmp = postDoc.Find(".detail > .title:contains(\"Plotas, m²:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ",") {
				tmp = strings.Split(tmp, ",")[0]
			} else {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = postDoc.Find("p.price:contains(\" €\")").Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		tmp = postDoc.Find(".detail > .title:contains(\"Kamb. sk.:\")").Next().Text()
		p.Rooms, _ = strconv.Atoi(tmp)

		// Extract year:
		tmp = postDoc.Find(".detail > .title:contains(\"Metai:\")").Next().Text()
		p.Year, _ = strconv.Atoi(tmp)

		go p.Handle()
	})

}
