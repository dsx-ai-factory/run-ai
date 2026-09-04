"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.escapeUnicode = exports.escapeClosingScriptTag = exports.sanitizeJSONString = exports.getPageHTML = exports.getObjectOrJSON = void 0;
const react_1 = require("react");
const redoc_1 = require("redoc");
const openapi_core_1 = require("@redocly/openapi-core");
const server_1 = require("react-dom/server");
const styled_components_1 = require("styled-components");
const handlebars_1 = require("handlebars");
const path_1 = require("path");
const fs_1 = require("fs");
const colorette_1 = require("colorette");
const utils_1 = require("../../utils");
function getObjectOrJSON(openapiOptions, config) {
    switch (typeof openapiOptions) {
        case 'object':
            return openapiOptions;
        case 'string':
            try {
                if (fs_1.existsSync(openapiOptions) && fs_1.lstatSync(openapiOptions).isFile()) {
                    return JSON.parse(fs_1.readFileSync(openapiOptions, 'utf-8'));
                }
                else {
                    return JSON.parse(openapiOptions);
                }
            }
            catch (e) {
                process.stderr.write(colorette_1.red(`Encountered error:\n\n${openapiOptions}\n\nis neither a file with a valid JSON object neither a stringified JSON object.`));
                utils_1.exitWithError(e);
            }
            break;
        default: {
            if (config) {
                process.stderr.write(`Found ${config.configFile} and using theme.openapi options`);
                return config.theme.openapi ? config.theme.openapi : {};
            }
            return {};
        }
    }
    return {};
}
exports.getObjectOrJSON = getObjectOrJSON;
function getPageHTML(api, pathToApi, { cdn, title, disableGoogleFont, templateFileName, templateOptions, redocOptions = {}, redocCurrentVersion, }) {
    return __awaiter(this, void 0, void 0, function* () {
        process.stderr.write('Prerendering docs');
        const apiUrl = redocOptions.specUrl || (openapi_core_1.isAbsoluteUrl(pathToApi) ? pathToApi : undefined);
        const store = yield redoc_1.createStore(api, apiUrl, redocOptions);
        const sheet = new styled_components_1.ServerStyleSheet();
        const html = server_1.renderToString(sheet.collectStyles(react_1.createElement(redoc_1.Redoc, { store })));
        const state = yield store.toJS();
        const css = sheet.getStyleTags();
        templateFileName = templateFileName ? templateFileName : path_1.join(__dirname, './template.hbs');
        const template = handlebars_1.compile(fs_1.readFileSync(templateFileName).toString());
        return template({
            redocHTML: `
      <div id="redoc">${html || ''}</div>
      <script>
      ${`const __redoc_state = ${sanitizeJSONString(JSON.stringify(state))};` || ''}

      var container = document.getElementById('redoc');
      Redoc.${'hydrate(__redoc_state, container)'};

      </script>`,
            redocHead: (cdn
                ? '<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>'
                : `<script src="https://cdn.redoc.ly/redoc/v${redocCurrentVersion}/bundles/redoc.standalone.js"></script>`) +
                css,
            title: title || api.info.title || 'ReDoc documentation',
            disableGoogleFont,
            templateOptions,
        });
    });
}
exports.getPageHTML = getPageHTML;
function sanitizeJSONString(str) {
    return escapeClosingScriptTag(escapeUnicode(str));
}
exports.sanitizeJSONString = sanitizeJSONString;
// see http://www.thespanner.co.uk/2011/07/25/the-json-specification-is-now-wrong/
function escapeClosingScriptTag(str) {
    return str.replace(/<\/script>/g, '<\\/script>');
}
exports.escapeClosingScriptTag = escapeClosingScriptTag;
// see http://www.thespanner.co.uk/2011/07/25/the-json-specification-is-now-wrong/
function escapeUnicode(str) {
    return str.replace(/\u2028|\u2029/g, (m) => '\\u202' + (m === '\u2028' ? '8' : '9'));
}
exports.escapeUnicode = escapeUnicode;
