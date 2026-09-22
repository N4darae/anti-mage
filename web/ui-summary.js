var AMSummary = (function () {
  "use strict";

  var DETERMINATIONS = { "not-evaluated": 1, insufficient: 1, coherent: 1, discrepant: 1, instrumented: 1 };

  var SCALE_NOTE =
    "One score for the whole scan, in steps of ten, raised only by evidence, and by how little other than a deliberate change accounts " +
    "for that evidence. Each further finding raises what the ones before it left, so the scale is not a total, never falls because " +
    "other readings agreed, and does not reach its top. The strongest statement this page makes is that an environment appears " +
    "modified; that describes the environment, not the person using it.";

  function determinationOf(det) {
    var d = det === undefined || det === null ? "" : String(det);
    return DETERMINATIONS[d] ? d : "not-evaluated";
  }

  function marker(det) {
    var d = determinationOf(det);
    return AMDom.h("span", { class: "det", "data-det": d, text: d });
  }

  function intOrDash(v) {
    return typeof v === "number" && isFinite(v) ? String(Math.round(v)) : "—";
  }

  function suppliedCount(a) {
    return a && Array.isArray(a.supplied) ? a.supplied.length : null;
  }

  function numberPair(label, v) {
    return AMDom.h("div", null, AMDom.h("span", { class: "n-label", text: label }), AMDom.h("span", { class: "n-value", text: intOrDash(v) }));
  }

  function failureBand(message) {
    return AMDom.h(
      "div",
      { class: "band", "data-det": "not-evaluated" },
      AMDom.h(
        "div",
        { class: "band-top" },
        AMDom.h("p", { class: "band-statement", text: "The local server returned no assessment." }),
        marker("not-evaluated")
      ),
      AMDom.h("p", { class: "band-note", text: message })
    );
  }

  function assessmentBand(a) {
    var shown = determinationOf(a.determination);
    return AMDom.h(
      "div",
      { class: "band", "data-det": shown },
      AMDom.h(
        "div",
        { class: "band-top" },
        AMDom.h("p", { class: "band-statement", text: String(a.statement || shown) }),
        marker(shown)
      ),
      AMDom.h("div", { class: "band-nums" }, numberPair("score", a.score), numberPair("observations supplied", suppliedCount(a))),
      AMDom.h("p", { class: "band-note", text: SCALE_NOTE })
    );
  }

  function band(assessment, serverError) {
    if (serverError) return failureBand(serverError);
    if (!assessment) return null;
    return assessmentBand(assessment);
  }

  function probeTally(payload) {
    if (!payload) return "not yet run";
    var counts = AMProbes.statusCounts(payload);
    return counts.ok + " reported, " + counts.unsupported + " unsupported, " + counts.error + " in error, of " + payload.ids.length;
  }

  function bootstrapNote(payload) {
    if (!payload) return "not yet read";
    return payload.bootstrapSource + (payload.nonce ? ", nonce " + payload.nonce : "");
  }

  function facts(payload) {
    var nodes = [];
    [
      ["page origin", location.origin],
      ["user agent, as reported", navigator.userAgent],
      ["report generated", new Date().toString()],
      ["probes", probeTally(payload)],
      ["page bootstrap", bootstrapNote(payload)]
    ].forEach(function (pr) {
      nodes.push(AMDom.h("dt", { text: pr[0] }));
      nodes.push(AMDom.h("dd", { text: String(pr[1]) }));
    });
    return nodes;
  }

  return { band: band, facts: facts, marker: marker };
})();
