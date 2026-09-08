# Protected Source

BlackLang `.black` files are high-value source assets. A compact source file can represent a much larger generated web or API application, so production workflows should avoid shipping source files when generated artifacts are enough.

The protected source MVP uses encrypted `.black.enc` files.

```bash
black security encrypted-source --json
BLACKLANG_SOURCE_KEY=local-key black security encrypt app.black --out app.black.enc --json
BLACKLANG_SOURCE_KEY=local-key black security decrypt app.black.enc --stdout
```

The encrypted file starts with a small plaintext header and stores the encrypted source body below it:

```text
BLACKLANG-ENC v1
alg AES-256-GCM
kdf SHA256-ENV
keyEnv BLACKLANG_SOURCE_KEY
nonce <base64>
---
<base64 ciphertext>
```

Key material comes only from the environment variable named by `keyEnv` or `--key-env`. The key value must not be written into `.black`, `.black.enc`, generated files, docs, or git history.

`black security decrypt` requires `--stdout`. It does not write decrypted source files. If plaintext editing is needed, decrypt inside a trusted developer or CI workspace, edit and format the `.black` file, then re-encrypt it.

These commands can read `.black.enc` source in memory when the key environment variable is set:

- `black parse`
- `black lint`
- `black validate`
- `black inspect`
- `black benchmark`
- `black security scan`
- `black build`

`black format` does not rewrite `.black.enc` files and returns `UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE`.

Production packages exclude both `.black` and `.black.enc` source files. Generated web output should be packaged and deployed instead of protected source.

AI agents should use JSON output and diagnostic codes:

```bash
black security encrypted-source --json
black security scan app.black.enc --json
black validate app.black.enc --json
black build app.black.enc --out generated --json
```

Common diagnostics:

- `MISSING_ENCRYPTION_KEY`: set the key environment variable named in the header or `--key-env`.
- `INVALID_ENCRYPTED_SOURCE`: recreate the encrypted file with `black security encrypt`.
- `DECRYPTION_FAILED`: check that the same key value is used.
- `MISSING_DECRYPT_STDOUT`: add `--stdout`; decrypt never writes plaintext by default.
- `UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE`: format plaintext source, then re-encrypt.
