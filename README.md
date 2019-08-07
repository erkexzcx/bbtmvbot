# BBTMV bot

This bot scans the most popular flat rent portals for latest posts, which will be sent to subscribed users using Telegram app.

Hardware requirements are so low that you can even run this completelly fine on a lowest-end SBC. On RPI0W, RAM usage is only about 8mb and CPU load is up to 25%, so you can run this on any _potato_ you want :)

See it in action (Lithuanian language): https://t.me/butuskelbimubot (yes, hosted on RPI0W)

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
Once you set-up bot, you should have your bot's Telegram **API key**.

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

Then create file `telegram.conf` and save your bot's Telegram **API key**:
```
echo "<your_api_key>" > telegram.conf
```

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

Time to time you wish to update this app. Pull latest changes & recompile:
```
git pull
go build src/*.go
```
