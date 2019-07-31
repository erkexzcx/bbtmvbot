# BBTMV bot

This bot scans the most popular flat rent portals for latest posts, which will be sent to subscribed users using Telegram app.

See it in action (only in Lithuanian language): https://t.me/butuskelbimubot

## Installation

### Set-up Telegram bot

Using [BotFather](https://t.me/BotFather) create your own Telegram bot.

Also using BotFather use command `/setcommands` and update your bot commands:
```
help - Information how to use
config - Your personal preferences
enable - Enable notifications
disable - Disable notifications
stats - Interesting statistics
```
Once you set-up bot, you should have **API key**.

### Set-up application

Installation will cover only Arch Linux, but it's basically the same for any other Linux distribution.

Download project:
```
git clone https://github.com/erkexzcx/BBTMV-bot.git
cd BBTMV-bot
```

Install required packages:
```
pacman -S go sqlite
```

Then install required go dependencies using below commands:
```
go get -u gopkg.in/tucnak/telebot.v2
go get github.com/mattn/go-sqlite3
go get github.com/PuerkitoBio/goquery
```

Then create file `telegram.conf` and save only your **API key** in there. Nothing else.

Then using below command create empty database:
```
cat createdb.sql | sqlite3 database.db
```

Then compile this application:
```
go build src/*.go
```

And finally, run:
```
./main
```