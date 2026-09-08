import { readFile } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

const requiredFiles = [
  "package.json",
  "extension.js",
  "language-configuration.json",
  "syntaxes/black.tmLanguage.json",
  "syntaxes/blackthm.tmLanguage.json",
  "README.md",
  "CHANGELOG.md",
];

for (const file of requiredFiles) {
  await readText(file);
}

const manifest = JSON.parse(await readText("package.json"));
assert(manifest.name === "blacklang-vscode", "package name must stay blacklang-vscode");
assert(manifest.main === "./extension.js", "extension main must be extension.js");
assert(Array.isArray(manifest.activationEvents), "activationEvents must be an array");
assert(manifest.activationEvents.includes("onLanguage:blacklang"), "blacklang activation is required");
assert(manifest.activationEvents.includes("onLanguage:blacklang-theme"), "blacklang-theme activation is required");
assert(manifest.contributes.languages.some((language) => language.id === "blacklang" && language.extensions.includes(".black")), ".black language contribution is required");
assert(manifest.contributes.languages.some((language) => language.id === "blacklang-theme" && language.extensions.includes(".blackthm")), ".blackthm language contribution is required");
assert(manifest.contributes.commands.some((command) => command.command === "blacklang.refreshDiagnostics"), "refresh diagnostics command is required");
assert(manifest.contributes.commands.some((command) => command.command === "blacklang.inspectAffected"), "inspect affected command is required");
assert(manifest.contributes.configuration.properties["blacklang.cliPath"], "blacklang.cliPath config is required");

JSON.parse(await readText("language-configuration.json"));
JSON.parse(await readText("syntaxes/black.tmLanguage.json"));
JSON.parse(await readText("syntaxes/blackthm.tmLanguage.json"));

const extensionSource = await readText("extension.js");
assert(extensionSource.includes('"ide", "--json"'), "extension must load black ide --json");
assert(extensionSource.includes('"ide", "diagnostics"'), "extension must load black ide diagnostics");
assert(extensionSource.includes('"inspect", editor.document.fileName, "--affected"'), "extension must call inspect --affected");
assert(extensionSource.includes("registerCompletionItemProvider"), "completion provider is required");
assert(extensionSource.includes("registerCodeActionsProvider"), "code action provider is required");
assert(extensionSource.includes("FORMAT_REQUIRED"), "format quick fix is required");

const syntaxCheck = spawnSync(process.execPath, ["--check", path.join(root, "extension.js")], {
  encoding: "utf8",
});
if (syntaxCheck.status !== 0) {
  throw new Error(syntaxCheck.stderr || syntaxCheck.stdout || "extension.js syntax check failed");
}

console.log("BlackLang VS Code extension validation passed");

async function readText(relativePath) {
  return readFile(path.join(root, relativePath), "utf8");
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
