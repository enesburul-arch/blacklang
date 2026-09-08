# Background Query Jobs

BlackLang v0.2 supports deterministic generated background jobs for read-only query work.

```black
query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
}
```

Generated output:

- `jobs/manifest.json` lists declared jobs, schedule metadata, query, source, limit, and safe execution notes
- `src/worker.ts` runs the declared query window from Prisma and logs compact JSON metadata
- `package.json` adds `jobs:run` for one execution and `jobs:loop` for a long-running worker process
- `openapi.json` includes root `x-blacklang-jobs` metadata for AI and tooling discovery

The generated worker applies the referenced query's `where`, `sort`, and `limit` rules and selects only record IDs. The runtime log includes the job name, schedule, query, source, count, limit, and timestamp.

The MVP accepts only:

```black
schedule every <integer> minutes|hours|days
run query <QueryName>
```

Supported ranges are:

- `1..1440 minutes`
- `1..168 hours`
- `1..365 days`

Jobs run in internal worker scope and do not expose a public HTTP endpoint. Queue providers, retries, delayed queues, mutating jobs, external calls, and provider schedulers are future adapter work.

Useful commands:

```bash
black docs job --json
black explain job --json
black inspect examples/warehouse/app.black --affected LowStockMonitor --json
black inspect examples/warehouse/app.black --affected LowStockProducts --json
black build examples/warehouse/app.black --out generated --json
cd generated && npm run jobs:run
```
