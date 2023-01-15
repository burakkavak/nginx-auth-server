#!/bin/bash

# This script is for development purposes only. It pushes binaries and 'docker_run.sh' to a remote server
# for Docker testing purposes.

if [ $# -eq 0 ]
  then
    echo "No arguments supplied. Correct syntax example: 'deploy_files.sh root@example.org /var/www/html'"
    exit 1
fi


REMOTE=$1 # e.g. root@example.org
REMOTE_PATH=$2 # e.g. /var/www/html
FILES=("bin/nginx-auth-server-linux-amd64.tar.gz" "scripts/docker_run.sh") # add files to the array as needed
SCRIPT_PATH=$(readlink -f "${BASH_SOURCE:-$0}")
SCRIPT_DIR=$(dirname "$SCRIPT_PATH")

for FILE in "${FILES[@]}"
do
  rsync --mkpath --progress "$SCRIPT_DIR/../$FILE" "$REMOTE:$REMOTE_PATH/$FILE"
done
