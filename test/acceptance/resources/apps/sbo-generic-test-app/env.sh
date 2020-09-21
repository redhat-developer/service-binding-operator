echo "Content-Type: application/json"

env_var=$(echo $PATH_INFO | cut -c2-)
value=$(jq -e ".${env_var}" /tmp/env.json)
if [ $? != 0 ]; then
  echo -e "Status: 404 Not Found\n"
else
  echo -e "\n$value"
fi