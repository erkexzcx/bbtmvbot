package website

import (
	"strings"
	"testing"
)

type PostData struct {
	Provided string
	Expected bool
}

var PostTestData = []PostData{
	{
		Provided: `
Jei butas tiks, bus įmamas vienkartinis agentūros mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
Pasirašant nuomos sutartį yra taikomas vienkartinis sutarties sudarymo mokestis agentūrai 250 eur. 
		`,
		Expected: true,
	},
	{
		Provided: `
Vienkartinis agentūros mokestis 200 eurų.
		`,
		Expected: true,
	},
	{
		Provided: `
Bus taikomas vienkartinis agentūros mokestis – 200 eur.
------------------------------------------------------------------------------------------------
		`,
		Expected: true,
	},
	{
		Provided: `
- Centrinis šildymas 
- Vienkartinis tarpininkavimo mokestis (jei butas tiks) 
 
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Objekto ID:10395 
		`,
		Expected: true,
	},
	{
		Provided: `
KAINA: 500 EUR
Vienkartinis tarpininkavimo mokestis (jei butas tiks).
		`,
		Expected: true,
	},
	{
		Provided: `
Daugiau informacijos suteiksime tel.867786879 Skambinkite Jums patogiu metu.
Jei butas tiks, bus taikomas minimalus vienkartinis agentros mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS. JEI BUTAS TIKS, BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS 
		`,
		Expected: true,
	},
	{
		Provided: `
KAINA: 340 EUR
- Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
		Expected: true,
	},
	{
		Provided: `
Butas išnuomojamas ilgam laikui. 
Vienkartinis agentūros mokestis 180 eurų.
		`,
		Expected: true,
	},
	{
		Provided: `
***************************************************************
Jei butas tiks bus imamas vienkartinis agentūros mokestis - 150 eurų
***************************************************************
		`,
		Expected: true,
	},
	{
		Provided: `
Centrinis-kolektorinis šildymas. Kitos paslaugos apie 17 €. 
Vienkartinis tarpininkavimo mokestis (jei butas tiks). 
 
		`,
		Expected: true,
	},
	{
		Provided: `
************************************* 
Jei butas tiks bus imamas vienkartinis agentūros mokestis! 
*************************************
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Objekto ID 10395 
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU
JEIGU BUTAS TIKS IR PATIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS.
JEI BUTAS TIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS
Objekto ID 9362
		`,
		Expected: true,
	},
	{
		Provided: `
Skambinkite Jums patogiu laiku, atsakysime į Jums rūpimus klausimus.
JEIGU BUTAS TIKS, BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
Per avere piu informazioni sul l'affitto di questo appartamento chiamate a qualsiasi ora.
		`,
		Expected: true,
	},
	{
		Provided: `
SKAMBINKITE JUMS PATOGIU LAIKU IR SUTEIKSIU DAUGIAU INFORMACIJOS.
JEI BUTAS TIKS BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS.
		`,
		Expected: true,
	},
	{
		Provided: `
******************************************************
Taikomas vienkartinis tarpininkavimo mokestis.
Nekilnojamo turto agentūra OPPA
		`,
		Expected: true,
	},
	{
		Provided: `
Jei butas tiks, bus taikomas vienkartinis tarpininkavimo mokestis.
Skambinkite Jums patogiu laiku, atsakysime į Jums rūpimus klausimus.
		`,
		Expected: true,
	},
	{
		Provided: `
	
– Šitam butui taikomas vienkartinis agentūros mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
stalas, šaldytuvas, skalbimo mašina. Bute plastikiniai langai. Nuomos kaina 120eur./mėn. (už komunalinės paslaugos mokėti nereikia).
Jei kambarys tiks, bus imamas vienkartinis tarpininkavimo mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
• KAINA: 450 €
• Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
		Expected: true,
	},
	{
		Provided: `
Jei butas tiks, bus taikomas vienkartinis agentūros mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
KAINA: 450 Eur
Vienkartinis tarpininkavimo mokestis (jei butas tiks)
		`,
		Expected: true,
	},
	{
		Provided: `
Jei butas tiks bus imamas vienkartinis tarpininkavimo mokestis.
		`,
		Expected: true,
	},
	{
		Provided: `
• Kitos paslaugos apie 20 €.
• Vienkartinis tarpininkavimo mokestis (jei butas tiks).
		`,
		Expected: true,
	},
	{
		Provided: `
JEI BUTAS TIKS - BUS TAIKOMAS VIENKARTINIS TARPININKAVIMO MOKESTIS
Galima skambinti ir poilsio dienomis, jei neatsiliepiu - perskambinu.
		`,
		Expected: true,
	},
	/* TEST FOR DESCRIPTIONS WITHOUT FEE */
	{
		Provided: `
Tarpininkavimo mokestis nera taikomas!
		`,
		Expected: false,
	},
	{
		Provided: `
Nėra tarpininkavimo mokesčio.
		`,
		Expected: false,
	},
	{
		Provided: `
nėra tarpininkavimo mokescio
		`,
		Expected: false,
	},
	{
		Provided: `
nėra sutarties sudarymo mokesčio
		`,
		Expected: false,
	},
	{
		Provided: `
tarpininkavimo mokescio nera.
		`,
		Expected: false,
	},
	{
		Provided: `
tarpininkavimo mokesčio nėra
		`,
		Expected: false,
	},
	{
		Provided: `
nėra taikomas tarpininkavimo mokestis
		`,
		Expected: false,
	},
	{
		Provided: `
nuomos mokestis + komunaliniai
		`,
		Expected: false,
	},
}

func TestHasFee(t *testing.T) {
	p := &Post{Price: 300}
	for _, v := range PostTestData {
		p.Description = v.Provided
		if v.Expected == true {
			if excl := p.IsExcludable(); excl != v.Expected {
				t.Errorf("Result was incorrect, '%s' expected '%t', got: '%t'.", strings.TrimSpace(v.Provided), v.Expected, excl)
			}
		} else {
			if excl := p.IsExcludable(); excl != v.Expected {
				t.Errorf("Result was incorrect, '%s' expected '%t', got: '%t'.", strings.TrimSpace(v.Provided), v.Expected, excl)
			}
		}

	}
}

func TestTestHasFee(t *testing.T) {
	for k, v := range PostTestData {
		for kk, vv := range PostTestData {
			if k != kk && strings.EqualFold(v.Provided, vv.Provided) {
				t.Errorf("Duplicating test data in rows %d and %d: '%s'.", k, kk, v.Provided)
			}
		}
	}
}
