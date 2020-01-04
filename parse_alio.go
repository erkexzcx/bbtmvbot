package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseAlio() {
	// Download page
	doc, err := fetchDocument(parseLinkAlio)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("#main_left_b > #main-content-center > div.result").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, ok := s.Attr("id")
		if !ok {
			return
		}
		p.Link = "https://www.alio.lt/skelbimai/ID" + strings.ReplaceAll(upstreamID, "lv_ad_id_", "") + ".html" // https://www.alio.lt/skelbimai/ID60331923.html

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

		// Extract phone:
		tmp := postDoc.Find("#phone_val_value").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("#adv_description_b > .a_line_val").Text()

		// Extract address:
		el := postDoc.Find(".data_moreinfo_b:contains(\"Adresas\")")
		if el.Length() != 0 {
			p.Address = el.Find(".a_line_val").Text()
		}

		// Extract heating:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.Heating = el.Find(".a_line_val").Text()
		}

		// Extract floor:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto aukštas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Aukštų skaičius pastate\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto plotas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kaina, €\")").First()
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			if strings.Contains(tmp, ".") {
				tmp = strings.Split(tmp, ".")[0]
			}
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kambarių skaičius\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Statybos metai\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			p.Year, _ = strconv.Atoi(tmp)
		}

		go p.Handle()
	})
}
