#!/bin/sh

set -ex

tmpdir=`ttpage ./demo/ticket/routemap.toml`

cp "$tmpdir/page/ticket/routemap.toml" ./demo/ticket/routemap.toml
cp "$tmpdir/page/ticket/handlers.go" ./demo/ticket/handlers.go
rsync -r "$tmpdir/page/ticket/templates/" ./demo/ticket/templates/
go generate ./demo/ticket