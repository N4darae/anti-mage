(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var attempt = core.attempt;

  function longOffsetFor(tz, d) {
    if (!tz) return null;
    return attempt(function () {
      var parts = new Intl.DateTimeFormat("en-US", { timeZone: tz, timeZoneName: "longOffset" }).formatToParts(d);
      for (var k = 0; k < parts.length; k++) {
        if (parts[k].type === "timeZoneName") return parts[k].value;
      }
      return null;
    }, null);
  }

  core.register([
    {
      id: "time.zone",
      group: "time",
      run: function () {
        if (!window.Intl || !Intl.DateTimeFormat) unsupported("Intl.DateTimeFormat is not available");
        var ro = Intl.DateTimeFormat().resolvedOptions();
        var now = new Date();
        function namePart(style) {
          return attempt(function () {
            var parts = new Intl.DateTimeFormat("en-US", { timeZone: ro.timeZone, timeZoneName: style }).formatToParts(now);
            for (var i = 0; i < parts.length; i++) {
              if (parts[i].type === "timeZoneName") return parts[i].value;
            }
            return null;
          }, null);
        }
        return {
          timeZone: ro.timeZone || null,
          locale: ro.locale || null,
          calendar: ro.calendar || null,
          numberingSystem: ro.numberingSystem || null,
          hourCycle: ro.hourCycle || null,
          longOffset: namePart("longOffset"),
          longName: namePart("long"),
          shortName: namePart("short"),
          currentOffsetMinutes: now.getTimezoneOffset(),
          dateToString: String(now),
          dateToLocaleString: attempt(function () {
            return now.toLocaleString();
          }, null),
          navigatorLanguage: attempt(function () {
            return navigator.language;
          }, null),
          navigatorLanguages: attempt(function () {
            return navigator.languages ? Array.prototype.slice.call(navigator.languages) : null;
          }, null)
        };
      }
    },
    {
      id: "time.offsets",
      group: "time",
      run: function () {
        var dates = core.bootstrap && core.bootstrap.dates;
        if (!dates || !dates.length) {
          unsupported("the page bootstrap supplied no sample date");
        }
        var tz = attempt(function () {
          return Intl.DateTimeFormat().resolvedOptions().timeZone;
        }, null);
        var rows = [];
        var skipped = 0;
        for (var i = 0; i < dates.length; i++) {
          var sample = dates[i];
          var raw = sample.date === null || sample.date === undefined ? "" : String(sample.date);
          var d;
          if (sample.epochMs !== null && sample.epochMs !== undefined) {
            d = new Date(sample.epochMs);
          } else {
            d = new Date(/^\d{4}-\d{2}-\d{2}$/.test(raw) ? raw + "T12:00:00Z" : raw);
          }
          if (isNaN(d.getTime())) {
            skipped++;
            continue;
          }
          rows.push({
            date: raw || d.toISOString().slice(0, 10),
            instant: sample.iso || d.toISOString(),
            epochMs: d.getTime(),
            offsetMinutes: d.getTimezoneOffset(),
            longOffset: longOffsetFor(tz, d),
            localString: attempt(function () {
              return String(d);
            }, null)
          });
        }
        if (!rows.length) unsupported("no supplied date parsed as a date");
        var distinct = {};
        rows.forEach(function (r) {
          distinct[r.offsetMinutes] = true;
        });
        core.meta.timeOffsets = {
          datesFrom: (core.bootstrap && core.bootstrap.source) || "unknown",
          timeZone: tz,
          sampleCount: rows.length,
          unparsedCount: skipped,
          distinctOffsets: Object.keys(distinct)
            .map(Number)
            .sort(function (a, b) {
              return a - b;
            })
        };
        return rows;
      }
    }
  ]);
})();
