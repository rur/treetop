/*eslint-env node, es6 */
const jsdom = require("jsdom");
const sinon = require("sinon");

const DEFAULT_HTML = "<html><head><title>Default Title</title></head><body></body></html>";
global.document = jsdom.jsdom(DEFAULT_HTML);
global.window = document.defaultView;
global.window.requestAnimationFrame = sinon.spy();
global.window.cancelAnimationFrame = sinon.spy();
global.navigator = window.navigator;