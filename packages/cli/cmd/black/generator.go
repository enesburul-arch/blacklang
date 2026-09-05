package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedComparisonOperators = setOf(
	"==",
	"!=",
	"<",
	"<=",
	">",
	">=",
)

func BuildWeb(program Program, outDir string) ([]GeneratedFile, []Diagnostic) {
	generator := webGenerator{
		program: program,
		outDir:  outDir,
		files:   []GeneratedFile{},
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, []Diagnostic{{
			Code:       "BUILD_OUTPUT_ERROR",
			Message:    err.Error(),
			Suggestion: "Choose a writable output directory with `--out <dir>`.",
		}}
	}

	generator.write("README.md", "documentation", generator.readme())
	generator.write(".env.example", "environment-example", generator.envExample())
	generator.write("package.json", "package", generator.packageJSON())
	generator.write("index.html", "html-entry", generator.indexHTML())
	generator.write("tsconfig.json", "typescript-config", generator.tsconfig())
	generator.write("vite.config.ts", "vite-config", generator.viteConfig())
	generator.write("prisma.config.ts", "prisma-config", generator.prismaConfig())
	generator.write("openapi.json", "api-spec", generator.openapiJSON())
	generator.write(filepath.Join("prisma", "schema.prisma"), "database-schema", generator.prismaSchema())
	generator.write(filepath.Join("src", "main.tsx"), "react-entry", generator.mainTSX())
	generator.write(filepath.Join("src", "App.tsx"), "react-app", generator.appTSX())
	generator.write(filepath.Join("src", "db.ts"), "database-client", generator.dbTS())
	generator.write(filepath.Join("src", "setup-db.ts"), "database-setup", generator.setupDBTS())
	generator.write(filepath.Join("src", "server.ts"), "api-server", generator.serverTS())
	generator.write(filepath.Join("src", "styles.css"), "stylesheet", generator.stylesCSS())
	generator.write(filepath.Join("src", "vite-env.d.ts"), "type-declarations", generator.viteEnv())
	generator.write(filepath.Join("src", "types.ts"), "types", generator.types())
	if program.Auth != nil {
		generator.write(filepath.Join("src", "auth", "AuthPage.tsx"), "auth-page", generator.authPageTSX())
		generator.write(filepath.Join("src", "routes", "auth.ts"), "auth-route", generator.authRouteTS())
		if len(program.Roles) > 0 {
			generator.write(filepath.Join("src", "auth", "UsersPage.tsx"), "auth-users-page", generator.authUsersPageTSX())
			generator.write(filepath.Join("src", "auth", "AuditPage.tsx"), "auth-audit-page", generator.authAuditPageTSX())
		}
	}

	for _, component := range program.Components {
		generator.write(filepath.Join("src", "components", component.Name+".tsx"), "react-component", generator.componentTSX(component))
	}

	for _, page := range program.Pages {
		entity, ok := generator.findEntity(page.Source)
		if !ok {
			continue
		}
		name := strings.ToLower(page.Source)
		generator.write(filepath.Join("src", "api", name+".ts"), "api-client", generator.apiClient(page, entity))
		generator.write(filepath.Join("src", "routes", name+".ts"), "api-route", generator.route(page, entity))
		generator.write(filepath.Join("src", "validation", name+".ts"), "validation", generator.validation(entity))
		generator.write(filepath.Join("src", "pages", page.Name+"Page.tsx"), "react-page", generator.page(page, entity))
	}

	return generator.files, generator.diagnostics
}

type webGenerator struct {
	program     Program
	outDir      string
	files       []GeneratedFile
	diagnostics []Diagnostic
}

type incomingRelation struct {
	entity EntityDecl
	field  FieldDecl
}

func (g *webGenerator) write(relativePath string, kind string, content string) {
	fullPath := filepath.Join(g.outDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		g.addError(fullPath, err)
		return
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		g.addError(fullPath, err)
		return
	}
	g.files = append(g.files, GeneratedFile{
		Path: fullPath,
		Kind: kind,
	})
}

func (g *webGenerator) addError(path string, err error) {
	g.diagnostics = append(g.diagnostics, Diagnostic{
		File:       path,
		Code:       "BUILD_WRITE_ERROR",
		Message:    err.Error(),
		Suggestion: "Check output directory permissions.",
	})
}

func (g *webGenerator) findEntity(name string) (EntityDecl, bool) {
	for _, entity := range g.program.Entities {
		if entity.Name == name {
			return entity, true
		}
	}
	return EntityDecl{}, false
}

func (g *webGenerator) isRelationField(field FieldDecl) bool {
	for _, entity := range g.program.Entities {
		if entity.Name == field.Type {
			return true
		}
	}
	return false
}

func (g *webGenerator) sqliteColumnName(field FieldDecl) string {
	if g.isRelationField(field) {
		return relationIDFieldName(field)
	}
	return field.Name
}

func (g *webGenerator) sqliteFieldType(field FieldDecl) string {
	if g.isRelationField(field) {
		return "TEXT"
	}
	return sqliteType(field.Type)
}

func (g *webGenerator) incomingRelations(entityName string) []incomingRelation {
	relations := []incomingRelation{}
	for _, sourceEntity := range g.program.Entities {
		for _, field := range sourceEntity.Fields {
			if field.Type == entityName {
				relations = append(relations, incomingRelation{
					entity: sourceEntity,
					field:  field,
				})
			}
		}
	}
	return relations
}

func (g *webGenerator) relationFields(entity EntityDecl) []FieldDecl {
	fields := []FieldDecl{}
	for _, field := range entity.Fields {
		if g.isRelationField(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (g *webGenerator) hasRelationFields(entity EntityDecl) bool {
	return len(g.relationFields(entity)) > 0
}

func (g *webGenerator) prismaIncludeLine(entity EntityDecl, indent string, trailingComma ...bool) string {
	withComma := true
	if len(trailingComma) > 0 {
		withComma = trailingComma[0]
	}
	relations := g.relationFields(entity)
	parts := []string{}
	for _, field := range relations {
		parts = append(parts, fmt.Sprintf("%s: true", field.Name))
	}
	line := fmt.Sprintf("%sinclude: { %s }", indent, strings.Join(parts, ", "))
	if withComma {
		line += ","
	}
	return line + "\n"
}

func (g *webGenerator) readme() string {
	return fmt.Sprintf(`# %s

This folder was generated by BlackLang.

Do not edit generated files manually. Change the BlackLang source file and run black build again.

## Database

Copy .env.example to .env for local development.

~~~bash
npm run db:generate
npm run db:validate
npm run db:setup
npm run db:push
npm run api:dev
~~~

## Generated Summary

- App: %s
- Entities: %d
- Pages: %d
`, g.program.App.Name, g.program.App.Name, len(g.program.Entities), len(g.program.Pages))
}

func (g *webGenerator) envExample() string {
	return `DATABASE_URL="file:./dev.db"
`
}

func (g *webGenerator) packageJSON() string {
	name := strings.ToLower(g.program.App.Name)
	return fmt.Sprintf(`{
  "name": "%s-generated",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "api:dev": "tsx src/server.ts",
    "build": "node -e \"require('fs').rmSync('dist', { recursive: true, force: true })\" && npm run db:generate && tsc && vite build",
    "db:generate": "prisma generate",
    "db:validate": "prisma validate",
    "db:setup": "tsx src/setup-db.ts",
    "db:push": "npm run db:setup",
    "db:push:native": "prisma db push"
  },
  "dependencies": {
    "@prisma/adapter-better-sqlite3": "7.10.0",
    "@prisma/client": "7.10.0",
    "better-sqlite3": "latest",
    "dotenv": "latest",
    "react": "latest",
    "react-dom": "latest",
    "express": "latest"
  },
  "devDependencies": {
    "@types/express": "latest",
    "@types/better-sqlite3": "latest",
    "@types/node": "latest",
    "@types/react": "latest",
    "@types/react-dom": "latest",
    "@vitejs/plugin-react": "latest",
    "prisma": "7.10.0",
    "tsx": "latest",
    "typescript": "latest",
    "vite": "latest"
  }
}
`, name)
}

func (g *webGenerator) prismaConfig() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { defineConfig } from "prisma/config";

export default defineConfig({
  schema: "prisma/schema.prisma",
  datasource: {
    url: process.env.DATABASE_URL ?? "file:./dev.db"
  }
});
`
}

func (g *webGenerator) viteConfig() string {
	return `// Generated by BlackLang. Do not edit manually.

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": "http://localhost:3001"
    }
  }
});
`
}

func (g *webGenerator) indexHTML() string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`, g.program.App.Name)
}

func (g *webGenerator) tsconfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["DOM", "DOM.Iterable", "ES2020"],
    "allowJs": false,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "include": ["src"]
}
`
}

func (g *webGenerator) mainTSX() string {
	return `// Generated by BlackLang. Do not edit manually.

import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
`
}

func (g *webGenerator) authPageTSX() string {
	emailPlaceholder := "you@example.com"
	namePlaceholder := "Full name"
	if g.program.Auth != nil {
		for _, field := range g.program.Auth.User.Fields {
			if field.Name == "email" {
				if value := modifierValue(field, "placeholder"); value != "" {
					emailPlaceholder = value
				}
			}
			if field.Name == "name" {
				if value := modifierValue(field, "placeholder"); value != "" {
					namePlaceholder = value
				}
			}
		}
	}

	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import { useState } from "react";
import type { FormEvent } from "react";

type AuthPageProps = {
  appName: string;
  onAuthenticated: (user: { id: string; name: string; email: string; role: string }) => void;
};

export function AuthPage({ appName, onAuthenticated }: AuthPageProps) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const input = {
      name: String(form.get("name") ?? ""),
      email: String(form.get("email") ?? ""),
      password: String(form.get("password") ?? "")
    };

    setSaving(true);
    setError(null);
    try {
      const response = await fetch("/api/auth/" + mode, {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input)
      });
      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Authentication failed" }));
        throw new Error(String(body.error ?? "Authentication failed"));
      }
      const body = await response.json();
      onAuthenticated(body.user);
    } catch (error) {
      setError(error instanceof Error ? error.message : "Authentication failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="auth-screen">
      <section className="auth-panel" aria-labelledby="auth-title">
        <div>
          <span className="eyebrow">{appName}</span>
          <h1 id="auth-title">{mode === "login" ? "Sign in" : "Create account"}</h1>
          <p className="muted">Generated from BlackLang auth intent with cookie session persistence.</p>
        </div>

        <div className="auth-tabs" role="tablist" aria-label="Authentication mode">
          <button className={mode === "login" ? "active" : "secondary"} type="button" onClick={() => setMode("login")}>Login</button>
          <button className={mode === "register" ? "active" : "secondary"} type="button" onClick={() => setMode("register")}>Register</button>
        </div>

        <form onSubmit={submit}>
          {error && <div className="error">{error}</div>}
          {mode === "register" && (
            <label>
              Name
              <input name="name" required placeholder=%q />
            </label>
          )}
          <label>
            Email
            <input name="email" required type="email" placeholder=%q />
          </label>
          <label>
            Password
            <input name="password" required minLength={8} type="password" placeholder="At least 8 characters" />
          </label>
          <button disabled={saving} type="submit">{saving ? "Please wait" : mode === "login" ? "Sign in" : "Create account"}</button>
        </form>
      </section>
    </main>
  );
}
`, namePlaceholder, emailPlaceholder)
}

func (g *webGenerator) authUsersPageTSX() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import { useEffect, useState } from "react";

type User = {
  id: string;
  name: string;
  email: string;
  role: string;
};

const roles = %s;

function csrfHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = document.cookie.split("; ").find((item) => item.startsWith("black_csrf="))?.split("=")[1] ?? "";
  if (token) headers["X-CSRF-Token"] = decodeURIComponent(token);
  return headers;
}

