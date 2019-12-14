package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseNuomininkai() {

	url := "https://nuomininkai.lt/paieska/?propery_type=butu-nuoma&propery_contract_type=&propery_location=461&imic_property_district=&new_quartals=&min_price=&max_price=&min_price_meter=&max_price_meter=&min_area=&max_area=&rooms_from=&rooms_to=&high_from=&high_to=&floor_type=&irengimas=&building_type=&house_year_from=&house_year_to=&zm_skaicius=&lot_size_from=&lot_size_to=&by_date="

	// Get content as Goquery Document:
	doc, err := getGoqueryDocument(url)
	if err != nil {
		log.Println(err)
		return
	}

	// For each post in page:
	doc.Find("#property_grid_holder > .property_element").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Find("h3 > a").Attr("href")
		if !exists {
			return
		}
		link := postUpstreamID // https://nuomininkai.lt/skelbimas/vilniaus-m-sav-vilniaus-m-pilaite-i-kanto-al-isnuomojamas-1-kambario-butas-pilaiteje/

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
		el := postDoc.Find("h4 > i.fa-mobile").Parent()
		el.Find("i").Remove()
		p.phone = strings.ReplaceAll(el.Text(), " ", "")

		// Extract description:
		// Extracts together with details table, but we dont care since
		// we dont store description anyway...
		p.description = postDoc.Find("#description").Text()

		// Extract address:
		detailsElement := postDoc.Find("#description > table.table-details")
		addrState := detailsElement.Find("td.table-details-name:contains(\"Mikrorajonas\")").Next().Text()
		addrStreet := detailsElement.Find("td.table-details-name:contains(\"Adresas\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.address = compileAddress(addrState, addrStreet)

		// Extract heating:
		// Not possible

		// Extract floor:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštų sk.\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.floorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Plotas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ".") {
				tmp = strings.Split(tmp, ".")[0]
			}
			p.area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kaina\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kambarių skaičius\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Metai\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.year, _ = strconv.Atoi(tmp)
		}

		go p.processPost()
	})

}
