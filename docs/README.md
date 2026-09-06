# BlackLang Documentation Site Plan

This folder will contain the official BlackLang learning and reference site.

The site should work for both humans and AI coding agents.

The first published static site source lives in `../website/`.

## Purpose

The documentation site is the public memory of the language.

It should explain:

- How to install BlackLang
- How `.black` files work
- How target platform and generated stack intent is declared
- How each keyword behaves
- How CLI commands behave
- How JSON diagnostics work
- How `.blackthm` theme profile files work
- How compact UI profile slots are read
- How standard UI mode groups work
- How inline UI intent is written near fields, forms, tables, and action buttons
- How page view order moves generated form, table, and detail sections
- How locked UI profiles enforce append-only slot changes
- How i18n locale and field label translations work
- How security CORS policy is declared without hardcoding deployment origins
- How Docker deployment intent generates production runtime files
- How BlackLang compares with common web languages without overstating current capability
- How measured scale benchmarks differ from positioning notes
- How AI agents should respect the current capability boundary
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
├── target.md
├── i18n.md
├── inline-ui.md
├── view.md
├── ai-agent-contract.md
├── deployment.md
├── security-cors.md
├── release-artifacts.md
├── npm-wrapper.md
├── ai-agents/
├── examples/
├── benchmarks/
│   └── language-comparison-v0.2.md
│   └── scale-5327-loc.md
├── llms.txt
└── llms-full.txt
```

## First Pages

- Learn: What is BlackLang?
- Learn: Quick start
- Reference: Syntax
- Reference: app
- Reference: target
- Reference: entity
- Reference: page
- Reference: table
- Reference: form
- Reference: view
- CLI: parse
- CLI: format
- CLI: lint
- CLI: validate
- CLI: docs all JSON export
- CLI: explain keyword JSON
- CLI: theme inspect JSON
- CLI: install paths
- CLI: GitHub publish policy
- CLI: release artifact layout
- CLI: npm wrapper plan
- AI Agents: Codex guide
- AI Agents: Capability boundary and official workflow
- Errors: Error code reference
- Reference: Theme profile
- Reference: UI profile rules
- Reference: UI modes
- Reference: Inline UI intent
- Reference: I18n

## Documentation Rule

Every language or CLI change should update the relevant documentation in the same change.
