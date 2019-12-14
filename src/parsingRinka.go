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

	url := "https://www.rinka.lt/nekilnojamojo-turto-skelbimai/butu-nuoma?filter%5BKainaForAll%5D%5Bmin%5D=&filter%5BKainaForAll%5D%5Bmax%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmin%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmax%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmin%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmax%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmin%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmax%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmin%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmax%5D=&filter%5BNTnuomaaukstas%5D%5Bmin%5D=&filter%5BNTnuomaaukstas%5D%5Bmax%5D=&cities%5B0%5D=2&cities%5B1%5D=3"

	// Get content as Goquery Document:
	doc, err := getGoqueryDocument(url)
	if err != nil {
		log.Println(err)
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

		// Extract details element
		detailsElement := postDoc.Find("#adFullBlock")

		// ------------------------------------------------------------
		p := post{url: link}
		var tmp string

		// Extract phone:
		tmp = postDoc.Find("#phone_val_value").Text()
		p.phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.description = postDoc.Find("[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := detailsElement.Find("dt:contains(\"Mikrorajonas / Gyvenvietė:\")").Next().Text()
		addrStreet := detailsElement.Find("dt:contains(\"Gatvė:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.address = compileAddress(addrState, addrStreet)

		// Extract heating:
		p.heating = detailsElement.Find("dt:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmp = detailsElement.Find("dt:contains(\"Kelintame aukšte:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		tmp = detailsElement.Find("dt:contains(\"Pastato aukštų skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.floorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		tmp = detailsElement.Find("dt:contains(\"Bendras plotas, m²:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = postDoc.Find("span.price:contains(\"Kaina: \")").Text()
		if tmp != "" {
			arr := regexRinkaPrice.FindStringSubmatch(tmp)
			if len(arr) == 2 {
				p.price, _ = strconv.Atoi(arr[1])
			} else if strings.Contains(tmp, "Nenurodyta") {
				p.price = -1 // so it gets ignored
			}
		}

		// Extract rooms:
		tmp = detailsElement.Find("dt:contains(\"Kambarių skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		tmp = detailsElement.Find("dt:contains(\"Statybos metai:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.year, _ = strconv.Atoi(tmp)
		}

		go p.processPost()
	})

}
