package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type generatedOpsObservabilityManifest struct {
	Version        string                          `json:"version"`
	App            string                          `json:"app"`
	Provider       string                          `json:"provider"`
	EndpointEnv    string                          `json:"endpointEnv"`
	Protocol       string                          `json:"protocol"`
	Mode           string                          `json:"mode"`
	Signal         string                          `json:"signal"`
	TraceContext   string                          `json:"traceContext"`
	Delivery       string                          `json:"delivery"`
	Exporter       generatedOpsObservabilityExport `json:"exporter"`
	GeneratedFiles []string                        `json:"generatedFiles"`
	Policy         string                          `json:"policy"`
	Steps          []string                        `json:"steps"`
}

type generatedOpsObservabilityExport struct {
	Type        string `json:"type"`
	Format      string `json:"format"`
	NonBlocking bool   `json:"nonBlocking"`
}

func (g *webGenerator) hasOps() bool {
	return g.program.Ops != nil
}

func (g *webGenerator) hasOpsHealth() bool {
	return g.program.Ops != nil && g.program.Ops.Health != nil && g.program.Ops.Health.Path != ""
}

func (g *webGenerator) hasOpsReadiness() bool {
	return g.program.Ops != nil && g.program.Ops.Readiness != nil && g.program.Ops.Readiness.Path != ""
}

func (g *webGenerator) hasOpsMetrics() bool {
	return g.program.Ops != nil && g.program.Ops.Metrics != nil && g.program.Ops.Metrics.Path != ""
}

func (g *webGenerator) hasOpsRequestLogging() bool {
	return g.program.Ops != nil && g.program.Ops.Logging == "requests"
}

func (g *webGenerator) hasOpsObserve() bool {
	return g.program.Ops != nil && g.program.Ops.Observe != nil && g.program.Ops.Observe.Provider != "" && g.program.Ops.Observe.Endpoint.Name != ""
}

func (g *webGenerator) hasOpsOTLPObserve() bool {
	return g.hasOpsObserve() && g.opsObserveProvider() == "otlp"
}

func (g *webGenerator) opsObserveProvider() string {
	if !g.hasOpsObserve() {
		return ""
	}
	return g.program.Ops.Observe.Provider
}

func (g *webGenerator) opsObserveEndpointEnv() string {
	if !g.hasOpsObserve() {
		return ""
	}
	return g.program.Ops.Observe.Endpoint.Name
}

func (g *webGenerator) opsObserveProtocol() string {
	switch g.opsObserveProvider() {
	case "otlp":
		return "otlp-http-json"
	case "webhook":
		return "http-json"
	default:
		return ""
	}
}

func (g *webGenerator) opsObserveMode() string {
	switch g.opsObserveProvider() {
	case "otlp":
		return "async-otlp-http-json-traces"
	case "webhook":
		return "async-request-event-webhook"
	default:
		return ""
	}
}

func (g *webGenerator) opsObserveSignal() string {
	switch g.opsObserveProvider() {
	case "otlp":
		return "traces"
	case "webhook":
		return "request-events"
	default:
		return ""
	}
}

