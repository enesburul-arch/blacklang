# BlackLang Security CORS

BlackLang keeps security policy in source while keeping deployment-specific values in environment variables.

```black
security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Rules:

- Use one top-level `security` block.
- Use one `cors` block inside `security`.
- `origins` must reference an environment variable.
- The environment variable contains comma-separated allowed browser origins.
- `credentials true` enables cookie-auth compatible CORS headers.
- Do not hardcode production domains, API keys, passwords, tokens, or connection strings directly in `.black` source.

Generated Express behavior:

- Reads `process.env.CORS_ORIGINS`.
- Allows requests without an `Origin` header.
- Rejects browser origins that are not listed.
- Sets `Access-Control-Allow-Origin` to the matched request origin.
- Sets `Access-Control-Allow-Headers` for JSON and CSRF token requests.
- Answers `OPTIONS` preflight requests with `204`.
