package website

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

// Ensure websites are not accessed to frequently, otherwise some of them
// might block the IP/something...
var websiteLAT = make(map[string]time.Time, 0) // Last Access Time
var websiteLATMux = sync.Mutex{}

const waitTime = 30 * time.Second

func GetResponse(link string, website string) (*http.Response, error) {
	websiteLATMux.Lock()
	lat := websiteLAT[website]
	websiteLATMux.Unlock()

	if !lat.IsZero() {
		sleepTime := waitTime - time.Since(lat)
		if sleepTime > 0 {
			log.Printf("%s sleeping for %s...\n", website, sleepTime.String())
			time.Sleep(sleepTime)
		}
	}

	websiteLATMux.Lock()
	websiteLAT[website] = time.Now()
	websiteLATMux.Unlock()

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	myURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", myURL.Host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) FxQuantum/122.0 AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		linkURL, err := url.Parse(link)
		if err != nil {
			return nil, errors.New("unable to parse link " + link)
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			return nil, errors.New("unable to parse HTTP header \"Location\" of link " + link + " after redirection")
		}
		newLink := linkURL.ResolveReference(redirectURL)
		return GetResponse(newLink.String(), website)
	}

	return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
}

func CompileAddress(district, street string) (address string) {
	if district == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + district
	} else {
		address = "Vilnius, " + district + ", " + street
	}
	return
}

func CompileAddressWithStreet(district, street, houseNumber string) (address string) {
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
