#!/usr/bin/env node

// This script will clear the given directory.
// Example: "node clear_directory.js ./src/css/"

const fs = require('fs');
const path = require('path');

// get dir path from command arguments
const dirPath = process.argv[2];

if (!dirPath || dirPath === '') {
  console.error("fatal error: no directory path provided. Correct syntax example: 'node clear_directory.js ./src/css/'");
  process.exit(1);
}

if (fs.lstatSync(dirPath).isDirectory()) {
  fs.readdir(dirPath, (err, files) => {
    if (err) throw err;

    for (const filePath of files) {
      // delete all files that do not start with '.'
      if (filePath.charAt(0) !== '.') {
        fs.unlink(path.join(dirPath, filePath), (unlinkErr) => {
          if (unlinkErr) throw unlinkErr;
        });
      }
    }
  });
} else {
  console.error(`fatal error: directory at path '${dirPath} does not exist or is not a directory'`);
}
