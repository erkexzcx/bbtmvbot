package aruodas

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Aruodas struct{}

const LINK = "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search"
const WEBSITE = "aruodas.lt"

var inProgress = false
var inProgressMux = sync.Mutex{}

func (obj *Aruodas) Retrieve(db *database.Database) []*website.Post {
	posts := make([]*website.Post, 0)

	// If in progress - simply skip current iteration
	inProgressMux.Lock()
	if inProgress {
		defer inProgressMux.Unlock()
		return posts
	}

	// Mark in progress
	inProgress = true
	inProgressMux.Unlock()

	// Mark not in progress after function ends
	defer func() {
		inProgressMux.Lock()
		inProgress = false
		inProgressMux.Unlock()
	}()

	res, err := website.GetResponse(LINK, WEBSITE)
	if err != nil {
		return posts
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return posts
	}

	doc.Find("ul.search-result-list-v2 > li.result-item-v3:not([style='display: none'])").Each(func(i int, s *goquery.Selection) {
		p := &website.Post{}

		upstreamID, ok := s.Attr("data-id")
		if !ok {
			log.Println("Post ID is not found in 'aruodas' website")
			return
		}
		p.Link = "https://m.aruodas.lt/" + strings.ReplaceAll(upstreamID, "loadObject", "") // https://m.aruodas.lt/4-919937

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
		p.Phone = postDoc.Find("a[data-id=\"subtitlePhone1\"][data-type=\"phone\"]").First().Text()

		// Extract description:
		p.Description = postDoc.Find("#advertInfoContainer > #collapsedTextBlock > #collapsedText").Text()

		// Extract address:
		p.Address = postDoc.Find(".show-advert-container > .advert-info-header > h1").Text()

		// Extract heating:
		el := postDoc.Find("dt:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.Heating = el.Next().Text()
		}

		// Extract floor:
		el = postDoc.Find("dt:contains(\"Aukštas\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.Floor, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Floor number from 'aruodas' post")
				return
			}
		}

		// Extract floor total:
		el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract FloorTotal number from 'aruodas' post")
				return
			}
		}

		// Extract area:
		el = postDoc.Find("dt:contains(\"Plotas\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ",") {
				tmp = strings.Split(tmp, ",")[0]
			} else {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'aruodas' post")
				return
			}
		}

		// Extract price:
		el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Price number from 'aruodas' post")
				return
			}
		}

		// Extract rooms:
		el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Rooms number from 'aruodas' post")
				return
			}
		}

		// Extract year:
		el = postDoc.Find("dt:contains(\"Metai\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, " ") {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Year, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Year number from 'aruodas' post")
				return
			}
		}

		p.TrimFields()
		posts = append(posts, p)
	})

	return posts
}

func init() {
	website.Add("aruodas", &Aruodas{})
}
