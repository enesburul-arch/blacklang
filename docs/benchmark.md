# Benchmark Command

`black benchmark` measures BlackLang source size, generated output size, AI task scenarios, token estimates, long-running AI eval corpus metadata, evidence-backed eval history, web coverage, and tracked issue exports.

```bash
black benchmark app.black --out generated --json
black benchmark app.black --out generated --ir
black benchmark tasks app.black --out generated --json
black benchmark tasks app.black --out generated --ir
black benchmark eval app.black --out generated --json
black benchmark eval app.black --out generated --ir
black benchmark eval-history --json
black benchmark eval-history --ir
black benchmark eval-history --history benchmarks/eval-history.blackdir --json
black benchmark coverage --json
black benchmark coverage --ir
black benchmark issues --json
black benchmark issues --ir
```

The command loads, parses, and validates the project. It then builds web output in a temporary directory and counts only files emitted by the BlackLang generator. It does not modify the configured generated output directory.

Source metrics include the primary `.black` source file and the configured `.blackthm` theme file when one exists.

The JSON result includes:

- `source`
- `generated`
- `ratios`
- `sourceFiles`
- `generatedFiles`
- `generatedKinds`

Use this command for benchmark notes and AI task reports instead of estimating generated size by hand.

`black benchmark tasks` reports deterministic AI task scenarios and token estimates for common generated web work:

- bounded report query
- row-level custom action
- deterministic fixture rows
- browser-check expectations
- owner or tenant row policy
- explicit webhook update handler
- generated ops probes
- local preview deploy and rollback metadata

Each scenario includes:

- `projectEvidence`
- `blacklangEdits`
- `generatedImpact`
- `commands`
- `estimatedTokens`

Token estimates are planning signals. They are derived from the current source/generated line and byte counts, fixed scenario context sizes, and deterministic text estimates. They are not billed-token measurements and they do not require calling an AI model.

`black benchmark eval` turns those task scenarios into a long-running AI eval corpus. It is read-only metadata: it does not call an AI model, mutate generated output, commit, push, deploy, or store secrets.

The JSON result includes:

- `suite`
- `cases`
- `totals`

Each case includes a reusable prompt, expected evidence, required validation commands, a 100-point scoring rubric, and the same deterministic token estimate signal from the task benchmark. Use `--ir` when an agent needs the compact corpus identity, case count, score ceiling, run count, and token totals.

`black benchmark eval-history` reads `benchmarks/eval-history.blackdir` and reports the current published-local eval result history without running a model, mutating generated output, committing, pushing, deploying, or storing secrets.

The JSON result includes:

- `history`
- `summary`
- `runs`
- `errors`

Each run records suite id, source, status, coverage signal, case/repeat counts, pass/fail totals, evidence paths, notes, and model or validation-source results. Local validation entries may record compiler and generated-app checks, but must not be presented as billed model benchmark scores. External model-to-model scores should be appended only after repeated harness execution, command evidence, and release review. Use `--history <file>` to read a non-default manifest.

`black benchmark coverage` reports the current weighted web coverage matrix:

- `completionPercent`
- `areas`
- `issues`
- `milestones`

Each area includes deterministic `weight`, `score`, `status`, `implemented`, `missing`, and `next` fields. The issue list turns roadmap gaps into stable IDs such as `WEB-LAYOUT-001`, `WEB-DATA-007`, `WEB-DATA-009`, `WEB-SEC-001`, `WEB-SEC-002`, `WEB-SEC-003`, `WEB-SEC-004`, `WEB-SEC-005`, `WEB-OPS-001`, `WEB-OPS-003`, `WEB-OPS-004`, `WEB-OPS-005`, `WEB-TEST-006`, `WEB-TEST-007`, `WEB-DOCS-003`, `WEB-DOCS-004`, `WEB-DOCS-005`, `WEB-ECO-003`, and `WEB-ECO-004`. Use this before claiming the web target is complete.

`black benchmark issues` reports the same tracked issue IDs as a smaller export:

- `summary.total`
- `summary.open`
- `summary.done`
- `issues`
- `openIssues`

Use it when an AI agent, CI job, or GitHub issue sync needs the open roadmap gaps without loading the full weighted area matrix.

Compiler tests also keep a compact Warehouse golden manifest under:

```text
packages/cli/cmd/black/testdata/golden/warehouse-build-manifest.json
```

The manifest records generated file paths, kinds, line counts, byte counts, and SHA-256 hashes. This catches unintended generator output changes without storing full generated files as test snapshots.
