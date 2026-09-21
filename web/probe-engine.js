(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;

  core.register([
    {
      id: "engine.features",
      group: "engine",
      run: function () {
        var checks = {
          textFit: function () {
            return CSS.supports("text-fit", "shrink");
          },
          bgClipBorderArea: function () {
            return CSS.supports("background-clip", "border-area");
          },
          rubyOverhang: function () {
            return CSS.supports("ruby-overhang", "auto");
          },
          animationEventAnimation: function () {
            return "animation" in AnimationEvent.prototype;
          },
          transitionEventAnimation: function () {
            return "animation" in TransitionEvent.prototype;
          },
          userMediaElement: function () {
            return "HTMLUserMediaElement" in window;
          },
          softNavigations: function () {
            return PerformanceObserver.supportedEntryTypes.indexOf("soft-navigation") >= 0;
          }
        };
        var out = {};
        var names = Object.keys(checks);
        for (var i = 0; i < names.length; i++) {
          var id = names[i];
          try {
            out[id] = Boolean(checks[id]());
          } catch (e) {}
        }
        if (!Object.keys(out).length) unsupported("no capability could be evaluated in this context");
        return out;
      }
    }
  ]);
})();
