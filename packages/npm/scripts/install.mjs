import { spawnSync } from "node:child_process";
import fs from "node:fs";
import { mkdtemp, readFile, rm, mkdir, copyFile, chmod } from "node:fs/promises";
import https from "node:https";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseChecksums, assertChecksum } from "../lib/checksum.mjs";
import { artifactName, githubReleaseURL, normalizeVersion } from "../lib/release.mjs";
import { assertSignature, publicKeyFromEnvOrFile, signatureName } from "../lib/signature.mjs";
import { platformTarget, vendorBinary, vendorDir } from "../lib/platform.mjs";

const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const packageJSON = JSON.parse(await readFile(path.join(packageRoot, "package.json"), "utf8"));
const version = normalizeVersion(process.env.npm_package_version || packageJSON.version);
const target = platformTarget(process.platform, process.arch);
const binaryPath = vendorBinary(packageRoot, target);

if (process.env.BLACKLANG_BINARY) {
  assertExecutable(process.env.BLACKLANG_BINARY);
  console.log(`blacklang: using BLACKLANG_BINARY=${process.env.BLACKLANG_BINARY}`);
  process.exit(0);
}

if (fs.existsSync(binaryPath)) {
  assertExecutable(binaryPath);
  console.log(`blacklang: using vendored binary ${binaryPath}`);
  process.exit(0);
}

if (process.env.BLACKLANG_SKIP_DOWNLOAD === "1" || process.env.BLACKLANG_ALLOW_DOWNLOAD !== "1") {
  console.log("blacklang: native binary not vendored; set BLACKLANG_BINARY or BLACKLANG_ALLOW_DOWNLOAD=1 for release download.");
  process.exit(0);
}

const repository = process.env.BLACKLANG_RELEASE_REPOSITORY || "enesburul-arch/blacklang";
const artifact = artifactName(version, target);
const releaseRoot = process.env.BLACKLANG_RELEASE_ROOT;
const tempRoot = await mkdtemp(path.join(os.tmpdir(), "blacklang-npm-"));

try {
  const archivePath = path.join(tempRoot, artifact);
  const checksumPath = path.join(tempRoot, "checksums.sha256");
  const detachedSignature = signatureName(artifact);
  const signaturePath = path.join(tempRoot, detachedSignature);
  if (releaseRoot) {
    await copyFile(path.join(releaseRoot, artifact), archivePath);
    await copyFile(path.join(releaseRoot, "checksums.sha256"), checksumPath);
    await copyFile(path.join(releaseRoot, detachedSignature), signaturePath);
  } else {
    await download(githubReleaseURL(repository, version, artifact), archivePath);
    await download(githubReleaseURL(repository, version, "checksums.sha256"), checksumPath);
    await download(githubReleaseURL(repository, version, detachedSignature), signaturePath);
  }

  const checksums = parseChecksums(await readFile(checksumPath, "utf8"));
  const expected = checksums.get(artifact);
  if (!expected) {
    throw new Error(`No checksum entry for ${artifact}`);
  }
  await assertChecksum(archivePath, expected);
  await assertSignature(archivePath, signaturePath, await publicKeyFromEnvOrFile());

  const extractedRoot = path.join(tempRoot, "extracted");
  await mkdir(extractedRoot, { recursive: true });
  extractArchive(archivePath, extractedRoot, target.os);

  const extractedBinary = findBinary(extractedRoot, target.binary);
  if (!extractedBinary) {
    throw new Error(`Release archive did not contain ${target.binary}`);
  }

  await mkdir(vendorDir(packageRoot, target), { recursive: true });
  await copyFile(extractedBinary, binaryPath);
  if (target.os !== "windows") {
    await chmod(binaryPath, 0o755);
  }
  assertExecutable(binaryPath);
  console.log(`blacklang: installed ${artifact} to ${binaryPath}`);
} finally {
  await rm(tempRoot, { recursive: true, force: true });
}

function assertExecutable(binary) {
  if (!fs.existsSync(binary)) {
    throw new Error(`BlackLang binary does not exist: ${binary}`);
  }
  const result = spawnSync(binary, ["version"], {
    encoding: "utf8",
    windowsHide: true,
  });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `BlackLang binary failed version check: ${binary}`);
  }
}

function extractArchive(archivePath, destination, osName) {
  if (osName === "windows") {
    const result = spawnSync("powershell", [
      "-NoProfile",
      "-Command",
      "Expand-Archive -LiteralPath $args[0] -DestinationPath $args[1] -Force",
      archivePath,
      destination,
    ], {
      encoding: "utf8",
      windowsHide: true,
    });
    if (result.status !== 0) {
      throw new Error(result.stderr || result.stdout || `Could not extract ${archivePath}`);
    }
    return;
  }

  const result = spawnSync("tar", ["-xzf", archivePath, "-C", destination], {
    encoding: "utf8",
    windowsHide: true,
  });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || `Could not extract ${archivePath}`);
  }
}

function findBinary(root, binaryName) {
  const entries = fs.readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(root, entry.name);
    if (entry.isFile() && entry.name === binaryName) {
      return fullPath;
    }
    if (entry.isDirectory()) {
      const found = findBinary(fullPath, binaryName);
      if (found) {
        return found;
      }
    }
  }
  return "";
}

function download(url, file) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        download(response.headers.location, file).then(resolve, reject);
        return;
      }
      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`Download failed ${response.statusCode}: ${url}`));
        return;
      }
      const stream = fs.createWriteStream(file);
      response.pipe(stream);
      stream.on("finish", () => {
        stream.close(resolve);
      });
      stream.on("error", reject);
    });
    request.on("error", reject);
  });
}
