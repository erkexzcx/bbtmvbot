package domoplius

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Domoplius struct{}

const LINK = "https://m.domoplius.lt/skelbimai/butai?action_type=3&address_1=461&sell_price_from=&sell_price_to=&qt="

var reExtractFloors = regexp.MustCompile(`(\d+), (\d+) `)

func (obj *Domoplius) Retrieve(db *database.Database) []*website.Post {
	posts := make([]*website.Post, 0)

	res, err := website.GetResponse(LINK)
	if err != nil {
		return posts
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return posts
	}

	doc.Find("ul.list > li[id^='ann_']").Each(func(i int, s *goquery.Selection) {
		p := &website.Post{}

		upstreamID, ok := s.Attr("id")
		if !ok {
			log.Println("Post ID is not found in 'domoplius' website")
			return
		}
		p.Link = "https://m.domoplius.lt/skelbimai/-" + strings.ReplaceAll(upstreamID, "ann_", "") + ".html" // https://m.domoplius.lt/skelbimai/-5806213.html

		if db.InDatabase(p.Link) {
			return
		}

		postRes, err := website.GetResponse(p.Link)
		if err != nil {
			return
		}
		defer postRes.Body.Close()
		postDoc, err := goquery.NewDocumentFromReader(postRes.Body)
		if err != nil {
			return
		}

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
			arr := reExtractFloors.FindStringSubmatch(tmp)
			p.Floor, _ = strconv.Atoi(tmp) // will be 0 on failure, will be number if success
			if len(arr) == 3 {
				p.Floor, err = strconv.Atoi(arr[1])
				if err != nil {
					log.Println("failed to extract Floor number from 'domoplius' post")
					return
				}
				p.FloorTotal, err = strconv.Atoi(arr[2])
				if err != nil {
					log.Println("failed to extract FloorTotal number from 'domoplius' post")
					return
				}
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
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'domoplius' post")
				return
			}
		}

		// Extract price:
		tmp = postDoc.Find(".field-price > .price-column > .h1").Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Price number from 'domoplius' post")
				return
			}
		}

		// Extract rooms:
		el = postDoc.Find(".view-field-title:contains(\"Kambarių skaičius:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = el.Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Rooms number from 'domoplius' post")
				return
			}
		}

		// Extract year:
		el = postDoc.Find(".view-field-title:contains(\"Statybos metai:\")")
		if el.Length() != 0 {
			el = el.Parent()
			el.Find("span").Remove()
			tmp = el.Text()
			tmp = strings.TrimSpace(tmp)
			p.Year, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Year number from 'domoplius' post")
				return
			}
		}

		p.TrimFields()
		posts = append(posts, p)
	})

	return posts
}

var reNumberMap = regexp.MustCompile(`(\w+)='([^']+)'`)
var reNumerSeq = regexp.MustCompile(`document\.write\(([\w+]+)\);`)

func domopliusDecodeNumber(str string) string {
	// Create map:
	arr := reNumberMap.FindAllStringSubmatch(str, -1)
	mymap := make(map[string]string, len(arr))
	for _, v := range arr {
		mymap[v[1]] = v[2]
	}

	// Create sequence:
	arr = reNumerSeq.FindAllStringSubmatch(str, -1)
	var seq string
	for _, v := range arr {
		seq += "+" + v[1]
	}
	seq = strings.TrimLeft(seq, "+")

	// Split sequence into array:
	splittedSeq := strings.Split(seq, "+")

	// Build final string:
	var msg string
	for _, v := range splittedSeq {
		msg += mymap[v]
	}

	// Remove spaces
	msg = strings.ReplaceAll(msg, " ", "")

	// Replace 00 in the beginning to +
	if strings.HasPrefix(msg, "00") {
		msg = strings.Replace(msg, "00", "+", 1)
	}

	// Replace 86 in the beginning to +3706
	if strings.HasPrefix(msg, "86") {
		msg = strings.Replace(msg, "86", "+3706", 1)
	}

	return msg
}

func init() {
	website.Add("domoplius", &Domoplius{})
}
