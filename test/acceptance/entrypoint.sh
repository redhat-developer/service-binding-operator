#!/bin/bash -xe
if [ $1 == "generate" ]; then
    allure generate /allure/results -o /allure/report --clean
elif [ $1 == "serve" ]; then
    allure serve -p 8080 /allure/results
else
    echo "Usage: $0 (generate|serve)"
fi