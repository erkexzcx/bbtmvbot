package aruodas

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"strings"
	"sync"
)

type Aruodas struct{}

const LINK = "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search"
const WEBSITE = "aruodas.lt"

var inProgress = false
var inProgressMux = sync.Mutex{}

func (obj *Aruodas) Retrieve(db *database.Database) []*website.Post {
	log.Println("Retrieving posts from 'aruodas' website...")

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

	// Open new playwright page
	page, err := website.PlaywrightContext.NewPage()
	if err != nil {
		log.Printf("could not create %s page: %v", WEBSITE, err)
		return posts
	}

	// Ensure page is closed after function ends
	defer func() {
		err = page.Close()
		if err != nil {
			log.Printf("could not close %s page: %v", WEBSITE, err)
		}
	}()

	// Go to url that contains list of posts
	if _, err = page.Goto(LINK); err != nil {
		log.Printf("could not goto %s: %v", WEBSITE, err)
		return posts
	}

	entries, _ := page.Locator("ul.search-result-list-big_thumbs > li.result-item-big-thumb:not([style='display: none'])").All()
	//fmt.Println(entries)
	for _, entry := range entries {
		p := &website.Post{}

		// Get post ID
		upstreamID, err := entry.GetAttribute("data-id")
		if err != nil {
			log.Println("Post ID is not found in 'aruodas' website")
			continue
		}
		p.Link = "https://m.aruodas.lt/" + strings.ReplaceAll(upstreamID, "loadobject", "") // https://m.aruodas.lt/4-919937

		//log.Println(p.Link)

		// If post is already in database - skip it
		if db.InDatabase(p.Link) {
			continue
		}

		// Go to post url
		if _, err = page.Goto(p.Link); err != nil {
			log.Printf("could not goto %s: %v", p.Link, err)
			return posts
		}

		// var tmp string

		p.Phone, _ = page.Locator("a[data-id=\"subtitlePhone1\"][data-type=\"phone\"]").First().TextContent()

		p.Description, _ = page.Locator("#advertInfoContainer > #collapsedTextBlock > #collapsedText").First().TextContent()

		p.Address, _ = page.Locator(".show-advert-container > .advert-info-header > h1").First().TextContent()

		p.Heating = page.Locator("dt:contains(\"Šildymas\")").First().TextContent()

		// // Extract heating:
		// el := postDoc.Find("dt:contains(\"Šildymas\")")
		// if el.Length() != 0 {
		// 	p.Heating = el.Next().Text()
		// }

		// // Extract floor:
		// el = postDoc.Find("dt:contains(\"Aukštas\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	p.Floor, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract Floor number from 'aruodas' post")
		// 		return
		// 	}
		// }

		// // Extract floor total:
		// el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	p.FloorTotal, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract FloorTotal number from 'aruodas' post")
		// 		return
		// 	}
		// }

		// // Extract area:
		// el = postDoc.Find("dt:contains(\"Plotas\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	if strings.Contains(tmp, ",") {
		// 		tmp = strings.Split(tmp, ",")[0]
		// 	} else {
		// 		tmp = strings.Split(tmp, " ")[0]
		// 	}
		// 	p.Area, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract Area number from 'aruodas' post")
		// 		return
		// 	}
		// }

		// // Extract price:
		// el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	tmp = strings.ReplaceAll(tmp, " ", "")
		// 	tmp = strings.ReplaceAll(tmp, "€", "")
		// 	p.Price, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract Price number from 'aruodas' post")
		// 		return
		// 	}
		// }

		// // Extract rooms:
		// el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	p.Rooms, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract Rooms number from 'aruodas' post")
		// 		return
		// 	}
		// }

		// // Extract year:
		// el = postDoc.Find("dt:contains(\"Metai\")")
		// if el.Length() != 0 {
		// 	tmp = el.Next().Text()
		// 	tmp = strings.TrimSpace(tmp)
		// 	if strings.Contains(tmp, " ") {
		// 		tmp = strings.Split(tmp, " ")[0]
		// 	}
		// 	p.Year, err = strconv.Atoi(tmp)
		// 	if err != nil {
		// 		log.Println("failed to extract Year number from 'aruodas' post")
		// 		return
		// 	}
		// }

		p.TrimFields()
		posts = append(posts, p)
	}

	return posts
}

func init() {
	website.Add("aruodas", &Aruodas{})
}
