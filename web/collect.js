
var AM_CORE = (function () {
  "use strict";

  var core = {
    bootstrap: { dates: null, fonts: null, nonce: "", source: "none" },
    meta: {},
    probes: [],
    ids: [],
    resets: []
  };

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

  var NATIVE_TAIL = /\{\s*\[native code\]\s*\}\s*$/;

  function mq(q) {
    return window.matchMedia(q).matches;
  }

  core.Unsupported = Unsupported;
  core.unsupported = unsupported;
  core.reasonOf = reasonOf;
  core.attempt = attempt;
  core.withTimeout = withTimeout;
  core.NATIVE_TAIL = NATIVE_TAIL;
  core.mq = mq;

  core.register = function (list) {
    for (var i = 0; i < list.length; i++) {
      core.probes.push(list[i]);
      core.ids.push(list[i].id);
    }
  };

  core.onReset = function (fn) {
    core.resets.push(fn);
  };

  return core;
})();

var AM = (function () {
  "use strict";

  var core = AM_CORE;
  var reasonOf = core.reasonOf;
  var withTimeout = core.withTimeout;
  var probes = core.probes;

  var PROBE_TIMEOUT_MS = 9000;

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

  function startProbe(p) {
    return new Promise(function (resolve) {
      resolve(p.prime ? p.prime() : p.run());
    });
  }

  function runProbe(p, started) {
    return withTimeout(
      started || startProbe(p),
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
    core.resets.forEach(function (fn) {
      fn();
    });
    core.meta = {};
    return loadBootstrap().then(function (b) {
      core.bootstrap = b;
      var out = {};
      var started = {};
      probes.forEach(function (p) {
        if (!p.prime) return;
        var pending = startProbe(p);
        pending.catch(function () {});
        started[p.id] = pending;
      });
      var chain = Promise.resolve();
      probes.forEach(function (p, i) {
        chain = chain.then(function () {
          if (onProgress) onProgress(i, probes.length, p.id);
          return runProbe(p, started[p.id]).then(function (r) {
            out[p.id] = r;
          });
        });
      });
      return chain.then(function () {
        if (onProgress) onProgress(probes.length, probes.length, null);
        return {
          probes: out,
          meta: core.meta,
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
    ids: core.ids,
    groupOf: groupOf
  };
})();

(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;

  function fontInput() {
    var d = AM_FONT_DATA;
    var b = core.bootstrap && core.bootstrap.fonts;
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

  core.onReset(function () {
    measureCache = null;
  });

  core.register([
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
        core.meta.font = {
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
        core.meta.fontCoverage = { inputSource: (input.sources && input.sources.coverage) || input.source, order: order };
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
        core.meta.fontControls = {
          inputSource: (input.sources && input.sources.controls) || input.source,
          order: d.controls.slice(),
          count: d.controls.length,
          widthResolved: widthResolved,
          checkAvailable: checkAvailable,
          checkTrueCount: checkTrue.length
        };
        return out;
      }
    }
  ]);
})();

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
