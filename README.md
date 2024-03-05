# bbtmvbot

[![License](https://img.shields.io/github/license/erkexzcx/bbtmvbot)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/erkexzcx/bbtmvbot)](https://goreportcard.com/report/github.com/erkexzcx/bbtmvbot)

This bot scans the most popular flat rent portals for latest posts in Vilnius, which will be sent to subscribed users using Telegram app. This Telegram bot is suppossed to be multi-user, which means that you can host your own instance of this app and let others use your bot - either directly (private chat) or in group chats, so you and your significant other can get latest posts directly to your group chat.

Only `x86_64` architecture and Docker is supported. I suggest having at least 1GB RAM for this application to work, since it uses headless chromium browser and opens 7 tabs (websites) all at once for parsing.

# Set-up Telegram bot

Using [BotFather](https://t.me/BotFather) create your own Telegram bot.

Also using BotFather use command `/setcommands` and update your bot commands:
```
info - Information about BBTMV Bot
enable - Enable notifications
disable - Disable notifications
config - Configure bot settings
```
Once you set-up bot, you should have your bot's Telegram **API key**.

# Configure

## Create config.yml

Out of `config.example.yml` create `config.yml` file with the following contents (see my added comments below):

```
# Log level - set it to `info`, unless you are developing it, then `debug`
log_level: info # debug, info, warn, error

# Data dir - leave it unchanged
data_dir: data

# Telegram API key - BotFather will tell you the key
telegram_api_key: 1234567890:6xYrZZ2s_jrki5qgr8OxVBS566z2ZGF4Co7

# User agent (prefer Windows desktop browser) - Ensure it uses __latest__ Windows Chrome user-agent to avoid suspicion by anti-bot systems
user_agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
```

In case I forget to update above config example - always see `config.example.yml` for latest changes.

## Create docker-compose.yml

Create the following files/folders structure:

```
bbtmvbot
├── bbtmvbot
│   ├── config.yml      # <-- this is file
│   └── data            # <-- this is folder
└── docker-compose.yml  # <-- see below
```

Now this is how your `docker-compose.yml` should look like:

```
services:

  bbtmvbot:
    image: ghcr.io/erkexzcx/bbtmvbot
    container_name: bbtmvbot
    restart: unless-stopped
    environment:
      - TZ=Europe/Vilnius
    volumes:
      - ./bbtmvbot/config.yml:/config.yml
      - ./bbtmvbot/data:/data
```

Simply start it, then go to Telegram and issue `/config` (or whatever) command for this bot. Enjoy!

## (Optional) set-up logs DB

If you have no plans to contribute to this project (either by raising issues or pull requests), there is no need to do this step. Simply enjoy the application and I will _try_ to keep it up to date.

Since there is `./bbtmvbot/data/bbtmvbot.log` file which contains JSON encoded logs, it makes much easier to track websites for getting broken. If suddenly there are `error` level logs - something broken. If price of a certain website is only `0`, or description is always empty - something broke and needs to be fixed.

One way to tackle this problem is to set-up logs database. I prefer [OpenObserve](https://github.com/openobserve/openobserve) for storing and analyzing logs, while [Vector](https://github.com/vectordotdev/vector) for collecting and pushing logs.

Here is how you can expand your `docker-compose.yml` by appending below contents:

```
  openobserve:
    image: public.ecr.aws/zinclabs/openobserve
    container_name: openobserve
    restart: always
    environment:
      - ZO_DATA_DIR=/data
      - ZO_ROOT_USER_EMAIL=example@example.com
      - ZO_ROOT_USER_PASSWORD=Example123
    ports:
      - 5080:5080
    volumes:
      - ./openobserve/data:/data

  vector:
    image: timberio/vector:latest-alpine
    container_name: vector
    restart: always
    volumes:
      - ./vector/vector.yaml:/etc/vector/vector.yaml:ro
      - ./bbtmvbot/data/bbtmvbot.log:/bbtmvbot.log:ro
```

But before running anything, create `./vector/vector.yaml` file with the following contents:

```yaml
sources:
  bbtmvbot-log:
    type: file
    include:
      - /bbtmvbot.log
    file_key: ""

transforms:
  bbtmvbot-log-parser:
    inputs:
      - "bbtmvbot-log"
    type: "remap"
    source: |-
      . = parse_json!(.message)
      ._timestamp = del(.ts)

sinks:
  elasticsearch:
    type: elasticsearch
    inputs: [bbtmvbot-log-parser]
    endpoints: ["http://openobserve:5080/api/default/"]
    bulk:
      index: bbtmvbot
    auth:
      strategy: "basic"
      user: example@example.com
      password: Example123
    compression: "gzip"
    encoding:
      timestamp_format: "rfc3339"
    healthcheck:
      enabled: false
```

Spin up the updated `docker-compose.yml` and enjoy the OpenObserve. It will take some learning to get used to it, but it allows you to nicely inspect the logs, setup views, dashboards and even alerts.

For example, create and save these views in OpenObserve to have a great overview of what is likelly broken (switch to SQL mode and set appropriate relative time window):

- Generic parsing errors: `SELECT website, msg, error FROM "bbtmvbot" WHERE error is not null`
- Post parsing errors (reasons of why post was not sent to users): `SELECT link, post_errors FROM "bbtmvbot" WHERE post_errors is not null`
- Post parsing warnings (post is sent to users, but without these useful details): `SELECT link, post_warnings FROM "bbtmvbot" WHERE post_warnings is not null`

# Development

Install playwright on your PC, change `Headless:       playwright.Bool(true),` to `Headless:       playwright.Bool(false),` to see stuff in action.

For Docker, install `buildx` module and use following:

```bash
# Build Docker image
docker buildx build --platform=linux/amd64 . -t bbtmvbot:custom --progress=plain --load

# Test Docker image (create /data and config.yml in advance)
docker run --rm --name bbtmvbot -v $(pwd)/config.yml:/config.yml -v $(pwd)/data:/data bbtmvbot:custom

# Run Docker image into /bin/sh
docker run -it --rm --entrypoint /bin/sh bbtmvbot:custom
```
