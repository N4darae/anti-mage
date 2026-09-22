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

var AMProbes = (function () {
  "use strict";

  function reader(payload) {
    function probe(id) {
      return payload && payload.probes ? payload.probes[id] : null;
    }

    function value(id) {
      var p = probe(id);
      return p && p.status === "ok" ? p.value : null;
    }

    function status(id) {
      var p = probe(id);
      return p ? p.status : "not reported";
    }

    function reason(id) {
      var p = probe(id);
      return p && p.value && p.value.reason ? p.value.reason : null;
    }

    function meta(name) {
      var m = payload && payload.meta;
      return m && m[name] ? m[name] : null;
    }

    function strip(ids) {
      var el = AMDom.h("div", { class: "probes" });
      ids.forEach(function (id) {
        var st = status(id);
        el.appendChild(
          AMDom.h(
            "span",
            { class: "pr s-" + st.replace(/\s+/g, "-"), title: reason(id) || "" },
            id + " ",
            AMDom.h("b", { text: st })
          )
        );
      });
      return el;
    }

    function absentBlock(title, id) {
      var why = reason(id);
      return AMDom.block(title, AMDom.h("p", { class: "absent", text: id + ": " + status(id) + (why ? " — " + why : "") }));
    }

    return {
      value: value,
      status: status,
      reason: reason,
      meta: meta,
      strip: strip,
      absentBlock: absentBlock
    };
  }

  function statusCounts(payload) {
    var counts = { ok: 0, unsupported: 0, error: 0 };
    if (!payload || !payload.probes) return counts;
    Object.keys(payload.probes).forEach(function (id) {
      var st = payload.probes[id].status;
      if (st === "ok") counts.ok += 1;
      else if (st === "unsupported") counts.unsupported += 1;
      else counts.error += 1;
    });
    return counts;
  }

  return { reader: reader, statusCounts: statusCounts };
})();

(function () {
  "use strict";

  var state = { payload: null, assessment: null, serverError: null };

  function el(id) {
    return document.getElementById(id);
  }

  function setStatus(s) {
    el("status").textContent = s;
  }

  function mountSections() {
    var host = el("sections");
    AMDom.clear(host);
    var nodes = AMSections.build(state.payload);
    var idx = el("index");
    AMDom.clear(idx);
    idx.appendChild(AMSections.index(nodes));
    nodes.forEach(function (n) {
      host.appendChild(n);
    });
  }

  function mountBand() {
    var host = el("band");
    AMDom.clear(host);
    AMDom.append(host, AMSummary.band(state.assessment, state.serverError));
  }

  function mountFacts() {
    var host = el("facts");
    AMDom.clear(host);
    AMDom.append(host, AMSummary.facts(state.payload));
  }

  function render() {
    mountSections();
    labelToggle();
    mountBand();
    mountFacts();
  }

  function sections() {
    return document.querySelectorAll("details.sec");
  }

  function anyClosed() {
    var list = sections();
    for (var i = 0; i < list.length; i++) {
      if (!list[i].hasAttribute("open")) return true;
    }
    return false;
  }

  function setAll(open) {
    var list = sections();
    for (var i = 0; i < list.length; i++) {
      if (open) list[i].setAttribute("open", "");
      else list[i].removeAttribute("open");
    }
    labelToggle();
  }

  function labelToggle() {
    el("toggle").textContent = anyClosed() ? "Expand all" : "Collapse all";
  }

  function reportProgress(done, total, id) {
    setStatus("running probes " + done + " of " + total + (id ? "   " + id : ""));
  }

  function probesDone(payload) {
    state.payload = payload;
    render();
    setStatus("probes complete; asking the local server to assess them");
    return AMTransport.assess(payload);
  }

  function assessmentDone(res) {
    state.assessment = res && res.v === 1 ? res : null;
    if (!state.assessment) state.serverError = "the local server returned a payload this page does not recognise";
    setStatus(state.assessment ? "complete" : "complete, without an assessment");
    render();
  }

  function scanFailed(e) {
    state.serverError = String((e && e.message) || e);
    setStatus("probes complete; " + state.serverError);
    render();
  }

  function runScan() {
    var run = el("run");
    run.disabled = true;
    state.assessment = null;
    state.serverError = null;
    setStatus("reading the page bootstrap");
    AM.run(reportProgress)
      .then(probesDone)
      .then(assessmentDone)
      .catch(scanFailed)
      .then(function () {
        run.disabled = false;
      });
  }

  function init() {
    el("run").addEventListener("click", runScan);
    el("toggle").addEventListener("click", function () {
      setAll(anyClosed());
    });
    document.addEventListener("toggle", function (e) {
      if (e.target && e.target.classList && e.target.classList.contains("sec")) labelToggle();
    }, true);
    mountFacts();
    runScan();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
