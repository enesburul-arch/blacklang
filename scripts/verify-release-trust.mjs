#!/usr/bin/env node

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(fileURLToPath(new URL("..", import.meta.url)));
const args = process.argv.slice(2);
const json = args.includes("--json");
const strict = args.includes("--strict");
const releaseDirArg = args.find((arg) => !arg.startsWith("-")) ?? "artifacts/releases/<version>";
const releaseDir = path.resolve(root, releaseDirArg);
const command = "release:verify-trust";

function relativePath(filePath) {
  return path.relative(root, filePath).split(path.sep).join("/");
}

function readText(filePath) {
  return fs.readFileSync(filePath, "utf8").replace(/^\uFEFF/, "");
}

function parseReleaseManifest(content) {
  const manifest = {
    release: "",
    channel: "",
    cli: "",
    artifacts: []
  };
  const errors = [];
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) {
      continue;
    }
    const parts = line.split(/\s+/);
    switch (parts[0]) {
      case "release":
        manifest.release = parts[1] ?? "";
        break;
      case "channel":
        manifest.channel = parts[1] ?? "";
        break;
      case "cli":
        manifest.cli = parts[1] ?? "";
        break;
      case "artifact":
        if (parts.length !== 7 || parts[5] !== "sha256") {
          errors.push(`invalid artifact line: ${line}`);
          break;
        }
        manifest.artifacts.push({
          os: parts[1],
          arch: parts[2],
          kind: parts[3],
          name: parts[4],
          sha256: parts[6]
        });
        break;
      default:
        errors.push(`unknown release.blackdir line: ${line}`);
    }
  }
  if (manifest.release === "") {
    errors.push("release.blackdir is missing release <version>");
  }
  if (manifest.channel === "") {
    errors.push("release.blackdir is missing channel <dev|preview|stable>");
  }
  if (manifest.cli === "") {
    errors.push("release.blackdir is missing cli <command>");
  }
  if (manifest.artifacts.length === 0) {
    errors.push("release.blackdir has no artifact lines");
  }
  return { manifest, errors };
}

function parseChecksums(content) {
  const checksums = new Map();
  const errors = [];
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    const match = line.match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
    if (!match) {
      errors.push(`invalid checksums.sha256 line: ${line}`);
      continue;
    }
    checksums.set(match[2].trim(), match[1].toLowerCase());
  }
  return { checksums, errors };
}

