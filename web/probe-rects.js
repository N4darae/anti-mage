(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;

  core.register([
{
  id: "rect.identities",
  group: "rect",
  run: function () {
    if (!document.body || typeof document.createElement !== "function") {
      unsupported("document.body is not available");
    }
    var shiftPx = 41;
    var style =
      "position:fixed;" +
      "top:0px;left:0px;" +
      "width:181px;height:97px;" +
      "margin:0px;border:0px none;padding:0px;" +
      "box-sizing:border-box;" +
      "visibility:hidden;" +
      "pointer-events:none;" +
      "transform:none;";
    var a = document.createElement("div");
    var b = document.createElement("div");
    a.setAttribute("style", style);
    b.setAttribute("style", style);
    function readRect(el) {
      var r = el.getBoundingClientRect();
      return {
        x: r.x,
        y: r.y,
        left: r.left,
        top: r.top,
        right: r.right,
        bottom: r.bottom,
        width: r.width,
        height: r.height
      };
    }
    try {
      if (typeof a.getBoundingClientRect !== "function") {
        unsupported("getBoundingClientRect is not available");
      }
      document.body.appendChild(a);
      document.body.appendChild(b);
      var base = readRect(a);
      var twin = readRect(b);
      a.style.transform = "translateX(" + shiftPx + "px)";
      var shifted = readRect(a);
      a.style.transform = "none";
      var restored = readRect(a);
      return {
        shiftPx: shiftPx,
        base: base,
        twin: twin,
        shifted: shifted,
        restored: restored
      };
    } finally {
      if (a.parentNode) a.parentNode.removeChild(a);
      if (b.parentNode) b.parentNode.removeChild(b);
    }
  }
},
{
  id: "text.metrics",
  group: "text",
  run: function () {
    var c = document.createElement("canvas");
    var ctx = c.getContext && c.getContext("2d");
    if (!ctx || typeof ctx.measureText !== "function") {
      unsupported("canvas 2d measureText is not available");
    }
    ctx.font = "32px monospace";
    ctx.textAlign = "left";
    ctx.textBaseline = "alphabetic";
    ctx.direction = "ltr";

    var probeString = "The Quick Fox";

    function numOrNull(v) {
      return typeof v === "number" && isFinite(v) ? v : null;
    }
    function boxOf(m) {
      return {
        actualBoundingBoxLeft: numOrNull(m.actualBoundingBoxLeft),
        actualBoundingBoxRight: numOrNull(m.actualBoundingBoxRight),
        actualBoundingBoxAscent: numOrNull(m.actualBoundingBoxAscent),
        actualBoundingBoxDescent: numOrNull(m.actualBoundingBoxDescent),
        fontBoundingBoxAscent: numOrNull(m.fontBoundingBoxAscent),
        fontBoundingBoxDescent: numOrNull(m.fontBoundingBoxDescent),
        emHeightAscent: numOrNull(m.emHeightAscent),
        emHeightDescent: numOrNull(m.emHeightDescent),
        hangingBaseline: numOrNull(m.hangingBaseline),
        alphabeticBaseline: numOrNull(m.alphabeticBaseline),
        ideographicBaseline: numOrNull(m.ideographicBaseline)
      };
    }

    var emptyMetrics = ctx.measureText("");
    var fullMetrics = ctx.measureText(probeString);
    var repeatMetrics = ctx.measureText(probeString);

    var prefixWidths = [];
    for (var i = 0; i <= probeString.length; i++) {
      prefixWidths.push(ctx.measureText(probeString.slice(0, i)).width);
    }

    return {
      empty: { width: emptyMetrics.width, box: boxOf(emptyMetrics) },
      full: { width: fullMetrics.width, box: boxOf(fullMetrics) },
      repeat: { width: repeatMetrics.width },
      prefixWidths: prefixWidths
    };
  }
}
  ]);
})();
