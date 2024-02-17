package nuomininkai

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

type Nuomininkai struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Nuomininkai{
		Link:   "https://nuomininkai.lt/paieska/?propery_type=butu-nuoma&propery_contract_type=&propery_location=461&imic_property_district=&new_quartals=&min_price=&max_price=&min_price_meter=&max_price_meter=&min_area=&max_area=&rooms_from=&rooms_to=&high_from=&high_to=&floor_type=&irengimas=&building_type=&house_year_from=&house_year_to=&zm_skaicius=&lot_size_from=&lot_size_to=&by_date=",
		Domain: "nuomininkai.lt",
	})
}

func (obj *Nuomininkai) GetDomain() string {
	return obj.Domain
}

func (obj *Nuomininkai) Retrieve(db *database.Database, c chan *website.Post) {
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
	entries, _ := page.Locator("div.property-listing > ul > li.property_element").All()
	if len(entries) == 0 {
		logger.Logger.Errorw("Could not find any posts", "website", obj.Domain)
		return
	}

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		tmplink, err := entry.Locator("div.property-info > h3 > a").GetAttribute("href")
		if err != nil {
			logger.Logger.Errorw("Could not get link of post", "website", obj.Domain, "error", err)
			continue
		}
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
		el := postDoc.Find("h4 > i.fa-mobile").Parent()
		el.Find("i").Remove()
		p.Phone = el.Text()

		// Extract description:
		// Extracts together with details table, but we dont care since
		// we dont store description anyway...
		p.Description = postDoc.Find("#description").Text()

		// Extract address:
		detailsElement := postDoc.Find("#description > table.table-details")
		addrState := detailsElement.Find("td.table-details-name:contains(\"Mikrorajonas\")").Next().Text()
		addrStreet := detailsElement.Find("td.table-details-name:contains(\"Adresas\")").Next().Text()
		addrState = strings.TrimSpace(addrState)
		addrStreet = strings.TrimSpace(addrStreet)
		p.Address = website.CompileAddress(addrState, addrStreet)

		// Extract heating:
		// Not possible

		// Extract floor:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Floor, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Floor number from 'nuomininkai' post")
				return
			}
		}

		// Extract floor total:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Aukštų sk.\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract FloorTotal number from 'nuomininkai' post")
				return
			}
		}

		// Extract area:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Plotas\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			if strings.Contains(tmp, ".") {
				tmp = strings.Split(tmp, ".")[0]
			}
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'nuomininkai' post")
				return
			}
		}

		// Extract price:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kaina\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			tmp = strings.ReplaceAll(tmp, " ", "")
			tmp = strings.ReplaceAll(tmp, "€", "")
			p.Price, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Price number from 'nuomininkai' post")
				return
			}
		}

		// Extract rooms:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Kambarių skaičius\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Rooms, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Rooms number from 'nuomininkai' post")
				return
			}
		}

		// Extract year:
		tmp = detailsElement.Find("td.table-details-name:contains(\"Metai\")").Next().Text()
		if tmp != "" {
			tmp = strings.TrimSpace(tmp)
			p.Year, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Year number from 'nuomininkai' post")
				return
			}
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
