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

## Deploy to GAE
### Cloud Shell上での準備
1. プロジェクト, GAEアプリの作成
```sh
export PROJECT_ID=blog-XXXXXX
gcloud projects create --name ${PROJECT_ID}
gcloud config set project ${PROJECT_ID}
gcloud app create
```

2. APIを有効化(Cloud Buildのために課金を有効にする)
```sh
gcloud services enable appengine.googleapis.com

gcloud alpha billing accounts list
gcloud alpha billing projects link ${PROJECT_ID} --billing-account YYYYYY-ZZZZZZ-AAAAAA
gcloud services enable cloudbilling.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

3. サービスアカウント, サービスアカウントキーの作成
```sh
export SA_NAME=githubactions
gcloud iam service-accounts create ${SA_NAME} \
  --description="used by GitHub Actions" \
  --display-name="${SA_NAME}"
gcloud iam service-accounts list

export IAM_ACCOUNT=${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com

gcloud iam service-accounts keys create ~/${PROJECT_ID}/${SA_NAME}/key.json \
  --iam-account ${IAM_ACCOUNT}
```

4. role付与
```sh
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${IAM_ACCOUNT}" \
  --role='roles/compute.storageAdmin'
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${IAM_ACCOUNT}" \
  --role='roles/cloudbuild.builds.editor'
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${IAM_ACCOUNT}" \
  --role='roles/appengine.deployer'
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${IAM_ACCOUNT}" \
  --role='roles/appengine.appAdmin'
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${IAM_ACCOUNT}" \
  --role='roles/cloudbuild.builds.builder'
```

### GitHub Secret
- LINE_CHANNEL_ID
- LINE_CHANNEL_SECRET
- LINE_CHANNEL_TOKEN
- GCP_PLACES_API_KEY
- GCP_PROJECT: プロジェクトID
- GCP_SA_KEY: サービスアカウントのJSON鍵をBase64エンコード
  ```sh
  # Cloud Shell
  openssl base64 -in ~/${PROJECT_ID}/${SA_NAME}/key.json
  ```

### GitHubへPush
masterブランチへ
