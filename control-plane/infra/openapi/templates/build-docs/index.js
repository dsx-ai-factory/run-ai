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
exports.handlerBuildCommand = void 0;
const redoc_1 = require("redoc");
const path_1 = require("path");
const fs_1 = require("fs");
const perf_hooks_1 = require("perf_hooks");
const utils_1 = require("./utils");
const openapi_core_1 = require("@redocly/openapi-core");
const utils_2 = require("../../utils");
const handlerBuildCommand = (argv) => __awaiter(void 0, void 0, void 0, function* () {
    var _a;
    const startedAt = perf_hooks_1.performance.now();
    const configFromFile = yield utils_2.loadConfigAndHandleErrors({ configPath: argv.config });
    const config = openapi_core_1.getMergedConfig(configFromFile, argv.api);
    const apis = yield utils_2.getFallbackApisOrExit(argv.api ? [argv.api] : [], config);
    const { path: pathToApi } = apis[0];
    const options = {
        output: argv.o,
        cdn: argv.cdn,
        title: argv.title,
        disableGoogleFont: argv.disableGoogleFont,
        templateFileName: argv.template,
        templateOptions: argv.templateOptions || {},
        redocOptions: utils_1.getObjectOrJSON((_a = argv.theme) === null || _a === void 0 ? void 0 : _a.openapi, config),
    };
    const redocCurrentVersion = require('../../../package.json').dependencies.redoc.substring(1); // remove ~
    try {
        const elapsed = utils_2.getExecutionTime(startedAt);
        const api = yield redoc_1.loadAndBundleSpec(openapi_core_1.isAbsoluteUrl(pathToApi)
            ? pathToApi
            : path_1.resolve(argv.config ? path_1.dirname(argv.config) : '', pathToApi));
        const pageHTML = yield utils_1.getPageHTML(api, pathToApi, Object.assign(Object.assign({}, options), { redocCurrentVersion }));
        fs_1.mkdirSync(path_1.dirname(options.output), { recursive: true });
        fs_1.writeFileSync(options.output, pageHTML);
        const sizeInKiB = Math.ceil(Buffer.byteLength(pageHTML) / 1024);
        process.stdout.write(`\n🎉 bundled successfully in: ${options.output} (${sizeInKiB} KiB) [⏱ ${elapsed}].\n`);
    }
    catch (e) {
        utils_2.exitWithError(e);
    }
});
exports.handlerBuildCommand = handlerBuildCommand;
