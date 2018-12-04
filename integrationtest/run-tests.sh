#!/bin/bash

TEST_ID=$(openssl rand -hex 5)
GIT_COMMIT=$(git rev-parse HEAD)
SHORT_COMMIT="${GIT_COMMIT:0:7}"

# Service metadata
APPV_FILE="../appv.json"
SVC_NAME=$(jq '.name' -r $APPV_FILE)
SVC_VERSION=$(jq '.version' -r $APPV_FILE)
SVC_REGISTRY=$(jq '.registry' -r $APPV_FILE)
SVC_IMAGE="$SVC_REGISTRY/$SVC_NAME:$SVC_VERSION"
SVC_CONTAINER_NAME="$SVC_NAME-$SVC_VERSION-$SHORT_COMMIT-$TEST_ID"

echo "Testing $SVC_NAME v: $SVC_VERSION commit: $SHORT_COMMIT. Test ID: $TEST_ID"
echo ""
sleep 1

NETWORK_NAME="$SVC_NAME-network-$SHORT_COMMIT-$TEST_ID"
docker network create $NETWORK_NAME

# Database metadata
DB_IMAGE='postgres:11.1-alpine'
DB_CONTAINER_NAME="directory-db-$TEST_ID"

# Database setup
echo "Starting database: $DB_CONTAINER_NAME"
docker run -d --rm --name $DB_CONTAINER_NAME --net $NETWORK_NAME \
   -e POSTGRES_PASSWORD=password $DB_IMAGE

echo "Sleeping for 2 seconds to make database ready"
sleep 2

echo 'Setup up database and user'
docker exec -i $DB_CONTAINER_NAME psql -U postgres < conf/db_setup.sql

echo 'Database ready'

# Service setup
PASSWORD_SECRETS_FILE='/etc/mimir/directory/password_secrets.json'
TOKEN_SECRETS_FILE='/etc/mimir/token_secrets.json'
SVC_PORT=$(sh ./conf/random_port.sh)

echo "Starting service: $SVC_CONTAINER_NAME on port: $SVC_PORT"
docker run -d --rm --name $SVC_CONTAINER_NAME \
    --network $NETWORK_NAME -p $SVC_PORT:8080 \
    -e PASSWORD_SECRETS_FILE=$PASSWORD_SECRETS_FILE \
    -e TOKEN_SECRETS_FILE=$TOKEN_SECRETS_FILE \
    -e SERVICE_PORT=8080 \
    -e DB_HOST=$DB_CONTAINER_NAME \
    -e DB_PORT=5432 \
    -e DB_NAME=$SVC_NAME \
    -e DB_USERNAME=$SVC_NAME \
    -e DB_PASSWORD='password' \
    -v "$PWD/conf/password_secrets.json":$PASSWORD_SECRETS_FILE:ro \
    -v "$PWD/conf/token_secrets.json":$TOKEN_SECRETS_FILE:ro \
    $SVC_IMAGE

echo "Running tests"
python test.py $SVC_PORT

# Stopping containers
docker stop $SVC_CONTAINER_NAME
docker stop $DB_CONTAINER_NAME
docker network rm $NETWORK_NAME