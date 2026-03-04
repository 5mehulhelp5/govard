import test from "node:test";
import assert from "node:assert/strict";

import {
  normalizeSettingsPayload,
  renderSettingsDrawer,
} from "../../desktop/frontend/modules/settings.js";

test("normalizeSettingsPayload maps settings payload", () => {
  const value = normalizeSettingsPayload({
    theme: "dark",
    proxyTarget: "govard.test",
    preferredBrowser: "firefox",
  });
  assert.deepEqual(value, {
    theme: "dark",
    proxyTarget: "govard.test",
    preferredBrowser: "firefox",
    codeEditor: "",
    dbClientPreference: "pma",
  });
});

test("normalizeSettingsPayload falls back to defaults", () => {
  const value = normalizeSettingsPayload({});
  assert.deepEqual(value, {
    theme: "system",
    proxyTarget: "",
    preferredBrowser: "",
    codeEditor: "",
    dbClientPreference: "pma",
  });
});

test("renderSettingsDrawer includes update controls", () => {
  const container = { innerHTML: "" };
  renderSettingsDrawer(container);

  assert.equal(
    container.innerHTML.includes('data-action="check-updates"'),
    true,
    "expected check-updates action in settings drawer",
  );
  assert.equal(
    container.innerHTML.includes('data-action="install-update"'),
    true,
    "expected install-update action in settings drawer",
  );
});
