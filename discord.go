package bbtmvbot

import (
	"bbtmvbot/website"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func filterPost(post *website.Post, priceFrom int, priceTo int, roomsFrom int, roomsTo int) bool {
	if post.Price <= priceFrom || post.Price >= priceTo || post.Rooms <= roomsFrom || post.Rooms >= roomsTo {
		return false
	}
	return true
}

func sendDiscord(post *website.Payload, url string) error {
	startTime := time.Now()

    jsonData, err := json.Marshal(post)
    if err != nil {
        return err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    _, err = io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
	elapsedTime = time.Since(startTime)

	time.Sleep(30*time.Millisecond - elapsedTime)
	return nil
}