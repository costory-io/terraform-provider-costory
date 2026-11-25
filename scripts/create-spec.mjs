#!/usr/bin/env node
/**
 * Create a new spec template under specs/.
 *
 * Usage:
 *   node scripts/create-spec.mjs <name> [description]
 *   yarn spec:new <name> [description]
 *
 * Example:
 *   yarn spec:new onboarding-v3 "Draft for the next onboarding flow"
 *   → specs/2026-05-18_onboarding-v3.md
 */

import { writeFileSync, mkdirSync } from "node:fs"
import { join, dirname } from "node:path"
import { fileURLToPath } from "node:url"

const __dirname = dirname(fileURLToPath(import.meta.url))
const SPECS_DIR = join(__dirname, "..", "specs")

function usage() {
  console.error(`Usage: yarn spec:new <name> [description]

  <name>         Slug for the file (e.g. onboarding-v3)
  [description]  Optional short summary (rendered in <small>)
`)
  process.exit(1)
}

function slugify(name) {
  return name
    .trim()
    .toLowerCase()
    .replace(/[\s_]+/g, "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "")
}

function toTitle(slug) {
  return slug
    .split("-")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ")
}

function formatDate(date) {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  return `${y}-${m}-${d}`
}

function buildContent(title, description) {
  const smallLine = description
    ? `<small>${description}</small>`
    : ""

  return `# ${title}
Status: Spec In Progress
${smallLine}

Related Files:
- 
- 

`
}

const [, , rawName, ...descriptionParts] = process.argv
const description = descriptionParts.join(" ").trim()

if (!rawName) {
  usage()
}

const slug = slugify(rawName)
if (!slug) {
  console.error("Error: name must contain at least one alphanumeric character.")
  process.exit(1)
}

const date = formatDate(new Date())
const filename = `${date}_${slug}.md`
const filepath = join(SPECS_DIR, filename)

mkdirSync(SPECS_DIR, { recursive: true })

const content = buildContent(toTitle(slug), description)
try {
  writeFileSync(filepath, content, { encoding: "utf8", flag: "wx" })
} catch (err) {
  if (err && err.code === "EEXIST") {
    console.error(`Error: spec already exists: ${filepath}`)
    process.exit(1)
  }
  throw err
}

console.log(`Created ${filepath}`)
