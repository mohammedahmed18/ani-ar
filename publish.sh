#!/usr/bin/bash

go build -o temp/ani-ar ./cmd
source ./dropbox.token

curl -X POST https://content.dropboxapi.com/2/files/upload \
 --header "Authorization: Bearer $DROPBOX_API_TOKEN" \
 --header "Dropbox-API-Arg: {\"path\": \"/ani-ar/ani-ar\",\"mode\": \"overwrite\",\"autorename\": true,\"mute\": false}" \
 --header "Content-Type: application/octet-stream" \
 --data-binary @temp/ani-ar

echo ""
echo "removing the generated build"
rm ./temp/ani-ar