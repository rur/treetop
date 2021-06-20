#!/bin/bash

set -e

echo "Fetching treetop client library "

modified=`date +%s`
content=`curl https://raw.githubusercontent.com/rur/treetop-client/v0.9.0/treetop.js | sed 's|\`|"|g'`


cat << EOF > javascript.go
package internal

// Code generated by go generate; DO NOT EDIT

const Modified = "$modified"

var ScriptContent = \`$content\`

EOF