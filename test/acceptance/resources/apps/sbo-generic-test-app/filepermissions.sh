#!/bin/bash

echo "Content-Type: text/plain"

[ -z "$SERVICE_BINDING_ROOT" ] && SERVICE_BINDING_ROOT=$(jq -r '.SERVICE_BINDING_ROOT // empty' /tmp/env.json)
[ -z "$SERVICE_BINDING_ROOT" ] && echo -e "Status: 404 SERVICE_BINDING_ROOT Not Found\n" && exit

filename=$(echo $PATH_INFO | cut -c2-)
for f in $(find $SERVICE_BINDING_ROOT -name $filename -type f); do
  BINDING_DIR=$(dirname $f)
  break
done

if [ ! -f $BINDING_DIR/$filename ]; then
  echo -e "Status: 500 file not found $filename\n\n"
  exit
fi

ls -lL $BINDING_DIR/$filename | cut -d " " -f 1 | grep "w"

if [ $? == 0 ]; then
  echo -e "Status: 500 file has wrong permission\n\n"
  ls -lL $BINDING_DIR/$filename
else
  echo -e "\nOK"
fi
