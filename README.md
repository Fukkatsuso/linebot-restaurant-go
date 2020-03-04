# Restaurant Search App By Golang

## Setup
### Env
```
$ touch line.env

# チャネルID
$ echo "LINE_CHANNEL_ID=hoge" >> line.env

# チャネルシークレット
$ echo "LINE_CHANNEL_SECRET=fuga" >> line.env

# チャネルアクセストークン
$ echo "LINE_CHANNEL_TOKEN=hogefuga" >> line.env
```

## Run and Debug
```
$ docker-compose up
(another tab) $ ngrok http 8080
```

## Deploy
