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

# Database metadata
DB_IMAGE='postgres:11.1-alpine'
DB_CONTAINER_NAME="directory-db-$SHORT_COMMIT"
EXTERNAL_DB_PORT=$(sh ./conf/random_port.sh)

echo "Building svc"
cd ../cmd
go build
cd ../integrationtest

# Database setup
echo "Starting database: $DB_CONTAINER_NAME"
docker run -d --rm --name $DB_CONTAINER_NAME --net mimir-net \
   -p $EXTERNAL_DB_PORT:5432 -e POSTGRES_PASSWORD=password $DB_IMAGE

echo 'Sleeping for 2 seconds to allow database to startup'
sleep 2

echo 'Setup up database and user'
docker exec -i $DB_CONTAINER_NAME psql -U postgres < conf/db_setup.sql

echo 'Database ready'

# Service setup
# PASSWORD_SECRETS_FILE='/etc/mimir/directory/password_secrets.json'
# TOKEN_SECRETS_FILE='/etc/mimir/directory/token_secrets.json'

cp ../cmd/cmd $SVC_CONTAINER_NAME
cp -R ../cmd/migrations migrations
cp ./conf/2__test_data.sql ./migrations/2__test_data.sql
SVC_PORT='52711' #$(sh ./conf/random_port.sh)

export PASSWORD_SECRETS_FILE="$PWD/conf/password_secrets.json"
export TOKEN_SECRETS_FILE="$PWD/conf/token_secrets.json"
export SERVICE_PORT=$SVC_PORT
export DB_HOST='127.0.0.1'
export DB_PORT=$EXTERNAL_DB_PORT
export DB_NAME=$SVC_NAME
export DB_USERNAME=$SVC_NAME
export DB_PASSWORD='password'

echo "Starting service: $SVC_CONTAINER_NAME on port: $SVC_PORT"
./$SVC_CONTAINER_NAME
rm $SVC_CONTAINER_NAME
rm -rf migrations/

#echo "Starting service: $SVC_CONTAINER_NAME"
#docker run --rm -t --name $SVC_CONTAINER_NAME \
#    --network mimir-net -p $SVC_PORT:$SVC_PORT \
#    -e PASSWORD_SECRETS_FILE=$PASSWORD_SECRETS_FILE \
#    -e TOKEN_SECRETS_FILE=$TOKEN_SECRETS_FILE \
#    -e SERVICE_PORT=$SVC_PORT \
#    -e DB_HOST=$DB_NAME \
#    -e DB_NAME=$SVC_NAME \
#    -e DB_USERNAME=$SVC_NAME \
#    -e DB_PASSWORD='password' \
#    -v "$PWD/conf/password_secrets.json":$PASSWORD_SECRETS_FILE:ro \
#    -v "$PWD/conf/token_secrets.json":$TOKEN_SECRETS_FILE:ro \
#    $SVC_IMAGE

docker stop $DB_CONTAINER_NAME