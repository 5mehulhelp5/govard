import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

import {
  normalizeRemotePreset,
  normalizeRemotesPayload,
  renderRemotes,
} from "../../desktop/frontend/modules/remotes.js";

const getPresetButtonTag = (html, preset) => {
  const match = html.match(
    new RegExp(`<button[^>]*data-preset="${preset}"[^>]*>`, "i"),
  );
  assert.ok(match, `missing ${preset} preset button`);
  return match[0];
};

test("normalizeRemotePreset canonicalizes aliases", () => {
  assert.equal(normalizeRemotePreset("file"), "files");
  assert.equal(normalizeRemotePreset("database"), "db");
  assert.equal(normalizeRemotePreset("full"), "full");
  assert.equal(normalizeRemotePreset("unknown"), "");
});

test("normalizeRemotesPayload maps mixed-case payload fields", () => {
  const payload = normalizeRemotesPayload({
    Project: "demo",
    Remotes: [
      {
        Name: "staging",
        Host: "staging.example.com",
        User: "deploy",
        Path: "/var/www/staging",
        Port: 22,
        Environment: "staging",
        Protected: false,
        AuthMethod: "keychain",
        LastSync: "2m ago",
        Capabilities: ["files", "media"],
      },
    ],
    Warnings: ["warn"],
  });

  assert.equal(payload.project, "demo");
  assert.equal(payload.remotes.length, 1);
  assert.equal(payload.remotes[0].name, "staging");
  assert.equal(payload.remotes[0].authMethod, "keychain");
  assert.equal(payload.remotes[0].lastSync, "2m ago");
  assert.deepEqual(payload.remotes[0].capabilities, ["files", "media"]);
  assert.deepEqual(payload.warnings, ["warn"]);
});

test("desktop layout exposes remotes section and actions", async () => {
  const html = await readFile(
    new URL("../../desktop/frontend/index.html", import.meta.url),
    "utf8",
  );
  assert.equal(html.includes('id="remotes"'), true, "missing remotes panel");
  assert.equal(
    html.includes('data-action="refresh-remotes"'),
    true,
    "missing remotes refresh action",
  );
});

test("remote card renders open-url button to the left of test connection", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "staging",
      host: "staging.example.com",
      environment: "staging",
      protected: false,
    },
  ]);

  const openIndex = container.innerHTML.indexOf('data-action="open-remote-url"');
  const testIndex = container.innerHTML.indexOf('data-action="remote-test"');

  assert.equal(openIndex >= 0, true, "missing open-remote-url action button");
  assert.equal(testIndex >= 0, true, "missing remote-test action button");
  assert.equal(
    openIndex < testIndex,
    true,
    "open-remote-url button should be rendered to the left of remote-test",
  );
  assert.equal(
    container.innerHTML.includes("CONNECTED"),
    false,
    "connected badge should no longer be rendered",
  );
  assert.equal(
    container.innerHTML.includes('data-action="open-remote-shell"'),
    true,
    "missing open remote shell action button",
  );
  assert.equal(
    container.innerHTML.includes('data-action="open-remote-db"'),
    true,
    "missing open remote db action button",
  );
  assert.equal(
    container.innerHTML.includes('data-action="open-remote-sftp"'),
    true,
    "missing open remote sftp action button",
  );
});

test("remote open action buttons include loading labels for pending UX", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "staging",
      host: "staging.example.com",
      environment: "staging",
      protected: false,
    },
  ]);

  assert.equal(
    container.innerHTML.includes('data-loading-label="Opening SSH..."'),
    true,
    "open ssh button should define loading label",
  );
  assert.equal(
    container.innerHTML.includes('data-loading-label="Opening Database..."'),
    true,
    "open database button should define loading label",
  );
  assert.equal(
    container.innerHTML.includes('data-loading-label="Opening SFTP..."'),
    true,
    "open sftp button should define loading label",
  );
  assert.equal(
    container.innerHTML.includes('data-role="label"'),
    true,
    "open action buttons should include label span for loading updates",
  );
});

test("remote card shows auth method summary", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "staging",
      host: "staging.example.com",
      authMethod: "ssh-agent",
      protected: false,
    },
  ]);

  assert.equal(
    container.innerHTML.includes("Auth: SSH Agent"),
    true,
    "expected auth summary to show normalized auth method",
  );
});

test("pull buttons are disabled when remote capability is missing", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "limited",
      host: "limited.example.com",
      environment: "staging",
      capabilities: ["files"],
    },
  ]);

  const dbButton = getPresetButtonTag(container.innerHTML, "db");
  const mediaButton = getPresetButtonTag(container.innerHTML, "media");

  assert.equal(
    dbButton.includes("disabled"),
    true,
    "db button should be disabled when db capability is missing",
  );
  assert.equal(
    mediaButton.includes("disabled"),
    true,
    "media button should be disabled when media capability is missing",
  );
});

test("pull buttons stay enabled when capability list is not declared", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "legacy",
      host: "legacy.example.com",
      environment: "staging",
    },
  ]);

  const dbButton = getPresetButtonTag(container.innerHTML, "db");
  const mediaButton = getPresetButtonTag(container.innerHTML, "media");

  assert.equal(
    dbButton.includes("disabled"),
    false,
    "db button should remain enabled when capabilities are absent",
  );
  assert.equal(
    mediaButton.includes("disabled"),
    false,
    "media button should remain enabled when capabilities are absent",
  );
});

test("protected warning copy does not hardcode production wording", () => {
  const container = { innerHTML: "" };
  renderRemotes(container, [
    {
      name: "staging-protected",
      host: "staging.example.com",
      environment: "staging",
      protected: true,
    },
  ]);

  assert.equal(
    container.innerHTML.includes("protected remote can overwrite local data"),
    true,
    "expected protected warning copy",
  );
  assert.equal(
    container.innerHTML.includes("from Production can overwrite"),
    false,
    "warning copy should not hardcode production wording",
  );
});
