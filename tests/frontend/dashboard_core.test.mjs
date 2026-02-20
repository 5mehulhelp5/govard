import test from "node:test"
import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"

import { normalizeDashboardPayload } from "../../desktop/frontend/modules/dashboard.js"

test("normalizeDashboardPayload maps core values", () => {
  const value = normalizeDashboardPayload({
    ActiveEnvironments: 2,
    RunningServices: 9,
    QueuedTasks: 1,
    ActiveSummary: "demo",
    Environments: [{ Name: "demo.test" }],
    Warnings: ["warn"],
  })

  assert.equal(value.active, 2)
  assert.equal(value.services, 9)
  assert.equal(value.queued, 1)
  assert.equal(value.activeSummary, "demo")
  assert.equal(value.environments.length, 1)
  assert.equal(value.warnings.length, 1)
})

test("normalizeDashboardPayload uses safe defaults", () => {
  const value = normalizeDashboardPayload({})
  assert.equal(value.active, 0)
  assert.equal(value.services, 0)
  assert.equal(value.queued, 0)
  assert.deepEqual(value.environments, [])
  assert.deepEqual(value.warnings, [])
})

test("core layout excludes removed heavyweight sections", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  for (const id of ["operations", "onboardingChecks", "workflowSync", "activityList"]) {
    assert.equal(html.includes(`id="${id}"`), false, `expected removed section ${id}`)
  }
})

test("quick actions uses structured action grids", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('class="panel quick-actions"'), true, "missing quick-actions panel class")
  assert.equal(
    html.includes('class="quick-actions__primary action-grid action-grid--primary"'),
    true,
    "missing quick actions primary grid",
  )
  assert.equal(
    html.includes('class="quick-actions__tools action-grid action-grid--tools"'),
    true,
    "missing quick actions tools grid",
  )
})

test("logs section uses two-row controls layout", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('class="panel panel--logs" id="logs"'), true, "missing logs panel class")
  assert.equal(html.includes('class="logs-controls__row logs-controls__row--top"'), true, "missing logs top row")
  assert.equal(
    html.includes('class="logs-controls__row logs-controls__row--bottom"'),
    true,
    "missing logs bottom row",
  )
})

test("metrics section is present for resource monitoring", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="metrics"'), true, "missing metrics section")
  assert.equal(
    html.includes('data-action="refresh-metrics"'),
    true,
    "missing metrics refresh action button",
  )
})
