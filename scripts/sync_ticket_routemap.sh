#!/bin/sh

ttpage ./examples/ticket/routemap.toml | xargs -I '{}' rsync -r "{}/page/ticket/" ./examples/ticket/