export function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    fetch("/api/auth/users", { credentials: "same-origin" })
      .then(async (response) => {
        if (!response.ok) {
          const body = await response.json().catch(() => ({ error: "Unable to load users" }));
          throw new Error(String(body.error ?? "Unable to load users"));
        }
        return response.json() as Promise<{ users: User[] }>;
      })
      .then((body) => {
        if (active) setUsers(body.users);
      })
      .catch((reason: unknown) => {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load users");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  async function updateRole(userId: string, role: string) {
    setSavingId(userId);
    setError(null);
    try {
      const response = await fetch("/api/auth/users/" + userId + "/role", {
        method: "PUT",
        credentials: "same-origin",
        headers: csrfHeaders(),
        body: JSON.stringify({ role })
      });
      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Unable to update role" }));
        throw new Error(String(body.error ?? "Unable to update role"));
      }
      const body = await response.json() as { user: User };
      setUsers((current) => current.map((user) => user.id === userId ? body.user : user));
    } catch (reason: unknown) {
      setError(reason instanceof Error ? reason.message : "Unable to update role");
    } finally {
      setSavingId(null);
    }
  }

  return (
    <main>
      <header>
        <h1>Users</h1>
        <span>Role management</span>
      </header>

      <section className="panel">
        <div className="toolbar">
          <div>
            <h2>Users</h2>
            <p className="muted">Generated from BlackLang role declarations.</p>
          </div>
        </div>
        {error && <div className="error" role="alert">{error}</div>}
        {loading && <div className="status">Loading users...</div>}
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Role</th>
            </tr>
          </thead>
          <tbody>
            {!loading && users.length === 0 && (
              <tr><td colSpan={3}>No users yet.</td></tr>
            )}
            {users.map((user) => (
              <tr key={user.id}>
                <td>{user.name}</td>
                <td>{user.email}</td>
                <td>
                  <select disabled={savingId === user.id} value={user.role} onChange={(event) => updateRole(user.id, event.target.value)}>
                    {roles.map((role) => (
                      <option key={role} value={role}>{role}</option>
                    ))}
                  </select>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}
`, tsStringArrayLiteral(g.roleNames()))
}

func (g *webGenerator) authAuditPageTSX() string {
	return `// Generated by BlackLang. Do not edit manually.

import { useEffect, useState } from "react";

type AuditLog = {
  id: string;
  actorUserId: string;
  actorRole: string;
  action: string;
  resource: string;
  resourceId: string;
  summary: string;
  createdAt: string;
};

export function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    fetch("/api/auth/audit", { credentials: "same-origin" })
      .then(async (response) => {
        if (!response.ok) {
          const body = await response.json().catch(() => ({ error: "Unable to load audit logs" }));
          throw new Error(String(body.error ?? "Unable to load audit logs"));
        }
        return response.json() as Promise<{ logs: AuditLog[] }>;
      })
      .then((body) => {
        if (active) setLogs(body.logs);
      })
      .catch((reason: unknown) => {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load audit logs");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  return (
    <main>
      <header>
        <h1>Audit</h1>
        <span>Security activity log</span>
      </header>

      <section className="panel">
        <div className="toolbar">
          <div>
            <h2>Audit Log</h2>
            <p className="muted">Generated from BlackLang auth, role, and action intent.</p>
          </div>
        </div>
        {error && <div className="error" role="alert">{error}</div>}
        {loading && <div className="status">Loading audit logs...</div>}
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Actor</th>
              <th>Action</th>
              <th>Resource</th>
              <th>Summary</th>
            </tr>
          </thead>
          <tbody>
            {!loading && logs.length === 0 && (
              <tr><td colSpan={5}>No audit logs yet.</td></tr>
            )}
            {logs.map((log) => (
              <tr key={log.id}>
                <td>{new Date(log.createdAt).toLocaleString()}</td>
                <td>{log.actorRole}</td>
                <td>{log.action}</td>
                <td>{log.resource}{log.resourceId ? " / " + log.resourceId : ""}</td>
                <td>{log.summary}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}
`
}

func (g *webGenerator) appTSX() string {
	if len(g.program.Pages) == 0 {
		return `// Generated by BlackLang. Do not edit manually.

export function App() {
  return <main>No pages generated.</main>;
}
`
	}

	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	if g.program.Auth != nil {
		builder.WriteString("import { useEffect, useState } from \"react\";\n")
	} else {
		builder.WriteString("import { useState } from \"react\";\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("import { AuthPage } from \"./auth/AuthPage\";\n")
		if len(g.program.Roles) > 0 {
			builder.WriteString("import { UsersPage } from \"./auth/UsersPage\";\n")
			builder.WriteString("import { AuditPage } from \"./auth/AuditPage\";\n")
		}
	}
	for _, page := range g.program.Pages {
		builder.WriteString(fmt.Sprintf("import { %sPage } from \"./pages/%sPage\";\n", page.Name, page.Name))
	}
	builder.WriteString("\n")
	if g.program.Auth != nil {
		builder.WriteString("type CurrentUser = { id: string; name: string; email: string; role: string };\n\n")
		builder.WriteString("type RolePermission = { effect: string; action: string; resource: string; fields: string[] };\n\n")
		builder.WriteString(fmt.Sprintf("const rolePermissions: Record<string, RolePermission[]> = %s;\n\n", g.rolePermissionsLiteral()))
		builder.WriteString("function permissionMatches(permission: RolePermission, action: string, resource: string, field?: string) {\n")
		builder.WriteString("  if (permission.action === \"all\") return true;\n")
		builder.WriteString("  const actionMatches = permission.action === action || permission.action === \"manage\";\n")
		builder.WriteString("  if (!actionMatches || permission.resource !== resource) return false;\n")
		builder.WriteString("  if (!field || permission.fields.length === 0) return true;\n")
		builder.WriteString("  return permission.fields.includes(field);\n")
		builder.WriteString("}\n\n")
	}

	navigationPages := g.navigationPages()
	builder.WriteString("const pages = [\n")
	for _, page := range navigationPages {
		builder.WriteString(fmt.Sprintf("  { name: %q, access: %s },\n", page.Name, tsStringArrayLiteral(page.Access)))
	}
	if g.program.Auth != nil && len(g.program.Roles) > 0 {
		builder.WriteString(fmt.Sprintf("  { name: \"Users\", access: %s },\n", tsStringArrayLiteral([]string{g.defaultAuthRole()})))
		builder.WriteString(fmt.Sprintf("  { name: \"Audit\", access: %s },\n", tsStringArrayLiteral([]string{g.defaultAuthRole()})))
	}
	builder.WriteString("];\n\n")
	builder.WriteString("export function App() {\n")
	builder.WriteString("  const [activePage, setActivePage] = useState(pages[0].name);\n")
	builder.WriteString("  const [navOpen, setNavOpen] = useState(false);\n")
	if g.program.Auth != nil {
		builder.WriteString("  const [authenticated, setAuthenticated] = useState<boolean | null>(null);\n")
		builder.WriteString("  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("  const visiblePages = currentUser ? pages.filter((item) => item.access.length === 0 || item.access.includes(\"authenticated\") || item.access.includes(currentUser.role)) : pages;\n")
	} else {
		builder.WriteString("  const visiblePages = pages;\n")
	}
	builder.WriteString("  const page = visiblePages.find((item) => item.name === activePage) ?? visiblePages[0];\n\n")
	if g.program.Auth != nil {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    fetch(\"/api/auth/me\", { credentials: \"same-origin\" })\n")
		builder.WriteString("      .then(async (response) => {\n")
		builder.WriteString("        if (!response.ok) {\n")
		builder.WriteString("          setCurrentUser(null);\n")
		builder.WriteString("          setAuthenticated(false);\n")
		builder.WriteString("          return;\n")
		builder.WriteString("        }\n")
		builder.WriteString("        const body = await response.json();\n")
		builder.WriteString("        setCurrentUser(body.user);\n")
		builder.WriteString("        setAuthenticated(true);\n")
		builder.WriteString("      })\n")
		builder.WriteString("      .catch(() => {\n")
		builder.WriteString("        setCurrentUser(null);\n")
		builder.WriteString("        setAuthenticated(false);\n")
		builder.WriteString("      });\n")
		builder.WriteString("  }, []);\n\n")
		builder.WriteString("  async function logout() {\n")
		builder.WriteString("    const csrfToken = document.cookie.split(\"; \").find((item) => item.startsWith(\"black_csrf=\"))?.split(\"=\")[1] ?? \"\";\n")
		builder.WriteString("    await fetch(\"/api/auth/logout\", { method: \"POST\", credentials: \"same-origin\", headers: csrfToken ? { \"X-CSRF-Token\": decodeURIComponent(csrfToken) } : {} });\n")
		builder.WriteString("    setCurrentUser(null);\n")
		builder.WriteString("    setAuthenticated(false);\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (authenticated === null) {\n")
		builder.WriteString("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">Checking session...</p></section></main>;\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (!authenticated) {\n")
		builder.WriteString("    return <AuthPage appName=\"")
		builder.WriteString(g.program.App.Name)
		builder.WriteString("\" onAuthenticated={(user) => { setCurrentUser(user); setAuthenticated(true); }} />;\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (!page) {\n")
		builder.WriteString("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">No pages are available for this role.</p><button className=\"secondary\" type=\"button\" onClick={logout}>Logout</button></section></main>;\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function navigateTo(pageName: string) {\n")
	builder.WriteString("    setActivePage(pageName);\n")
	builder.WriteString("    setNavOpen(false);\n")
	builder.WriteString("  }\n\n")
	if g.program.Auth != nil {
		builder.WriteString("  function canAccessAction(action: string, resource: string) {\n")
		builder.WriteString("    if (!currentUser) return false;\n")
		builder.WriteString("    const permissions = rolePermissions[currentUser.role] ?? [];\n")
		builder.WriteString("    const denied = permissions.some((permission) => permission.effect === \"deny\" && permission.fields.length === 0 && permissionMatches(permission, action, resource));\n")
		builder.WriteString("    if (denied) return false;\n")
		builder.WriteString("    return permissions.some((permission) => permission.effect === \"allow\" && permissionMatches(permission, action, resource));\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  function canAccessField(action: string, resource: string, field: string) {\n")
		builder.WriteString("    if (!currentUser) return false;\n")
		builder.WriteString("    const permissions = rolePermissions[currentUser.role] ?? [];\n")
		builder.WriteString("    const denied = permissions.some((permission) => permission.effect === \"deny\" && permissionMatches(permission, action, resource, field));\n")
		builder.WriteString("    if (denied) return false;\n")
		builder.WriteString("    return permissions.some((permission) => permission.effect === \"allow\" && permissionMatches(permission, action, resource, field));\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  function pagePermissions(resource: string, fields: string[]) {\n")
		builder.WriteString("    const fieldAccess = Object.fromEntries(fields.map((field) => [field, canAccessField(\"read\", resource, field)]));\n")
		builder.WriteString("    const writeAccess = Object.fromEntries(fields.map((field) => [field, canAccessField(\"update\", resource, field)]));\n")
		builder.WriteString("    return {\n")
		builder.WriteString("      read: canAccessAction(\"read\", resource),\n")
		builder.WriteString("      create: canAccessAction(\"create\", resource),\n")
		builder.WriteString("      update: canAccessAction(\"update\", resource),\n")
		builder.WriteString("      delete: canAccessAction(\"delete\", resource),\n")
		builder.WriteString("      fields: fieldAccess,\n")
		builder.WriteString("      writeFields: writeAccess\n")
		builder.WriteString("    };\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function renderPage() {\n")
	builder.WriteString("    switch (page.name) {\n")
	for _, page := range g.program.Pages {
		builder.WriteString(fmt.Sprintf("      case %q:\n", page.Name))
		if g.program.Auth != nil {
			if entity, ok := g.findEntity(page.Source); ok {
				builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo} permissions={pagePermissions(%q, %s)} />;\n", page.Name, page.Source, tsStringArrayLiteral(fieldNames(entity))))
			}
		} else {
			builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo} />;\n", page.Name))
		}
	}
	if g.program.Auth != nil && len(g.program.Roles) > 0 {
		builder.WriteString("      case \"Users\":\n")
		builder.WriteString("        return <UsersPage />;\n")
		builder.WriteString("      case \"Audit\":\n")
		builder.WriteString("        return <AuditPage />;\n")
	}
	builder.WriteString("      default:\n")
	if g.program.Auth != nil {
		if entity, ok := g.findEntity(g.program.Pages[0].Source); ok {
			builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo} permissions={pagePermissions(%q, %s)} />;\n", g.program.Pages[0].Name, g.program.Pages[0].Source, tsStringArrayLiteral(fieldNames(entity))))
		}
	} else {
		builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo} />;\n", g.program.Pages[0].Name))
	}
	builder.WriteString("    }\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  return (\n")
	builder.WriteString("    <div className={navOpen ? \"app-shell nav-open\" : \"app-shell\"}>\n")
	builder.WriteString("      {navOpen && <button className=\"nav-backdrop\" type=\"button\" aria-label=\"Close navigation\" onClick={() => setNavOpen(false)} />}\n")
	builder.WriteString("      <aside className=\"app-sidebar\">\n")
	builder.WriteString(fmt.Sprintf("        <div className=\"app-brand\">%s</div>\n", g.program.App.Name))
	builder.WriteString("        <nav className=\"app-nav\" aria-label=\"Primary navigation\">\n")
	builder.WriteString("          {visiblePages.map((item) => (\n")
	builder.WriteString("            <button key={item.name} className={item.name === page.name ? \"active\" : \"secondary\"} type=\"button\" onClick={() => navigateTo(item.name)}>{item.name}</button>\n")
	builder.WriteString("          ))}\n")
	builder.WriteString("        </nav>\n")
	builder.WriteString("      </aside>\n")
	builder.WriteString("      <div className=\"app-workspace\">\n")
	builder.WriteString("        <div className=\"app-topbar\">\n")
	builder.WriteString("          <button className=\"menu-button secondary\" type=\"button\" aria-expanded={navOpen} onClick={() => setNavOpen(true)}>Menu</button>\n")
	builder.WriteString("          <div>\n")
	builder.WriteString(fmt.Sprintf("            <span className=\"breadcrumb\">%s / {page.name}</span>\n", g.program.App.Name))
	builder.WriteString("            <h1>{page.name}</h1>\n")
	builder.WriteString("          </div>\n")
	if g.program.Auth != nil {
		builder.WriteString("          <div className=\"user-menu\">\n")
		builder.WriteString("            {currentUser && <span className=\"role-badge\">{currentUser.name} / {currentUser.role}</span>}\n")
		builder.WriteString("          <button className=\"secondary\" type=\"button\" onClick={logout}>Logout</button>\n")
		builder.WriteString("          </div>\n")
	}
	builder.WriteString("        </div>\n")
	builder.WriteString("        <div className=\"app-content\">{renderPage()}</div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </div>\n")
	builder.WriteString("  );\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) serverTS() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import express from \"express\";\n")
	builder.WriteString("import path from \"node:path\";\n")
	if g.program.Auth != nil {
		builder.WriteString("import { authRouter, requireAuth, requireCsrf } from \"./routes/auth\";\n")
	}
	for _, page := range g.program.Pages {
		lower := strings.ToLower(page.Source)
		builder.WriteString(fmt.Sprintf("import { %sRouter } from \"./routes/%s\";\n", lower, lower))
	}
	builder.WriteString("\n")
	builder.WriteString("const app = express();\n")
	builder.WriteString("const port = 3001;\n\n")
	builder.WriteString("const rateWindowMs = 60_000;\n")
	builder.WriteString("const rateLimit = 120;\n")
	builder.WriteString("const requestCounts = new Map<string, { count: number; resetAt: number }>();\n\n")
	builder.WriteString("app.disable(\"x-powered-by\");\n")
	builder.WriteString("app.use((_req, res, next) => {\n")
	builder.WriteString("  res.setHeader(\"X-Content-Type-Options\", \"nosniff\");\n")
	builder.WriteString("  res.setHeader(\"X-Frame-Options\", \"DENY\");\n")
	builder.WriteString("  res.setHeader(\"Referrer-Policy\", \"no-referrer\");\n")
	builder.WriteString("  res.setHeader(\"Permissions-Policy\", \"geolocation=(), microphone=(), camera=()\");\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	builder.WriteString("app.use((req, res, next) => {\n")
	builder.WriteString("  const now = Date.now();\n")
	builder.WriteString("  const key = req.ip ?? \"unknown\";\n")
	builder.WriteString("  const current = requestCounts.get(key);\n")
	builder.WriteString("  if (!current || current.resetAt <= now) {\n")
	builder.WriteString("    requestCounts.set(key, { count: 1, resetAt: now + rateWindowMs });\n")
	builder.WriteString("    next();\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString("  if (current.count >= rateLimit) {\n")
	builder.WriteString("    res.status(429).json({ error: \"Too many requests\" });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString("  current.count += 1;\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	builder.WriteString("app.use(express.json({ limit: \"100kb\" }));\n")
	builder.WriteString("app.get(\"/openapi.json\", (_req, res) => {\n")
	builder.WriteString("  res.sendFile(path.join(process.cwd(), \"openapi.json\"));\n")
	builder.WriteString("});\n")
	if g.program.Auth != nil {
		builder.WriteString("app.use(\"/api/auth\", authRouter);\n")
		builder.WriteString("app.use(\"/api\", requireAuth);\n")
		builder.WriteString("app.use(\"/api\", requireCsrf);\n")
	}
	for _, page := range g.program.Pages {
		lower := strings.ToLower(page.Source)
		builder.WriteString(fmt.Sprintf("app.use(\"/api\", %sRouter);\n", lower))
	}
	builder.WriteString("\n")
	builder.WriteString("app.listen(port, () => {\n")
	builder.WriteString("  console.log(`BlackLang API server running on http://localhost:${port}`);\n")
	builder.WriteString("});\n")
	return builder.String()
}

func (g *webGenerator) authRouteTS() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import crypto from "node:crypto";
import express from "express";
import Database from "better-sqlite3";

const databaseUrl = process.env.DATABASE_URL ?? "file:./dev.db";
const filePath = databaseUrl.startsWith("file:") ? databaseUrl.slice(5) : databaseUrl;
const db = new Database(filePath);
const sessionCookie = "black_session";
const csrfCookie = "black_csrf";
const defaultRole = %q;
const allowedRoles = %s;

type RolePermission = {
  effect: "allow" | "deny";
  action: string;
  resource: string;
  fields: string[];
};

const rolePermissions: Record<string, RolePermission[]> = %s;

type UserRow = {
  id: string;
  name: string;
  email: string;
  role: string;
  passwordHash: string;
};

function hashPassword(password: string, salt = crypto.randomBytes(16).toString("hex")) {
  const hash = crypto.pbkdf2Sync(password, salt, 120_000, 32, "sha256").toString("hex");
  return salt + ":" + hash;
}

function verifyPassword(password: string, stored: string) {
  const [salt, originalHash] = stored.split(":");
  if (!salt || !originalHash) return false;
  const nextHash = hashPassword(password, salt).split(":")[1];
  return crypto.timingSafeEqual(Buffer.from(originalHash, "hex"), Buffer.from(nextHash, "hex"));
}

function readCookie(cookieHeader: string | undefined, name: string) {
  if (!cookieHeader) return "";
  const cookies = cookieHeader.split(";").map((item) => item.trim());
  const cookie = cookies.find((item) => item.startsWith(name + "="));
  return cookie ? decodeURIComponent(cookie.slice(name.length + 1)) : "";
}

function readSessionToken(cookieHeader: string | undefined) {
  return readCookie(cookieHeader, sessionCookie);
}

function setAuthCookies(res: express.Response, token: string, csrfToken: string) {
  res.cookie(sessionCookie, token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
  res.cookie(csrfCookie, csrfToken, {
    httpOnly: false,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
}

function publicUser(user: UserRow) {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    role: user.role
  };
}

export function requireAuth(req: express.Request, res: express.Response, next: express.NextFunction) {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  const user = db.prepare("SELECT u.id, u.name, u.email, u.role, u.passwordHash FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')").get(token) as UserRow | undefined;
  if (!user) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  (req as any).blackUser = publicUser(user);
  next();
}

export function requireCsrf(req: express.Request, res: express.Response, next: express.NextFunction) {
  if (["GET", "HEAD", "OPTIONS"].includes(req.method)) {
    next();
    return;
  }

  const headerToken = String(req.headers["x-csrf-token"] ?? "");
  const cookieToken = readCookie(req.headers.cookie, csrfCookie);
  if (headerToken && cookieToken && headerToken === cookieToken) {
    next();
    return;
  }

  res.status(403).json({ error: "Invalid CSRF token" });
}

export function requirePageAccess(allowedRoles: string[]) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    if (allowedRoles.includes("authenticated") || allowedRoles.includes(user.role)) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

function permissionMatches(permission: RolePermission, action: string, resource: string, field?: string) {
  if (permission.action === "all") return true;
  const actionMatches = permission.action === action || permission.action === "manage";
  if (!actionMatches || permission.resource !== resource) return false;
  if (!field || permission.fields.length === 0) return true;
  return permission.fields.includes(field);
}

export function canAccessAction(role: string, action: string, resource: string) {
  const permissions = rolePermissions[role] ?? [];
  const denied = permissions.some((permission) => permission.effect === "deny" && permission.fields.length === 0 && permissionMatches(permission, action, resource));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource));
}

export function canAccessField(role: string, action: string, resource: string, field: string) {
  const permissions = rolePermissions[role] ?? [];
  const denied = permissions.some((permission) => permission.effect === "deny" && permissionMatches(permission, action, resource, field));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource, field));
}

export function requirePermission(action: string, resource: string) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    if (canAccessAction(user.role, action, resource)) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

export function filterWritableFields(role: string, action: string, resource: string, input: Record<string, unknown>) {
  const output: Record<string, unknown> = {};
  for (const [field, value] of Object.entries(input)) {
    if (canAccessField(role, action, resource, field)) {
      output[field] = value;
    }
  }
  return output;
}

export function writeAuditLog(actor: ReturnType<typeof publicUser> | undefined, action: string, resource: string, resourceId: string, summary = "") {
  if (!actor) return;
  db.prepare("INSERT INTO \"BlackAuditLog\" (id, actorUserId, actorRole, action, resource, resourceId, summary) VALUES (?, ?, ?, ?, ?, ?, ?)").run(
    crypto.randomUUID(),
    actor.id,
    actor.role,
    action,
    resource,
    resourceId,
    summary
  );
}

export const authRouter = express.Router();

authRouter.post("/register", (req, res) => {
  const name = String(req.body?.name ?? "").trim();
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");

  if (!name || !email || password.length < 8) {
    res.status(400).json({ error: "Name, email, and an 8 character password are required" });
    return;
  }

  const existing = db.prepare("SELECT id FROM \"BlackUser\" WHERE email = ?").get(email);
  if (existing) {
    res.status(409).json({ error: "Email is already registered" });
    return;
  }

  const user: UserRow = {
    id: crypto.randomUUID(),
    name,
    email,
    role: defaultRole,
    passwordHash: hashPassword(password)
  };
  db.prepare("INSERT INTO \"BlackUser\" (id, name, email, role, passwordHash) VALUES (?, ?, ?, ?, ?)").run(user.id, user.name, user.email, user.role, user.passwordHash);

  const token = crypto.randomUUID();
  const csrfToken = crypto.randomUUID();
  db.prepare("INSERT INTO \"BlackSession\" (id, userId, expiresAt) VALUES (?, ?, datetime('now', '+7 days'))").run(token, user.id);
  writeAuditLog(publicUser(user), "register", "BlackUser", user.id, "User registered");
  setAuthCookies(res, token, csrfToken);
  res.status(201).json({ user: publicUser(user) });
});

authRouter.post("/login", (req, res) => {
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");
  const user = db.prepare("SELECT id, name, email, role, passwordHash FROM \"BlackUser\" WHERE email = ?").get(email) as UserRow | undefined;

  if (!user || !verifyPassword(password, user.passwordHash)) {
    res.status(401).json({ error: "Invalid email or password" });
    return;
  }

  const token = crypto.randomUUID();
  const csrfToken = crypto.randomUUID();
  db.prepare("INSERT INTO \"BlackSession\" (id, userId, expiresAt) VALUES (?, ?, datetime('now', '+7 days'))").run(token, user.id);
  setAuthCookies(res, token, csrfToken);
  res.json({ user: publicUser(user) });
});

authRouter.post("/logout", requireCsrf, (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (token) {
    db.prepare("DELETE FROM \"BlackSession\" WHERE id = ?").run(token);
  }
  res.clearCookie(sessionCookie, { path: "/" });
  res.clearCookie(csrfCookie, { path: "/" });
  res.status(204).send();
});

authRouter.get("/me", (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  const user = db.prepare("SELECT u.id, u.name, u.email, u.role, u.passwordHash FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')").get(token) as UserRow | undefined;

  if (!user) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  res.json({ user: publicUser(user) });
});

authRouter.get("/users", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {
  const users = db.prepare("SELECT id, name, email, role, passwordHash FROM \"BlackUser\" ORDER BY createdAt DESC").all() as UserRow[];
  res.json({ users: users.map(publicUser) });
});

authRouter.get("/audit", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {
  const logs = db.prepare("SELECT id, actorUserId, actorRole, action, resource, resourceId, summary, createdAt FROM \"BlackAuditLog\" ORDER BY createdAt DESC LIMIT 200").all();
  res.json({ logs });
});

authRouter.put("/users/:id/role", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {
  const role = String(req.body?.role ?? "").trim();
  if (!allowedRoles.includes(role)) {
    res.status(400).json({ error: "Unsupported role" });
    return;
  }

  const result = db.prepare("UPDATE \"BlackUser\" SET role = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?").run(role, String(req.params.id));
  if (result.changes === 0) {
    res.status(404).json({ error: "User not found" });
    return;
  }

  const user = db.prepare("SELECT id, name, email, role, passwordHash FROM \"BlackUser\" WHERE id = ?").get(String(req.params.id)) as UserRow;
  writeAuditLog((req as any).blackUser, "role.update", "BlackUser", user.id, "Role changed to " + role);
  res.json({ user: publicUser(user) });
});
`, g.defaultAuthRole(), tsStringArrayLiteral(g.roleNames()), g.rolePermissionsLiteral())
}

func (g *webGenerator) dbTS() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { PrismaBetterSqlite3 } from "@prisma/adapter-better-sqlite3";
import { PrismaClient } from "./generated/prisma/client";

const adapter = new PrismaBetterSqlite3({
  url: process.env.DATABASE_URL ?? "file:./dev.db"
});

export const prisma = new PrismaClient({ adapter });
`
}

func (g *webGenerator) setupDBTS() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import \"dotenv/config\";\n")
	builder.WriteString("import Database from \"better-sqlite3\";\n\n")
	builder.WriteString("const databaseUrl = process.env.DATABASE_URL ?? \"file:./dev.db\";\n")
	builder.WriteString("const filePath = databaseUrl.startsWith(\"file:\") ? databaseUrl.slice(5) : databaseUrl;\n")
	builder.WriteString("const db = new Database(filePath);\n\n")
	if g.program.Auth != nil {
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackUser\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"name\" TEXT NOT NULL,\n")
		builder.WriteString("  \"email\" TEXT NOT NULL UNIQUE,\n")
		builder.WriteString(fmt.Sprintf("  \"role\" TEXT NOT NULL DEFAULT %q,\n", g.defaultAuthRole()))
		builder.WriteString("  \"passwordHash\" TEXT NOT NULL,\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
		builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
		builder.WriteString("const blackUserColumns = db.prepare(`PRAGMA table_info(\"BlackUser\")`).all() as Array<{ name: string }>;\n")
		builder.WriteString("if (!blackUserColumns.some((column) => column.name === \"role\")) {\n")
		builder.WriteString(fmt.Sprintf("  db.prepare(`ALTER TABLE \"BlackUser\" ADD COLUMN \"role\" TEXT NOT NULL DEFAULT %q`).run();\n", g.defaultAuthRole()))
		builder.WriteString("}\n\n")
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackSession\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"userId\" TEXT NOT NULL REFERENCES \"BlackUser\"(\"id\"),\n")
		builder.WriteString("  \"expiresAt\" DATETIME NOT NULL,\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackAuditLog\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"actorUserId\" TEXT NOT NULL,\n")
		builder.WriteString("  \"actorRole\" TEXT NOT NULL,\n")
		builder.WriteString("  \"action\" TEXT NOT NULL,\n")
		builder.WriteString("  \"resource\" TEXT NOT NULL,\n")
		builder.WriteString("  \"resourceId\" TEXT NOT NULL,\n")
		builder.WriteString("  \"summary\" TEXT NOT NULL DEFAULT '',\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
	}
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("db.exec(`\nCREATE TABLE IF NOT EXISTS \"%s\" (\n", entity.Name))
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		for index, field := range entity.Fields {
			line := fmt.Sprintf("  \"%s\" %s%s", g.sqliteColumnName(field), g.sqliteFieldType(field), sqliteRequired(field))
			if defaultValue := modifierValue(field, "default"); defaultValue != "" {
				line += fmt.Sprintf(" DEFAULT %s", sqliteDefaultValue(field, defaultValue))
			}
			if hasModifier(field, "unique") {
				line += " UNIQUE"
			}
			if g.isRelationField(field) {
				line += fmt.Sprintf(" REFERENCES \"%s\"(\"id\")", field.Type)
			}
			line += ","
			builder.WriteString(line + "\n")
			if index == len(entity.Fields)-1 {
				builder.WriteString("  \"archivedAt\" DATETIME,\n")
				builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
				builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
			}
		}
		if len(entity.Fields) == 0 {
			builder.WriteString("  \"archivedAt\" DATETIME,\n")
			builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
			builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		}
		builder.WriteString(");\n`);\n\n")
	}
	builder.WriteString("db.close();\n")
	builder.WriteString("console.log(`SQLite database ready at ${filePath}`);\n")
	return builder.String()
}

