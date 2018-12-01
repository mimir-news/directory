#!/bin/bash

GIT_COMMIT=$(git rev-parse HEAD)
SHORT_COMMIT="${GIT_COMMIT:0:7}"

# Service metadata
APPV_FILE="../appv.json"
SVC_NAME=$(jq '.name' -r $APPV_FILE)
SVC_VERSION=$(jq '.version' -r $APPV_FILE)
SVC_REGISTRY=$(jq '.registry' -r $APPV_FILE)
SVC_IMAGE="$SVC_REGISTRY/$SVC_NAME:$SVC_VERSION"
SVC_CONTAINER_NAME="$SVC_NAME-$SVC_VERSION-$SHORT_COMMIT"
SVC_PORT=$(sh ./conf/random_port.sh)

# Database metadata
DB_IMAGE='postgres:11.1-alpine'
DB_NAME="directory-db-$SHORT_COMMIT"

# Database setup
echo "Starting database: $DB_NAME"
docker run -d --rm --name $DB_NAME --net mimir-net \
   -e POSTGRES_PASSWORD=password $DB_IMAGE

echo 'Sleeping for 2 seconds to allow database to startup'
sleep 2

echo 'Setup up database and user'
docker exec -i $DB_NAME psql -U postgres < conf/db_setup.sql

echo 'Database ready'

# Service setup
PASSWORD_SECRETS_FILE='/etc/mimir/directory/password_secrets.json'
TOKEN_SECRETS_FILE='/etc/mimir/directory/token_secrets.json'

echo "Starting service: $SVC_CONTAINER_NAME"
docker run -d --name $SVC_CONTAINER_NAME \
    --network mimir-net -p $SVC_PORT:$SVC_PORT \
    -e PASSWORD_SECRETS_FILE=$PASSWORD_SECRETS_FILE \
    -e TOKEN_SECRETS_FILE=$TOKEN_SECRETS_FILE \
    -e SERVICE_PORT=$SVC_PORT \
    -e DB_HOST=$DB_NAME\
    -e DB_DATABASE=$SVC_NAME \
    -e DB_USERNAME=$SVC_NAME \
    -e DB_PASSWORD='password' \
    -v "$PWD/conf/password_secrets.json":$PASSWORD_SECRETS_FILE:ro \
    -v "$PWD/conf/token_secrets.json":$TOKEN_SECRETS_FILE:ro \
    $SVC_IMAGE
