#!/bin/bash

echo "Content-Type: text/plain"

[ -z "$SERVICE_BINDING_ROOT" ] && SERVICE_BINDING_ROOT=$(jq -r '.SERVICE_BINDING_ROOT // empty' /tmp/env.json)

[ -z "$SERVICE_BINDING_ROOT" ] && echo -e "Status: 404 SERVICE_BINDING_ROOT Not Found\n" && exit

for f in $(find $SERVICE_BINDING_ROOT -name type -type f); do
  if [ $(cat $f) == "mysql" ]; then
    BINDING_DIR=$(dirname $f)
    break
  fi
done

[ -z "$BINDING_DIR" ] && echo -e "Status: 404 MySQL bindings Not Found\n" && exit

PORT=3306
[ -r "$BINDING_DIR/port" ] && PORT=$(cat $BINDING_DIR/port)

mysql -h $(cat $BINDING_DIR/host) -P $PORT -u$(cat $BINDING_DIR/username) -p$(cat $BINDING_DIR/password) -e 'use '$(cat $BINDING_DIR/database) >/tmp/mysql 2>&1

if [ $? != 0 ]; then
  echo -e "Status: 500 cannot connect\n\n"
  cat /tmp/mysql
else
  echo -e "\nOK"
fi