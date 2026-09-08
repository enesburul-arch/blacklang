import fs from "node:fs";
import path from "node:path";

const platformMap = new Map([
  ["win32:x64", { os: "windows", arch: "amd64", binary: "black.exe" }],
  ["win32:arm64", { os: "windows", arch: "arm64", binary: "black.exe" }],
  ["linux:x64", { os: "linux", arch: "amd64", binary: "black" }],
  ["linux:arm64", { os: "linux", arch: "arm64", binary: "black" }],
  ["darwin:x64", { os: "darwin", arch: "amd64", binary: "black" }],
  ["darwin:arm64", { os: "darwin", arch: "arm64", binary: "black" }],
]);

export function platformTarget(platform = process.platform, arch = process.arch) {
  const target = platformMap.get(`${platform}:${arch}`);
  if (!target) {
    throw new Error(`Unsupported BlackLang platform: ${platform}/${arch}`);
  }
  return target;
}

export function vendorDir(packageRoot, target) {
  return path.join(packageRoot, "vendor", `${target.os}-${target.arch}`);
}

export function vendorBinary(packageRoot, target = platformTarget()) {
  return path.join(vendorDir(packageRoot, target), target.binary);
}

export function resolveBinary(options = {}) {
  const env = options.env ?? process.env;
  if (env.BLACKLANG_BINARY) {
    return env.BLACKLANG_BINARY;
  }
  const target = platformTarget(options.platform, options.arch);
  const binary = vendorBinary(options.packageRoot, target);
  if (fs.existsSync(binary)) {
    return binary;
  }
  return target.binary;
}

export function supportedTargets() {
  return Array.from(platformMap.values()).map((target) => ({ ...target }));
}
