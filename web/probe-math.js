(function () {
  "use strict";

  var core = AM_CORE;

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

  core.register([exactProbe, repeatProbe]);
})();
