(function () {
  "use strict";

  var core = AM_CORE;
  var unsupported = core.unsupported;
  var attempt = core.attempt;
  var withTimeout = core.withTimeout;

  core.register([
    {
      id: "canvas.serial",
      group: "canvas",
      run: function () {
        var W = 220, H = 90;
        var canvas = attempt(function () {
          return document.createElement("canvas");
        }, null);
        if (!canvas || typeof canvas.getContext !== "function") unsupported("this document cannot create a canvas element");
        canvas.width = W;
        canvas.height = H;
        var ctx = attempt(function () {
          return canvas.getContext("2d");
        }, null);
        if (!ctx) unsupported("no two-dimensional drawing context was granted");

        var grad = ctx.createLinearGradient(0, 0, W, H);
        grad.addColorStop(0, "#1b3a6b");
        grad.addColorStop(0.5, "#c8452a");
        grad.addColorStop(1, "#e9e4d0");
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, W, H);
        ctx.font = "24px serif";
        ctx.fillStyle = "#ffffff";
        ctx.fillText("anti-mage 0123", 8, 40);
        ctx.strokeStyle = "rgba(0,0,0,0.55)";
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.arc(W - 30, H - 30, 20, 0, Math.PI * 1.5);
        ctx.stroke();

        function fnv1a(bytes) {
          var h = 0x811c9dc5;
          for (var i = 0; i < bytes.length; i++) {
            h = h ^ bytes[i];
            h = (h * 0x01000193) >>> 0;
          }
          var hex = h.toString(16);
          while (hex.length < 8) hex = "0" + hex;
          return hex;
        }
        function hashOf(buffer) {
          return fnv1a(new Uint8Array(buffer));
        }
        function pixelsOf(c) {
          return c.getContext("2d").getImageData(0, 0, W, H).data;
        }
        function decodeHash(src, revoke) {
          return new Promise(function (resolve) {
            var img = new Image();
            img.onload = function () {
              var out = document.createElement("canvas");
              out.width = W;
              out.height = H;
              try {
                out.getContext("2d").drawImage(img, 0, 0);
                resolve(fnv1a(pixelsOf(out)));
              } catch (e) {
                resolve(null);
              }
              if (revoke) attempt(function () {
                return URL.revokeObjectURL(src);
              }, null);
            };
            img.onerror = function () {
              resolve(null);
              if (revoke) attempt(function () {
                return URL.revokeObjectURL(src);
              }, null);
            };
            img.src = src;
          });
        }

        var out = { dataUrlAvailable: false, blobAvailable: false, rawHash: fnv1a(pixelsOf(canvas)) };

        var dataUrl = attempt(function () {
          return canvas.toDataURL("image/png");
        }, null);
        var dataStep = Promise.resolve();
        if (typeof dataUrl === "string" && dataUrl.indexOf("data:image/png") === 0) {
          out.dataUrlAvailable = true;
          dataStep = fetch(dataUrl)
            .then(function (r) {
              return r.arrayBuffer();
            })
            .then(function (buf) {
              out.dataUrlHash = hashOf(buf);
              out.dataUrlLength = buf.byteLength;
              return decodeHash(dataUrl, false);
            })
            .then(function (h) {
              out.dataUrlPixelsMatchRaw = h !== null && h === out.rawHash;
            })
            .catch(function (e) {
              out.dataUrlAvailable = false;
            });
        }

        var blobStep = dataStep.then(function () {
          if (typeof canvas.toBlob !== "function") return null;
          return withTimeout(
            new Promise(function (resolve) {
              canvas.toBlob(resolve, "image/png");
            }),
            5000,
            "canvas.toBlob"
          ).catch(function (e) {
            return null;
          });
        });

        return blobStep
          .then(function (blob) {
            if (!blob || typeof blob.arrayBuffer !== "function") return null;
            out.blobAvailable = true;
            return blob.arrayBuffer().then(function (buf) {
              out.blobHash = hashOf(buf);
              out.blobLength = buf.byteLength;
              return decodeHash(URL.createObjectURL(blob), true);
            });
          })
          .then(function (h) {
            if (out.blobAvailable) out.blobPixelsMatchRaw = h !== null && h === out.rawHash;
            return out;
          })
          .catch(function (e) {
            return out;
          });
      }
    }
  ]);
})();
