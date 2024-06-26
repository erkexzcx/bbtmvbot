package website

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Post struct {
	Website     string
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
	Fee         bool
}

type Payload struct {
	Content any `json:"content"`
    Embeds []Embed `json:"embeds"`
}

type Embed struct {
	Title 		string `json:"title"`
	Description string `json:"description"`
	Link 		string `json:"url"`
	Color 		int `json:"color"`
	Timestamp	string `json:"timestamp"`
	Fields 		[]EmbedField `json:"fields"`
}

type EmbedField struct {
	Name 	string `json:"name"`
	Value 	string `json:"value"`
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

func (p *Post) DetectFee() {
	// Default value is false.

	// Process description
	processedDescription := strings.ToLower(p.Description)
	processedDescription = lithuanianReplacer.Replace(processedDescription)

	// Check against keywords
	for _, v := range feeKeywords {
		if strings.Contains(processedDescription, v) {
			p.Fee = true
			return
		}
	}

	// Check against regexes
	for _, v := range feeRegexes {
		if v.MatchString(processedDescription) {
			p.Fee = true
			return
		}
	}
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

func (p *Post) FormatDiscordMessage(IDInDatabase int64) *Payload  {
	var embed = &Embed{
		Title: fmt.Sprintf("Rastas naujas butas (ID: %d)", IDInDatabase),
		Description: "",
		Link: p.Link,
		Color: 0x33ff00,
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: []EmbedField{},
	}

	if p.Address != "" {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Adresas",
			Value: fmt.Sprintf("[%s](https://maps.google.com/?q=%s)", p.Address, url.QueryEscape(p.Address)),
		})
	}

	if p.Price != 0 && p.Area != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Kaina",
			Value: fmt.Sprintf("%d€ *(%.2f€/m²)*", p.Price, float64(p.Price)/float64(p.Area)),
		})
	} else if p.Price != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Kaina",
			Value: fmt.Sprintf("%d€", p.Price),
		})
	}

	if p.Rooms != 0 && p.Area != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Kambariai",
			Value: fmt.Sprintf("%d *(%dm²)*", p.Rooms, p.Area),
		})
	} else if p.Rooms != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Kambariai",
			Value: fmt.Sprintf("%d", p.Rooms),
		})
	}
	
	if p.Phone != "" {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Tel. numeris",
			Value: p.Phone,
		})
	}

	if p.Year != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Statybos metai",
			Value: fmt.Sprintf("%d", p.Year),
		})
	}

	if p.Heating != "" {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Šildymo tipas",
			Value: p.Heating,
		})
	}

	if p.Floor != 0 && p.FloorTotal != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Aukštas",
			Value: fmt.Sprintf("%d/%d", p.Floor, p.FloorTotal),
		})
	} else if p.Floor != 0 {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Aukštas",
			Value: fmt.Sprintf("%d", p.Floor),
		})
	}

	if p.Link != "" {
		embed.Fields = append(embed.Fields, EmbedField{
			Name: "Nuoroda",
			Value: fmt.Sprintf("[%s](%s)", p.Website, p.Link),
		})
	}

	return &Payload{
		Content: nil,
		Embeds: []Embed{*embed},
	}
}

func (p *Post) ProcessFields() {
	p.Address = strings.TrimSpace(p.Address)
	p.Heating = strings.TrimSpace(p.Heating)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Phone = strings.ReplaceAll(p.Phone, " ", "")
}
