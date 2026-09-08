import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../..", import.meta.url));

const files = [
  {
    path: "packages/registry/package-index.blackdir",
    required: [
      "registry blacklang-packages",
      "package npm:blacklang",
      "package vscode:blacklang-vscode",
      "package openvsx:blacklang-vscode",
      "package cursor:blacklang-vscode",
      "trust native-cli-forwarder true",
      "trust compiler-backed-diagnostics true",
      "trust signed-release-verification true",
      "trust release-signature-algorithm ed25519",
      "trust release-transparency-log true",
      "trust release-key-rotation true"
    ]
  },
  {
    path: "packages/registry/public-index.blackdir",
    required: [
      "publicIndex blacklang-ecosystem-public-index",
      "path website/ecosystem-index.json",
      "format json",
      "format blackir",
      "package npm:blacklang",
      "package openvsx:blacklang-vscode",
      "package cursor:blacklang-vscode",
      "adapter deploy-cloud:docker-provider",
      "adapter observability:otlp",
      "adapter editor:open-vsx",
      "adapter editor:cursor-compatible",
      "adapter target:web-react-node-mysql",
      "trust compiler-owned-ecosystem-source true",
      "trust no-secret-values true",
      "publish external-deploy-after-release-owner-review"
    ]
  },
  {
    path: "packages/registry/trust-policy.blackdir",
    required: [
      "policy package-registry-trust",
      "require release-manifest release.blackdir",
      "require checksum-file checksums.sha256",
      "require detached-signature-pattern \"<artifact>.sig\"",
      "require release-signature-algorithm ed25519",
      "require release-trust-command \"node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict\"",
      "require release-transparency-policy packages/registry/release-transparency.blackdir",
      "require release-key-rotation-policy packages/registry/key-rotation-policy.blackdir",
      "require no-language-reimplementation true"
    ]
  },
  {
    path: "packages/registry/release-transparency.blackdir",
    required: [
      "policy release-transparency-log",
      "log transparency.blackdir",
      "hash sha256",
      "append-only true",
      "entry field release",
      "entry field artifact",
      "entry field keyId",
      "entry field previousEntryHash",
      "entry field entryHash",
      "require entry-hash-covers \"release artifact sha256 signature keyId previousEntryHash\"",
      "require no-secret-values true"
    ]
  },
  {
    path: "packages/registry/key-rotation-policy.blackdir",
    required: [
      "policy release-key-rotation",
      "key-id-format sha256-public-key-spki-prefix",
      "minimum-overlap-days 30",
      "revocation-file key-revocations.blackdir",
      "require new-public-key-listed-before-signing true",
      "require old-new-key-overlap-days 30",
      "require revoked-key-refused-for-new-release true",
      "require private-key-outside-repository true",
      "require no-secret-values true"
    ]
  },
  {
    path: "packages/registry/key-revocations.blackdir",
    required: [
      "revocations release-keys",
      "status prepared-local",
      "No revoked release keys"
    ]
  },
  {
    path: "adapters/marketplace/adapter-index.blackdir",
    required: [
      "marketplace provider-adapters",
      "adapter deploy-cloud:docker-provider",
      "adapter target:web-react-node-mysql",
      "adapter observability:webhook",
      "adapter observability:otlp",
      "adapter editor:vscode",
      "adapter editor:open-vsx",
      "adapter editor:cursor-compatible",
      "capability \"MySQL Docker Compose service\"",
      "capability \"VSIX package reuse\"",
      "trust env-only-secrets true",
      "capability ops/observability.json",
      "capability scripts/cloud-exec.mjs",
      "capability \"explicit apply provider CLI runner\"",
      "trust signed-package-verification true",
      "trust explicit-apply-before-mutation true",
      "trust release-signature-algorithm ed25519",
      "trust release-transparency-log true",
      "trust release-key-rotation true"
    ]
  },
  {
    path: "adapters/marketplace/trust-policy.blackdir",
    required: [
      "policy adapter-marketplace-trust",
      "require env-only-secrets true",
      "require read-only-plan-before-mutation true",
      "require json-and-ir-discovery true",
      "require signed-package-verification true",
      "require release-trust-command \"node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict\"",
      "require release-transparency-policy packages/registry/release-transparency.blackdir",
      "require release-key-rotation-policy packages/registry/key-rotation-policy.blackdir"
    ]
  },
  {
    path: "website/ecosystem-index.json",
    required: [
      "\"id\": \"public-index:blacklang-ecosystem\"",
      "\"path\": \"website/ecosystem-index.json\"",
      "\"openvsx:blacklang-vscode\"",
      "\"cursor:blacklang-vscode\"",
      "\"editor:open-vsx\"",
      "\"editor:cursor-compatible\"",
      "\"target:web-react-node-mysql\"",
      "\"Public index content must contain no provider credentials"
    ]
  },
  {
    path: "scripts/verify-release-trust.mjs",
    required: [
      "const command = \"release:verify-trust\"",
      "signatureAlgorithm: \"ed25519\"",
      "BLACKLANG_RELEASE_PUBLIC_KEY",
      "checksums.sha256",
      "release.blackdir"
    ]
  }
];

const errors = [];
for (const file of files) {
  const fullPath = path.join(root, ...file.path.split("/"));
  let content = "";
  try {
    content = fs.readFileSync(fullPath, "utf8");
  } catch (error) {
    errors.push(`${file.path}: ${error.message}`);
    continue;
  }
  for (const marker of file.required) {
    if (!content.includes(marker)) {
      errors.push(`${file.path}: missing ${marker}`);
    }
  }
}

if (errors.length > 0) {
  console.error(JSON.stringify({ success: false, command: "validate-registry", errors }, null, 2));
  process.exit(1);
}

console.log(JSON.stringify({
  success: true,
  command: "validate-registry",
  manifests: files.map((file) => file.path),
  packages: ["npm:blacklang", "vscode:blacklang-vscode", "openvsx:blacklang-vscode", "cursor:blacklang-vscode"],
  adapters: ["deploy-cloud:docker-provider", "observability:webhook", "observability:otlp", "editor:vscode", "editor:open-vsx", "editor:cursor-compatible", "target:web-react-node-mysql"],
  indexes: ["public-index:blacklang-ecosystem"],
  policies: ["package-registry-trust", "release-transparency-log", "release-key-rotation", "adapter-marketplace-trust"]
}, null, 2));
