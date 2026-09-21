(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;
  var withTimeout = core.withTimeout;

  core.register([
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
    }
  ]);
})();
