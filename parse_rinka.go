package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var regexRinkaPrice = regexp.MustCompile(`Kaina: ([\d,]+),\d+ €`)

func parseRinka() {
	// Download page
	doc, err := fetchDocument(parseLinkRinka)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("[id=\"adsBlock\"]").First().Find(".ad").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, exists := s.Find("a[itemprop=\"url\"]").Attr("href")
		if !exists {
			return
		}
		p.Link = upstreamID // https://www.rinka.lt/skelbimas/isnuomojamas-1-kambarys-3-kambariu-bute-id-4811032

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

		// Extract details element
		detailsElement := postDoc.Find("#adFullBlock")

		// ------------------------------------------------------------

		var tmp string

		// Extract phone:
		tmp = postDoc.Find("#phone_val_value").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := detailsElement.Find("dt:contains(\"Mikrorajonas / Gyvenvietė:\")").Next().Text()
		addrStreet := detailsElement.Find("dt:contains(\"Gatvė:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.Address = compileAddress(addrState, addrStreet)

		// Extract heating:
		p.Heating = detailsElement.Find("dt:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmp = detailsElement.Find("dt:contains(\"Kelintame aukšte:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		tmp = detailsElement.Find("dt:contains(\"Pastato aukštų skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		tmp = detailsElement.Find("dt:contains(\"Bendras plotas, m²:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = postDoc.Find("span.price:contains(\"Kaina: \")").Text()
		if tmp != "" {
			arr := regexRinkaPrice.FindStringSubmatch(tmp)
			if len(arr) == 2 {
				p.Price, _ = strconv.Atoi(arr[1])
			} else if strings.Contains(tmp, "Nenurodyta") {
				p.Price = -1 // so it gets ignored
			}
		}

		// Extract rooms:
		tmp = detailsElement.Find("dt:contains(\"Kambarių skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		tmp = detailsElement.Find("dt:contains(\"Statybos metai:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Year, _ = strconv.Atoi(tmp)
		}

		go p.Handle()
	})

}