func (g *webGenerator) opsObservabilityManifestJSON() string {
	if !g.hasOpsObserve() {
		return "{}\n"
	}
	manifest := generatedOpsObservabilityManifest{
		Version:      version,
		App:          g.program.App.Name,
		Provider:     g.opsObserveProvider(),
		EndpointEnv:  g.opsObserveEndpointEnv(),
		Protocol:     g.opsObserveProtocol(),
		Mode:         g.opsObserveMode(),
		Signal:       g.opsObserveSignal(),
		TraceContext: "w3c-traceparent",
		Delivery:     "non-blocking",
		Exporter: generatedOpsObservabilityExport{
			Type:        g.opsObserveProvider(),
			Format:      g.opsObserveProtocol(),
			NonBlocking: true,
		},
		GeneratedFiles: []string{
			"src/server.ts",
			"openapi.json",
			"src/blacklang.contract.test.ts",
			"src/blacklang.api.test.ts",
			"security/secrets.json",
			".env.example",
		},
		Policy: "observability endpoints must be environment references and generated delivery must not block the user response path",
		Steps: []string{
			"Set the endpoint environment variable in the deployment environment.",
			"Run npm run security:secrets:plan to verify the endpoint reference without printing its value.",
			"Run npm test so generated contract/API smoke tests verify observability metadata and local delivery behavior.",
		},
	}
	if g.hasOpsOTLPObserve() {
		manifest.Steps = append(manifest.Steps, "For OTLP, point the endpoint at an OTLP HTTP traces receiver such as /v1/traces.")
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) opsHealthPath() string {
	if !g.hasOpsHealth() {
		return ""
	}
	return g.program.Ops.Health.Path
}

func (g *webGenerator) opsReadinessPath() string {
	if !g.hasOpsReadiness() {
		return ""
	}
	return g.program.Ops.Readiness.Path
}

func (g *webGenerator) opsMetricsPath() string {
	if !g.hasOpsMetrics() {
		return ""
	}
	return g.program.Ops.Metrics.Path
}

func (g *webGenerator) dockerHealthcheckCommand() string {
	path := g.opsHealthPath()
	if path == "" {
		return ""
	}
	return fmt.Sprintf(
		"fetch('http://127.0.0.1:' + (process.env[%s] || %s) + %s).then(function (response) { process.exit(response.ok ? 0 : 1); }).catch(function () { process.exit(1); })",
		contractJSONString(g.deployPortEnv()),
		contractJSONString(g.deployPortDefault()),
		contractJSONString(path),
	)
}

func (g *webGenerator) dockerfileHealthcheck() string {
	command := g.dockerHealthcheckCommand()
	if command == "" {
		return ""
	}
	return fmt.Sprintf("HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD node -e %s\n\n", contractJSONString(command))
}

func (g *webGenerator) composeHealthcheck(indent string) string {
	command := g.dockerHealthcheckCommand()
	if command == "" {
		return ""
	}
	return fmt.Sprintf("%shealthcheck:\n%stest: [\"CMD\", \"node\", \"-e\", %s]\n%sinterval: 30s\n%stimeout: 5s\n%sretries: 3\n", indent, indent+"  ", contractJSONString(command), indent+"  ", indent+"  ", indent+"  ")
}

func (g *webGenerator) opsServerState() string {
	if !g.hasOps() {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("const startedAt = new Date();\n")
	if g.hasOpsMetrics() {
		builder.WriteString("const opsCounters = { requests: 0, errors: 0 };\n")
		builder.WriteString("const opsStatusCounts = new Map<number, number>();\n")
	}
	if g.hasOpsObserve() {
		builder.WriteString(fmt.Sprintf("const opsObserveProvider = %s;\n", contractJSONString(g.opsObserveProvider())))
		builder.WriteString("function opsObserveEndpoint() {\n")
		builder.WriteString(fmt.Sprintf("  return process.env[%s] ?? \"\";\n", contractJSONString(g.opsObserveEndpointEnv())))
		builder.WriteString("}\n")
		builder.WriteString("type OpsTraceContext = {\n")
		builder.WriteString("  traceId: string;\n")
		builder.WriteString("  spanId: string;\n")
		builder.WriteString("  parentSpanId?: string;\n")
		builder.WriteString("  traceFlags: string;\n")
		builder.WriteString("  traceparent: string;\n")
		builder.WriteString("};\n")
		builder.WriteString("function opsRandomHex(bytes: number) {\n")
		builder.WriteString("  return crypto.randomBytes(bytes).toString(\"hex\");\n")
		builder.WriteString("}\n")
		builder.WriteString("function opsValidHex(value: string, length: number) {\n")
		builder.WriteString("  return new RegExp(\"^[a-f0-9]{\" + length + \"}$\").test(value) && !/^0+$/.test(value);\n")
		builder.WriteString("}\n")
		builder.WriteString("function opsTraceContext(header: string | string[] | undefined): OpsTraceContext {\n")
		builder.WriteString("  const traceparent = Array.isArray(header) ? header[0] ?? \"\" : header ?? \"\";\n")
		builder.WriteString("  const parts = traceparent.split(\"-\");\n")
		builder.WriteString("  const candidateTraceId = parts[1] ?? \"\";\n")
		builder.WriteString("  const candidateParentSpanId = parts[2] ?? \"\";\n")
		builder.WriteString("  const candidateTraceFlags = parts[3] ?? \"\";\n")
		builder.WriteString("  const inheritsTrace = parts[0] === \"00\" && opsValidHex(candidateTraceId, 32) && opsValidHex(candidateParentSpanId, 16);\n")
		builder.WriteString("  const traceId = inheritsTrace ? candidateTraceId : opsRandomHex(16);\n")
		builder.WriteString("  const parentSpanId = inheritsTrace ? candidateParentSpanId : undefined;\n")
		builder.WriteString("  const traceFlags = /^[a-f0-9]{2}$/.test(candidateTraceFlags) ? candidateTraceFlags : \"01\";\n")
		builder.WriteString("  const spanId = opsRandomHex(8);\n")
		builder.WriteString("  return { traceId, spanId, parentSpanId, traceFlags, traceparent: `00-${traceId}-${spanId}-${traceFlags}` };\n")
		builder.WriteString("}\n")
		if g.hasOpsOTLPObserve() {
			builder.WriteString("function opsUnixNano(ms: number) {\n")
			builder.WriteString("  return String(BigInt(ms) * 1000000n);\n")
			builder.WriteString("}\n")
			builder.WriteString("function opsOTLPValue(value: unknown) {\n")
			builder.WriteString("  if (typeof value === \"number\" && Number.isFinite(value)) return { intValue: String(Math.trunc(value)) };\n")
			builder.WriteString("  if (typeof value === \"boolean\") return { boolValue: value };\n")
			builder.WriteString("  return { stringValue: String(value ?? \"\") };\n")
			builder.WriteString("}\n")
			builder.WriteString("function opsOTLPAttributes(event: Record<string, unknown>) {\n")
			builder.WriteString("  return Object.entries(event).filter(([, value]) => value !== undefined).map(([key, value]) => ({ key, value: opsOTLPValue(value) }));\n")
			builder.WriteString("}\n")
			builder.WriteString("function opsOTLPTracePayload(event: Record<string, unknown>, traceContext: OpsTraceContext, started: number) {\n")
			builder.WriteString("  const statusCode = Number(event.status ?? 0);\n")
			builder.WriteString("  const span: Record<string, unknown> = {\n")
			builder.WriteString("    traceId: traceContext.traceId,\n")
			builder.WriteString("    spanId: traceContext.spanId,\n")
			builder.WriteString("    name: String(event.method ?? \"HTTP\") + \" \" + String(event.path ?? \"/\"),\n")
			builder.WriteString("    kind: 2,\n")
			builder.WriteString("    startTimeUnixNano: opsUnixNano(started),\n")
			builder.WriteString("    endTimeUnixNano: opsUnixNano(Date.now()),\n")
			builder.WriteString("    attributes: opsOTLPAttributes(event),\n")
			builder.WriteString("    status: { code: statusCode >= 500 ? 2 : 1 }\n")
			builder.WriteString("  };\n")
			builder.WriteString("  if (traceContext.parentSpanId) span.parentSpanId = traceContext.parentSpanId;\n")
			builder.WriteString("  return {\n")
			builder.WriteString("    resourceSpans: [{\n")
			builder.WriteString("      resource: { attributes: [\n")
			builder.WriteString(fmt.Sprintf("        { key: \"service.name\", value: { stringValue: %s } },\n", contractJSONString(g.program.App.Name)))
			builder.WriteString(fmt.Sprintf("        { key: \"service.version\", value: { stringValue: %s } },\n", contractJSONString(version)))
			builder.WriteString("        { key: \"telemetry.sdk.name\", value: { stringValue: \"blacklang\" } }\n")
			builder.WriteString("      ] },\n")
			builder.WriteString(fmt.Sprintf("      scopeSpans: [{ scope: { name: \"blacklang.generated.ops\", version: %s }, spans: [span] }]\n", contractJSONString(version)))
			builder.WriteString("    }]\n")
			builder.WriteString("  };\n")
			builder.WriteString("}\n")
		}
		builder.WriteString("function opsObservePayload(event: Record<string, unknown>, traceContext: OpsTraceContext, started: number) {\n")
		if g.hasOpsOTLPObserve() {
			builder.WriteString("  return opsOTLPTracePayload(event, traceContext, started);\n")
		} else {
			builder.WriteString("  return event;\n")
		}
		builder.WriteString("}\n")
	}
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) opsServerMiddleware() string {
	if !g.hasOpsMetrics() && !g.hasOpsRequestLogging() && !g.hasOpsObserve() {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("app.use((req, res, next) => {\n")
	builder.WriteString("  const started = Date.now();\n")
	if g.hasOpsObserve() {
		builder.WriteString("  const traceContext = opsTraceContext(req.headers.traceparent);\n")
		builder.WriteString("  res.setHeader(\"traceparent\", traceContext.traceparent);\n")
	}
	builder.WriteString("  res.on(\"finish\", () => {\n")
	if g.hasOpsMetrics() {
		builder.WriteString("    opsCounters.requests += 1;\n")
		builder.WriteString("    if (res.statusCode >= 500) opsCounters.errors += 1;\n")
		builder.WriteString("    opsStatusCounts.set(res.statusCode, (opsStatusCounts.get(res.statusCode) ?? 0) + 1);\n")
	}
	if g.hasOpsRequestLogging() {
		builder.WriteString("    console.log(JSON.stringify({\n")
		builder.WriteString("      at: new Date().toISOString(),\n")
		builder.WriteString("      method: req.method,\n")
		builder.WriteString("      path: req.originalUrl,\n")
		builder.WriteString("      status: res.statusCode,\n")
		builder.WriteString("      durationMs: Date.now() - started\n")
		builder.WriteString("    }));\n")
	}
	if g.hasOpsObserve() {
		builder.WriteString("    const observeEndpoint = opsObserveEndpoint();\n")
		builder.WriteString("    if (observeEndpoint) {\n")
		builder.WriteString("      const event = {\n")
		builder.WriteString("        kind: \"request\",\n")
		builder.WriteString("        provider: opsObserveProvider,\n")
		builder.WriteString(fmt.Sprintf("        app: %s,\n", contractJSONString(g.program.App.Name)))
		builder.WriteString(fmt.Sprintf("        version: %s,\n", contractJSONString(version)))
		builder.WriteString("        at: new Date().toISOString(),\n")
		builder.WriteString("        method: req.method,\n")
		builder.WriteString("        path: req.originalUrl,\n")
		builder.WriteString("        status: res.statusCode,\n")
		builder.WriteString("        durationMs: Date.now() - started,\n")
		builder.WriteString("        traceId: traceContext.traceId,\n")
		builder.WriteString("        spanId: traceContext.spanId,\n")
		builder.WriteString("        parentSpanId: traceContext.parentSpanId,\n")
		builder.WriteString("        traceFlags: traceContext.traceFlags\n")
		builder.WriteString("      };\n")
		builder.WriteString("      const observePayload = opsObservePayload(event, traceContext, started);\n")
		builder.WriteString("      void fetch(observeEndpoint, {\n")
		builder.WriteString("        method: \"POST\",\n")
		builder.WriteString("        headers: { \"content-type\": \"application/json\" },\n")
		builder.WriteString("        body: JSON.stringify(observePayload)\n")
		builder.WriteString("      }).catch((error) => {\n")
		builder.WriteString("        console.error(JSON.stringify({ at: new Date().toISOString(), kind: \"observability_delivery_error\", provider: opsObserveProvider, message: error instanceof Error ? error.message : String(error) }));\n")
		builder.WriteString("      });\n")
		builder.WriteString("    }\n")
	}
	builder.WriteString("  });\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	return builder.String()
}

func (g *webGenerator) opsServerRoutes() string {
	if !g.hasOps() {
		return ""
	}
	var builder strings.Builder
	if path := g.opsHealthPath(); path != "" {
		builder.WriteString(fmt.Sprintf("app.get(%s, (_req, res) => {\n", contractJSONString(path)))
		builder.WriteString(fmt.Sprintf("  res.json({ status: \"ok\", app: %s, version: %s, uptimeSeconds: Math.floor(process.uptime()), startedAt: startedAt.toISOString() });\n", contractJSONString(g.program.App.Name), contractJSONString(version)))
		builder.WriteString("});\n")
	}
	if path := g.opsReadinessPath(); path != "" {
		builder.WriteString(fmt.Sprintf("app.get(%s, async (_req, res) => {\n", contractJSONString(path)))
		builder.WriteString("  try {\n")
		builder.WriteString("    await prisma.$queryRawUnsafe(\"SELECT 1\");\n")
		builder.WriteString(fmt.Sprintf("    res.json({ status: \"ready\", app: %s, database: \"ok\", uptimeSeconds: Math.floor(process.uptime()) });\n", contractJSONString(g.program.App.Name)))
		builder.WriteString("  } catch (_error) {\n")
		builder.WriteString(fmt.Sprintf("    res.status(503).json({ status: \"not_ready\", app: %s, database: \"error\" });\n", contractJSONString(g.program.App.Name)))
		builder.WriteString("  }\n")
		builder.WriteString("});\n")
	}
	if path := g.opsMetricsPath(); path != "" {
		builder.WriteString(fmt.Sprintf("app.get(%s, (_req, res) => {\n", contractJSONString(path)))
		builder.WriteString("  res.json({\n")
		builder.WriteString(fmt.Sprintf("    app: %s,\n", contractJSONString(g.program.App.Name)))
		builder.WriteString(fmt.Sprintf("    version: %s,\n", contractJSONString(version)))
		builder.WriteString("    uptimeSeconds: Math.floor(process.uptime()),\n")
		builder.WriteString("    startedAt: startedAt.toISOString(),\n")
		builder.WriteString("    requests: opsCounters.requests,\n")
		builder.WriteString("    errors: opsCounters.errors,\n")
		builder.WriteString("    statusCodes: Object.fromEntries(Array.from(opsStatusCounts.entries()).map(([status, count]) => [String(status), count]))\n")
		builder.WriteString("  });\n")
		builder.WriteString("});\n")
	}
	if builder.Len() > 0 {
		builder.WriteString("\n")
	}
	return builder.String()
}

func (g *webGenerator) addOpenAPIOpsPaths(paths map[string]any) {
	if !g.hasOps() {
		return
	}
	if path := g.opsHealthPath(); path != "" {
		paths[path] = map[string]any{
			"get": openapiOpsOperation("health", "Runtime health probe", []string{"status", "app", "version", "uptimeSeconds", "startedAt"}),
		}
	}
	if path := g.opsReadinessPath(); path != "" {
		paths[path] = map[string]any{
			"get": openapiOpsOperation("readiness", "Runtime readiness probe", []string{"status", "app", "database", "uptimeSeconds"}),
		}
	}
	if path := g.opsMetricsPath(); path != "" {
		paths[path] = map[string]any{
			"get": openapiOpsOperation("metrics", "Runtime metrics snapshot", []string{"app", "version", "uptimeSeconds", "startedAt", "requests", "errors", "statusCodes"}),
		}
	}
}

func (g *webGenerator) openAPIObservabilityExtension() map[string]any {
	if !g.hasOpsObserve() {
		return nil
	}
	return map[string]any{
		"provider":     g.opsObserveProvider(),
		"endpointEnv":  g.opsObserveEndpointEnv(),
		"protocol":     g.opsObserveProtocol(),
		"mode":         g.opsObserveMode(),
		"signal":       g.opsObserveSignal(),
		"traceContext": "w3c-traceparent",
		"delivery":     "non-blocking",
	}
}

func openapiOpsOperation(kind string, summary string, required []string) map[string]any {
	return map[string]any{
		"summary":            summary,
		"x-blacklang-ops":    kind,
		"x-blacklang-public": true,
		"responses": openapiJSONResponses(openapiObjectSchema(map[string]any{
			"status":        map[string]any{"type": "string"},
			"app":           map[string]any{"type": "string"},
			"version":       map[string]any{"type": "string"},
			"database":      map[string]any{"type": "string"},
			"uptimeSeconds": map[string]any{"type": "integer"},
			"startedAt":     map[string]any{"type": "string", "format": "date-time"},
			"requests":      map[string]any{"type": "integer"},
			"errors":        map[string]any{"type": "integer"},
			"statusCodes":   map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "integer"}},
		}, required)),
	}
}