func (g *webGenerator) stylesCSS() string {
	return `:root {
  color: #172026;
  background: #f6f8fa;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  margin: 0;
}

.app-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  min-height: 100vh;
}

.app-sidebar {
  background: #172026;
  color: #ffffff;
  padding: 20px;
}

.nav-backdrop {
  display: none;
}

.app-brand {
  font-size: 18px;
  font-weight: 750;
  margin-bottom: 20px;
}

.app-nav {
  display: grid;
  gap: 8px;
}

.app-nav button {
  justify-content: flex-start;
  text-align: left;
  width: 100%;
}

.app-nav button.active {
  background: #ffffff;
  border-color: #ffffff;
  color: #172026;
}

.app-nav button.secondary {
  background: transparent;
  border-color: #3d4b59;
  color: #d8dee4;
}

.app-workspace {
  min-width: 0;
}

.app-topbar {
  align-items: center;
  background: #ffffff;
  border-bottom: 1px solid #d8dee4;
  display: flex;
  justify-content: space-between;
  min-height: 72px;
  padding: 16px 24px;
}

.menu-button {
  display: none;
}

.breadcrumb {
  color: #57606a;
  display: block;
  font-size: 13px;
  font-weight: 650;
  margin-bottom: 4px;
}

.app-content main {
  max-width: 1080px;
  margin: 0 auto;
  padding: 24px 20px 32px;
}

main > header {
  display: none;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

h1,
h2 {
  margin: 0;
}

.panel {
  background: #ffffff;
  border: 1px solid #d8dee4;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

label {
  display: grid;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
}

input,
select {
  border: 1px solid #c9d1d9;
  border-radius: 6px;
  font: inherit;
  padding: 9px 10px;
}

input[type="checkbox"] {
  height: 16px;
  width: 16px;
}

.inline-control {
  align-items: center;
  display: inline-flex;
  gap: 6px;
  grid-template-columns: none;
}

.field-note {
  color: #57606a;
  font-size: 12px;
  font-weight: 500;
}

.field-error {
  color: #a40e26;
  font-size: 12px;
  font-weight: 650;
}

.field-preview {
  display: flex;
  margin-top: 6px;
}

.relation-status {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.pagination {
  align-items: center;
  justify-content: flex-end;
}

.column-controls {
  align-items: center;
  flex-wrap: wrap;
}

.filter-controls {
  align-items: end;
  flex-wrap: wrap;
}

.filter-controls label {
  min-width: 180px;
}

.auth-screen {
  align-items: center;
  background: #f6f8fa;
  display: flex;
  min-height: 100vh;
  padding: 24px;
}

.auth-panel {
  background: #ffffff;
  border: 1px solid #d8dee4;
  border-radius: 8px;
  box-shadow: 0 16px 36px rgba(23, 32, 38, 0.08);
  display: grid;
  gap: 18px;
  margin: 0 auto;
  max-width: 420px;
  padding: 24px;
  width: 100%;
}

.auth-panel form {
  display: grid;
  gap: 14px;
}

.auth-tabs {
  display: grid;
  gap: 8px;
  grid-template-columns: 1fr 1fr;
}

.user-menu {
  align-items: center;
  display: flex;
  gap: 10px;
}

.role-badge {
  background: #eef2f6;
  border: 1px solid #d8dee4;
  border-radius: 999px;
  color: #172026;
  font-size: 13px;
  font-weight: 650;
  padding: 6px 10px;
}

.eyebrow {
  color: #57606a;
  display: block;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 6px;
  text-transform: uppercase;
}

button {
  border: 1px solid #1f6feb;
  border-radius: 6px;
  background: #1f6feb;
  color: #ffffff;
  cursor: pointer;
  font: inherit;
  font-weight: 650;
  padding: 9px 12px;
}

button.secondary {
  border-color: #c9d1d9;
  background: #ffffff;
  color: #172026;
}

button.danger {
  border-color: #cf222e;
  background: #cf222e;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.status,
.error {
  border-radius: 6px;
  margin-bottom: 12px;
  padding: 10px 12px;
}

.status {
  background: #e7f5ff;
  color: #0b4678;
}

.error {
  background: #fff1f3;
  color: #a40e26;
}

.muted {
  color: #57606a;
}

.detail-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  margin: 16px 0 0;
}

.detail-grid div {
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 10px;
}

.detail-grid dt {
  color: #57606a;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.detail-grid dd {
  margin: 0;
  overflow-wrap: anywhere;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  border-bottom: 1px solid #d8dee4;
  padding: 10px;
  text-align: left;
}

th {
  color: #57606a;
  font-size: 13px;
}

.select-cell {
  width: 40px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.black-component {
  border: 1px solid #d8dee4;
  border-radius: 999px;
  display: inline-flex;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  min-width: 34px;
  padding: 5px 8px;
}

.black-component-stock-badge-low {
  background: #fff1f3;
  border-color: #ffccd5;
  color: #a40e26;
}

.black-component-stock-badge-normal {
  background: #e6fcf5;
  border-color: #b7ebd8;
  color: #087f5b;
}

@media (max-width: 760px) {
  .app-shell {
    display: block;
  }

  .app-sidebar {
    bottom: 0;
    left: 0;
    max-width: 280px;
    position: fixed;
    top: 0;
    transform: translateX(-100%);
    transition: transform 160ms ease;
    width: 78vw;
    z-index: 20;
  }

  .app-shell.nav-open .app-sidebar {
    transform: translateX(0);
  }

  .nav-backdrop {
    background: rgba(23, 32, 38, 0.44);
    border: 0;
    bottom: 0;
    display: block;
    left: 0;
    padding: 0;
    position: fixed;
    right: 0;
    top: 0;
    z-index: 10;
  }

  .menu-button {
    display: inline-flex;
  }

  .app-topbar {
    gap: 12px;
    justify-content: flex-start;
    padding: 14px 16px;
  }

  .app-content main {
    padding: 16px;
  }
}
`
}

func (g *webGenerator) viteEnv() string {
	return `/// <reference types="vite/client" />
`
}

func (g *webGenerator) componentTSX(component ComponentDecl) string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString(fmt.Sprintf("type %sProps = {\n", component.Name))
	for _, input := range component.Inputs {
		builder.WriteString(fmt.Sprintf("  %s: %s;\n", input.Name, g.componentInputTSType(input)))
	}
	builder.WriteString("};\n\n")
	builder.WriteString(fmt.Sprintf("export function %s(props: %sProps) {\n", component.Name, component.Name))
	builder.WriteString("  const variant = resolveVariant(props);\n\n")
	builder.WriteString(fmt.Sprintf("  return <span className={\"black-component black-component-%s black-component-%s-\" + variant}>{String(%s)}</span>;\n", kebabCase(component.Name), kebabCase(component.Name), g.componentDisplayExpression(component)))
	builder.WriteString("}\n\n")
	builder.WriteString(fmt.Sprintf("function resolveVariant(props: %sProps) {\n", component.Name))
	for _, variant := range component.Variants {
		if condition := componentVariantConditionExpression(variant.Condition, component.Inputs); condition != "" {
			builder.WriteString(fmt.Sprintf("  if (%s) return %q;\n", condition, variant.Name))
		}
	}
	builder.WriteString("  return \"default\";\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) componentInputTSType(input ComponentInput) string {
	inputType := tsType(input.Type)
	if _, ok := g.findEntity(input.Type); ok {
		inputType = input.Type
	}
	if input.List {
		return inputType + "[]"
	}
	return inputType
}

func (g *webGenerator) componentDisplayExpression(component ComponentDecl) string {
	if len(component.Inputs) == 0 {
		return "variant"
	}
	for _, input := range component.Inputs {
		if !input.List && input.Type != "boolean" {
			return "props." + input.Name
		}
	}
	return "variant"
}

func (g *webGenerator) componentsForPage(page PageDecl, entity EntityDecl) []ComponentDecl {
	components := []ComponentDecl{}
	seen := map[string]bool{}
	fieldNames := append([]string{}, page.Table.Columns...)
	fieldNames = append(fieldNames, page.Form.Fields...)
	for _, fieldName := range fieldNames {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		component, ok := g.componentForField(field)
		if !ok || seen[component.Name] {
			continue
		}
		seen[component.Name] = true
		components = append(components, component)
	}
	return components
}

func (g *webGenerator) componentForField(field FieldDecl) (ComponentDecl, bool) {
	for _, component := range g.program.Components {
		if len(component.Inputs) != 1 {
			continue
		}
		input := component.Inputs[0]
		if input.Name == field.Name && input.Type == field.Type && !input.List {
			return component, true
		}
	}
	return ComponentDecl{}, false
}

func (g *webGenerator) itemRenderExpression(itemName string, field FieldDecl) string {
	component, ok := g.componentForField(field)
	if !ok {
		return g.itemDisplayExpression(itemName, field)
	}
	input := component.Inputs[0]
	return fmt.Sprintf("<%s %s={%s} />", component.Name, input.Name, componentPropExpression(itemName, input))
}

func (g *webGenerator) formComponentPreview(field FieldDecl) string {
	component, ok := g.componentForField(field)
	if !ok {
		return ""
	}
	input := component.Inputs[0]
	return fmt.Sprintf("<span className=\"field-preview\"><%s %s={%s} /></span>\n              ", component.Name, input.Name, formComponentPropExpression(field, input))
}

func componentPropExpression(itemName string, input ComponentInput) string {
	switch input.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("Number(%s.%s ?? 0)", itemName, input.Name)
	case "boolean":
		return fmt.Sprintf("Boolean(%s.%s)", itemName, input.Name)
	default:
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, input.Name)
	}
}

func formComponentPropExpression(field FieldDecl, input ComponentInput) string {
	switch input.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("Number(form.%s || 0)", field.Name)
	case "boolean":
		return fmt.Sprintf("form.%s === \"true\"", field.Name)
	default:
		return fmt.Sprintf("String(form.%s || \"\")", field.Name)
	}
}

func componentVariantConditionExpression(condition string, inputs []ComponentInput) string {
	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return ""
	}
	input, ok := componentInputByName(inputs, parts[0])
	if !ok || input.List {
		return ""
	}
	operator := parts[1]
	if !supportedComparisonOperators[operator] {
		return ""
	}
	right := parts[2]
	switch input.Type {
	case "number", "integer", "decimal", "money":
		if !isNumericLiteral(right) {
			return ""
		}
		return fmt.Sprintf("Number(props.%s) %s %s", input.Name, operator, right)
	case "boolean":
		if right != "true" && right != "false" {
			return ""
		}
		return fmt.Sprintf("props.%s %s %s", input.Name, operator, right)
	default:
		if operator != "==" && operator != "!=" {
			return ""
		}
		return fmt.Sprintf("String(props.%s) %s %q", input.Name, operator, right)
	}
}

