import test from "node:test"
import assert from "node:assert/strict"

import {
  formatMetricMB,
  formatMetricPercent,
  normalizeMetricsPayload,
} from "../../desktop/frontend/modules/metrics.js"

test("normalizeMetricsPayload maps summary and project metrics", () => {
  const payload = normalizeMetricsPayload({
    Summary: {
      activeProjects: 2,
      cpuPercent: 24.5,
      memoryMB: 1536.2,
      netRxMB: 44.2,
      netTxMB: 22.1,
      oomProjects: 1,
    },
    Projects: [
      { project: "demo", status: "running", cpuPercent: 10.1, memoryMB: 200.4, memoryPercent: 52.2, netRxMB: 8, netTxMB: 2 },
    ],
    Warnings: ["OOM detected in demo"],
  })

  assert.equal(payload.summary.activeProjects, 2)
  assert.equal(payload.summary.cpuPercent, 24.5)
  assert.equal(payload.projects.length, 1)
  assert.equal(payload.projects[0].project, "demo")
  assert.equal(payload.warnings.length, 1)
})

test("normalizeMetricsPayload falls back to safe defaults", () => {
  const payload = normalizeMetricsPayload({})
  assert.equal(payload.summary.activeProjects, 0)
  assert.equal(payload.summary.cpuPercent, 0)
  assert.equal(payload.summary.memoryMB, 0)
  assert.equal(payload.summary.netRxMB, 0)
  assert.equal(payload.summary.netTxMB, 0)
  assert.equal(payload.summary.oomProjects, 0)
  assert.deepEqual(payload.projects, [])
  assert.deepEqual(payload.warnings, [])
})

test("formatMetric helpers render stable strings", () => {
  assert.equal(formatMetricPercent(12.345), "12.3%")
  assert.equal(formatMetricPercent(0), "0.0%")
  assert.equal(formatMetricMB(1024.08), "1024.1 MB")
  assert.equal(formatMetricMB(0), "0.0 MB")
})
