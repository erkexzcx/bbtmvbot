package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var regexDomopliusExtractFloors = regexp.MustCompile(`(\d+), (\d+) `)

func parseDomoplius() {
	// Download page
	doc, err := fetchDocument(parseLinkDomoplius)
	if err != nil {
		log.Println(err)
		return
	}

	// Iterate posts in webpage
	doc.Find("ul.list > li[id^=\"ann_\"]").Each(func(i int, s *goquery.Selection) {

		p := &Post{}

		upstreamID, ok := s.Attr("id")
		if !ok {
			return
		}
		p.Link = "https://m.domoplius.lt/skelbimai/-" + strings.ReplaceAll(upstreamID, "ann_", "") + ".html" // https://m.domoplius.lt/skelbimai/-5806213.html

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
		tmp, err := postDoc.Find("#phone_button_4").Html()
		if err == nil {
			tmp = domopliusDecodeNumber(tmp)
			p.Phone = strings.ReplaceAll(tmp, " ", "")
		}

		// Extract description:
		p.Description = postDoc.Find("div.container > div.group-comments").Text()

		// Extract address:
		tmp = postDoc.Find(".panel > .container > .container > h1").Text()
		if tmp != "" {
			p.Address = strings.Split(tmp, "nuoma ")[1]
		}

		// Extract heating:
		el := postDoc.Find(".view-field-title:contains(\"Šildymas:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			p.Heating = el.Text()
		}

		// Extract floor and floor total:
		el = postDoc.Find(".view-field-title:contains(\"Aukštas:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = strings.TrimSpace(el.Text())
			arr := regexDomopliusExtractFloors.FindStringSubmatch(tmp)
			p.Floor, _ = strconv.Atoi(tmp) // will be 0 on failure, will be number if success
			if len(arr) == 3 {
				p.Floor, _ = strconv.Atoi(arr[1])
				p.FloorTotal, _ = strconv.Atoi(arr[2])
			}
		}

		// Extract area:
		el = postDoc.Find(".view-field-title:contains(\"Buto plotas (kv. m):\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = el.Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, ".")[0]
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		tmp = postDoc.Find(".field-price > .price-column > .h1").Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find(".view-field-title:contains(\"Kambarių skaičius:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = el.Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find(".view-field-title:contains(\"Statybos metai:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = el.Text()
			tmp = strings.TrimSpace(tmp)
			p.Year, _ = strconv.Atoi(tmp)
		}

		go p.Handle()
	})
}

var regexDomopliusExtractNumberMap = regexp.MustCompile(`(\w+)='([^']+)'`)
var regexDomopliusExtractNumberSeq = regexp.MustCompile(`document\.write\(([\w+]+)\);`)

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
