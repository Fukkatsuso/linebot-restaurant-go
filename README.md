# Restaurant Search App By Golang
[![Deploy](https://github.com/Fukkatsuso/linebot-restaurant-go/workflows/Deploy/badge.svg)](https://github.com/Fukkatsuso/linebot-restaurant-go/actions?query=workflow%3ADeploy)

## Setup
### Env
```sh
cat <<EOS >> ./go-app/secret.env
LINE_CHANNEL_ID=hoge
LINE_CHANNEL_SECRET=fuga
LINE_CHANNEL_TOKEN=hogefuga
GCP_PLACES_API_KEY=AAAAA
DATASTORE_PROJECT_ID=restaurant-search-XXXXXX
EOS

cat <<EOS >> ./datastore/secret.env
DATASTORE_PROJECT_ID=restaurant-search-XXXXXX
EOS
```


## Run and Debug
```sh
docker-compose up
(another tab) ngrok http 8080
```

## Deploy to Cloud Run
### Cloud Shell上での準備
1. プロジェクトの作成
```sh
export PROJECT_ID=restaurant-search-XXXXXX
export REGION=asia-northeast1
gcloud projects create --name ${PROJECT_ID}
gcloud config set project ${PROJECT_ID}
gcloud config set run/region ${REGION}
```

2. APIを有効化(課金も有効にする)
```sh
gcloud services enable run.googleapis.com

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

4. role付与 - [参考](https://cloud.google.com/run/docs/reference/iam/roles?hl=ja#additional-configuration)
```sh
gcloud projects add-iam-policy-binding ${PROJECT_ID} --member="serviceAccount:${IAM_ACCOUNT}" \
  --role="roles/run.admin"

export PROJECT_NUMBER=XXXXXXXXXXXX
gcloud iam service-accounts add-iam-policy-binding ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com --member="serviceAccount:${IAM_ACCOUNT}" \
  --role="roles/iam.serviceAccountUser"
```

### GitHub Secret
- LINE_CHANNEL_ID
- LINE_CHANNEL_SECRET
- LINE_CHANNEL_TOKEN
- GCP_PLACES_API_KEY
- GCP_PROJECT: プロジェクトID
- GCP_REGION: Cloud Runのリージョン
- GCP_SA_KEY: サービスアカウントのJSON鍵をBase64エンコード
  ```sh
  # Cloud Shell
  openssl base64 -in ~/${PROJECT_ID}/${SA_NAME}/key.json
  ```

### GitHubへPush
- masterブランチへ
- Google App Engineを有効化していないとDatastoreが使用できないので注意
