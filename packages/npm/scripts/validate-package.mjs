import { generateKeyPairSync, sign } from "node:crypto";
import { readFile, mkdtemp, writeFile, chmod, rm } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseChecksums, sha256File, assertChecksum } from "../lib/checksum.mjs";
import { artifactName, githubReleaseURL, normalizeVersion } from "../lib/release.mjs";
import { assertSignature, publicKeyFromEnv, signatureName } from "../lib/signature.mjs";
import { platformTarget, supportedTargets, vendorBinary, resolveBinary } from "../lib/platform.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const packageJSON = JSON.parse(await readFile(path.join(root, "package.json"), "utf8"));

assert(packageJSON.name === "blacklang", "package name must stay blacklang");
assert(packageJSON.type === "module", "package must stay ESM");
assert(packageJSON.bin.black === "./bin/blacklang.mjs", "black bin must point to wrapper");
assert(packageJSON.bin.blacklang === "./bin/blacklang.mjs", "blacklang bin must point to wrapper");
assert(packageJSON.scripts.postinstall === "node scripts/install.mjs", "postinstall must run installer");
assert(packageJSON.files.includes("vendor/"), "vendor directory must be included in package files");

assert(normalizeVersion("0.1.0") === "v0.1.0", "version normalization failed");
assert(normalizeVersion("v0.1.0") === "v0.1.0", "version normalization should keep v prefix");

const windows = platformTarget("win32", "x64");
assert(windows.os === "windows" && windows.arch === "amd64" && windows.binary === "black.exe", "Windows platform mapping failed");
const linux = platformTarget("linux", "arm64");
assert(linux.os === "linux" && linux.arch === "arm64" && linux.binary === "black", "Linux ARM platform mapping failed");
assert(supportedTargets().length >= 6, "expected common platform targets");
assert(artifactName("0.1.0", windows) === "blacklang-v0.1.0-windows-amd64.zip", "Windows artifact naming failed");
assert(artifactName("0.1.0", linux) === "blacklang-v0.1.0-linux-arm64.tar.gz", "Linux artifact naming failed");
assert(signatureName("blacklang-v0.1.0-windows-amd64.zip") === "blacklang-v0.1.0-windows-amd64.zip.sig", "signature naming failed");
assert(githubReleaseURL("owner/repo", "0.1.0", "artifact.zip") === "https://github.com/owner/repo/releases/download/v0.1.0/artifact.zip", "release URL failed");

const checksums = parseChecksums("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  blacklang.zip\n");
assert(checksums.get("blacklang.zip") === "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "checksum parser failed");

const tempRoot = await mkdtemp(path.join(os.tmpdir(), "blacklang-npm-test-"));
try {
  const fakeBinary = path.join(tempRoot, process.platform === "win32" ? "black.cmd" : "black");
  const fakeSource = process.platform === "win32"
    ? "@echo off\r\necho 0.1.0-dev\r\n"
    : "#!/usr/bin/env sh\nprintf '0.1.0-dev\\n'\n";
  await writeFile(fakeBinary, fakeSource, "utf8");
  if (process.platform !== "win32") {
    await chmod(fakeBinary, 0o755);
  }
  assert(resolveBinary({ env: { BLACKLANG_BINARY: fakeBinary }, packageRoot: root }) === fakeBinary, "BLACKLANG_BINARY override failed");

  const checksumFile = path.join(tempRoot, "payload.txt");
  await writeFile(checksumFile, "blacklang", "utf8");
  const actual = await sha256File(checksumFile);
  await assertChecksum(checksumFile, actual);

  const { publicKey, privateKey } = generateKeyPairSync("ed25519");
  const signatureFile = path.join(tempRoot, "payload.txt.sig");
  const signature = sign(null, Buffer.from("blacklang"), privateKey).toString("base64");
  await writeFile(signatureFile, signature, "utf8");
  const publicKeyPem = publicKey.export({ type: "spki", format: "pem" });
  await assertSignature(checksumFile, signatureFile, publicKeyPem);
  assert(publicKeyFromEnv({ BLACKLANG_RELEASE_PUBLIC_KEY: publicKeyPem.replace(/\n/g, "\\n") }).includes("PUBLIC KEY"), "public key env parser failed");
} finally {
  await rm(tempRoot, { recursive: true, force: true });
}

const binPath = path.join(root, "bin", "blacklang.mjs");
const installPath = path.join(root, "scripts", "install.mjs");
for (const file of [binPath, installPath]) {
  const result = spawnSync(process.execPath, ["--check", file], { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `${file} syntax check failed`);
  }
}

for (const file of [path.join(root, "lib", "signature.mjs")]) {
  const result = spawnSync(process.execPath, ["--check", file], { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `${file} syntax check failed`);
  }
}

console.log("BlackLang npm wrapper validation passed");

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
