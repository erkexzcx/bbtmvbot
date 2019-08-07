package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var regexDomopliusExtractNumberMap = regexp.MustCompile(`(\w+)='([^']+)'`)
var regexDomopliusExtractNumberSeq = regexp.MustCompile(`document\.write\(([\w+]+)\);`)
var regexDomopliusExtractFloors = regexp.MustCompile(`(\d+), (\d+) `)

func parseDomoplius() {

	url := "https://m.domoplius.lt/skelbimai/butai?action_type=3&address_1=461&sell_price_from=&sell_price_to=&qt="

	// Get content as Goquery Document:
	doc, err := downloadAsGoqueryDocument(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For each post in page:
	doc.Find("ul.list > li[id^=\"ann_\"]").Each(func(i int, s *goquery.Selection) {

		// Get postURL:
		postUpstreamID, exists := s.Attr("id")
		if !exists {
			return
		}
		link := "https://m.domoplius.lt/skelbimai/-" + strings.ReplaceAll(postUpstreamID, "ann_", "") + ".html" // https://m.domoplius.lt/skelbimai/-5806213.html

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
		phone, err = postDoc.Find("#phone_button_4").Html()
		if err == nil {
			phone = domopliusDecodeNumber(phone)
			phone = strings.ReplaceAll(phone, " ", "")
		}

		// Extract description:
		descr = postDoc.Find("div.container > div.group-comments").Text()

		// Extract address:
		addr = postDoc.Find(".panel > .container > .container > h1").Text()
		if addr != "" {
			addr = strings.Split(addr, "nuoma ")[1]
		}

		// Extract heating:
		el := postDoc.Find(".view-field-title:contains(\"Šildymas:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			heating = el.Text()
		}

		// Extract floor and floor total:
		el = postDoc.Find(".view-field-title:contains(\"Aukštas:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp := strings.TrimSpace(el.Text())
			arr := regexDomopliusExtractFloors.FindStringSubmatch(tmp)
			floor, _ = strconv.Atoi(tmp) // will be 0 on failure, will be number if success
			if len(arr) == 3 {
				floor, _ = strconv.Atoi(arr[1])
				floorTotal, _ = strconv.Atoi(arr[2])
			}
		}

		// Extract area:
		el = postDoc.Find(".view-field-title:contains(\"Buto plotas (kv. m):\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmpStr = el.Text()
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.Split(tmpStr, ".")[0]
			area, _ = strconv.Atoi(tmpStr)
		}

		// Extract price:
		tmpStr = postDoc.Find(".field-price > .price-column > .h1").Text()
		if tmpStr != "" {
			tmpStr = strings.TrimSpace(tmpStr)
			tmpStr = strings.ReplaceAll(tmpStr, " ", "")
			tmpStr = strings.ReplaceAll(tmpStr, "€", "")
			price, _ = strconv.Atoi(tmpStr)
		}

		// Extract rooms:
		el = postDoc.Find(".view-field-title:contains(\"Kambarių skaičius:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmpStr = el.Text()
			tmpStr = strings.TrimSpace(tmpStr)
			rooms, _ = strconv.Atoi(tmpStr)
		}

		// Extract year:
		el = postDoc.Find(".view-field-title:contains(\"Statybos metai:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmpStr = el.Text()
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

		go processPost(p)
	})

}

func domopliusDecodeNumber(str string) string {
	// Create map:
	arr := regexDomopliusExtractNumberMap.FindAllSubmatch([]byte(str), -1)
	mymap := make(map[string]string, len(arr))
	for _, v := range arr {
		mymap[string(v[1])] = string(v[2])
	}

	// Create sequence:
	arr = regexDomopliusExtractNumberSeq.FindAllSubmatch([]byte(str), -1)
	var seq string
	for _, v := range arr {
		seq += "+" + string(v[1])
	}
	seq = strings.TrimLeft(seq, "+")

	// Split sequence into array:
	splittedSeq := strings.Split(seq, "+")

	// Build final string:
	var msg string
	for _, v := range splittedSeq {
		msg += mymap[v]
	}

	return strings.ReplaceAll(msg, " ", "")
}
