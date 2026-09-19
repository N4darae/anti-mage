(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var attempt = core.attempt;
  var mq = core.mq;

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

  core.register(mediaPathProbes);
})();
