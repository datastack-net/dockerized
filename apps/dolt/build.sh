DOLT_VERSION=$1
apk add --no-cache bash curl
wget -O - https://github.com/dolthub/dolt/releases/${DOLT_VERSION}/download/install.sh | bash