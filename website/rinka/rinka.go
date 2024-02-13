package rinka

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Rinka struct{}

const LINK = "https://www.rinka.lt/nekilnojamojo-turto-skelbimai/butu-nuoma?filter%5BKainaForAll%5D%5Bmin%5D=&filter%5BKainaForAll%5D%5Bmax%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmin%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmax%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmin%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmax%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmin%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmax%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmin%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmax%5D=&filter%5BNTnuomaaukstas%5D%5Bmin%5D=&filter%5BNTnuomaaukstas%5D%5Bmax%5D=&cities%5B0%5D=2&cities%5B1%5D=3"
const WEBSITE = "rinka.lt"

var rePrice = regexp.MustCompile(`Kaina: ([\d,]+),\d+ €`)

func (obj *Rinka) Retrieve(db *database.Database) []*website.Post {
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

	doc.Find("[id='adsBlock']").First().Find(".ad").Each(func(i int, s *goquery.Selection) {
		p := &website.Post{}

		upstreamID, exists := s.Find("a[itemprop='url']").Attr("href")
		if !exists {
			return
		}
		p.Link = upstreamID // https://www.rinka.lt/skelbimas/isnuomojamas-1-kambarys-3-kambariu-bute-id-4811032

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

		// Extract details element
		detailsElement := postDoc.Find("#adFullBlock")
		var tmp string

		// Extract phone:
		tmp = postDoc.Find("#phone_val_value").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("[itemprop=\"description\"]").Text()

		// Extract address:
		addrState := detailsElement.Find("dt:contains(\"Mikrorajonas / Gyvenvietė:\")").Next().Text()
		addrStreet := detailsElement.Find("dt:contains(\"Gatvė:\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.Address = website.CompileAddress(addrState, addrStreet)

		// Extract heating:
		p.Heating = detailsElement.Find("dt:contains(\"Šildymas:\")").Next().Text()

		// Extract floor:
		tmp = detailsElement.Find("dt:contains(\"Kelintame aukšte:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Floor, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Floor number from 'rinka' post")
				return
			}
		}

		// Extract floor total:
		tmp = detailsElement.Find("dt:contains(\"Pastato aukštų skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract FloorTotal number from 'rinka' post")
				return
			}
		}

		// Extract area:
		tmp = detailsElement.Find("dt:contains(\"Bendras plotas, m²:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'rinka' post")
				return
			}
		}

		// Extract price:
		tmp = postDoc.Find("span.price:contains(\"Kaina: \")").Text()
		if tmp != "" {
			arr := rePrice.FindStringSubmatch(tmp)
			if len(arr) == 2 {
				p.Price, err = strconv.Atoi(arr[1])
				if err != nil {
					log.Println("failed to extract Price number from 'rinka' post")
					return
				}
			} else if strings.Contains(tmp, "Nenurodyta") {
				p.Price = -1 // so it gets ignored
			}
		}

		// Extract rooms:
		tmp = detailsElement.Find("dt:contains(\"Kambarių skaičius:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Rooms, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Rooms number from 'rinka' post")
				return
			}
		}

		// Extract year:
		tmp = detailsElement.Find("dt:contains(\"Statybos metai:\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Year, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Year number from 'rinka' post")
				return
			}
		}

		p.TrimFields()
		posts = append(posts, p)
	})

	return posts
}

func init() {
	website.Add("rinka", &Rinka{})
}
