#!/bin/sh

# handle scrapy commands
scrapy_commands="$(scrapy | awk '/Available commands/,/^$/ { if (NF && $1 != "Available") { print $1 }}')"
if (echo $scrapy_commands | grep -qw "$1"); then
    set -- scrapy $@
fi

exec $@