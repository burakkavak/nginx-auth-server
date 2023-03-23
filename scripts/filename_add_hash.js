#!/usr/bin/env node

// This script will hash the content of a file and add it to the filename.
// Example: "node filename_add_hash.js ./src/css/main.css"
// The example will (MD5) hash the contents of "main.css" and
// move "main.css" to "main.<md5_hash>.css"

const fs = require('fs');
const crypto = require('crypto');

// get file path from command arguments
const filePath = process.argv[2];

if (!filePath || filePath === '') {
  console.error("fatal error: no file path provided. Correct syntax example: 'node filename_add_hash.js ./src/css/main.css'");
  process.exit(1);
}

const fileExtension = filePath.split('.').pop();
const filePathWithoutExtension = filePath.substring(0, filePath.length - (fileExtension.length + 1));

if (fs.existsSync(filePath)) {
  const buffer = fs.readFileSync(filePath);
  const hash = crypto.createHash('md5').update(buffer).digest('hex');

  const newFilePath = `${filePathWithoutExtension}.${hash}.${fileExtension}`;

  fs.rename(filePath, newFilePath, (err) => {
    if (err) console.error(`fatal error: could not rename file${err}`);
  });
} else {
  console.error(`fatal error: file at path '${filePath} does not exist or is not accessible'`);
}
