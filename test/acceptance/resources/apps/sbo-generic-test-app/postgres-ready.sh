#!/bin/bash

urlencode() {
    # urlencode <string>
    local length="${#1}"
    for (( i = 0; i < length; i++ )); do
        local c="${1:i:1}"
        case $c in
            [a-zA-Z0-9.~_-]) printf "$c" ;;
            *) printf '%%%02X' "'$c" ;;
        esac
    done
}

echo "Content-Type: text/plain"

[ -z "$SERVICE_BINDING_ROOT" ] && SERVICE_BINDING_ROOT=$(jq -r '.SERVICE_BINDING_ROOT // empty' /tmp/env.json)

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

ENC_PASSWD=$(urlencode "$(cat $BINDING_DIR/password)")

psql postgresql://$(cat $BINDING_DIR/username):${ENC_PASSWD}@$(cat $BINDING_DIR/host):$PORT/$(cat $BINDING_DIR/database) -c '\conninfo' >/tmp/psql 2>&1

if [ $? != 0 ]; then
  echo -e "Status: 500 cannot connect\n\n"
  cat /tmp/psql
else
  echo -e "\nOK"
fi