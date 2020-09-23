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

# builder
FROM base AS builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o /go/bin/server

# release
FROM alpine:latest AS release

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /go/bin/server .

CMD [ "./server" ]
