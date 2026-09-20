(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;

  core.register([
    {
      id: "perm.state",
      group: "perm",
      run: function () {
        if (!navigator.permissions || typeof navigator.permissions.query !== "function") {
          unsupported("navigator.permissions.query is not available");
        }
        var names = ["geolocation", "notifications", "camera", "microphone", "clipboard-read", "midi", "persistent-storage"];
        var rows = [];
        var chain = Promise.resolve();
        names.forEach(function (n) {
          chain = chain.then(function () {
            return navigator.permissions
              .query({ name: n })
              .then(function (p) {
                rows.push({ name: n, state: p.state, error: null });
              })
              .catch(function (e) {
                rows.push({ name: n, state: null, error: reasonOf(e) });
              });
          });
        });
        return chain.then(function () {
          var byName = {};
          rows.forEach(function (r) {
            byName[r.name] = r.state;
          });
          var notif = {
            apiPresent: typeof Notification !== "undefined",
            apiValue:
              typeof Notification !== "undefined"
                ? attempt(function () {
                    return Notification.permission;
                  }, null)
                : null,
            queryState: byName.notifications || null
          };
          var equivalent = { granted: "granted", denied: "denied", prompt: "default" };
          notif.agree =
            notif.apiValue === null ||
            notif.queryState === null ||
            !Object.prototype.hasOwnProperty.call(equivalent, notif.queryState)
              ? null
              : equivalent[notif.queryState] === notif.apiValue;
          var pair = {};
          if (typeof notif.queryState === "string" && notif.queryState) pair.query = notif.queryState;
          if (typeof notif.apiValue === "string" && notif.apiValue) pair.actual = notif.apiValue;

          var geoState = byName.geolocation || null;
          var geo = { queryState: geoState, apiPresent: !!navigator.geolocation, exercised: false, outcome: null, agree: null };
          if (!geo.apiPresent || geoState !== "denied") {
            geo.outcome = "not exercised";
            return { states: rows, notifications: pair, notification: notif, geolocation: geo };
          }
          geo.exercised = true;
          return new Promise(function (resolve) {
            var settled = false;
            function finish() {
              if (settled) return;
              settled = true;
              resolve({ states: rows, notifications: pair, notification: notif, geolocation: geo });
            }
            var t = setTimeout(function () {
              geo.outcome = "no answer within 3000 ms";
              finish();
            }, 3000);
            navigator.geolocation.getCurrentPosition(
              function () {
                clearTimeout(t);
                geo.outcome = "position returned";
                geo.agree = false;
                finish();
              },
              function (err) {
                clearTimeout(t);
                geo.outcome = "error code " + err.code;
                geo.agree = err.code === 1;
                finish();
              },
              { timeout: 2500, maximumAge: 0 }
            );
          });
        });
      }
    }
  ]);
})();
