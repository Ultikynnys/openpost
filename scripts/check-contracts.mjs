import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const env = {
  ...process.env,
  GOCACHE: process.env.GOCACHE || path.join("/tmp", "openpost-go-cache"),
};

run("node", ["scripts/sync-docs-openapi.mjs"]);
run("pnpm", ["--filter", "@openpost/web", "generate:types"]);
run("git", [
  "diff",
  "--exit-code",
  "--",
  "frontend/openapi.json",
  "frontend/src/lib/api/types.d.ts",
  "docs-site/.generated/openapi.json",
  "docs-site/public/openapi.json",
  "docs-site/reference/cli.md",
]);

console.log("Generated API, TypeScript, docs, and CLI contracts are current.");

function run(command, args) {
  const result = spawnSync(command, args, { cwd: root, env, stdio: "inherit" });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}
