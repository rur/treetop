/* global window, document, XMLHttpRequest, ActiveXObject */

window.treetop = (function ($, config) {
    "use strict";
    var onLoad = $.simpleSignal();

    /**
     * Treetop API Constructor
     *
     * @constructor
     * @param {Array|Treetop} setup GA style initialization
     */
    function Treetop(setup) {
        if (setup instanceof Treetop) {
            this._setup = setup._setup;
        } else if (setup instanceof Array) {
            this._setup = setup;
        }
        this._setup = (this._setup || []).slice();
        $.bindComponentsAsync(this._setup);
    }

    // setup inheritance for extensions
    function TreetopProto() {}
    TreetopProto.prototype = $.extensions;
    Treetop.prototype = new TreetopProto();

    /**
     * Add a component definition
     * @param  {Object} def Dict containing component
     */
    Treetop.prototype.push = function (def) {
        if (def) this._setup.push(def);
        $.bindComponentsAsync(this._setup);
    };

    /**
     * trigger mount event on a node and it's subtree
     * @param  {HTMLElement} el
     */
    Treetop.prototype.mount = function (el) {
        $.mount(el);
    };

    /**
     * trigger mount event on a node and it's subtree
     * @param  {HTMLElement} el
     */
    Treetop.prototype.unmount = function (el) {
        $.unmount(el);
    };

    /**
     * Send XHR request to Treetop endpoint. The response is handled
     * internally, the response is handled internally.
     *
     * @public
     * @param  {string} method The request method GET|POST|...
     * @param  {string} url    The url
     */
    Treetop.prototype.request = function (method, url, data, encoding) {
        if ($.METHODS[method.toUpperCase()]) {
            throw new Error("Treetop: Unknown request method '" + method + "'");
        }
        var req = (XMLHttpRequest) ? new XMLHttpRequest() : new ActiveXObject("MSXML2.XMLHTTP");
        req.open(method.toUpperCase(), url, true);
        req.setRequestHeader("Accept", $.HJ_CONTENT_TYPE);
        if (data) {
            req.setRequestHeader("Content-Type", encoding || "application/x-www-form-urlencoded");
        }
        req.onload = function () {
            $.ajaxSuccess(req);
            onLoad.trigger();
        };
        req.send(data || null);
    };

    Treetop.prototype.onLoad = onLoad.add;

    // api
    return new Treetop(config);
}({
    //
    // Private
    //
    /**
     * Store the component definitions by tagName
     * @type {DefaultDict}
     */
    bindTagName: null,

    /**
     * Store the component definitions by attrName
     * @type {DefaultDict}
     */
    bindAttrName: null,

    /**
     * Treetop library extensions, object attached to the
     * prototype chain of the main Treetop library
     * @type {Object}
     */
    extensions: {},

    /**
     * White-list of request methods types
     * @type {Array}
     */
    METHODS: {"POST": true, "GET": true, "PUT": true, "PATCH": true, "DELETE": true},

    /**
     * List of HTML element for which there can be only one
     * @type {Array}
     */
    SINGLETONS: {"TITLE": true},

    /**
     * Content-Type for Treetop partials
     *
     * This will be set as the `Accept` header for Treetop mediated XHR requests. The
     * server must respond with the same value as `Content-Type` or a client error result.
     *
     * With respect to the media type value, we are taking advantage of the unregistered 'x.' tree while
     * Treetop is a proof-of-concept project. Should a stable API emerge at a later point, then registering a personal
     * or vendor MEME-type would be considered. See https://tools.ietf.org/html/rfc6838#section-3.4
     *
     * @type {String}
     */
    HJ_CONTENT_TYPE: "application/x.treetop-html-partial+xml",

    /**
     * XHR onload handler
     *
     * This will convert the response HTML into nodes and
     * figure out how to attached them to the DOM
     *
     * @param {XMLHttpRequest} xhr The xhr instance used to make the request
     */
    ajaxSuccess: function (xhr) {
        "use strict";
        var $ = this;
        var i, len, temp, child, old, nodes;
        if (xhr.getResponseHeader("content-type") !== $.HJ_CONTENT_TYPE) {
            throw Error("Content-Type is not supported by Treetop '" + xhr.getResponseHeader("content-type") + "'");
        }
        temp = document.createElement("div");
        temp.innerHTML = xhr.responseText;
        nodes = new Array(len);
        for (i = 0, len = temp.children.length; i < len; i++) {
            nodes[i] = temp.children[i];
        }
        node_loop:
            for (i = 0, len = nodes.length; i < len; i++) {
                child = nodes[i];
                if ($.SINGLETONS[child.nodeName.toUpperCase()]) {
                    old = document.getElementsByTagName(child.nodeName)[0];
                    if (old) {
                        old.parentNode.replaceChild(child, old);
                        $.unmount(old);
                        $.mount(child);
                        continue node_loop;
                    }
                }
                if (child.id) {
                    old = document.getElementById(child.id);
                    if (old) {
                        old.parentNode.replaceChild(child, old);
                        $.unmount(old);
                        $.mount(child);
                        continue node_loop;
                    }
                }
            }
    },

    /**
     * Attach an external component to an element and its children depending
     * on the node name or its attributes
     *
     * @param  {HTMLElement} el
     */
    mount: function (el) {
        "use strict";
        var $ = this;
        var i, len, j, comps, comp, attr;
        if (el.nodeType !== 1 && el.nodeType !== 10) {
            return;
        }
        for (i = 0; i < el.children.length; i++) {
            $.mount(el.children[i]);
        }
        el._treetopComponents = (el._treetopComponents || []);
        comps = $.bindTagName.get(el.tagName);
        len = comps.length;
        for (i = 0; i < len; i++) {
            comp = comps[i];
            if (comp && typeof comp.mount === "function" &&
                (!(el._treetopComponents instanceof Array) || el._treetopComponents.indexOf(comp) === -1)
            ) {
                comp.mount(el);
                (el._treetopComponents = (el._treetopComponents || [])).push(comp);
            }
        }
        for (j = el.attributes.length - 1; j >= 0; j--) {
            attr = el.attributes[j];
            comps = $.bindAttrName.get(attr.name);
            len = comps.length;
            for (i = 0; i < len; i++) {
                comp = comps[i];
                if (comp && typeof comp.mount === "function" &&
                    (!(el._treetopComponents instanceof Array) || el._treetopComponents.indexOf(comp) === -1)
                ) {
                    comp.mount(el);
                    (el._treetopComponents = (el._treetopComponents || [])).push(comp);
                }
            }
        }
    },

    /**
     * Trigger unmount handler on all Treetop mounted components attached
     * to a DOM Element
     *
     * @param  {HTMLElement} el
     */
    unmount: function (el) {
        "use strict";
        var $ = this;
        var i, comp;
        // TODO: do this with a stack not recursion
        for (i = 0; i < el.children.length; i++) {
            $.unmount(el.children[i]);
        }
        if (el._treetopComponents instanceof Array) {
            for (i = el._treetopComponents.length - 1; i >= 0; i--) {
                comp = el._treetopComponents[i];
                if (comp && typeof comp.unmount === "function") {
                    comp.unmount(el);
                }
            }
            el._treetopComponents = null;
        }
    },

    /**
     * index all component definitions and mount the full document
     *
     * @param  {Array} setup List of component definitions
     */
    bindComponents: function (setup) {
        "use strict";
        var $ = this;
        var def, i, len = setup.length;
        $.bindTagName = $.index();
        $.bindAttrName = $.index();
        for (i = 0; i < len; i++) {
            def = setup[i];
            if (def.extensionName && def.hasOwnProperty("extension")) {
                $.extensions[def.extensionName] = def.extension;
            }
            if (def.tagName) {
                $.bindTagName.get(def.tagName.toUpperCase()).push(def);
            }
            if (def.attrName) {
                $.bindAttrName.get(def.attrName.toUpperCase()).push(def);
            }
        }
        $.mount(document.body);
    },

    /**
     * index all component definitions some time before the next rendering frame
     *
     * @param  {Array} setup List of component definitions
     */
    bindComponentsAsync: (function () {
        "use strict";
        var id = null;
        return function (setup) {
            var $ = this;
            $.cancelAnimationFrame(id);
            id = $.requestAnimationFrame(function () {
                $.bindComponents(setup);
            });
        };
    }()),


    /**
     * x-browser requestAnimationFrame shim
     *
     * see: https://gist.github.com/paulirish/1579671
     */
    requestAnimationFrame: (function () {
        "use strict";
        var requestAnimationFrame = window.requestAnimationFrame;
        var lastTime = 0;
        var vendors = ["ms", "moz", "webkit", "o"];
        for (var i = 0; i < vendors.length && !requestAnimationFrame; ++i) {
            requestAnimationFrame = window[vendors[i] + "RequestAnimationFrame"];
        }

        if (!requestAnimationFrame) {
            requestAnimationFrame = function (callback) {
                var currTime = new Date().getTime();
                var timeToCall = Math.max(0, 16 - (currTime - lastTime));
                var id = window.setTimeout(function () {
                    callback(currTime + timeToCall);
                }, timeToCall);
                lastTime = currTime + timeToCall;
                return id;
            };
        }

        return function (cb) {
            // must be bound to window object
            return requestAnimationFrame.call(window, cb);
        };
    }()),

    /**
     * x-browser cancelAnimationFrame shim
     *
     * see: https://gist.github.com/paulirish/1579671
     */
    cancelAnimationFrame: (function () {
        "use strict";
        var cancelAnimationFrame = window.cancelAnimationFrame;
        var vendors = ["ms", "moz", "webkit", "o"];
        for (var i = 0; i < vendors.length && !cancelAnimationFrame; ++i) {
            cancelAnimationFrame = window[vendors[i] + "CancelAnimationFrame"] || window[vendors[i] + "CancelRequestAnimationFrame"];
        }

        if (!cancelAnimationFrame) {
            cancelAnimationFrame = function (id) {
                clearTimeout(id);
            };
        }

        return function (cb) {
            // must be bound to window object
            return cancelAnimationFrame.call(window, cb);
        };
    }()),


    /**
     * Create a case insensitive dictionary
     *
     * @returns {Object} implementing { get(string)Array, has(string)bool }
     */
    index: function () {
        "use strict";
        var _store = {};
        return {
            get: function (key) {
                if (typeof key != "string" || key === "") {
                    throw new Error("Index: invalid key (" + key + ")");
                }
                // underscore used to avoid collisions with Object prototype
                var _key = ("_" + key).toUpperCase();
                if (_store.hasOwnProperty(_key)) {
                    return _store[_key];
                } else {
                    return (_store[_key] = []);
                }
            },
            has: function (key) {
                if (typeof key != "string" || key === "") {
                    throw new Error("Index: invalid key (" + key + ")");
                }
                var _key = ("_" + key).toUpperCase();
                return _store.hasOwnProperty(_key);
            }
        };
    },

    /**
     * The dumbest event dispatcher I can think of
     *
     * @return {Object} Object implementing the { add(Function)Function, trigger() } interface
     */
    simpleSignal: function () {
        var listeners = [];
        return {
            add: function (f) {
                var i = listeners.indexOf(f);
                if (i === -1) {
                    i = listeners.push(f) - 1;
                }
                return function remove() {
                    listeners[i] = null;
                };
            },
            trigger: function () {
                for (var i = 0; i < listeners.length; i++) {
                    if (typeof listeners[i] === "function") {
                        listeners[i]();
                    }
                }
            }
        };
    }
}, window.treetop));