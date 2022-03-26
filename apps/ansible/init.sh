#!/bin/sh

set -e

cp -r /dockerized/host/home/.ssh /root/.ssh
chmod -R go-rwx /root/.ssh

"$@"
