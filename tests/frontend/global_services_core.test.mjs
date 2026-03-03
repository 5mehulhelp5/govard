import test from "node:test";
import assert from "node:assert/strict";

import {
  normalizeGlobalServicesSnapshot,
  renderGlobalServices,
} from "../../desktop/frontend/modules/global-services.js";

test("normalizeGlobalServicesSnapshot keeps DNSMasq directly below Caddy", () => {
  const snapshot = normalizeGlobalServicesSnapshot({
    Services: [
      { ID: "mail", Name: "Mailpit" },
      { ID: "dnsmasq", Name: "DNSMasq" },
      { ID: "portainer", Name: "Portainer" },
      { ID: "caddy", Name: "Caddy Proxy" },
      { ID: "pma", Name: "PHPMyAdmin" },
    ],
  });

  const ids = snapshot.services.map((item) => item.id);
  const caddyIndex = ids.indexOf("caddy");
  const dnsmasqIndex = ids.indexOf("dnsmasq");
  assert.equal(caddyIndex >= 0, true);
  assert.equal(dnsmasqIndex, caddyIndex + 1);
});

test("renderGlobalServices shows routing warning when caddy or dnsmasq is stopped", () => {
  const container = { innerHTML: "" };
  renderGlobalServices(container, [
    {
      id: "caddy",
      name: "Caddy Proxy",
      containerName: "caddy",
      status: "stopped",
      state: "stopped",
      running: false,
      openable: true,
    },
    {
      id: "dnsmasq",
      name: "DNSMasq",
      containerName: "dnsmasq",
      status: "exited",
      state: "exited",
      running: false,
      openable: false,
    },
  ]);

  assert.equal(
    container.innerHTML.includes("Routing warning: Caddy Proxy is stopped."),
    true,
  );
  assert.equal(
    container.innerHTML.includes("Routing warning: DNSMasq is stopped."),
    true,
  );
});
