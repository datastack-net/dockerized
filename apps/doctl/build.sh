DOCTL_VERSION=$1
cd ~
wget https://github.com/digitalocean/doctl/releases/download/v${DOCTL_VERSION}/doctl-${DOCTL_VERSION}-linux-amd64.tar.gz
tar xf ~/doctl-${DOCTL_VERSION}-linux-amd64.tar.gz
mv ~/doctl /usr/local/bin
