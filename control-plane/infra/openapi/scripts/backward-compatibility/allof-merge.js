const fs = require('fs');
const yaml = require('js-yaml');
const { merge } = require('allof-merge');

var args = process.argv.slice(2);

if (args.length !== 2) {
    console.error('Usage: node allof-merge.js <input-file.yaml> <output-file.yaml>');
    process.exit(1);
}

const inputFilePath = args[0];
const outputFilePath = args[1];
// Load your OpenAPI YAML file
const openApiDoc = yaml.load(fs.readFileSync(inputFilePath, 'utf8'));

// Merge allOf schemas
const mergedDoc = merge(openApiDoc);

// Save the flattened OpenAPI YAML file
fs.writeFileSync(outputFilePath, yaml.dump(mergedDoc));