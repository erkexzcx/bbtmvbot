package website

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func GetResponse(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	myURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", myURL.Host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36")
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
		return GetResponse(newLink.String())
	}

	return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
}
