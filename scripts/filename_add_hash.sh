#!/bin/sh

# This script will hash the content of a file and add it to the filename.
# Example: "filename_add_hash.sh ./src/css/main.css"
# The example will (MD5) hash the contents of "main.css" and
# move "main.css" to "main.<md5_hash>.css"

if [ $# -eq 0 ]
then
  echo "No arguments supplied. Correct syntax example: 'filename_add_hash.sh ./src/css/main.css'"
  exit 1
fi

if ! command -v md5 &> /dev/null
then
    echo "fatal error: 'md5' binary not found in PATH"
    exit 1
fi

FILE=$1

if [ ! -f "$FILE" ]
then
  echo "$FILE: No such file"
  exit 1
fi

FILE_PARENT_DIR=$(dirname "$FILE")
FILENAME=$(basename -- "$FILE")
EXTENSION="${FILENAME##*.}"
FILENAME="${FILENAME%.*}"
MD5=$(md5 -q "$FILE")

mv "$FILE" "$FILE_PARENT_DIR/$FILENAME.$MD5.$EXTENSION"
