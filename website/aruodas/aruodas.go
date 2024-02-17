package aruodas

import (
	"bbtmvbot/database"
	"bbtmvbot/logger"
	"bbtmvbot/website"
	"fmt"
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
	logger.Logger.Infow("Retrieve started", "website", obj.Domain)

	// Open new playwright blank page
	page, err := website.PlaywrightContext.NewPage()
	if err != nil {
		logger.Logger.Errorw("Could not create blank Playwright page", "website", obj.Domain, "error", err)
		return
	}

	// Ensure page is closed after function ends
	defer func() {
		err = page.Close()
		if err != nil {
			logger.Logger.Errorw("Could not close Playwright page", "website", obj.Domain, "error", err)
		}
	}()

	// Go to website query URL that contains list of posts
	if _, err = page.Goto(obj.Link); err != nil {
		logger.Logger.Errorw("Could not go to website", "website", obj.Domain, "error", err)
		return
	}

	// Extract entries
	entries, _ := page.Locator("ul.search-result-list-big_thumbs > li.result-item-big-thumb:not([style='display: none'])").All()
	if len(entries) == 0 {
		logger.Logger.Errorw("Could not find any posts", "website", obj.Domain)
		return
	}

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.GetAttribute("data-id")
		if err != nil {
			logger.Logger.Errorw("Could not get data-id attribute", "website", obj.Domain, "error", err)
			continue
		}
		tmplink := "https://m.aruodas.lt/" + strings.ReplaceAll(upstreamID, "loadobject", "") // https://m.aruodas.lt/4-919937
		links = append(links, tmplink)
	}
	logger.Logger.Debugw(fmt.Sprintf("Found %d links to be processed", len(entries)), "website", obj.Domain)

	for _, link := range links {
		logger.Logger.Debugw(fmt.Sprintf("Processing post %s", link), "website", obj.Domain)

		// Create new post
		p := &website.Post{Link: link}

		// If post is already in database - skip it
		if db.InDatabase(p.Link) {
			logger.Logger.Debugw(fmt.Sprintf("Post %s is already in database - skipping", p.Link), "website", obj.Domain)
			continue
		}

		// Avoid being blocked/ratelimited/detected
		logger.Logger.Debugw("Sleeping for 30 seconds to avoid being blocked", "website", obj.Domain)
		time.Sleep(30 * time.Second)

		// Go to post url
		if _, err = page.Goto(p.Link); err != nil {
			logger.Logger.Errorw("Could not go to post", "website", obj.Domain, "error", err)
			continue
		}

		// Get HTML code of the loaded page
		content, err := page.Content()
		if err != nil {
			logger.Logger.Errorw("Could not get content of post", "website", obj.Domain, "error", err)
			continue
		}

		// Create goquery document from HTML code
		// This is done because goquery has much more advanced selectors than playwright
		postDoc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			logger.Logger.Errorw("Could not create goquery document from post", "website", obj.Domain, "error", err)
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
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find("dt:contains(\"Aukštų sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
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
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		el = postDoc.Find("dt:contains(\"Kaina mėn.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find("dt:contains(\"Kambarių sk.\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find("dt:contains(\"Metai\")")
		if el.Length() != 0 {
			tmp = el.Next().Text()
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, " ") {
				tmp = strings.Split(tmp, " ")[0]
			}
			p.Year, _ = strconv.Atoi(tmp)
		}

		// Trim fields
		p.TrimFields()

		logger.Logger.Infow(
			"Post processed",
			"website", obj.Domain,
			"link", p.Link,
			"phone", p.Phone,
			"description_length", len(p.Description),
			"address", p.Address,
			"heating", p.Heating,
			"floor", p.Floor,
			"floor_total", p.FloorTotal,
			"area", p.Area,
			"price", p.Price,
			"rooms", p.Rooms,
			"year", p.Year,
		)

		// Send post to channel
		c <- p
	}
}
