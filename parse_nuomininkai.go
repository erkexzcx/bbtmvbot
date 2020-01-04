package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseNuomininkai() {
	// Download page
	doc, err := fetchDocument(parseLinkNuomininkai)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("#property_grid_holder > .property_element").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, exists := s.Find("h3 > a").Attr("href")
		if !exists {
			return
		}
		p.Link = upstreamID // https://nuomininkai.lt/skelbimas/vilniaus-m-sav-vilniaus-m-pilaite-i-kanto-al-isnuomojamas-1-kambario-butas-pilaiteje/

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
		el := postDoc.Find("h4 > i.fa-mobile").Parent()
		el.Find("i").Remove()
		p.Phone = strings.ReplaceAll(el.Text(), " ", "")

		// Extract description:
		// Extracts together with details table, but we dont care since
		// we dont store description anyway...
		p.Description = postDoc.Find("#description").Text()

		// Extract address:
		detailsElement := postDoc.Find("#description > table.table-details")
		addrState := detailsElement.Find("td.table-details-name:contains(\"Mikrorajonas\")").Next().Text()
		addrStreet := detailsElement.Find("td.table-details-name:contains(\"Adresas\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.Address = compileAddress(addrState, addrStreet)

		// Extract heating:
		// Not possible

		// Extract floor:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštų sk.\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Plotas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ".") {
				tmp = strings.Split(tmp, ".")[0]
			}
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kaina\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kambarių skaičius\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Metai\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Year, _ = strconv.Atoi(tmp)
		}

		go p.Handle()
	})

}
