package skelbiu

import (
	"bbtmvbot/database"
	"bbtmvbot/logger"
	"bbtmvbot/website"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Skelbiu struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Skelbiu{
		Link:   "https://www.skelbiu.lt/skelbimai/?cities=465&category_id=322&cities=465&district=0&cost_min=&cost_max=&status=0&space_min=&space_max=&rooms_min=&rooms_max=&building=0&year_min=&year_max=&floor_min=&floor_max=&floor_type=0&user_type=0&type=1&orderBy=1&import=2&keywords=",
		Domain: "skelbiu.lt",
	})
}

func (obj *Skelbiu) GetDomain() string {
	return obj.Domain
}

func (obj *Skelbiu) Retrieve(db *database.Database, c chan *website.Post) {
	logger.Logger.Debugw("Retrieve started", "website", obj.Domain)

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
	entries, _ := page.Locator("div.standard-list-container a.standard-list-item").All()
	if len(entries) == 0 {
		logger.Logger.Errorw("Could not find any posts", "website", obj.Domain)
		return
	}

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.First().GetAttribute("id")
		if err != nil {
			logger.Logger.Errorw("Could not get id attribute", "website", obj.Domain, "error", err)
			continue
		}
		tmplink := "https://skelbiu.lt/skelbimai/" + strings.ReplaceAll(upstreamID, "ads", "") + ".html" // https://skelbiu.lt/42588321.html
		links = append(links, tmplink)
	}
	logger.Logger.Debugw(fmt.Sprintf("Found %d links to be processed", len(entries)), "website", obj.Domain)

	for _, link := range links {
		logger.Logger.Debugw(fmt.Sprintf("Processing post %s", link), "website", obj.Domain)

		// Create new post
		p := &website.Post{Website: obj.Domain, Link: link}

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

		// Reveal phone number
		page.Locator("div.phone-button[onclick]").First().Click()
		page.Locator(".js-number-to-show:visible .phone-button .js-phone-place").First().WaitFor()

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
		p.Phone = postDoc.Find(".js-phone-place").First().Text()

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

		// Send post to channel
		c <- p

	}
}
