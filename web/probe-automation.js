(function () {
  "use strict";

  var core = AM_CORE;
  var attempt = core.attempt;
  var NATIVE_TAIL = core.NATIVE_TAIL;

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

  core.register([
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
    }
  ]);
})();
