package nuomininkai

import (
	"bbtmvbot/database"
	"bbtmvbot/website"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Nuomininkai struct{}

const LINK = "https://nuomininkai.lt/paieska/?propery_type=butu-nuoma&propery_contract_type=&propery_location=461&imic_property_district=&new_quartals=&min_price=&max_price=&min_price_meter=&max_price_meter=&min_area=&max_area=&rooms_from=&rooms_to=&high_from=&high_to=&floor_type=&irengimas=&building_type=&house_year_from=&house_year_to=&zm_skaicius=&lot_size_from=&lot_size_to=&by_date="
const WEBSITE = "nuomininkai.lt"

var inProgress = false
var inProgressMux = sync.Mutex{}

func (obj *Nuomininkai) Retrieve(db *database.Database) []*website.Post {
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

	doc.Find("div.property-listing > ul > li.property_element").Each(func(i int, s *goquery.Selection) {
		p := &website.Post{}

		upstreamID, exists := s.Find("h3 > a").Attr("href")
		if !exists {
			log.Println("unable to find 'id' of the post in 'nuomininkai' portal")
			return
		}
		p.Link = upstreamID // https://nuomininkai.lt/skelbimas/vilniaus-m-sav-vilniaus-m-pilaite-i-kanto-al-isnuomojamas-1-kambario-butas-pilaiteje/

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
		el := postDoc.Find("h4 > i.fa-mobile").Parent()
		el.Find("i").Remove()
		p.Phone = strings.ReplaceAll(el.Text(), " ", "")

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

		p.TrimFields()
		posts = append(posts, p)
	})

	return posts
}

func init() {
	website.Add("nuomininkai", &Nuomininkai{})
}
