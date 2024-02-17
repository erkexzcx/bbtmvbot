package domoplius

import (
	"bbtmvbot/database"
	"bbtmvbot/logger"
	"bbtmvbot/website"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Domoplius struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Domoplius{
		Link:   "https://m.domoplius.lt/skelbimai/butai?action_type=3&address_1=461&sell_price_from=&sell_price_to=&qt=",
		Domain: "domoplius.lt",
	})
}

func (obj *Domoplius) GetDomain() string {
	return obj.Domain
}

var reExtractFloors = regexp.MustCompile(`(\d+), (\d+) `)

func (obj *Domoplius) Retrieve(db *database.Database, c chan *website.Post) {
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
	entries, _ := page.Locator("ul.list > li[id^='ann_']").All()
	if len(entries) == 0 {
		logger.Logger.Errorw("Could not find any posts", "website", obj.Domain)
		return
	}

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.GetAttribute("id")
		if err != nil {
			logger.Logger.Errorw("Could not get id attribute", "website", obj.Domain, "error", err)
			continue
		}
		tmplink := "https://m.domoplius.lt/skelbimai/-" + strings.ReplaceAll(upstreamID, "ann_", "") + ".html" // https://m.domoplius.lt/skelbimai/-5806213.html
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

		// Extract phone:
		tmp, err := postDoc.Find(".decode-real-value").Html()
		if err == nil {
			p.Phone = strings.ReplaceAll(tmp, " ", "")
		}

		// Extract description:
		p.Description = postDoc.Find("div.container > div.group-comments").Text()

		// Extract address:
		tmp = postDoc.Find(".panel > .container > h1").Text()
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
