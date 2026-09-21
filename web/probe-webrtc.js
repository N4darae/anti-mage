(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;
  var withTimeout = core.withTimeout;

  var ICE_DEADLINE_MS = 5000;
  var ICE_POLL_MS = 100;

  function sdpCandidateLines(pc) {
    var sdp = attempt(function () {
      return pc.localDescription ? pc.localDescription.sdp : null;
    }, null);
    if (typeof sdp !== "string") return -1;
    return (sdp.match(/^a=candidate:/gm) || []).length;
  }

  function countLocalCandidates(pc) {
    var found = sdpCandidateLines(pc);
    var lines = found < 0 ? null : found;
    var stats = attempt(function () {
      return pc.getStats();
    }, null);
    if (!stats || typeof stats.then !== "function") {
      return Promise.resolve({ sdpCandidateLines: lines, statsLocalCandidates: null, statsRead: false });
    }
    return withTimeout(stats, 2000, "getStats").then(
      function (report) {
        var n = 0;
        attempt(function () {
          report.forEach(function (entry) {
            if (entry && entry.type === "local-candidate") n += 1;
          });
          return true;
        }, null);
        return { sdpCandidateLines: lines, statsLocalCandidates: n, statsRead: true };
      },
      function () {
        return { sdpCandidateLines: lines, statsLocalCandidates: null, statsRead: false };
      }
    );
  }

  core.register([
    {
      id: "webrtc.ice",
      group: "webrtc",
      prime: function () {
        var Ctor = window.RTCPeerConnection || window.webkitRTCPeerConnection;
        if (typeof Ctor !== "function") unsupported("no peer connection interface is available");
        var pc;
        try {
          pc = new Ctor({ iceServers: [] });
        } catch (e) {
          unsupported("a peer connection could not be created: " + reasonOf(e));
        }
        var sawStates = [];
        var candidateTypes = [];
        var gatheringEvents = 0;
        var candidateEvents = 0;
        function noteState() {
          var st = attempt(function () {
            return pc.iceGatheringState;
          }, null);
          if (st && sawStates[sawStates.length - 1] !== st) sawStates.push(st);
        }
        noteState();
        var settle;
        var finished = new Promise(function (resolve) {
          settle = resolve;
        });
        pc.onicegatheringstatechange = function () {
          gatheringEvents += 1;
          noteState();
          if (attempt(function () {
            return pc.iceGatheringState;
          }, null) === "complete") settle(true);
        };
        pc.onicecandidate = function (ev) {
          candidateEvents += 1;
          if (!ev.candidate) {
            noteState();
            settle(true);
            return;
          }
          var t = attempt(function () {
            return ev.candidate.type;
          }, null);
          candidateTypes.push(typeof t === "string" ? t : "unknown");
        };
        try {
          pc.createDataChannel("am-probe");
        } catch (e) {}
        var localDescriptionSet = false;
        var describeError = null;
        var offered = Promise.resolve()
          .then(function () {
            return pc.createOffer();
          })
          .then(function (offer) {
            return pc.setLocalDescription(offer);
          })
          .then(function () {
            localDescriptionSet = true;
          })
          .catch(function (e) {
            describeError = reasonOf(e);
          });
        var settled = false;
        finished.then(function () {
          settled = true;
        });
        var poll = setInterval(function () {
          if (settled) {
            clearInterval(poll);
            return;
          }
          if (attempt(function () {
            return pc.iceGatheringState;
          }, null) !== "new") return;
          if (sdpCandidateLines(pc) > 0) {
            clearInterval(poll);
            settle(true);
          }
        }, ICE_POLL_MS);

        return offered
          .then(function () {
            return withTimeout(finished, ICE_DEADLINE_MS, "ice gathering").then(
              function () {
                return false;
              },
              function () {
                return true;
              }
            );
          })
          .then(function (timedOut) {
            clearInterval(poll);
            noteState();
            var signalingState = attempt(function () {
              return pc.signalingState;
            }, null);
            return countLocalCandidates(pc).then(function (own) {
              try {
                pc.close();
              } catch (e) {}
              return {
                finalState: sawStates.length ? sawStates[sawStates.length - 1] : null,
                timedOut: timedOut,
                sawStates: sawStates,
                candidateTypes: candidateTypes,
                localDescriptionSet: localDescriptionSet,
                describeError: describeError,
                signalingState: signalingState,
                gatheringEvents: gatheringEvents,
                candidateEvents: candidateEvents,
                sdpCandidateLines: own.sdpCandidateLines,
                statsLocalCandidates: own.statsLocalCandidates,
                statsRead: own.statsRead
              };
            });
          });
      }
    }
  ]);
})();
