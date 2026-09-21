(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var attempt = core.attempt;
  var mq = core.mq;

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

  core.register([
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
    }
  ]);
})();
