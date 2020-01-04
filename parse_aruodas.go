package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseAruodas() {
	// Download page
	doc, err := fetchDocument(parseLinkAruodas)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("ul.search-result-list-v2 > li.result-item-v3:not([style='display: none'])").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, ok := s.Attr("data-id")
		if !ok {
			return
		}
		p.Link = "https://m.aruodas.lt/" + strings.ReplaceAll(upstreamID, "loadObject", "") // https://m.aruodas.lt/4-919937

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
		p.Phone = postDoc.Find("a[data-id=\"subtitlePhone1\"][data-type=\"phone\"]").First().Text()

		// Extract description:
		p.Description = postDoc.Find("#advertInfoContainer > #collapsedTextBlock > #collapsedText").Text()

		// Extract address:
		p.Address = postDoc.Find(".show-advert-container > .advert-info-header > h1").Text()

		// Extract heating:
		el := postDoc.Find("dt:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.Heating = el.Next().Text()
		}

		// Extract floor:
		el = postDoc.Find("dt:contains(\"Aukštas\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
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
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find("dt:contains(\"Metai\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, " ") {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Year, _ = strconv.Atoi(tmp)
		}

		go p.Handle()
	})
}
