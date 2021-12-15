#!/bin/bash

echo "Content-Type: text/plain"

SERVICE_BINDING_ROOT=$(jq -r '.SERVICE_BINDING_ROOT // empty' /tmp/env.json)

[ -z "$SERVICE_BINDING_ROOT" ] && echo -e "Status: 404 SERVICE_BINDING_ROOT Not Found\n" && exit

for f in $(find $SERVICE_BINDING_ROOT -name type -type f); do
  if [ $(cat $f) == "postgresql" ]; then
    BINDING_DIR=$(dirname $f)
    break
  fi
done

[ -z "$BINDING_DIR" ] && echo -e "Status: 404 Postgres bindings Not Found\n" && exit

PORT=5432
[ -r "$BINDING_DIR/port" ] && PORT=$(cat $BINDING_DIR/port)

psql postgresql://$(cat $BINDING_DIR/username):$(cat $BINDING_DIR/password)@$(cat $BINDING_DIR/host):$PORT/$(cat $BINDING_DIR/database) -c '\conninfo' >/tmp/psql 2>&1

if [ $? != 0 ]; then
  echo -e "Status: 500 cannot connect\n\n"
  cat /tmp/psql
else
  echo -e "\nOK"
fi