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

test("quick actions exposes desktop action contracts", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  for (const action of ["open-folder", "open-ide", "open-db-client", "open-mail-client"]) {
    assert.equal(html.includes(`data-action="${action}"`), true, `missing quick action ${action}`)
  }
  assert.equal(html.includes('data-action="open-mail"'), false, "legacy quick action should stay removed")
})

test("project management workspace keeps core IDs for controllers", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="envList"'), true, "missing environment list container")
  assert.equal(html.includes('id="metricsList"'), true, "missing metrics list container")
  assert.equal(html.includes('id="onboardingModal"'), true, "missing onboarding modal")
  assert.equal(html.includes('data-action="open-onboarding"'), true, "missing onboarding open action")
})

test("logs section exposes filtering and streaming controls", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="tab-logs"'), true, "missing logs tab")
  assert.equal(html.includes('id="logServiceSelector"'), true, "missing log service selector")
  assert.equal(html.includes('id="logSeverity"'), true, "missing log severity selector")
  assert.equal(html.includes('data-action="refresh-logs"'), true, "missing refresh logs action")
  assert.equal(html.includes('data-action="toggle-live"'), true, "missing toggle live action")
})

test("metrics section is present for resource monitoring", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  for (const id of ["metricCPU", "metricMemory", "metricsList"]) {
    assert.equal(html.includes(`id="${id}"`), true, `missing metrics field ${id}`)
  }
})
