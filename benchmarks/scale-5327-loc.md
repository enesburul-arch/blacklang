# BlackLang Scale Benchmark: 5327 LOC Synthetic Web Source

This is a measured local scale test, not a qualitative positioning note.

## Purpose

The goal is to check whether the current BlackLang compiler can parse and generate a large `.black` source file around 5000 lines.

This benchmark answers a narrow question:

> Can BlackLang currently handle a `.black` source file with roughly 5000 source lines, and how much generated web output does that produce?

## Test Shape

The synthetic source contains:

- One `app` declaration.
- One `target web` block.
- 313 entities.
- 313 pages.
- Each page has one table, one form, and create/edit/delete actions.

The generated output excludes `node_modules`, `dist`, and generated Prisma client output from line counting.

## Result

| Metric | Value |
|---|---:|
| BlackLang source lines, non-empty | 5327 |
| Entities | 313 |
| Pages | 313 |
| Build success | yes |
| Generated files | 1268 |
| Generated non-empty lines | 236610 |
| Generator time | 1.15 seconds |

## Interpretation

The current compiler has no explicit source line limit at 5000 lines.

In this synthetic test, a 5327-line `.black` file generated 1268 files and 236610 non-empty generated web-code lines successfully.

This does not prove production readiness for a real 5000-line application. Real applications add complex relations, permissions, workflows, custom UI, API contracts, validation, and deployment constraints. Those areas should be benchmarked separately.

## Claim Boundary

Safe claim:

> The current BlackLang compiler can successfully generate a large synthetic CRUD/admin-style web project from a `.black` source file above 5000 non-empty lines.

Unsafe claim:

> Every real 5000-line BlackLang production application will build correctly without additional scale work.

