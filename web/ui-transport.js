var AMTransport = (function () {
  "use strict";

  var ENDPOINT = "./api/scan";

  function requestBody(payload) {
    var body = { v: 1, probes: payload.probes };
    if (payload.nonce) body.nonce = payload.nonce;
    return body;
  }

  function assess(payload) {
    return fetch(ENDPOINT, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
      credentials: "omit",
      body: JSON.stringify(requestBody(payload))
    }).then(function (r) {
      if (!r.ok) throw new Error("the local server answered HTTP " + r.status);
      return r.json();
    });
  }

  return { assess: assess };
})();
