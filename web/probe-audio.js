(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var reasonOf = core.reasonOf;
  var attempt = core.attempt;
  var withTimeout = core.withTimeout;

function audioOfflineCtor() {
  if (typeof OfflineAudioContext === "function") return OfflineAudioContext;
  if (typeof window !== "undefined" && typeof window.webkitOfflineAudioContext === "function") {
    return window.webkitOfflineAudioContext;
  }
  return null;
}

function audioCloseContext(ctx) {
  if (!ctx || typeof ctx.close !== "function") return;
  try {
    var p = ctx.close();
    if (p && typeof p.then === "function") p.then(null, function () {});
  } catch (e) {
  }
}

function audioBuildDeterministicGraph(ctx) {
  var osc = ctx.createOscillator();
  osc.type = "sine";
  osc.frequency.value = 440;
  osc.connect(ctx.destination);
  osc.start(0);
}

function audioChannelsServed(buffer) {
  if (!buffer || typeof buffer.getChannelData !== "function") return null;
  var n = 0;
  while (true) {
    try {
      buffer.getChannelData(n);
    } catch (e) {
      break;
    }
    n++;
    if (n > 64) break; 
  }
  return n;
}

function audioCompareChannelArrays(a, b, acc) {
  var n = Math.min(a.length, b.length);
  for (var i = 0; i < n; i++) {
    acc.sampleCount++;
    var d = Math.abs(a[i] - b[i]);
    if (d !== 0) {
      acc.differingSampleCount++;
      if (d > acc.maxAbsoluteDifference) acc.maxAbsoluteDifference = d;
    }
    if (!acc.ratios) continue;
    if (a[i] === 0) {
      if (b[i] !== 0) acc.zerosAltered++;
      continue;
    }
    acc.ratios.push(b[i] / a[i]);
  }
}

function audioScaleFit(acc) {
  if (!acc.ratios || acc.ratios.length === 0) {
    return { fitted: false, reason: "no sample with a non-zero first reading was compared" };
  }
  var sorted = acc.ratios.slice().sort(function (x, y) {
    return x - y;
  });
  var factor = sorted[Math.floor(sorted.length / 2)];
  if (!isFinite(factor) || factor === 0) {
    return { fitted: false, reason: "the ratio between the two readings did not fit one finite factor" };
  }
  var maxRelativeResidual = 0;
  for (var i = 0; i < sorted.length; i++) {
    var residual = Math.abs(sorted[i] / factor - 1);
    if (residual > maxRelativeResidual) maxRelativeResidual = residual;
  }
  return {
    fitted: true,
    factor: factor,
    comparedSamples: sorted.length,
    maxRelativeResidual: maxRelativeResidual,
    ratioMin: sorted[0],
    ratioMax: sorted[sorted.length - 1],
    zerosAltered: acc.zerosAltered
  };
}

function audioBufferDescription(buffer) {
  return {
    sampleRateHz: buffer.sampleRate,
    numberOfChannels: buffer.numberOfChannels,
    lengthFrames: buffer.length,
    durationSeconds: buffer.duration
  };
}

var AUDIO_VIEWS_SAMPLE_RATE_HZ = 8000;
var AUDIO_VIEWS_NUM_CHANNELS = 2;
var AUDIO_VIEWS_DURATION_SECONDS = 0.02; 

var AUDIO_REPEAT_SAMPLE_RATE_HZ = 8000;
var AUDIO_REPEAT_NUM_CHANNELS = 1;
var AUDIO_REPEAT_DURATION_SECONDS = 0.02;

var AUDIO_PROBES = [
  {
    id: "audio.views",
    group: "audio",
    run: function () {
      var Ctor = audioOfflineCtor();
      if (!Ctor) unsupported("OfflineAudioContext is not available");

      var sampleRateHz = AUDIO_VIEWS_SAMPLE_RATE_HZ;
      var numberOfChannels = AUDIO_VIEWS_NUM_CHANNELS;
      var lengthFrames = Math.max(1, Math.round(sampleRateHz * AUDIO_VIEWS_DURATION_SECONDS));

      var ctx;
      try {
        ctx = new Ctor(numberOfChannels, lengthFrames, sampleRateHz);
        audioBuildDeterministicGraph(ctx);
      } catch (e) {
        unsupported("could not construct or graph the offline render: " + reasonOf(e));
      }

      var rendering;
      try {
        rendering = ctx.startRendering();
      } catch (e) {
        audioCloseContext(ctx);
        unsupported("startRendering did not run: " + reasonOf(e));
      }

      return withTimeout(rendering, 4000, "audio.views render").then(
        function (buffer) {
          var served = audioChannelsServed(buffer);
          var out = {
            requested: { sampleRateHz: sampleRateHz, numberOfChannels: numberOfChannels, lengthFrames: lengthFrames },
            rendered: audioBufferDescription(buffer)
          };
          if (served !== null) out.channelsServed = served;

          var hasCopy = typeof buffer.copyFromChannel === "function";
          out.copyFromChannelAvailable = hasCopy;
          if (hasCopy && served !== null) {
            var acc = { sampleCount: 0, differingSampleCount: 0, maxAbsoluteDifference: 0, zerosAltered: 0, ratios: [] };
            var channels = Math.min(out.rendered.numberOfChannels, served);
            var ranAny = false;
            for (var c = 0; c < channels; c++) {
              try {
                var direct = buffer.getChannelData(c);
                var dest = new Float32Array(direct.length);
                buffer.copyFromChannel(dest, c, 0);
                audioCompareChannelArrays(direct, dest, acc);
                ranAny = true;
              } catch (e) {
              }
            }
            if (ranAny) {
              out.views = {
                compared: true,
                agree: acc.differingSampleCount === 0,
                sampleCount: acc.sampleCount,
                differingSampleCount: acc.differingSampleCount,
                maxAbsoluteDifference: acc.maxAbsoluteDifference,
                scale: audioScaleFit(acc)
              };
            } else {
              out.views = { compared: false };
            }
          }

          audioCloseContext(ctx);
          return out;
        },
        function (e) {
          audioCloseContext(ctx);
          throw e;
        }
      );
    }
  },
  {
    id: "audio.repeat",
    group: "audio",
    run: function () {
      var Ctor = audioOfflineCtor();
      if (!Ctor) unsupported("OfflineAudioContext is not available");

      var sampleRateHz = AUDIO_REPEAT_SAMPLE_RATE_HZ;
      var numberOfChannels = AUDIO_REPEAT_NUM_CHANNELS;
      var lengthFrames = Math.max(1, Math.round(sampleRateHz * AUDIO_REPEAT_DURATION_SECONDS));

      function renderOnce() {
        var ctx;
        try {
          ctx = new Ctor(numberOfChannels, lengthFrames, sampleRateHz);
          audioBuildDeterministicGraph(ctx);
        } catch (e) {
          unsupported("could not construct or graph the offline render: " + reasonOf(e));
        }
        var rendering;
        try {
          rendering = ctx.startRendering();
        } catch (e) {
          audioCloseContext(ctx);
          unsupported("startRendering did not run: " + reasonOf(e));
        }
        return withTimeout(rendering, 4000, "audio.repeat render").then(
          function (buffer) {
            audioCloseContext(ctx);
            return buffer;
          },
          function (e) {
            audioCloseContext(ctx);
            throw e;
          }
        );
      }

      return renderOnce().then(function (firstBuffer) {
        var out = {
          requested: { sampleRateHz: sampleRateHz, numberOfChannels: numberOfChannels, lengthFrames: lengthFrames },
          rendered: audioBufferDescription(firstBuffer),
          secondRenderCompleted: false
        };
        var firstChannel = attempt(
          function () {
            return firstBuffer.getChannelData(0);
          },
          null
        );
        if (!firstChannel) return out;

        return renderOnce().then(
          function (secondBuffer) {
            out.secondRenderCompleted = true;
            var secondChannel = attempt(
              function () {
                return secondBuffer.getChannelData(0);
              },
              null
            );
            if (!secondChannel) return out;
            var acc = { sampleCount: 0, differingSampleCount: 0, maxAbsoluteDifference: 0, zerosAltered: 0, ratios: [] };
            audioCompareChannelArrays(firstChannel, secondChannel, acc);
            out.repeat = {
              compared: true,
              agree: acc.differingSampleCount === 0,
              sampleCount: acc.sampleCount,
              differingSampleCount: acc.differingSampleCount,
              maxAbsoluteDifference: acc.maxAbsoluteDifference
            };
            return out;
          },
          function () {
            return out;
          }
        );
      });
    }
  }
];

  core.register(AUDIO_PROBES);
})();
