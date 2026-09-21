
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
