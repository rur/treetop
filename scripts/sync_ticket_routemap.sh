#!/bin/sh

set -ex

tmpdir=`ttpage ./examples/ticket/routemap.toml`

cp "$tmpdir/page/ticket/routemap.toml" ./examples/ticket/routemap.toml
cp "$tmpdir/page/ticket/handlers.go" ./examples/ticket/handlers.go
rsync -r "$tmpdir/page/ticket/templates" ./examples/ticket/templates
go generate ./examples/ticket