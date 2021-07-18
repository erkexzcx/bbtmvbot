package website

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

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

var lithuanianReplacer = strings.NewReplacer(
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

func (p *Post) IsExcludable() bool {
	processedDescription := strings.ToLower(p.Description)
	processedDescription = lithuanianReplacer.Replace(processedDescription)

	// Check against keywords
	for _, v := range feeKeywords {
		if strings.Contains(processedDescription, v) {
			return true
		}
	}

	// Check against regexes
	for _, v := range feeRegexes {
		if v.MatchString(processedDescription) {
			return true
		}
	}

	// Ignore 0 eur price
	return p.Price == 0
}

func (p *Post) FormatTelegramMessage(IDInDatabase int64) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%d. %s\n", IDInDatabase, p.Link)

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

func (p *Post) TrimFields() {
	p.Address = strings.TrimSpace(p.Address)
	p.Heating = strings.TrimSpace(p.Heating)
	p.Phone = strings.TrimSpace(p.Phone)
}
