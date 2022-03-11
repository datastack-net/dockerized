#!/bin/sh

if [ ! -f ~/.ssh/id_rsa ]; then
    ssh-keygen -t rsa -b 4096 -C "dockerized@${HOST_HOSTNAME}" -f ~/.ssh/id_rsa -N ""
fi

"$@"
