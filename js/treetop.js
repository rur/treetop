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
     * @param  {string} body   Encoded request body
     * @param  {string} contentType    Encoding of the request body
     */
    Treetop.prototype.request = function (method, url, body, contentType) {
        if (!$.METHODS[method.toUpperCase()]) {
            throw new Error("Treetop: Unknown request method '" + method + "'");
        }
        var req = (XMLHttpRequest) ? new XMLHttpRequest() : new ActiveXObject("MSXML2.XMLHTTP");
        req.open(method.toUpperCase(), url, true);
        req.setRequestHeader("accept", [$.PARTIAL_CONTENT_TYPE, $.FRAGMENT_CONTENT_TYPE].join(", "));
        if (contentType) {
            req.setRequestHeader("content-type", contentType);
        }
        req.onload = function () {
            $.xhrLoad(req, method);
            onLoad.trigger();
        };
        req.send(body || null);
    };

    Treetop.prototype.onLoad = onLoad.add;

    /**
     * treetop.submit will trigger an XHR request derived from the state
     * of a supplied HTML Form element.
     */
    Treetop.prototype.submit = function (formElement, submitInputElement) {
        function dataHandler(fdata) {
            window.setTimeout(function () {
                window.treetop.request(
                    fdata.method,
                    fdata.action,
                    fdata.data,
                    fdata.enctype
                );
            }, 0);
        }
        new $.FormSerializer(formElement, dataHandler);
    }

    Treetop.prototype.PARTIAL_CONTENT_TYPE = $.PARTIAL_CONTENT_TYPE;
    Treetop.prototype.FRAGMENT_CONTENT_TYPE = $.FRAGMENT_CONTENT_TYPE;

    // api
    return new Treetop(config);
}({
    //
    // Private
    //
    /**
     * Store the component definitions by tagName
     * @type {Object} index
     */
    bindTagName: null,

    /**
     * Store the component definitions by attrName
     * @type {Object} index
     */
    bindAttrName: null,

    /**
     * Store the treetop composition definitions
     * @type {Object} object reference
     */
    composition: {},

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
    PARTIAL_CONTENT_TYPE: "application/x.treetop-html-partial+xml",
    FRAGMENT_CONTENT_TYPE: "application/x.treetop-html-fragment+xml",

    /**
     * XHR onload handler
     *
     * This will convert the response HTML into nodes and
     * figure out how to attached them to the DOM
     *
     * @param {XMLHttpRequest} xhr The xhr instance used to make the request
     * @param {string} method HTTP method used
     */
    xhrLoad: function (xhr, method) {
        "use strict";
        var $ = this;
        var i, len, temp, child, old, nodes;
        i = len = temp = child = old = nodes = undefined;
        var responseContentType = xhr.getResponseHeader("content-type");
        var responseURL = xhr.getResponseHeader("x-response-url") || xhr.responseURL;
        if (responseContentType != $.PARTIAL_CONTENT_TYPE && responseContentType != $.FRAGMENT_CONTENT_TYPE) {
            // TODO: Consider checking for 4xx and 5xx status and handle these responses differently
            //       since the 'responseURL' may be a POST only or a treetop fragment route hence not accessible
            //       the following way.
            if (xhr.getResponseHeader("x-treetop-redirect") != null) {
                if (method.toUpperCase() == "GET") {
                    window.location = xhr.getResponseHeader("x-treetop-redirect");
                    return;
                } else {
                    throw Error("Treetop client cannot redirect a non-GET request");
                }
            }
            throw Error("Non-treetop response from URL: " + responseURL);
        }

        if (responseContentType == $.PARTIAL_CONTENT_TYPE && window.history) {
            window.history.pushState({
                treetop: true,
            }, "", responseURL);
        }

        temp = document.createElement("div");
        temp.innerHTML = xhr.responseText;
        nodes = new Array(temp.children.length);
        for (i = 0, len = temp.children.length; i < len; i++) {
            nodes[i] = temp.children[i];
        }
        node_loop:
            for (i = 0, len = nodes.length; i < len; i++) {
                child = nodes[i];
                if ($.SINGLETONS[child.nodeName.toUpperCase()]) {
                    old = document.getElementsByTagName(child.nodeName)[0];
                    if (old) {
                        $.compose(child, old);
                        continue node_loop;
                    }
                }
                if (child.id) {
                    old = document.getElementById(child.id);
                    if (old) {
                        $.compose(child, old);
                        continue node_loop;
                    }
                }
            }
    },

    /**
     * Default treetop composition method
     *
     * @param  {HTMLElement} next The element recently loaded from the API
     * @param  {HTMLElement} prev The element currently within the DOM
     */
    defaultComposition: function(next, prev) {
        prev.parentNode.replaceChild(next, prev);
    },

    /**
     * Apply a recently loaded element to an existing one attached to the DOM
     *
     * @param  {HTMLElement} next The element recently loaded from the API
     * @param  {HTMLElement} prev The element currently within the DOM
    */
    compose: function(next, prev) {
        var $ = this;
        var nextCompose = next.getAttribute("treetop-compose");
        var prevCompose = prev.getAttribute("treetop-compose");
        var compose = $.defaultComposition;
        if (typeof nextCompose === "string" && typeof prevCompose === "string") {
            nextCompose = nextCompose.toLowerCase();
            prevCompose = prevCompose.toLowerCase();
            if (nextCompose === prevCompose && nextCompose in $.composition && typeof $.composition[nextCompose] === "function") {
                compose = $.composition[nextCompose];
            }
        }

        var asyncMount = compose(next, prev);
        if (typeof asyncMount === "function") {
            asyncMount($.asyncMountFn(next, prev));
        } else {
            $.mount(next);
            $.unmount(prev);
        }
    },

    asyncMountFn: function (n, p) {
        var $ = this;
        return function () {
            if (n !== null && p !== null) {
                $.mount(n);
                $.unmount(p);
                n = null;
                p = null;
            }
        };
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
            if (def.composition instanceof Object) {
                for (var prop in def.composition) {
                    if (def.composition.hasOwnProperty(prop) && typeof def.composition[prop] === "function") {
                        $.composition[prop.toLowerCase()] = def.composition[prop];
                    }
                }
            } else {
                if (def.tagName) {
                    $.bindTagName.get(def.tagName.toUpperCase()).push(def);
                }
                if (def.attrName) {
                    $.bindAttrName.get(def.attrName.toUpperCase()).push(def);
                }
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
    },

    FormSerializer: (function () {
        "use strict";
        /**
         * techniques:
         */
        var URLEN_GET = 0;   // GET method
        var URLEN_POST = 1;  // POST method, enctype is application/x-www-form-urlencoded (default)
        var PLAIN_POST = 2;  // POST method, enctype is text/plain
        var MULTI_POST = 3;  // POST method, enctype is multipart/form-data

        /**
         * @private
         * @constructor
         * @param {FormElement}   elm       The form to be serialized
         * @param {Function}      callback  Called when the serialization is complete (may be sync or async)
         */
        function FormSerializer(elm, callback) {
            if (!(this instanceof FormSerializer)) {
                return new FormSerializer(elm, callback);
            }

            var nFile, sFieldType, oField, oSegmReq, oFile;
            var bIsPost = elm.method.toLowerCase() === "post";
            var fFilter = window.encodeURIComponent;

            this.onRequestReady = callback;
            this.receiver = elm.action;
            this.status = 0;
            this.segments = [];

            if (bIsPost) {
                this.contentType = elm.enctype ? elm.enctype : "application\/x-www-form-urlencoded";
                switch (this.contentType) {
                case "multipart\/form-data":
                    this.technique = MULTI_POST;

                    try {
                        // ...to let FormData do all the work
                        this.data = new window.FormData(elm);
                        if (this.data) {
                            this.processStatus();
                            return;
                        }
                    } catch (_) {
                        "pass";
                    }

                    break;

                case "text\/plain":
                    this.technique = PLAIN_POST;
                    fFilter = plainEscape;
                    break;

                default:
                    this.technique = URLEN_POST;
                }
            } else {
                this.technique = URLEN_GET;
            }

            for (var i = 0, len = elm.elements.length; i < len; i++) {
                oField = elm.elements[i];
                if (!oField.hasAttribute("name")) { continue; }
                sFieldType = oField.nodeName.toUpperCase() === "INPUT" ? oField.getAttribute("type").toUpperCase() : "TEXT";
                if (sFieldType === "FILE" && oField.files.length > 0) {
                    if (this.technique === MULTI_POST) {
                        if (!window.FileReader) {
                            throw new Error("Operation not supported: cannot upload a document via AJAX if FileReader is not supported");
                        }
                        /* enctype is multipart/form-data */
                        for (nFile = 0; nFile < oField.files.length; nFile++) {
                            oFile = oField.files[nFile];
                            oSegmReq = new window.FileReader();
                            oSegmReq.onload = this.fileReadHandler(oField, oFile);
                            oSegmReq.readAsBinaryString(oFile);
                        }
                    } else {
                        /* enctype is application/x-www-form-urlencoded or text/plain or method is GET: files will not be sent! */
                        for (nFile = 0; nFile < oField.files.length; this.segments.push(fFilter(oField.name) + "=" + fFilter(oField.files[nFile++].name)));
                    }
                } else if ((sFieldType !== "RADIO" && sFieldType !== "CHECKBOX") || oField.checked) {
                    /* field type is not FILE or is FILE but is empty */
                    this.segments.push(
                        this.technique === MULTI_POST ? /* enctype is multipart/form-data */
                            "Content-Disposition: form-data; name=\"" + oField.name + "\"\r\n\r\n" + oField.value + "\r\n"
                        : /* enctype is application/x-www-form-urlencoded or text/plain or method is GET */
                            fFilter(oField.name) + "=" + fFilter(oField.value)
                    );
                }
            }
            this.processStatus();
        }

        /**
         * Create FileReader onload handler
         *
         * @return {function}
         */
        FormSerializer.prototype.fileReadHandler = function (field, file) {
            var self = this;
            var index = self.segments.length;
            self.segments.push(
                "Content-Disposition: form-data; name=\"" + field.name + "\"; " +
                "filename=\""+ file.name + "\"\r\n" +
                "Content-Type: " + file.type + "\r\n\r\n");
            self.status++;
            return function (oFREvt) {
                self.segments[index] += oFREvt.target.result + "\r\n";
                self.status--;
                self.processStatus();
            };
        };

        /**
         * Is called when a pass of serialization has completed.
         *
         * It will be called asynchronously if file reading is taking place.
         */
        FormSerializer.prototype.processStatus = function () {
            if (this.status > 0) { return; }
            /* the form is now totally serialized! prepare the data to be sent to the server... */
            var sBoundary, method, url, hash, data, enctype;

            switch (this.technique) {
            case URLEN_GET:
                method = "GET";
                url = this.receiver.split("#");
                hash = url.length > 1 ? "#" + url.splice(1).join("#") : "";  // preserve the hash
                url = url[0].replace(/(?:\?.*)?$/, this.segments.length > 0 ? "?" + this.segments.join("&") : "") + hash;
                data = null;
                enctype = null;
                break;

            case URLEN_POST:
            case PLAIN_POST:
                method = "POST";
                url = this.receiver;
                enctype =  this.contentType;
                data  = this.segments.join(this.technique === PLAIN_POST ? "\r\n" : "&");
                break;

            case MULTI_POST:
                method = "POST";
                url = this.receiver;
                if (this.data) {
                    // use native FormData multipart data
                    data = this.data;
                    enctype = null;
                } else {
                    // construct serialized multipart data manually
                    sBoundary = "---------------------------" + Date.now().toString(16);
                    enctype = "multipart\/form-data; boundary=" + sBoundary;
                    data = "--" + sBoundary + "\r\n" + this.segments.join("--" + sBoundary + "\r\n") + "--" + sBoundary + "--\r\n";
                    if (window.Uint8Array) {
                        data = createArrayBuffer(data);
                    }
                }
                break;
            }

            this.onRequestReady({
                method: method,
                action: url,
                data: data,
                enctype: enctype
            });
        };

        /**
         * Used to escape strings for encoding text/plain
         *
         * eg. "4\3\7 - Einstein said E=mc2" ----> "4\\3\\7\ -\ Einstein\ said\ E\=mc2"
         *
         * @param  {stirng} sText
         * @return {string}
         */
        function plainEscape(sText) {
            return sText.replace(/[\s\=\\]/g, "\\$&");
        }

        /**
         * @param  {string} str
         * @return {ArrayBuffer}
         */
        function createArrayBuffer(str) {
            var nBytes = str.length;
            var ui8Data = new window.Uint8Array(nBytes);
            for (var i = 0; i < nBytes; i++) {
                ui8Data[i] = str.charCodeAt(i) & 0xff;
            }
            return ui8Data;
        }

        return FormSerializer;
    }())
}, window.treetop));


