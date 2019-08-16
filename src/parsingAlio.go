package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseAlio() {

	url := "https://www.alio.lt/paieska/?category_id=1393&city_id=228626&search_block=1&search[eq][adresas_1]=228626&order=ad_id"

	// Get content as Goquery Document:
	doc, err := downloadAsGoqueryDocument(url)
	if err != nil {
		fmt.Println(err)
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
		exists, err := databasePostExists(post{url: link})
		if exists {
			return
		}

		// Get post's content as Goquery Document:
		postDoc, err := downloadAsGoqueryDocument(link)
		if err != nil {
			fmt.Println(err)
			return
		}

		//-------------------------------------------------
		// Define variables:
		var phone, descr, addr, heating, tmpStr string
		var floor, floorTotal, area, price, rooms, year int

		// Extract phone:
		phone = postDoc.Find("#phone_val_value").Text()
		phone = strings.ReplaceAll(phone, " ", "")

		// Extract description:
		descr = postDoc.Find("#adv_description_b > .a_line_val").Text()

		// Extract address:
		el := postDoc.Find(".data_moreinfo_b:contains(\"Adresas\")")
		if el.Length() != 0 {
			addr = el.Find(".a_line_val").Text()
		}

		// Extract heating:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Šildymas\")")
		if el.Length() != 0 {
			heating = el.Find(".a_line_val").Text()
		}

		// Extract floor:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto aukštas\")")
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			floor, _ = strconv.Atoi(tmpStr)
		}

		// Extract floor total:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Aukštų skaičius pastate\")")
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			floorTotal, _ = strconv.Atoi(tmpStr)
		}

		// Extract area:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto plotas\")")
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.Split(tmpStr, " ")[0]
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kaina, €\")").First()
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.ReplaceAll(tmpStr, " ", "")
			tmpStr = strings.ReplaceAll(tmpStr, "€", "")
			if strings.Contains(tmpStr, ".") {
				tmpStr = strings.Split(tmpStr, ".")[0]
			}
			price, _ = strconv.Atoi(tmpStr)
		}

		// Extract rooms:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kambarių skaičius\")")
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			rooms, _ = strconv.Atoi(tmpStr)
		}

		// Extract year:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Statybos metai\")")
		if el.Length() != 0 {
			tmpStr = el.Find(".a_line_val").Text()
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.Split(tmpStr, " ")[0]
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

		go p.processPost()
	})

}
