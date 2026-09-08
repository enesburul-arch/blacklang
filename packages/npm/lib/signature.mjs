import { createPublicKey, verify } from "node:crypto";
import { readFile } from "node:fs/promises";
import path from "node:path";

export function signatureName(artifact) {
  return `${artifact}.sig`;
}

export async function assertSignature(archivePath, signaturePath, publicKeyText) {
  if (!publicKeyText || publicKeyText.trim() === "") {
    throw new Error("BLACKLANG_RELEASE_PUBLIC_KEY or BLACKLANG_RELEASE_PUBLIC_KEY_FILE is required for release signature verification");
  }
  const publicKey = createPublicKey(publicKeyText.replace(/\\n/g, "\n"));
  const archiveBytes = await readFile(archivePath);
  const signature = decodeSignature(await readFile(signaturePath));
  if (!verify(null, archiveBytes, publicKey, signature)) {
    throw new Error(`Ed25519 signature verification failed for ${path.basename(archivePath)}`);
  }
  return true;
}

export function publicKeyFromEnv(env = process.env) {
  if (env.BLACKLANG_RELEASE_PUBLIC_KEY && env.BLACKLANG_RELEASE_PUBLIC_KEY.trim() !== "") {
    return env.BLACKLANG_RELEASE_PUBLIC_KEY.replace(/\\n/g, "\n");
  }
  return "";
}

export async function publicKeyFromEnvOrFile(env = process.env) {
  const inline = publicKeyFromEnv(env);
  if (inline !== "") {
    return inline;
  }
  if (env.BLACKLANG_RELEASE_PUBLIC_KEY_FILE && env.BLACKLANG_RELEASE_PUBLIC_KEY_FILE.trim() !== "") {
    return await readFile(env.BLACKLANG_RELEASE_PUBLIC_KEY_FILE, "utf8");
  }
  return "";
}

function decodeSignature(bytes) {
  const text = bytes.toString("utf8").trim();
  if (/^[A-Za-z0-9+/= \r\n]+$/.test(text) && text.length > 0) {
    return Buffer.from(text.replace(/\s+/g, ""), "base64");
  }
  return bytes;
}
