package aruodas

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Aruodas struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Aruodas{
		Link:   "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search",
		Domain: "aruodas.lt",
	})
}

func (obj *Aruodas) GetDomain() string {
	return obj.Domain
}

func (obj *Aruodas) Retrieve(db *database.Database, c chan *website.Post) {

	// Open new playwright blank page
	page, err := website.PlaywrightContext.NewPage()
	if err != nil {
		log.Printf("Could not create blank page for %s: %v\n", obj.Domain, err)
		return
	}

	// Ensure page is closed after function ends
	defer func() {
		err = page.Close()
		if err != nil {
			log.Printf("Could not close %s page: %v\n", obj.Domain, err)
		}
	}()

	log.Printf("Retrieving posts from %s", obj.Domain)

	// Go to website query URL that contains list of posts
	if _, err = page.Goto(obj.Link); err != nil {
		log.Printf("could not goto %s: %v\n", obj.Link, err)
		return
	}

	// Extract entries
	entries, _ := page.Locator("ul.search-result-list-big_thumbs > li.result-item-big-thumb:not([style='display: none'])").All()
	log.Printf("Found %d entries in %s\n", len(entries), obj.Domain)

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.GetAttribute("data-id")
		if err != nil {
			log.Printf("could not get data-id attribute from %s: %v\n", obj.Domain, err)
			continue
		}
		tmplink := "https://m.aruodas.lt/" + strings.ReplaceAll(upstreamID, "loadobject", "") // https://m.aruodas.lt/4-919937
		links = append(links, tmplink)
		log.Printf("Crafted post link for %s: %s\n", obj.Domain, tmplink)
	}

	for _, link := range links {
		// Create new post
		p := &website.Post{Link: link}

		// If post is already in database - skip it
		if db.InDatabase(p.Link) {
			log.Printf("Post %s is already in database - skipping...\n", p.Link)
			continue
		}

		// Avoid being blocked/ratelimited/detected
		log.Printf("Sleeping for 30 seconds to avoid being blocked by %s\n", obj.Domain)
		time.Sleep(30 * time.Second)

		// Go to post url
		if _, err = page.Goto(p.Link); err != nil {
			log.Printf("could not goto %s: %v\n", p.Link, err)
			continue
		}

		// Get HTML code of the loaded page
		content, err := page.Content()
		if err != nil {
			log.Printf("could not get content of %s: %v\n", p.Link, err)
			continue
		}

		// Create goquery document from HTML code
		postDoc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Printf("could not create goquery document from %s: %v\n", p.Link, err)
			continue
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
				continue
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
				continue
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
				continue
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
				continue
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
				continue
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
				continue
			}
		}

		// Trim fields
		p.TrimFields()

		// Send post to channel
		c <- p
	}
}
