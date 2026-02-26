import test from "node:test"
import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"

import { normalizeOnboardingFramework } from "../../desktop/frontend/modules/onboarding.js"

test("normalizeOnboardingFramework canonicalizes empty and aliases", () => {
  assert.equal(normalizeOnboardingFramework(""), "")
  assert.equal(normalizeOnboardingFramework("auto"), "")
  assert.equal(normalizeOnboardingFramework("m2"), "magento2")
  assert.equal(normalizeOnboardingFramework("magento2"), "magento2")
  assert.equal(normalizeOnboardingFramework("custom"), "custom")
})

test("desktop layout exposes onboarding section and actions", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="onboardingModal"'), true, "missing onboarding modal")
  assert.equal(html.includes('data-action="browse-project"'), true, "missing browse action")
  assert.equal(html.includes('data-action="add-project"'), true, "missing add project action")
})
