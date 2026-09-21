(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;

  var CODECS = [
    ["H.264 baseline, MP4", 'video/mp4; codecs="avc1.42E01E"', "video"],
    ["H.264 high, MP4", 'video/mp4; codecs="avc1.640028"', "video"],
    ["HEVC, MP4", 'video/mp4; codecs="hev1.1.6.L93.B0"', "video"],
    ["VP8, WebM", 'video/webm; codecs="vp8"', "video"],
    ["VP9, WebM", 'video/webm; codecs="vp9"', "video"],
    ["AV1, WebM", 'video/webm; codecs="av01.0.04M.08"', "video"],
    ["Theora, Ogg", 'video/ogg; codecs="theora"', "video"],
    ["AAC-LC, MP4", 'audio/mp4; codecs="mp4a.40.2"', "audio"],
    ["MP3", "audio/mpeg", "audio"],
    ["Opus, WebM", 'audio/webm; codecs="opus"', "audio"],
    ["Vorbis, Ogg", 'audio/ogg; codecs="vorbis"', "audio"],
    ["FLAC", "audio/flac", "audio"]
  ];

  var RECORDER_PREFIX = "MediaRecorder ";

  var RECORDER_TYPES = [
    "video/webm",
    "video/webm;codecs=vp8",
    "video/webm;codecs=vp9",
    "video/webm;codecs=h264",
    "video/mp4",
    "audio/webm;codecs=opus",
    "audio/ogg;codecs=opus"
  ];

  core.register([
    {
      id: "media.matrix",
      group: "media",
      run: function () {
        var video = attempt(function () {
          return document.createElement("video");
        }, null);
        var audio = attempt(function () {
          return document.createElement("audio");
        }, null);
        var canPlay = !!(video && typeof video.canPlayType === "function");
        var hasMSE = typeof MediaSource !== "undefined" && typeof MediaSource.isTypeSupported === "function";
        var hasRecorder = typeof MediaRecorder !== "undefined" && typeof MediaRecorder.isTypeSupported === "function";
        var hasCaps = !!(navigator.mediaCapabilities && typeof navigator.mediaCapabilities.decodingInfo === "function");
        if (!canPlay && !hasMSE && !hasRecorder && !hasCaps) {
          unsupported("no media capability interface is available");
        }
        var rows = CODECS.map(function (c) {
          var el = c[2] === "video" ? video : audio;
          return {
            label: c[0],
            contentType: c[1],
            kind: c[2],
            canPlayType:
              canPlay && el
                ? attempt(function () {
                    return el.canPlayType(c[1]);
                  }, null)
                : null,
            mediaSource: hasMSE
              ? attempt(function () {
                  return MediaSource.isTypeSupported(c[1]);
                }, null)
              : null,
            decodingInfo: null
          };
        });
        var recorder = RECORDER_TYPES.map(function (t) {
          return {
            contentType: t,
            supported: hasRecorder
              ? attempt(function () {
                  return MediaRecorder.isTypeSupported(t);
                }, null)
              : null
          };
        });
        var facilities = { canPlayType: canPlay, mediaSource: hasMSE, mediaRecorder: hasRecorder, mediaCapabilities: hasCaps };
        core.meta.media = {
          facilities: facilities,
          codecOrder: rows.map(function (r) {
            return r.label;
          }),
          recorderOrder: recorder.map(function (r) {
            return RECORDER_PREFIX + r.contentType;
          })
        };
        function flatten() {
          var out = { "interfaces available": facilities };
          rows.forEach(function (r) {
            var rec = { contentType: r.contentType, kind: r.kind };
            if (r.canPlayType !== null) rec.canPlayType = r.canPlayType;
            if (r.mediaSource !== null) rec.mediaSource = r.mediaSource;
            var di = r.decodingInfo;
            if (di) {
              if (di.error) rec.decodingInfoError = di.error;
              else {
                rec.decodingInfoSupported = di.supported;
                rec.smooth = di.smooth;
                rec.powerEfficient = di.powerEfficient;
              }
            }
            out[r.label] = rec;
          });
          recorder.forEach(function (r) {
            if (r.supported === null) return;
            out[RECORDER_PREFIX + r.contentType] = { isTypeSupported: r.supported };
          });
          return out;
        }
        if (!hasCaps) return flatten();
        var chain = Promise.resolve();
        rows.forEach(function (r) {
          chain = chain.then(function () {
            var config =
              r.kind === "video"
                ? { type: "file", video: { contentType: r.contentType, width: 1280, height: 720, bitrate: 2000000, framerate: 30 } }
                : { type: "file", audio: { contentType: r.contentType, channels: 2, bitrate: 128000, samplerate: 48000 } };
            return navigator.mediaCapabilities
              .decodingInfo(config)
              .then(function (info) {
                r.decodingInfo = { supported: !!info.supported, smooth: !!info.smooth, powerEfficient: !!info.powerEfficient };
              })
              .catch(function (e) {
                r.decodingInfo = { error: reasonOf(e) };
              });
          });
        });
        return chain.then(function () {
          return flatten();
        });
      }
    }
  ]);
})();
