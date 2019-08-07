package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func compileAddressWithStreet(state, street, houseNumber string) (address string) {
	if state == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + state
	} else if houseNumber == "" {
		address = "Vilnius, " + state + ", " + street
	} else {
		address = "Vilnius, " + state + ", " + street + " " + houseNumber
	}
	return
}

func compileAddress(state, street string) (address string) {
	if state == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + state
	} else {
		address = "Vilnius, " + state + ", " + street
	}
	return
}

func downloadAsBytes(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func downloadAsGoqueryDocument(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return goquery.NewDocumentFromReader(res.Body)
}
