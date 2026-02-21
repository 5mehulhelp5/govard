import test from "node:test"
import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"

import { normalizeMailpitURL } from "../../desktop/frontend/modules/mail.js"

test("normalizeMailpitURL falls back to default", () => {
  assert.equal(normalizeMailpitURL(""), "https://mail.govard.test")
})

test("normalizeMailpitURL normalizes proxy targets", () => {
  assert.equal(normalizeMailpitURL("workspace.internal"), "https://mail.workspace.internal")
  assert.equal(normalizeMailpitURL("https://workspace.internal/"), "https://mail.workspace.internal")
})

test("desktop layout removes mailpit inbox section", async () => {
  const html = await readFile(new URL("../../desktop/frontend/index.html", import.meta.url), "utf8")
  assert.equal(html.includes('id="mailpit"'), false, "mailpit panel should be removed")
  assert.equal(html.includes('id="mailFrame"'), false, "mail iframe should be removed")
  assert.equal(html.includes('data-action="refresh-mail"'), false, "mail refresh action should be removed")
  assert.equal(html.includes('data-action="open-mail-external"'), false, "mail external action should be removed")
})
