package shared

var (
	// TreetopJS is a snapshot of the treetop-client source for use with the example app
	TreetopJS string = `
/* global window, document, XMLHttpRequest, ActiveXObject, history, setTimeout, URLSearchParams */

// Web browser client API for [Treetop request library](https://github.com/rur/treetop).
//
// For an introduction and API docs see https://github.com/rur/treetop-client
//
// This script is written with the following goals:
//      1. Work out-of-the-box, without a build tool or wrapper;
//      2. Maximize compatibility for both modern and legacy browsers;
//      3. Minimize browser footprint to accommodate the use of other JS libraries and frameworks.
//
// Compatibility Caveats
//      The following modern browser APIs are essential. A 'polyfil' must available when necessary.
//       * 'history.pushState' is required so that the location can be updated following partial navigation;
//       * 'HTMLTemplateElement' is required for reliable decoding of HTML strings.
//       * 'FormData' and 'URLSearchParams' are required to use the built-in form element encoding for XHR
//
// Global browser footprint of this script:
//      * Assigns 'window.treetop' with Treetop API instance;
//      * Assigns 'window.onpopstate' with a handler that refreshes the page when a treetop entry is popped from the browser history;
//      * Built-in components attach various event listeners when mounted. (Built-ins can be disabled, see docs)
//

window.treetop = (function ($) {
    "use strict";
    if (window.treetop !== void 0) {
        // throwing an error here is important since it prevents window.treetop from being reassigned
        throw Error("Treetop: treetop global is already defined");
    }

    // First check browser support for essential modern features
    if (typeof window.HTMLTemplateElement === "undefined") {
        throw Error("Treetop: HTMLTemplateElement not supported, a polyfil should be used");
    }
    if (!$.supportsHistory()) {
        throw Error("Treetop: HTML5 History pushState not supported, a polyfil should be used");
    }

    function init(_config) {
        var config = _config instanceof Object ? _config : {};

        // Feature flags for built-in component. Note default values.
        var treetopAttr = true;
        var treetopLinkAttr = true;
        var treetopSubmitterAttr = true;

        for (var key in config) {
            if (!config.hasOwnProperty(key)) {
                continue;
            }
            switch (key.toLowerCase()) {
            case "mountattr":
            case "mountattrs":
                $.mountAttrs = $.copyConfig(config[key]);
                break;
            case "unmountattr":
            case "unmountattrs":
                $.unmountAttrs = $.copyConfig(config[key]);
                break;
            case "merge":
                $.merge = $.copyConfig(config[key]);
                break;
            case "onnetworkerror":
                if (typeof config[key] === "function") {
                    $.onNetworkError = config[key];
                }
                break;
            case "onunsupported":
                if (typeof config[key] === "function") {
                    $.onUnsupported = config[key];
                }
                break;
            case "treetopattr":
                treetopAttr = !(config[key] === false);
                continue;
            case "treetoplinkattr":
                treetopLinkAttr = !(config[key] === false);
                continue;
            case "treetopsubmitterattr":
                treetopSubmitterAttr = !(config[key] === false);
                continue;
            case "mounttags":
            case "unmounttags":
                try {
                    throw new Error("Treetop: Mounting components based upon tag name is no longer supported");
                } catch (err) {
                    // throw error later allowing init to finish its work
                    $.throwErrorAsync(err);
                }
                break;
            default:
                try {
                    throw new Error("Treetop: unknown configuration property '" + key + "'");
                } catch (err) {
                    // throw error later allowing init to finish its work
                    $.throwErrorAsync(err);
                }
            }
        }

        // Add built-in component to configuration.
        // Notice that conflicting custom components will be clobbered.
        if (treetopAttr) {
            document.body.setAttribute("treetop-attr", "enabled")
            $.mountAttrs["treetop-attr"] = $.bind($.bodyMount, $);
        }
        if (treetopLinkAttr) {
            $.mountAttrs["treetop-link"] = $.bind($.linkMount, $);
        }
        if (treetopSubmitterAttr) {
            $.mountAttrs["treetop-submitter"] = $.bind($.submitterMount, $);
        }

        window.onpopstate = function (evt) {
            // Taken from https://github.com/ReactTraining/history/blob/master/modules/createBrowserHistory.js
            var stateFromHistory = (history && history.state) || null;
            var isPageLoadPopState = (evt.state === null) && !!stateFromHistory;

            // Ignore extraneous popstate events in WebKit.
            if (isPageLoadPopState || $.isExtraneousPopstateEvent(evt)) {
                return;
            }
            if (!history.state || !history.state.treetop) {
                // not a treetop state, skip
                return
            }
            $.browserPopState(evt);
        };

        // normalize initial history state
        history.replaceState({treetop: true}, window.document.title, window.location.href)
        $.traverseApply($.wrapElement(document.body), $.mountAttrs);
    }

    /**
        * Treetop API Constructor
        *
        * @constructor
        */
    function Treetop() {}

    var initialized = false;
    /**
        * Configure treetop and mount document.body.
        *
        * @param  {Object} config Dict containing complete page configuration.
        * @throws  {Error} If a config property isn't recognized or 'init' was
        *                  triggered previously
        */
    Treetop.prototype.init = function(config) {
        // Since the DOM is 'stateful', mounting is not a
        // reversible operation. It is crucial therefore that
        // the initial setup process only ever happens once during
        // the lifetime of a page. After that elements will only
        // be mounted and unmounted when being attached or detached
        // from the DOM.
        if (initialized) {
            throw Error("Treetop: Failed attempt to re-initialize. Treetop client is already in use.");
        }
        initialized = true;
        // see https://plainjs.com/javascript/events/running-code-when-the-document-is-ready-15/
        if (document.readyState != "loading") {
            // async used for the sake of consistency with other conditions
            setTimeout(function () {
                init(config);
            });
        } else if (document.addEventListener) {
            // modern browsers
            document.addEventListener("DOMContentLoaded", function(){
                init(config);
            });
        } else {
            // IE <= 8
            document.attachEvent("onreadystatechange", function(){
                if (document.readyState == "complete") init(config);
            });
        }
    };

    /**
        * Update a existing DOM node with a new element. The elements will be merged
        * and (un)mounted in the normal Treetop way.
        *
        * @param {HTMLElement} next: HTMLElement, not yet attached to the DOM
        * @param {HTMLElement} prev: node currently attached to the DOM
        *
        * @throws Error if the elements provided are not valid in some obvious way
        */
    Treetop.prototype.updateElement = function (next, prev) {
        var _next = $.wrapElement(next)
        var _prev = $.wrapElement(prev)
        // make sure an error is raise if initialization happens after the API is used
        initialized = true;
        if (_next.notAnElement() || _prev.notAnElement()) {
            throw new Error("Treetop: Expecting two HTMLElements");
        }
        if (_prev.parentElement().notAnElement()) {
            throw new Error(
                "Treetop: Cannot update an element that is not attached to the DOM"
            );
        }
        $.updateElement(_next, _prev);
    };


    /**
        * Appends a node to a parent and mounts treetop components.
        *
        * @param {HTMLElement} child: HTMLElement, not yet attached to the DOM
        * @param {HTMLElement} mountedParent: node currently attached to the DOM
        *
        * @throws Error if the elements provided are not valid in some obvious way
        */
    Treetop.prototype.mountChild = function(child, mountedParent) {
        var _child = $.wrapElement(child)
        var _mounted = $.wrapElement(mountedParent)
        if (_child.notAnElement() || _mounted.notAnElement()) {
            throw new Error("Treetop: Expecting two HTMLElements");
        }
        _mounted.appendChild(child);
        $.traverseApply(_child, $.mountAttrs);
    };

    /**
        * Inserts new node as a sibling before an element already attached to a parent node.
        * The new node will be mounted.
        *
        * @param {HTMLElement} newSibling: HTMLElement, not yet attached to the DOM
        * @param {HTMLElement} mountedSibling: node currently attached to the DOM
        *
        * @throws Error if the elements provided are not valid in some obvious way
        */
    Treetop.prototype.mountBefore = function(newSibling, mountedSibling) {
        var _new = $.wrapElement(newSibling)
        var _sibling = $.wrapElement(mountedSibling)
        if (_new.notAnElement() || _sibling.notAnElement()) {
            throw new Error("Treetop: Expecting two HTMLElements");
        }
        var parent = _sibling.parentElement();
        if (parent.notAnElement()) {
            throw new Error(
                "Treetop: Cannot mount before a sibling node that is not attached to a parent."
            );
        }
        parent.insertBefore(_new.element, _sibling.element);
        $.traverseApply(_new, $.mountAttrs);
    };

    /**
        * Removes and un-mounts an element from the DOM
        *
        * @param {HTMLElement} mountedElement: HTMLElement, not attached and mounted to the DOM
        *
        * @throws Error if the elements provided is not attached to a parent node
        */
    Treetop.prototype.unmount = function(mountedElement) {
        var _mounted = $.wrapElement(mountedElement)
        if (_mounted.notAnElement()) {
            throw new Error("Treetop: Expecting a HTMLElement to umount");
        }
        var parent = _mounted.parentElement();
        if (parent.notAnElement()) {
            throw new Error(
                "Treetop: Cannot unmount a node that is not attached to a parent."
            );
        }
        parent.removeChild(_mounted.element);
        $.traverseApply(_mounted, $.unmountAttrs);
    };

    /**
        * Get a copy of the treetop configuration,
        * useful for debugging.
        *
        * Note, mutating this object will not affect the configuration.
        *
        * @returns {Object} copy of internal configuration
        */
    Treetop.prototype.config = function () {
        return {
            mountAttrs: $.copyConfig($.mountAttrs),
            unmountAttrs: $.copyConfig($.unmountAttrs),
            merge: $.copyConfig($.merge),
            onNetworkError: $.onNetworkError,
            onUnsupported: $.onUnsupported,
        };
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
        * @param  {array} headers    List of header field-name and field-value pairs
        */
    Treetop.prototype.request = function (method, url, body, contentType, headers) {
        // make sure an error is raise if initialization happens after the API is used
        initialized = true;
        if (!$.METHODS[method.toUpperCase()]) {
            throw new Error("Treetop: Unknown request method '" + method + "'");
        }

        var xhr = $.createXMLHTTPObject();
        if (!xhr) {
            throw new Error("Treetop: XHR is not supported by this browser");
        }
        var requestID = $.lastRequestID = $.lastRequestID + 1;
        xhr.open(method.toUpperCase(), url, true);
        if (headers instanceof Array) {
            for (var i = 0; i < headers.length; i++) {
                xhr.setRequestHeader(headers[i][0], headers[i][1]);
            }
        }
        xhr.setRequestHeader("accept", $.TEMPLATE_CONTENT_TYPE);
        if (contentType) {
            xhr.setRequestHeader("content-type", contentType);
        }
        xhr.onreadystatechange = function() {
            if (xhr.readyState !== 4) {
                return;
            }
            $.endRequest(requestID);
            if (xhr.status < 100) {
                // error occurred, do not attempt to process contents
                return;
            }
            // check if the response can be processed by treetop client library,
            // otherwise trigger 'onUnsupported' signal
            if (xhr.getResponseHeader("x-treetop-redirect") === "SeeOther") {
                // force browser redirect to Location header value
                // if it is defined, otherwise do nothing
                var location = xhr.getResponseHeader("Location")
                if (location !== null) {
                    // Redirect browser window
                    window.location = location;
                }
                return
            }

            if(xhr.getResponseHeader("content-type") === $.TEMPLATE_CONTENT_TYPE) {
                var pageURL = xhr.getResponseHeader("X-Page-URL")
                if (pageURL !== null) {
                    // this response is part of a larger page, add a history entry before processing
                    var responseURL = pageURL;
                    var responseHistory = xhr.getResponseHeader("x-response-history");
                    // NOTE: This HTML5 feature will require a polyfil for some browsers
                    if (typeof responseHistory === "string"
                        && responseHistory.toLowerCase() === "replace"
                        && typeof history.replaceState === "function"
                    ) {
                        // update the current history with a new URL
                        history.replaceState({
                            treetop: true,
                        }, "", responseURL);
                    } else {
                        // add a new history entry using response URL
                        history.pushState({
                            treetop: true,
                        }, "", responseURL);
                    }
                }
                $.xhrProcess(xhr, requestID, pageURL !== null);
                return
            }
            
            if(typeof $.onUnsupported === "function") {
                // Fall through; this is not a response that treetop supports.
                // Allow developer to handle.
                $.onUnsupported(xhr, url);
            }
        };
        xhr.onerror = function () {
            if(typeof $.onNetworkError === "function") {
                // Network level error, likely a connection problem
                $.onNetworkError(xhr);
            }
        };
        xhr.send(body || null);
        $.startRequest(requestID);
    };

    /**
        * treetop.submit will trigger an XHR request derived from the state
        * of a supplied HTML Form element. Request will be sent asynchronously.
        *
        * @public
        * @param {HTMLFormElement} formElement Reference to a HTML form element whose state is to be encoded into a treetop request
        * @param {HTMLElement} submitter Optional element that is capable of adding an input value and overriding form behaviour
        * @throws {Error} If an XHR request cannot derived from the element supplied for any reason.
        */
    Treetop.prototype.submit = function (formElement, submitter) {
        initialized = true;  // ensure that late arriving configuration will be rejected
        var params = $.encodeForm($.wrapElement(formElement), $.wrapElement(submitter));
        if (params) {
            window.treetop.request(
                params["method"],
                params["action"],
                params["data"],
                params["enctype"]
            );
        }
    };

    Treetop.prototype.TEMPLATE_CONTENT_TYPE = $.TEMPLATE_CONTENT_TYPE;

    var api = new Treetop();
    if (window.hasOwnProperty("TREETOP_CONFIG")) {
        // support passive initialization
        api.init(window.TREETOP_CONFIG);
    }
    // api
    return api;
}({
    //
    // Treetop Internal
    //
    /**
        * Store configuration
        */
    mountTags: {},
    mountAttrs: {},
    unmountTags: {},
    unmountAttrs: {},
    onUnsupported: null,
    onNetworkError: null,

    /**
        * Store the treetop custom merge functions
        * @type {Object} object reference
        */
    merge: {},

    /**
        * Track order of requests as well as the elements that were updated.
        * This is necessary because under certain situations late arriving
        * responses should be ignored.
        */
    lastRequestID: 0,
    /**
        * Dictionary is used to track the last request ID that was successfully resolved
        * to a given element "id"
        */
    updates: {},

    /**
        * Track the number of active XHR requests.
        */
    activeCount: 0,

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
        * Content-Type for Treetop templates
        *
        * This will be set as the 'Accept' header for Treetop mediated XHR requests. The
        * server must respond with the same value as 'Content-Type' or a client error result.
        *
        * With respect to the media type value, we are taking advantage of the unregistered 'x.' tree while
        * Treetop is a proof-of-concept project. Should a stable API emerge at a later point, then registering a personal
        * or vendor MEME-type would be considered. See https://tools.ietf.org/html/rfc6838#section-3.4
        *
        * @type {String}
        */
    TEMPLATE_CONTENT_TYPE: "application/x.treetop-html-template+xml",

    START: "treetopstart",
    COMPLETE: "treetopcomplete",

    startRequest: function () {
        "use strict";
        this.activeCount++;
        if (this.activeCount === 1) {
            var event = document.createEvent("Event");
            event.initEvent(this.START, false, false);
            document.dispatchEvent(event);
        }
    },

    endRequest: function () {
        "use strict";
        this.activeCount--;
        if (this.activeCount === 0) {
            var event = document.createEvent("Event");
            event.initEvent(this.COMPLETE, false, false);
            document.dispatchEvent(event);
        }
    },

    /**
        * XHR onload handler
        *
        * This will convert the response HTML into nodes and
        * figure out how to attached them to the DOM
        *
        * @param {XMLHttpRequest} xhr The xhr instance used to make the request
        * @param {number} requestID The number of this request
        * @param {boolean} isPagePartial Flag which will be true if the request response is part of a page
        */
    xhrProcess: function (xhr, requestID, isPagePartial) {
        "use strict";
        var i, len, tmpl, neu, old, matches, targetID;

        // this will require a polyfil for browsers that do not support HTMLTemplateElement
        tmpl = document.createElement("template");
        tmpl.innerHTML = xhr.responseText;
        if (tmpl.content.children.length === 1 && tmpl.content.firstChild.tagName === "TEMPLATE") {
            tmpl = tmpl.content.firstChild
        }
        matches = []
        for (i = 0, len = tmpl.content.children.length; i < len; i++) {
            neu = this.wrapElement(tmpl.content.children[i]);
            if (neu.notAnElement()) {
                continue;
            }
            targetID = neu.id();
            old = new this.ElementWrapper(null)
            if (this.SINGLETONS[neu.tagName().toUpperCase()]) {
                old.element = document.getElementsByTagName(neu.tagName())[0];
            } else if (targetID) {
                old.element = document.getElementById(targetID);
            }
            if (old.notAnElement()) {
                // no match was found for this incoming element, do nothing
                continue;
            }
            var oldParent = old.parentElement()
            if (oldParent.notAnElement()) {
                // for some strange reason the matched element does not have a parent, do nothing
                // TODO: consider whether or not this should throw an error
                continue;
            }
            // Check enclosing nodes have not already been updated by a more recent request
            if (requestID >= this.getLastUpdate(oldParent)) {
                if (isPagePartial) {
                    this.updates["BODY"] = requestID;
                } else if (targetID) {
                    this.updates["#" + targetID] = requestID;
                }
                matches.push(neu, old);
            }
        }
        for (i = 0; i < matches.length; i += 2) {
            this.updateElement(matches[i], matches[i+1]);
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
        window.location.reload();
    },

    /**
        * Given a HTMLELement node attached to the DOM, this will
        * the most recent update requestID for this node and all of its
        * parent nodes.
        *
        * @param node (ElementWrapper)
        * @returns number: the most recent request ID that either this or one of
        *                  its ancestor nodes were updated
        */
    getLastUpdate: function(node) {
        "use strict";
        node.assertElement()
        var updatedID = 0;
        var nodeID = node.id();
        if (node.element === document.body) {
            if ("BODY" in this.updates) {
                updatedID = this.updates["BODY"];
            }
            // don't descent further
            return updatedID;
        } else if (nodeID && "#" + nodeID in this.updates) {
            updatedID = this.updates["#" + nodeID];
        }
        var parent = node.parentElement()
        if (parent.notAnElement()) {
            return updatedID;
        }
        return Math.max(this.getLastUpdate(parent), updatedID)
    },

    /**
        * Default treetop merge method. Replace element followed by sync
        * mount of next and unmount of previous elements.
        *
        * @param  {Element} next The element recently loaded from the API
        * @param  {Element} prev The element currently within the DOM
        */
    defaultComposition: function(next, prev) {
        "use strict";
        var _next = this.wrapElement(next)
        var _prev = this.wrapElement(prev)
        _next.assertElement()
        _prev.assertElement()
        parent = _prev.parentElement()
        if (parent.notAnElement()) {
            // 'prev' is not attached to the DOM
            return
        }
        this.traverseApply(_prev, this.unmountAttrs);
        parent.replaceChild(next, prev);
        this.traverseApply(_next, this.mountAttrs);
    },

    /**
        * Apply a recently loaded element to an existing one attached to the DOM
        *
        * @param  {ElementWrapper} next The element recently loaded from the API
        * @param  {ElementWrapper} prev The element currently within the DOM
    */
    updateElement: function(next, prev) {
        "use strict";
        next.assertElement()
        prev.assertElement()
        var nextValue = next.getAttribute("treetop-merge");
        var prevValue = prev.getAttribute("treetop-merge");
        if (typeof nextValue === "string" &&
            typeof prevValue === "string" &&
            nextValue !== ""
        ) {
            nextValue = nextValue.toLowerCase();
            prevValue = prevValue.toLowerCase();
            if (
                nextValue === prevValue &&
                this.merge.hasOwnProperty(nextValue) &&
                typeof this.merge[nextValue] === "function"
            ) {
                // all criteria have been met, delegate update to custom merge function.
                var mergeFn = this.merge[nextValue];
                mergeFn(next.element, prev.element);
                return;
            }
        }
        this.defaultComposition(next.element, prev.element);
    },

    /**
        * Execute function on elements where the element attributes match
        * a key in supplied hash. Children of head element are traversed in depth first order.
        *
        * @param  {ElementWrapper} head subtree root to descend into
        * @param  {Object} attrFns Functions to apply to elements when the object keys match an attribute name
        */
    traverseApply: function (head, attrFns) {
        "use strict";
        head.assertElement();
        var i, j, comp, name, child, attrs;
        // depth first recursion
        var children = head.children()
        for (i = 0; i < children.length; i++) {
            child = this.wrapElement(children[i]);
            if (child.notAnElement()) continue
            this.traverseApply(child, attrFns);
        }
        // scan element attribute names to match a function
        attrs = head.attributes()
        for (j = attrs.length - 1; j >= 0; j--) {
            name = attrs[j].name.toLowerCase();
            if (attrFns.hasOwnProperty(name)) {
                comp = attrFns[name];
                if (typeof comp === "function") {
                    try {
                        comp(head.element);
                    } catch (err) {
                        this.throwErrorAsync(err)
                    }
                }
            }
        }
    },

    // see https://www.quirksmode.org/js/xmlhttp.html
    XMLHttpFactories: [
        function () {return new XMLHttpRequest();},
        function () {return new ActiveXObject("Msxml2.XMLHTTP");},
        function () {return new ActiveXObject("Msxml3.XMLHTTP");},
        function () {return new ActiveXObject("Microsoft.XMLHTTP");}
    ],

    createXMLHTTPObject: function() {
        "use strict";
        var xmlhttp = false;
        for (var i = 0; i < this.XMLHttpFactories.length; i++) {
            try {
                xmlhttp = this.XMLHttpFactories[i]();
            }
            catch (e) {
                continue;
            }
            break;
        }
        return xmlhttp;
    },

    /**
        * Create copy of config object, all keys are transformed to lowercase.
        * Non-function type config values will be ignored.
        *
        * @param {object} source Dict {String => Function}
        *
        */
    copyConfig: function (source) {
        "use strict";
        var target = {};
        // snippet from
        // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object/assign#Polyfill
        for (var key in source) {
            if (typeof source[key] !== "function") {
                continue;
            }
            // Avoid bugs when hasOwnProperty is shadowed
            if (Object.prototype.hasOwnProperty.call(source, key)) {
                target[key.toLowerCase()] = source[key];
            }
        }
        return target;
    },

    /**
        * Used to throw non-fatal errors.
        *
        * @param {Error} err Error instance to rethrow
        */
    throwErrorAsync: function(err) {
        "use strict";
        setTimeout(function(){
            throw err;
        });
    },


    /**
        * Returns true if the HTML5 history API is supported. Taken from Modernizr via ReactTraining.
        *
        * https://github.com/ReactTraining/history/blob/master/LICENSE
        * https://github.com/ReactTraining/history/blob/master/modules/DOMUtils.js
        * https://github.com/Modernizr/Modernizr/blob/master/LICENSE
        * https://github.com/Modernizr/Modernizr/blob/master/feature-detects/history.js
        */
    supportsHistory: function() {
        "use strict";
        var ua = window.navigator.userAgent

        if ((ua.indexOf('Android 2.') !== -1 || ua.indexOf('Android 4.0') !== -1) &&
        ua.indexOf('Mobile Safari') !== -1 &&
        ua.indexOf('Chrome') === -1 &&
        ua.indexOf('Windows Phone') === -1
        ) {
            return false;
        }
        return window.history && 'pushState' in window.history
    },
    /**
        * Taken from ReactTraining history.
        * https://github.com/ReactTraining/history/blob/master/LICENSE
        * https://github.com/ReactTraining/history/blob/master/modules/DOMUtils.js
        *
        * Returns true if a given popstate event is an extraneous WebKit event.
        * Accounts for the fact that Chrome on iOS fires real popstate events
        * containing undefined state when pressing the back button.
        *
        */
    isExtraneousPopstateEvent: function (event) {
        "use strict";
        return event.state === undefined && window.navigator.userAgent.indexOf('CriOS') === -1;
    },

    /**
        * Convert a form element into parameters for treetop.request API method.
        *
        * This is provided as a utility, it relies upon FormData and URLSearchParams APIs for serialization.
        * If those are not availalble and cannot be polyfilled, this feature is of no use
        * and the programmer must implement their own brand of request data serialization.
        *
        * @param {ElementWrapper} form Required, valid HTML form element with state to be used for treetop request
        * @param {ElementWrapper} submitter Optional element designated as the form 'submitter'
        * @returns Array parameters for treetop request method
        * @throws Error if the target form cannot be encoded for any reason
        */
    encodeForm: function(form, submitter) {
        "use strict";
        if (!(form.element instanceof HTMLFormElement)) {
            throw new Error("Treetop: Expecting HTMLFormElement for encoding, got " + form.element);
        }
        var noValidate = form.hasAttribute("noValidate");
        var method = form.getAttribute("method");
        var action = form.getAttribute("action");
        var enctype = form.getAttribute("enctype");
        if (!submitter.notAnElement()) {
            if (submitter.hasAttribute("formnovalidate")) {
                noValidate =  true;
            }
            if (submitter.hasAttribute("formmethod")) {
                method = submitter.getAttribute("formmethod");
            }
            if (submitter.hasAttribute("formaction")) {
                action = submitter.getAttribute("formaction");
            }
            if (submitter.hasAttribute("formenctype")) {
                enctype = submitter.getAttribute("formenctype");
            }
        }
        if (!noValidate && !form.nativeFormValidate()) {
            // native validiation return 'false'
            return null;
        }

        if (!method) {
            // default method
            method = "GET"
        } else {
            method = method.toUpperCase();
        }

        if (typeof window.FormData === "undefined") {
            throw Error("Treetop: An implementation of FormData is not available. Form cannot be encoded for XHR.");
        }
        var data = new window.FormData(form.element)
        if (!submitter.notAnElement() && submitter.getAttribute("name")) {
            // if a submitter element was supplied adopt that element as an input
            // regardless of the element type. If it has a non-empty "name" attribute
            // add "name" and "value" attribute values to the form data
            data.append(submitter.getAttribute("name"), submitter.getAttribute("value"));
        }

        if (method === "GET") {
            // add form data to the action URL
            if (typeof window.URLSearchParams === "undefined") {
                throw Error("Treetop: An implementation of URLSearchParams is not available. Form cannot be encmded for XHR.");
            }
            data = (new URLSearchParams(data)).toString()
            // strip any existing query and hash, this is the behaviour observed on modern browser
            action = action.split("#")[0].split("?")[0]
            if (data) {
                action = action + "?" + data;
            }
            data = null; // body is null for GET request
        } else {
            if (!enctype) {
                // default encoding
                enctype = "application\/x-www-form-urlencoded"
            } else {
                enctype = enctype.toLowerCase();
            }

            switch (enctype) {
                case "application\/x-www-form-urlencoded":
                    if (typeof window.URLSearchParams === "undefined") {
                        throw Error("Treetop: An implementation of URLSearchParams is not available. Form cannot be encoded for XHR.");
                    }
                    data = (new URLSearchParams(data)).toString();

                    break;

                case "multipart/form-data":
                    // this will be set by the FormData instance which will include a form boundary for the encoding
                    enctype = void 0;
                    data = data
                    break;

                default:
                    // fall-through
                    throw Error("Treetop: Cannot submit form as XHR request with method " + method + " and encoding type " + enctype);
            }
        }

        return {
            method: method,
            action: action,
            data: data,
            enctype: enctype
        }
    },

    // handlers:
    /**
        * This is the implementation of the 'treetop' attributes what can be used to overload
        * html anchors and form elements. It works by registering event handlers on
        * the body element.
        *
        * @type {Object} dictionary with 'mount' and 'unmount' function
        *
        */
    documentClick: function (_evt) {
        "use strict";
        if (!this.attrEquals(document.body, "treetop-attr", "enabled")) {
            return
        }
        var evt = _evt || window.event;
        var elm = this.wrapElement(evt.target || evt.srcElement);
        if (elm.notAnElement()) {
            return;
        }
        var parent = null;
        while (elm.tagName().toUpperCase() !== "A") {
            var parent = elm.parentElement();
            if (parent.notAnElement()) {
                return; // this is not an anchor click
            } else {
                // step up to check if enclosing node is an Anchor tag
                elm = parent;
            }
        }
        // use MouseEvent properties to check for modifiers
        // if engaged, allow default action to proceed
        // see https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/MouseEvent
        if (evt.ctrlKey || evt.shiftKey || evt.altKey || evt.metaKey ||
            (elm.getAttribute("treetop") || "").toLowerCase() === "disabled" ||
            !elm.hasAttribute("href") || !elm.hasAttribute("treetop")
        ) {
            // Use default browser behavior when a modifier key is pressed
            // or treetop has been explicity disabled
            return;
        }
        // hijack standard link click, extract href of link and
        // trigger a Treetop XHR request instead
        evt.preventDefault();
        window.treetop.request("GET", elm.getAttribute("href"));
        return false;
    },

    onSubmit: function (_evt) {
        "use strict";
        if (!this.attrEquals(document.body, "treetop-attr", "enabled")) {
            return
        }
        var evt = _evt || window.event;
        var elm = this.wrapElement(evt.target || evt.srcElement);
        if (!(elm.element instanceof HTMLFormElement)) return;
        if (elm.action() && elm.hasAttribute("treetop") && elm.getAttribute("treetop").toLowerCase() != "disabled") {
            evt.preventDefault();
            // treetop API will serialize the state of the form using FormData and send as a tt request
            window.treetop.submit(elm.element);
            return false;
        }
    },

    linkClick: function (_evt) {
        "use strict";
        var evt = _evt || window.event;
        var elm = this.wrapElement(evt.currentTarget);
        if (elm.notAnElement()) return;
        if (elm.hasAttribute("treetop-link")) {
            var href = elm.getAttribute("treetop-link");
            window.treetop.request("GET", href);
        }
    },

    /**
        * Click event hander for elements with the 'treetop-submitter' attribute.
        *
        * treetop-submitter designates an element as a submitter.
        * Hence, when the element is clicked the state of the targeted form
        * will be used to trigger a Treetop XHR request.
        *
        * Explicit or implicit behavior declared on the form element can be overridden
        * by the designated submitter, using the "formaction" attribute for example.
        *
        * The "form" attribute is also supported where the target form does not enclose the submitter.
        */
    submitClick: function (_evt) {
        "use strict";
        var evt = _evt || window.event;
        var target = this.wrapElement(evt.currentTarget);
        if (target.notAnElement()) return
        var form = null;
        if (target.hasAttribute("treetop-submitter") && target.getAttribute("treetop-submitter") !== "disabled") {
            if (target.hasAttribute("form")) {
                var formID = target.getAttribute("form");
                if (!formID) {
                    return false;
                }
                form = document.getElementById(formID);
            } else {
                // scan up DOM lineage for an enclosing form element
                var cursor = target;
                while (!cursor.notAnElement()) {
                    if (cursor.element instanceof HTMLFormElement) {
                        form = cursor.element;
                        break;
                    }
                    cursor = cursor.parentElement();
                }
            }
            if (!(form instanceof HTMLFormElement)) return false;
            // pass click target as 'submitter'
            window.treetop.submit(form, target.element)
            evt.preventDefault();
            return false;
        }
        // fall-through, default click behaviour not prevented
    },

    bodyMount: function (el) {
        "use strict";
        var _elmt = this.wrapElement(el);
        _elmt.addEventListener("click", this.bind(this.documentClick, this), false);
        _elmt.addEventListener("submit", this.bind(this.onSubmit, this), false);
    },
    linkMount: function (el) {
        "use strict";
        var _elmt = this.wrapElement(el);
        _elmt.addEventListener("click", this.bind(this.linkClick, this), false);
    },
    submitterMount: function (el) {
        "use strict";
        var _elmt = this.wrapElement(el);
        _elmt.addEventListener("click", this.bind(this.submitClick, this), false);
    },

    /**
        *
        * @param {Element} el DOM element to test for attribute value
        * @param {String} attr Attribute name
        * @param {String} expect value for case insensitive comparison
        */
    attrEquals: function (el, attr, expect) {
        "use strict";
        var _elmt = this.wrapElement(el);
        if (_elmt.notAnElement()) return false;
        if (_elmt.hasAttribute(attr)) {
            var value = _elmt.getAttribute(attr);
            if (!value && !expect) {
                return true;
            } else if (typeof value === "string" && typeof expect === "string") {
                return value.toLowerCase() === expect.toLowerCase();
            }
        }
        return false;
    },

    /**
        * Cheap and cheerful bind implementation
        */
    bind: function (f, that) {
        "use strict";
        return function () {
            switch (arguments.length) {
            case 0:
                return f.call(that)
            case 1:
                return f.call(that, arguments[0])
            case 2:
                return f.call(that, arguments[0], arguments[1])
            case 3:
                return f.call(that, arguments[0], arguments[1], arguments[2])
            case 4:
                return f.call(that, arguments[0], arguments[1], arguments[2], arguments[3])
            }
            var args = []
            for (let i = 0; i < arguments.length; i++) {
                args.push(arguments[i]);
            }
            return f.apply(that, args)
        }
    },

    /**
        * As more testing and QC was carried out, there has been an increasing need
        * for DOM API related shims. Edge cases were starting to clog up the library code
        * so this wrapper was created to localize all that nastiness.
        *
        * This code is for internal use only, hence it can be minimal for the treetop-client use case.
        * However, I still expect the code to expand as more conditions are factored out and
        * more issues present themselves during development.
        *
        * see https://github.com/rur/treetop-client/issues/5
        */
    ElementWrapper: (function() {
        "use strict";
        function Wrap(e) {
            if (e instanceof Wrap) {
                throw new Error("Treetop: Double wrapped element, " + e)
            }
            this.element = e
        }
        Wrap.prototype = {
            // getters
            // these will throw an error if the underlying element is not defined
            action: function () {
                return this.deshadow("action");
            },
            checked: function () {
                return this.deshadow("checked");
            },
            children: function () {
                return this.deshadow("children");
            },
            attributes: function () {
                return this.deshadow("attributes");
            },
            elements: function () {
                return this.deshadow("elements");
            },
            id: function () {
                return this.deshadow("id");
            },
            name: function () {
                return this.deshadow("name");
            },
            nodeName: function () {
                return this.deshadow("nodeName");
            },
            parentElement: function () {
                this.assertElement()
                if (this.element instanceof window.HTMLFormElement) {
                    // HTML form element can have their properties shadowed by
                    // named inputs. Attempt to ensure that the parentElement
                    // property is telling the truth.
                    if (!this.element.parentElement) {
                        return new Wrap(this.element.parentElement);
                    } else if (!this.element.parentNode) {
                        return new Wrap(this.element.parentNode);
                    } else if (this.element.parentElement === this.element.parentNode) {
                        // this properties cannot have been shadowed by an input since one input element cannot have two different names
                        return new Wrap(this.element.parentElement);
                    } else if (Array.prototype.indexOf.call(this.element.parentElement.children, this.element) !== -1) {
                        // we have proven that this element is a child of the parentElement node.
                        // likely parentNode was shadowed
                        return new Wrap(this.element.parentElement);
                    } else if (Array.prototype.indexOf.call(this.element.parentNode.children, this.element) !== -1) {
                        // we have proven that this element is a child of the parentNode node.
                        // likely parentElement was shadowed
                        return new Wrap(this.element.parentNode);
                    } else {
                        // both parentElement and parentNode have been shadowed, all bets are off
                        throw new Error("Form input names are shadowing the DOM API. Please rename inputs.")
                    }
                }
                return new Wrap(this.element.parentElement);
            },
            tagName: function () {
                return this.deshadow("tagName");
            },
            value: function () {
                return this.deshadow("value");
            },

            // element methods
            // these will throw an error if the underlying element is not defined
            addEventListener: function(event, listener, capture) {
                this.assertElement()
                if (EventTarget && EventTarget.prototype.addEventListener instanceof Function) {
                    return EventTarget.prototype.addEventListener.call(this.element, event, listener, capture)
                } else if (this.element.__proto__.attachEvent instanceof Function) {
                    this.element.__proto__.attachEvent.call(this.element, event, listener)
                } else {
                    throw new Error("addEventListener is not supported by this user agent")
                }
            },
            appendChild: function(nue) {
                this.assertElement()
                return Node.prototype.appendChild.call(this.element, nue)
            },
            insertBefore: function(nue, child) {
                this.assertElement()
                return Node.prototype.insertBefore.call(this.element, nue, child)
            },
            removeChild: function(old) {
                this.assertElement()
                return Node.prototype.removeChild.call(this.element, old)
            },
            replaceChild: function(nue, old) {
                this.assertElement()
                return Node.prototype.replaceChild.call(this.element, nue, old)
            },
            getAttribute: function(name) {
                this.assertElement()
                return Element.prototype.getAttribute.call(this.element, name)
            },
            hasAttribute: function(name) {
                this.assertElement()
                return Element.prototype.hasAttribute.call(this.element, name)
            },

            // wrapper specific API
            notAnElement: function (){
                return !(this.element instanceof window.Element);
            },
            /**
                * @throws Error: throws assertion error if this is not an element
                */
            assertElement: function (){
                if (this.notAnElement()) {
                    throw new Error("Assertion error, " + this.element + " is not an element")
                }
            },
            /**
                * Hack to access element informational properties which might have been shadowed
                * @throws Error: if wrapped value is not an element
                */
            deshadow: function (name) {
                if (this.element instanceof window.HTMLFormElement) {
                    if (this.element[name] instanceof window.Element) {
                        // This hack temporarily removes the input element to obtain
                        // access to the property.
                        // NOTE: This is an expensive operation and should be avoided for performance sensitive code.
                        var input = this.element[name];
                        var inputParent = input.parentElement; // this is unlikely to be shadowed since <form> is not a valid input
                        var placeholder = document.createElement("span");
                        inputParent.replaceChild(placeholder, input)
                        // access un-shadowed property
                        var value = this.element[name]
                        inputParent.replaceChild(input, placeholder)
                        return value
                    }
                }
                this.assertElement()
                return this.element[name]
            },
            /**
                * attempt to trigger native form validation and return a boolean flag
                * to indicate if the form is valid or not. false === invalid
                * @throws Error: if the element being wrapped is not a HTMLFormElement
                */
            nativeFormValidate: function( ) {
                if (this.element instanceof window.HTMLFormElement) {
                    if (typeof HTMLFormElement.prototype.reportValidity === "function") {
                        if (!HTMLFormElement.prototype.reportValidity.call(this.element)) {
                            return false;
                        }
                    } else if (typeof HTMLFormElement.prototype.checkValidity === "function") {
                        if (!HTMLFormElement.prototype.checkValidity.call(this.element)) {
                            return false;
                        }
                    }
                } else {
                    throw new Error("Treetop: Cannot validate " + this.element + ", node is not a HTMLFormElement")
                }
                return true
            }
        }
        return Wrap;
    }()),
    wrapElement: function(e) {
        "use strict";
        return new this.ElementWrapper(e)
    }
}));`
)
