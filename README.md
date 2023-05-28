# Welcome to Kenobi exchange

## Overview

With this adorable Golang currency exchange application you can:

- get current BTCUAH rate (Hryvnia amount you need to buy a single Bitcoin);
- subscribe to our mail to receive an exchange rate every single time corresponding endpoint will be triggered;
- trigger endpoint and send current BTCUAH rate to all subscribed folks (a lot of fun).

#### In the nearest future we will add some new features amoung wich

- even more currencies: you'll be able to choose several currencies to subscribe on;
- custom mail notifications: set up timing that fits your taste;
- rate history by time: choose time interval and observe the dynamics;
- some tests;
- more to come...

## Tech part

### App modes

There are two and a half modes you can run this beautiful app:

- clone the repo and run it locally with:
  - binary
  - docker;
- pull the docker image and run it wherever its need.

**Env variables:**

- mode:
  - binary: create `.env` file in the root directory;
  - docker: pass env variables with `docker run` command;
- variables:
  - `EMAIL_SENDER_NAME` - sender name;
  - `EMAIL_SENDER_ADDRESS` - sender email;
  - `EMAIL_SENDER_PASSWORD` - app password generated in mail provider settings.

#### Clone repo mode

1. clone the [repo](https://github.com/andrpech/ses-gen-tech) with:

```bash
git clone https://github.com/andrpech/ses-gen-tech
```

2. run the app:
   a. run it locally

   1. navigate inside the directory

   ```bash
   cd ses-gen-tech
   ```

   2. run the entry point

   ```bash
   go run cmd/gses3_btc_application/main.go
   ```

   b. run it in docker container

   1. build docker image

   ```bash
   docker build -t gses3_btc_application .
   ```

   2. check image built

   ```bash
   docker image ls
   ```

   3. run docker container

   ```bash
   docker run -p 8080:8080 -e EMAIL_SENDER_NAME={sender_name} -e EMAIL_SENDER_ADDRESS={email} -e EMAIL_SENDER_PASSWORD={email_app_password} gses3_btc_application
   ```

3. use Postman or curl to interact with app

#### Pull an image

1. clone the [image](https://hub.docker.com/r/andrpech/gses3-btc-application) from Docker Hub:

```bash
docker pull andrpech/gses3-btc-application
```

2. run docker container

   ```bash
   docker run -p 8080:8080 -e EMAIL_SENDER_NAME={sender_name} -e EMAIL_SENDER_ADDRESS={email} -e EMAIL_SENDER_PASSWORD={email_app_password} gses3_btc_application
   ```

3. use Postman or curl to interact with app

### API

According to the [tech task](https://github.com/AndriiPopovych/gses/blob/main/gses2swagger.yaml), at the moment application supports 3 needed routes and the healthcheck:

- `GET /api/rate`: calls [Binance](https://www.binance.com/en/trade/BTC_UAH) api and returns current BTCUAH rate as `float64`;
- `POST /api/subscribe`:
  - consumes `email` param as form data with `application/x-www-form-urlencoded` type;
  - checks:
    - if the email param exists;
    - if the email param is only in payload;
    - if the email already subscribed;
  - if email is already subscribed - returns corresponding answer with subscription time;
  - if email is not subscribed - saves with current time to `db/emails.json` and returns corresponding answer;
- `POST /api/sendEmail`:
  - retrieves current rate;
  - iterates over subscribed emails and sends the rate for email separately;
  - after all returns an array of emails received an update.
- `GET /api/kenobi`:
  - simple healthcheck for docker container;
  - returns 200;

### Project style

Used [standard](https://github.com/golang-standards/project-layout) Golang project structure:

```
├── cmd/
│   └── gses3_btc_application/
│       └── main.go                # entry point
├── db/
│   └── emails.json                # emails storage
├── internal/
│   └── app/
│       ├── endpoints/
│       │   ├── assets/            # static files
│       │   │   └── email_template.html # email template that will be embed during build time
│       │   └── endpoint.go        # application endpoints
│       ├── middleware/
│       │   └── parse_form_data.go # middleware for parsing form data
│       ├── service/               # application service
│       │   ├── fs/
│       │   │   └── json.go        # json io operations service
│       │   ├── mail/
│       │   │   └── sender.go      # mail service
│       │   ├── rate/
│       │   │   └── binance.go     # binance api service
│       │   └── service.go         # service interface
│       └── pkg/
│           └── app.go             # application interface
├── tools/
│   └── check_email.go             # simple script for checking email
└── util/
    └── config.go                  # config struct
```
