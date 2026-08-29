
(function () {
  "use strict";

  var state = { payload: null, assessment: null, serverError: null };
  var uidCounter = 0;

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

  function vtext(v) {
    if (v === null || v === undefined) return { s: "—", c: "v-none" };
    if (v === true) return { s: "true", c: "v-true" };
    if (v === false) return { s: "false", c: "v-false" };
    if (Array.isArray(v)) return v.length ? { s: v.join(", "), c: "" } : { s: "(none)", c: "v-none" };
    if (typeof v === "object") return { s: JSON.stringify(v), c: "" };
    if (v === "") return { s: "(empty)", c: "v-none" };
    return { s: String(v), c: "" };
  }

  function vcell(v, cls) {
    cls = cls || "";
    var r = vtext(v);
    var base = /prose/.test(cls) ? "" : "v";
    return h("td", { class: [base, cls, r.c].join(" ").trim(), text: r.s });
  }

  function kvTable(pairs) {
    var tb = h("tbody");
    pairs.forEach(function (p) {
      if (!p) return;
      tb.appendChild(h("tr", null, h("th", { text: p[0] }), vcell(p[1], p[2])));
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
          row.appendChild(vcell(c.v, [cls, c.cls || ""].join(" ").trim()));
        } else {
          row.appendChild(vcell(c, cls));
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

  function absentBlock(title, id) {
    var reason = preason(id);
    return block(title, h("p", { class: "absent", text: id + ": " + pstatus(id) + (reason ? " — " + reason : "") }));
  }

  var DETERMINATIONS = { "not-evaluated": 1, insufficient: 1, coherent: 1, discrepant: 1, instrumented: 1 };

  function detMarker(det) {
    var d = det && DETERMINATIONS[det] ? det : "not-evaluated";
    return h("span", { class: "det", "data-det": d, text: d });
  }

  function probe(id) {
    return state.payload && state.payload.probes ? state.payload.probes[id] : null;
  }
  function pv(id) {
    var p = probe(id);
    return p && p.status === "ok" ? p.value : null;
  }
  function pstatus(id) {
    var p = probe(id);
    return p ? p.status : "not reported";
  }
  function preason(id) {
    var p = probe(id);
    return p && p.value && p.value.reason ? p.value.reason : null;
  }
  function pmeta(name) {
    var m = state.payload && state.payload.meta;
    return m && m[name] ? m[name] : null;
  }

  function probeStrip(ids) {
    var strip = h("div", { class: "probes" });
    ids.forEach(function (id) {
      var st = pstatus(id);
      strip.appendChild(
        h("span", { class: "pr s-" + st.replace(/\s+/g, "-"), title: preason(id) || "" }, id + " ", h("b", { text: st }))
      );
    });
    return strip;
  }

  function renderPlatform() {
    var m = pv("scope.main");
    if (!m) return [absentBlock("Platform claim", "scope.main")];
    return [
      block(
        "what this browser says it is",
        kvTable([
          ["navigator.userAgent", m.userAgent],
          ["navigator.platform", m.platform],
          ["navigator.userAgentData.platform", m.uaDataPlatform],
          ["navigator.userAgentData.mobile", m.uaDataMobile],
          ["navigator.language", m.language],
          ["navigator.languages", m.languages],
          ["navigator.hardwareConcurrency", m.hardwareConcurrency],
          ["navigator.deviceMemory present", m.deviceMemoryPresent],
          ["navigator.deviceMemory", m.deviceMemory],
          ["navigator.webdriver", m.webdriver],
          ["isSecureContext", m.isSecureContext],
          ["location.origin", m.origin]
        ]),
        "These are the surfaces on which a browser names its operating system. Every other section reads its observations against this claim rather than against a claim of its own."
      )
    ];
  }

  function renderGeom() {
    var out = [];
    var g = pv("geom.screen");
    var c = pv("geom.css");
    if (!g) out.push(absentBlock("Screen and window", "geom.screen"));
    else {
      out.push(
        block(
          "screen and window, from JavaScript",
          kvTable([
            ["screen.width", g.width],
            ["screen.height", g.height],
            ["screen.availWidth", g.availWidth],
            ["screen.availHeight", g.availHeight],
            ["screen.colorDepth", g.colorDepth],
            ["screen.pixelDepth", g.pixelDepth],
            ["window.innerWidth", g.innerWidth],
            ["window.innerHeight", g.innerHeight],
            ["window.outerWidth", g.outerWidth],
            ["window.outerHeight", g.outerHeight],
            ["window.screenX", g.screenX],
            ["window.screenY", g.screenY],
            ["window.screenLeft", g.screenLeft],
            ["window.screenTop", g.screenTop],
            ["window.devicePixelRatio", g.devicePixelRatio],
            ["screen.orientation.type", g.orientationType],
            ["screen.orientation.angle", g.orientationAngle],
            ["visualViewport.scale", g.visualViewportScale]
          ])
        )
      );
    }
    if (!c) out.push(absentBlock("The same numbers, as CSS reports them", "geom.css"));
    else {
      out.push(
        block(
          "the same numbers, as CSS reports them",
          kvTable([
            ["resolution, narrowed in dppx", c.resolution],
            ["-webkit-device-pixel-ratio, narrowed", c.webkitDevicePixelRatio],
            ["device-width, narrowed in px", c.deviceWidthPx],
            ["device-height, narrowed in px", c.deviceHeightPx],
            ["viewport width, narrowed in px", c.viewportWidthPx],
            ["viewport height, narrowed in px", c.viewportHeightPx],
            ["color, bits per component", c.colorBitsPerComponent],
            ["monochrome, bits", c.monochromeBits],
            ["(resolution: devicePixelRatio) matches exactly", c.exactResolutionMatchesDpr],
            ["(-webkit-device-pixel-ratio: devicePixelRatio) matches exactly", c.exactWebkitMatchesDpr],
            ["orientation", c.orientation],
            ["pointer", c.pointer],
            ["any-pointer", c.anyPointer],
            ["hover", c.hover],
            ["any-hover", c.anyHover],
            ["prefers-color-scheme", c.prefersColorScheme],
            ["prefers-reduced-motion", c.prefersReducedMotion],
            ["forced-colors", c.forcedColors],
            ["display-mode", c.displayMode],
            ["scripting", c.scripting]
          ]),
          "Range media features are narrowed by bisection over min- queries. A feature that matches across the whole search range, or across none of it, has located no value and is reported as having located none."
        )
      );
    }
    if (g && c) {
      var near = function (a, b, tol) {
        if (a === null || a === undefined || b === null || b === undefined) return null;
        return Math.abs(a - b) <= tol;
      };
      var cssBits = c.colorBitsPerComponent === null || c.colorBitsPerComponent === undefined ? null : c.colorBitsPerComponent * 3;
      var pairs = [
        ["device pixel ratio", g.devicePixelRatio, c.resolution, near(g.devicePixelRatio, c.resolution, 0.01)],
        ["device pixel ratio, prefixed feature", g.devicePixelRatio, c.webkitDevicePixelRatio, near(g.devicePixelRatio, c.webkitDevicePixelRatio, 0.01)],
        ["screen width in px", g.width, c.deviceWidthPx, near(g.width, c.deviceWidthPx, 1)],
        ["screen height in px", g.height, c.deviceHeightPx, near(g.height, c.deviceHeightPx, 1)],
        ["viewport width in px", g.innerWidth, c.viewportWidthPx, near(g.innerWidth, c.viewportWidthPx, 1)],
        ["viewport height in px", g.innerHeight, c.viewportHeightPx, near(g.innerHeight, c.viewportHeightPx, 1)],
        ["colour depth in bits", g.colorDepth, cssBits, near(g.colorDepth, cssBits, 0)]
      ];
      out.push(
        block(
          "read two ways",
          dataTable(
            ["quantity", { label: "JavaScript", cls: "num" }, { label: "CSS media query", cls: "num" }, "agree"],
            pairs.map(function (p) {
              return { cells: [p[0], { v: p[1], cls: "num" }, { v: p[2], cls: "num" }, { v: p[3], cls: p[3] === false ? "v-flag" : "" }], flag: p[3] === false };
            })
          ),
          "The CSS column comes from the narrowing above, so each row is one quantity asked for through two interfaces."
        )
      );
    }
    return out;
  }

  function renderTime() {
    var out = [];
    var z = pv("time.zone");
    var o = pv("time.offsets");
    var ometa = pmeta("timeOffsets");
    if (!z) out.push(absentBlock("Zone and locale", "time.zone"));
    else {
      out.push(
        block(
          "zone and locale",
          kvTable([
            ["Intl.DateTimeFormat timeZone", z.timeZone],
            ["Intl.DateTimeFormat locale", z.locale],
            ["calendar", z.calendar],
            ["numbering system", z.numberingSystem],
            ["hour cycle", z.hourCycle],
            ["zone name, long offset", z.longOffset],
            ["zone name, long", z.longName],
            ["zone name, short", z.shortName],
            ["getTimezoneOffset() now, in minutes", z.currentOffsetMinutes],
            ["Date.prototype.toString()", z.dateToString],
            ["Date.prototype.toLocaleString()", z.dateToLocaleString],
            ["navigator.language", z.navigatorLanguage],
            ["navigator.languages", z.navigatorLanguages]
          ])
        )
      );
    }
    if (!o || !o.length) out.push(absentBlock("Offsets across dates", "time.offsets"));
    else {
      out.push(
        block(
          "offsets across dates",
          kvTable([
            ["dates supplied by", ometa ? ometa.datesFrom : null],
            ["zone they were read in", ometa ? ometa.timeZone : null],
            ["dates sampled", o.length],
            ["dates that did not parse", ometa ? ometa.unparsedCount : null],
            ["distinct offsets seen, in minutes", ometa ? ometa.distinctOffsets : null]
          ]),
          "The dates come from the server with the page, so the set asked about is not fixed in this script. A bare calendar date is sampled at noon UTC. Each date is read twice: once through the browser's own offset arithmetic, and once through the formatter for the zone the browser names."
        )
      );
    }
    return out;
  }

  function renderAuto() {
    var a = pv("auto.residue");
    if (!a) return [absentBlock("Remote-control surface", "auto.residue")];
    var out = [];
    var wd = a.webdriverDescriptor || {};
    out.push(
      block(
        "navigator.webdriver",
        kvTable([
          ["value", a.webdriver === undefined ? null : a.webdriver],
          ["property present on navigator", a.navigatorWebdriverPresent],
          ["descriptor on Navigator.prototype", wd.onPrototype === undefined ? null : wd.onPrototype],
          ["descriptor kind", wd.kind || null],
          ["configurable", wd.configurable === undefined ? null : wd.configurable],
          ["getter ends in [native code]", wd.getterNative === undefined ? null : wd.getterNative, wd.getterNative === false ? "v-flag" : ""]
        ]),
        "The HTML standard defines navigator.webdriver as the browser's own statement about whether it is under remote control. A browser with no such member reports no value here."
      )
    );

    function stackShape(s) {
      if (!s) return null;
      return [s.present ? "a string" : "not a string", s.name || "no name", (s.frameCount === undefined ? "?" : s.frameCount) + " frames", s.framesWellFormed === false ? "a frame line is malformed" : "every frame line well formed"].join(", ");
    }
    var hits = a.patternHits || [];
    out.push(
      block(
        "error stack shape",
        kvTable([
          ["stack of a property read on null", stackShape(a.errorStackNative), a.errorStackNative && a.errorStackNative.framesWellFormed === false ? "v-flag" : ""],
          ["stack of a throw two calls deep", stackShape(a.errorStackUser), a.errorStackUser && a.errorStackUser.framesWellFormed === false ? "v-flag" : ""],
          ["Error.prototype.stack descriptor", a.errorStackDescriptor ? JSON.stringify(a.errorStackDescriptor) : null],
          ["typeof Error.captureStackTrace", a.captureStackTraceType],
          ["typeof Error.prepareStackTrace", a.prepareStackTraceType],
          ["Error.stackTraceLimit", a.stackTraceLimit],
          [
            "global names matching an injected-hook shape",
            hits.length
              ? hits.map(function (x) {
                  return x.name;
                })
              : "none",
            hits.length ? "v-flag" : ""
          ]
        ]),
        "A stack is produced twice, once by the engine's own failure and once by a throw of this script's own, so the two forms can be read against each other. The hook shapes probed describe how a name is formed rather than where it came from."
      )
    );

    var prefixed = (a.windowPrefixedKeys || []).concat(a.documentPrefixedKeys || []);
    out.push(
      block(
        "counts and other globals",
        kvTable([
          ["own property names on window", a.windowKeyCount],
          ["own property names on document", a.documentKeyCount],
          ["globals whose name starts with _ or $", prefixed.length ? prefixed : null]
        ]),
        "An extension the reader installed on purpose puts names of that last shape on the page, so they are listed as an observation."
      )
    );
    return out;
  }

  function renderPerm() {
    var p = pv("perm.state");
    if (!p) return [absentBlock("Permissions", "perm.state")];
    var out = [];
    var states = p.states || [];
    out.push(
      block(
        "Permissions.query",
        kvTable([
          ["permission names queried", states.length],
          [
            "states returned",
            states.length
              ? Object.keys(
                  states.reduce(function (acc, r) {
                    acc[r.state || "no state"] = 1;
                    return acc;
                  }, {})
                ).sort()
              : null
          ],
          [
            "names this browser did not recognise",
            states.filter(function (r) {
              return !!r.error;
            }).length
          ]
        ]),
        "A name this browser does not recognise rejects the query. That is what the browser supports, and it is counted here rather than listed."
      )
    );
    var n = p.notification || {};
    var g = p.geolocation || {};
    out.push(
      block(
        "the query against the interface itself",
        kvTable([
          ["Notification.permission", n.apiValue],
          ["Permissions.query('notifications')", n.queryState],
          ["the two agree", n.agree, n.agree === false ? "v-flag" : ""],
          ["Permissions.query('geolocation')", g.queryState],
          ["Geolocation interface present", g.apiPresent],
          ["interface called", g.exercised],
          ["outcome", g.outcome],
          ["outcome matches the query", g.agree, g.agree === false ? "v-flag" : ""]
        ]),
        "The geolocation interface is called only where the query has already reported the permission denied, which is the one state in which the call is answered from that state alone: it raises no dialog, and it asks nothing off this machine."
      )
    );
    return out;
  }

  var GROUPS = [
    { id: "platform", title: "Platform claim", probes: ["scope.main"], render: renderPlatform, scale: null },
    { id: "geom", title: "Display geometry", probes: ["geom.screen", "geom.css"], render: renderGeom, scale: null },
    {
      id: "time",
      title: "Time zone and offsets",
      probes: ["time.zone", "time.offsets"],
      render: renderTime,
      scale: function () {
        var o = pv("time.offsets");
        return o && o.length ? o.length + " dates sampled" : null;
      }
    },
    { id: "auto", title: "Remote-control surface", probes: ["auto.residue"], render: renderAuto, scale: null },
    { id: "perm", title: "Permissions", probes: ["perm.state"], render: renderPerm, scale: null }
  ];

  function sectionNode(group) {
    var content = group.render();
    if (content === null && !(group.probes || []).length) return null;
    var d = h("details", { class: "sec", id: "sec-" + group.id });
    var scale = group.scale ? group.scale() : null;
    d.appendChild(h("summary", null, h("h2", { text: group.title }), scale ? h("span", { class: "sec-scale", text: scale }) : null));
    var body = h("div", { class: "sec-body" }, probeStrip(group.probes));
    append(body, content);
    d.appendChild(body);
    return d;
  }

  function intOrDash(v) {
    return typeof v === "number" && isFinite(v) ? String(Math.round(v)) : "—";
  }

  function assessmentNode() {
    var host = document.getElementById("band");
    clear(host);
    if (state.serverError) {
      host.appendChild(
        h(
          "div",
          { class: "band", "data-det": "not-evaluated" },
          h(
            "div",
            { class: "band-top" },
            h("p", { class: "band-statement", text: "The local server returned no assessment." }),
            detMarker("not-evaluated")
          ),
          h("p", { class: "band-note", text: state.serverError })
        )
      );
      return;
    }
    var a = state.assessment;
    if (!a) return;
    var det = String(a.determination === undefined || a.determination === null ? "" : a.determination);
    var shown = DETERMINATIONS[det] ? det : "not-evaluated";
    host.appendChild(
      h(
        "div",
        { class: "band", "data-det": shown },
        h(
          "div",
          { class: "band-top" },
          h("p", { class: "band-statement", text: String(a.statement || shown) }),
          h("span", { class: "det", "data-det": shown, text: shown })
        ),
        h(
          "div",
          { class: "band-nums" },
          h("div", null, h("span", { class: "n-label", text: "score" }), h("span", { class: "n-value", text: intOrDash(a.score) })),
          h("div", null, h("span", { class: "n-label", text: "observations supplied" }), h("span", { class: "n-value", text: intOrDash(suppliedCount(a)) }))
        ),
        h("p", {
          class: "band-note",
          text:
            "One score for the whole scan, in steps of ten, raised only by evidence that disagrees with itself. It is not a probability, " +
            "and it is not a total of per-reading weights. The strongest statement this page makes is that an environment appears modified; " +
            "that describes the environment, not the person using it."
        })
      )
    );
  }

  function suppliedCount(a) {
    return a && Array.isArray(a.supplied) ? a.supplied.length : null;
  }

  function indexNode(sections) {
    var ol = h("ol");
    sections.forEach(function (sec) {
      var title = sec.querySelector("summary h2");
      var scale = sec.querySelector("summary .sec-scale");
      ol.appendChild(
        h(
          "li",
          null,
          h("a", { href: "#" + sec.id, text: title ? title.textContent : sec.id }),
          scale ? h("span", { class: "idx-count", text: scale.textContent }) : null
        )
      );
    });
    return h("nav", { class: "index" }, h("h2", { text: "what was observed" }), ol);
  }

  function renderAll() {
    var host = document.getElementById("sections");
    clear(host);
    var nodes = GROUPS.map(sectionNode).filter(function (n) {
      return n !== null;
    });
    var idx = document.getElementById("index");
    clear(idx);
    idx.appendChild(indexNode(nodes));
    nodes.forEach(function (n) {
      host.appendChild(n);
    });
    labelToggle();
    assessmentNode();
    renderFacts();
  }

  function renderFacts() {
    var dl = document.getElementById("facts");
    clear(dl);
    var p = state.payload;
    var ok = 0,
      uns = 0,
      err = 0;
    if (p) {
      Object.keys(p.probes).forEach(function (k) {
        var st = p.probes[k].status;
        if (st === "ok") ok += 1;
        else if (st === "unsupported") uns += 1;
        else err += 1;
      });
    }
    [
      ["page origin", location.origin],
      ["user agent, as reported", navigator.userAgent],
      ["report generated", new Date().toString()],
      ["probes", p ? ok + " reported, " + uns + " unsupported, " + err + " in error, of " + p.ids.length : "not yet run"],
      ["page bootstrap", p ? p.bootstrapSource + (p.nonce ? ", nonce " + p.nonce : "") : "not yet read"]
    ].forEach(function (pr) {
      dl.appendChild(h("dt", { text: pr[0] }));
      dl.appendChild(h("dd", { text: String(pr[1]) }));
    });
  }

  function setStatus(s) {
    document.getElementById("status").textContent = s;
  }

  function requestBody() {
    var body = { v: 1, probes: state.payload.probes };
    if (state.payload.nonce) body.nonce = state.payload.nonce;
    return body;
  }

  function post() {
    return fetch("./api/scan", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
      credentials: "omit",
      body: JSON.stringify(requestBody())
    }).then(function (r) {
      if (!r.ok) throw new Error("the local server answered HTTP " + r.status);
      return r.json();
    });
  }

  function runScan() {
    var run = document.getElementById("run");
    run.disabled = true;
    state.assessment = null;
    state.serverError = null;
    setStatus("reading the page bootstrap");
    AM.run(function (done, total, id) {
      setStatus("running probes " + done + " of " + total + (id ? "   " + id : ""));
    })
      .then(function (payload) {
        state.payload = payload;
        renderAll();
        setStatus("probes complete; asking the local server to assess them");
        return post();
      })
      .then(function (res) {
        state.assessment = res && res.v === 1 ? res : null;
        if (!state.assessment) state.serverError = "the local server returned a payload this page does not recognise";
        setStatus(state.assessment ? "complete" : "complete, without an assessment");
        renderAll();
      })
      .catch(function (e) {
        state.serverError = String((e && e.message) || e);
        setStatus("probes complete; " + state.serverError);
        renderAll();
      })
      .then(function () {
        run.disabled = false;
      });
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
    document.getElementById("toggle").textContent = anyClosed() ? "Expand all" : "Collapse all";
  }

  function init() {
    document.getElementById("run").addEventListener("click", runScan);
    document.getElementById("toggle").addEventListener("click", function () {
      setAll(anyClosed());
    });
    document.addEventListener("toggle", function (e) {
      if (e.target && e.target.classList && e.target.classList.contains("sec")) labelToggle();
    }, true);
    renderFacts();
    runScan();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
