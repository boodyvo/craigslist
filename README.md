# Description 

Monitor SF pages from craigslist for cars and trucks from owners only
and push them into MySQL as fast, as they are discovered.  

## Build

[Docker](https://docs.docker.com/engine/install/) and [docker-compose](https://docs.docker.com/compose/install/) should be installed.
It will build the env with the image.

To run all services (with db) run:

```bash
docker-compose build && docker-compose up -d
```

To run only the scrapper run:

```bash
docker-compose build app && docker-compose up -d app
```

## Setting up own db

To set custom MySQL connection set up env variables in docker-compose.yml: 

- `CRAIGSLIST_DB_USER` - MySQL user with permission to write into the `table`, default is `root`;
- `CRAIGSLIST_DB_PASSWORD` - MySQL user's password, default is `somepassword`;
- `CRAIGSLIST_DB_HOST` - MySQL host, default is `db`;
- `CRAIGSLIST_DB_DATABASE` - MySQL database that contains the `table`, default is `test`;
- `CRAIGSLIST_DB_TABLE` - MySQL table, default is `urls`;
- `CRAIGSLIST_DB_FIELD_NAME` - MySQL field name in the `table`, default is `url`;
- `CRAIGSLIST_TELEGRAM_BOT_TOKEN` - a token for telegram bot. Should be set in the docker env;
- `CRAIGSLIST_TELEGRAM_CHANNEL` - list of channels to push found messages to in format "@chan1,@chan2";
- `CRAIGSLIST_START_TIME` - UTC time to start notifications in telegram, default is "02:00PM";
- `CRAIGSLIST_STOP_TIME` - UTC time to stop notifications in telegram, default is "07:00AM";