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
### Make 'secret.yaml'
```sh
$ cat go-app/secret.yaml
env_variables:
  LINE_CHANNEL_ID: "hoge"
  LINE_CHANNEL_SECRET: "fuga"
  LINE_CHANNEL_TOKEN: "hogefuga"
  GCP_PLACES_API_KEY: "AAAAA"
```

### Login gcloud and Deploy
```sh
# login gcloud
/go-app $ docker-compose up
(another tab) /go-app $ docker container exec -it linebot-restaurant-go bash
(inside the container) $ gcloud auth login


# gcloud app information
(inside the container) $ gcloud app describe


# deploy first time
(inside the container) $ gcloud init
(inside the container) $ gcloud app create --project=*your project ID*
...
Please enter your numeric choice:  3 (asia-northeast2)
...
(inside the container) $ gcloud app deploy


# deploy after the first time
(inside the container) $ gcloud config set project *your project ID*
(inside the container) $ gcloud app deploy
```
