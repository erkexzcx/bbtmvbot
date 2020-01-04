# BBTMV bot

[![License](https://img.shields.io/github/license/erkexzcx/BBTMV-bot)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/erkexzcx/BBTMV-bot)](https://goreportcard.com/report/github.com/erkexzcx/BBTMV-bot)

This bot scans the most popular flat rent portals for latest posts in Vilnius, which will be sent to subscribed users using Telegram app.

Hardware requirements are so low that you can even run this completelly fine on a lowest-end SBC. On RPI0W, RAM usage is only about 8mb and CPU load is only few percent, so you can run this on any _potato_ you want :)

See it in action (Lithuanian language): https://t.me/butuskelbimubot (hosted on RPI3B+)

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

Install `go` and `sqlite` packages. Note that you need latest version of `go` (most debian/ubuntu distributions has outdated version).

Then download the project:
```
git clone https://github.com/erkexzcx/BBTMV-bot.git
cd bbtmvbot
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
go build -o bbtmvbot
```

And finally, run:
```
./bbtmvbot
```

To update this app, simply pull latest changes & recompile:
```
git pull
go build -o bbtmvbot
```

### (Optional) InfluxDB/Grafana monitoring

Use `<ipaddress>:3999/influx` to print influxdb (line protocol)[https://docs.influxdata.com/influxdb/v1.7/write_protocols/line_protocol_tutorial/] output. Grafana, InfluxDB, simple bash script to evaluate this output + cronjob every minute = amazing statistics & charts. :)
