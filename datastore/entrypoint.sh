#!/bin/bash

set -x

gcloud config set project "$DATASTORE_PROJECT_ID"

gcloud beta emulators datastore env-init

gcloud beta emulators datastore start \
  --data-dir=/datastore/.data \
  --host-port="$DATASTORE_LISTEN_ADDRESS"
