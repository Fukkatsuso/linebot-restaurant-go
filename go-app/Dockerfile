# base
FROM golang:1.14.4-alpine3.12 AS base

RUN apk --no-cache add \
        git 

WORKDIR $GOPATH/src/github.com/Fukkatsuso/linebot-restaurant-go/go-app

RUN go get -u -v \
        github.com/line/line-bot-sdk-go/linebot

# dev
FROM base AS dev

RUN go get -v \
        github.com/oxequa/realize