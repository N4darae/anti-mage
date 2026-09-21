(function () {
  "use strict";

  var core = AM_CORE;
  var Unsupported = core.Unsupported;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;

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

  core.onReset(function () {
    scopeCache = null;
  });

  core.register([
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
    }
  ]);
})();