func (g *webGenerator) openapiJSON() string {
	components := map[string]any{"schemas": g.openapiSchemas()}
	if g.program.Auth != nil {
		components["securitySchemes"] = map[string]any{
			"cookieAuth": map[string]any{
				"type": "apiKey",
				"in":   "cookie",
				"name": "black_session",
			},
		}
	}
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   g.program.App.Name + " API",
			"version": "0.1.0",
		},
		"paths":      g.openapiPaths(),
		"components": components,
	}
	content, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "{\n  \"openapi\": \"3.1.0\"\n}\n"
	}
	return string(content) + "\n"
}

func (g *webGenerator) openapiPaths() map[string]any {
	paths := map[string]any{}
	if g.program.Auth != nil {
		paths["/api/auth/register"] = map[string]any{
			"post": map[string]any{
				"summary":     "Register a user",
				"requestBody": openapiRequestBody("#/components/schemas/AuthRegisterInput"),
				"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		paths["/api/auth/login"] = map[string]any{
			"post": map[string]any{
				"summary":     "Login a user",
				"requestBody": openapiRequestBody("#/components/schemas/AuthLoginInput"),
				"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		paths["/api/auth/logout"] = map[string]any{
			"post": map[string]any{
				"summary":   "Logout current user",
				"responses": openapiJSONResponses(map[string]any{"type": "object"}),
			},
		}
		paths["/api/auth/me"] = map[string]any{
			"get": map[string]any{
				"summary":   "Read current user session",
				"responses": openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		if len(g.program.Roles) > 0 {
			paths["/api/auth/users"] = map[string]any{
				"get": map[string]any{
					"summary":  "List authenticated users",
					"security": []any{map[string]any{"cookieAuth": []any{}}},
					"responses": openapiJSONResponses(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"users": map[string]any{
								"type":  "array",
								"items": map[string]any{"$ref": "#/components/schemas/AuthUser"},
							},
						},
						"required": []string{"users"},
					}),
				},
			}
			paths["/api/auth/audit"] = map[string]any{
				"get": map[string]any{
					"summary":  "List audit log entries",
					"security": []any{map[string]any{"cookieAuth": []any{}}},
					"responses": openapiJSONResponses(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"logs": map[string]any{
								"type":  "array",
								"items": map[string]any{"$ref": "#/components/schemas/AuditLog"},
							},
						},
						"required": []string{"logs"},
					}),
				},
			}
			paths["/api/auth/users/{id}/role"] = map[string]any{
				"put": map[string]any{
					"summary":     "Update authenticated user role",
					"security":    []any{map[string]any{"cookieAuth": []any{}}},
					"parameters":  []any{openapiIDParameter()},
					"requestBody": openapiRequestBody("#/components/schemas/AuthRoleUpdateInput"),
					"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
				},
			}
		}
	}
	for _, page := range g.program.Pages {
		entity, ok := g.findEntity(page.Source)
		if !ok {
			continue
		}
		pathName := "/api/" + strings.ToLower(page.Name)
		entityRef := "#/components/schemas/" + entity.Name
		inputRef := "#/components/schemas/" + entity.Name + "Input"
		collectionOperations := map[string]any{
			"get": map[string]any{
				"summary": "List " + entity.Name + " records",
				"parameters": []any{
					map[string]any{
						"name": "archived",
						"in":   "query",
						"schema": map[string]any{
							"type": "string",
							"enum": []string{"all"},
						},
					},
				},
				"responses": openapiJSONResponses(map[string]any{
					"type":  "array",
					"items": map[string]any{"$ref": entityRef},
				}),
			},
		}
		if hasAction(page, "create") {
			collectionOperations["post"] = map[string]any{
				"summary":     "Create a " + entity.Name + " record",
				"requestBody": openapiRequestBody(inputRef),
				"responses":   openapiJSONResponses(map[string]any{"$ref": entityRef}),
			}
		}
		if hasAction(page, "delete") {
			collectionOperations["delete"] = map[string]any{
				"summary":     "Bulk delete " + entity.Name + " records",
				"requestBody": openapiArrayRequestBody("ids", "string"),
				"responses":   openapiJSONResponses(map[string]any{"type": "object"}),
			}
		}
		paths[pathName] = collectionOperations

		itemPath := pathName + "/{id}"
		itemOperations := map[string]any{
			"get": map[string]any{
				"summary":    "Read a " + entity.Name + " record",
				"parameters": []any{openapiIDParameter()},
				"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
			},
		}
		if hasAction(page, "edit") {
			itemOperations["put"] = map[string]any{
				"summary":     "Update a " + entity.Name + " record",
				"parameters":  []any{openapiIDParameter()},
				"requestBody": openapiRequestBody(inputRef),
				"responses":   openapiJSONResponses(map[string]any{"$ref": entityRef}),
			}
		}
		if hasAction(page, "delete") {
			itemOperations["delete"] = map[string]any{
				"summary":    "Delete a " + entity.Name + " record",
				"parameters": []any{openapiIDParameter()},
				"responses":  openapiJSONResponses(map[string]any{"type": "object"}),
			}
		}
		paths[itemPath] = itemOperations

		if hasAction(page, "archive") {
			paths[itemPath+"/archive"] = map[string]any{
				"patch": map[string]any{
					"summary":    "Archive a " + entity.Name + " record",
					"parameters": []any{openapiIDParameter()},
					"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
				},
			}
		}
		if hasAction(page, "restore") {
			paths[itemPath+"/restore"] = map[string]any{
				"patch": map[string]any{
					"summary":    "Restore a " + entity.Name + " record",
					"parameters": []any{openapiIDParameter()},
					"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
				},
			}
		}
		if g.hasRuntimePermissions() {
			for _, workflow := range g.workflowsForEntity(entity.Name) {
				for _, transition := range workflow.Transitions {
					paths[itemPath+"/workflow/"+transition.Name] = map[string]any{
						"post": map[string]any{
							"summary":    "Run " + workflow.Name + " transition " + transition.Name,
							"security":   []any{map[string]any{"cookieAuth": []any{}}},
							"parameters": []any{openapiIDParameter()},
							"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
						},
					}
				}
			}
		}
	}
	for _, api := range g.program.APIs {
		method := strings.ToLower(api.Method)
		if method == "" || api.Path == "" {
			continue
		}
		operations, ok := paths[api.Path].(map[string]any)
		if !ok {
			operations = map[string]any{}
		}
		operation := map[string]any{
			"summary":            api.Name,
			"x-blacklang-api":    api.Name,
			"x-blacklang-access": explicitAPIAccess(api),
			"responses":          openapiJSONResponses(map[string]any{"type": "object"}),
		}
		if api.Webhook {
			operation["x-blacklang-webhook"] = true
		}
		if explicitAPIAccess(api) == "private" && g.program.Auth != nil {
			operation["security"] = []any{map[string]any{"cookieAuth": []any{}}}
		}
		parameters := []any{}
		for _, param := range api.Params {
			parameters = append(parameters, openapiExplicitParameter(param, "path", true))
		}
		for _, query := range api.Queries {
			parameters = append(parameters, openapiExplicitParameter(query, "query", false))
		}
		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}
		if method == "post" || method == "put" || method == "patch" {
			operation["requestBody"] = openapiRequestBody("#/components/schemas/" + api.Name + "Input")
		}
		operations[method] = operation
		paths[api.Path] = operations
	}
	return paths
}

func (g *webGenerator) openapiSchemas() map[string]any {
	schemas := map[string]any{}
	if g.program.Auth != nil {
		schemas["AuthRegisterInput"] = openapiObjectSchema(map[string]any{
			"name":     map[string]any{"type": "string"},
			"email":    map[string]any{"type": "string", "format": "email"},
			"password": map[string]any{"type": "string", "minLength": 8},
		}, []string{"name", "email", "password"})
		schemas["AuthLoginInput"] = openapiObjectSchema(map[string]any{
			"email":    map[string]any{"type": "string", "format": "email"},
			"password": map[string]any{"type": "string", "minLength": 8},
		}, []string{"email", "password"})
		schemas["AuthUser"] = openapiObjectSchema(map[string]any{
			"id":    map[string]any{"type": "string"},
			"name":  map[string]any{"type": "string"},
			"email": map[string]any{"type": "string", "format": "email"},
			"role":  map[string]any{"type": "string"},
		}, []string{"id", "name", "email", "role"})
		if len(g.program.Roles) > 0 {
			schemas["AuthRoleUpdateInput"] = openapiObjectSchema(map[string]any{
				"role": map[string]any{
					"type": "string",
					"enum": g.roleNames(),
				},
			}, []string{"role"})
			schemas["AuditLog"] = openapiObjectSchema(map[string]any{
				"id":          map[string]any{"type": "string"},
				"actorUserId": map[string]any{"type": "string"},
				"actorRole":   map[string]any{"type": "string"},
				"action":      map[string]any{"type": "string"},
				"resource":    map[string]any{"type": "string"},
				"resourceId":  map[string]any{"type": "string"},
				"summary":     map[string]any{"type": "string"},
				"createdAt":   map[string]any{"type": "string"},
			}, []string{"id", "actorUserId", "actorRole", "action", "resource", "resourceId", "summary", "createdAt"})
		}
		schemas["AuthUserResponse"] = openapiObjectSchema(map[string]any{
			"user": map[string]any{"$ref": "#/components/schemas/AuthUser"},
		}, []string{"user"})
	}
	for _, api := range g.program.APIs {
		method := strings.ToUpper(api.Method)
		if method == "POST" || method == "PUT" || method == "PATCH" {
			schemas[api.Name+"Input"] = map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			}
		}
	}
	for _, entity := range g.program.Entities {
		properties := map[string]any{
			"id":         map[string]any{"type": "string"},
			"archivedAt": map[string]any{"type": []string{"string", "null"}},
		}
		inputProperties := map[string]any{}
		required := []string{"id"}
		inputRequired := []string{}
		for _, field := range entity.Fields {
			name := field.Name
			if g.isRelationField(field) {
				name = relationIDFieldName(field)
			}
			schema := openapiFieldSchema(field, g.isRelationField(field))
			properties[name] = schema
			inputProperties[name] = schema
			if hasModifier(field, "required") {
				required = append(required, name)
				inputRequired = append(inputRequired, name)
			}
		}
		schemas[entity.Name] = openapiObjectSchema(properties, required)
		schemas[entity.Name+"Input"] = openapiObjectSchema(inputProperties, inputRequired)
	}
	return schemas
}

func openapiFieldSchema(field FieldDecl, relation bool) map[string]any {
	if relation {
		return map[string]any{"type": "string"}
	}
	switch field.Type {
	case "number", "integer":
		return map[string]any{"type": "integer"}
	case "decimal", "money":
		return map[string]any{"type": "number"}
	case "boolean":
		return map[string]any{"type": "boolean"}
	case "date", "datetime":
		return map[string]any{"type": "string", "format": "date-time"}
	case "email":
		return map[string]any{"type": "string", "format": "email"}
	default:
		return map[string]any{"type": "string"}
	}
}

func openapiObjectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func openapiRequestBody(schemaRef string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": schemaRef},
			},
		},
	}
}

func openapiArrayRequestBody(name string, itemType string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						name: map[string]any{
							"type":  "array",
							"items": map[string]any{"type": itemType},
						},
					},
					"required": []string{name},
				},
			},
		},
	}
}

func openapiIDParameter() map[string]any {
	return map[string]any{
		"name":     "id",
		"in":       "path",
		"required": true,
		"schema":   map[string]any{"type": "string"},
	}
}

func openapiExplicitParameter(param APIParamDecl, location string, required bool) map[string]any {
	return map[string]any{
		"name":     param.Name,
		"in":       location,
		"required": required,
		"schema":   openapiAPIParamSchema(param.Type),
	}
}

func openapiAPIParamSchema(paramType string) map[string]any {
	switch paramType {
	case "number", "integer":
		return map[string]any{"type": "integer"}
	case "decimal", "money":
		return map[string]any{"type": "number"}
	case "boolean":
		return map[string]any{"type": "boolean"}
	case "date":
		return map[string]any{"type": "string", "format": "date"}
	case "datetime":
		return map[string]any{"type": "string", "format": "date-time"}
	case "email":
		return map[string]any{"type": "string", "format": "email"}
	default:
		return map[string]any{"type": "string"}
	}
}

func explicitAPIAccess(api APIDecl) string {
	if api.Access != "" {
		return api.Access
	}
	return "private"
}

func openapiJSONResponses(schema map[string]any) map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "OK",
			"content": map[string]any{
				"application/json": map[string]any{"schema": schema},
			},
		},
		"400": map[string]any{"description": "Validation error"},
		"404": map[string]any{"description": "Not found"},
	}
}

func (g *webGenerator) prismaSchema() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("generator client {\n")
	builder.WriteString("  provider = \"prisma-client\"\n")
	builder.WriteString("  output   = \"../src/generated/prisma\"\n")
	builder.WriteString("}\n\n")
	builder.WriteString("datasource db {\n")
	builder.WriteString("  provider = \"sqlite\"\n")
	builder.WriteString("}\n\n")
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("model %s {\n", entity.Name))
		builder.WriteString("  id String @id @default(cuid())\n")
		for _, field := range entity.Fields {
			if g.isRelationField(field) {
				builder.WriteString(fmt.Sprintf("  %s String%s%s\n", relationIDFieldName(field), prismaRelationIDOptional(field), prismaAttributes(field)))
				builder.WriteString(fmt.Sprintf("  %s %s%s @relation(%q, fields: [%s], references: [id])\n", field.Name, field.Type, prismaOptional(field), relationName(entity, field), relationIDFieldName(field)))
				continue
			}
			builder.WriteString(fmt.Sprintf("  %s %s%s%s\n", field.Name, prismaType(field.Type), prismaOptional(field), prismaAttributes(field)))
		}
		for _, relation := range g.incomingRelations(entity.Name) {
			builder.WriteString(fmt.Sprintf("  %s %s[] @relation(%q)\n", relationBackFieldName(relation.entity, relation.field), relation.entity.Name, relationName(relation.entity, relation.field)))
		}
		builder.WriteString("  archivedAt DateTime?\n")
		builder.WriteString("  createdAt DateTime @default(now())\n")
		builder.WriteString("  updatedAt DateTime @updatedAt\n")
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func (g *webGenerator) types() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("export type %s = {\n", entity.Name))
		builder.WriteString("  id: string;\n")
		for _, field := range entity.Fields {
			if g.isRelationField(field) {
				optional := ""
				if !hasModifier(field, "required") {
					optional = "?"
				}
				builder.WriteString(fmt.Sprintf("  %s%s: string;\n", relationIDFieldName(field), optional))
				builder.WriteString(fmt.Sprintf("  %s?: %s | null;\n", field.Name, field.Type))
				continue
			}
			optional := ""
			if hasModifier(field, "optional") || (!hasModifier(field, "required") && !hasModifier(field, "default")) {
				optional = "?"
			}
			builder.WriteString(fmt.Sprintf("  %s%s: %s;\n", field.Name, optional, tsType(field.Type)))
		}
		builder.WriteString("  archivedAt?: string | null;\n")
		builder.WriteString("};\n\n")
	}
	return builder.String()
}

