import test from "node:test"
import assert from "node:assert/strict"

import { resolveLogTarget } from "../../desktop/frontend/modules/logs.js"

test("resolveLogTarget returns selected project and service", () => {
  const value = resolveLogTarget({
    project: "demo",
    service: "php",
  })
  assert.equal(value.project, "demo")
  assert.equal(value.service, "php")
})

test("resolveLogTarget applies defaults", () => {
  const value = resolveLogTarget({})
  assert.equal(value.project, "")
  assert.equal(value.service, "web")
})

