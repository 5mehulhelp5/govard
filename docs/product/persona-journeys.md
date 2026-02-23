# Persona Journeys

## Developer Journey

1. Run `govard doctor fix-deps`.
2. Run `govard init`.
3. Inspect profile with `govard config profile`.
4. Apply profile if needed: `govard config profile apply`.
5. Start runtime: `govard env up`.
6. Use remote operations (`remote`, `sync`, `db`) during development.

## PM Journey

1. Install Govard from project onboarding docs.
2. Run `govard doctor fix-deps` and follow dependency guidance.
3. Run `govard env up`.
4. Open project URLs from desktop or `govard open`.
5. Use read-only environment checks (`status`, desktop dashboard).

## Tester Journey

1. Run `govard doctor fix-deps`.
2. Run `govard env up`.
3. Open app/admin tools with `govard open`.
4. Use snapshots for repeatable test states (`snapshot create`, `snapshot restore`).
5. Escalate failures with command output from `doctor`, `profile`, and operation logs.