func (g *webGenerator) route(page PageDecl, entity EntityDecl) string {
	lower := strings.ToLower(entity.Name)
	path := strings.ToLower(page.Name)
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import express from \"express\";\n")
	builder.WriteString("import { prisma } from \"../db\";\n")
	if g.hasRuntimePermissions() {
		builder.WriteString("import { canAccessField, filterWritableFields, requirePageAccess, requirePermission, writeAuditLog } from \"./auth\";\n")
	}
	builder.WriteString(fmt.Sprintf("import { validate%sInput } from \"../validation/%s\";\n\n", entity.Name, lower))
	builder.WriteString(fmt.Sprintf("export const %sRouter = express.Router();\n\n", lower))
	builder.WriteString(fmt.Sprintf("const %sModel = prisma.%s;\n\n", lower, lower))
	if g.program.Auth != nil && len(page.Access) > 0 {
		builder.WriteString(fmt.Sprintf("%sRouter.use(requirePageAccess(%s));\n\n", lower, tsStringArrayLiteral(page.Access)))
	}
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("function sanitize%s(item: any, role: string) {\n", entity.Name))
		builder.WriteString("  if (!item) return item;\n")
		builder.WriteString("  const output = { ...item };\n")
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("  if (!canAccessField(role, \"read\", %q, %q)) delete output.%s;\n", entity.Name, field.Name, g.sqliteColumnName(field)))
			if g.isRelationField(field) {
				builder.WriteString(fmt.Sprintf("  if (!canAccessField(role, \"read\", %q, %q)) delete output.%s;\n", entity.Name, field.Name, field.Name))
			}
		}
		builder.WriteString("  return output;\n")
		builder.WriteString("}\n\n")
		builder.WriteString("function currentUser(req: express.Request) {\n")
		builder.WriteString("  return (req as any).blackUser as { id: string; name: string; email: string; role: string } | undefined;\n")
		builder.WriteString("}\n\n")
		builder.WriteString("function currentRole(req: express.Request) {\n")
		builder.WriteString("  return currentUser(req)?.role ?? \"\";\n")
		builder.WriteString("}\n\n")
	}
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("read", entity.Name)))
	builder.WriteString("  const includeArchived = req.query.archived === \"all\";\n")
	builder.WriteString(fmt.Sprintf("  const items = await %sModel.findMany({\n", lower))
	builder.WriteString("    where: includeArchived ? {} : { archivedAt: null },\n")
	if g.hasRelationFields(entity) {
		builder.WriteString(g.prismaIncludeLine(entity, "    "))
	}
	builder.WriteString("    orderBy: { createdAt: \"desc\" }\n")
	builder.WriteString("  });\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(items.map((item) => sanitize%s(item, currentRole(req))));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(items);\n")
	}
	builder.WriteString("});\n\n")
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s/:id\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("read", entity.Name)))
	builder.WriteString(fmt.Sprintf("  const item = await %sModel.findUnique({\n", lower))
	builder.WriteString("    where: { id: String(req.params.id) }")
	if g.hasRelationFields(entity) {
		builder.WriteString(",\n")
		builder.WriteString(g.prismaIncludeLine(entity, "    ", false))
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("  });\n")
	builder.WriteString("  if (!item) {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(sanitize%s(item, currentRole(req)));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(item);\n")
	}
	builder.WriteString("});\n\n")
	builder.WriteString(fmt.Sprintf("%sRouter.post(\"/%s\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("create", entity.Name)))
	builder.WriteString(fmt.Sprintf("  const validation = validate%sInput(req.body);\n", entity.Name))
	builder.WriteString("  if (!validation.valid) {\n")
	builder.WriteString("    res.status(400).json({ error: validation.errors });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  const writableValue = filterWritableFields(currentRole(req), \"create\", %q, validation.value as Record<string, unknown>);\n", entity.Name))
		builder.WriteString("  if (Object.keys(writableValue).length === 0) {\n")
		builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString(fmt.Sprintf("  const item = await %sModel.create({\n", lower))
	if g.hasRuntimePermissions() {
		builder.WriteString("    data: writableValue as any")
	} else {
		builder.WriteString("    data: validation.value as any")
	}
	if g.hasRelationFields(entity) {
		builder.WriteString(",\n")
		builder.WriteString(g.prismaIncludeLine(entity, "    ", false))
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("  });\n\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), \"create\", %q, item.id, %q);\n", entity.Name, entity.Name+" record created"))
		builder.WriteString(fmt.Sprintf("  res.status(201).json(sanitize%s(item, currentRole(req)));\n", entity.Name))
	} else {
		builder.WriteString("  res.status(201).json(item);\n")
	}
	builder.WriteString("});\n\n")
	builder.WriteString(fmt.Sprintf("%sRouter.put(\"/%s/:id\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("update", entity.Name)))
	builder.WriteString(fmt.Sprintf("  const validation = validate%sInput(req.body);\n", entity.Name))
	builder.WriteString("  if (!validation.valid) {\n")
	builder.WriteString("    res.status(400).json({ error: validation.errors });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  const writableValue = filterWritableFields(currentRole(req), \"update\", %q, validation.value as Record<string, unknown>);\n", entity.Name))
		builder.WriteString("  if (Object.keys(writableValue).length === 0) {\n")
		builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  try {\n")
	builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", lower))
	builder.WriteString("      where: { id: String(req.params.id) },\n")
	if g.hasRuntimePermissions() {
		builder.WriteString("      data: writableValue as any")
	} else {
		builder.WriteString("      data: validation.value as any")
	}
	if g.hasRelationFields(entity) {
		builder.WriteString(",\n")
		builder.WriteString(g.prismaIncludeLine(entity, "      ", false))
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("    });\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"update\", %q, item.id, %q);\n", entity.Name, entity.Name+" record updated"))
		builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRole(req)));\n", entity.Name))
	} else {
		builder.WriteString("    res.json(item);\n")
	}
	builder.WriteString("  } catch {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("  }\n")
	builder.WriteString("});\n\n")
	if hasAction(page, "archive") {
		builder.WriteString(fmt.Sprintf("%sRouter.patch(\"/%s/:id/archive\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("update", entity.Name)))
		builder.WriteString("  try {\n")
		builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", lower))
		builder.WriteString("      where: { id: String(req.params.id) },\n")
		builder.WriteString("      data: { archivedAt: new Date() }")
		if g.hasRelationFields(entity) {
			builder.WriteString(",\n")
			builder.WriteString(g.prismaIncludeLine(entity, "      ", false))
		} else {
			builder.WriteString("\n")
		}
		builder.WriteString("    });\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"archive\", %q, item.id, %q);\n", entity.Name, entity.Name+" record archived"))
			builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRole(req)));\n", entity.Name))
		} else {
			builder.WriteString("    res.json(item);\n")
		}
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n\n")
	}
	if hasAction(page, "restore") {
		builder.WriteString(fmt.Sprintf("%sRouter.patch(\"/%s/:id/restore\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("update", entity.Name)))
		builder.WriteString("  try {\n")
		builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", lower))
		builder.WriteString("      where: { id: String(req.params.id) },\n")
		builder.WriteString("      data: { archivedAt: null }")
		if g.hasRelationFields(entity) {
			builder.WriteString(",\n")
			builder.WriteString(g.prismaIncludeLine(entity, "      ", false))
		} else {
			builder.WriteString("\n")
		}
		builder.WriteString("    });\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"restore\", %q, item.id, %q);\n", entity.Name, entity.Name+" record restored"))
			builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRole(req)));\n", entity.Name))
		} else {
			builder.WriteString("    res.json(item);\n")
		}
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n\n")
	}
	builder.WriteString(g.workflowTransitionRoutes(page, entity))
	builder.WriteString(fmt.Sprintf("%sRouter.delete(\"/%s\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("delete", entity.Name)))
	builder.WriteString("  const ids = Array.isArray(req.body?.ids)\n")
	builder.WriteString("    ? req.body.ids.filter((id: unknown) => typeof id === \"string\")\n")
	builder.WriteString("    : [];\n\n")
	builder.WriteString("  if (ids.length === 0) {\n")
	builder.WriteString("    res.status(400).json({ error: \"ids are required\" });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	builder.WriteString(fmt.Sprintf("  const result = await %sModel.deleteMany({\n", lower))
	builder.WriteString("    where: {\n")
	builder.WriteString("      id: { in: ids }\n")
	builder.WriteString("    }\n")
	builder.WriteString("  });\n\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), \"bulkDelete\", %q, ids.join(\",\"), String(result.count) + \" %s records deleted\");\n", entity.Name, strings.ToLower(entity.Name)))
	}
	builder.WriteString("  res.json({ deleted: result.count });\n")
	builder.WriteString("});\n\n")
	builder.WriteString(fmt.Sprintf("%sRouter.delete(\"/%s/:id\", %sasync (req, res) => {\n", lower, path, g.permissionMiddleware("delete", entity.Name)))
	builder.WriteString("  try {\n")
	builder.WriteString(fmt.Sprintf("    await %sModel.delete({\n", lower))
	builder.WriteString("      where: { id: String(req.params.id) }\n")
	builder.WriteString("    });\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"delete\", %q, String(req.params.id), %q);\n", entity.Name, entity.Name+" record deleted"))
	}
	builder.WriteString("    res.status(204).send();\n")
	builder.WriteString("  } catch {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("  }\n")
	builder.WriteString("});\n")
	return builder.String()
}

