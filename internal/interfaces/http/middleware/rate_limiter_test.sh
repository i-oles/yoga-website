#!/bin/bash

URL="http://localhost:8080/pending_bookings"

for i in {1..30}; do
    EMAIL="test$(openssl rand -hex 4)@gmail.com"

    echo "Request #$i -> $EMAIL"

    curl -s -o /dev/null -w "HTTP %{http_code}\n" \
        --location "$URL" \
        --form "class_id=7badf60a-7225-4904-b8b9-186686299b48" \
        --form "first_name=igor" \
        --form "last_name=oles" \
        --form "email=$EMAIL"

    sleep 0.2
done
