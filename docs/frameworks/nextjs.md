# Next.js

Govard supports Next.js development with a streamlined Node.js stack and automated project initialization.

## Requirements

- Node.js 20 or 24 (default 24)
- No PHP or MariaDB required by default

## Detection

Govard detects Next.js when `next` is present in `package.json`.

## Default Stack

- **Runtime**: Node.js `24`
- **Port**: `3000` (internal)
- **Web Root**: Project root

## Bootstrap

### Fresh Install

When running `govard bootstrap --fresh --framework nextjs`, Govard:

1. Runs `npx create-next-app@latest .` with:
   - TypeScript
   - Tailwind CSS
   - ESLint
   - App Router
   - No `src/` directory
2. Runs `npm install`
3. Creates a default `.env.local`

### Clone Workflow

When cloning a Next.js project, Govard:

1. Runs `npm install`
2. Creates `.env.local` and `.env.example` if they are missing

## Commands

Use `npm` for Next.js tasks:

```bash
govard tool npm run dev
govard tool npm run build
```

## Examples

```bash
# Start the development server
govard env up

# Run a production build inside the container
govard tool npm run build
```
