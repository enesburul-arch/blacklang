# BlackLang Documentation Site Plan

This folder will contain the official BlackLang learning and reference site.

The site should work for both humans and AI coding agents.

The first published static site source lives in `../website/`.

## Purpose

The documentation site is the public memory of the language.

It should explain:

- How to install BlackLang
- How `.black` files work
- How each keyword behaves
- How CLI commands behave
- How JSON diagnostics work
- How AI agents should use the language

## Planned Structure

```text
docs/
├── learn/
├── reference/
├── cli/
├── errors/
├── install.md
├── github-publish.md
├── release-artifacts.md
├── npm-wrapper.md
├── ai-agents/
├── examples/
├── benchmarks/
├── llms.txt
└── llms-full.txt
```

## First Pages

- Learn: What is BlackLang?
- Learn: Quick start
- Reference: Syntax
- Reference: app
- Reference: entity
- Reference: page
- Reference: table
- Reference: form
- CLI: parse
- CLI: format
- CLI: lint
- CLI: validate
- CLI: docs all JSON export
- CLI: explain keyword JSON
- CLI: install paths
- CLI: GitHub publish policy
- CLI: release artifact layout
- CLI: npm wrapper plan
- AI Agents: Codex guide
- Errors: Error code reference

## Documentation Rule

Every language or CLI change should update the relevant documentation in the same change.
