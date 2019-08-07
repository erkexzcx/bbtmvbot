package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseNuomininkai() {

	url := "https://nuomininkai.lt/paieska/?propery_type=butu-nuoma&propery_contract_type=&propery_location=461&imic_property_district=&new_quartals=&min_price=&max_price=&min_price_meter=&max_price_meter=&min_area=&max_area=&rooms_from=&rooms_to=&high_from=&high_to=&floor_type=&irengimas=&building_type=&house_year_from=&house_year_to=&zm_skaicius=&lot_size_from=&lot_size_to=&by_date="

	// Get content as Goquery Document:
	doc, err := downloadAsGoqueryDocument(url)
	if err != nil {
		fmt.Println(err)
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
		var phone, descr, addr, tmpStr string
		var floor, floorTotal, area, price, rooms, year int

		// Extract phone:
		el := postDoc.Find("h4 > i.fa-mobile").Parent()
		el.Find("i").Remove()
		phone = strings.ReplaceAll(el.Text(), " ", "")

		// Extract description:
		// Extracts together with details table, but we dont care since
		// we dont store description anyway...
		descr = postDoc.Find("#description").Text()

		// Extract address:
		detailsElement := postDoc.Find("#description > table.table-details")
		addrState := detailsElement.Find("td.table-details-name:contains(\"Mikrorajonas\")").Next().Text()
		addrStreet := detailsElement.Find("td.table-details-name:contains(\"Adresas\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addr = compileAddress(addrState, addrStreet)

		// Extract heating:
		// Not possible

		// Extract floor:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Aukštas\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			floor, _ = strconv.Atoi(tmpStr)
		}

		// Extract floor total:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Aukštų sk.\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			floorTotal, _ = strconv.Atoi(tmpStr)
		}

		// Extract area:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Plotas\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			if strings.Contains(tmpStr, ".") {
				tmpStr = strings.Split(tmpStr, ".")[0]
			}
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Kaina\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.ReplaceAll(tmpStr, " ", "")
			tmpStr = strings.ReplaceAll(tmpStr, "€", "")
			price, _ = strconv.Atoi(tmpStr)
		}

		// Extract rooms:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Kambarių skaičius\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			rooms, _ = strconv.Atoi(tmpStr)
		}

		// Extract year:
		tmpStr = detailsElement.Find("td.table-details-name:contains(\"Metai\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			year, _ = strconv.Atoi(tmpStr)
		}

		p := post{
			url:         link,
			phone:       strings.TrimSpace(phone),
			description: strings.TrimSpace(descr),
			address:     strings.TrimSpace(addr),
			heating:     "", // not possible
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
