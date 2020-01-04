package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func parseKampas() {
	type kampasPosts struct {
		Hits []struct {
			ID          int      `json:"id"`
			Title       string   `json:"title"`
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
	var results kampasPosts

	// Download page
	contents, err := fetch(parseLinkKampas)
	if err != nil {
		log.Println(err)
		return
	}

	// Decode json
	json.Unmarshal(contents, &results)

	// Iterate through posts
	for _, v := range results.Hits {

		p := &Post{}

		p.Link = fmt.Sprintf("https://www.kampas.lt/skelbimai/%d", v.ID) // https://www.kampas.lt/skelbimai/504506

		// Skip if already in database:
		if p.InDatabase() {
			return
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
		p.Heating = strings.ReplaceAll(p.Heating, "thermostat", "termostatas")

		//p.Phone = "" // Impossible
		p.Description = strings.ReplaceAll(v.Description, "<br/>", "\n")
		p.Address = v.Title
		p.Floor = v.Objectfloor
		p.FloorTotal = v.Totalfloors
		p.Area = v.Objectarea
		p.Price = v.Objectprice
		p.Rooms = v.Totalrooms
		p.Year = v.Yearbuilt

		go p.Handle()
	}
}
