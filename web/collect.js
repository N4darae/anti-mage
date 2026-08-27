
var AM = (function () {
  "use strict";

  var PROBE_TIMEOUT_MS = 9000;

  function Unsupported(reason) {
    var e = new Error(reason || "facility not available");
    e.amUnsupported = true;
    return e;
  }
  function unsupported(reason) {
    throw Unsupported(reason);
  }
  function reasonOf(e) {
    if (!e) return "unknown";
    if (typeof e === "string") return e;
    var m = e.message || String(e);
    return String(m).slice(0, 300);
  }
  function attempt(fn, fallback) {
    try {
      var v = fn();
      return v === undefined ? fallback : v;
    } catch (e) {
      return fallback;
    }
  }
  function withTimeout(promise, ms, label) {
    return new Promise(function (resolve, reject) {
      var done = false;
      var t = setTimeout(function () {
        if (done) return;
        done = true;
        var e = new Error("timed out after " + ms + " ms: " + label);
        e.amTimeout = true;
        reject(e);
      }, ms);
      Promise.resolve(promise).then(
        function (v) {
          if (done) return;
          done = true;
          clearTimeout(t);
          resolve(v);
        },
        function (e) {
          if (done) return;
          done = true;
          clearTimeout(t);
          reject(e);
        }
      );
    });
  }

  var bootstrap = { dates: null, fonts: null, nonce: "", source: "none" };

  var meta = {};

  function readInlineBootstrap() {
    var el = document.getElementById("am-bootstrap");
    if (!el || !el.textContent) return null;
    var txt = el.textContent.trim();
    if (!txt || txt === "{}") return null;
    try {
      return JSON.parse(txt);
    } catch (e) {
      return null;
    }
  }

  function fetchJSON(url) {
    if (typeof fetch !== "function") return Promise.resolve(null);
    return fetch(url, { cache: "no-store", credentials: "omit" })
      .then(function (r) {
        return r.ok ? r.json() : null;
      })
      .catch(function () {
        return null;
      });
  }

  function normaliseSamples(x) {
    if (!Array.isArray(x) || !x.length) return null;
    var out = [];
    x.forEach(function (e) {
      if (typeof e === "string" && e) {
        out.push({ date: e, epochMs: null });
        return;
      }
      if (e && typeof e === "object") {
        var date = typeof e.date === "string" ? e.date : typeof e.iso === "string" ? e.iso : null;
        var ms = typeof e.epochMs === "number" && isFinite(e.epochMs) ? e.epochMs : null;
        if (date || ms !== null) out.push({ date: date, epochMs: ms, iso: typeof e.iso === "string" ? e.iso : null });
      }
    });
    return out.length ? out : null;
  }

  function normaliseBootstrap(b) {
    if (!b || typeof b !== "object") return null;
    var t = b.time && typeof b.time === "object" ? b.time : {};
    var dates =
      normaliseSamples(t.offsets) ||
      normaliseSamples(t.dates) ||
      normaliseSamples(t.offsetDates) ||
      normaliseSamples(b.dates) ||
      normaliseSamples(b.offsetDates) ||
      null;
    var fonts = null;
    [b.fonts, b.font].forEach(function (f) {
      if (!f || typeof f !== "object") return;
      fonts = fonts || {};
      ["measureString", "bases", "families", "controls", "coverage"].forEach(function (k) {
        if (fonts[k] === undefined && f[k] !== undefined) fonts[k] = f[k];
      });
    });
    if (Array.isArray(b.fontControls) && b.fontControls.length) {
      fonts = fonts || {};
      if (!Array.isArray(fonts.controls) || !fonts.controls.length) fonts.controls = b.fontControls;
    }
    var nonce = typeof b.nonce === "string" ? b.nonce : "";
    if (!dates && !fonts && !nonce) return null;
    return { dates: dates, fonts: fonts, nonce: nonce };
  }

  function loadBootstrap() {
    var inline = normaliseBootstrap(readInlineBootstrap());
    if (inline) {
      inline.source = "inline page bootstrap";
      return Promise.resolve(inline);
    }
    var glob = typeof window !== "undefined" ? normaliseBootstrap(window.AM_BOOTSTRAP) : null;
    if (glob) {
      glob.source = "page global";
      return Promise.resolve(glob);
    }
    return fetchJSON("./api/bootstrap")
      .then(function (v) {
        var n = normaliseBootstrap(v);
        if (n) {
          n.source = "./api/bootstrap";
          return n;
        }
        return fetchJSON("./bootstrap.json").then(function (v2) {
          var n2 = normaliseBootstrap(v2);
          if (n2) {
            n2.source = "./bootstrap.json";
            return n2;
          }
          return { dates: null, fonts: null, nonce: "", source: "none" };
        });
      })
      .catch(function () {
        return { dates: null, fonts: null, nonce: "", source: "none" };
      });
  }

  function fontInput() {
    var d = AM_FONT_DATA;
    var b = bootstrap && bootstrap.fonts;
    if (!b) return { data: d, source: "built-in table" };
    var sources = {};
    function pick(key, supplied) {
      if (supplied) {
        sources[key] = "page bootstrap";
        return supplied;
      }
      sources[key] = "built-in table";
      return d[key];
    }
    var data = {
      measureString: pick("measureString", typeof b.measureString === "string" && b.measureString ? b.measureString : null),
      bases: pick("bases", Array.isArray(b.bases) && b.bases.length ? b.bases : null),
      families: pick("families", Array.isArray(b.families) && b.families.length ? b.families : null),
      controls: pick("controls", Array.isArray(b.controls) && b.controls.length ? b.controls : null),
      coverage: pick("coverage", Array.isArray(b.coverage) && b.coverage.length ? b.coverage : null)
    };
    return { data: data, source: sources.families, sources: sources };
  }

  var measureCache = null;

  function measurer() {
    if (measureCache) return measureCache;
    var c = document.createElement("canvas");
    var ctx = c.getContext && c.getContext("2d");
    if (!ctx || typeof ctx.measureText !== "function") {
      unsupported("canvas 2d measureText is not available");
    }
    var baseWidths = {};
    var m = {
      width: function (spec, text) {
        ctx.font = "72px " + spec;
        return ctx.measureText(text).width;
      },
      base: function (base, text) {
        var k = base + "\u0000" + text;
        if (!(k in baseWidths)) baseWidths[k] = m.width(base, text);
        return baseWidths[k];
      }
    };
    measureCache = m;
    return m;
  }

  function quoteFamily(f) {
    return '"' + String(f).replace(/"/g, '\\"') + '"';
  }

  function differsFromGenerics(m, family, bases, text) {
    var q = quoteFamily(family);
    for (var i = 0; i < bases.length; i++) {
      var b = bases[i];
      var withFamily = m.width(q + "," + b, text);
      var generic = m.base(b, text);
      if (Math.abs(withFamily - generic) > 0.02) return true;
    }
    return false;
  }

  function fontsCheck(family) {
    try {
      if (!document.fonts || typeof document.fonts.check !== "function") return null;
      return document.fonts.check("72px " + quoteFamily(family));
    } catch (e) {
      return null;
    }
  }

  var nativeCache = null;

  function nativeTargets() {
    var list = [];
    function add(t) {
      list.push(t);
    }
    function proto(name) {
      try {
        return window[name] && window[name].prototype;
      } catch (e) {
        return null;
      }
    }

    var navProto = proto("Navigator");
    if (navProto) {
      ["userAgent", "platform", "hardwareConcurrency", "deviceMemory", "language", "languages", "plugins", "webdriver"].forEach(function (n) {
        add({ key: "navigator." + n, owner: navProto, name: n, kind: "getter", instance: navigator });
      });
    }
    var scrProto = proto("Screen");
    if (scrProto) {
      ["width", "height", "availWidth", "availHeight", "colorDepth", "pixelDepth"].forEach(function (n) {
        add({ key: "screen." + n, owner: scrProto, name: n, kind: "getter", instance: screen });
      });
    }
    var dprDesc = attempt(function () {
      return Object.getOwnPropertyDescriptor(window, "devicePixelRatio");
    }, null);
    if (dprDesc && typeof dprDesc.get === "function") {
      add({ key: "window.devicePixelRatio", owner: window, name: "devicePixelRatio", kind: "getter", instance: window });
    }
    var nodeProto = proto("Node");
    var probeDiv = attempt(function () {
      return document.createElement("div");
    }, null);
    if (nodeProto && probeDiv) {
      ["nodeType", "textContent"].forEach(function (n) {
        add({ key: "Node.prototype." + n, owner: nodeProto, name: n, kind: "getter", instance: probeDiv });
      });
    }
    var elementProto = proto("Element");
    if (elementProto && probeDiv) {
      add({ key: "Element.prototype.tagName", owner: elementProto, name: "tagName", kind: "getter", instance: probeDiv });
      if (typeof elementProto.getAttribute === "function") {
        add({
          key: "Element.prototype.getAttribute",
          owner: elementProto,
          name: "getAttribute",
          kind: "method",
          instance: probeDiv,
          args: ["id"]
        });
      }
    }
    var targetProto = proto("EventTarget");
    if (targetProto && probeDiv && typeof targetProto.addEventListener === "function") {
      add({
        key: "EventTarget.prototype.addEventListener",
        owner: targetProto,
        name: "addEventListener",
        kind: "method",
        instance: probeDiv,
        args: ["am-probe", null]
      });
    }
    var htmlProto = proto("HTMLElement");
    if (htmlProto && probeDiv && typeof htmlProto.click === "function") {
      add({ key: "HTMLElement.prototype.click", owner: htmlProto, name: "click", kind: "method", instance: probeDiv, args: [] });
    }
    var docProto = proto("Document");
    if (docProto && typeof docProto.createElement === "function") {
      add({
        key: "Document.prototype.createElement",
        owner: docProto,
        name: "createElement",
        kind: "method",
        instance: document,
        args: ["div"]
      });
    }
    var elemProto = proto("Element");
    if (elemProto && typeof elemProto.getBoundingClientRect === "function") {
      add({
        key: "Element.prototype.getBoundingClientRect",
        owner: elemProto,
        name: "getBoundingClientRect",
        kind: "method",
        instance: document.createElement("div"),
        args: []
      });
    }
    add({ key: "Date.prototype.getTimezoneOffset", owner: Date.prototype, name: "getTimezoneOffset", kind: "method", instance: new Date(), args: [] });
    add({ key: "Function.prototype.toString", owner: Function.prototype, name: "toString", kind: "method", instance: function amSample() {}, args: [] });
    if (window.Intl && Intl.DateTimeFormat) {
      add({
        key: "Intl.DateTimeFormat.prototype.resolvedOptions",
        owner: Intl.DateTimeFormat.prototype,
        name: "resolvedOptions",
        kind: "method",
        instance: new Intl.DateTimeFormat(),
        args: []
      });
    }
    var mediaProto = proto("HTMLMediaElement");
    if (mediaProto) {
      add({
        key: "HTMLMediaElement.prototype.canPlayType",
        owner: mediaProto,
        name: "canPlayType",
        kind: "method",
        instance: document.createElement("video"),
        args: ["audio/mpeg"]
      });
    }
    var canvasProto = proto("HTMLCanvasElement");
    if (canvasProto) {
      var cv = document.createElement("canvas");
      cv.width = 8;
      cv.height = 8;
      add({ key: "HTMLCanvasElement.prototype.toDataURL", owner: canvasProto, name: "toDataURL", kind: "method", instance: cv, args: [] });
      var ctx = attempt(function () {
        return cv.getContext("2d");
      }, null);
      var ctxProto = proto("CanvasRenderingContext2D");
      if (ctx && ctxProto) {
        add({
          key: "CanvasRenderingContext2D.prototype.measureText",
          owner: ctxProto,
          name: "measureText",
          kind: "method",
          instance: ctx,
          args: ["sample"]
        });
      }
    }
    var glProto = proto("WebGLRenderingContext");
    var gl = attempt(function () {
      var g = document.createElement("canvas");
      return g.getContext("webgl") || g.getContext("experimental-webgl");
    }, null);
    if (gl && glProto) {
      add({ key: "WebGLRenderingContext.prototype.getParameter", owner: glProto, name: "getParameter", kind: "method", instance: gl, args: [0x1f00] });
    }
    var abProto = proto("AudioBuffer");
    var audioBuf = attempt(function () {
      var OAC = window.OfflineAudioContext || window.webkitOfflineAudioContext;
      if (!OAC) return null;
      return new OAC(1, 4096, 44100).createBuffer(1, 4096, 44100);
    }, null);
    if (audioBuf && abProto) {
      add({ key: "AudioBuffer.prototype.getChannelData", owner: abProto, name: "getChannelData", kind: "method", instance: audioBuf, args: [0] });
    }
    var perfProto = proto("Performance");
    if (perfProto && window.performance) {
      add({ key: "Performance.prototype.now", owner: perfProto, name: "now", kind: "method", instance: performance, args: [] });
    }
    if (typeof window.matchMedia === "function") {
      add({
        key: "window.matchMedia",
        owner: window,
        name: "matchMedia",
        kind: "method",
        instance: window,
        args: ["(min-width: 0px)"]
      });
    }
    if (typeof window.getComputedStyle === "function" && document.documentElement) {
      add({
        key: "window.getComputedStyle",
        owner: window,
        name: "getComputedStyle",
        kind: "method",
        instance: window,
        args: [document.documentElement]
      });
    }
    return list;
  }

  var NATIVE_TAIL = /\{\s*\[native code\]\s*\}\s*$/;

  function sortedKeys(list) {
    return list
      .map(function (k) {
        return typeof k === "symbol" ? k.toString() : String(k);
      })
      .sort();
  }

  function describeTarget(t) {
    var out = { key: t.key, kind: t.kind, status: "ok" };
    var desc;
    try {
      desc = Object.getOwnPropertyDescriptor(t.owner, t.name);
    } catch (e) {
      out.status = "error";
      out.reason = reasonOf(e);
      return out;
    }
    if (!desc) {
      out.status = "unsupported";
      out.reason = "not an own property of the interface prototype in this browser";
      return out;
    }
    var fn = t.kind === "getter" ? desc.get : desc.value;

    var d = {
      onPrototype: true,
      kind: desc.get || desc.set ? "accessor" : "data",
      enumerable: !!desc.enumerable,
      configurable: !!desc.configurable,
      writable: desc.get || desc.set ? null : !!desc.writable,
      hasGetter: !!desc.get,
      hasSetter: !!desc.set
    };
    if (t.instance && t.instance !== t.owner) {
      d.ownOnInstance = attempt(function () {
        return !!Object.getOwnPropertyDescriptor(t.instance, t.name);
      }, null);
    } else {
      d.ownOnInstance = null;
    }
    d.expectedKind = t.kind === "getter" ? "accessor" : "data";
    d.kindAsExpected = d.kind === d.expectedKind;
    out.descriptor = d;

    if (typeof fn !== "function") {
      out.status = "unsupported";
      out.reason = "the descriptor carries no function for a " + t.kind;
      return out;
    }
    out.prototypeIsFunctionPrototype = attempt(function () {
      return Object.getPrototypeOf(fn) === Function.prototype;
    }, null);

    var ts = { status: "ok" };
    try {
      ts.source = Function.prototype.toString.call(fn);
    } catch (e) {
      ts.status = "error";
      ts.reason = reasonOf(e);
    }
    if (ts.status === "ok") {
      ts.native = NATIVE_TAIL.test(ts.source);
      ts.length = ts.source.length;
      ts.fnName = attempt(function () {
        return fn.name;
      }, null);
      ts.fnLength = attempt(function () {
        return fn.length;
      }, null);
      var own = attempt(function () {
        return fn.toString();
      }, null);
      ts.selfAgrees = own === ts.source;
      ts.hasOwnToString = attempt(function () {
        return Object.prototype.hasOwnProperty.call(fn, "toString");
      }, null);
    }
    out.tostring = ts;

    var ok = { status: "ok" };
    try {
      var all = Reflect.ownKeys(fn);
      ok.ownKeys = sortedKeys(
        all.filter(function (k) {
          return typeof k !== "symbol";
        })
      );
      ok.symbolKeys = sortedKeys(
        all.filter(function (k) {
          return typeof k === "symbol";
        })
      );
      ok.getOwnPropertyNames = sortedKeys(Object.getOwnPropertyNames(fn));
      ok.descriptors = sortedKeys(Object.keys(Object.getOwnPropertyDescriptors(fn)));
      ok.kind = t.kind;
      ok.agree =
        JSON.stringify(ok.ownKeys) === JSON.stringify(ok.getOwnPropertyNames) &&
        JSON.stringify(ok.getOwnPropertyNames) === JSON.stringify(ok.descriptors);
    } catch (e) {
      ok.status = "error";
      ok.reason = reasonOf(e);
    }
    out.ownkeys = ok;

    var rc = { status: "ok", skipped: false };
    var sane = true;
    if (t.instance) {
      try {
        if (t.kind === "getter") fn.call(t.instance);
        else fn.apply(t.instance, t.args || []);
      } catch (e) {
        sane = false;
        rc.status = "unsupported";
        rc.skipped = true;
        rc.reason = "the member did not run with its own receiver: " + reasonOf(e);
      }
    }
    if (sane) {
      try {
        var r = fn.apply({}, t.kind === "getter" ? [] : t.args || []);
        rc.threw = false;
        rc.isTypeError = false;
        rc.resultType = typeof r;
      } catch (e2) {
        rc.threw = true;
        rc.name = attempt(function () {
          return e2.name;
        }, null);
        rc.isTypeError = e2 instanceof TypeError;
        rc.message = reasonOf(e2);
      }
    }
    out.receiver = rc;
    return out;
  }

  function collectNatives() {
    if (nativeCache) return nativeCache;
    var targets = nativeTargets();
    if (!targets.length) unsupported("no interface member was reachable");
    var results = [];
    for (var i = 0; i < targets.length; i++) {
      try {
        results.push(describeTarget(targets[i]));
      } catch (e) {
        results.push({ key: targets[i].key, kind: targets[i].kind, status: "error", reason: reasonOf(e) });
      }
    }
    var globals = {
      toStringOfToString: attempt(function () {
        return Function.prototype.toString.call(Function.prototype.toString);
      }, null),
      toStringOwnKeys: attempt(function () {
        return sortedKeys(Reflect.ownKeys(Function.prototype.toString));
      }, null),
      toStringDescriptor: attempt(function () {
        var d = Object.getOwnPropertyDescriptor(Function.prototype, "toString");
        return d ? { enumerable: !!d.enumerable, configurable: !!d.configurable, writable: !!d.writable } : null;
      }, null)
    };
    nativeCache = { globals: globals, targets: results };
    return nativeCache;
  }

  function amFacts(tag, g) {
    var out = { tag: tag };
    function get(name, fn) {
      try {
        var v = fn();
        out[name] = v === undefined ? null : v;
      } catch (e) {
        out[name] = null;
      }
    }
    var nav = g.navigator;
    get("scope", function () {
      return g.constructor && g.constructor.name;
    });
    get("userAgent", function () {
      return nav.userAgent;
    });
    get("platform", function () {
      return nav.platform;
    });
    get("hardwareConcurrency", function () {
      return nav.hardwareConcurrency;
    });
    get("deviceMemoryPresent", function () {
      return "deviceMemory" in nav;
    });
    get("deviceMemory", function () {
      return nav.deviceMemory;
    });
    get("language", function () {
      return nav.language;
    });
    get("languages", function () {
      return nav.languages ? Array.prototype.slice.call(nav.languages) : null;
    });
    get("webdriver", function () {
      return "webdriver" in nav ? nav.webdriver : null;
    });
    get("uaDataPlatform", function () {
      return nav.userAgentData ? nav.userAgentData.platform : null;
    });
    get("uaDataMobile", function () {
      return nav.userAgentData ? nav.userAgentData.mobile : null;
    });
    get("timeZone", function () {
      return new g.Intl.DateTimeFormat().resolvedOptions().timeZone;
    });
    get("locale", function () {
      return new g.Intl.DateTimeFormat().resolvedOptions().locale;
    });
    get("calendar", function () {
      return new g.Intl.DateTimeFormat().resolvedOptions().calendar;
    });
    get("timezoneOffsetMinutes", function () {
      return new g.Date().getTimezoneOffset();
    });
    get("isSecureContext", function () {
      return g.isSecureContext;
    });
    get("origin", function () {
      return g.location ? g.location.origin : null;
    });
    get("hasWindow", function () {
      return typeof g.window !== "undefined";
    });
    return out;
  }

  var scopeCache = null;

  function collectWorker() {
    return new Promise(function (resolve, reject) {
      if (typeof Worker !== "function" || typeof Blob !== "function" || !window.URL || !URL.createObjectURL) {
        reject(Unsupported("dedicated workers or blob URLs are not available in this context"));
        return;
      }
      var src =
        String(amFacts) +
        "\nself.onmessage=function(){try{self.postMessage({ok:true,data:amFacts('worker',self)});}" +
        "catch(e){self.postMessage({ok:false,error:String(e&&e.message||e)});}};";
      var url, w;
      try {
        url = URL.createObjectURL(new Blob([src], { type: "text/javascript" }));
        w = new Worker(url);
      } catch (e) {
        reject(Unsupported("the worker could not be created: " + reasonOf(e)));
        return;
      }
      var settled = false;
      function done(fn, arg) {
        if (settled) return;
        settled = true;
        try {
          w.terminate();
        } catch (e) {}
        try {
          URL.revokeObjectURL(url);
        } catch (e) {}
        fn(arg);
      }
      w.onmessage = function (ev) {
        if (ev.data && ev.data.ok) done(resolve, ev.data.data);
        else done(reject, Unsupported("the worker reported: " + (ev.data && ev.data.error)));
      };
      w.onerror = function (ev) {
        done(reject, Unsupported("the worker did not start: " + ((ev && ev.message) || "blocked or failed")));
      };
      try {
        w.postMessage("collect");
      } catch (e) {
        done(reject, Unsupported("the worker did not accept a message: " + reasonOf(e)));
      }
      setTimeout(function () {
        done(reject, Unsupported("the worker did not answer within 5000 ms"));
      }, 5000);
    });
  }

  function collectIframe() {
    return new Promise(function (resolve, reject) {
      var f;
      try {
        f = document.createElement("iframe");
        f.setAttribute("title", "probe frame");
        f.setAttribute("aria-hidden", "true");
        f.setAttribute("tabindex", "-1");
        f.style.cssText = "position:absolute;left:-9999px;top:-9999px;width:0;height:0;border:0";
        document.body.appendChild(f);
      } catch (e) {
        reject(Unsupported("a frame could not be created: " + reasonOf(e)));
        return;
      }
      setTimeout(function () {
        var settledValue = null;
        var failure = null;
        try {
          var g = f.contentWindow;
          if (!g || !g.navigator) failure = Unsupported("the frame's realm was not reachable");
          else settledValue = amFacts("iframe", g);
        } catch (e) {
          failure = Unsupported("the frame's realm could not be read: " + reasonOf(e));
        }
        try {
          if (f.parentNode) f.parentNode.removeChild(f);
        } catch (e2) {}
        if (failure) reject(failure);
        else resolve(settledValue);
      }, 60);
    });
  }

  function collectNestedWorker() {
    return new Promise(function (resolve, reject) {
      if (typeof Worker !== "function" || typeof Blob !== "function" || !window.URL || !URL.createObjectURL) {
        reject(Unsupported("dedicated workers or blob URLs are not available in this context"));
        return;
      }
      var inner =
        String(amFacts) +
        ";self.onmessage=function(){try{self.postMessage({ok:true,data:amFacts('workerNested',self)});}" +
        "catch(e){self.postMessage({ok:false,error:String(e&&e.message||e)});}};";
      var outer =
        "self.onmessage=function(){try{" +
        "var u=URL.createObjectURL(new Blob([" + JSON.stringify(inner) + "],{type:'text/javascript'}));" +
        "var w=new Worker(u);" +
        "w.onmessage=function(ev){try{w.terminate();URL.revokeObjectURL(u);}catch(e){}self.postMessage(ev.data);};" +
        "w.onerror=function(){try{w.terminate();URL.revokeObjectURL(u);}catch(e){}" +
        "self.postMessage({ok:false,error:'the inner worker did not start'});};" +
        "w.postMessage('collect');" +
        "}catch(e){self.postMessage({ok:false,error:String(e&&e.message||e)});}};";
      var url, w;
      try {
        url = URL.createObjectURL(new Blob([outer], { type: "text/javascript" }));
        w = new Worker(url);
      } catch (e) {
        reject(Unsupported("the outer worker could not be created: " + reasonOf(e)));
        return;
      }
      var settled = false;
      function done(fn, arg) {
        if (settled) return;
        settled = true;
        try {
          w.terminate();
        } catch (e) {}
        try {
          URL.revokeObjectURL(url);
        } catch (e) {}
        fn(arg);
      }
      w.onmessage = function (ev) {
        if (ev.data && ev.data.ok) done(resolve, ev.data.data);
        else done(reject, Unsupported("the nested worker reported: " + (ev.data && ev.data.error)));
      };
      w.onerror = function (ev) {
        done(reject, Unsupported("the outer worker did not start: " + ((ev && ev.message) || "blocked or failed")));
      };
      setTimeout(function () {
        done(reject, Unsupported("the nested worker did not answer in time"));
      }, 4000);
      try {
        w.postMessage("collect");
      } catch (e) {
        done(reject, Unsupported("the outer worker did not accept a message: " + reasonOf(e)));
      }
    });
  }

  function collectFrameVariant(label, prepare) {
    return new Promise(function (resolve, reject) {
      var f, cleanupUrl = null;
      try {
        f = document.createElement("iframe");
        f.setAttribute("title", "probe frame");
        f.setAttribute("aria-hidden", "true");
        f.setAttribute("tabindex", "-1");
        f.style.cssText = "position:absolute;left:-9999px;top:-9999px;width:0;height:0;border:0";
        cleanupUrl = prepare(f);
        document.body.appendChild(f);
      } catch (e) {
        reject(Unsupported("a frame could not be created: " + reasonOf(e)));
        return;
      }
      setTimeout(function () {
        var settledValue = null;
        var failure = null;
        try {
          var g = f.contentWindow;
          if (!g || !g.navigator) failure = Unsupported("the frame's realm was not reachable");
          else settledValue = amFacts(label, g);
        } catch (e) {
          failure = Unsupported("the frame's realm could not be read: " + reasonOf(e));
        }
        try {
          if (f.parentNode) f.parentNode.removeChild(f);
        } catch (e2) {}
        try {
          if (cleanupUrl) URL.revokeObjectURL(cleanupUrl);
        } catch (e3) {}
        if (failure) reject(failure);
        else resolve(settledValue);
      }, 80);
    });
  }

  function collectSrcdocIframe() {
    return collectFrameVariant("iframeSrcdoc", function (f) {
      f.setAttribute("srcdoc", "<!doctype html><title>probe</title>");
      return null;
    });
  }

  function collectBlobIframe() {
    return collectFrameVariant("iframeBlob", function (f) {
      if (typeof Blob !== "function" || !window.URL || !URL.createObjectURL) {
        throw Unsupported("blob URLs are not available in this context");
      }
      var u = URL.createObjectURL(new Blob(["<!doctype html><title>probe</title>"], { type: "text/html" }));
      f.setAttribute("src", u);
      return u;
    });
  }

  function collectScopes() {
    if (scopeCache) return scopeCache;
    scopeCache = Promise.all([
      Promise.resolve()
        .then(function () {
          return { status: "ok", value: amFacts("main", window) };
        })
        .catch(function (e) {
          return { status: "error", reason: reasonOf(e) };
        }),
      collectWorker()
        .then(function (v) {
          return { status: "ok", value: v };
        })
        .catch(function (e) {
          return { status: e && e.amUnsupported ? "unsupported" : "error", reason: reasonOf(e) };
        }),
      collectIframe()
        .then(function (v) {
          return { status: "ok", value: v };
        })
        .catch(function (e) {
          return { status: e && e.amUnsupported ? "unsupported" : "error", reason: reasonOf(e) };
        }),
      collectNestedWorker()
        .then(function (v) {
          return { status: "ok", value: v };
        })
        .catch(function (e) {
          return { status: e && e.amUnsupported ? "unsupported" : "error", reason: reasonOf(e) };
        }),
      collectSrcdocIframe()
        .then(function (v) {
          return { status: "ok", value: v };
        })
        .catch(function (e) {
          return { status: e && e.amUnsupported ? "unsupported" : "error", reason: reasonOf(e) };
        }),
      collectBlobIframe()
        .then(function (v) {
          return { status: "ok", value: v };
        })
        .catch(function (e) {
          return { status: e && e.amUnsupported ? "unsupported" : "error", reason: reasonOf(e) };
        })
    ]).then(function (r) {
      return {
        main: r[0],
        worker: r[1],
        iframe: r[2],
        workerNested: r[3],
        iframeSrcdoc: r[4],
        iframeBlob: r[5]
      };
    });
    return scopeCache;
  }

  function scopeSlice(name) {
    return collectScopes().then(function (all) {
      var s = all[name];
      if (s.status === "ok") return s.value;
      if (s.status === "unsupported") unsupported(s.reason);
      throw new Error(s.reason);
    });
  }

  function mq(q) {
    return window.matchMedia(q).matches;
  }

  function bisectFloat(pred, lo, hi, iterations) {
    if (!attempt(function () {
      return pred(lo);
    }, false)) {
      return null;
    }
    if (attempt(function () {
      return pred(hi);
    }, true)) {
      return null;
    }
    for (var i = 0; i < iterations; i++) {
      var mid = (lo + hi) / 2;
      if (pred(mid)) lo = mid;
      else hi = mid;
    }
    return Math.round(((lo + hi) / 2) * 100000) / 100000;
  }

  function bisectInt(pred, lo, hi) {
    if (!attempt(function () {
      return pred(lo);
    }, false)) {
      return null;
    }
    if (attempt(function () {
      return pred(hi);
    }, true)) {
      return null;
    }
    while (hi - lo > 1) {
      var mid = Math.floor((lo + hi) / 2);
      if (pred(mid)) lo = mid;
      else hi = mid;
    }
    return lo;
  }

  function firstMatching(queries) {
    for (var i = 0; i < queries.length; i++) {
      var q = queries[i][1];
      var hit = attempt(function () {
        return mq(q);
      }, false);
      if (hit) return queries[i][0];
    }
    return null;
  }

  var CODECS = [
    ["H.264 baseline, MP4", 'video/mp4; codecs="avc1.42E01E"', "video"],
    ["H.264 high, MP4", 'video/mp4; codecs="avc1.640028"', "video"],
    ["HEVC, MP4", 'video/mp4; codecs="hev1.1.6.L93.B0"', "video"],
    ["VP8, WebM", 'video/webm; codecs="vp8"', "video"],
    ["VP9, WebM", 'video/webm; codecs="vp9"', "video"],
    ["AV1, WebM", 'video/webm; codecs="av01.0.04M.08"', "video"],
    ["Theora, Ogg", 'video/ogg; codecs="theora"', "video"],
    ["AAC-LC, MP4", 'audio/mp4; codecs="mp4a.40.2"', "audio"],
    ["MP3", "audio/mpeg", "audio"],
    ["Opus, WebM", 'audio/webm; codecs="opus"', "audio"],
    ["Vorbis, Ogg", 'audio/ogg; codecs="vorbis"', "audio"],
    ["FLAC", "audio/flac", "audio"]
  ];

  var RECORDER_PREFIX = "MediaRecorder ";

  var RECORDER_TYPES = [
    "video/webm",
    "video/webm;codecs=vp8",
    "video/webm;codecs=vp9",
    "video/webm;codecs=h264",
    "video/mp4",
    "audio/webm;codecs=opus",
    "audio/ogg;codecs=opus"
  ];

  var RESIDUE_PATTERNS = [
    ["random-suffixed injected symbol", /^\$?cdc_[A-Za-z0-9]{16,}$/],
    ["hook verb form", /^__[A-Za-z]+_(evaluate|unwrapped|script_fn|script_func|script_function)$/],
    ["webdriver-prefixed global", /^__(web)?driver[_A-Z]/],
    ["automation-controller global", /^__domAutomation/]
  ];

  function scanKeys(keys) {
    var hits = [];
    for (var i = 0; i < keys.length; i++) {
      for (var p = 0; p < RESIDUE_PATTERNS.length; p++) {
        if (RESIDUE_PATTERNS[p][1].test(keys[i])) {
          hits.push({ name: keys[i], pattern: RESIDUE_PATTERNS[p][0] });
          break;
        }
      }
    }
    return hits;
  }

  function prefixedKeys(keys) {
    return keys
      .filter(function (k) {
        return /^[_$]/.test(k) && k !== "$" && k !== "_";
      })
      .slice(0, 60);
  }

  var probes = [
    {
      id: "font.resolved",
      group: "font",
      run: function () {
        var input = fontInput();
        var d = input.data;
        var m = measurer();
        var families = {};
        var resolvedCount = 0;
        for (var i = 0; i < d.families.length; i++) {
          var f = d.families[i];
          var w = differsFromGenerics(m, f, d.bases, d.measureString);
          families[f] = { width: w, check: fontsCheck(f) };
          if (w) resolvedCount += 1;
        }
        meta.font = {
          inputSource: input.source,
          inputSources: input.sources || null,
          measureString: d.measureString,
          bases: d.bases,
          probed: d.families.length,
          resolvedCount: resolvedCount
        };
        return families;
      }
    },
    {
      id: "font.coverage",
      group: "font",
      run: function () {
        var input = fontInput();
        var d = input.data;
        if (!d.coverage || !d.coverage.length) unsupported("no coverage input was supplied");
        var m = measurer();
        var out = {};
        var order = [];
        for (var i = 0; i < d.coverage.length; i++) {
          var c = d.coverage[i];
          var covers = differsFromGenerics(m, c.family, d.bases, c.text);
          var width = differsFromGenerics(m, c.family, d.bases, d.measureString);
          out[c.family] = {
            covers: covers,
            width: width,
            codepoints: c.codepoints,
            scripts: c.scripts,
            agree: covers === width
          };
          order.push(c.family);
        }
        meta.fontCoverage = { inputSource: (input.sources && input.sources.coverage) || input.source, order: order };
        return out;
      }
    },
    {
      id: "font.controls",
      group: "font",
      run: function () {
        var input = fontInput();
        var d = input.data;
        if (!d.controls || !d.controls.length) unsupported("no control name was supplied");
        var m = measurer();
        var out = {};
        var widthResolved = [];
        var checkTrue = [];
        var checkAvailable = false;
        for (var i = 0; i < d.controls.length; i++) {
          var name = d.controls[i];
          var w = differsFromGenerics(m, name, d.bases, d.measureString);
          var ch = fontsCheck(name);
          out[name] = { width: w, check: ch };
          if (w) widthResolved.push(name);
          if (ch === true) checkTrue.push(name);
          if (ch !== null) checkAvailable = true;
        }
        meta.fontControls = {
          inputSource: (input.sources && input.sources.controls) || input.source,
          order: d.controls.slice(),
          count: d.controls.length,
          widthResolved: widthResolved,
          checkAvailable: checkAvailable,
          checkTrueCount: checkTrue.length
        };
        return out;
      }
    },
    {
      id: "native.tostring",
      group: "native",
      run: function () {
        var n = collectNatives();
        meta.native = { globals: n.globals, order: n.targets.map(function (t) { return t.key; }) };
        var out = {};
        n.targets.forEach(function (t) {
          var rec = { status: t.status };
          if (t.reason) rec.reason = t.reason;
          if (t.tostring) {
            if (t.tostring.status === "ok") {
              rec.source = t.tostring.source;
              rec.native = t.tostring.native;
              rec.fnName = t.tostring.fnName;
              rec.fnLength = t.tostring.fnLength;
              rec.chars = t.tostring.length;
              rec.selfAgrees = t.tostring.selfAgrees;
              rec.hasOwnToString = t.tostring.hasOwnToString;
            } else if (t.tostring.reason) {
              rec.reason = t.tostring.reason;
            }
          }
          out[t.key] = rec;
        });
        return out;
      }
    },
    {
      id: "native.ownkeys",
      group: "native",
      run: function () {
        var n = collectNatives();
        var out = {};
        n.targets.forEach(function (t) {
          var rec = { status: t.status };
          if (t.reason) rec.reason = t.reason;
          if (t.ownkeys && t.ownkeys.status === "ok") {
            rec.kind = t.ownkeys.kind;
            rec.ownKeys = t.ownkeys.ownKeys;
            rec.symbolKeys = t.ownkeys.symbolKeys;
            rec.getOwnPropertyNames = t.ownkeys.getOwnPropertyNames;
            rec.descriptors = t.ownkeys.descriptors;
            rec.agree = t.ownkeys.agree;
          } else if (t.ownkeys && t.ownkeys.reason) {
            rec.reason = t.ownkeys.reason;
          }
          out[t.key] = rec;
        });
        return out;
      }
    },
    {
      id: "native.descriptor",
      group: "native",
      run: function () {
        var n = collectNatives();
        var out = {};
        n.targets.forEach(function (t) {
          var rec = { status: t.status, declaredAs: t.kind };
          if (t.reason) rec.reason = t.reason;
          var d = t.descriptor;
          if (d) {
            rec.onPrototype = d.onPrototype;
            rec.kind = d.kind;
            rec.expectedKind = d.expectedKind;
            rec.kindAsExpected = d.kindAsExpected;
            rec.enumerable = d.enumerable;
            rec.configurable = d.configurable;
            rec.writable = d.writable;
            rec.hasGetter = d.hasGetter;
            rec.hasSetter = d.hasSetter;
            rec.shadowedOnInstance = d.ownOnInstance;
          }
          if (t.prototypeIsFunctionPrototype !== undefined) {
            rec.prototypeIsFunctionPrototype = t.prototypeIsFunctionPrototype;
          }
          out[t.key] = rec;
        });
        return out;
      }
    },
    {
      id: "native.receiver",
      group: "native",
      run: function () {
        var n = collectNatives();
        var out = {};
        n.targets.forEach(function (t) {
          var rec = { status: t.status };
          if (t.reason) rec.reason = t.reason;
          var r = t.receiver;
          if (r) {
            rec.skipped = !!r.skipped;
            if (!r.skipped) {
              rec.threw = r.threw;
              rec.isTypeError = r.isTypeError;
              if (r.name) rec.name = r.name;
              if (r.message) rec.message = r.message;
              if (r.resultType) rec.resultType = r.resultType;
            } else if (r.reason) {
              rec.reason = r.reason;
            }
          } else {
            rec.skipped = true;
          }
          out[t.key] = rec;
        });
        return out;
      }
    },
    {
      id: "scope.main",
      group: "scope",
      run: function () {
        return scopeSlice("main");
      }
    },
    {
      id: "scope.worker",
      group: "scope",
      run: function () {
        return scopeSlice("worker");
      }
    },
    {
      id: "scope.workerNested",
      group: "scope",
      run: function () {
        return scopeSlice("workerNested");
      }
    },
    {
      id: "scope.iframeSrcdoc",
      group: "scope",
      run: function () {
        return scopeSlice("iframeSrcdoc");
      }
    },
    {
      id: "scope.iframeBlob",
      group: "scope",
      run: function () {
        return scopeSlice("iframeBlob");
      }
    },
    {
      id: "scope.iframe",
      group: "scope",
      run: function () {
        return scopeSlice("iframe");
      }
    },
    {
      id: "scope.availability",
      group: "scope",
      run: function () {
        return collectScopes().then(function (all) {
          var out = {};
          ["main", "worker", "iframe", "workerNested", "iframeSrcdoc", "iframeBlob"].forEach(function (k) {
            out[k] = { created: all[k].status === "ok", status: all[k].status, reason: all[k].reason || null };
          });
          return out;
        });
      }
    },
    {
      id: "gpu.renderer",
      group: "gpu",
      run: function () {
        var canvas = attempt(function () {
          return document.createElement("canvas");
        }, null);
        if (!canvas || typeof canvas.getContext !== "function") {
          unsupported("this document cannot create a canvas element");
        }
        var gl = attempt(function () {
          return canvas.getContext("webgl") || canvas.getContext("experimental-webgl");
        }, null);
        if (!gl) unsupported("no WebGL rendering context was granted");
        var info = attempt(function () {
          return gl.getExtension("WEBGL_debug_renderer_info");
        }, null);
        var param = function (name) {
          return attempt(function () {
            var value = gl.getParameter(gl[name]);
            return typeof value === "string" ? value : null;
          }, null);
        };
        var unmasked = function (key) {
          if (!info) return null;
          return attempt(function () {
            return gl.getParameter(info[key]);
          }, null);
        };
        return {
          vendor: param("VENDOR"),
          renderer: param("RENDERER"),
          version: param("VERSION"),
          shadingLanguageVersion: param("SHADING_LANGUAGE_VERSION"),
          unmaskedVendor: unmasked("UNMASKED_VENDOR_WEBGL"),
          unmaskedRenderer: unmasked("UNMASKED_RENDERER_WEBGL"),
          debugRendererInfo: !!info
        };
      }
    },
    {
      id: "gpu.adapter",
      group: "gpu",
      run: function () {
        var base = { present: !!(navigator.gpu && typeof navigator.gpu.requestAdapter === "function"), secureContext: !!isSecureContext, variants: {} };
        if (!base.present) return base;
        var requests = [
          ["default", undefined],
          ["highPerformance", { powerPreference: "high-performance" }],
          ["lowPower", { powerPreference: "low-power" }],
          ["fallback", { forceFallbackAdapter: true }]
        ];
        var chain = Promise.resolve();
        requests.forEach(function (r) {
          chain = chain.then(function () {
            return withTimeout(Promise.resolve(navigator.gpu.requestAdapter(r[1])), 5000, "requestAdapter " + r[0])
              .then(function (adapter) {
                if (!adapter) {
                  base.variants[r[0]] = { adapter: false };
                  return;
                }
                var info = adapter.info || {};
                base.variants[r[0]] = {
                  adapter: true,
                  vendor: typeof info.vendor === "string" ? info.vendor : null,
                  architecture: typeof info.architecture === "string" ? info.architecture : null,
                  device: typeof info.device === "string" ? info.device : null,
                  maxTextureDimension2D: adapter.limits ? adapter.limits.maxTextureDimension2D : null
                };
              })
              .catch(function (e) {
                base.variants[r[0]] = { error: reasonOf(e) };
              });
          });
        });
        return chain.then(function () {
          return base;
        });
      }
    },
    {
      id: "webrtc.ice",
      group: "webrtc",
      run: function () {
        var Ctor = window.RTCPeerConnection || window.webkitRTCPeerConnection;
        if (typeof Ctor !== "function") unsupported("no peer connection interface is available");
        var pc;
        try {
          pc = new Ctor({ iceServers: [] });
        } catch (e) {
          unsupported("a peer connection could not be created: " + reasonOf(e));
        }
        var sawStates = [];
        var candidateTypes = [];
        function noteState() {
          var st = attempt(function () {
            return pc.iceGatheringState;
          }, null);
          if (st && sawStates[sawStates.length - 1] !== st) sawStates.push(st);
        }
        noteState();
        var settle;
        var finished = new Promise(function (resolve) {
          settle = resolve;
        });
        pc.onicegatheringstatechange = function () {
          noteState();
          if (attempt(function () {
            return pc.iceGatheringState;
          }, null) === "complete") settle(true);
        };
        pc.onicecandidate = function (ev) {
          if (!ev.candidate) {
            noteState();
            settle(true);
            return;
          }
          var t = attempt(function () {
            return ev.candidate.type;
          }, null);
          candidateTypes.push(typeof t === "string" ? t : "unknown");
        };
        try {
          pc.createDataChannel("am-probe");
        } catch (e) {}
        var offered = Promise.resolve()
          .then(function () {
            return pc.createOffer();
          })
          .then(function (offer) {
            return pc.setLocalDescription(offer);
          })
          .catch(function (e) {
            return null;
          });
        return offered
          .then(function () {
            return withTimeout(finished, 5000, "ice gathering").then(
              function () {
                return false;
              },
              function () {
                return true;
              }
            );
          })
          .then(function (timedOut) {
            noteState();
            try {
              pc.close();
            } catch (e) {}
            return {
              finalState: sawStates.length ? sawStates[sawStates.length - 1] : null,
              timedOut: timedOut,
              sawStates: sawStates,
              candidateTypes: candidateTypes
            };
          });
      }
    },
    {
      id: "canvas.serial",
      group: "canvas",
      run: function () {
        var W = 220, H = 90;
        var canvas = attempt(function () {
          return document.createElement("canvas");
        }, null);
        if (!canvas || typeof canvas.getContext !== "function") unsupported("this document cannot create a canvas element");
        canvas.width = W;
        canvas.height = H;
        var ctx = attempt(function () {
          return canvas.getContext("2d");
        }, null);
        if (!ctx) unsupported("no two-dimensional drawing context was granted");

        var grad = ctx.createLinearGradient(0, 0, W, H);
        grad.addColorStop(0, "#1b3a6b");
        grad.addColorStop(0.5, "#c8452a");
        grad.addColorStop(1, "#e9e4d0");
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, W, H);
        ctx.font = "24px serif";
        ctx.fillStyle = "#ffffff";
        ctx.fillText("anti-mage 0123", 8, 40);
        ctx.strokeStyle = "rgba(0,0,0,0.55)";
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.arc(W - 30, H - 30, 20, 0, Math.PI * 1.5);
        ctx.stroke();

        function fnv1a(bytes) {
          var h = 0x811c9dc5;
          for (var i = 0; i < bytes.length; i++) {
            h = h ^ bytes[i];
            h = (h * 0x01000193) >>> 0;
          }
          var hex = h.toString(16);
          while (hex.length < 8) hex = "0" + hex;
          return hex;
        }
        function hashOf(buffer) {
          return fnv1a(new Uint8Array(buffer));
        }
        function pixelsOf(c) {
          return c.getContext("2d").getImageData(0, 0, W, H).data;
        }
        function decodeHash(src, revoke) {
          return new Promise(function (resolve) {
            var img = new Image();
            img.onload = function () {
              var out = document.createElement("canvas");
              out.width = W;
              out.height = H;
              try {
                out.getContext("2d").drawImage(img, 0, 0);
                resolve(fnv1a(pixelsOf(out)));
              } catch (e) {
                resolve(null);
              }
              if (revoke) attempt(function () {
                return URL.revokeObjectURL(src);
              }, null);
            };
            img.onerror = function () {
              resolve(null);
              if (revoke) attempt(function () {
                return URL.revokeObjectURL(src);
              }, null);
            };
            img.src = src;
          });
        }

        var out = { dataUrlAvailable: false, blobAvailable: false, rawHash: fnv1a(pixelsOf(canvas)) };

        var dataUrl = attempt(function () {
          return canvas.toDataURL("image/png");
        }, null);
        var dataStep = Promise.resolve();
        if (typeof dataUrl === "string" && dataUrl.indexOf("data:image/png") === 0) {
          out.dataUrlAvailable = true;
          dataStep = fetch(dataUrl)
            .then(function (r) {
              return r.arrayBuffer();
            })
            .then(function (buf) {
              out.dataUrlHash = hashOf(buf);
              out.dataUrlLength = buf.byteLength;
              return decodeHash(dataUrl, false);
            })
            .then(function (h) {
              out.dataUrlPixelsMatchRaw = h !== null && h === out.rawHash;
            })
            .catch(function (e) {
              out.dataUrlAvailable = false;
            });
        }

        var blobStep = dataStep.then(function () {
          if (typeof canvas.toBlob !== "function") return null;
          return withTimeout(
            new Promise(function (resolve) {
              canvas.toBlob(resolve, "image/png");
            }),
            5000,
            "canvas.toBlob"
          ).catch(function (e) {
            return null;
          });
        });

        return blobStep
          .then(function (blob) {
            if (!blob || typeof blob.arrayBuffer !== "function") return null;
            out.blobAvailable = true;
            return blob.arrayBuffer().then(function (buf) {
              out.blobHash = hashOf(buf);
              out.blobLength = buf.byteLength;
              return decodeHash(URL.createObjectURL(blob), true);
            });
          })
          .then(function (h) {
            if (out.blobAvailable) out.blobPixelsMatchRaw = h !== null && h === out.rawHash;
            return out;
          })
          .catch(function (e) {
            return out;
          });
      }
    },
    {
      id: "geom.screen",
      group: "geom",
      run: function () {
        if (typeof screen === "undefined") unsupported("screen is not exposed");
        var o = attempt(function () {
          return screen.orientation;
        }, null);
        return {
          width: screen.width,
          height: screen.height,
          availWidth: screen.availWidth,
          availHeight: screen.availHeight,
          colorDepth: screen.colorDepth,
          pixelDepth: screen.pixelDepth,
          innerWidth: window.innerWidth,
          innerHeight: window.innerHeight,
          outerWidth: window.outerWidth,
          outerHeight: window.outerHeight,
          screenX: window.screenX,
          screenY: window.screenY,
          screenLeft: attempt(function () {
            return window.screenLeft;
          }, null),
          screenTop: attempt(function () {
            return window.screenTop;
          }, null),
          devicePixelRatio: window.devicePixelRatio,
          orientationType: o ? o.type : null,
          orientationAngle: o ? o.angle : null,
          visualViewportScale: attempt(function () {
            return window.visualViewport ? window.visualViewport.scale : null;
          }, null)
        };
      }
    },
    {
      id: "geom.css",
      group: "geom",
      run: function () {
        if (typeof window.matchMedia !== "function") unsupported("matchMedia is not available");
        if (!attempt(function () {
          return mq("(min-width: 0px)");
        }, false)) {
          unsupported("media queries did not evaluate");
        }
        var dpr = window.devicePixelRatio;
        return {
          resolution: bisectFloat(function (v) {
            return mq("(min-resolution: " + v + "dppx)");
          }, 0.02, 32, 34),
          webkitDevicePixelRatio: bisectFloat(function (v) {
            return mq("(-webkit-min-device-pixel-ratio: " + v + ")");
          }, 0.02, 32, 34),
          deviceWidthPx: bisectInt(function (v) {
            return mq("(min-device-width: " + v + "px)");
          }, 1, 40000),
          deviceHeightPx: bisectInt(function (v) {
            return mq("(min-device-height: " + v + "px)");
          }, 1, 40000),
          viewportWidthPx: bisectInt(function (v) {
            return mq("(min-width: " + v + "px)");
          }, 1, 40000),
          viewportHeightPx: bisectInt(function (v) {
            return mq("(min-height: " + v + "px)");
          }, 1, 40000),
          colorBitsPerComponent: bisectInt(function (v) {
            return mq("(min-color: " + v + ")");
          }, 1, 64),
          monochromeBits: bisectInt(function (v) {
            return mq("(min-monochrome: " + v + ")");
          }, 1, 64),
          exactResolutionMatchesDpr: attempt(function () {
            return mq("(resolution: " + dpr + "dppx)");
          }, null),
          exactWebkitMatchesDpr: attempt(function () {
            return mq("(-webkit-device-pixel-ratio: " + dpr + ")");
          }, null),
          orientation: firstMatching([["portrait", "(orientation: portrait)"], ["landscape", "(orientation: landscape)"]]),
          pointer: firstMatching([["fine", "(pointer: fine)"], ["coarse", "(pointer: coarse)"], ["none", "(pointer: none)"]]),
          anyPointer: firstMatching([["fine", "(any-pointer: fine)"], ["coarse", "(any-pointer: coarse)"], ["none", "(any-pointer: none)"]]),
          hover: firstMatching([["hover", "(hover: hover)"], ["none", "(hover: none)"]]),
          anyHover: firstMatching([["hover", "(any-hover: hover)"], ["none", "(any-hover: none)"]]),
          prefersColorScheme: firstMatching([["dark", "(prefers-color-scheme: dark)"], ["light", "(prefers-color-scheme: light)"]]),
          prefersReducedMotion: firstMatching([["reduce", "(prefers-reduced-motion: reduce)"], ["no-preference", "(prefers-reduced-motion: no-preference)"]]),
          forcedColors: firstMatching([["active", "(forced-colors: active)"], ["none", "(forced-colors: none)"]]),
          displayMode: firstMatching([
            ["fullscreen", "(display-mode: fullscreen)"],
            ["standalone", "(display-mode: standalone)"],
            ["minimal-ui", "(display-mode: minimal-ui)"],
            ["browser", "(display-mode: browser)"]
          ]),
          scripting: firstMatching([
            ["enabled", "(scripting: enabled)"],
            ["initial-only", "(scripting: initial-only)"],
            ["none", "(scripting: none)"]
          ])
        };
      }
    },
    {
      id: "time.zone",
      group: "time",
      run: function () {
        if (!window.Intl || !Intl.DateTimeFormat) unsupported("Intl.DateTimeFormat is not available");
        var ro = Intl.DateTimeFormat().resolvedOptions();
        var now = new Date();
        function namePart(style) {
          return attempt(function () {
            var parts = new Intl.DateTimeFormat("en-US", { timeZone: ro.timeZone, timeZoneName: style }).formatToParts(now);
            for (var i = 0; i < parts.length; i++) {
              if (parts[i].type === "timeZoneName") return parts[i].value;
            }
            return null;
          }, null);
        }
        return {
          timeZone: ro.timeZone || null,
          locale: ro.locale || null,
          calendar: ro.calendar || null,
          numberingSystem: ro.numberingSystem || null,
          hourCycle: ro.hourCycle || null,
          longOffset: namePart("longOffset"),
          longName: namePart("long"),
          shortName: namePart("short"),
          currentOffsetMinutes: now.getTimezoneOffset(),
          dateToString: String(now),
          dateToLocaleString: attempt(function () {
            return now.toLocaleString();
          }, null),
          navigatorLanguage: attempt(function () {
            return navigator.language;
          }, null),
          navigatorLanguages: attempt(function () {
            return navigator.languages ? Array.prototype.slice.call(navigator.languages) : null;
          }, null)
        };
      }
    },
    {
      id: "time.offsets",
      group: "time",
      run: function () {
        var dates = bootstrap && bootstrap.dates;
        if (!dates || !dates.length) {
          unsupported("the page bootstrap supplied no sample date");
        }
        var tz = attempt(function () {
          return Intl.DateTimeFormat().resolvedOptions().timeZone;
        }, null);
        var rows = [];
        var skipped = 0;
        for (var i = 0; i < dates.length; i++) {
          var sample = dates[i];
          var raw = sample.date === null || sample.date === undefined ? "" : String(sample.date);
          var d;
          if (sample.epochMs !== null && sample.epochMs !== undefined) {
            d = new Date(sample.epochMs);
          } else {
            d = new Date(/^\d{4}-\d{2}-\d{2}$/.test(raw) ? raw + "T12:00:00Z" : raw);
          }
          if (isNaN(d.getTime())) {
            skipped++;
            continue;
          }
          rows.push({
            date: raw || d.toISOString().slice(0, 10),
            instant: sample.iso || d.toISOString(),
            epochMs: d.getTime(),
            offsetMinutes: d.getTimezoneOffset(),
            longOffset: longOffsetFor(tz, d),
            localString: attempt(function () {
              return String(d);
            }, null)
          });
        }
        if (!rows.length) unsupported("no supplied date parsed as a date");
        var distinct = {};
        rows.forEach(function (r) {
          distinct[r.offsetMinutes] = true;
        });
        meta.timeOffsets = {
          datesFrom: (bootstrap && bootstrap.source) || "unknown",
          timeZone: tz,
          sampleCount: rows.length,
          unparsedCount: skipped,
          distinctOffsets: Object.keys(distinct)
            .map(Number)
            .sort(function (a, b) {
              return a - b;
            })
        };
        return rows;
      }
    },
    {
      id: "media.matrix",
      group: "media",
      run: function () {
        var video = attempt(function () {
          return document.createElement("video");
        }, null);
        var audio = attempt(function () {
          return document.createElement("audio");
        }, null);
        var canPlay = !!(video && typeof video.canPlayType === "function");
        var hasMSE = typeof MediaSource !== "undefined" && typeof MediaSource.isTypeSupported === "function";
        var hasRecorder = typeof MediaRecorder !== "undefined" && typeof MediaRecorder.isTypeSupported === "function";
        var hasCaps = !!(navigator.mediaCapabilities && typeof navigator.mediaCapabilities.decodingInfo === "function");
        if (!canPlay && !hasMSE && !hasRecorder && !hasCaps) {
          unsupported("no media capability interface is available");
        }
        var rows = CODECS.map(function (c) {
          var el = c[2] === "video" ? video : audio;
          return {
            label: c[0],
            contentType: c[1],
            kind: c[2],
            canPlayType:
              canPlay && el
                ? attempt(function () {
                    return el.canPlayType(c[1]);
                  }, null)
                : null,
            mediaSource: hasMSE
              ? attempt(function () {
                  return MediaSource.isTypeSupported(c[1]);
                }, null)
              : null,
            decodingInfo: null
          };
        });
        var recorder = RECORDER_TYPES.map(function (t) {
          return {
            contentType: t,
            supported: hasRecorder
              ? attempt(function () {
                  return MediaRecorder.isTypeSupported(t);
                }, null)
              : null
          };
        });
        var facilities = { canPlayType: canPlay, mediaSource: hasMSE, mediaRecorder: hasRecorder, mediaCapabilities: hasCaps };
        meta.media = {
          facilities: facilities,
          codecOrder: rows.map(function (r) {
            return r.label;
          }),
          recorderOrder: recorder.map(function (r) {
            return RECORDER_PREFIX + r.contentType;
          })
        };
        function flatten() {
          var out = { "interfaces available": facilities };
          rows.forEach(function (r) {
            var rec = { contentType: r.contentType, kind: r.kind };
            if (r.canPlayType !== null) rec.canPlayType = r.canPlayType;
            if (r.mediaSource !== null) rec.mediaSource = r.mediaSource;
            var di = r.decodingInfo;
            if (di) {
              if (di.error) rec.decodingInfoError = di.error;
              else {
                rec.decodingInfoSupported = di.supported;
                rec.smooth = di.smooth;
                rec.powerEfficient = di.powerEfficient;
              }
            }
            out[r.label] = rec;
          });
          recorder.forEach(function (r) {
            if (r.supported === null) return;
            out[RECORDER_PREFIX + r.contentType] = { isTypeSupported: r.supported };
          });
          return out;
        }
        if (!hasCaps) return flatten();
        var chain = Promise.resolve();
        rows.forEach(function (r) {
          chain = chain.then(function () {
            var config =
              r.kind === "video"
                ? { type: "file", video: { contentType: r.contentType, width: 1280, height: 720, bitrate: 2000000, framerate: 30 } }
                : { type: "file", audio: { contentType: r.contentType, channels: 2, bitrate: 128000, samplerate: 48000 } };
            return navigator.mediaCapabilities
              .decodingInfo(config)
              .then(function (info) {
                r.decodingInfo = { supported: !!info.supported, smooth: !!info.smooth, powerEfficient: !!info.powerEfficient };
              })
              .catch(function (e) {
                r.decodingInfo = { error: reasonOf(e) };
              });
          });
        });
        return chain.then(function () {
          return flatten();
        });
      }
    },
    {
      id: "auto.residue",
      group: "auto",
      run: function () {
        var winKeys = attempt(function () {
          return Object.getOwnPropertyNames(window);
        }, []);
        var docKeys = attempt(function () {
          return Object.getOwnPropertyNames(document);
        }, []);
        var navDesc = attempt(function () {
          var d = window.Navigator && Object.getOwnPropertyDescriptor(Navigator.prototype, "webdriver");
          if (!d) return null;
          return {
            onPrototype: true,
            kind: d.get ? "accessor" : "data",
            configurable: !!d.configurable,
            getterNative: d.get ? NATIVE_TAIL.test(Function.prototype.toString.call(d.get)) : null
          };
        }, null);

        function stackShape(e) {
          if (!e) return null;
          var s = attempt(function () {
            return e.stack;
          }, null);
          if (typeof s !== "string") return { present: false, type: typeof s };
          var lines = s.split("\n");
          var frames = lines.slice(1);
          return {
            present: true,
            name: attempt(function () {
              return e.name;
            }, null),
            firstLine: lines[0],
            frameCount: frames.length,
            framesWellFormed:
              frames.length > 0 &&
              frames.every(function (l) {
                return /^\s+at\s/.test(l);
              }),
            length: s.length
          };
        }
        var nativeErr = null;
        try {
          null.probeProperty;
        } catch (e) {
          nativeErr = e;
        }
        var userErr = null;
        try {
          (function outer() {
            (function inner() {
              throw new Error("stack shape probe");
            })();
          })();
        } catch (e2) {
          userErr = e2;
        }

        var hits = scanKeys(winKeys).concat(scanKeys(docKeys));
        var hitNames = hits.map(function (x) {
          return x.name;
        });
        var flag = attempt(function () {
          return "webdriver" in navigator ? navigator.webdriver : null;
        }, null);

        var out = {
          driverProperties: hitNames,
          navigatorWebdriverPresent: attempt(function () {
            return "webdriver" in navigator;
          }, null),
          webdriverDescriptor: navDesc,
          windowKeyCount: winKeys.length,
          documentKeyCount: docKeys.length,
          patternHits: hits,
          windowPrefixedKeys: prefixedKeys(winKeys),
          documentPrefixedKeys: prefixedKeys(docKeys),
          errorStackNative: stackShape(nativeErr),
          errorStackUser: stackShape(userErr),
          errorStackDescriptor: attempt(function () {
            var d = Object.getOwnPropertyDescriptor(Error.prototype, "stack");
            return d ? { hasGetter: !!d.get, configurable: !!d.configurable } : null;
          }, null),
          captureStackTraceType: typeof Error.captureStackTrace,
          prepareStackTraceType: typeof Error.prepareStackTrace,
          stackTraceLimit: attempt(function () {
            return Error.stackTraceLimit;
          }, null)
        };
        if (typeof flag === "boolean") out.webdriver = flag;
        return out;
      }
    },
    {
      id: "perm.state",
      group: "perm",
      run: function () {
        if (!navigator.permissions || typeof navigator.permissions.query !== "function") {
          unsupported("navigator.permissions.query is not available");
        }
        var names = ["geolocation", "notifications", "camera", "microphone", "clipboard-read", "midi", "persistent-storage"];
        var rows = [];
        var chain = Promise.resolve();
        names.forEach(function (n) {
          chain = chain.then(function () {
            return navigator.permissions
              .query({ name: n })
              .then(function (p) {
                rows.push({ name: n, state: p.state, error: null });
              })
              .catch(function (e) {
                rows.push({ name: n, state: null, error: reasonOf(e) });
              });
          });
        });
        return chain.then(function () {
          var byName = {};
          rows.forEach(function (r) {
            byName[r.name] = r.state;
          });
          var notif = {
            apiPresent: typeof Notification !== "undefined",
            apiValue:
              typeof Notification !== "undefined"
                ? attempt(function () {
                    return Notification.permission;
                  }, null)
                : null,
            queryState: byName.notifications || null
          };
          var equivalent = { granted: "granted", denied: "denied", prompt: "default" };
          notif.agree =
            notif.apiValue === null ||
            notif.queryState === null ||
            !Object.prototype.hasOwnProperty.call(equivalent, notif.queryState)
              ? null
              : equivalent[notif.queryState] === notif.apiValue;
          var pair = {};
          if (typeof notif.queryState === "string" && notif.queryState) pair.query = notif.queryState;
          if (typeof notif.apiValue === "string" && notif.apiValue) pair.actual = notif.apiValue;

          var geoState = byName.geolocation || null;
          var geo = { queryState: geoState, apiPresent: !!navigator.geolocation, exercised: false, outcome: null, agree: null };
          if (!geo.apiPresent || geoState !== "denied") {
            geo.outcome = "not exercised";
            return { states: rows, notifications: pair, notification: notif, geolocation: geo };
          }
          geo.exercised = true;
          return new Promise(function (resolve) {
            var settled = false;
            function finish() {
              if (settled) return;
              settled = true;
              resolve({ states: rows, notifications: pair, notification: notif, geolocation: geo });
            }
            var t = setTimeout(function () {
              geo.outcome = "no answer within 3000 ms";
              finish();
            }, 3000);
            navigator.geolocation.getCurrentPosition(
              function () {
                clearTimeout(t);
                geo.outcome = "position returned";
                geo.agree = false;
                finish();
              },
              function (err) {
                clearTimeout(t);
                geo.outcome = "error code " + err.code;
                geo.agree = err.code === 1;
                finish();
              },
              { timeout: 2500, maximumAge: 0 }
            );
          });
        });
      }
    }
  ];

  probes.push.apply(probes, (function () {
  function amMathEncode(x) {
    if (typeof x !== "number") return "type:" + typeof x;
    if (Number.isNaN(x)) return "NaN";
    if (x === Infinity) return "Infinity";
    if (x === -Infinity) return "-Infinity";
    if (Object.is(x, -0)) return "-0";
    return String(x);
  }

  var exactProbe = {
    id: "math.exact",
    group: "math",
    run: function () {
      var out = {};
      function put(key, fn) {
        try {
          var v = fn();
          if (typeof v !== "number") return;
          out[key] = amMathEncode(v);
        } catch (e) {
        }
      }

      put("abs.nan", function () { return Math.abs(NaN); });
      put("abs.negZero", function () { return Math.abs(-0); });
      put("abs.negInf", function () { return Math.abs(-Infinity); });
      put("abs.neg", function () { return Math.abs(-7); });
      put("abs.pos", function () { return Math.abs(7); });

      put("sign.nan", function () { return Math.sign(NaN); });
      put("sign.posZero", function () { return Math.sign(0); });
      put("sign.negZero", function () { return Math.sign(-0); });
      put("sign.pos", function () { return Math.sign(5); });
      put("sign.neg", function () { return Math.sign(-5); });

      put("floor.nan", function () { return Math.floor(NaN); });
      put("floor.posInf", function () { return Math.floor(Infinity); });
      put("floor.negInf", function () { return Math.floor(-Infinity); });
      put("floor.negZero", function () { return Math.floor(-0); });
      put("floor.fracNeg", function () { return Math.floor(-0.5); });
      put("floor.fracPos", function () { return Math.floor(2.7); });

      put("ceil.nan", function () { return Math.ceil(NaN); });
      put("ceil.posInf", function () { return Math.ceil(Infinity); });
      put("ceil.negInf", function () { return Math.ceil(-Infinity); });
      put("ceil.posZero", function () { return Math.ceil(0); });
      put("ceil.fracNeg", function () { return Math.ceil(-0.5); });
      put("ceil.fracPos", function () { return Math.ceil(2.3); });

      put("trunc.nan", function () { return Math.trunc(NaN); });
      put("trunc.fracNeg", function () { return Math.trunc(-0.9); });
      put("trunc.fracPos", function () { return Math.trunc(0.9); });
      put("trunc.negInt", function () { return Math.trunc(-3.9); });

      put("round.nan", function () { return Math.round(NaN); });
      put("round.negZero", function () { return Math.round(-0); });
      put("round.halfNeg", function () { return Math.round(-0.5); });
      put("round.halfPos", function () { return Math.round(0.5); });
      put("round.negHalfInt", function () { return Math.round(-2.5); });
      put("round.posInf", function () { return Math.round(Infinity); });
      put("round.negInf", function () { return Math.round(-Infinity); });

      put("min.nan", function () { return Math.min(1, NaN); });
      put("min.zero", function () { return Math.min(-0, 0); });
      put("min.basic", function () { return Math.min(3, 1, 2); });
      put("min.empty", function () { return Math.min(); });

      put("max.nan", function () { return Math.max(1, NaN); });
      put("max.zero", function () { return Math.max(-0, 0); });
      put("max.basic", function () { return Math.max(3, 1, 2); });
      put("max.empty", function () { return Math.max(); });

      put("fround.nan", function () { return Math.fround(NaN); });
      put("fround.negZero", function () { return Math.fround(-0); });
      put("fround.overflow", function () { return Math.fround(Math.pow(2, 150)); });
      put("fround.tieToEven", function () { return Math.fround(16777217); });
      put("fround.exact", function () { return Math.fround(0.5); });

      put("clz32.zero", function () { return Math.clz32(0); });
      put("clz32.one", function () { return Math.clz32(1); });
      put("clz32.negOne", function () { return Math.clz32(-1); });
      put("clz32.nan", function () { return Math.clz32(NaN); });

      put("imul.basic", function () { return Math.imul(3, 4); });
      put("imul.overflow", function () { return Math.imul(0xffffffff, 5); });
      put("imul.bigxbig", function () { return Math.imul(0x7fffffff, 0x7fffffff); });
      put("imul.nan", function () { return Math.imul(NaN, 5); });

      put("sqrt.nan", function () { return Math.sqrt(NaN); });
      put("sqrt.negative", function () { return Math.sqrt(-1); });
      put("sqrt.negZero", function () { return Math.sqrt(-0); });
      put("sqrt.posInf", function () { return Math.sqrt(Infinity); });
      put("sqrt.perfect", function () { return Math.sqrt(4); });
      put("sqrt.exactFraction", function () { return Math.sqrt(6.25); });

      return out;
    }
  };

  var repeatProbe = {
    id: "math.repeat",
    group: "math",
    run: function () {
      var out = {};
      function pair(key, fn) {
        var a, b, gotA = true, gotB = true;
        try { a = fn(); } catch (e) { gotA = false; }
        try { b = fn(); } catch (e) { gotB = false; }
        if (!gotA || !gotB) return;
        if (typeof a !== "number" || typeof b !== "number") return;
        out[key] = { a: amMathEncode(a), b: amMathEncode(b) };
      }

      pair("abs", function () { return Math.abs(-7); });
      pair("sign", function () { return Math.sign(-7); });
      pair("floor", function () { return Math.floor(2.7); });
      pair("ceil", function () { return Math.ceil(2.3); });
      pair("trunc", function () { return Math.trunc(-2.7); });
      pair("round", function () { return Math.round(2.5); });
      pair("min", function () { return Math.min(3, 1, 2); });
      pair("max", function () { return Math.max(3, 1, 2); });
      pair("fround", function () { return Math.fround(1.1); });
      pair("clz32", function () { return Math.clz32(1000); });
      pair("imul", function () { return Math.imul(123, 456); });
      pair("sqrt", function () { return Math.sqrt(2); });

      pair("sin", function () { return Math.sin(0.7); });
      pair("cos", function () { return Math.cos(0.7); });
      pair("tan", function () { return Math.tan(0.7); });
      pair("exp", function () { return Math.exp(0.7); });
      pair("log", function () { return Math.log(0.7); });
      pair("log2", function () { return Math.log2(0.7); });
      pair("log10", function () { return Math.log10(0.7); });
      pair("pow", function () { return Math.pow(1.3, 2.7); });
      pair("hypot", function () { return Math.hypot(3, 4, 5); });
      pair("cbrt", function () { return Math.cbrt(30); });
      pair("atan2", function () { return Math.atan2(0.3, 0.7); });
      pair("expm1", function () { return Math.expm1(0.7); });
      pair("log1p", function () { return Math.log1p(0.7); });
      pair("sinh", function () { return Math.sinh(0.7); });
      pair("cosh", function () { return Math.cosh(0.7); });
      pair("tanh", function () { return Math.tanh(0.7); });
      pair("asinh", function () { return Math.asinh(0.7); });
      pair("acosh", function () { return Math.acosh(1.7); });
      pair("atanh", function () { return Math.atanh(0.7); });

      return out;
    }
  };

    return [exactProbe, repeatProbe];
  })());

  probes.push.apply(probes, (function () {
  "use strict";

  function amThrowsCtorGlobal(name) {
    switch (name) {
      case "TypeError":
        return typeof TypeError !== "undefined" ? TypeError : null;
      case "RangeError":
        return typeof RangeError !== "undefined" ? RangeError : null;
      case "SyntaxError":
        return typeof SyntaxError !== "undefined" ? SyntaxError : null;
      case "URIError":
        return typeof URIError !== "undefined" ? URIError : null;
      case "DOMException":
        return typeof DOMException !== "undefined" ? DOMException : null;
      default:
        return null;
    }
  }

  function amThrowsObserveInto(rec, e, ctorName) {
    try {
      rec.name = e && typeof e.name === "string" ? e.name : null;
    } catch (ignore1) {
      rec.name = null;
    }
    try {
      var ctor = e && e.constructor;
      rec.ctor = ctor && typeof ctor.name === "string" ? ctor.name : null;
    } catch (ignore2) {
      rec.ctor = null;
    }
    var globalCtor = null;
    try {
      globalCtor = amThrowsCtorGlobal(ctorName);
      rec.ctorGlobalAvailable = !!globalCtor;
    } catch (ignore3) {
      rec.ctorGlobalAvailable = false;
    }
    try {
      rec.instanceOf = !!(globalCtor && e instanceof globalCtor);
    } catch (ignore4) {
      rec.instanceOf = false;
    }
    try {
      rec.toStringTag = Object.prototype.toString.call(e);
    } catch (ignore5) {
      rec.toStringTag = null;
    }
    try {
      rec.hasStack = !!e && typeof e.stack === "string";
    } catch (ignore6) {
      rec.hasStack = false;
    }
  }

  function amThrowsRunMandated(c) {
    var avail = true;
    try {
      avail = !!c.available();
    } catch (ignoreAvail) {
      avail = false;
    }
    if (!avail) {
      return { available: false };
    }

    var rec = { available: true, threw: false };
    try {
      c.trigger();
    } catch (e1) {
      rec.threw = true;
      amThrowsObserveInto(rec, e1, c.ctor);
    }

    var attempt2 = { threw: false };
    try {
      c.trigger();
    } catch (e2) {
      attempt2.threw = true;
      amThrowsObserveInto(attempt2, e2, c.ctor);
    }
    rec.attempt2 = attempt2;
    return rec;
  }

  var AM_THROWS_MANDATED_CASES = [
    {
      id: "atob.invalidChars",
      ctor: "DOMException",
      available: function () {
        return typeof atob === "function";
      },
      trigger: function () {
        atob("a");
      }
    },
    {
      id: "createElement.invalidName",
      ctor: "DOMException",
      available: function () {
        return typeof document !== "undefined" && typeof document.createElement === "function";
      },
      trigger: function () {
        document.createElement("");
      }
    },
    {
      id: "json.malformed",
      ctor: "SyntaxError",
      available: function () {
        return typeof JSON !== "undefined" && typeof JSON.parse === "function";
      },
      trigger: function () {
        JSON.parse("{");
      }
    },
    {
      id: "property.accessOnNull",
      ctor: "TypeError",
      available: function () {
        return true;
      },
      trigger: function () {
        var n = null;
        return n.x;
      }
    },
    {
      id: "call.nonCallable",
      ctor: "TypeError",
      available: function () {
        return true;
      },
      trigger: function () {
        var notAFunction = {};
        notAFunction();
      }
    },
    {
      id: "array.negativeLength",
      ctor: "RangeError",
      available: function () {
        return true;
      },
      trigger: function () {
        return new Array(-1);
      }
    },
    {
      id: "decodeURIComponent.malformed",
      ctor: "URIError",
      available: function () {
        return true;
      },
      trigger: function () {
        decodeURIComponent("%");
      }
    },
    {
      id: "structuredClone.function",
      ctor: "DOMException",
      available: function () {
        return typeof structuredClone === "function";
      },
      trigger: function () {
        structuredClone(function () {});
      }
    }
  ];

  var AM_THROWS_NAME_BATTERY = [
    {
      id: "typedArray.negativeLength",
      trigger: function () {
        return new Uint8Array(-1);
      }
    },
    {
      id: "map.nonIterablePrimitive",
      trigger: function () {
        return new Map(1);
      }
    },
    {
      id: "reflect.applyNonFunction",
      trigger: function () {
        return Reflect.apply({}, null, []);
      }
    },
    {
      id: "arrayBuffer.negativeLength",
      trigger: function () {
        return new ArrayBuffer(-1);
      }
    },
    {
      id: "weakmap.primitiveKey",
      trigger: function () {
        var m = new WeakMap();
        m.set(1, 1);
      }
    }
  ];

  var AM_THROWS_PROBES = [
    {
      id: "throw.mandated",
      group: "throw",
      run: function () {
        var out = {};
        for (var i = 0; i < AM_THROWS_MANDATED_CASES.length; i++) {
          var c = AM_THROWS_MANDATED_CASES[i];
          out[c.id] = amThrowsRunMandated(c);
        }
        return out;
      }
    },
    {
      id: "throw.names",
      group: "throw",
      run: function () {
        var out = {};
        for (var i = 0; i < AM_THROWS_NAME_BATTERY.length; i++) {
          var b = AM_THROWS_NAME_BATTERY[i];
          var rec = { threw: false };
          try {
            b.trigger();
          } catch (e) {
            rec.threw = true;
            try {
              rec.name = e && typeof e.name === "string" ? e.name : null;
            } catch (ignore) {
              rec.name = null;
            }
          }
          out[b.id] = rec;
        }
        return out;
      }
    }
  ];

    return AM_THROWS_PROBES;
  })());

  probes.push.apply(probes, (function () {
function mediaPathSalt() {
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

var mediaPathHideStyle =
  "position:absolute !important;" +
  "left:-9999px !important; top:0 !important;" +
  "width:0 !important; height:0 !important;" +
  "min-width:0 !important; min-height:0 !important;" +
  "max-width:0 !important; max-height:0 !important;" +
  "margin:0 !important; padding:0 !important; border:0 !important;" +
  "overflow:hidden !important; visibility:hidden !important;" +
  "pointer-events:none !important;";

var mediaPathStylesheetProbe = {
  id: "media.stylesheet",
  group: "media",
  run: function () {
    if (typeof window.matchMedia !== "function") unsupported("matchMedia is not available");
    var host = document.head || document.documentElement;
    if (!host) unsupported("no document to attach a stylesheet to");
    if (typeof window.getComputedStyle !== "function") unsupported("getComputedStyle is not available");

    var style = document.createElement("style");
    style.setAttribute("data-am-probe", "media-stylesheet");
    var inserted = [];
    var created = [];

    try {
      var salt = mediaPathSalt();
      var idx = 0;
      function nextId() {
        idx++;
        return "am-mp-" + idx + "-" + salt;
      }

      var cssParts = [];
      function plant(id, query, sentinel) {
        cssParts.push("#" + id + "{--am-mp:0}");
        if (query === null) {
          cssParts.push("#" + id + "{--am-mp:" + sentinel + " !important}");
        } else {
          cssParts.push("@media " + query + "{#" + id + "{--am-mp:" + sentinel + " !important}}");
        }
      }

      var controlId = nextId();
      plant(controlId, null, 9);

      var widthTautLowId = nextId();
      plant(widthTautLowId, "(min-width: 0px)", 1);
      var widthTautHighId = nextId();
      plant(widthTautHighId, "(max-width: 100000px)", 1);
      var heightTautLowId = nextId();
      plant(heightTautLowId, "(min-height: 0px)", 1);
      var heightTautHighId = nextId();
      plant(heightTautHighId, "(max-height: 100000px)", 1);

      var widthPx = [1, 320, 768, 1200, 2560, 4000];
      var heightPx = [1, 480, 900, 1600];
      var numericPlan = [];
      widthPx.forEach(function (px) {
        ["min", "max"].forEach(function (op) {
          var id = nextId();
          var q = "(" + op + "-width: " + px + "px)";
          plant(id, q, 1);
          numericPlan.push({ feature: "width", op: op, px: px, id: id, query: q });
        });
      });
      heightPx.forEach(function (px) {
        ["min", "max"].forEach(function (op) {
          var id = nextId();
          var q = "(" + op + "-height: " + px + "px)";
          plant(id, q, 1);
          numericPlan.push({ feature: "height", op: op, px: px, id: id, query: q });
        });
      });

      var discreteFeatures = [
        { feature: "orientation", values: ["portrait", "landscape"] },
        { feature: "hover", values: ["hover", "none"] },
        { feature: "pointer", values: ["fine", "coarse", "none"] },
        { feature: "prefers-color-scheme", values: ["dark", "light"] },
        { feature: "prefers-reduced-motion", values: ["reduce", "no-preference"] }
      ];
      var discretePlan = [];
      discreteFeatures.forEach(function (f) {
        f.values.forEach(function (val) {
          var id = nextId();
          var q = "(" + f.feature + ": " + val + ")";
          plant(id, q, 1);
          discretePlan.push({ feature: f.feature, value: val, id: id, query: q });
        });
      });

      style.textContent = cssParts.join("\n");
      host.appendChild(style);
      inserted.push(style);

      var allIds = [controlId, widthTautLowId, widthTautHighId, heightTautLowId, heightTautHighId]
        .concat(
          numericPlan.map(function (p) {
            return p.id;
          })
        )
        .concat(
          discretePlan.map(function (p) {
            return p.id;
          })
        );

      var elById = {};
      for (var i = 0; i < allIds.length; i++) {
        var id = allIds[i];
        var el = document.createElement("am-mp-probe");
        el.id = id;
        el.setAttribute("aria-hidden", "true");
        el.style.cssText = mediaPathHideStyle;
        host.appendChild(el);
        created.push(el);
        elById[id] = el;
      }

      function marker(id) {
        return attempt(function () {
          return window
            .getComputedStyle(elById[id])
            .getPropertyValue("--am-mp")
            .replace(/^\s+|\s+$/g, "");
        }, "");
      }

      var controlOk = marker(controlId) === "9";
      var widthValid = controlOk && marker(widthTautLowId) === "1" && marker(widthTautHighId) === "1";
      var heightValid = controlOk && marker(heightTautLowId) === "1" && marker(heightTautHighId) === "1";

      function cssBool(id) {
        var v = marker(id);
        if (v === "1") return true;
        if (v === "0") return false;
        return null;
      }

      var numeric = numericPlan.map(function (p) {
        return {
          feature: p.feature,
          op: p.op,
          px: p.px,
          jsMatches: attempt(
            function () {
              return mq(p.query);
            },
            null
          ),
          cssMatches: cssBool(p.id)
        };
      });

      var discrete = discretePlan.map(function (p) {
        return {
          feature: p.feature,
          value: p.value,
          jsMatches: attempt(
            function () {
              return mq(p.query);
            },
            null
          ),
          cssMatches: cssBool(p.id)
        };
      });

      return {
        controlOk: controlOk,
        widthValid: widthValid,
        heightValid: heightValid,
        numeric: numeric,
        discrete: discrete
      };
    } finally {
      for (var c = 0; c < created.length; c++) {
        if (created[c].parentNode) created[c].parentNode.removeChild(created[c]);
      }
      for (var s = 0; s < inserted.length; s++) {
        if (inserted[s].parentNode) inserted[s].parentNode.removeChild(inserted[s]);
      }
    }
  }
};

var mediaPathComplementProbe = {
  id: "media.complement",
  group: "media",
  run: function () {
    if (typeof window.matchMedia !== "function") unsupported("matchMedia is not available");

    var innerWidth = attempt(
      function () {
        var w = window.innerWidth;
        return typeof w === "number" && isFinite(w) ? w : null;
      },
      null
    );
    var innerHeight = attempt(
      function () {
        var h = window.innerHeight;
        return typeof h === "number" && isFinite(h) ? h : null;
      },
      null
    );

    var complementQueries = [
      "(min-width: 1px)",
      "(min-width: 100000px)",
      "(min-height: 1px)",
      "(min-height: 100000px)",
      "(orientation: portrait)",
      "(prefers-reduced-motion: reduce)"
    ];
    var complements = complementQueries.map(function (q) {
      return {
        query: q,
        matches: attempt(
          function () {
            return mq(q);
          },
          null
        ),
        negationMatches: attempt(
          function () {
            return mq("not " + q);
          },
          null
        )
      };
    });

    function bracketFor(feature, value) {
      if (typeof value !== "number" || !isFinite(value)) {
        return {
          feature: feature,
          value: null,
          insideBelowPx: null,
          insideAbovePx: null,
          minInside: null,
          maxInside: null,
          minOutside: null,
          maxOutside: null
        };
      }
      var inLo = Math.floor(value) - 1;
      var inHi = Math.ceil(value) + 1;
      var outLo = Math.floor(value) + 2;
      var outHi = Math.ceil(value) - 2;
      function ask(q) {
        return attempt(function () {
          return mq(q);
        }, null);
      }
      return {
        feature: feature,
        value: value,
        insideBelowPx: inLo,
        insideAbovePx: inHi,
        minInside: inLo >= 0 ? ask("(min-" + feature + ": " + inLo + "px)") : null,
        maxInside: ask("(max-" + feature + ": " + inHi + "px)"),
        minOutside: outLo >= 0 ? ask("(min-" + feature + ": " + outLo + "px)") : null,
        maxOutside: outHi >= 0 ? ask("(max-" + feature + ": " + outHi + "px)") : null
      };
    }

    return {
      innerWidth: innerWidth,
      innerHeight: innerHeight,
      complements: complements,
      brackets: [bracketFor("width", innerWidth), bracketFor("height", innerHeight)]
    };
  }
};

var mediaPathProbes = [mediaPathStylesheetProbe, mediaPathComplementProbe];

    return mediaPathProbes;
  })());

  probes.push.apply(probes, (function () {
function audioOfflineCtor() {
  if (typeof OfflineAudioContext === "function") return OfflineAudioContext;
  if (typeof window !== "undefined" && typeof window.webkitOfflineAudioContext === "function") {
    return window.webkitOfflineAudioContext;
  }
  return null;
}

function audioCloseContext(ctx) {
  if (!ctx || typeof ctx.close !== "function") return;
  try {
    var p = ctx.close();
    if (p && typeof p.then === "function") p.then(null, function () {});
  } catch (e) {
  }
}

function audioBuildDeterministicGraph(ctx) {
  var osc = ctx.createOscillator();
  osc.type = "sine";
  osc.frequency.value = 440;
  osc.connect(ctx.destination);
  osc.start(0);
}

function audioChannelsServed(buffer) {
  if (!buffer || typeof buffer.getChannelData !== "function") return null;
  var n = 0;
  while (true) {
    try {
      buffer.getChannelData(n);
    } catch (e) {
      break;
    }
    n++;
    if (n > 64) break; 
  }
  return n;
}

function audioCompareChannelArrays(a, b, acc) {
  var n = Math.min(a.length, b.length);
  for (var i = 0; i < n; i++) {
    acc.sampleCount++;
    var d = Math.abs(a[i] - b[i]);
    if (d !== 0) {
      acc.differingSampleCount++;
      if (d > acc.maxAbsoluteDifference) acc.maxAbsoluteDifference = d;
    }
  }
}

function audioBufferDescription(buffer) {
  return {
    sampleRateHz: buffer.sampleRate,
    numberOfChannels: buffer.numberOfChannels,
    lengthFrames: buffer.length,
    durationSeconds: buffer.duration
  };
}

var AUDIO_VIEWS_SAMPLE_RATE_HZ = 8000;
var AUDIO_VIEWS_NUM_CHANNELS = 2;
var AUDIO_VIEWS_DURATION_SECONDS = 0.02; 

var AUDIO_REPEAT_SAMPLE_RATE_HZ = 8000;
var AUDIO_REPEAT_NUM_CHANNELS = 1;
var AUDIO_REPEAT_DURATION_SECONDS = 0.02;

var AUDIO_PROBES = [
  {
    id: "audio.views",
    group: "audio",
    run: function () {
      var Ctor = audioOfflineCtor();
      if (!Ctor) unsupported("OfflineAudioContext is not available");

      var sampleRateHz = AUDIO_VIEWS_SAMPLE_RATE_HZ;
      var numberOfChannels = AUDIO_VIEWS_NUM_CHANNELS;
      var lengthFrames = Math.max(1, Math.round(sampleRateHz * AUDIO_VIEWS_DURATION_SECONDS));

      var ctx;
      try {
        ctx = new Ctor(numberOfChannels, lengthFrames, sampleRateHz);
        audioBuildDeterministicGraph(ctx);
      } catch (e) {
        unsupported("could not construct or graph the offline render: " + reasonOf(e));
      }

      var rendering;
      try {
        rendering = ctx.startRendering();
      } catch (e) {
        audioCloseContext(ctx);
        unsupported("startRendering did not run: " + reasonOf(e));
      }

      return withTimeout(rendering, 4000, "audio.views render").then(
        function (buffer) {
          var served = audioChannelsServed(buffer);
          var out = {
            requested: { sampleRateHz: sampleRateHz, numberOfChannels: numberOfChannels, lengthFrames: lengthFrames },
            rendered: audioBufferDescription(buffer)
          };
          if (served !== null) out.channelsServed = served;

          var hasCopy = typeof buffer.copyFromChannel === "function";
          out.copyFromChannelAvailable = hasCopy;
          if (hasCopy && served !== null) {
            var acc = { sampleCount: 0, differingSampleCount: 0, maxAbsoluteDifference: 0 };
            var channels = Math.min(out.rendered.numberOfChannels, served);
            var ranAny = false;
            for (var c = 0; c < channels; c++) {
              try {
                var direct = buffer.getChannelData(c);
                var dest = new Float32Array(direct.length);
                buffer.copyFromChannel(dest, c, 0);
                audioCompareChannelArrays(direct, dest, acc);
                ranAny = true;
              } catch (e) {
              }
            }
            if (ranAny) {
              out.views = {
                compared: true,
                agree: acc.differingSampleCount === 0,
                sampleCount: acc.sampleCount,
                differingSampleCount: acc.differingSampleCount,
                maxAbsoluteDifference: acc.maxAbsoluteDifference
              };
            } else {
              out.views = { compared: false };
            }
          }

          audioCloseContext(ctx);
          return out;
        },
        function (e) {
          audioCloseContext(ctx);
          throw e;
        }
      );
    }
  },
  {
    id: "audio.repeat",
    group: "audio",
    run: function () {
      var Ctor = audioOfflineCtor();
      if (!Ctor) unsupported("OfflineAudioContext is not available");

      var sampleRateHz = AUDIO_REPEAT_SAMPLE_RATE_HZ;
      var numberOfChannels = AUDIO_REPEAT_NUM_CHANNELS;
      var lengthFrames = Math.max(1, Math.round(sampleRateHz * AUDIO_REPEAT_DURATION_SECONDS));

      function renderOnce() {
        var ctx;
        try {
          ctx = new Ctor(numberOfChannels, lengthFrames, sampleRateHz);
          audioBuildDeterministicGraph(ctx);
        } catch (e) {
          unsupported("could not construct or graph the offline render: " + reasonOf(e));
        }
        var rendering;
        try {
          rendering = ctx.startRendering();
        } catch (e) {
          audioCloseContext(ctx);
          unsupported("startRendering did not run: " + reasonOf(e));
        }
        return withTimeout(rendering, 4000, "audio.repeat render").then(
          function (buffer) {
            audioCloseContext(ctx);
            return buffer;
          },
          function (e) {
            audioCloseContext(ctx);
            throw e;
          }
        );
      }

      return renderOnce().then(function (firstBuffer) {
        var out = {
          requested: { sampleRateHz: sampleRateHz, numberOfChannels: numberOfChannels, lengthFrames: lengthFrames },
          rendered: audioBufferDescription(firstBuffer),
          secondRenderCompleted: false
        };
        var firstChannel = attempt(
          function () {
            return firstBuffer.getChannelData(0);
          },
          null
        );
        if (!firstChannel) return out;

        return renderOnce().then(
          function (secondBuffer) {
            out.secondRenderCompleted = true;
            var secondChannel = attempt(
              function () {
                return secondBuffer.getChannelData(0);
              },
              null
            );
            if (!secondChannel) return out;
            var acc = { sampleCount: 0, differingSampleCount: 0, maxAbsoluteDifference: 0 };
            audioCompareChannelArrays(firstChannel, secondChannel, acc);
            out.repeat = {
              compared: true,
              agree: acc.differingSampleCount === 0,
              sampleCount: acc.sampleCount,
              differingSampleCount: acc.differingSampleCount,
              maxAbsoluteDifference: acc.maxAbsoluteDifference
            };
            return out;
          },
          function () {
            return out;
          }
        );
      });
    }
  }
];

    return AUDIO_PROBES;
  })());

  probes.push.apply(probes, [
{
  id: "rect.identities",
  group: "rect",
  run: function () {
    if (!document.body || typeof document.createElement !== "function") {
      unsupported("document.body is not available");
    }
    var shiftPx = 41;
    var style =
      "position:fixed;" +
      "top:0px;left:0px;" +
      "width:181px;height:97px;" +
      "margin:0px;border:0px none;padding:0px;" +
      "box-sizing:border-box;" +
      "visibility:hidden;" +
      "pointer-events:none;" +
      "transform:none;";
    var a = document.createElement("div");
    var b = document.createElement("div");
    a.setAttribute("style", style);
    b.setAttribute("style", style);
    function readRect(el) {
      var r = el.getBoundingClientRect();
      return {
        x: r.x,
        y: r.y,
        left: r.left,
        top: r.top,
        right: r.right,
        bottom: r.bottom,
        width: r.width,
        height: r.height
      };
    }
    try {
      if (typeof a.getBoundingClientRect !== "function") {
        unsupported("getBoundingClientRect is not available");
      }
      document.body.appendChild(a);
      document.body.appendChild(b);
      var base = readRect(a);
      var twin = readRect(b);
      a.style.transform = "translateX(" + shiftPx + "px)";
      var shifted = readRect(a);
      a.style.transform = "none";
      var restored = readRect(a);
      return {
        shiftPx: shiftPx,
        base: base,
        twin: twin,
        shifted: shifted,
        restored: restored
      };
    } finally {
      if (a.parentNode) a.parentNode.removeChild(a);
      if (b.parentNode) b.parentNode.removeChild(b);
    }
  }
},
{
  id: "text.metrics",
  group: "text",
  run: function () {
    var c = document.createElement("canvas");
    var ctx = c.getContext && c.getContext("2d");
    if (!ctx || typeof ctx.measureText !== "function") {
      unsupported("canvas 2d measureText is not available");
    }
    ctx.font = "32px monospace";
    ctx.textAlign = "left";
    ctx.textBaseline = "alphabetic";
    ctx.direction = "ltr";

    var probeString = "The Quick Fox";

    function numOrNull(v) {
      return typeof v === "number" && isFinite(v) ? v : null;
    }
    function boxOf(m) {
      return {
        actualBoundingBoxLeft: numOrNull(m.actualBoundingBoxLeft),
        actualBoundingBoxRight: numOrNull(m.actualBoundingBoxRight),
        actualBoundingBoxAscent: numOrNull(m.actualBoundingBoxAscent),
        actualBoundingBoxDescent: numOrNull(m.actualBoundingBoxDescent),
        fontBoundingBoxAscent: numOrNull(m.fontBoundingBoxAscent),
        fontBoundingBoxDescent: numOrNull(m.fontBoundingBoxDescent),
        emHeightAscent: numOrNull(m.emHeightAscent),
        emHeightDescent: numOrNull(m.emHeightDescent),
        hangingBaseline: numOrNull(m.hangingBaseline),
        alphabeticBaseline: numOrNull(m.alphabeticBaseline),
        ideographicBaseline: numOrNull(m.ideographicBaseline)
      };
    }

    var emptyMetrics = ctx.measureText("");
    var fullMetrics = ctx.measureText(probeString);
    var repeatMetrics = ctx.measureText(probeString);

    var prefixWidths = [];
    for (var i = 0; i <= probeString.length; i++) {
      prefixWidths.push(ctx.measureText(probeString.slice(0, i)).width);
    }

    return {
      empty: { width: emptyMetrics.width, box: boxOf(emptyMetrics) },
      full: { width: fullMetrics.width, box: boxOf(fullMetrics) },
      repeat: { width: repeatMetrics.width },
      prefixWidths: prefixWidths
    };
  }
}
  ]);

  function longOffsetFor(tz, d) {
    if (!tz) return null;
    return attempt(function () {
      var parts = new Intl.DateTimeFormat("en-US", { timeZone: tz, timeZoneName: "longOffset" }).formatToParts(d);
      for (var k = 0; k < parts.length; k++) {
        if (parts[k].type === "timeZoneName") return parts[k].value;
      }
      return null;
    }, null);
  }

  function runProbe(p) {
    return withTimeout(
      new Promise(function (resolve) {
        resolve(p.run());
      }),
      PROBE_TIMEOUT_MS,
      p.id
    ).then(
      function (v) {
        return { status: "ok", value: v };
      },
      function (e) {
        if (e && e.amUnsupported) return { status: "unsupported", value: { reason: reasonOf(e) } };
        return { status: "error", value: { reason: reasonOf(e) } };
      }
    );
  }

  function run(onProgress) {
    measureCache = null;
    nativeCache = null;
    scopeCache = null;
    meta = {};
    return loadBootstrap().then(function (b) {
      bootstrap = b;
      var out = {};
      var chain = Promise.resolve();
      probes.forEach(function (p, i) {
        chain = chain.then(function () {
          if (onProgress) onProgress(i, probes.length, p.id);
          return runProbe(p).then(function (r) {
            out[p.id] = r;
          });
        });
      });
      return chain.then(function () {
        if (onProgress) onProgress(probes.length, probes.length, null);
        return {
          probes: out,
          meta: meta,
          nonce: b.nonce || "",
          bootstrapSource: b.source,
          ids: probes.map(function (p) {
            return p.id;
          })
        };
      });
    });
  }

  function groupOf(id) {
    for (var i = 0; i < probes.length; i++) {
      if (probes[i].id === id) return probes[i].group;
    }
    return String(id).split(".")[0];
  }

  return {
    run: run,
    ids: probes.map(function (p) {
      return p.id;
    }),
    groupOf: groupOf
  };
})();
