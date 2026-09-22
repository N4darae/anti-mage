var AMDom = (function () {
  "use strict";

  function h(tag, attrs) {
    var e = document.createElement(tag);
    if (attrs) {
      for (var k in attrs) {
        if (!Object.prototype.hasOwnProperty.call(attrs, k)) continue;
        var v = attrs[k];
        if (v === null || v === undefined) continue;
        if (k === "text") e.textContent = String(v);
        else if (k === "class") e.className = v;
        else e.setAttribute(k, String(v));
      }
    }
    for (var i = 2; i < arguments.length; i++) append(e, arguments[i]);
    return e;
  }

  function append(parent, c) {
    if (c === null || c === undefined || c === false) return;
    if (Array.isArray(c)) {
      c.forEach(function (x) {
        append(parent, x);
      });
      return;
    }
    if (typeof c === "string" || typeof c === "number") {
      parent.appendChild(document.createTextNode(String(c)));
      return;
    }
    parent.appendChild(c);
  }

  function clear(node) {
    while (node.firstChild) node.removeChild(node.firstChild);
  }

  function valueText(v) {
    if (v === null || v === undefined) return { s: "—", c: "v-none" };
    if (v === true) return { s: "true", c: "v-true" };
    if (v === false) return { s: "false", c: "v-false" };
    if (Array.isArray(v)) return v.length ? { s: v.join(", "), c: "" } : { s: "(none)", c: "v-none" };
    if (typeof v === "object") return { s: JSON.stringify(v), c: "" };
    if (v === "") return { s: "(empty)", c: "v-none" };
    return { s: String(v), c: "" };
  }

  function valueCell(v, cls) {
    cls = cls || "";
    var r = valueText(v);
    var base = /prose/.test(cls) ? "" : "v";
    return h("td", { class: [base, cls, r.c].join(" ").trim(), text: r.s });
  }

  function kvTable(pairs) {
    var tb = h("tbody");
    pairs.forEach(function (p) {
      if (!p) return;
      tb.appendChild(h("tr", null, h("th", { text: p[0] }), valueCell(p[1], p[2])));
    });
    return h("div", { class: "scroll" }, h("table", { class: "kv" }, tb));
  }

  function dataTable(headers, rows) {
    var tr = h("tr");
    headers.forEach(function (hd) {
      tr.appendChild(h("th", { text: typeof hd === "string" ? hd : hd.label }));
    });
    var tb = h("tbody");
    rows.forEach(function (r) {
      var cells = Array.isArray(r) ? r : r.cells;
      var row = h("tr");
      if (!Array.isArray(r) && r.flag) row.setAttribute("data-flag", "1");
      cells.forEach(function (c, i) {
        var cls = (headers[i] && headers[i].cls) || "";
        if (c && typeof c === "object" && !Array.isArray(c) && Object.prototype.hasOwnProperty.call(c, "v")) {
          row.appendChild(valueCell(c.v, [cls, c.cls || ""].join(" ").trim()));
        } else {
          row.appendChild(valueCell(c, cls));
        }
      });
      tb.appendChild(row);
    });
    return h("div", { class: "scroll" }, h("table", null, h("thead", null, tr), tb));
  }

  function block(title, nodes, prose) {
    var b = h("section", { class: "block" }, h("h3", { text: title }));
    if (prose) b.appendChild(h("p", { text: prose }));
    append(b, nodes);
    return b;
  }

  return {
    h: h,
    append: append,
    clear: clear,
    valueText: valueText,
    valueCell: valueCell,
    kvTable: kvTable,
    dataTable: dataTable,
    block: block
  };
})();
