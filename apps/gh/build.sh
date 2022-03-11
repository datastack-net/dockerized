GH_VERSION=$1
mkdir /gh
cd /gh
wget https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_386.tar.gz -O ghcli.tar.gz
tar --strip-components=1 -xf ghcli.tar.gz
ln -s /gh/bin/gh /usr/local/bin/gh
