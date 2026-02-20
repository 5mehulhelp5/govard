import test from "node:test"
import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"

import {
  normalizeRemotePreset,
  normalizeRemotesPayload,
} from "../../desktop/frontend/modules/remotes.js"

test("normalizeRemotePreset canonicalizes aliases", () => {
  assert.equal(normalizeRemotePreset("file"), "files")
  assert.equal(normalizeRemotePreset("database"), "db")
  assert.equal(normalizeRemotePreset("full"), "full")
  assert.equal(normalizeRemotePreset("unknown"), "")
})

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
        Capabilities: ["files", "media"],
      },
    ],
    Warnings: ["warn"],
  })

  assert.equal(payload.project, "demo")
  assert.equal(payload.remotes.length, 1)
  assert.equal(payload.remotes[0].name, "staging")
  assert.equal(payload.remotes[0].authMethod, "keychain")
  assert.deepEqual(payload.remotes[0].capabilities, ["files", "media"])
  assert.deepEqual(payload.warnings, ["warn"])
})

test("desktop layout exposes remotes section and actions", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="remotes"'), true, "missing remotes panel")
  assert.equal(html.includes('data-action="refresh-remotes"'), true, "missing remotes refresh action")
  assert.equal(html.includes('data-action="save-remote"'), true, "missing remote save action")
})