/**
 * Register treetop delegation event handlers on the document.body
 */
window.treetop.push(function ($) {
    "use strict";

    // handlers:
    function documentClick(_evt) {
        var evt = _evt || window.event;
        var elm = _evt.target || _evt.srcElement;
        while (elm.tagName.toUpperCase() !== "A") {
            if (elm.parentElement) {
                elm = elm.parentElement;
            } else {
                return; // this is not an anchor click
            }
        }
        $.anchorClicked(evt, elm);
    }

    function updateModifiers(_kevt) {
        var kevt = _kevt || window.event;
        $.shiftKey =  kevt.shiftKey;
        $.ctrlKey =  kevt.ctrlKey;
        $.metaKey =  kevt.metaKey;
    }

    function onSubmit(_evt) {
        var evt = _evt || window.event;
        var elm = _evt.target || _evt.srcElement;
        $.formSubmit(evt, elm);
    }

    function onPopState(_evt) {
        var evt = _evt || window.event;
        $.browserPopState(evt);
    }

    /**
     * treetop event delegation component definition
     */
    return {
        tagName: "body",
        mount: function (el) {
            if (el.addEventListener) {
                el.addEventListener("click", documentClick, false);
                el.addEventListener("submit", onSubmit, false);
                el.addEventListener("keydown", updateModifiers, false);
                el.addEventListener("keyup", updateModifiers, false);
            } else if (el.attachEvent) {
                el.attachEvent("onclick", documentClick);
                el.attachEvent("onsubmit", onSubmit);
                el.attachEvent("onkeydown", updateModifiers);
                el.attachEvent("onkeyup", updateModifiers);
            } else {
                throw new Error("Treetop Events: Event delegation is not supported in this browser!");
            }
            window.onpopstate = onPopState;
        },
        unmount: function (el) {
            if (el.removeEventListener) {
                el.removeEventListener("click", documentClick);
                el.removeEventListener("submit", onSubmit);
                el.removeEventListener("keydown", updateModifiers);
                el.removeEventListener("keyup", updateModifiers);
            } else if (el.detachEvent) {
                el.detachEvent("onclick", documentClick);
                el.detachEvent("onsubmit", onSubmit);
                el.detachEvent("onkeydown", updateModifiers);
                el.detachEvent("onkeyup", updateModifiers);
            }
            if(window.onpopstate === onPopState) {
                window.onpopstate = null;
            }
        }
    };
}({
    //
    // Private
    //

    // track modifier key state
    shiftKey: false,
    ctrlKey: false,
    metaKey: false,

    /**
     * document submit event handler
     *
     * @param {Event} evt
     */
    anchorClicked: function (evt, elm) {
        "use strict";
        if (this.shiftKey || this.ctrlKey || this.metaKey ||
            (elm.getAttribute("treetop") || "").toLowerCase() === "disabled"
        ) {
            // Use default browser behaviour when a modifier key is pressed
            // or treetop has been explicity disabled
            return
        }
        if (elm.hasAttribute("treetop-link")) {
            // 'treetop-link' attribute can be used as an alternative to 'href' attribute.
            // This is useful when default 'href' behavior is undesirable.
            evt.preventDefault();
            window.treetop.request("GET", elm.getAttribute("treetop-link"));
            return false;
        } else if (elm.href && elm.hasAttribute("treetop")) {
            // hijack standard link click, extract href of link and
            // trigger a Treetop XHR request instead
            evt.preventDefault();
            window.treetop.request("GET", elm.href);
            return false;
        }
    },

    /**
     * document submit event handler
     *
     * @param {Event} evt
     */
    formSubmit: function (evt, elm) {
        "use strict";
        if (elm.action && elm.hasAttribute("treetop") && elm.getAttribute("treetop").toLowerCase() != "disabled") {
            evt.preventDefault();

            // Serialize HTML form including file inputs and trigger a treetop request.
            // The request will be executed asynchronously.
            // TODO: If an error occurs during serialization there should be some logging/recovery mechanism in the API
            window.treetop.submit(elm);

            return false;
        }
    },

    /**
     * document history pop state event handler
     *
     * @param {PopStateEvent} e
     */
    browserPopState: function () {
        "use strict";
        // force browser to refresh the page when the back
        // nav is triggered, seems to be the best thing to do
        location.reload();
    }
}));

