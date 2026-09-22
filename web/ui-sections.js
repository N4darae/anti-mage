var AMSections = (function () {
  "use strict";

  function renderPlatform(p) {
    var m = p.value("scope.main");
    if (!m) return [p.absentBlock("Platform claim", "scope.main")];
    return [
      AMDom.block(
        "what this browser says it is",
        AMDom.kvTable([
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
        ])
      )
    ];
  }

  function screenBlock(g) {
    return AMDom.block(
      "screen and window, from JavaScript",
      AMDom.kvTable([
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
    );
  }

  function cssBlock(c) {
    return AMDom.block(
      "the same numbers, as CSS reports them",
      AMDom.kvTable([
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
      ])
    );
  }

  function agreementBlock(g, c) {
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
    return AMDom.block(
      "read two ways",
      AMDom.dataTable(
        ["quantity", { label: "JavaScript", cls: "num" }, { label: "CSS media query", cls: "num" }, "agree"],
        pairs.map(function (p) {
          return { cells: [p[0], { v: p[1], cls: "num" }, { v: p[2], cls: "num" }, { v: p[3], cls: p[3] === false ? "v-flag" : "" }], flag: p[3] === false };
        })
      )
    );
  }

  function renderGeom(p) {
    var out = [];
    var g = p.value("geom.screen");
    var c = p.value("geom.css");
    if (!g) out.push(p.absentBlock("Screen and window", "geom.screen"));
    else out.push(screenBlock(g));
    if (!c) out.push(p.absentBlock("The same numbers, as CSS reports them", "geom.css"));
    else out.push(cssBlock(c));
    if (g && c) out.push(agreementBlock(g, c));
    return out;
  }

  function renderTime(p) {
    var out = [];
    var z = p.value("time.zone");
    var o = p.value("time.offsets");
    var ometa = p.meta("timeOffsets");
    if (!z) out.push(p.absentBlock("Zone and locale", "time.zone"));
    else {
      out.push(
        AMDom.block(
          "zone and locale",
          AMDom.kvTable([
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
    if (!o || !o.length) out.push(p.absentBlock("Offsets across dates", "time.offsets"));
    else {
      out.push(
        AMDom.block(
          "offsets across dates",
          AMDom.kvTable([
            ["dates supplied by", ometa ? ometa.datesFrom : null],
            ["zone they were read in", ometa ? ometa.timeZone : null],
            ["dates sampled", o.length],
            ["dates that did not parse", ometa ? ometa.unparsedCount : null],
            ["distinct offsets seen, in minutes", ometa ? ometa.distinctOffsets : null]
          ])
        )
      );
    }
    return out;
  }

  function stackShape(s) {
    if (!s) return null;
    return [s.present ? "a string" : "not a string", s.name || "no name", (s.frameCount === undefined ? "?" : s.frameCount) + " frames", s.framesWellFormed === false ? "a frame line is malformed" : "every frame line well formed"].join(", ");
  }

  function renderAuto(p) {
    var a = p.value("auto.residue");
    if (!a) return [p.absentBlock("Remote-control surface", "auto.residue")];
    var out = [];
    var wd = a.webdriverDescriptor || {};
    out.push(
      AMDom.block(
        "navigator.webdriver",
        AMDom.kvTable([
          ["value", a.webdriver === undefined ? null : a.webdriver],
          ["property present on navigator", a.navigatorWebdriverPresent],
          ["descriptor on Navigator.prototype", wd.onPrototype === undefined ? null : wd.onPrototype],
          ["descriptor kind", wd.kind || null],
          ["configurable", wd.configurable === undefined ? null : wd.configurable],
          ["getter ends in [native code]", wd.getterNative === undefined ? null : wd.getterNative, wd.getterNative === false ? "v-flag" : ""]
        ])
      )
    );

    var hits = a.patternHits || [];
    out.push(
      AMDom.block(
        "error stack shape",
        AMDom.kvTable([
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
        ])
      )
    );

    var prefixed = (a.windowPrefixedKeys || []).concat(a.documentPrefixedKeys || []);
    out.push(
      AMDom.block(
        "counts and other globals",
        AMDom.kvTable([
          ["own property names on window", a.windowKeyCount],
          ["own property names on document", a.documentKeyCount],
          ["globals whose name starts with _ or $", prefixed.length ? prefixed : null]
        ])
      )
    );
    return out;
  }

  function renderPerm(p) {
    var perm = p.value("perm.state");
    if (!perm) return [p.absentBlock("Permissions", "perm.state")];
    var out = [];
    var states = perm.states || [];
    out.push(
      AMDom.block(
        "Permissions.query",
        AMDom.kvTable([
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
        ])
      )
    );
    var n = perm.notification || {};
    var g = perm.geolocation || {};
    out.push(
      AMDom.block(
        "the query against the interface itself",
        AMDom.kvTable([
          ["Notification.permission", n.apiValue],
          ["Permissions.query('notifications')", n.queryState],
          ["the two agree", n.agree, n.agree === false ? "v-flag" : ""],
          ["Permissions.query('geolocation')", g.queryState],
          ["Geolocation interface present", g.apiPresent],
          ["interface called", g.exercised],
          ["outcome", g.outcome],
          ["outcome matches the query", g.agree, g.agree === false ? "v-flag" : ""]
        ])
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
      scale: function (p) {
        var o = p.value("time.offsets");
        return o && o.length ? o.length + " dates sampled" : null;
      }
    },
    { id: "auto", title: "Remote-control surface", probes: ["auto.residue"], render: renderAuto, scale: null },
    { id: "perm", title: "Permissions", probes: ["perm.state"], render: renderPerm, scale: null }
  ];

  function sectionNode(group, p) {
    var content = group.render(p);
    if (content === null && !(group.probes || []).length) return null;
    var d = AMDom.h("details", { class: "sec", id: "sec-" + group.id });
    var scale = group.scale ? group.scale(p) : null;
    d.appendChild(
      AMDom.h("summary", null, AMDom.h("h2", { text: group.title }), scale ? AMDom.h("span", { class: "sec-scale", text: scale }) : null)
    );
    var body = AMDom.h("div", { class: "sec-body" }, p.strip(group.probes));
    AMDom.append(body, content);
    d.appendChild(body);
    return d;
  }

  function build(payload) {
    var p = AMProbes.reader(payload);
    return GROUPS.map(function (group) {
      return sectionNode(group, p);
    }).filter(function (n) {
      return n !== null;
    });
  }

  function index(sections) {
    var ol = AMDom.h("ol");
    sections.forEach(function (sec) {
      var title = sec.querySelector("summary h2");
      var scale = sec.querySelector("summary .sec-scale");
      ol.appendChild(
        AMDom.h(
          "li",
          null,
          AMDom.h("a", { href: "#" + sec.id, text: title ? title.textContent : sec.id }),
          scale ? AMDom.h("span", { class: "idx-count", text: scale.textContent }) : null
        )
      );
    });
    return AMDom.h("nav", { class: "index" }, AMDom.h("h2", { text: "what was observed" }), ol);
  }

  return { build: build, index: index };
})();
