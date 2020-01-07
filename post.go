package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Post contains sendable data to the user
type Post struct {
	Link        string
	Phone       string
	Description string
	Address     string
	Heating     string
	Floor       int
	FloorTotal  int
	Area        int
	Price       int
	Rooms       int
	Year        int
}

// Ensure these are lowercase
var feeKeywords = []string{
	"(yra mokestis)",
	"mokestis (jei butas",
	"\ntaikomas tarpininkavimas",
	"tiks vienkartinis tarpinink",
}

// Ensure these are lowercase
var feeRegexes = map[string]*regexp.Regexp{
	"regex1": regexp.MustCompile(`(agent|tarpinink|vienkart)\S+ mokestis[\s:-]{0,3}\d+`),
	"regex2": regexp.MustCompile(`\d+\s{0,1}\S+ (agent|tarpinink|vienkart)\S+ (tarp|mokest)\S+`),
	"regex3": regexp.MustCompile(`\W(ira|bus) (taikoma(s|)|imama(s|)|vienkartinis|agent\S+)( vienkartinis|) (agent|tarpinink|mokest)\S+`),
	"regex4": regexp.MustCompile(`\Wtiks[^\s\w]{0,1}\s{0,1}(bus|ira|) (taikoma(s|)|imama(s|))`),
	"regex5": regexp.MustCompile(`\W(ira |)(taikoma(s|)|imama(s|)|vienkartinis|sutarties)( sutarties|) sudar\S+ mokestis`),
	"regex6": regexp.MustCompile(`(ui|ir) (ira |)(taikoma(s|)|imama(s|)) (vienkart|agent|tarpinink|mokest)\S+`),
	"regex7": regexp.MustCompile(`(vienkartinis |)(agent|tarpinink)\S+ mokest\S+,{0,1} jei`),
	"regex8": regexp.MustCompile(`[^\w\s](\s|)(taikoma(s|)|imama(s|)|vienkartinis|agent\S+)( vienkartinis|) (agent|tarpinink|mokest)\S+`),
}

// Fee returns true (and a reason) if post contain broker's fee.
func (p *Post) hasFee() (excluded bool, reason string) {

	// Process description:
	d := processDescription(p.Description)

	// Check against keywords
	for _, v := range feeKeywords {
		if !strings.Contains(d, v) {
			continue
		}
		return true, v
	}

	// Check against regexes
	for k, v := range feeRegexes {
		arr := v.FindStringSubmatch(d)
		if len(arr) >= 1 {
			return true, k
		}
	}

	// Ignore 0 eur price
	if p.Price == 0 {
		return true, "0 eur price"
	}

	return false, ""
}

func processDescription(d string) string {
	// Convert description to lowercase
	d = strings.ToLower(d)

	// Remove diacrytics from Lithuanian language
	r := strings.NewReplacer(
		"ą", "a",
		"č", "c",
		"ę", "e",
		"ė", "e",
		"į", "i",
		"š", "s",
		"ų", "u",
		"ū", "u",
		"ž", "z",
		"y", "i", // Replace y with i, because some people are bad at writting
	)
	d = r.Replace(d)

	return d
}

// Message compiles post information to a sendable Telegram message (string).
func (p *Post) message(ID int64) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%d. %s\n", ID, p.Link)

	if p.Phone != "" {
		fmt.Fprintf(&sb, "» *Tel. numeris:* [%s](tel:%s)\n", p.Phone, p.Phone)
	}

	if p.Address != "" {
		fmt.Fprintf(&sb, "» *Adresas:* [%s](https://maps.google.com/?q=%s)\n", p.Address, url.QueryEscape(p.Address))
	}

	if p.Price != 0 && p.Area != 0 {
		fmt.Fprintf(&sb, "» *Kaina:* `%d€ (%.2f€/m²)`\n", p.Price, float64(p.Price)/float64(p.Area))
	} else if p.Price != 0 {
		fmt.Fprintf(&sb, "» *Kaina:* `%d€`\n", p.Price)
	}

	if p.Rooms != 0 && p.Area != 0 {
		fmt.Fprintf(&sb, "» *Kambariai:* `%d (%dm²)`\n", p.Rooms, p.Area)
	} else if p.Rooms != 0 {
		fmt.Fprintf(&sb, "» *Kambariai:* `%d`\n", p.Rooms)
	}

	if p.Year != 0 {
		fmt.Fprintf(&sb, "» *Statybos metai:* `%d`\n", p.Year)
	}

	if p.Heating != "" {
		fmt.Fprintf(&sb, "» *Šildymo tipas:* `%s`\n", p.Heating)
	}

	if p.Floor != 0 && p.FloorTotal != 0 {
		fmt.Fprintf(&sb, "» *Aukštas:* `%d/%d`\n", p.Floor, p.FloorTotal)
	} else if p.Floor != 0 {
		fmt.Fprintf(&sb, "» *Aukštas:* `%d`\n", p.Floor)
	}

	return sb.String()
}

func (p *Post) debugMessage(rowID int64) string {
	return fmt.Sprintf(
		"\tID:%d Link:%d Tel:%s Desc:%d Addr:%d Heat:%d Fl:%d FlTot:%d Area:%d Price:%d Room:%d Year:%d",
		rowID, len(p.Link), p.Phone, len(p.Description), len(p.Address), len(p.Heating), p.Floor, p.FloorTotal, p.Area, p.Price, p.Rooms, p.Year,
	)
}

// Send sends post to Telegram subscribers.
func (p *Post) send(postID int64) {
	query := `
	SELECT id FROM users WHERE
	enabled=1 AND
	((price_from <= ? AND price_to >= ?) OR ? = 0) AND
	((rooms_from <= ? AND rooms_to >= ?) OR ? = 0) AND
	(year_from <= ? OR ? = 0)`

	rows, err := db.Query(query,
		p.Price, p.Price,
		p.Price, p.Rooms, p.Rooms,
		p.Rooms, p.Year, p.Year)
	if err != nil {
		panic(err)
		//return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err = rows.Scan(&userID); err != nil {
			panic(err)
			//return
		}

		sendTo(&tb.User{ID: userID}, p.message(postID)) // Send to user
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
}

// InDatabase returns TRUE if post is already in database.
func (p *Post) InDatabase() bool {
	query := "SELECT COUNT(*) AS count FROM posts WHERE url=? LIMIT 1"
	var count int
	err := db.QueryRow(query, p.Link).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count > 0
}

// ToDatabase saves post to database, so it won't be sent again.
func (p *Post) toDatabase(hasFee bool, reason string) (rowID int64) {
	query := "INSERT INTO posts(url, excluded, reason) values (?, ?, ?)"
	res, err := db.Exec(query, p.Link, hasFee, reason)
	if err != nil {
		panic(err)
	}

	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	return lastInsertedID
}

// Handle handles post (e.g. sends to subscribers).
func (p *Post) Handle() {
	p.Phone = strings.TrimSpace(p.Phone)
	p.Description = strings.TrimSpace(p.Description)
	p.Address = strings.TrimSpace(strings.Title(p.Address))
	p.Heating = strings.TrimSpace(p.Heating)

	// Exclude
	if excluded, reason := p.hasFee(); excluded {
		rowID := p.toDatabase(true, reason)
		log.Println("// Excluded ", rowID, "|", reason)
		return
	}

	rowID := p.toDatabase(false, "")

	p.send(rowID)

	// Debug info
	log.Println(p.debugMessage(rowID))
}
