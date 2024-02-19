package kampas

import (
	"bbtmvbot/database"
	"bbtmvbot/logger"
	"bbtmvbot/website"
	"encoding/json"
	"fmt"
	"strings"
)

type Kampas struct {
	Link   string
	Domain string
}

func init() {
	website.Add(&Kampas{
		Link:   "https://www.kampas.lt/api/classifieds/search-new?query={%22municipality%22%3A%2258%22%2C%22settlement%22%3A19220%2C%22page%22%3A1%2C%22sort%22%3A%22new%22%2C%22section%22%3A%22bustas-nuomai%22%2C%22type%22%3A%22flat%22}",
		Domain: "kampas.lt",
	})
}

func (obj *Kampas) GetDomain() string {
	return obj.Domain
}

type kampasPosts struct {
	Hits []struct {
		ID          int      `json:"id"`
		Title       string   `json:"title"`
		Phone       string   `json:"phone"`
		Objectprice int      `json:"objectprice"`
		Objectarea  int      `json:"objectarea"`
		Totalfloors int      `json:"totalfloors"`
		Totalrooms  int      `json:"totalrooms"`
		Objectfloor int      `json:"objectfloor"`
		Yearbuilt   int      `json:"yearbuilt"`
		Description string   `json:"description"`
		Features    []string `json:"features"`
	} `json:"hits"`
}

func (obj *Kampas) Retrieve(db *database.Database, c chan *website.Post) {
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

	// Get HTML code of the loaded page
	content, err := page.Locator("pre").First().InnerHTML() // Because Chromium (?) surrounds json by HTML tags, so using Selector here...
	if err != nil {
		logger.Logger.Errorw("Could not get content of post", "website", obj.Domain, "error", err)
		return
	}

	var results kampasPosts
	err = json.Unmarshal([]byte(content), &results)
	if err != nil {
		logger.Logger.Errorw("Could not unmarshal JSON", "website", obj.Domain, "error", err)
		return
	}

	for _, v := range results.Hits {
		link := fmt.Sprintf("https://www.kampas.lt/skelbimai/%d", v.ID) // https://www.kampas.lt/skelbimai/504506

		logger.Logger.Debugw(fmt.Sprintf("Processing post %s", link), "website", obj.Domain)

		p := &website.Post{Link: link}

		// If post is already in database - skip it
		if db.InDatabase(p.Link) {
			logger.Logger.Debugw(fmt.Sprintf("Post %s is already in database - skipping", p.Link), "website", obj.Domain)
			continue
		}

		// Extract heating
		for _, feature := range v.Features {
			if strings.HasSuffix(feature, "_heating") {
				p.Heating = strings.ReplaceAll(feature, "_heating", "")
				break
			}
		}
		p.Heating = strings.ReplaceAll(p.Heating, "gas", "dujinis")
		p.Heating = strings.ReplaceAll(p.Heating, "central", "centrinis")
		p.Heating = strings.ReplaceAll(p.Heating, "city", "miesto")
		p.Heating = strings.ReplaceAll(p.Heating, "thermostat", "termostatas")

		p.Address = v.Title
		p.Description = strings.ReplaceAll(v.Description, "<br/>", "\n")
		p.Phone = v.Phone
		p.Floor = v.Objectfloor
		p.FloorTotal = v.Totalfloors
		p.Area = v.Objectarea
		p.Price = v.Objectprice
		p.Rooms = v.Totalrooms
		p.Year = v.Yearbuilt

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
