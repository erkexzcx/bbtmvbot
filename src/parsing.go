package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func downloadAsReader(url string) (io.ReadCloser, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return res.Body, nil
}

func downloadAsBytes(url string) ([]byte, error) {
	r, err := downloadAsReader(url)
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}
