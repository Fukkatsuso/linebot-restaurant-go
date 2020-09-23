# base
FROM golang:1.14.4-alpine3.12 AS base

RUN apk --no-cache add \
        git 

WORKDIR /go/src/github.com/Fukkatsuso/linebot-restaurant-go/go-app

COPY go.mod .

RUN go mod download

# dev
FROM base AS dev

RUN go get \
        github.com/pilu/fresh

CMD [ "fresh" ]
