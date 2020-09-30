FROM google/cloud-sdk:alpine

RUN apk add --update --no-cache openjdk8-jre \
    && gcloud components install cloud-datastore-emulator beta --quiet

WORKDIR /datastore
