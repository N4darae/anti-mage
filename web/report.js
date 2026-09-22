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
