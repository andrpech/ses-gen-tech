FROM golang:1.20-alpine as build-stage

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .
COPY db /db

RUN CGO_ENABLED=0 GOOS=linux go build -o /gses3_btc_application github.com/andrpech/ses-gen-tech/cmd/gses3_btc_application

FROM alpine:latest 

COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /gses3_btc_application /gses3_btc_application
COPY --from=build-stage /db /db

ENTRYPOINT ["/gses3_btc_application"]
