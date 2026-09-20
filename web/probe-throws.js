(function () {
  "use strict";

  var core = AM_CORE;

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

  core.register(AM_THROWS_PROBES);
})();
