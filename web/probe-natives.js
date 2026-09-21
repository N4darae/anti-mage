(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;
  var NATIVE_TAIL = core.NATIVE_TAIL;

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

  core.onReset(function () {
    nativeCache = null;
  });

  core.register([
    {
      id: "native.tostring",
      group: "native",
      run: function () {
        var n = collectNatives();
        core.meta.native = { globals: n.globals, order: n.targets.map(function (t) { return t.key; }) };
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
    }
  ]);
})();
