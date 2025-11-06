#!/bin/bash

# Usage:
# ./crud.sh <port> <memory> <CRUD> <key> <value>


port_num=$1		# Server port
mem=$2			# C = with cache, D = without cache
CRUD=$3			# any of the crud operation but in small letters
key=$4
val=$5

# Base URL
base_url="http://localhost:${port_num}/${mem}/${CRUD}"

# Perform action
if [[ "$CRUD" == "create" || "$CRUD" == "update" ]]; then
    echo "POST on ${base_url} with key : ${key} and value : ${val}"
    curl -X POST "${base_url}" \
        -H "Content-Type: application/json" \
        -d "{\"key\": ${key}, \"value\": \"${val}\"}"
    echo
elif [[ "$CRUD" == "delete" || "$CRUD" == "read" ]]; then
    echo "GET on ${base_url} with key : ${key}"
    curl "${base_url}?key=${key}"
    echo
fi

