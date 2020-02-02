const fs = require('fs');

fs.readFile('messages.json', 'utf-8', (err, json) => {
  const obj = JSON.parse(json);
  fs.writeFile('messages.json', JSON.stringify(obj, null, ' '), err => {
    if (err) {
      throw err;
    }
  });
});
