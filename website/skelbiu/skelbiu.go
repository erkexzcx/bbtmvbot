package skelbiu

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Skelbiu struct{}

const LINK = "https://www.skelbiu.lt/skelbimai/?cities=465&category_id=322&cities=465&district=0&cost_min=&cost_max=&status=0&space_min=&space_max=&rooms_min=&rooms_max=&building=0&year_min=&year_max=&floor_min=&floor_max=&floor_type=0&user_type=0&type=1&orderBy=1&import=2&keywords="
const WEBSITE = "skelbiu.lt"

func (obj *Skelbiu) Retrieve(db *database.Database) []*website.Post {
	posts := make([]*website.Post, 0)

	res, err := website.GetResponse(LINK, WEBSITE)
	if err != nil {
		return posts
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return posts
	}

	doc.Find("#itemsList > ul > li.simpleAds:not(.passivatedItem)").Each(func(i int, s *goquery.Selection) {
		p := &website.Post{}

		upstreamID, exists := s.Find("a.adsImage[data-item-id]").Attr("data-item-id")
		if !exists {
			return
		}
		p.Link = "https://skelbiu.lt/skelbimai/" + upstreamID + ".html" // https://skelbiu.lt/42588321.html

		if db.InDatabase(p.Link) {
			return
		}

		postRes, err := website.GetResponse(p.Link, WEBSITE)
		if err != nil {
			return
		}
		defer postRes.Body.Close()
		postDoc, err := goquery.NewDocumentFromReader(postRes.Body)
		if err != nil {
			return
		}

		var tmp string

		// Extract phone:
		tmp = postDoc.Find("div.phone-button > div.primary").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("div[itemprop='description']").Text()

		// Extract address:
		addrState := postDoc.Find(".detail > .title:contains('Mikrorajonas:')").Next().Text()
		addrStreet := postDoc.Find(".detail > .title:contains('Gatvė:')").Next().Text()
		addrHouseNum := postDoc.Find(".detail > .title:contains('Namo numeris:')").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		addrHouseNum = strings.TrimSpace(addrHouseNum)
		p.Address = website.CompileAddressWithStreet(addrState, addrStreet, addrHouseNum)

		// Extract heating:
		p.Heating = postDoc.Find(".detail > .title:contains('Šildymas:')").Next().Text()

		// Extract floor:
		tmp = postDoc.Find(".detail > .title:contains('Aukštas:')").Next().Text()
		p.Floor, err = strconv.Atoi(tmp)
		if err != nil {
			log.Println("failed to extract Floor number from 'skelbiu' post")
			return
		}

		// Extract floor total:
		tmp = postDoc.Find(".detail > .title:contains('Aukštų skaičius:')").Next().Text()
		p.FloorTotal, err = strconv.Atoi(tmp)
		if err != nil {
			log.Println("failed to extract FloorTotal number from 'skelbiu' post")
			return
		}

		// Extract area:
		tmp = postDoc.Find(".detail > .title:contains('Plotas, m²:')").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ",") {
				tmp = strings.Split(tmp, ",")[0]
			} else {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'skelbiu' post")
				return
			}
		}

		// Extract price:
		tmp = postDoc.Find("p.price:contains(' €')").Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Price number from 'skelbiu' post")
				return
			}
		}

		// Extract rooms:
		tmp = postDoc.Find(".detail > .title:contains('Kamb. sk.:')").Next().Text()
		p.Rooms, err = strconv.Atoi(tmp)
		if err != nil {
			log.Println("failed to extract Rooms number from 'skelbiu' post")
			return
		}

		// Extract year:
		tmp = postDoc.Find(".detail > .title:contains('Metai:')").Next().Text()
		p.Year, err = strconv.Atoi(tmp)
		if err != nil {
			log.Println("failed to extract Year number from 'skelbiu' post")
			return
		}

		p.TrimFields()
		posts = append(posts, p)
	})

	return posts
}

func init() {
	website.Add("skelbiu", &Skelbiu{})
}
