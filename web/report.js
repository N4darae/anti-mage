
(function () {
  "use strict";

  var state = { payload: null, assessment: null, serverError: null };
  var uidCounter = 0;

  function uid() {
    uidCounter += 1;
    return "am-" + uidCounter;
  }

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

  function dataTable(headers, rows, opts) {
    opts = opts || {};
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
    var tall = opts.tall || rows.length > 28;
    return h("div", { class: "scroll" + (tall ? " tall" : "") }, h("table", null, h("thead", null, tr), tb));
  }

  function withFilter(container, label, toggleLabel) {
    var table = container.querySelector("table");
    if (!table || !table.tBodies.length) return container;
    var rows = Array.prototype.slice.call(table.tBodies[0].rows);
    var input = h("input", { type: "text", placeholder: "filter", "aria-label": "filter " + label });
    var count = h("span", { class: "count" });
    var toggle = null;
    var toggleId = uid();
    var bar = h("div", { class: "filter" }, input);
    if (toggleLabel) {
      toggle = h("input", { type: "checkbox", id: toggleId });
      bar.appendChild(h("label", { for: toggleId }, toggle, toggleLabel));
    }
    bar.appendChild(count);

    function apply() {
      var q = input.value.toLowerCase();
      var onlyFlag = toggle && toggle.checked;
      var shown = 0;
      rows.forEach(function (row) {
        var ok = true;
        if (onlyFlag && row.getAttribute("data-flag") !== "1") ok = false;
        if (ok && q && row.textContent.toLowerCase().indexOf(q) < 0) ok = false;
        row.style.display = ok ? "" : "none";
        if (ok) shown += 1;
      });
      count.textContent = shown + " of " + rows.length + " rows";
    }
    input.addEventListener("input", apply);
    if (toggle) toggle.addEventListener("change", apply);
    apply();
    return h("div", null, bar, container);
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

  function sortedNames(obj, order) {
    var out = [];
    var seen = {};
    (order || []).forEach(function (k) {
      if (Object.prototype.hasOwnProperty.call(obj, k)) {
        out.push(k);
        seen[k] = 1;
      }
    });
    Object.keys(obj)
      .filter(function (k) {
        return !seen[k];
      })
      .sort()
      .forEach(function (k) {
        out.push(k);
      });
    return out;
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

  function renderFont() {
    var out = [];
    var fam = pv("font.resolved");
    var fmeta = pmeta("font");
    if (!fam) out.push(absentBlock("Width probe", "font.resolved"));
    else {
      var names = Object.keys(fam);
      var resolvedCount = names.filter(function (n) {
        return fam[n] && fam[n].width;
      }).length;
      out.push(
        block(
          "width probe",
          kvTable([
            ["candidate families", fmeta ? fmeta.probed : names.length],
            ["families resolved", resolvedCount],
            ["probe string", fmeta ? fmeta.measureString : null],
            ["generic families compared", fmeta ? fmeta.bases : null],
            ["candidate list from", fmeta ? fmeta.inputSource : null],
            [
              "probe inputs",
              fmeta && fmeta.inputSources
                ? Object.keys(fmeta.inputSources)
                    .map(function (k) {
                      return k + ": " + fmeta.inputSources[k];
                    })
                    .join("   ")
                : null
            ]
          ]),
          "A family counts as resolved when text set in it measures a different advance width from the same text set in a generic family alone. The comparison runs against all three CSS generic families, because a family whose metrics match one generic still differs from the other two."
        )
      );
    }

    var ctrl = pv("font.controls");
    var cmeta = pmeta("fontControls");
    if (!ctrl) out.push(absentBlock("Control names", "font.controls"));
    else {
      var cnames = sortedNames(ctrl, cmeta ? cmeta.order : null);
      var resolvedControls = cnames.filter(function (n) {
        return ctrl[n] && ctrl[n].width;
      });
      var checkTrue = cnames.filter(function (n) {
        return ctrl[n] && ctrl[n].check === true;
      });
      var checkSeen = cnames.some(function (n) {
        return ctrl[n] && ctrl[n].check !== null && ctrl[n].check !== undefined;
      });
      out.push(
        block(
          "control names",
          [
            kvTable([
              ["control names probed", cnames.length],
              ["resolved by width probe", resolvedControls.length, resolvedControls.length ? "v-flag" : ""],
              ["names resolved", resolvedControls.length ? resolvedControls : null],
              ["document.fonts.check() answered", checkSeen],
              ["names it answered true for", checkSeen ? checkTrue.length : null],
              ["control list from", cmeta ? cmeta.inputSource : null]
            ]),
            dataTable(
              ["control name", "width probe resolved", "document.fonts.check()"],
              cnames.map(function (n) {
                return { cells: [n, { v: ctrl[n].width, cls: ctrl[n].width ? "v-flag" : "" }, ctrl[n].check], flag: !!ctrl[n].width };
              })
            )
          ],
          "These names are invented, so no font by any of them is installed. They are carried through every font probe as its control. document.fonts.check() answers true for a name that does not exist, so its answer is recorded beside each width result and is not read as a result on its own."
        )
      );
    }

    var cov = pv("font.coverage");
    var covmeta = pmeta("fontCoverage");
    if (!cov) out.push(absentBlock("Script coverage", "font.coverage"));
    else {
      var covNames = sortedNames(cov, covmeta ? covmeta.order : null);
      out.push(
        block(
          "script coverage",
          dataTable(
            [
              "family",
              "codepoints asked for",
              "resolves for its own script",
              "resolves for Latin text",
              "the two agree",
              { label: "documented scripts", cls: "prose" }
            ],
            covNames.map(function (n) {
              var r = cov[n];
              return {
                cells: [n, r.codepoints, { v: r.covers }, { v: r.width }, { v: r.agree, cls: r.agree ? "" : "v-flag" }, { v: r.scripts, cls: "prose" }],
                flag: !r.agree
              };
            })
          ),
          "Each family is asked for codepoints from a script its vendor documents it as carrying. A width probe run on Latin text cannot see a font that draws only icons, or only one non-Latin script, so the two measurements are reported side by side."
        )
      );
    }

    if (fam) {
      out.push(
        block(
          "families, one row each",
          withFilter(
            dataTable(
              ["family", "width probe resolved", "document.fonts.check()"],
              Object.keys(fam)
                .sort()
                .map(function (n) {
                  return { cells: [n, { v: fam[n].width }, fam[n].check], flag: !!fam[n].width };
                }),
              { tall: true }
            ),
            "families",
            "resolved only"
          )
        )
      );
    }
    return out;
  }

  function renderNative() {
    var ts = pv("native.tostring");
    var ok = pv("native.ownkeys");
    var ds = pv("native.descriptor");
    var rc = pv("native.receiver");
    var nmeta = pmeta("native");
    if (!ts && !ok && !ds && !rc) return [absentBlock("Interface members", "native.tostring")];
    var out = [];

    var order = (nmeta && nmeta.order) || [];
    var keySet = {};
    [ts, ok, ds, rc].forEach(function (v) {
      if (!v) return;
      Object.keys(v).forEach(function (k) {
        keySet[k] = 1;
      });
    });
    var keys = sortedNames(keySet, order);

    var rows = keys.map(function (k) {
      var d = (ds && ds[k]) || null;
      var r = (rc && rc[k]) || null;
      var s = (ts && ts[k]) || null;
      var o = (ok && ok[k]) || null;
      var receiverText = null;
      var receiverFlag = "";
      if (r && !r.skipped && r.threw !== undefined) {
        if (r.threw) {
          receiverText = r.isTypeError ? "TypeError" : r.name || "threw";
          receiverFlag = r.isTypeError ? "" : "v-flag";
        } else {
          receiverText = "returned " + r.resultType;
          receiverFlag = "v-flag";
        }
      }
      var notes = [];
      [s, o, d, r].forEach(function (x) {
        if (x && x.reason && notes.indexOf(x.reason) < 0) notes.push(x.reason);
      });
      var onProto = d && d.onPrototype !== undefined ? d.onPrototype : null;
      var kindOK = d && d.kindAsExpected !== undefined ? d.kindAsExpected : null;
      var flagged =
        onProto === false ||
        kindOK === false ||
        (d && d.shadowedOnInstance === true) ||
        (s && s.native === false) ||
        (o && o.agree === false) ||
        receiverFlag === "v-flag";
      return {
        cells: [
          k,
          (d && d.declaredAs) || null,
          { v: onProto, cls: onProto === false ? "v-flag" : "" },
          { v: d ? d.kind || null : null, cls: kindOK === false ? "v-flag" : "" },
          { v: d ? (d.shadowedOnInstance === undefined ? null : d.shadowedOnInstance) : null, cls: d && d.shadowedOnInstance === true ? "v-flag" : "" },
          { v: s ? (s.native === undefined ? null : s.native) : null, cls: s && s.native === false ? "v-flag" : "" },
          { v: o ? (o.agree === undefined ? null : o.agree) : null, cls: o && o.agree === false ? "v-flag" : "" },
          { v: receiverText, cls: receiverFlag },
          { v: notes.length ? notes.join("; ") : null, cls: "prose" }
        ],
        flag: !!flagged
      };
    });

    out.push(
      block(
        "members, read four ways",
        dataTable(
          [
            "member",
            "declared as",
            "on the interface prototype",
            "descriptor kind",
            "shadowed on the instance",
            "toString ends in [native code]",
            "three key enumerators agree",
            "alien receiver",
            { label: "note", cls: "prose" }
          ],
          rows
        ),
        "WebIDL puts an interface member on the interface prototype object: an attribute as an accessor, an operation as a data property, and either one behind a brand check that makes an alien receiver a TypeError. Each member below is read against those four requirements. A member this browser does not have contributes no answer."
      )
    );

    if (ts) {
      out.push(
        block(
          "Function.prototype.toString output",
          dataTable(
            ["member", "name", { label: "length", cls: "num" }, { label: "chars", cls: "num" }, "own toString agrees", "source text"],
            keys
              .filter(function (k) {
                return ts[k];
              })
              .map(function (k) {
                var s = ts[k];
                return {
                  cells: [
                    k,
                    { v: s.fnName === undefined ? null : s.fnName },
                    { v: s.fnLength === undefined ? null : s.fnLength, cls: "num" },
                    { v: s.chars === undefined ? null : s.chars, cls: "num" },
                    { v: s.selfAgrees === undefined ? null : s.selfAgrees, cls: s.selfAgrees === false ? "v-flag" : "" },
                    { v: s.source === undefined ? s.reason || null : s.source, cls: "wide" }
                  ],
                  flag: s.native === false
                };
              })
          ),
          "ECMA-262 requires the string returned for a built-in function to be a NativeFunction. The function's own toString is read as well, so a member whose two answers differ is visible as such."
        )
      );
    }

    var g = nmeta && nmeta.globals;
    if (g) {
      out.push(
        block(
          "the reader itself",
          kvTable([
            ["Function.prototype.toString, described by itself", g.toStringOfToString],
            ["its own keys", g.toStringOwnKeys],
            ["its descriptor on Function.prototype", g.toStringDescriptor ? JSON.stringify(g.toStringDescriptor) : null]
          ]),
          "Every toString result above was produced by this function, so the function is described by itself first."
        )
      );
    }

    if (ok) {
      var keyRows = keys
        .filter(function (k) {
          return ok[k] && ok[k].ownKeys;
        })
        .map(function (k) {
          var o = ok[k];
          return {
            cells: [k, { v: o.ownKeys }, { v: o.getOwnPropertyNames }, { v: o.descriptors }, { v: o.symbolKeys }, { v: o.agree, cls: o.agree === false ? "v-flag" : "" }],
            flag: o.agree === false
          };
        });
      if (keyRows.length) {
        out.push(
          block(
            "own keys, by three enumerators",
            dataTable(
              ["member", "Reflect.ownKeys, string keys", "Object.getOwnPropertyNames", "Object.getOwnPropertyDescriptors", "symbol keys", "agree"],
              keyRows
            ),
            "The three enumerators walk the same own-property list by three different routes. Symbol keys are listed on their own, because only the first enumerator returns them."
          )
        );
      }
    }
    return out;
  }

  var SCOPE_COLUMNS = [
    ["scope.main", "main window", "main"],
    ["scope.worker", "dedicated worker", "worker"],
    ["scope.iframe", "same-origin frame", "iframe"]
  ];

  var SCOPE_FACTS = [
    "scope",
    "userAgent",
    "platform",
    "hardwareConcurrency",
    "deviceMemoryPresent",
    "deviceMemory",
    "language",
    "languages",
    "webdriver",
    "uaDataPlatform",
    "uaDataMobile",
    "timeZone",
    "locale",
    "calendar",
    "timezoneOffsetMinutes",
    "isSecureContext",
    "origin",
    "hasWindow"
  ];

  var SCOPE_DEFINING = { scope: 1, hasWindow: 1, origin: 1, isSecureContext: 1, deviceMemory: 1, deviceMemoryPresent: 1, uaDataPlatform: 1, uaDataMobile: 1 };

  function renderScope() {
    var out = [];
    var avail = pv("scope.availability");
    var values = SCOPE_COLUMNS.map(function (c) {
      return pv(c[0]);
    });
    if (!values.some(Boolean)) return [absentBlock("Execution scopes", "scope.main")];

    var rows = SCOPE_FACTS.map(function (fact) {
      var cells = [fact];
      var seen = [];
      values.forEach(function (v) {
        var val = v ? (fact in v ? v[fact] : null) : null;
        cells.push({ v: val });
        if (v) seen.push(JSON.stringify(val));
      });
      var agree = null;
      if (!SCOPE_DEFINING[fact] && seen.length > 1) {
        agree = seen.every(function (x) {
          return x === seen[0];
        });
      }
      cells.push({ v: agree, cls: agree === false ? "v-flag" : "" });
      return { cells: cells, flag: agree === false };
    });

    var headers = ["fact"];
    SCOPE_COLUMNS.forEach(function (c, i) {
      headers.push(c[1] + (values[i] ? "" : " (" + pstatus(c[0]) + ")"));
    });
    headers.push("agree");

    out.push(
      block(
        "the same question in three scopes",
        dataTable(headers, rows),
        "One function is read in three realms. The window and the frame are read through the frame's own global object; the worker receives that same function's source text. A scope that could not be created reports its status in the column heading and contributes no value. Facts gated on a scope's own origin, and facts that say which scope is being read, are shown without an agreement column because an honest browser answers them differently in each."
      )
    );

    if (avail) {
      out.push(
        block(
          "scope availability",
          dataTable(
            ["scope", "created", "status", { label: "reason", cls: "prose" }],
            Object.keys(avail).map(function (k) {
              return [k, { v: avail[k].created }, avail[k].status, { v: avail[k].reason, cls: "prose" }];
            })
          ),
          "A content policy on the page, or a browser that does not offer the scope, lands here. That is a property of how the page was delivered."
        )
      );
    }
    return out;
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
          [
            kvTable([
              ["dates supplied by", ometa ? ometa.datesFrom : null],
              ["zone they were read in", ometa ? ometa.timeZone : null],
              ["dates sampled", o.length],
              ["dates that did not parse", ometa ? ometa.unparsedCount : null],
              ["distinct offsets seen, in minutes", ometa ? ometa.distinctOffsets : null]
            ]),
            dataTable(
              ["date", "instant sampled", { label: "getTimezoneOffset(), minutes", cls: "num" }, "long offset for the named zone", "Date.prototype.toString()"],
              o.map(function (r) {
                return [r.date, r.instant, { v: r.offsetMinutes, cls: "num" }, r.longOffset, { v: r.localString, cls: "wide" }];
              })
            )
          ],
          "The dates come from the server with the page, so the set asked about is not fixed in this script. A bare calendar date is sampled at noon UTC. Each date is read twice: once through the browser's own offset arithmetic, and once through the formatter for the zone the browser names."
        )
      );
    }
    return out;
  }

  var RECORDER_PREFIX = "MediaRecorder ";
  var FACILITY_KEY = "interfaces available";

  function renderMedia() {
    var m = pv("media.matrix");
    if (!m) return [absentBlock("Codec matrix", "media.matrix")];
    var mmeta = pmeta("media");
    var out = [];
    var fac = m[FACILITY_KEY] || (mmeta && mmeta.facilities) || null;
    if (fac) {
      out.push(
        block(
          "interfaces available",
          kvTable([
            ["HTMLMediaElement.canPlayType", fac.canPlayType],
            ["MediaSource.isTypeSupported", fac.mediaSource],
            ["MediaRecorder.isTypeSupported", fac.mediaRecorder],
            ["MediaCapabilities.decodingInfo", fac.mediaCapabilities]
          ])
        )
      );
    }
    var codecKeys = sortedNames(m, mmeta ? mmeta.codecOrder : null).filter(function (k) {
      return k !== FACILITY_KEY && k.indexOf(RECORDER_PREFIX) !== 0;
    });
    out.push(
      block(
        "one codec per row, three interfaces per codec",
        dataTable(
          ["codec", "content type", "canPlayType", "MediaSource", "decodingInfo supported", "smooth", "power efficient"],
          codecKeys.map(function (k) {
            var r = m[k] || {};
            return [
              k,
              { v: r.contentType === undefined ? null : r.contentType, cls: "wide" },
              { v: r.canPlayType === undefined ? null : r.canPlayType },
              { v: r.mediaSource === undefined ? null : r.mediaSource },
              { v: r.decodingInfoError !== undefined ? r.decodingInfoError : r.decodingInfoSupported === undefined ? null : r.decodingInfoSupported },
              { v: r.smooth === undefined ? null : r.smooth },
              { v: r.powerEfficient === undefined ? null : r.powerEfficient }
            ];
          })
        ),
        "The three interfaces answer for the same content type, and a build that carries a decoder answers for it consistently across them."
      )
    );
    var recKeys = Object.keys(m)
      .filter(function (k) {
        return k.indexOf(RECORDER_PREFIX) === 0;
      })
      .sort();
    if (recKeys.length) {
      out.push(
        block(
          "MediaRecorder",
          dataTable(
            ["content type", "isTypeSupported"],
            recKeys.map(function (k) {
              return [{ v: k.slice(RECORDER_PREFIX.length), cls: "wide" }, { v: m[k].isTypeSupported }];
            })
          )
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

    function stackRow(label, s) {
      if (!s) return [label, { v: null }, { v: null }, { v: null }, { v: null }, { v: null }];
      return [
        label,
        { v: s.present },
        { v: s.name === undefined ? null : s.name },
        { v: s.frameCount === undefined ? null : s.frameCount, cls: "num" },
        { v: s.framesWellFormed === undefined ? null : s.framesWellFormed, cls: s.framesWellFormed === false ? "v-flag" : "" },
        { v: s.firstLine === undefined ? null : s.firstLine, cls: "wide" }
      ];
    }
    out.push(
      block(
        "error stack shape",
        [
          dataTable(
            ["thrown by", "stack is a string", "error name", { label: "frames", cls: "num" }, "every frame line well formed", "first line"],
            [stackRow("a property read on null", a.errorStackNative), stackRow("a throw two calls deep", a.errorStackUser)]
          ),
          kvTable([
            ["Error.prototype.stack descriptor", a.errorStackDescriptor ? JSON.stringify(a.errorStackDescriptor) : null],
            ["typeof Error.captureStackTrace", a.captureStackTraceType],
            ["typeof Error.prepareStackTrace", a.prepareStackTraceType],
            ["Error.stackTraceLimit", a.stackTraceLimit]
          ])
        ],
        "A stack is produced twice, once by the engine's own failure and once by a throw of this script's own, so the two forms can be read against each other."
      )
    );

    var hits = a.patternHits || [];
    out.push(
      block(
        "global names matching an injected-hook shape",
        hits.length
          ? dataTable(
              ["name", "shape matched"],
              hits.map(function (x) {
                return [{ v: x.name, cls: "v-flag" }, x.pattern];
              })
            )
          : h("p", { class: "absent", text: "no name on window or document matched any of the shapes probed" }),
        "The shapes probed describe how a name is formed rather than where it came from: a symbol whose suffix is a long random run, and the underscore-prefixed verb forms a host uses when it hangs a hook on a page."
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
        dataTable(
          ["permission name", "state", { label: "note", cls: "prose" }],
          states.map(function (r) {
            return [r.name, r.state, { v: r.error, cls: "prose" }];
          })
        ),
        "A name this browser does not recognise rejects the query. That is what the browser supports, and it is reported as the note on the row."
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
        "The geolocation interface is called only where the query has already reported a settled state, which is the case in which calling it answers from that state rather than raising a dialog."
      )
    );
    return out;
  }

  var GROUPS = [
    { id: "platform", title: "Platform claim", probes: ["scope.main"], render: renderPlatform, scale: null },
    {
      id: "font",
      title: "Fonts",
      probes: ["font.resolved", "font.coverage", "font.controls"],
      render: renderFont,
      scale: function () {
        var m = pmeta("font");
        return m ? m.resolvedCount + " of " + m.probed + " families resolved" : null;
      }
    },
    {
      id: "native",
      title: "Interface members",
      probes: ["native.tostring", "native.ownkeys", "native.descriptor", "native.receiver"],
      render: renderNative,
      scale: function () {
        var m = pmeta("native");
        return m && m.order ? m.order.length + " members" : null;
      }
    },
    {
      id: "scope",
      title: "Execution scopes",
      probes: ["scope.main", "scope.worker", "scope.iframe", "scope.availability"],
      render: renderScope,
      scale: function () {
        var n = 0;
        SCOPE_COLUMNS.forEach(function (c) {
          if (pv(c[0])) n += 1;
        });
        return n + " of " + SCOPE_COLUMNS.length + " scopes answered";
      }
    },
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
    {
      id: "media",
      title: "Media capabilities",
      probes: ["media.matrix"],
      render: renderMedia,
      scale: function () {
        var m = pmeta("media");
        return m && m.codecOrder ? m.codecOrder.length + " codecs" : null;
      }
    },
    { id: "auto", title: "Remote-control surface", probes: ["auto.residue"], render: renderAuto, scale: null },
    { id: "perm", title: "Permissions", probes: ["perm.state"], render: renderPerm, scale: null }
  ];

  function sectionNode(group) {
    var d = h("details", { class: "sec", id: "sec-" + group.id, open: "" });
    var scale = group.scale ? group.scale() : null;
    d.appendChild(h("summary", null, h("h2", { text: group.title }), scale ? h("span", { class: "sec-scale", text: scale }) : null));
    var body = h("div", { class: "sec-body" }, probeStrip(group.probes));
    append(body, group.render());
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
    var nodes = GROUPS.map(sectionNode);
    var idx = document.getElementById("index");
    clear(idx);
    idx.appendChild(indexNode(nodes));
    nodes.forEach(function (n) {
      host.appendChild(n);
    });
    assessmentNode();
    renderFacts();
    renderPayload();
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

  function renderPayload() {
    var host = document.getElementById("payload");
    clear(host);
    if (!state.payload) return;
    host.appendChild(
      h(
        "details",
        { class: "sec" },
        h("summary", null, h("h2", { text: "Probe payload, as sent" })),
        h("div", { class: "sec-body" }, h("pre", { class: "payload", text: JSON.stringify(requestBody(), null, 1) }))
      )
    );
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

  function setAll(open) {
    var list = document.querySelectorAll("details.sec");
    for (var i = 0; i < list.length; i++) {
      if (open) list[i].setAttribute("open", "");
      else list[i].removeAttribute("open");
    }
  }

  function init() {
    document.getElementById("run").addEventListener("click", runScan);
    document.getElementById("expand").addEventListener("click", function () {
      setAll(true);
    });
    document.getElementById("collapse").addEventListener("click", function () {
      setAll(false);
    });
    renderFacts();
    runScan();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
