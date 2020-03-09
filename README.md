# Restaurant Search App By Golang

## Setup
### Env
```sh
$ touch ./go-app/line.env

# LINE チャネルID
$ echo "LINE_CHANNEL_ID=hoge" >> ./go-app/line.env

# LINE チャネルシークレット
$ echo "LINE_CHANNEL_SECRET=fuga" >> ./go-app/line.env

# LINE チャネルアクセストークン
$ echo "LINE_CHANNEL_TOKEN=hogefuga" >> ./go-app/line.env

$ touch ./go-app/gcp.env

# GCP Places API キー
$ echo "GCP_PLACES_API_KEY=AAAAA" >> ./go-app/gcp.env
```


## Run and Debug
```sh
$ cd go-app/
/go-app $ docker-compose up
(another tab) $ ngrok http 8080
```


## Deploy
```sh
$ cat go-app/secret.yaml
env_variables:
  LINE_CHANNEL_ID: "hoge"
  LINE_CHANNEL_SECRET: "fuga"
  LINE_CHANNEL_TOKEN: "hogefuga"
  GCP_PLACES_API_KEY: "AAAAA"
```