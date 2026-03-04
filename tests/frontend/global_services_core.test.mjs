import test from "node:test";
import assert from "node:assert/strict";

import {
  buildRoutingWarningMessage,
  normalizeGlobalServicesSnapshot,
  renderGlobalServices,
  summarizeActionMessage,
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

test("buildRoutingWarningMessage includes detected port conflict list without hardcoded stacks", () => {
  const message = buildRoutingWarningMessage([
    { id: "caddy", status: "created", state: "created", running: false },
    { id: "dnsmasq", status: "created", state: "created", running: false },
  ], [
    "Port conflict 80/tcp: docker container warden-nginx-1 (project: warden)",
    "Port conflict 53/udp: host process dnsmasq (pid: 845)",
  ]);

  assert.equal(message.includes("Warden"), false);
  assert.equal(message.split("\n").length, 3);
  assert.equal(
    message.includes("Missing bindings:"),
    true,
  );
  assert.equal(
    message.includes("Occupied by:"),
    true,
  );
  assert.equal(
    message.includes("warden-nginx-1 (80/tcp)"),
    true,
  );
  assert.equal(
    message.includes("dnsmasq (53/udp)"),
    true,
  );
  assert.equal(
    message.includes("Resolve conflicts, then click Restart All or Start All."),
    true,
  );
});

test("buildRoutingWarningMessage warns when services look running but bindings are degraded", () => {
  const message = buildRoutingWarningMessage([
    { id: "caddy", status: "running", state: "running", running: true },
    { id: "dnsmasq", status: "running", state: "running", running: true },
  ], [
    "Port conflict 80/tcp: Caddy Proxy is running but govard-proxy-caddy is not published on host",
  ]);

  assert.equal(
    message.includes("Routing services are running but port bindings are degraded."),
    true,
  );
  assert.equal(message.split("\n").length, 3);
  assert.equal(
    message.includes("Missing bindings: Caddy Proxy (80/tcp)."),
    true,
  );
  assert.equal(
    message.includes("is running but govard-proxy-caddy is not published on host"),
    false,
  );
});

test("summarizeActionMessage keeps only the first line from multiline command output", () => {
  const message = summarizeActionMessage(
    "Global services restarted.\nContainer govard-proxy-caddy Started\nContainer govard-proxy-pma Started",
    "fallback",
  );

  assert.equal(message, "Global services restarted.");
});

test("summarizeActionMessage uses fallback when message is empty", () => {
  const message = summarizeActionMessage("   \n\t", "Global restart completed.");

  assert.equal(message, "Global restart completed.");
});
