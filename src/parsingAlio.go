package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseAlio() {

	url := "https://www.alio.lt/paieska/?category_id=1393&city_id=228626&search_block=1&search[eq][adresas_1]=228626&order=ad_id"

	// Get content as Goquery Document:
	doc, err := getGoqueryDocument(url)
	if err != nil {
		log.Println(err)
		return
	}

	// For each post in page:
	doc.Find("#main_left_b > #main-content-center > div.result").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Attr("id")
		if !exists {
			return
		}
		link := "https://www.alio.lt/skelbimai/ID" + strings.ReplaceAll(postUpstreamID, "lv_ad_id_", "") + ".html" // https://www.alio.lt/skelbimai/ID60331923.html

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

		// Extract phone:
		tmp := postDoc.Find("#phone_val_value").Text()
		p.phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.description = postDoc.Find("#adv_description_b > .a_line_val").Text()

		// Extract address:
		el := postDoc.Find(".data_moreinfo_b:contains(\"Adresas\")")
		if el.Length() != 0 {
			p.address = el.Find(".a_line_val").Text()
		}

		// Extract heating:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.heating = el.Find(".a_line_val").Text()
		}

		// Extract floor:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto aukštas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Aukštų skaičius pastate\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.floorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto plotas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			p.area, _ = strconv.Atoi(tmp)
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
			p.price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kambarių skaičius\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Statybos metai\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			p.year, _ = strconv.Atoi(tmp)
		}

		go p.processPost()
	})

}
