package alio

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

type Alio struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Alio{
		Link:   "https://www.alio.lt/paieska/?category_id=1393&city_id=228626&search_block=1&search[eq][adresas_1]=228626&order=ad_id",
		Domain: "alio.lt",
	})
}

func (obj *Alio) GetDomain() string {
	return obj.Domain
}

func (obj *Alio) Retrieve(db *database.Database, c chan *website.Post) {
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
	entries, _ := page.Locator("#main_left_b > #main-content-center > div.result").All()
	if len(entries) == 0 {
		logger.Logger.Errorw("Could not find any posts", "website", obj.Domain)
		return
	}

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.GetAttribute("id")
		if err != nil {
			logger.Logger.Errorw("Could not get data-id attribute", "website", obj.Domain, "error", err)
			continue
		}
		tmplink := "https://www.alio.lt/skelbimai/ID" + strings.ReplaceAll(upstreamID, "lv_ad_id_", "") + ".html" // https://www.alio.lt/skelbimai/ID60331923.html
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
		tmp := postDoc.Find("#phone_val_value").Text()
		p.Phone = strings.ReplaceAll(tmp, " ", "")

		// Extract description:
		p.Description = postDoc.Find("#adv_description_b > .a_line_val").Text()

		// Extract address:
		el := postDoc.Find(".data_moreinfo_b:contains(\"Adresas\")")
		if el.Length() != 0 {
			p.Address = el.Find(".a_line_val").Text()
		}

		// Extract heating:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Šildymas\")")
		if el.Length() != 0 {
			p.Heating = el.Find(".a_line_val").Text()
		}

		// Extract floor:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto aukštas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.Floor, _ = strconv.Atoi(tmp)
		}

		// Extract floor total:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Aukštų skaičius pastate\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, _ = strconv.Atoi(tmp)
		}

		// Extract area:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto plotas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			tmp = strings.Split(tmp, ".")[0]
			p.Area, _ = strconv.Atoi(tmp)
		}

		// Extract price:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kaina, €\")").First()
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			if strings.Contains(tmp, ".") {
				tmp = strings.Split(tmp, ".")[0]
			}
			p.Price, _ = strconv.Atoi(tmp)
		}

		// Extract rooms:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kambarių skaičius\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, _ = strconv.Atoi(tmp)
		}

		// Extract year:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Statybos metai\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
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
