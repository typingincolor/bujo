// Gmail → Bujo Bookmarklet
// Readable source for the bookmarklet embedded in install.html.
// This file is not executed directly — it's minified into a javascript: URL.

void (function () {
  try {
    var s = document.querySelector("h2.hP");
    var e =
      document.querySelector("span.go span.gD[email]") ||
      document.querySelector("span.gD[email]");
    var b = document.querySelector("div.a3s.aiL");

    if (!s) {
      alert("Open an email first");
      return;
    }

    var subject = s.innerText.trim();
    var sender = e ? e.getAttribute("email") : "unknown";
    var body = b
      ? b.innerText
          .trim()
          .replace(
            /^This Message Is From an External Sender\.?\s*(This message came from outside your organization\.?\s*)?/i,
            ""
          )
          .replace(
            /^(CAUTION:?\s*)?This (email|message) (originated|came) from outside[^.]*\.\s*/i,
            ""
          )
          .replace(/^\[?EXTERNAL\]?:?\s*/i, "")
          .trim()
          .substring(0, 200)
      : "";
    var url = window.location.href;

    fetch("http://127.0.0.1:8743/api/entries", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        entries: [
          {
            type: "task",
            content:
              "Follow up: " + subject + " @" + sender.split("@")[0] + " #email",
            children: [
              { type: "note", content: "Context: " + body },
              { type: "note", content: "Email: " + url },
            ],
          },
        ],
      }),
    })
      .then(function (r) {
        return r.json();
      })
      .then(function (d) {
        if (d.success) {
          var t = document.createElement("div");
          t.style.cssText =
            "position:fixed;top:20px;right:20px;background:#1a1a2e;color:#e94560;padding:12px 20px;border-radius:8px;z-index:99999;font-family:sans-serif;font-size:14px";
          t.textContent = "Saved to Bujo";
          document.body.appendChild(t);
          setTimeout(function () {
            t.remove();
          }, 3000);
        } else {
          alert("Error: " + (d.error || "Unknown"));
        }
      })
      .catch(function () {
        alert("Bujo not running. Start the app first.");
      });
  } catch (err) {
    alert("Error: " + err.message);
  }
})();
