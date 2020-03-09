FROM google/cloud-sdk:latest

ENV GO_VERSION 1.11

RUN curl -Lso go.tar.gz "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH /usr/local/go/bin:$PATH

ENV GOPATH /go

WORKDIR /go/src/linebot-restaurant-go/go-app

RUN go get github.com/line/line-bot-sdk-go/linebot