function sha256File(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

function loadPublicKey() {
  const inline = process.env.BLACKLANG_RELEASE_PUBLIC_KEY;
  if (inline && inline.trim() !== "") {
    return {
      present: true,
      source: "env:BLACKLANG_RELEASE_PUBLIC_KEY",
      key: inline.replace(/\\n/g, "\n")
    };
  }

  const file = process.env.BLACKLANG_RELEASE_PUBLIC_KEY_FILE;
  if (file && file.trim() !== "") {
    const fullPath = path.resolve(process.cwd(), file);
    return {
      present: fs.existsSync(fullPath),
      source: `env:BLACKLANG_RELEASE_PUBLIC_KEY_FILE:${file}`,
      key: fs.existsSync(fullPath) ? readText(fullPath) : ""
    };
  }

  return { present: false, source: "", key: "" };
}

function readSignature(filePath) {
  const bytes = fs.readFileSync(filePath);
  const text = bytes.toString("utf8").trim();
  if (/^[A-Za-z0-9+/= \r\n]+$/.test(text) && text.length > 0) {
    try {
      return Buffer.from(text.replace(/\s+/g, ""), "base64");
    } catch {
      return bytes;
    }
  }
  return bytes;
}

function verifyEd25519(bytes, signature, publicKeyPem) {
  const publicKey = crypto.createPublicKey(publicKeyPem);
  return crypto.verify(null, bytes, publicKey, signature);
}

const releaseManifestPath = path.join(releaseDir, "release.blackdir");
const checksumsPath = path.join(releaseDir, "checksums.sha256");
const publicKey = loadPublicKey();
const missing = [];
const errors = [];
const checks = [];
let manifest = {
  release: "",
  channel: "",
  cli: "",
  artifacts: []
};
let checksums = new Map();

if (!fs.existsSync(releaseDir)) {
  missing.push(relativePath(releaseDir));
}

if (!fs.existsSync(releaseManifestPath)) {
  missing.push(relativePath(releaseManifestPath));
} else {
  const parsed = parseReleaseManifest(readText(releaseManifestPath));
  manifest = parsed.manifest;
  errors.push(...parsed.errors);
}

if (!fs.existsSync(checksumsPath)) {
  missing.push(relativePath(checksumsPath));
} else {
  const parsed = parseChecksums(readText(checksumsPath));
  checksums = parsed.checksums;
  errors.push(...parsed.errors);
}

for (const artifact of manifest.artifacts) {
  const artifactPath = path.join(releaseDir, artifact.name);
  const signatureName = `${artifact.name}.sig`;
  const signaturePath = path.join(releaseDir, signatureName);
  const check = {
    name: artifact.name,
    os: artifact.os,
    arch: artifact.arch,
    kind: artifact.kind,
    sha256: artifact.sha256,
    exists: fs.existsSync(artifactPath),
    checksumListed: checksums.has(artifact.name),
    checksumMatches: false,
    signature: signatureName,
    signaturePresent: fs.existsSync(signaturePath),
    signatureVerified: null
  };

  if (!check.exists) {
    missing.push(relativePath(artifactPath));
  }

  if (!check.checksumListed) {
    missing.push(`${relativePath(checksumsPath)}:${artifact.name}`);
  }

  if (check.exists) {
    const actual = sha256File(artifactPath);
    const listed = checksums.get(artifact.name);
    check.checksumMatches = actual === artifact.sha256.toLowerCase() && listed === artifact.sha256.toLowerCase();
    if (!check.checksumMatches) {
      errors.push(`${artifact.name}: checksum mismatch between archive, release.blackdir, and checksums.sha256`);
    }
  }

  if (!check.signaturePresent) {
    missing.push(relativePath(signaturePath));
  } else if (publicKey.present && check.exists) {
    try {
      check.signatureVerified = verifyEd25519(
        fs.readFileSync(artifactPath),
        readSignature(signaturePath),
        publicKey.key
      );
      if (!check.signatureVerified) {
        errors.push(`${artifact.name}: Ed25519 signature verification failed`);
      }
    } catch (error) {
      check.signatureVerified = false;
      errors.push(`${artifact.name}: ${error.message}`);
    }
  }

  checks.push(check);
}

if (checks.length > 0 && !publicKey.present) {
  missing.push("BLACKLANG_RELEASE_PUBLIC_KEY or BLACKLANG_RELEASE_PUBLIC_KEY_FILE");
}

const uniqueMissing = [...new Set(missing)];
const result = {
  success: errors.length === 0,
  command,
  releaseDir: relativePath(releaseDir),
  release: manifest.release,
  channel: manifest.channel,
  cli: manifest.cli,
  signatureAlgorithm: "ed25519",
  signatureFilePattern: "<artifact>.sig",
  publicKey: {
    env: "BLACKLANG_RELEASE_PUBLIC_KEY",
    fileEnv: "BLACKLANG_RELEASE_PUBLIC_KEY_FILE",
    present: publicKey.present,
    source: publicKey.source
  },
  ready: errors.length === 0 && uniqueMissing.length === 0 && checks.length > 0 && checks.every((check) => check.signatureVerified === true),
  checks,
  missing: uniqueMissing,
  errors,
  policy: "release archives must match release.blackdir, checksums.sha256, detached Ed25519 signatures, and a trusted public key before public install paths are trusted"
};

if (json) {
  console.log(JSON.stringify(result, null, 2));
} else {
  console.log(`${command} ${result.ready ? "ready" : "not-ready"}`);
  console.log(`release: ${result.release || "(unknown)"}`);
  console.log(`artifacts: ${result.checks.length}`);
  if (result.missing.length > 0) {
    console.log("missing:");
    for (const item of result.missing) {
      console.log(`- ${item}`);
    }
  }
  if (result.errors.length > 0) {
    console.log("errors:");
    for (const item of result.errors) {
      console.log(`- ${item}`);
    }
  }
}

if (strict && (!result.success || !result.ready)) {
  process.exit(1);
}
