#!/bin/sh

if [ ! -f ~/.ssh/dockerized ]; then
    ssh-keygen -t rsa -b 4096 -C "dockerized@${HOST_HOSTNAME}" -f ~/.ssh/dockerized -N ""
fi

"$@"
