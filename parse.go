package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	parseLinkAlio        = "https://www.alio.lt/paieska/?category_id=1393&city_id=228626&search_block=1&search[eq][adresas_1]=228626&order=ad_id"
	parseLinkAruodas     = "https://m.aruodas.lt/?obj=4&FRegion=461&FDistrict=1&FOrder=AddDate&from_search=1&detailed_search=1&FShowOnly=FOwnerDbId0%2CFOwnerDbId1&act=search"
	parseLinkDomoplius   = "https://m.domoplius.lt/skelbimai/butai?action_type=3&address_1=461&sell_price_from=&sell_price_to=&qt="
	parseLinkKampas      = "https://www.kampas.lt/api/classifieds/search-new?query={%22municipality%22%3A%2258%22%2C%22settlement%22%3A19220%2C%22page%22%3A1%2C%22sort%22%3A%22new%22%2C%22section%22%3A%22bustas-nuomai%22%2C%22type%22%3A%22flat%22}"
	parseLinkNuomininkai = "https://nuomininkai.lt/paieska/?propery_type=butu-nuoma&propery_contract_type=&propery_location=461&imic_property_district=&new_quartals=&min_price=&max_price=&min_price_meter=&max_price_meter=&min_area=&max_area=&rooms_from=&rooms_to=&high_from=&high_to=&floor_type=&irengimas=&building_type=&house_year_from=&house_year_to=&zm_skaicius=&lot_size_from=&lot_size_to=&by_date="
	parseLinkRinka       = "https://www.rinka.lt/nekilnojamojo-turto-skelbimai/butu-nuoma?filter%5BKainaForAll%5D%5Bmin%5D=&filter%5BKainaForAll%5D%5Bmax%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmin%5D=&filter%5BNTnuomakambariuskaiciusButai%5D%5Bmax%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmin%5D=&filter%5BNTnuomabendrasplotas%5D%5Bmax%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmin%5D=&filter%5BNTnuomastatybosmetai%5D%5Bmax%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmin%5D=&filter%5BNTnuomaaukstuskaicius%5D%5Bmax%5D=&filter%5BNTnuomaaukstas%5D%5Bmin%5D=&filter%5BNTnuomaaukstas%5D%5Bmax%5D=&cities%5B0%5D=2&cities%5B1%5D=3"
	parseLinkSkelbiu     = "https://www.skelbiu.lt/skelbimai/?cities=465&category_id=322&cities=465&district=0&cost_min=&cost_max=&status=0&space_min=&space_max=&rooms_min=&rooms_max=&building=0&year_min=&year_max=&floor_min=&floor_max=&floor_type=0&user_type=0&type=1&orderBy=1&import=2&keywords="
)

func compileAddressWithStreet(district, street, houseNumber string) (address string) {
	if district == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + district
	} else if houseNumber == "" {
		address = "Vilnius, " + district + ", " + street
	} else {
		address = "Vilnius, " + district + ", " + street + " " + houseNumber
	}
	return
}

func compileAddress(district, street string) (address string) {
	if district == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + district
	} else {
		address = "Vilnius, " + district + ", " + street
	}
	return
}

var httpClient = &http.Client{Timeout: time.Second * 30}

func fetch(link string) ([]byte, error) {
	res, err := httpClient.Get(link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		u, _ := url.Parse(link)
		return nil, fmt.Errorf("%s returned: %s %s", u.Host, res.Status, string(content))
	}

	return content, nil
}

func fetchDocument(link string) (*goquery.Document, error) {
	res, err := httpClient.Get(link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		u, _ := url.Parse(link)
		return nil, fmt.Errorf("%s returned: %s", u.Host, res.Status)
	}

	return goquery.NewDocumentFromReader(res.Body)
}
