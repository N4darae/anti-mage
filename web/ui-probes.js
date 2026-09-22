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
