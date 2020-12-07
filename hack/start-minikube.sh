#!/usr/bin/env bash

minikube start --addons=registry --insecure-registry=0.0.0.0/0 "$@"
