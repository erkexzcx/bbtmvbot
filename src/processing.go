package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type post struct {
	url         string
	phone       string
	description string
	address     string
	heating     string
	floor       int
	floorTotal  int
	area        int
	price       int
	rooms       int
	year        int
}

// Must be lowercase!!!
var exclusionKeywords = []string{
	"(yra mokestis)",
	"mokestis (jei butas",
	"\ntaikomas tarpininkavimas",
	"tiks vienkartinis tarpinink",
}

var exlusionRegexes = map[string]*regexp.Regexp{
	"regex1": regexp.MustCompile(`(agent|tarpinink|vienkart)\S+ mokestis[\s:]{0,3}\d+`),
	"regex2": regexp.MustCompile(`\d+\s{0,1}\S+ (agent|tarpinink|vienkart)\S+ (tarp|mokest)\S+`),
	"regex3": regexp.MustCompile(`\W(yra|bus) (taikoma(s|)|imama(s|)|vienkartinis|agent\S+)( vienkartinis|) (agent|tarpinink|mokest)\S+`),
	"regex4": regexp.MustCompile(`\Wtiks[^\s\w]{0,1}\s{0,1}(bus|yra|) (taikoma(s|)|imama(s|))`),
	"regex5": regexp.MustCompile(`\W(yra |)(taikoma(s|)|imama(s|)|vienkartinis|sutarties)( sutarties|) sudar\S+ mokestis`),
	"regex6": regexp.MustCompile(`(ui|ir) (yra |)(taikoma(s|)|imama(s|)) (vienkart|agent|tarpinink|mokest)\S+`),
	"regex7": regexp.MustCompile(`(vienkartinis |)(agent|tarpinink)\S+ mokest\S+,{0,1} jei`),
}

// Note that post is already checked against DB in parsing functions!
func (p post) processPost() {

	// Check if we need to exclude this post
	excluded, reason := p.isExcluded()
	if excluded {
		rowID := p.addToDB(true, reason)
		log.Println("// Excluded ", rowID, "|", reason)
		return
	}

	// Add to database, so it won't be sent again
	rowID := p.addToDB(false, "")

	// Send to users
	p.sendToUsers(rowID)

	// Show debug info
	log.Printf(
		"{ID:%d URL:%d Phon:%s Desc:%d Addr:%d Heat:%d Floor:%d FlTot:%d Area:%d Price:%d Room:%d Year:%d}\n",
		rowID, len(p.url), p.phone, len(p.description), len(p.address), len(p.heating), p.floor, p.floorTotal, p.area, p.price, p.rooms, p.year,
	)
}

func (p post) isExcluded() (excluded bool, reason string) {

	// Convert description to lowercase and store here
	desc := strings.ToLower(p.description)

	// Check if description contains exclusion keyword
	for _, v := range exclusionKeywords {
		if !strings.Contains(desc, v) {
			continue
		}
		return true, v
	}

	// Now check against regex rules
	for k, v := range exlusionRegexes {
		arr := v.FindStringSubmatch(desc)
		if len(arr) >= 1 {
			return true, "/" + k + "/"
		}
	}

	// Skip posts without price
	if p.price == 0 {
		return true, "0eur price"
	}

	return false, ""
}

func (p *post) compileMessage(ID int64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%d. %s\n", ID, p.url)

	if p.phone != "" {
		fmt.Fprintf(&b, "» *Tel. numeris:* [%s](tel:%s)\n", p.phone, p.phone)
	}

	if p.address != "" {
		fmt.Fprintf(&b, "» *Adresas:* [%s](https://maps.google.com/?q=%s)\n", p.address, url.QueryEscape(p.address))
	}

	if p.price != 0 && p.area != 0 {
		fmt.Fprintf(&b, "» *Kaina:* `%d€ (%.2f€/m²)`\n", p.price, float64(p.price)/float64(p.area))
	} else if p.price != 0 {
		fmt.Fprintf(&b, "» *Kaina:* `%d€`\n", p.price)
	}

	if p.rooms != 0 && p.area != 0 {
		fmt.Fprintf(&b, "» *Kambariai:* `%d (%dm²)`\n", p.rooms, p.area)
	} else if p.rooms != 0 {
		fmt.Fprintf(&b, "» *Kambariai:* `%d`\n", p.rooms)
	}

	if p.year != 0 {
		fmt.Fprintf(&b, "» *Statybos metai:* `%d`\n", p.year)
	}

	if p.heating != "" {
		fmt.Fprintf(&b, "» *Šildymo tipas:* `%s`\n", p.heating)
	}

	if p.floor != 0 && p.floorTotal != 0 {
		fmt.Fprintf(&b, "» *Aukštas:* `%d/%d`\n", p.floor, p.floorTotal)
	} else if p.floor != 0 {
		fmt.Fprintf(&b, "» *Aukštas:* `%d`\n", p.floor)
	}

	return b.String()
}

func (p post) addToDB(excluded bool, reason string) int64 {

	var excludedVal int
	if excluded {
		excludedVal = 1
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO posts(url, excluded, reason) values (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(p.url, excludedVal, reason)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return -1
	}
	return lastInsertedID

}

func (p post) postExistsInDB() (bool, error) {
	var count int // Will store count here

	sql := fmt.Sprintf("SELECT COUNT(*) AS count FROM posts WHERE url=\"%s\" LIMIT 1", p.url)
	err := db.QueryRow(sql).Scan(&count)

	if err != nil {
		log.Println(err)
		return false, err
	}
	if count != 1 {
		return false, nil
	}
	return true, nil

}

func (p post) sendToUsers(postID int64) {

	sql := fmt.Sprintf(`
	SELECT id FROM users WHERE
	enabled=1 AND
	((price_from <= %d AND price_to >= %d) OR %d = 0) AND
	((rooms_from <= %d AND rooms_to >= %d) OR %d = 0) AND
	(year_from <= %d OR %d = 0)`,
		p.price, p.price, p.price,
		p.rooms, p.rooms, p.rooms,
		p.year, p.year)

	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		err = rows.Scan(&userID)
		if err != nil {
			log.Println(err)
			return
		}

		// Send to user:
		sendTo(&tb.User{ID: userID}, p.compileMessage(postID))
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}

}
