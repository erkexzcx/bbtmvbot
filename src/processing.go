package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
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
	" bus taikomas vienkartinis agentūros mokestis",
	" bus taikomas vienkartinis agentūrinis mokestis",
	" bus taikomas vienkartinis agenturos mokestis",
	" bus taikomas vienkartinis agenturinis mokestis",
	" bus taikomas vienkartinis tarpininkavimo mokestis",
	",bus taikomas vienkartinis agentūros mokestis",
	",bus taikomas vienkartinis agentūrinis mokestis",
	",bus taikomas vienkartinis agenturos mokestis",
	",bus taikomas vienkartinis agenturinis mokestis",
	",bus taikomas vienkartinis tarpininkavimo mokestis",
	"yra taikomas vienkartinis agentūros mokestis",
	"yra taikomas vienkartinis agentūrinis mokestis",
	"yra taikomas vienkartinis agenturos mokestis",
	"yra taikomas vienkartinis agenturinis mokestis",
	"yra taikomas vienkartinis tarpininkavimo mokestis",
	"vienkartinis agentūros mokestis jei",
	"vienkartinis agentūrinis mokestis jei",
	"vienkartinis agenturos mokestis jei",
	"vienkartinis agenturinis mokestis jei",
	"vienkartinis tarpininkavimo mokestis jei",
	" tiks, bus taikoma",
	" tiks bus taikoma",
	" tiks, yra taikoma",
	" tiks yra taikoma",
	"taikomas vienkartinis tarpininkavimo mokestis",
	"tiks vienkartinis tarpininkavimo mokestis",
	"tarpininkavimo mokestis-",
	"tarpininkavimo mokestis -",
	"(yra mokestis)",
	" bus imamas vienkartinis",
	" bus imamas tarpininkavimo",
	" bus taikomas vienkartinis",
	".bus taikomas vienkartinis",
	",bus taikomas vienkartinis",
	" bus taikomas tarpininkavimo",
	"mokestis (jei butas tiks)",
	"ir imamas vienkartinis mokestis",
	",yra vienkartinis agent",
	" yra vienkartinis agent",
	".yra vienkartinis agent",
	"ui taikomas vienkartinis agent",
	"ui taikomas agent",
	"\nyra vienkartinis agent",
	"\nyra tarpininkavimo mokest",
	"\nyra vienkartinis tarpinink",
	"\ntaikomas tarpinink",
	"\ntaikomas vienkartinis tarpinink",
	"\ntaikomas vienkartinis agent",
	" vienkartinis sutarties sudarymo mokestis",
	"\nvienkartinis sutarties sudarymo mokestis",
	".vienkartinis sutarties sudarymo mokestis",
	" taikomas sutarties sudarymo mokestis",
	"\ntaikomas sutarties sudarymo mokestis",
	".taikomas sutarties sudarymo mokestis",
	" yra sutarties sudarymo mokestis",
	"\nyra sutarties sudarymo mokestis",
	".yra sutarties sudarymo mokestis",
}

var regexExclusion1 = regexp.MustCompile(`(agenturos|agentūros|agenturinis|agentūrinis|tarpininkavimo) mokestis[\s:]{0,3}\d+`)

// Note that post is already checked against DB in parsing functions!
func processPost(p post) {

	// Add to database, so it won't be sent again:
	insertedRowID := databaseAddPost(p)

	// Check if description contains exclusion keyword
	desc := strings.ToLower(p.description)
	for _, v := range exclusionKeywords {
		if strings.Contains(desc, v) {
			fmt.Println(">> Excluding", insertedRowID, "reason:", v)
			return
		}
	}

	// Passed blacklisted keywords test, so let's do some regex tests
	arr := regexExclusion1.FindStringSubmatch(desc)
	if len(arr) >= 1 {
		fmt.Println(">> Excluding", insertedRowID, "reason: /regex1/")
		return
	}

	// Skip posts without price and let user know:
	if p.price == 0 {
		fmt.Println(">> 0eur price", p.url)
		return
	}

	// Send to users
	databaseGetUsersAndSendThem(p, insertedRowID)

	// Show debug info:
	fmt.Printf(
		"{ID:%d URL:%d Phon:%s Desc:%d Addr:%d Heat:%d Floor:%d FlTot:%d Area:%d Price:%d Room:%d Year:%d}\n",
		insertedRowID, len(p.url), p.phone, len(p.description), len(p.address), len(p.heating), p.floor, p.floorTotal, p.area, p.price, p.rooms, p.year,
	)
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
