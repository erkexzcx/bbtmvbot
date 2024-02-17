package alio

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
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
	entries, _ := page.Locator("#main_left_b > #main-content-center > div.result").All()
	log.Printf("Found %d entries in %s\n", len(entries), obj.Domain)

	// Extract all the links now, so we can re-use browser page later
	links := []string{}
	for _, entry := range entries {
		upstreamID, err := entry.GetAttribute("id")
		if err != nil {
			log.Printf("could not get id attribute from %s: %v\n", obj.Domain, err)
			continue
		}
		tmplink := "https://www.alio.lt/skelbimai/ID" + strings.ReplaceAll(upstreamID, "lv_ad_id_", "") + ".html" // https://www.alio.lt/skelbimai/ID60331923.html
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
			p.Floor, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Floor number from 'alio' post")
				return
			}
		}

		// Extract floor total:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Aukštų skaičius pastate\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.FloorTotal, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract FloorTotal number from 'alio' post")
				return
			}
		}

		// Extract area:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Buto plotas\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			tmp = strings.Split(tmp, ".")[0]
			p.Area, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Area number from 'alio' post")
				return
			}
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
			p.Price, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Price number from 'alio' post")
				return
			}
		}

		// Extract rooms:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Kambarių skaičius\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			p.Rooms, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Rooms number from 'alio' post")
				return
			}
		}

		// Extract year:
		el = postDoc.Find(".data_moreinfo_b:contains(\"Statybos metai\")")
		if el.Length() != 0 {
			tmp = el.Find(".a_line_val").Text()
			tmp = strings.TrimSpace(tmp)
			tmp = strings.Split(tmp, " ")[0]
			p.Year, err = strconv.Atoi(tmp)
			if err != nil {
				log.Println("failed to extract Year number from 'alio' post")
				return
			}
		}

		p.TrimFields()

		c <- p
	}

}
