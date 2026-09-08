export function normalizeVersion(version) {
  if (!version) {
    throw new Error("BlackLang package version is required.");
  }
  return version.startsWith("v") ? version : `v${version}`;
}

export function artifactName(version, target) {
  const normalized = normalizeVersion(version);
  const extension = target.os === "windows" ? "zip" : "tar.gz";
  return `blacklang-${normalized}-${target.os}-${target.arch}.${extension}`;
}

export function githubReleaseURL(repository, version, artifact) {
  const normalized = normalizeVersion(version);
  const repo = repository || "enesburul-arch/blacklang";
  return `https://github.com/${repo}/releases/download/${normalized}/${artifact}`;
}
