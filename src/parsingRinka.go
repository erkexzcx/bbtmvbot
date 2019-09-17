package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var regexRinkaPrice = regexp.MustCompile(`Kaina: ([\d,]+),\d+ €`)

func parseRinka() {

	url := "https://www.rinka.lt/nekilnojamojo-turto-skelbimai/butu-nuoma?filter%5BKainaForAll%5D%5Bmin%5D=&filter%5BKainaForAll%5D%5Bmax%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmin%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmax%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmin%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmax%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmin%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmax%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmin%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmax%5D=&filter%5BNTnuomaaukstas%5D%5Bmin%5D=&filter%5BNTnuomaaukstas%5D%5Bmax%5D=&cities%5B0%5D=2&cities%5B1%5D=3"

	// Get content as Goquery Document:
	doc, err := downloadAsGoqueryDocument(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For each post in page:
	doc.Find("[id=\"adsBlock\"]").First().Find(".ad").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Find("a[itemprop=\"url\"]").Attr("href")
		if !exists {
			return
		}
		link := postUpstreamID // https://www.rinka.lt/skelbimas/isnuomojamas-1-kambarys-3-kambariu-bute-id-4811032

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

		// Extract details element
		detailsElement := postDoc.Find("#adFullBlock")

		//-------------------------------------------------
		// Define variables:
		var phone, descr, addr, heating, tmpStr string
		var floor, floorTotal, area, price, rooms, year int

		// Extract phone:
		phone = postDoc.Find("#phone_val_value").Text()
		phone = strings.ReplaceAll(phone, " ", "")

		// Extract description:
		descr = postDoc.Find("[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := detailsElement.Find("dt:contains(\"Mikrorajonas / Gyvenvietė:\")").Next().Text()
		addrStreet := detailsElement.Find("dt:contains(\"Gatvė:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addr = compileAddress(addrState, addrStreet)

		// Extract heating:
		heating = detailsElement.Find("dt:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmpStr = detailsElement.Find("dt:contains(\"Kelintame aukšte:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			floor, _ = strconv.Atoi(tmpStr)
		}

		// Extract floor total:
		tmpStr = detailsElement.Find("dt:contains(\"Pastato aukštų skaičius:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			floorTotal, _ = strconv.Atoi(tmpStr)
		}

		// Extract area:
		tmpStr = detailsElement.Find("dt:contains(\"Bendras plotas, m²:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		tmpStr = postDoc.Find("span.price:contains(\"Kaina: \")").Text()
		if tmpStr != "" {
			arr := regexRinkaPrice.FindStringSubmatch(tmpStr)
			if len(arr) == 2 {
				price, _ = strconv.Atoi(arr[1])
			} else if strings.Contains(tmpStr, "Nenurodyta") {
				price = -1 // so it gets ignored
			}
		}

		// Extract rooms:
		tmpStr = detailsElement.Find("dt:contains(\"Kambarių skaičius:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			rooms, _ = strconv.Atoi(tmpStr)
		}

		// Extract year:
		tmpStr = detailsElement.Find("dt:contains(\"Statybos metai:\")").Next().Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
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
