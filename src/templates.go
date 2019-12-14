package main

const statsTemplate = `Boto statistinė informacija:
» *Lankytojai:* %d (iš jų %d įjungę pranešimus)
» *Nuscreipinta skelbimų:* %d
» *Vidutiniai kainų nustatymai:* Nuo %d€ iki %d€
» *Vidutiniai kambarių sk. nustatymai:* Nuo %d iki %d`

const helpText = `*Galimos komandos:*
/help - Pagalba
/config - Konfiguruoti pranešimus
/enable - Įjungti pranešimus
/disable - Išjungti pranešimus
/stats - Boto statistika

*Aprašymas:*
Tai yra botas (scriptas), kuris skenuoja įvairius populiariausius būtų nuomos portalus ir ieško būtų Vilniuje, kuriems (potencialiai) nėra taikomas tarpininkavimo mokestis. Jeigu kyla klausimų arba pasitaikė pranešimas, kuriame yra tarpininkavimo mokestis - chat grupė https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA`

const errorText = `Įvyko duomenų bazės klaida! Praneškite apie tai chat grupėje https://t.me/joinchat/G2hnjQ80K5qZaeHTEOFrDA`

const configText = "Naudokite tokį formatą:\n\n```\n/config <kaina_nuo> <kaina_iki> <kambariai_nuo> <kambariai_iki> <metai_nuo>\n```\nPavyzdys:\n```\n/config 200 330 1 2 2000\n```"

const configErrorText = "Neteisinga įvestis! " + configText

const activeSettings = `
Jūsų aktyvūs nustatymai:
» *Pranešimai:* %s
» *Kaina:* Nuo %d€ iki %d€
» *Kambarių sk.:* Nuo %d iki %d
» *Metai nuo:* %d`
