import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

import { createUpdateNotifierController } from "../../desktop/frontend/modules/update-notifier.js";

const createClassList = (initial = []) => {
  const values = new Set(initial);
  return {
    add: (...classes) => {
      classes.forEach((cls) => values.add(cls));
    },
    remove: (...classes) => {
      classes.forEach((cls) => values.delete(cls));
    },
    toggle: (cls, force) => {
      if (force === undefined) {
        if (values.has(cls)) {
          values.delete(cls);
        } else {
          values.add(cls);
        }
        return values.has(cls);
      }
      if (force) {
        values.add(cls);
      } else {
        values.delete(cls);
      }
      return values.has(cls);
    },
    contains: (cls) => values.has(cls),
  };
};

const createElement = (initialClasses = []) => ({
  classList: createClassList(initialClasses),
  textContent: "",
  innerHTML: "",
  disabled: false,
  attributes: {},
  setAttribute(name, value) {
    this.attributes[name] = String(value);
  },
});

const createRefs = () => ({
  updatePrompt: createElement(["hidden"]),
  updatePromptCurrent: createElement(),
  updatePromptLatest: createElement(),
  updatePromptMessage: createElement(),
  installUpdatePromptButton: createElement(),
});

test("checkForUpdatesInBackground shows prompt when update is available", async () => {
  const refs = createRefs();
  const statuses = [];
  const settingsController = {
    async checkForUpdates() {
      return {
        skipped: false,
        failed: false,
        outdated: true,
        currentVersion: "v1.0.0",
        latestVersion: "v1.1.0",
        message: "Update available: v1.0.0 -> v1.1.0",
      };
    },
    async installLatestUpdate() {
      return { ok: true };
    },
  };

  const controller = createUpdateNotifierController({
    refs,
    settingsController,
    onStatus: (message) => statuses.push(message),
  });

  await controller.checkForUpdatesInBackground();

  assert.equal(refs.updatePrompt.classList.contains("hidden"), false);
  assert.equal(refs.updatePromptCurrent.textContent, "v1.0.0");
  assert.equal(refs.updatePromptLatest.textContent, "v1.1.0");
  assert.equal(
    refs.updatePromptMessage.textContent.includes("v1.0.0 -> v1.1.0"),
    true,
  );
  assert.deepEqual(statuses, ["Update available."]);
});

test("checkForUpdatesInBackground keeps prompt hidden when no update", async () => {
  const refs = createRefs();
  const settingsController = {
    async checkForUpdates() {
      return {
        skipped: false,
        failed: false,
        outdated: false,
        currentVersion: "v1.1.0",
        latestVersion: "v1.1.0",
        message: "Govard Desktop is up to date (v1.1.0).",
      };
    },
    async installLatestUpdate() {
      return { ok: true };
    },
  };

  const controller = createUpdateNotifierController({
    refs,
    settingsController,
    onStatus: () => {},
  });

  await controller.checkForUpdatesInBackground();

  assert.equal(refs.updatePrompt.classList.contains("hidden"), true);
});

test("dismissPrompt suppresses repeated prompt for same latest version", async () => {
  const refs = createRefs();
  const settingsController = {
    async checkForUpdates() {
      return {
        skipped: false,
        failed: false,
        outdated: true,
        currentVersion: "v1.0.0",
        latestVersion: "v1.2.0",
        message: "Update available: v1.0.0 -> v1.2.0",
      };
    },
    async installLatestUpdate() {
      return { ok: true };
    },
  };

  const controller = createUpdateNotifierController({
    refs,
    settingsController,
    onStatus: () => {},
  });

  await controller.checkForUpdatesInBackground();
  assert.equal(refs.updatePrompt.classList.contains("hidden"), false);

  controller.dismissPrompt();
  assert.equal(refs.updatePrompt.classList.contains("hidden"), true);

  await controller.checkForUpdatesInBackground();
  assert.equal(refs.updatePrompt.classList.contains("hidden"), true);
});

test("installLatestUpdateFromPrompt delegates to settings installer and hides prompt on success", async () => {
  const refs = createRefs();
  let installCalled = 0;
  const settingsController = {
    async checkForUpdates() {
      return {
        skipped: false,
        failed: false,
        outdated: true,
        currentVersion: "v1.0.0",
        latestVersion: "v1.3.0",
        message: "Update available: v1.0.0 -> v1.3.0",
      };
    },
    async installLatestUpdate() {
      installCalled += 1;
      return { ok: true, skipped: false };
    },
  };

  const controller = createUpdateNotifierController({
    refs,
    settingsController,
    onStatus: () => {},
  });

  await controller.checkForUpdatesInBackground();
  assert.equal(refs.updatePrompt.classList.contains("hidden"), false);

  const outcome = await controller.installLatestUpdateFromPrompt();
  assert.equal(Boolean(outcome?.ok), true);
  assert.equal(installCalled, 1);
  assert.equal(refs.updatePrompt.classList.contains("hidden"), true);
});

test("desktop shell contains update prompt actions", async () => {
  const html = await readFile(
    new URL("../../desktop/frontend/index.html", import.meta.url),
    "utf8",
  );

  assert.equal(
    html.includes('id="updatePrompt"'),
    true,
    "missing update prompt container",
  );
  assert.equal(
    html.includes('data-action="install-update-from-prompt"'),
    true,
    "missing install-update-from-prompt action",
  );
  assert.equal(
    html.includes('data-action="dismiss-update-prompt"'),
    true,
    "missing dismiss-update-prompt action",
  );
});