func (g *webGenerator) apiClient(page PageDecl, entity EntityDecl) string {
	lower := strings.ToLower(entity.Name)
	path := strings.ToLower(page.Name)
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import type { %s } from "../types";

export type %sInput = Omit<%s, "id">;

const endpoint = "/api/%s";

function csrfHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = document.cookie.split("; ").find((item) => item.startsWith("black_csrf="))?.split("=")[1] ?? "";
  if (token) headers["X-CSRF-Token"] = decodeURIComponent(token);
  return headers;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "same-origin",
    headers: csrfHeaders(),
    ...options
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Request failed");
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export const %sApi = {
  list: (includeArchived = false) =>
    request<%s[]>(includeArchived ? endpoint + "?archived=all" : endpoint),
  get: (id: string) => request<%s>(endpoint + "/" + id),
  create: (input: %sInput) =>
    request<%s>(endpoint, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  update: (id: string, input: %sInput) =>
    request<%s>(endpoint + "/" + id, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  bulkDelete: (ids: string[]) =>
    request<{ deleted: number }>(endpoint, {
      method: "DELETE",
      body: JSON.stringify({ ids })
    }),
  archive: (id: string) =>
    request<%s>(endpoint + "/" + id + "/archive", {
      method: "PATCH"
    }),
  restore: (id: string) =>
    request<%s>(endpoint + "/" + id + "/restore", {
      method: "PATCH"
    }),
  delete: (id: string) =>
    request<void>(endpoint + "/" + id, {
      method: "DELETE"
    })%s
};
`, entity.Name, entity.Name, entity.Name, path, lower, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, g.workflowClientMethods(entity))
}

func (g *webGenerator) workflowTransitionRoutes(page PageDecl, entity EntityDecl) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	lower := strings.ToLower(entity.Name)
	path := strings.ToLower(page.Name)
	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf("%sRouter.post(\"/%s/:id/workflow/%s\", %sasync (req, res) => {\n", lower, path, transition.Name, g.permissionMiddleware("update", entity.Name)))
			if len(transition.Allow) > 0 {
				builder.WriteString("  const user = currentUser(req);\n")
				if containsString(transition.Allow, "authenticated") {
					builder.WriteString("  if (!user) {\n")
				} else {
					builder.WriteString(fmt.Sprintf("  if (!user || !%s.includes(user.role)) {\n", tsStringArrayLiteral(transition.Allow)))
				}
				builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
				builder.WriteString("    return;\n")
				builder.WriteString("  }\n\n")
			}
			builder.WriteString(fmt.Sprintf("  const existing = await %sModel.findUnique({ where: { id: String(req.params.id) } });\n", lower))
			builder.WriteString("  if (!existing) {\n")
			builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
			builder.WriteString(fmt.Sprintf("  if (String(existing.status ?? \"\") !== %q) {\n", transition.From))
			builder.WriteString(fmt.Sprintf("    res.status(409).json({ error: %q });\n", fmt.Sprintf("Transition %s requires status %s", transition.Name, transition.From)))
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
			builder.WriteString(fmt.Sprintf("  const item = await %sModel.update({\n", lower))
			builder.WriteString("    where: { id: String(req.params.id) },\n")
			builder.WriteString(fmt.Sprintf("    data: { status: %q }", transition.To))
			if g.hasRelationFields(entity) {
				builder.WriteString(",\n")
				builder.WriteString(g.prismaIncludeLine(entity, "    ", false))
			} else {
				builder.WriteString("\n")
			}
			builder.WriteString("  });\n\n")
			builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), %q, %q, item.id, %q);\n", "workflow."+transition.Name, entity.Name, fmt.Sprintf("%s: %s -> %s", workflow.Name, transition.From, transition.To)))
			builder.WriteString(fmt.Sprintf("  res.json(sanitize%s(item, currentRole(req)));\n", entity.Name))
			builder.WriteString("});\n\n")
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowClientMethods(entity EntityDecl) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf(",\n  transition%s: (id: string) =>\n    request<%s>(endpoint + \"/\" + id + \"/workflow/%s\", {\n      method: \"POST\"\n    })", title(transition.Name), entity.Name, transition.Name))
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowsForEntity(entityName string) []WorkflowDecl {
	workflows := []WorkflowDecl{}
	for _, workflow := range g.program.Workflows {
		if workflow.Source == entityName {
			workflows = append(workflows, workflow)
		}
	}
	return workflows
}

func (g *webGenerator) workflowPageActionFunctions(entity EntityDecl) string {
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	lower := strings.ToLower(entity.Name)
	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			functionName := workflowActionFunctionName(transition)
			builder.WriteString(fmt.Sprintf("  async function %s(item: %s) {\n", functionName, entity.Name))
			builder.WriteString(fmt.Sprintf("    if (!canUpdate || item.archivedAt || String(item.status ?? \"\") !== %q) return;\n", transition.From))
			builder.WriteString("    setSaving(true);\n")
			builder.WriteString("    setError(null);\n")
			builder.WriteString("    try {\n")
			builder.WriteString(fmt.Sprintf("      const updated = await %sApi.transition%s(item.id);\n", lower, title(transition.Name)))
			builder.WriteString("      setItems((current) => current.map((existing) => existing.id === item.id ? updated : existing));\n")
			builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? updated : current);\n")
			builder.WriteString("    } catch (reason: unknown) {\n")
			builder.WriteString(fmt.Sprintf("      setError(reason instanceof Error ? reason.message : %q);\n", "Unable to run transition "+transition.Name))
			builder.WriteString("    } finally {\n")
			builder.WriteString("      setSaving(false);\n")
			builder.WriteString("    }\n")
			builder.WriteString("  }\n\n")
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowPageActionButtons(entity EntityDecl) string {
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf("                  {canUpdate && !item.archivedAt && String(item.status ?? \"\") === %q && <button className=\"secondary\" type=\"button\" disabled={saving} onClick={() => %s(item)}>%s</button>}\n", transition.From, workflowActionFunctionName(transition), identifierLabel(transition.Name)))
		}
	}
	return builder.String()
}

func (g *webGenerator) validation(entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("type ValidationResult<T> =\n")
	builder.WriteString("  | { valid: true; value: T }\n")
	builder.WriteString("  | { valid: false; errors: string[] };\n\n")
	builder.WriteString(fmt.Sprintf("export function validate%sInput(input: any): ValidationResult<Record<string, unknown>> {\n", entity.Name))
	builder.WriteString("  const errors: string[] = [];\n")
	builder.WriteString("  const value: any = {};\n\n")
	for _, field := range entity.Fields {
		builder.WriteString(g.validationFieldBlock(field))
	}
	builder.WriteString(g.entityValidationBlocks(entity))
	builder.WriteString("\n  if (errors.length > 0) {\n")
	builder.WriteString("    return { valid: false, errors };\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  return { valid: true, value };\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) validationFieldBlock(field FieldDecl) string {
	if !g.isRelationField(field) {
		return validationFieldBlock(field)
	}

	var builder strings.Builder
	name := field.Name
	valueName := relationIDFieldName(field)
	builder.WriteString(fmt.Sprintf("  if (input.%s === undefined || input.%s === null || input.%s === \"\") {\n", valueName, valueName, valueName))
	if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.push(%q);\n", fieldValidationMessage(field, name+" is required")))
	} else {
		builder.WriteString(fmt.Sprintf("    value.%s = undefined;\n", valueName))
	}
	builder.WriteString("  } else {\n")
	builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", valueName, valueName))
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) page(page PageDecl, entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import { useEffect, useMemo, useState } from \"react\";\n")
	builder.WriteString("import type { FormEvent } from \"react\";\n")
	builder.WriteString(fmt.Sprintf("import { %sApi } from \"../api/%s\";\n", strings.ToLower(entity.Name), strings.ToLower(entity.Name)))
	for _, field := range g.relationFields(entity) {
		builder.WriteString(fmt.Sprintf("import { %sApi } from \"../api/%s\";\n", strings.ToLower(field.Type), strings.ToLower(field.Type)))
	}
	for _, component := range g.componentsForPage(page, entity) {
		builder.WriteString(fmt.Sprintf("import { %s } from \"../components/%s\";\n", component.Name, component.Name))
	}
	builder.WriteString(fmt.Sprintf("import type { %s } from \"../types\";\n\n", g.typeImportList(page, entity)))
	builder.WriteString("type FormState = Record<string, string>;\n\n")
	builder.WriteString("type FormErrors = Record<string, string>;\n\n")
	builder.WriteString("type PagePermissions = {\n")
	builder.WriteString("  read: boolean;\n")
	builder.WriteString("  create: boolean;\n")
	builder.WriteString("  update: boolean;\n")
	builder.WriteString("  delete: boolean;\n")
	builder.WriteString("  fields: Record<string, boolean>;\n")
	builder.WriteString("  writeFields: Record<string, boolean>;\n")
	builder.WriteString("};\n\n")
	builder.WriteString("const defaultPermissions: PagePermissions = { read: true, create: true, update: true, delete: true, fields: {}, writeFields: {} };\n\n")
	builder.WriteString("type PageProps = {\n")
	builder.WriteString("  onNavigate?: (page: string) => void;\n")
	builder.WriteString("  permissions?: PagePermissions;\n")
	builder.WriteString("};\n\n")
	builder.WriteString(fmt.Sprintf("const emptyForm: FormState = %s;\n\n", formStateLiteral(page.Form.Fields, entity)))
	builder.WriteString(g.formValidationFunction(page, entity))
	builder.WriteString(fmt.Sprintf("export function %sPage({ onNavigate, permissions = defaultPermissions }: PageProps) {\n", page.Name))
	builder.WriteString(fmt.Sprintf("  const [items, setItems] = useState<%s[]>([]);\n", entity.Name))
	builder.WriteString("  const [search, setSearch] = useState(\"\");\n\n")
	if len(page.Table.Filters) > 0 {
		builder.WriteString(fmt.Sprintf("  const [filters, setFilters] = useState<Record<string, string>>(%s);\n", tableFiltersLiteral(page.Table.Filters)))
	}
	builder.WriteString("  const [showArchived, setShowArchived] = useState(false);\n")
	builder.WriteString(fmt.Sprintf("  const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>(%s);\n", columnVisibilityLiteral(page.Table.Columns)))
	if page.Table.Paginate > 0 {
		builder.WriteString("  const [currentPage, setCurrentPage] = useState(1);\n")
	}
	builder.WriteString("  const [form, setForm] = useState<FormState>(emptyForm);\n")
	builder.WriteString("  const [touchedFields, setTouchedFields] = useState<Record<string, boolean>>({});\n")
	builder.WriteString("  const [submitted, setSubmitted] = useState(false);\n")
	builder.WriteString("  const [editingId, setEditingId] = useState<string | null>(null);\n\n")
	builder.WriteString(fmt.Sprintf("  const [selectedItem, setSelectedItem] = useState<%s | null>(null);\n", entity.Name))
	for _, field := range g.relationFields(entity) {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState<%s[]>([]);\n", relationOptionsStateName(field), title(relationOptionsStateName(field)), field.Type))
	}
	state, hasState := g.stateForPage(page)
	if hasState {
		builder.WriteString(g.pageStateHooks(state))
	}
	builder.WriteString("  const [selectedIds, setSelectedIds] = useState<string[]>([]);\n")
	builder.WriteString("  const [loading, setLoading] = useState(true);\n")
	builder.WriteString("  const [reading, setReading] = useState(false);\n")
	builder.WriteString("  const [saving, setSaving] = useState(false);\n")
	builder.WriteString("  const [error, setError] = useState<string | null>(null);\n\n")
	builder.WriteString("  useEffect(() => {\n")
	builder.WriteString("    let active = true;\n")
	builder.WriteString("    setLoading(true);\n")
	builder.WriteString("    setError(null);\n")
	builder.WriteString(fmt.Sprintf("    %sApi.list(showArchived)\n", strings.ToLower(entity.Name)))
	builder.WriteString("      .then((records) => {\n")
	builder.WriteString("        if (active) setItems(records);\n")
	builder.WriteString("      })\n")
	builder.WriteString("      .catch((reason: unknown) => {\n")
	builder.WriteString("        if (active) setError(reason instanceof Error ? reason.message : \"Unable to load records\");\n")
	builder.WriteString("      })\n")
	builder.WriteString("      .finally(() => {\n")
	builder.WriteString("        if (active) setLoading(false);\n")
	builder.WriteString("      });\n\n")
	builder.WriteString("    return () => {\n")
	builder.WriteString("      active = false;\n")
	builder.WriteString("    };\n")
	builder.WriteString("  }, [showArchived]);\n\n")
	for _, field := range g.relationFields(entity) {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    let active = true;\n")
		builder.WriteString(fmt.Sprintf("    %sApi.list()\n", strings.ToLower(field.Type)))
		builder.WriteString("      .then((records) => {\n")
		builder.WriteString(fmt.Sprintf("        if (active) set%s(records);\n", title(relationOptionsStateName(field))))
		builder.WriteString("      })\n")
		builder.WriteString("      .catch((reason: unknown) => {\n")
		builder.WriteString("        if (active) setError(reason instanceof Error ? reason.message : \"Unable to load relation options\");\n")
		builder.WriteString("      });\n\n")
		builder.WriteString("    return () => {\n")
		builder.WriteString("      active = false;\n")
		builder.WriteString("    };\n")
		builder.WriteString("  }, []);\n\n")
	}
	if page.Table.Paginate > 0 {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    setCurrentPage(1);\n")
		if len(page.Table.Filters) > 0 {
			builder.WriteString("  }, [search, showArchived, filters]);\n\n")
		} else {
			builder.WriteString("  }, [search, showArchived]);\n\n")
		}
	}
	builder.WriteString("  const visibleItems = useMemo(() => {\n")
	builder.WriteString("    const query = search.trim().toLowerCase();\n")
	builder.WriteString("    const searchedItems = query ? items.filter((item) =>\n")
	builder.WriteString(g.searchExpression(page.Table.Search, entity))
	builder.WriteString("    ) : items;\n")
	builder.WriteString(g.filterExpression(page.Table.Filters, entity))
	builder.WriteString(g.sortExpression(page.Table.Sort, entity))
	if len(page.Table.Filters) > 0 {
		builder.WriteString("  }, [items, search, filters]);\n\n")
	} else {
		builder.WriteString("  }, [items, search]);\n\n")
	}
	if page.Table.Paginate > 0 {
		builder.WriteString(fmt.Sprintf("  const pageSize = %d;\n", page.Table.Paginate))
		builder.WriteString("  const totalPages = Math.max(1, Math.ceil(visibleItems.length / pageSize));\n")
		builder.WriteString("  const safeCurrentPage = Math.min(currentPage, totalPages);\n")
		builder.WriteString("  const paginatedItems = useMemo(() => {\n")
		builder.WriteString("    const start = (safeCurrentPage - 1) * pageSize;\n")
		builder.WriteString("    return visibleItems.slice(start, start + pageSize);\n")
		builder.WriteString("  }, [visibleItems, safeCurrentPage]);\n\n")
	} else {
		builder.WriteString("  const paginatedItems = visibleItems;\n\n")
	}
	builder.WriteString("  const visibleItemIds = useMemo(() => paginatedItems.map((item) => item.id), [paginatedItems]);\n")
	builder.WriteString("  const allVisibleSelected = visibleItemIds.length > 0 && visibleItemIds.every((id) => selectedIds.includes(id));\n\n")
	builder.WriteString("  const visibleColumnCount = Object.entries(visibleColumns).filter(([column, visible]) => visible && permissions.fields[column] !== false).length;\n")
	builder.WriteString(fmt.Sprintf("  const tableColspan = visibleColumnCount + %d;\n\n", g.staticTableExtraColumns(page, entity)))
	builder.WriteString("  const canCreate = permissions.create;\n")
	builder.WriteString("  const canUpdate = permissions.update;\n")
	builder.WriteString("  const canDelete = permissions.delete;\n")
	builder.WriteString("  const canSave = editingId ? canUpdate : canCreate;\n\n")
	builder.WriteString("  const formErrors = useMemo(() => validateForm(form), [form]);\n")
	builder.WriteString("  const visibleFormErrors = useMemo(() => Object.fromEntries(Object.entries(formErrors).filter(([field]) => touchedFields[field] || submitted)), [formErrors, touchedFields, submitted]);\n\n")
	builder.WriteString(fmt.Sprintf("  const missingRequiredRelations = %s;\n\n", g.missingRequiredRelationsExpression(page, entity)))
	builder.WriteString("  function updateField(field: string, value: string) {\n")
	builder.WriteString("    setForm((current) => ({ ...current, [field]: value }));\n")
	builder.WriteString("    setTouchedFields((current) => ({ ...current, [field]: true }));\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  function toggleItemSelection(id: string) {\n")
	builder.WriteString("    setSelectedIds((current) => current.includes(id) ? current.filter((value) => value !== id) : [...current, id]);\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  function toggleVisibleSelection() {\n")
	builder.WriteString("    if (allVisibleSelected) {\n")
	builder.WriteString("      setSelectedIds((current) => current.filter((id) => !visibleItemIds.includes(id)));\n")
	builder.WriteString("      return;\n")
	builder.WriteString("    }\n\n")
	builder.WriteString("    setSelectedIds((current) => Array.from(new Set([...current, ...visibleItemIds])));\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  function toggleColumn(column: string) {\n")
	builder.WriteString("    setVisibleColumns((current) => ({ ...current, [column]: !current[column] }));\n")
	builder.WriteString("  }\n\n")
	if len(page.Table.Filters) > 0 {
		builder.WriteString("  function updateFilter(field: string, value: string) {\n")
		builder.WriteString("    setFilters((current) => ({ ...current, [field]: value }));\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function resetForm() {\n")
	builder.WriteString("    setForm(emptyForm);\n")
	builder.WriteString("    setTouchedFields({});\n")
	builder.WriteString("    setSubmitted(false);\n")
	builder.WriteString("    setEditingId(null);\n")
	if hasState {
		builder.WriteString(g.resetPageModals(state))
	}
	builder.WriteString("  }\n\n")
	if hasState {
		builder.WriteString(g.pageModalHelpers(state))
	}
	builder.WriteString("  async function viewItem(id: string) {\n")
	builder.WriteString("    setReading(true);\n")
	builder.WriteString("    setError(null);\n")
	builder.WriteString("    try {\n")
	builder.WriteString(fmt.Sprintf("      const item = await %sApi.get(id);\n", strings.ToLower(entity.Name)))
	builder.WriteString("      setSelectedItem(item);\n")
	builder.WriteString("    } catch (reason: unknown) {\n")
	builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to load record details\");\n")
	builder.WriteString("    } finally {\n")
	builder.WriteString("      setReading(false);\n")
	builder.WriteString("    }\n")
	builder.WriteString("  }\n\n")
	if hasAction(page, "create") || hasAction(page, "edit") {
		builder.WriteString("  async function saveItem(event: FormEvent<HTMLFormElement>) {\n")
		builder.WriteString("    event.preventDefault();\n")
		builder.WriteString("    if (!canSave) return;\n")
		builder.WriteString("    setSubmitted(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    const nextErrors = validateForm(form);\n")
		builder.WriteString("    if (Object.keys(nextErrors).length > 0) return;\n\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    const input = {\n")
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			inputName := field.Name
			inputValue := formValueExpression(field)
			if g.isRelationField(field) {
				inputName = relationIDFieldName(field)
				inputValue = fmt.Sprintf("form.%s", field.Name)
			}
			builder.WriteString(fmt.Sprintf("      %s: %s,\n", inputName, inputValue))
		}
		builder.WriteString("    };\n\n")
		builder.WriteString("    try {\n")
		builder.WriteString("      if (editingId) {\n")
		builder.WriteString(fmt.Sprintf("        const saved = await %sApi.update(editingId, input);\n", strings.ToLower(entity.Name)))
		builder.WriteString("        setItems((current) => current.map((existing) => existing.id === editingId ? saved : existing));\n")
		builder.WriteString("        setSelectedItem((current) => current?.id === saved.id ? saved : current);\n")
		builder.WriteString("      } else {\n")
		builder.WriteString(fmt.Sprintf("        const saved = await %sApi.create(input);\n", strings.ToLower(entity.Name)))
		builder.WriteString("        setItems((current) => [...current, saved]);\n")
		builder.WriteString("      }\n")
		builder.WriteString("      resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to save record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "edit") {
		builder.WriteString(fmt.Sprintf("  function editItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		if hasState {
			builder.WriteString(g.closePageModals(state))
		}
		builder.WriteString("    setEditingId(item.id);\n")
		builder.WriteString("    setTouchedFields({});\n")
		builder.WriteString("    setSubmitted(false);\n")
		builder.WriteString("    setForm({\n")
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			itemValue := fmt.Sprintf("item.%s", field.Name)
			if g.isRelationField(field) {
				itemValue = fmt.Sprintf("item.%s", relationIDFieldName(field))
			}
			builder.WriteString(fmt.Sprintf("      %s: String(%s ?? \"\"),\n", fieldName, itemValue))
		}
		builder.WriteString("    });\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "delete") {
		builder.WriteString("  async function deleteItem(id: string) {\n")
		builder.WriteString("    if (!canDelete) return;\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      await %sApi.delete(id);\n", strings.ToLower(entity.Name)))
		builder.WriteString("      setItems((current) => current.filter((item) => item.id !== id));\n")
		builder.WriteString("      setSelectedIds((current) => current.filter((value) => value !== id));\n")
		builder.WriteString("      setSelectedItem((current) => current?.id === id ? null : current);\n")
		builder.WriteString("      if (editingId === id) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to delete record\");\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "delete") {
		builder.WriteString("  async function bulkDeleteSelected() {\n")
		builder.WriteString("    if (!canDelete) return;\n")
		builder.WriteString("    if (selectedIds.length === 0) return;\n")
		builder.WriteString("    const ids = [...selectedIds];\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      await %sApi.bulkDelete(ids);\n", strings.ToLower(entity.Name)))
		builder.WriteString("      setItems((current) => current.filter((item) => !ids.includes(item.id)));\n")
		builder.WriteString("      setSelectedIds((current) => current.filter((id) => !ids.includes(id)));\n")
		builder.WriteString("      setSelectedItem((current) => current && ids.includes(current.id) ? null : current);\n")
		builder.WriteString("      if (editingId && ids.includes(editingId)) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to delete selected records\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "archive") {
		builder.WriteString(fmt.Sprintf("  async function archiveItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      const archived = await %sApi.archive(item.id);\n", strings.ToLower(entity.Name)))
		builder.WriteString("      if (showArchived) {\n")
		builder.WriteString("        setItems((current) => current.map((existing) => existing.id === item.id ? archived : existing));\n")
		builder.WriteString("      } else {\n")
		builder.WriteString("        setItems((current) => current.filter((existing) => existing.id !== item.id));\n")
		builder.WriteString("      }\n")
		builder.WriteString("      setSelectedIds((current) => current.filter((id) => id !== item.id));\n")
		builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? archived : current);\n")
		builder.WriteString("      if (editingId === item.id) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to archive record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "restore") {
		builder.WriteString(fmt.Sprintf("  async function restoreItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      const restored = await %sApi.restore(item.id);\n", strings.ToLower(entity.Name)))
		builder.WriteString("      setItems((current) => current.map((existing) => existing.id === item.id ? restored : existing));\n")
		builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? restored : current);\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to restore record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString(g.workflowPageActionFunctions(entity))
	builder.WriteString("  return (\n")
	builder.WriteString("    <main>\n")
	builder.WriteString("      <header>\n")
	builder.WriteString(fmt.Sprintf("        <h1>%s</h1>\n", page.Name))
	builder.WriteString(fmt.Sprintf("        <span>%s source: %s</span>\n", g.program.App.Name, entity.Name))
	builder.WriteString("      </header>\n\n")
	builder.WriteString("      <section className=\"panel\">\n")
	builder.WriteString("        <div className=\"toolbar\">\n")
	builder.WriteString("          <input\n")
	builder.WriteString("        value={search}\n")
	builder.WriteString("        onChange={(event) => setSearch(event.target.value)}\n")
	builder.WriteString("        placeholder=\"Search\"\n")
	builder.WriteString("      />\n\n")
	if hasAction(page, "archive") || hasAction(page, "restore") {
		builder.WriteString("          <label className=\"inline-control\"><input checked={showArchived} type=\"checkbox\" onChange={(event) => setShowArchived(event.target.checked)} /> Show archived</label>\n")
	}
	if hasAction(page, "delete") {
		builder.WriteString("          {canDelete && <button className=\"danger\" type=\"button\" disabled={saving || selectedIds.length === 0} onClick={bulkDeleteSelected}>Delete Selected ({selectedIds.length})</button>}\n")
	}
	if hasState && hasAction(page, "create") {
		builder.WriteString(g.openCreateModalButton(state, entity))
	}
	builder.WriteString("        </div>\n")
	if len(page.Table.Filters) > 0 {
		builder.WriteString("        <div className=\"toolbar filter-controls\">\n")
		for _, fieldName := range page.Table.Filters {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			builder.WriteString(fmt.Sprintf("          <label>\n            %s Filter\n            <input type=\"text\" value={filters.%s} onChange={(event) => updateFilter(%q, event.target.value)} placeholder=\"Filter %s\" />\n          </label>\n", fieldLabel(field), field.Name, field.Name, fieldLabel(field)))
		}
		builder.WriteString("        </div>\n")
	}
	builder.WriteString("        <div className=\"toolbar column-controls\">\n")
	builder.WriteString("          <span className=\"muted\">Columns</span>\n")
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if !ok {
			continue
		}
		builder.WriteString(fmt.Sprintf("          {permissions.fields.%s !== false && <label className=\"inline-control\"><input checked={visibleColumns.%s} type=\"checkbox\" onChange={() => toggleColumn(%q)} /> %s</label>}\n", field.Name, field.Name, field.Name, fieldLabel(field)))
	}
	builder.WriteString("        </div>\n")
	builder.WriteString("        {error && <div className=\"error\" role=\"alert\">{error}</div>}\n")
	builder.WriteString("        {loading && <div className=\"status\">Loading records...</div>}\n")
	builder.WriteString(g.paginationToolbar(page.Table.Paginate, "        "))
	builder.WriteString("        <table>\n")
	builder.WriteString("        <thead>\n")
	builder.WriteString("          <tr>\n")
	if hasAction(page, "delete") {
		builder.WriteString("            {canDelete && <th className=\"select-cell\"><input aria-label=\"Select visible records\" checked={allVisibleSelected} type=\"checkbox\" onChange={toggleVisibleSelection} /></th>}\n")
	}
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if !ok {
			continue
		}
		builder.WriteString(fmt.Sprintf("            {visibleColumns.%s && permissions.fields.%s !== false && <th>%s</th>}\n", field.Name, field.Name, fieldLabel(field)))
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		builder.WriteString("            <th>Status</th>\n")
	}
	if g.hasRowActions(page, entity) {
		builder.WriteString("            <th>Actions</th>\n")
	}
	builder.WriteString("          </tr>\n")
	builder.WriteString("        </thead>\n")
	builder.WriteString("        <tbody>\n")
	builder.WriteString("          {!loading && visibleItems.length === 0 && (\n")
	builder.WriteString(fmt.Sprintf("            <tr><td colSpan={tableColspan}>No %s records yet.</td></tr>\n", strings.ToLower(entity.Name)))
	builder.WriteString("          )}\n")
	builder.WriteString("          {paginatedItems.map((item) => (\n")
	builder.WriteString("            <tr key={item.id}>\n")
	if hasAction(page, "delete") {
		builder.WriteString("              {canDelete && <td className=\"select-cell\"><input aria-label=\"Select record\" checked={selectedIds.includes(item.id)} type=\"checkbox\" onChange={() => toggleItemSelection(item.id)} /></td>}\n")
	}
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if !ok {
			continue
		}
		builder.WriteString(fmt.Sprintf("              {visibleColumns.%s && permissions.fields.%s !== false && <td>{%s}</td>}\n", field.Name, field.Name, g.itemRenderExpression("item", field)))
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		builder.WriteString("              <td>{item.archivedAt ? \"Archived\" : \"Active\"}</td>\n")
	}
	if g.hasRowActions(page, entity) {
		builder.WriteString("              <td>\n")
		builder.WriteString("                <div className=\"actions\">\n")
		builder.WriteString("                  <button className=\"secondary\" type=\"button\" onClick={() => viewItem(item.id)}>View</button>\n")
		if hasAction(page, "archive") {
			builder.WriteString("                  {canUpdate && !item.archivedAt && <button className=\"secondary\" type=\"button\" onClick={() => archiveItem(item)}>Archive</button>}\n")
		}
		if hasAction(page, "restore") {
			builder.WriteString("                  {canUpdate && item.archivedAt && <button className=\"secondary\" type=\"button\" onClick={() => restoreItem(item)}>Restore</button>}\n")
		}
		if hasAction(page, "edit") {
			builder.WriteString("                  {canUpdate && <button className=\"secondary\" type=\"button\" onClick={() => editItem(item)}>Edit</button>}\n")
		}
		if hasAction(page, "delete") {
			builder.WriteString("                  {canDelete && <button className=\"danger\" type=\"button\" onClick={() => deleteItem(item.id)}>Delete</button>}\n")
		}
		builder.WriteString(g.workflowPageActionButtons(entity))
		builder.WriteString("                </div>\n")
		builder.WriteString("              </td>\n")
	}
	builder.WriteString("            </tr>\n")
	builder.WriteString("          ))}\n")
	builder.WriteString("        </tbody>\n")
	builder.WriteString("      </table>\n\n")
	builder.WriteString(g.paginationToolbar(page.Table.Paginate, "        "))
	builder.WriteString("      </section>\n\n")
	builder.WriteString("      <section className=\"panel\">\n")
	builder.WriteString(fmt.Sprintf("        <h2>%s Details</h2>\n", entity.Name))
	builder.WriteString("        {reading && <div className=\"status\">Loading details...</div>}\n")
	builder.WriteString("        {!reading && !selectedItem && <p className=\"muted\">Select a record to view details.</p>}\n")
	builder.WriteString("        {selectedItem && (\n")
	builder.WriteString("          <dl className=\"detail-grid\">\n")
	builder.WriteString("            <div><dt>ID</dt><dd>{selectedItem.id}</dd></div>\n")
	builder.WriteString("            <div><dt>Status</dt><dd>{selectedItem.archivedAt ? \"Archived\" : \"Active\"}</dd></div>\n")
	for _, field := range entity.Fields {
		builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && <div><dt>%s</dt><dd>{%s}</dd></div>}\n", field.Name, fieldLabel(field), g.itemRenderExpression("selectedItem", field)))
	}
	builder.WriteString("          </dl>\n")
	builder.WriteString("        )}\n")
	builder.WriteString("      </section>\n\n")
	if hasAction(page, "create") || hasAction(page, "edit") {
		builder.WriteString(fmt.Sprintf("      {%s && (\n", g.formVisibleExpression(page, entity, state, hasState)))
		builder.WriteString("      <section className=\"panel\">\n")
		builder.WriteString(fmt.Sprintf("        <h2>{editingId ? \"Edit %s\" : \"Create %s\"}</h2>\n", entity.Name, entity.Name))
		builder.WriteString("        {missingRequiredRelations && (\n")
		builder.WriteString("          <div className=\"status relation-status\">\n")
		builder.WriteString(fmt.Sprintf("            <span>%s</span>\n", g.missingRequiredRelationsMessage(page, entity)))
		builder.WriteString(g.relationNavigationButtons(page, entity))
		builder.WriteString("          </div>\n")
		builder.WriteString("        )}\n")
		builder.WriteString("        <form noValidate onSubmit={saveItem}>\n")
		builder.WriteString("          <div className=\"form-grid\">\n")
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			if g.isRelationField(field) {
				optionsName := relationOptionsStateName(field)
				builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && (editingId ? permissions.writeFields.%s !== false : canCreate) && <label>\n              %s\n              <select%s disabled={%s.length === 0} value={form.%s} onChange={(event) => updateField(\"%s\", event.target.value)}>\n                <option value=\"\">%s</option>\n                {%s.map((option) => (\n                  <option key={option.id} value={option.id}>{%s}</option>\n                ))}\n              </select>\n              {visibleFormErrors.%s && <span className=\"field-error\">{visibleFormErrors.%s}</span>}\n              %s{%s.length === 0 && <span className=\"field-note\">Create a %s record before selecting %s.</span>}\n            </label>}\n", field.Name, field.Name, fieldLabel(field), selectAttributes(field), optionsName, field.Name, field.Name, fieldPlaceholder(field, "Select "+fieldLabel(field)), optionsName, g.relationOptionLabelExpression("option", field.Type), field.Name, field.Name, helpElement(field), optionsName, field.Type, fieldLabel(field)))
				continue
			}
			builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && (editingId ? permissions.writeFields.%s !== false : canCreate) && <label>\n              %s\n              <input%s%s value={form.%s} onChange={(event) => updateField(\"%s\", event.target.value)} />\n              {visibleFormErrors.%s && <span className=\"field-error\">{visibleFormErrors.%s}</span>}\n              %s%s</label>}\n", field.Name, field.Name, fieldLabel(field), inputAttributes(field), placeholderAttribute(field), field.Name, field.Name, field.Name, field.Name, g.formComponentPreview(field), helpElement(field)))
		}
		builder.WriteString("          </div>\n")
		builder.WriteString("          <div className=\"toolbar\">\n")
		builder.WriteString("            <button type=\"submit\" disabled={saving || missingRequiredRelations}>{saving ? \"Saving...\" : editingId ? \"Save Changes\" : \"Create\"}</button>\n")
		builder.WriteString(fmt.Sprintf("            {%s && <button className=\"secondary\" type=\"button\" onClick={resetForm}>Cancel</button>}\n", g.cancelVisibleExpression(entity, state, hasState)))
		builder.WriteString("          </div>\n")
		builder.WriteString("        </form>\n")
		builder.WriteString("      </section>\n")
		builder.WriteString("      )}\n")
	}
	builder.WriteString("    </main>\n")
	builder.WriteString("  );\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) searchExpression(fields []string, entity EntityDecl) string {
	if len(fields) == 0 {
		return "      false\n"
	}
	lines := []string{}
	for _, fieldName := range fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("      %s.toLowerCase().includes(query)", g.itemDisplayExpression("item", field)))
	}
	if len(lines) == 0 {
		return "      false\n"
	}
	return strings.Join(lines, " ||\n") + "\n"
}

func (g *webGenerator) sortExpression(sort SortDecl, entity EntityDecl) string {
	if sort.Field == "" {
		return "    return filteredItems;\n"
	}
	field, ok := findField(entity, sort.Field)
	if !ok {
		return "    return filteredItems;\n"
	}
	direction := 1
	if sort.Direction == "desc" {
		direction = -1
	}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("    return [...filteredItems].sort((left, right) => (Number(left.%s ?? 0) - Number(right.%s ?? 0)) * %d);\n", field.Name, field.Name, direction)
	default:
		return fmt.Sprintf("    return [...filteredItems].sort((left, right) => %s.localeCompare(%s) * %d);\n", g.itemDisplayExpression("left", field), g.itemDisplayExpression("right", field), direction)
	}
}

func (g *webGenerator) filterExpression(filters []string, entity EntityDecl) string {
	if len(filters) == 0 {
		return "    const filteredItems = searchedItems;\n"
	}
	lines := []string{}
	for _, fieldName := range filters {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("      (filters.%s.trim() === \"\" || %s.toLowerCase().includes(filters.%s.trim().toLowerCase()))", field.Name, g.itemDisplayExpression("item", field), field.Name))
	}
	if len(lines) == 0 {
		return "    const filteredItems = searchedItems;\n"
	}
	return "    const filteredItems = searchedItems.filter((item) =>\n" + strings.Join(lines, " &&\n") + "\n    );\n"
}

func (g *webGenerator) paginationToolbar(pageSize int, indent string) string {
	if pageSize <= 0 {
		return ""
	}
	return fmt.Sprintf("%s<div className=\"toolbar pagination\">\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage <= 1} onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}>Previous</button>\n%s  <span className=\"muted\">Page {safeCurrentPage} of {totalPages}</span>\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage >= totalPages} onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}>Next</button>\n%s</div>\n", indent, indent, indent, indent, indent)
}

func (g *webGenerator) formValidationFunction(page PageDecl, entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString("function validateForm(form: FormState): FormErrors {\n")
	builder.WriteString("  const errors: FormErrors = {};\n\n")
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		builder.WriteString(g.formValidationFieldBlock(field))
	}
	builder.WriteString(g.formEntityValidationBlocks(entity, page.Form.Fields))
	builder.WriteString("  return errors;\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) formValidationFieldBlock(field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	label := fieldLabel(field)
	if g.isRelationField(field) {
		if hasModifier(field, "required") {
			builder.WriteString(fmt.Sprintf("  if (form.%s.trim() === \"\") {\n", name))
			builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", name, fieldValidationMessage(field, label+" is required")))
			builder.WriteString("  }\n\n")
		}
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("  if (form.%s.trim() === \"\") {\n", name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString("    // Empty input will use the field default on save.\n")
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", name, fieldValidationMessage(field, label+" is required")))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer", "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(form.%s);\n", name))
		builder.WriteString(fmt.Sprintf("    if (Number.isNaN(parsed)) errors.%s = %q;\n", name, fieldValidationMessage(field, label+" must be a number")))
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed < %s) errors.%s = %q;\n", minValue, name, fieldValidationMessage(field, label+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed > %s) errors.%s = %q;\n", maxValue, name, fieldValidationMessage(field, label+" must be at most "+maxValue)))
		}
	case "email":
		builder.WriteString(fmt.Sprintf("    if (!form.%s.includes(\"@\")) errors.%s = %q;\n", name, name, fieldValidationMessage(field, label+" must be an email")))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    else if (form.%s.length < %d || form.%s.length > %d) errors.%s = %q;\n", name, minLength, name, maxLength, name, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s && !(new RegExp(%q)).test(form.%s)) errors.%s = %q;\n", name, pattern, name, name, fieldValidationMessage(field, label+" has an invalid format")))
		}
	default:
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (form.%s.length < %d || form.%s.length > %d) errors.%s = %q;\n", name, minLength, name, maxLength, name, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
		}
		if hasModifier(field, "url") {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s) {\n", name))
			builder.WriteString(fmt.Sprintf("      try { new URL(form.%s); } catch { errors.%s = %q; }\n", name, name, fieldValidationMessage(field, label+" must be a valid URL")))
			builder.WriteString("    }\n")
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s && !(new RegExp(%q)).test(form.%s)) errors.%s = %q;\n", name, pattern, name, name, fieldValidationMessage(field, label+" has an invalid format")))
		}
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func emptyStateColspan(page PageDecl) int {
	colspan := len(page.Table.Columns)
	if len(page.Actions) > 0 {
		colspan++
	}
	if hasAction(page, "delete") {
		colspan++
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		colspan++
	}
	return colspan
}

func (g *webGenerator) hasRowActions(page PageDecl, entity EntityDecl) bool {
	return len(page.Actions) > 0 || len(g.workflowsForEntity(entity.Name)) > 0
}

func staticTableExtraColumns(page PageDecl) int {
	count := 0
	if len(page.Actions) > 0 {
		count++
	}
	if hasAction(page, "delete") {
		count++
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		count++
	}
	return count
}

func (g *webGenerator) staticTableExtraColumns(page PageDecl, entity EntityDecl) int {
	count := staticTableExtraColumns(page)
	if len(page.Actions) == 0 && len(g.workflowsForEntity(entity.Name)) > 0 {
		count++
	}
	return count
}

func relationIDFieldName(field FieldDecl) string {
	return field.Name + "Id"
}

func relationOptionsStateName(field FieldDecl) string {
	return field.Name + "Options"
}

func relationName(entity EntityDecl, field FieldDecl) string {
	return entity.Name + "_" + field.Name
}

func relationBackFieldName(entity EntityDecl, field FieldDecl) string {
	return strings.ToLower(entity.Name[:1]) + entity.Name[1:] + title(field.Name) + "Items"
}

func (g *webGenerator) typeImportList(page PageDecl, entity EntityDecl) string {
	names := []string{entity.Name}
	seen := map[string]bool{entity.Name: true}
	for _, field := range g.relationFields(entity) {
		if seen[field.Type] {
			continue
		}
		seen[field.Type] = true
		names = append(names, field.Type)
	}
	if state, ok := g.stateForPage(page); ok {
		for _, field := range state.Fields {
			if _, exists := g.findEntity(field.Type); !exists || seen[field.Type] {
				continue
			}
			seen[field.Type] = true
			names = append(names, field.Type)
		}
	}
	return strings.Join(names, ", ")
}

func (g *webGenerator) stateForPage(page PageDecl) (StateDecl, bool) {
	expectedNames := map[string]bool{
		page.Name + "State":     true,
		page.Name + "PageState": true,
	}
	for _, state := range g.program.States {
		if expectedNames[state.Name] {
			return state, true
		}
	}
	return StateDecl{}, false
}

func (g *webGenerator) pageStateHooks(state StateDecl) string {
	var builder strings.Builder
	for _, field := range state.Fields {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState<%s>(%s);\n", field.Name, title(field.Name), g.stateFieldTSType(field), stateFieldDefaultValue(field)))
	}
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState(%t);\n", modalStateName(modal), title(modalStateName(modal)), modal.Default == "open"))
	}
	return builder.String()
}

func (g *webGenerator) stateFieldTSType(field StateField) string {
	fieldType := tsType(field.Type)
	if _, ok := g.findEntity(field.Type); ok {
		fieldType = field.Type
	}
	if field.List {
		return fieldType + "[]"
	}
	return fieldType
}

func stateFieldDefaultValue(field StateField) string {
	if field.List {
		return "[]"
	}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return "0"
	case "boolean":
		return "false"
	default:
		return "\"\""
	}
}

func (g *webGenerator) pageModalHelpers(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		stateName := modalStateName(modal)
		builder.WriteString(fmt.Sprintf("  function open%s() {\n", title(modal.Name)))
		builder.WriteString(fmt.Sprintf("    set%s(true);\n", title(stateName)))
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  function close%s() {\n", title(modal.Name)))
		builder.WriteString(fmt.Sprintf("    set%s(false);\n", title(stateName)))
		builder.WriteString("  }\n\n")
	}
	return builder.String()
}

func (g *webGenerator) resetPageModals(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("    close%s();\n", title(modal.Name)))
	}
	return builder.String()
}

func (g *webGenerator) closePageModals(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("    close%s();\n", title(modal.Name)))
	}
	return builder.String()
}

func (g *webGenerator) openCreateModalButton(state StateDecl, entity EntityDecl) string {
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name {
			return fmt.Sprintf("          {canCreate && !editingId && !%s && <button className=\"secondary\" type=\"button\" onClick={open%s}>New %s</button>}\n", modalStateName(modal), title(modal.Name), entity.Name)
		}
	}
	return ""
}

func (g *webGenerator) formVisibleExpression(page PageDecl, entity EntityDecl, state StateDecl, hasState bool) string {
	base := "(canCreate || canUpdate)"
	if !hasState {
		return base
	}
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name && hasAction(page, "create") {
			return fmt.Sprintf("((canUpdate && editingId) || (canCreate && %s))", modalStateName(modal))
		}
	}
	return base
}

func (g *webGenerator) cancelVisibleExpression(entity EntityDecl, state StateDecl, hasState bool) string {
	if !hasState {
		return "editingId"
	}
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name {
			return "(editingId || " + modalStateName(modal) + ")"
		}
	}
	return "editingId"
}

func modalStateName(modal StateModal) string {
	return modal.Name + "Open"
}

func (g *webGenerator) missingRequiredRelationsExpression(page PageDecl, entity EntityDecl) string {
	parts := []string{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s.length === 0", relationOptionsStateName(field)))
	}
	if len(parts) == 0 {
		return "false"
	}
	return strings.Join(parts, " || ")
}

func (g *webGenerator) missingRequiredRelationsMessage(page PageDecl, entity EntityDecl) string {
	targets := []string{}
	seen := map[string]bool{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") || seen[field.Type] {
			continue
		}
		seen[field.Type] = true
		targets = append(targets, field.Type)
	}
	if len(targets) == 0 {
		return ""
	}
	if len(targets) == 1 {
		return fmt.Sprintf("Create a %s record before creating %s.", targets[0], entity.Name)
	}
	return fmt.Sprintf("Create required related records (%s) before creating %s.", strings.Join(targets, ", "), entity.Name)
}

func (g *webGenerator) relationNavigationButtons(page PageDecl, entity EntityDecl) string {
	buttons := []string{}
	seen := map[string]bool{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") || seen[field.Type] {
			continue
		}
		targetPage, ok := g.pageForEntity(field.Type)
		if !ok {
			continue
		}
		seen[field.Type] = true
		buttons = append(buttons, fmt.Sprintf("              <button className=\"secondary\" type=\"button\" onClick={() => onNavigate(%q)}>Open %s</button>", targetPage.Name, targetPage.Name))
	}
	if len(buttons) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("            {onNavigate && (\n")
	builder.WriteString("              <div className=\"actions\">\n")
	for _, button := range buttons {
		builder.WriteString(button)
		builder.WriteString("\n")
	}
	builder.WriteString("              </div>\n")
	builder.WriteString("            )}\n")
	return builder.String()
}

func (g *webGenerator) pageForEntity(entityName string) (PageDecl, bool) {
	for _, page := range g.program.Pages {
		if page.Source == entityName {
			return page, true
		}
	}
	return PageDecl{}, false
}

func (g *webGenerator) navigationPages() []PageDecl {
	if len(g.program.Layouts) == 0 || len(g.program.Layouts[0].Sidebar.Items) == 0 {
		return g.program.Pages
	}

	pageIndex := map[string]PageDecl{}
	for _, page := range g.program.Pages {
		pageIndex[page.Name] = page
	}

	ordered := []PageDecl{}
	seen := map[string]bool{}
	for _, item := range g.program.Layouts[0].Sidebar.Items {
		page, ok := pageIndex[item]
		if !ok || seen[item] {
			continue
		}
		seen[item] = true
		ordered = append(ordered, page)
	}
	for _, page := range g.program.Pages {
		if seen[page.Name] {
			continue
		}
		ordered = append(ordered, page)
	}
	if len(ordered) == 0 {
		return g.program.Pages
	}
	return ordered
}

func (g *webGenerator) itemDisplayExpression(itemName string, field FieldDecl) string {
	if !g.isRelationField(field) {
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, field.Name)
	}
	target, ok := g.findEntity(field.Type)
	if !ok {
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, relationIDFieldName(field))
	}
	labelField := relationLabelField(target)
	return fmt.Sprintf("String(%s.%s?.%s ?? %s.%s ?? \"\")", itemName, field.Name, labelField, itemName, relationIDFieldName(field))
}

func (g *webGenerator) relationOptionLabelExpression(optionName string, entityName string) string {
	target, ok := g.findEntity(entityName)
	if !ok {
		return fmt.Sprintf("String(%s.id)", optionName)
	}
	labelField := relationLabelField(target)
	return fmt.Sprintf("String(%s.%s ?? %s.id)", optionName, labelField, optionName)
}

func relationLabelField(entity EntityDecl) string {
	for _, preferred := range []string{"name", "title", "email", "sku"} {
		for _, field := range entity.Fields {
			if field.Name == preferred {
				return field.Name
			}
		}
	}
	if len(entity.Fields) > 0 {
		return entity.Fields[0].Name
	}
	return "id"
}

func tsType(fieldType string) string {
	switch fieldType {
	case "number", "integer", "decimal", "money":
		return "number"
	case "boolean":
		return "boolean"
	case "date", "datetime":
		return "string"
	default:
		return "string"
	}
}

func prismaType(fieldType string) string {
	switch fieldType {
	case "number", "integer":
		return "Int"
	case "decimal", "money":
		return "Float"
	case "boolean":
		return "Boolean"
	case "date", "datetime":
		return "DateTime"
	default:
		return "String"
	}
}

func sqliteType(fieldType string) string {
	switch fieldType {
	case "number", "integer":
		return "INTEGER"
	case "decimal", "money":
		return "REAL"
	case "boolean":
		return "INTEGER"
	case "date", "datetime":
		return "DATETIME"
	default:
		return "TEXT"
	}
}

func sqliteRequired(field FieldDecl) string {
	if hasModifier(field, "required") || hasModifier(field, "default") {
		return " NOT NULL"
	}
	return ""
}

func sqliteDefaultValue(field FieldDecl, value string) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return value
	case "boolean":
		if value == "true" {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%q", value)
	}
}

func prismaOptional(field FieldDecl) string {
	if hasModifier(field, "required") || hasModifier(field, "default") {
		return ""
	}
	return "?"
}

func prismaRelationIDOptional(field FieldDecl) string {
	if hasModifier(field, "required") {
		return ""
	}
	return "?"
}

func prismaAttributes(field FieldDecl) string {
	attributes := []string{}
	if hasModifier(field, "unique") {
		attributes = append(attributes, "@unique")
	}
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		attributes = append(attributes, fmt.Sprintf("@default(%s)", prismaDefaultValue(field, defaultValue)))
	}
	if len(attributes) == 0 {
		return ""
	}
	return " " + strings.Join(attributes, " ")
}

func prismaDefaultValue(field FieldDecl, value string) string {
	switch field.Type {
	case "text", "email", "date", "datetime":
		return fmt.Sprintf("%q", value)
	default:
		return value
	}
}

func hasModifier(field FieldDecl, name string) bool {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return true
		}
	}
	return false
}

func hasAction(page PageDecl, name string) bool {
	for _, action := range page.Actions {
		if action == name {
			return true
		}
	}
	return false
}

func (g *webGenerator) defaultAuthRole() string {
	if len(g.program.Roles) > 0 {
		return g.program.Roles[0].Name
	}
	return "authenticated"
}

func (g *webGenerator) roleNames() []string {
	names := []string{}
	for _, role := range g.program.Roles {
		names = append(names, role.Name)
	}
	return names
}

func (g *webGenerator) hasRuntimePermissions() bool {
	return g.program.Auth != nil && len(g.program.Roles) > 0
}

func (g *webGenerator) permissionMiddleware(action string, resource string) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	return fmt.Sprintf("requirePermission(%q, %q), ", action, resource)
}

func (g *webGenerator) rolePermissionsLiteral() string {
	type generatedPermission struct {
		Effect   string   `json:"effect"`
		Action   string   `json:"action"`
		Resource string   `json:"resource"`
		Fields   []string `json:"fields"`
	}
	permissions := map[string][]generatedPermission{}
	for _, role := range g.program.Roles {
		for _, permission := range role.Permissions {
			fields := permission.Fields
			if fields == nil {
				fields = []string{}
			}
			permissions[role.Name] = append(permissions[role.Name], generatedPermission{
				Effect:   permission.Effect,
				Action:   permission.Action,
				Resource: permission.Resource,
				Fields:   fields,
			})
		}
	}
	content, err := json.Marshal(permissions)
	if err != nil {
		return "{}"
	}
	return string(content)
}

func tsStringArrayLiteral(values []string) string {
	parts := []string{}
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%q", value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func fieldNames(entity EntityDecl) []string {
	names := []string{}
	for _, field := range entity.Fields {
		names = append(names, field.Name)
	}
	return names
}

func findField(entity EntityDecl, name string) (FieldDecl, bool) {
	for _, field := range entity.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldDecl{}, false
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func componentInputByName(inputs []ComponentInput, name string) (ComponentInput, bool) {
	for _, input := range inputs {
		if input.Name == name {
			return input, true
		}
	}
	return ComponentInput{}, false
}

func isNumericLiteral(value string) bool {
	if value == "" {
		return false
	}
	dotSeen := false
	for index, char := range value {
		if char == '-' && index == 0 {
			continue
		}
		if char == '.' && !dotSeen {
			dotSeen = true
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != "-" && value != "." && value != "-."
}

func kebabCase(value string) string {
	if value == "" {
		return value
	}
	var builder strings.Builder
	for index, char := range value {
		if index > 0 && char >= 'A' && char <= 'Z' {
			builder.WriteByte('-')
		}
		builder.WriteRune(char)
	}
	return strings.ToLower(builder.String())
}

func workflowActionFunctionName(transition TransitionDecl) string {
	return "run" + title(transition.Name) + "Workflow"
}

func identifierLabel(value string) string {
	if value == "" {
		return value
	}
	words := []rune{}
	for index, char := range value {
		if index > 0 && char >= 'A' && char <= 'Z' {
			words = append(words, ' ')
		}
		words = append(words, char)
	}
	return title(string(words))
}

func formStateLiteral(fields []string, entity EntityDecl) string {
	if len(fields) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, fieldName := range fields {
		value := ""
		if field, ok := findField(entity, fieldName); ok {
			value = modifierValue(field, "default")
		}
		parts = append(parts, fmt.Sprintf("%s: %q", fieldName, value))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func columnVisibilityLiteral(columns []string) string {
	if len(columns) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, column := range columns {
		parts = append(parts, fmt.Sprintf("%s: true", column))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func tableFiltersLiteral(filters []string) string {
	if len(filters) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, filter := range filters {
		parts = append(parts, fmt.Sprintf("%s: %q", filter, ""))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func formValueExpression(field FieldDecl) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		if defaultValue := modifierValue(field, "default"); defaultValue != "" {
			return fmt.Sprintf("form.%s === \"\" ? %s : Number(form.%s)", field.Name, defaultValue, field.Name)
		}
		if hasModifier(field, "required") {
			return fmt.Sprintf("Number(form.%s)", field.Name)
		}
		return fmt.Sprintf("form.%s === \"\" ? undefined : Number(form.%s)", field.Name, field.Name)
	case "boolean":
		return fmt.Sprintf("form.%s === \"true\"", field.Name)
	default:
		return fmt.Sprintf("form.%s", field.Name)
	}
}

func fieldLabel(field FieldDecl) string {
	if label := modifierValue(field, "label"); label != "" {
		return label
	}
	return title(field.Name)
}

func fieldPlaceholder(field FieldDecl, fallback string) string {
	if placeholder := modifierValue(field, "placeholder"); placeholder != "" {
		return placeholder
	}
	return fallback
}

func placeholderAttribute(field FieldDecl) string {
	if placeholder := modifierValue(field, "placeholder"); placeholder != "" {
		return fmt.Sprintf(" placeholder=%q", placeholder)
	}
	return ""
}

func helpElement(field FieldDecl) string {
	if help := modifierValue(field, "help"); help != "" {
		return fmt.Sprintf("<span className=\"field-note\">%s</span>\n              ", help)
	}
	return ""
}

func inputAttributes(field FieldDecl) string {
	attributes := []string{}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		attributes = append(attributes, " type=\"number\"")
		if minValue := modifierValue(field, "min"); minValue != "" {
			attributes = append(attributes, fmt.Sprintf(" min=%q", minValue))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			attributes = append(attributes, fmt.Sprintf(" max=%q", maxValue))
		}
	case "email":
		attributes = append(attributes, " type=\"email\"")
	case "date":
		attributes = append(attributes, " type=\"date\"")
	case "datetime":
		attributes = append(attributes, " type=\"datetime-local\"")
	default:
		if hasModifier(field, "url") {
			attributes = append(attributes, " type=\"url\"")
		} else {
			attributes = append(attributes, " type=\"text\"")
		}
	}
	if hasModifier(field, "required") {
		attributes = append(attributes, " required")
	}
	if minLength, maxLength, ok := fieldLengthBounds(field); ok {
		attributes = append(attributes, fmt.Sprintf(" minLength={%d}", minLength))
		attributes = append(attributes, fmt.Sprintf(" maxLength={%d}", maxLength))
	}
	if pattern := modifierValue(field, "regex"); pattern != "" {
		attributes = append(attributes, fmt.Sprintf(" pattern=%q", pattern))
	}
	return strings.Join(attributes, "")
}

func fieldLengthBounds(field FieldDecl) (int, int, bool) {
	value := modifierValue(field, "length")
	if value == "" {
		return 0, 0, false
	}
	return parseLengthConstraint(value)
}

func selectAttributes(field FieldDecl) string {
	if hasModifier(field, "required") {
		return " required"
	}
	return ""
}

func validationFieldBlock(field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	builder.WriteString(fmt.Sprintf("  if (input.%s === undefined || input.%s === null || input.%s === \"\") {\n", name, name, name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString(fmt.Sprintf("    value.%s = %s;\n", name, typedLiteral(field, defaultValue)))
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.push(%q);\n", fieldValidationMessage(field, name+" is required")))
	} else {
		builder.WriteString(fmt.Sprintf("    value.%s = undefined;\n", name))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer", "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(input.%s);\n", name))
		builder.WriteString("    if (Number.isNaN(parsed)) {\n")
		builder.WriteString(fmt.Sprintf("      errors.push(%q);\n", fieldValidationMessage(field, name+" must be a number")))
		builder.WriteString("    } else {\n")
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed < %s) errors.push(%q);\n", minValue, fieldValidationMessage(field, name+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed > %s) errors.push(%q);\n", maxValue, fieldValidationMessage(field, name+" must be at most "+maxValue)))
		}
		builder.WriteString(fmt.Sprintf("      value.%s = parsed;\n", name))
		builder.WriteString("    }\n")
	case "boolean":
		builder.WriteString(fmt.Sprintf("    value.%s = Boolean(input.%s);\n", name, name))
	case "email":
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		builder.WriteString(fmt.Sprintf("    if (!value.%s.includes(\"@\")) errors.push(%q);\n", name, fieldValidationMessage(field, name+" must be an email")))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (value.%s.length < %d || value.%s.length > %d) errors.push(%q);\n", name, minLength, name, maxLength, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", name, minLength, maxLength))))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!(new RegExp(%q)).test(value.%s)) errors.push(%q);\n", pattern, name, fieldValidationMessage(field, name+" has an invalid format")))
		}
	default:
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (value.%s.length < %d || value.%s.length > %d) errors.push(%q);\n", name, minLength, name, maxLength, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", name, minLength, maxLength))))
		}
		if hasModifier(field, "url") {
			builder.WriteString(fmt.Sprintf("    try { new URL(value.%s); } catch { errors.push(%q); }\n", name, fieldValidationMessage(field, name+" must be a valid URL")))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!(new RegExp(%q)).test(value.%s)) errors.push(%q);\n", pattern, name, fieldValidationMessage(field, name+" has an invalid format")))
		}
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) entityValidationBlocks(entity EntityDecl) string {
	if len(entity.Validations) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, validation := range entity.Validations {
		if validation.Required && validation.When != nil {
			condition := g.validationConditionExpression(entity, *validation.When, "value")
			if condition == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("  if ((%s) && (value.%s === undefined || value.%s === null || String(value.%s).trim() === \"\")) errors.push(%q);\n", condition, validation.Left, validation.Left, validation.Left, entityValidationMessage(entity, validation)))
			continue
		}
		expression := g.entityValidationExpression(entity, validation, "value")
		if expression == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("  if (!(%s)) errors.push(%q);\n", expression, entityValidationMessage(entity, validation)))
	}
	if builder.Len() > 0 {
		builder.WriteString("\n")
	}
	return builder.String()
}

func (g *webGenerator) formEntityValidationBlocks(entity EntityDecl, formFields []string) string {
	if len(entity.Validations) == 0 {
		return ""
	}
	formFieldIndex := map[string]bool{}
	for _, field := range formFields {
		formFieldIndex[field] = true
	}
	var builder strings.Builder
	for _, validation := range entity.Validations {
		if validation.Required && validation.When != nil {
			if !formFieldIndex[validation.Left] || !formFieldIndex[validation.When.Left] {
				continue
			}
			condition := g.validationConditionExpression(entity, *validation.When, "form")
			if condition == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("  if ((%s) && form.%s.trim() === \"\") {\n", condition, validation.Left))
			builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", validation.Left, entityValidationMessage(entity, validation)))
			builder.WriteString("  }\n\n")
			continue
		}
		if !formFieldIndex[validation.Left] || !formFieldIndex[validation.Right] {
			continue
		}
		expression := g.entityValidationExpression(entity, validation, "form")
		if expression == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("  if (form.%s.trim() !== \"\" && form.%s.trim() !== \"\" && !(%s)) {\n", validation.Left, validation.Right, expression))
		builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", validation.Left, entityValidationMessage(entity, validation)))
		builder.WriteString("  }\n\n")
	}
	return builder.String()
}

func (g *webGenerator) validationConditionExpression(entity EntityDecl, condition ValidationConditionDecl, objectName string) string {
	left, leftOK := findField(entity, condition.Left)
	if !leftOK {
		return ""
	}
	if right, rightOK := findField(entity, condition.Right); rightOK {
		if numberLikeField(left) && numberLikeField(right) {
			return fmt.Sprintf("Number(%s.%s) %s Number(%s.%s)", objectName, condition.Left, condition.Operator, objectName, condition.Right)
		}
		if condition.Operator == "==" || condition.Operator == "!=" {
			return fmt.Sprintf("String(%s.%s) %s String(%s.%s)", objectName, condition.Left, condition.Operator, objectName, condition.Right)
		}
		return ""
	}
	if numberLikeField(left) {
		return fmt.Sprintf("Number(%s.%s) %s %s", objectName, condition.Left, condition.Operator, condition.Right)
	}
	if condition.Operator == "==" || condition.Operator == "!=" {
		return fmt.Sprintf("String(%s.%s) %s %q", objectName, condition.Left, condition.Operator, condition.Right)
	}
	return ""
}

func (g *webGenerator) entityValidationExpression(entity EntityDecl, validation EntityValidationDecl, objectName string) string {
	left, leftOK := findField(entity, validation.Left)
	right, rightOK := findField(entity, validation.Right)
	if !leftOK || !rightOK {
		return ""
	}
	if numberLikeField(left) && numberLikeField(right) {
		return fmt.Sprintf("Number(%s.%s) %s Number(%s.%s)", objectName, validation.Left, validation.Operator, objectName, validation.Right)
	}
	if validation.Operator == "==" || validation.Operator == "!=" {
		return fmt.Sprintf("String(%s.%s) %s String(%s.%s)", objectName, validation.Left, validation.Operator, objectName, validation.Right)
	}
	return ""
}

func entityValidationMessage(entity EntityDecl, validation EntityValidationDecl) string {
	if validation.Message != "" {
		return validation.Message
	}
	if validation.Required && validation.When != nil {
		return fmt.Sprintf("%s.%s is required when %s.%s %s %s", entity.Name, validation.Left, entity.Name, validation.When.Left, validation.When.Operator, validation.When.Right)
	}
	return fmt.Sprintf("%s.%s must be %s %s.%s", entity.Name, validation.Left, validation.Operator, entity.Name, validation.Right)
}

func typedLiteral(field FieldDecl, value string) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return value
	case "boolean":
		if value == "true" {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%q", value)
	}
}

func modifierValue(field FieldDecl, name string) string {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return modifier.Value
		}
	}
	return ""
}

func fieldValidationMessage(field FieldDecl, fallback string) string {
	if message := modifierValue(field, "message"); message != "" {
		return message
	}
	return fallback
}

func title(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
