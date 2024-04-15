package website

import (
	"strings"
	"testing"
)

type PostData struct {
	Provided string
	Expected bool
}

var PostTestDataFee = []string{
	`
Jei butas tiks, bus įmamas vienkartinis agentūros mokestis.
		`,
	`
Pasirašant nuomos sutartį yra taikomas vienkartinis sutarties sudarymo mokestis agentūrai 250 eur. 
		`,
	`
Vienkartinis agentūros mokestis 200 eurų.
		`,
	`
Bus taikomas vienkartinis agentūros mokestis – 200 eur.
------------------------------------------------------------------------------------------------
		`,
	`
- Centrinis šildymas 
- Vienkartinis tarpininkavimo mokestis (jei butas tiks) 
 
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Objekto ID:10395 
		`,
	`
KAINA: 500 EUR
Vienkartinis tarpininkavimo mokestis (jei butas tiks).
		`,
	`
Daugiau informacijos suteiksime tel.867786879 Skambinkite Jums patogiu metu.
Jei butas tiks, bus taikomas minimalus vienkartinis agentros mokestis.
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS. JEI BUTAS TIKS, BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS 
		`,
	`
KAINA: 340 EUR
- Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
	`
Butas išnuomojamas ilgam laikui. 
Vienkartinis agentūros mokestis 180 eurų.
		`,
	`
***************************************************************
Jei butas tiks bus imamas vienkartinis agentūros mokestis - 150 eurų
***************************************************************
		`,
	`
Centrinis-kolektorinis šildymas. Kitos paslaugos apie 17 €. 
Vienkartinis tarpininkavimo mokestis (jei butas tiks). 
 
		`,
	`
************************************* 
Jei butas tiks bus imamas vienkartinis agentūros mokestis! 
*************************************
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Objekto ID 10395 
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS.
JEI BUTAS TIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS
Objekto ID 9362
		`,
	`
Skambinkite Jums patogiu laiku, atsakysime į Jums rūpimus klausimus.
JEIGU BUTAS TIKS, BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Per avere piu informazioni sul l'affitto di questo appartamento chiamate a qualsiasi ora.
		`,
	`
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS.
JEI BUTAS TIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
		`,
	`
******************************************************
Taikomas vienkartinis tarpininkavimo mokestis.
Nekilnojamo turto agentūra OPPA
		`,
	`
Jei butas tiks, bus taikomas vienkartinis tarpininkavimo mokestis.
Skambinkite Jums patogiu laiku, atsakysime į Jums rūpimus klausimus.
		`,
	`
	
– Šitam butui taikomas vienkartinis agentūros mokestis.
		`,
	`
stalas, šaldytuvas, skalbimo mašina. Bute plastikiniai langai. Nuomos kaina 120eur./mėn. (už komunalinės paslaugos mokėti nereikia).
Jei kambarys tiks, bus imamas vienkartinis tarpininkavimo mokestis.
		`,
	`
• KAINA: 450 €
• Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
	`
Jei butas tiks, bus taikomas vienkartinis agentūros mokestis.
		`,
	`
KAINA: 450 Eur
Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
	`
Jei butas tiks bus imamas vienkartinis tarpininkavimo mokestis.
		`,
	`
• Kitos paslaugos apie 20 €.
• Vienkartinis tarpininkavimo mokestis (jei butas tiks).
		`,
	`
JEI BUTAS TIKS - BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS
Galima skambinti ir poilsio dienomis, jei neatsiliepiu - perskambinu.
		`,
	`
Bus sudaroma nuomos sutartis ,
Jei butas tiks bus pritaikytas vienkartinis tarpininkavimo mokestis-100€`,
}

var PostTestDataNoFee = []string{
	`
Tarpininkavimo mokestis nera taikomas!
		`,
	`
Nėra tarpininkavimo mokesčio.
		`,
	`
nėra tarpininkavimo mokescio
		`,
	`
nėra sutarties sudarymo mokesčio
		`,
	`
tarpininkavimo mokescio nera.
		`,
	`
tarpininkavimo mokesčio nėra
		`,
	`
nėra taikomas tarpininkavimo mokestis
		`,
}

func TestDetectFee(t *testing.T) {
	var p *Post
	for _, desc := range PostTestDataFee {
		p = &Post{Description: desc}
		p.DetectFee()
		if p.Fee == false {
			t.Errorf("Result was incorrect, '%s' expected '%t', got: '%t'.", strings.TrimSpace(desc), true, false)
		}
	}
	for _, desc := range PostTestDataNoFee {
		p = &Post{Description: desc}
		p.DetectFee()
		if p.Fee == true {
			t.Errorf("Result was incorrect, '%s' expected '%t', got: '%t'.", strings.TrimSpace(desc), true, false)
		}
	}
}

// Test for duplicate test data - has Fee
func TestTestHasFee(t *testing.T) {
	for k, v := range PostTestDataFee {
		for kk, vv := range PostTestDataFee {
			if k != kk && strings.EqualFold(v, vv) {
				t.Errorf("Duplicating test data in rows %d and %d: '%s'.", k, kk, v)
			}
		}
	}
}

// Test for duplicate test data - has no Fee
func TestTestHasNoFee(t *testing.T) {
	for k, v := range PostTestDataNoFee {
		for kk, vv := range PostTestDataNoFee {
			if k != kk && strings.EqualFold(v, vv) {
				t.Errorf("Duplicating test data in rows %d and %d: '%s'.", k, kk, v)
			}
		}
	}
}
