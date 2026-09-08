# BlackLang Web Geliştirme Yol Haritası

## Amaç

BlackLang, AI-native deterministic intent language olarak tasarlanır.

Yani temel hedefi; yapay zeka ajanlarının uygulama niyetini kısa, net, doğrulanabilir ve deterministik bir kaynak dille yazmasını, sonra bu niyetten çalışan yazılım üretmesini sağlamaktır.

BlackLang'in ilk büyük hedefi, yapay zeka ajanlarının web tabanlı uygulamaları daha az token, daha az dosya değişikliği ve daha düşük hata oranıyla geliştirebilmesini sağlamaktır.

Bu belge, web geliştirme için gereken tüm temel ihtiyaçların BlackLang'e hangi sırayla ekleneceğini açıklar.

BlackLang'in web tarafındaki amacı sadece React, Node.js veya SQL üretmek değildir. Asıl amaç, AI ajanının uygulamayı daha kolay okuyacağı, anlayacağı, değiştireceği ve doğrulayacağı bir kaynak temsil oluşturmaktır.

## Ana Tasarım İlkesi

BlackLang şu soruya göre tasarlanmalıdır:

> Eğer ben bir yapay zeka ajanı olsaydım, bir yazılım projesini nasıl daha hızlı, daha az enerjiyle ve daha az belirsizlikle anlardım?

Bu yüzden dilin temel ilkeleri şunlardır:

- Tek işi tek blok anlatır.
- Aynı davranış için birden fazla yazım şekli olmaz.
- Kısa ama şifreli olmayan syntax kullanılır.
- AI'nin tahmin etmesi gereken alanlar azaltılır.
- Dosya yapısı sabit ve tahmin edilebilir olur.
- Generated kod ile BlackLang kaynak kodu kesin ayrılır.
- Hata mesajları hem insan hem AI için anlaşılır olur.
- CLI komutları JSON çıktı verebilir.
- Dil, önce iş niyetini anlatır; teknik detayları generator çözer.

## Genel Mimari

```text
Human Request
      ↓
AI Coding Agent
      ↓
BlackLang Source
      ↓
Parser
      ↓
AST
      ↓
Validator
      ↓
IR
      ↓
Code Generator
      ↓
Web Application
```

İlk hedef web uygulamasıdır:

```text
BlackLang Core
      ↓
Web Target
      ↓
React / API / Database / Auth / Tests / Deploy
```

İleride aynı core temsil şu hedeflere genişleyebilir:

```text
BlackLang Core
├── Web Target
├── Mobile Target
├── Desktop Target
├── API-only Target
└── Automation Target
```

## Katmanlı Uzun Vadeli Dil Hedefi

Web target için yakın hedef JavaScript runtime parity değildir. Yakın hedef, deterministic web capability parity olmalıdır: CRUD/admin/SaaS/API/jobs/deploy/test gibi yaygın web işleri BlackLang kaynak niyetiyle kısa, doğrulanabilir ve generated-output-safe şekilde yapılabilmelidir.

Bu ara hedef uzun vadeli genel amaçlı dil hedefini iptal etmez. Web completion içinde sınırlı bir Core Program Logic alt kümesi başlamıştır: computed/action/API arithmetic, action/API local `value`, indentation-based `if`/`else` ve compound `and`/`or`/`not` condition'lar. Web tarafı olgunlaştıktan sonra BlackLang ayrı bir Core Program Logic v1 katmanına ihtiyaç duyar. Bu katman, hazır declarative web blokları yetmediğinde kullanıcının kontrollü ve deterministik program mantığını row/API handler sınırının dışına taşımasını sağlar.

Post-web Core Program Logic v1 için ayrı takip edilecek başlıklar:

- literals
- general-purpose variables
- assignment
- full expression grammar
- operator precedence
- boolean logic
- conditional logic outside bounded action/API handlers
- functions
- parameters
- return
- lexical scope
- collections
- loops
- error values / error handling
- bounded async operations

Bu çekirdek web completion yüzdesine karıştırılmamalıdır. Web %100 hedefi production web uygulaması üretme kapasitesini ölçer; Core Program Logic v1 ise BlackLang'in daha sonra Python/JavaScript benzeri genel amaçlı alana büyümesini sağlar.

Core eval milestone'ları ayrı izlenmelidir:

```text
CORE-EVAL-001  Calculator with local expression state
CORE-EVAL-002  Todo with local state
CORE-EVAL-003  Conditional pricing logic
CORE-EVAL-004  Loop over collection
CORE-EVAL-005  Bounded async API call
```

Bu eval'ların kabul kriteri şudur: AI ajanı yalnızca resmi BlackLang kaynak syntax'ı ve compiler davranışı üzerinden çalışmalı; generated JavaScript/TypeScript dosyalarını çözüm olarak elle düzenlememelidir.

## Web %100 Kapsam Yol Haritası

Bu bölümün amacı, BlackLang'in web tarafında uzun vadede "webde yapamayacağı şey kalmaması" hedefini ölçülebilir hale getirmektir.

Buradaki `%100`, dünyadaki her framework'ün her özel kullanımını birebir çekirdeğe gömmek anlamına gelmez. Daha doğru tanım şudur:

> Modern bir web projesinde ihtiyaç duyulan her ana davranış, BlackLang kaynak diliyle veya resmi BlackLang extension/plugin sistemiyle deterministik şekilde tarif edilebilmeli, doğrulanabilmeli, üretilebilmeli ve deploy edilebilmelidir.

Yani hedef yalnızca CRUD üretmek değildir. Hedef; frontend, backend, database, güvenlik, dosya, ödeme, realtime, test, deployment, observability ve ekip çalışması dahil production web geliştirme alanının tamamını kapsayan bir kaynak temsil oluşturmaktır.

### Kapsam Prensibi

BlackLang webde %100'e giderken iki katmanlı ilerlemelidir:

```text
BlackLang Core
  ortak, deterministik, güvenli web kavramları

BlackLang Extensions
  payment provider, cloud storage, email provider, AI service, özel framework adapterleri
```

Core içine her şeyi doldurmak dili ağırlaştırır. Ama her şeyi plugin'e bırakmak da dilin gücünü azaltır. Bu yüzden karar kuralı:

- Her web uygulamasında sık görülen davranışlar core'a girer.
- Provider'a veya şirkete özel davranışlar extension/plugin olarak bağlanır.
- Her extension aynı parser, validator, docs, JSON/BlackIR ve affected graph disiplinine uyar.
- AI ajanı bir extension'ı kullanmadan önce o extension'ın kısa öğrenme paketini okuyabilir.

### %100 İçin Ana Kapsam Alanları

| Alan | Hedef | Durum |
|---|---|---|
| App structure | Proje, target, generated sınırı, config, paketleme | Kısmen var |
| Data model | Entity, relation, index, constraint, migration, seed | Kısmen var |
| CRUD | Create, read, update, delete, archive, restore, bulk actions | Kısmen var |
| Query | Filter, sort, pagination, custom query, aggregate, join, search | Kısmen var: stored-field custom query + aggregate summary MVP; join yok |
| UI composition | Section, panel, grid, stack, tabs, modal, drawer, navbar, footer | Güçlü MVP: order, grid/stack, tabs, modal/drawer, nested groups, accessibility audit, interaction triggers |
| Styling/theme | Token, theme profile, responsive rule, state style, animation | Kısmen var: theme profile, inline UI, migration check |
| Forms | Inputs, validation, wizard, dynamic field, file field, dependent select | Kısmen var |
| Routing | Nested route, dynamic route, protected route, layout route | Başlangıç gerekli |
| State | Page state, global state, URL state, persisted state | Kısmen var |
| Components | Input/output, variant, slot, event, composition | Kısmen var |
| Frontend logic | if, loop, computed display, event handler, client action | Kısmen var: computed display + generated custom action panel |
| Backend logic | service, command, transaction, domain rule, scheduled job | Güçlü MVP: deterministic row-level custom action, top-level transaction blocks, explicit API service grouping ve read-only background query worker |
| API | REST, OpenAPI, custom endpoint, webhook, GraphQL/gRPC adapter | Güçlü MVP: CRUD, query, custom action, explicit API declared runtime ve service metadata |
| Auth | Email/password, OAuth, 2FA, password reset, session/JWT | Kısmen var |
| Authorization | Role, permission, ownership, tenant, policy, row-level access | Güçlü MVP: role/permission + multi-role kullanıcı ataması + tenant admin UI + owner/tenant policy |
| Security | CORS, CSRF, headers, rate limit, secrets, audit, scanning, protected source | Güçlü MVP: CORS, headers, rate limit, audit, source scan, `.black.enc`, secret reference manifest, provider preflight execution, signed release trust, release transparency, key rotation policy |
| Database runtime | SQLite, PostgreSQL, MySQL, Redis/cache, migrations | SQLite, PostgreSQL ve MySQL runtime tamamlandı; MySQL rename migration runner bilinçli sınır |
| File/media | Upload, image processing, storage provider, download permission | Güçlü yerel MVP: `file`/`image` stored field, data URL form input, table/detail preview, API validation ve OpenAPI metadata; external storage, image processing ve download permission provider işi olarak eksik |
| Email/notification | SMTP, provider adapter, template, queue, retry | Eksik |
| Realtime | WebSocket, SSE, presence, live table updates, notifications | Eksik |
| Payments | Product, price, checkout, subscription, invoice, webhook | Eksik |
| Background jobs | Queue, cron, retry, delayed task, worker target | Başlangıç MVP: deterministic `job` syntax, query worker, manifest, package scripts; queue/retry/provider scheduler eksik |
| Search | Full-text, external search provider, indexing, facets | Eksik |
| Testing | Unit, API, UI, e2e, fixture, generated benchmark tests | Güçlü MVP: generated contract/API/frontend/browser-check smoke tests + real browser e2e execution + cross-browser e2e matrix runner + deterministic seed fixture runtime + benchmark task/token reports |
| Observability | Logging, metrics, tracing, health check, error reporting | Güçlü MVP: ops health/readiness/metrics + request logs + webhook hook |
| Deployment | Docker, VPS, preview, cloud adapter, env management, rollback | Mature MVP: Docker + ops healthcheck + local preview compose + rollback metadata + cloud adapter plan/preflight + explicit provider CLI apply runner |
| Performance | Bundle policy, cache, CDN, lazy loading, query optimization | Eksik |
| Accessibility | Semantic UI rules, keyboard nav, contrast, aria validation | Başlangıç gerekli |
| SEO/content | Meta, sitemap, robots, CMS/content pages, markdown routes | Başlangıç gerekli |
| Plugin ecosystem | Extension manifest, versioning, docs, install, trust model | Güçlü MVP: compiler-owned ecosystem discovery + npm/editor package metadata + local registry/marketplace manifests + trust validator + signed package trust workflow + multi-editor package channel metadata |
| IDE support | Syntax highlight, autocomplete, diagnostics, refactor UI | Güçlü MVP: compiler-owned IDE metadata + packageable VS Code/Open VSX/Cursor-compatible bridge metadata + diagnostics/completion/snippet/format/affected workflow |
| Migration/refactor | Rename, move, split, schema migration, UI profile migration | Başlangıç var: UI profile migration check |

### Yüzde Kilometre Taşları

Bu yüzdeler kesin matematik değil, web kapasitesini takip etmek için çalışma ölçüsüdür.

```text
%20  MVP web app
     entity, page, CRUD, basic auth, basic role, validation, generated app

%40  Admin/dashboard web
     relation, workflow, table tools, i18n label, inline UI, Docker, docs, affected graph

%60  Flexible product UI
     layout composition, routing, modal/tabs/drawer, responsive rules, richer components

%75  Business application platform
     custom query, custom action, backend service, transaction, local media fields, read-only query jobs, email

%90  Production web platform
     PostgreSQL/MySQL, migrations, tests, observability, cache, search, realtime, payments

%100 Full web coverage model
     plugin ecosystem, cloud/provider adapters, IDE support, versioned migrations,
     security hardening, performance/a11y/SEO policies, enterprise deployment workflows
```

Mevcut durum `black benchmark coverage --json` ile makine okunur şekilde raporlanır. Güncel weighted signal %100'dür. Bu oran pazarlama iddiası değil; data model/CRUD, frontend UI/layout, backend/API/actions, auth/security, database runtime, deployment/ops, testing/benchmarks, docs/AI ergonomics ve ecosystem alanlarının ağırlıklı planlama ölçüsüdür. Dış registry publish ve canlı hosted index deployment hâlâ release-owner/review sonrası adımlardır.

Coverage report ayrıca roadmap gap'lerini stable ID'lere çevirir:

```text
WEB-UI-001       UI profile migration rules (done)
WEB-LAYOUT-001   modal/drawer section display (done)
WEB-LAYOUT-002   nested section groups with local stack/grid composition (done)
WEB-LAYOUT-003   accessibility audit policy checks for overlays and nested groups (done)
WEB-LAYOUT-004   advanced interactive composition triggers (done)
WEB-LAYOUT-005   reusable page component sections bound to selected/first records (done)
WEB-LAYOUT-006   generated JSX DOM order follows effective page view order (done)
WEB-LAYOUT-007   collection-backed page component sections with bind each (done)
WEB-I18N-001     runtime language switching for field labels (done)
WEB-I18N-002     placeholder/help/message/date/number/currency/RTL support (done)
WEB-DATA-001     generated schema migration plan before destructive database changes (done)
WEB-DATA-002     PostgreSQL runtime support (done)
WEB-DATA-003     applied migration runtime and first-class rename/refactor syntax (done)
WEB-DATA-006     generated online migration runner safeguards (done)
WEB-DATA-009     MySQL target runtime, seed, and Docker Compose provider output (done)
WEB-DATA-007     relation response load policies for generated routes (done)
WEB-DATA-008     single-hop bulk relation prefetch and nested relation response sanitization (done)
WEB-SEC-001      ownership, tenant, and policy language (done)
WEB-SEC-002      multiple roles per authenticated user (done)
WEB-SEC-003      tenant administration UI for tenant policies (done)
WEB-SEC-004      generated secret reference manifest and provider handoff plan (done)
WEB-SEC-007      read-only secret manager provider preflight execution (done)
WEB-SEC-005      signed compiler and release verification trust workflow (done)
WEB-SEC-006      release transparency log and key rotation policy metadata (done)
WEB-API-001      explicit API and webhook declared-runtime routes (done)
WEB-API-002      explicit API handler syntax for business side effects (done)
WEB-API-003      aggregate summaries for page-bound custom queries (done)
WEB-API-004      deterministic background query worker jobs (done)
WEB-API-005      API-only target generation (done)
WEB-API-006      explicit API service/module grouping (done)
WEB-OPS-001      local preview deployment target and rollback metadata (done)
WEB-OPS-002      health/readiness/metrics probes and request logging (done)
WEB-OPS-003      cloud adapters and external observability provider hooks (done)
WEB-OPS-004      provider CLI deployment preflight and explicit apply runner (done)
WEB-OPS-005      distributed trace context and OTLP exporter metadata (done)
WEB-TEST-001     AI task benchmarks and token estimate reports (done)
WEB-TEST-002     first-class fixture and seed syntax (done)
WEB-TEST-003     first-class browser-check syntax and generated browser checks (done)
WEB-TEST-004     full browser e2e execution with real browser automation (done)
WEB-TEST-005     generated cross-browser e2e matrix plan and runner (done)
WEB-TEST-006     long-running AI eval corpus metadata (done)
WEB-TEST-007     evidence-backed AI eval result history export (done)
WEB-DOCS-001     coverage status on docs site (done)
WEB-DOCS-002     tracked web coverage issue export (done)
WEB-DOCS-003     IDE diagnostics and autocomplete support (done)
WEB-DOCS-004     IDE refactor workflows and packaged editor extension (done)
WEB-DOCS-005     multi-editor marketplace package channel metadata (done)
WEB-ECO-001      installable artifacts and plugin/adapter discovery (done)
WEB-ECO-002      prepared package registries and provider adapter marketplace (done)
WEB-ECO-003      signed package and provider adapter verification workflow (done)
WEB-ECO-004      prepared public ecosystem index source for hosted package and adapter discovery (done)
```

### Öncelikli Büyük Aşamalar

%100'e giderken en doğru sıra şu olmalıdır:

1. UI composition ve layout dilini kur.
2. Routing ve sayfa hiyerarşisini genişlet.
3. Custom query ve custom action sistemini ekle.
4. Backend service/logic syntax'ını ekle.
5. Database migration ve PostgreSQL runtime desteğini tamamla.
6. File upload ve storage adapterlerini ekle.
7. Email, notification ve provider-backed job queue sistemini ekle.
8. Realtime, cache ve full-text search katmanını ekle.
9. Test generator ve benchmark komutlarını ekle.
10. Payment ve webhook provider adapterlerini ekle.
11. Observability, health check ve production logging ekle.
12. Plugin/extension sistemini standartlaştır.
13. IDE autocomplete, diagnostics ve refactor araçlarını ekle.
14. Accessibility, SEO ve performance policy kontrollerini dil seviyesine çıkar.
15. Çoklu deployment hedeflerini ve rollback akışlarını tamamla.

### Dilde Kalması Gereken Denge

BlackLang webde %100'e giderken şu riski sürekli kontrol etmelidir:

```text
çok özellik
  +
çok syntax
  =
AI için tekrar zorlaşan dil
```

Bu yüzden her yeni web yeteneği eklenmeden önce şu sorular sorulmalıdır:

- Bu özellik web projelerinde sık mı kullanılıyor?
- AI bunu normal stack'te yazarken çok dosya ve çok token harcıyor mu?
- BlackLang bunu daha kısa ve deterministik anlatabiliyor mu?
- Validator hatayı erken yakalayabiliyor mu?
- Generated output test edilebilir mi?
- Bu core özellik mi, yoksa extension mı olmalı?
- AI ajanı bu özelliği kısa bir docs/explain çıktısından öğrenebilir mi?

### Nihai Hedef

Nihai web hedefi şudur:

> Bir kullanıcı veya AI ajanı, modern bir web uygulamasının veri modelini, arayüzünü, iş mantığını, güvenliğini, entegrasyonlarını, testlerini ve deployment akışını BlackLang ile tarif edebilmeli; generator da bunu seçilen web target için çalışan, test edilebilir ve production'a hazırlanabilir çıktıya dönüştürebilmelidir.

## Aşama 0: Proje Zemini

Bu aşamada henüz tam bir dil yoktur. Ama proje standardı oluşturulur.

### Eklenecekler

- GitHub repository
- `README.md`
- `ROADMAP.md`
- `SPEC.md`
- `BLACKLANG.md`
- `AGENTS.md`
- `LICENSE`
- `CONTRIBUTING.md`
- Örnek `.black` dosyaları
- Örnek proje klasörü
- Dil hedefleri
- Yasaklı belirsizlikler
- AI ajanları için çalışma kuralları

### Amaç

Codex gibi bir ajan proje klasörünü açtığında önce ne okuyacağını, hangi dosyaları değiştireceğini ve hangi dosyalara dokunmayacağını bilmelidir.

### Örnek Klasör

```text
blacklang/
├── README.md
├── ROADMAP.md
├── SPEC.md
├── BLACKLANG.md
├── AGENTS.md
├── LICENSE
├── CONTRIBUTING.md
├── packages/
│   └── cli/
├── examples/
│   ├── warehouse/
│   ├── crm/
│   └── inventory/
├── docs/
├── benchmarks/
└── generated/
```

## Aşama 0.1: Dağıtım ve Ekosistem Altyapısı

BlackLang sadece yerel bir compiler denemesi olarak kalmamalıdır. En baştan indirilebilir, denenebilir, incelenebilir ve AI ajanları tarafından referans alınabilir bir ekosistem olarak kurulmalıdır.

### Ana Yayın Kanalları

- GitHub repository
- npm paketi
- GitHub Releases
- Dokümantasyon sitesi
- Örnek proje galerisi
- Benchmark raporları
- AI agent kullanım rehberleri

### GitHub Repository

GitHub, BlackLang'in ana merkezi olmalıdır.

Kullanım amaçları:

- Kaynak kodu yayınlamak
- Issue ve feature request toplamak
- Roadmap göstermek
- Release geçmişini tutmak
- Katkı kabul etmek
- npm paketiyle güven ilişkisi kurmak
- Codex ve benzeri ajanların dokümantasyon okuyabileceği sabit kaynak oluşturmak

İlk repo yapısı:

```text
blacklang/
├── README.md
├── ROADMAP.md
├── SPEC.md
├── BLACKLANG.md
├── AGENTS.md
├── LICENSE
├── CONTRIBUTING.md
├── packages/
│   └── cli/
├── examples/
│   ├── warehouse/
│   ├── crm/
│   └── inventory/
├── docs/
├── benchmarks/
└── .github/
    └── workflows/
        ├── test.yml
        └── release.yml
```

### CLI Dağıtımı

Nihai hedef, BlackLang'i tek binary olarak dağıtmaktır:

```text
Windows: black.exe
Linux:   black
macOS:   black
```

Kullanıcı deneyimi:

```bash
black init
black validate
black build
black dev
black test
```

Bu modelde BlackLang compiler'ı çalıştırmak için Python veya Node.js kurulu olması gerekmez.

### npm Üzerinden Kurulum

Yaygınlaşmak için npm desteği de olmalıdır.

```bash
npm install -g blacklang
```

veya:

```bash
npm install -D blacklang
npx black validate
npx black build
```

npm paketi kendi içinde platforma uygun binary'yi çalıştırabilir.

Örnek yapı:

```text
blacklang npm package
├── bin/
│   └── black.js
└── binaries/
    ├── windows-x64/
    │   └── black.exe
    ├── linux-x64/
    │   └── black
    ├── darwin-x64/
    │   └── black
    └── darwin-arm64/
        └── black
```

Daha profesyonel modelde platform paketleri ayrılabilir:

```text
blacklang
@blacklang/windows-x64
@blacklang/linux-x64
@blacklang/darwin-x64
@blacklang/darwin-arm64
```

### GitHub Releases

Her sürümde binary dosyaları GitHub Releases üzerinden yayınlanmalıdır.

Release çıktıları:

```text
black-windows-x64.zip
black-linux-x64.tar.gz
black-darwin-x64.tar.gz
black-darwin-arm64.tar.gz
checksums.txt
```

### CI/CD

GitHub Actions ile otomatik test ve release sistemi kurulmalıdır.

Gerekli işler:

- Parser testleri
- Validator testleri
- Generator snapshot testleri
- CLI komut testleri
- Windows/Linux/macOS build
- Release artifact üretimi
- npm publish

### Dokümantasyon Sitesi

İleride `blacklang.dev` gibi bir alan adıyla dokümantasyon sitesi oluşturulmalıdır.

İçerikler:

- Quick start
- Installation
- Language spec
- CLI reference
- Web generator guide
- AI agent guide
- Examples
- Error codes
- Benchmark reports

İlk aşamada GitHub Pages, Cloudflare Pages, Netlify veya Vercel yeterlidir.

### Package Manager Yayılımı

npm ve GitHub Releases oturduktan sonra daha fazla kanal eklenebilir.

v0.2 içinde dış registry yayını yapılmadan önce yerel hazırlık katmanı eklendi: `packages/registry/package-index.blackdir`, `packages/registry/trust-policy.blackdir`, `adapters/marketplace/adapter-index.blackdir`, `adapters/marketplace/trust-policy.blackdir` ve `node packages/registry/scripts/validate-registry.mjs`. Böylece AI ajanı package/adapter publish işine başlamadan önce manifestleri ve trust kurallarını deterministik şekilde okuyabilir.

Hedefler:

- Homebrew
- Scoop
- Winget
- Chocolatey
- Docker image

Örnek:

```bash
brew install blacklang
winget install blacklang
scoop install blacklang
docker run blacklang/black validate
```

### AI Agent Kullanım Rehberleri

BlackLang'in ana hedef kitlesi AI coding agentları olduğu için, her popüler agent için ayrı kullanım rehberi hazırlanmalıdır.

İlk hedefler:

- Codex
- Claude Code
- Cursor
- Windsurf
- GitHub Copilot coding agent

Örnek kural:

```md
Before editing a BlackLang project:

1. Read `AGENTS.md`.
2. Run `black inspect --json`.
3. Modify only `.black` source files.
4. Run `black validate --json`.
5. Run `black build`.
6. Do not manually edit generated files.
```

### Benchmark Yayını

BlackLang'in iddiası ölçülerek yayınlanmalıdır.

Benchmark alanları:

- Satır sayısı
- Dosya sayısı
- Input token
- Output token
- Değişiklik süresi
- Hata sayısı
- Test başarısı
- Agent'ın dokunduğu dosya sayısı

Bu raporlar GitHub ve dokümantasyon sitesinde yayınlanmalıdır.

### Kendi Server Ne Zaman Gerekir?

İlk aşamada BlackLang dağıtımı için özel server gerekmez.

Yeterli altyapı:

- GitHub
- GitHub Actions
- GitHub Releases
- npm registry
- Statik dokümantasyon hosting

Kendi backend server ancak şu özellikler gelirse gerekir:

- Cloud build
- Kullanıcı hesabı
- Lisans yönetimi
- Telemetry dashboard
- Online template marketplace
- AI API proxy
- Hosted project registry

İlk sürümde bu alanlar kapsam dışı tutulmalıdır.

## Aşama 0.2: BlackLang Öğrenme ve Referans Sitesi

BlackLang geliştirilirken aynı anda resmi bir öğrenme ve referans sitesi de hazırlanmalıdır.

Bu site sadece insanlara tanıtım yapan bir web sitesi olmamalıdır. Python dokümantasyonu, MDN veya W3Schools gibi hem insanların hem de AI ajanlarının başvurabileceği güncel bir bilgi kaynağı olmalıdır.

### Ana Amaç

Site şu iki kullanıcıyı aynı anda hedeflemelidir:

- BlackLang öğrenmek isteyen insan geliştirici
- BlackLang projesinde çalışan AI coding agent

Bu yüzden her konu hem açıklamalı hem de makine tarafından kolay taranabilir şekilde yazılmalıdır.

### Site Mantığı

```text
blacklang.dev
├── Learn
├── Reference
├── Examples
├── CLI
├── Errors
├── AI Agents
├── Roadmap
└── Benchmarks
```

### İlk Sayfalar

- BlackLang nedir?
- Kurulum
- İlk `.black` dosyan
- `app` kullanımı
- `entity` kullanımı
- Field tipleri
- Field modifier kuralları
- `page` kullanımı
- `table` kullanımı
- `form` kullanımı
- `actions` kullanımı
- CLI komutları
- JSON hata çıktıları
- Codex ile kullanım

### Reference Bölümü

Her keyword için ayrı referans sayfası olmalıdır.

Örnek:

```text
/reference/app
/reference/entity
/reference/page
/reference/table
/reference/form
/reference/actions
/reference/search
```

Her referans sayfası aynı şablonu kullanmalıdır:

```text
Keyword
Ne işe yarar?
Syntax
Parametreler
Geçerli kullanım
Hatalı kullanım
Üretilen web karşılığı
AI agent notları
İlgili hata kodları
```

### AI Agent Bölümü

AI ajanları için özel bir bölüm olmalıdır.

İçerikler:

- Codex ile BlackLang kullanımı
- Claude Code ile BlackLang kullanımı
- Cursor ile BlackLang kullanımı
- Generated dosyalara dokunmama kuralı
- `black inspect --json` kullanımı
- `black validate --json` kullanımı
- Hata kodlarını yorumlama
- Bir değişikliğin etkisini öğrenme

Örnek:

```md
When editing a BlackLang project:

1. Read the local `AGENTS.md`.
2. Read the relevant docs page.
3. Modify only `.black` files.
4. Run `black validate --json`.
5. Run `black build`.
6. Do not manually edit generated files.
```

### AI İçin Güncel Bilgi Kaynağı

AI modelleri BlackLang'i eğitim verilerinden bilmeyebilir. Bu yüzden site, AI'nin webden araştırarak güncel kuralları öğrenebileceği resmi kaynak olmalıdır.

Bu amaçla:

- Sayfalar açık başlıklarla yazılmalıdır.
- Her keyword için örnekler küçük tutulmalıdır.
- Hata kodları stabil olmalıdır.
- Eski syntax değişiklikleri migration notlarıyla açıklanmalıdır.
- Her sürümün değişiklik notu yayınlanmalıdır.
- `llms.txt` ve `llms-full.txt` gibi AI dostu indeks dosyaları eklenmelidir.

### llms.txt Fikri

Site kökünde AI ajanları için kısa bir indeks dosyası bulunabilir:

```text
https://blacklang.dev/llms.txt
```

İçerik örneği:

```text
# BlackLang

BlackLang is an AI-native application language.

Read these pages first:
- /reference/syntax
- /reference/entity
- /reference/page
- /cli/validate
- /ai-agents/codex
```

Daha geniş tek dosya referansı:

```text
https://blacklang.dev/llms-full.txt
```

Bu dosya, AI ajanlarının tek seferde güncel dil özetini alabilmesi için hazırlanabilir.

### AI Öğrenme Maliyeti Riski

BlackLang yeni bir dil olduğu için ilk aşamada AI ajanları dili eğitim verilerinden bilmeyecektir.

Bu gerçek bir risktir:

```text
İlk görevde:
Normal web stack → düşük öğrenme maliyeti
BlackLang        → dokümantasyon okuma maliyeti
```

Bu yüzden BlackLang'in başarısı şu dengeye bağlıdır:

```text
İlk öğrenme maliyeti
  <
Tekrar eden görevlerde kazanılan token ve hata avantajı
```

Örnek hipotez:

```text
TypeScript:
Başlangıç öğrenmesi:        0 token
100 görev × 10.000 token = 1.000.000 token

BlackLang:
Dil öğrenme paketi:        20.000 token
100 görev × 2.000 token = 200.000 token
Toplam:                   220.000 token
```

Bu rakamlar kanıt değil, test edilmesi gereken hipotezdir.

### Bu Riski Azaltma Stratejisi

BlackLang şu yöntemlerle ilk öğrenme maliyetini düşürmelidir:

- `BLACKLANG.md` kısa tutulur.
- `AGENTS.md` net bir çalışma protokolü verir.
- `SPEC.md` örnek odaklı yazılır.
- Her keyword için küçük referans sayfası olur.
- `llms.txt` AI'ye önce ne okuyacağını söyler.
- `black docs <keyword> --json` sadece gerekli bilgiyi döner.
- `black explain <concept> --for-agent` kısa açıklama verir.
- `black validate --json` hatayı kod, dosya, satır ve öneriyle anlatır.
- Syntax tamamen yabancı sembollerden oluşmaz.
- Mevcut programlama dillerinden tanıdık kelimeler kullanılır.

### Sürüm Bilinci

Her BlackLang projesinde sürüm bilgisi olmalıdır:

```toml
version = "0.1"
target = "web"
```

AI ajanı projeye girdiğinde önce bu sürümü öğrenmeli, sonra o sürüme ait syntax kurallarını uygulamalıdır.

Planlanan komutlar:

```bash
black version
black docs --version 0.1 --agent
black docs entity --json
black explain table --for-agent
```

Bu sayede modelin eğitim verisi eski kalsa bile, agent güncel BlackLang davranışını yerel CLI veya resmi docs sitesi üzerinden öğrenebilir.

### Dokümantasyon Güncelleme Kuralı

Dil değiştikçe site de güncellenmelidir.

Kural:

> BlackLang syntax'ında veya CLI davranışında yapılan her değişiklik, aynı PR içinde dokümantasyon sitesine de yansıtılmalıdır.

### İnsan İçin Öğrenme Akışı

Siteye giren insan geliştirici şu sırayla ilerleyebilmelidir:

1. BlackLang nedir?
2. Neden AI-native?
3. Nasıl kurulur?
4. İlk uygulama nasıl yazılır?
5. `.black` dosyası nasıl parse edilir?
6. Nasıl validate edilir?
7. Nasıl build edilir?
8. Üretilen web uygulaması nasıl çalıştırılır?

### AI İçin Öğrenme Akışı

AI ajanı şu sırayla bilgi alabilmelidir:

1. `llms.txt` oku.
2. İlgili keyword referansını oku.
3. Projedeki `AGENTS.md` dosyasını oku.
4. `black inspect --json` çalıştır.
5. `.black` dosyasını değiştir.
6. `black validate --json` çalıştır.
7. Hata koduna göre ilgili docs sayfasını oku.
8. `black build` çalıştır.

### İlk Site Teknolojisi

İlk aşamada site statik dokümantasyon sitesi olabilir.

Uygun seçenekler:

- VitePress
- Astro
- Docusaurus
- Next.js
- Sites hosting

Başlangıç için önemli olan teknoloji değil, içerik yapısının düzenli ve AI tarafından kolay okunabilir olmasıdır.

### İlk Site MVP'si

İlk site MVP'si şunları içermelidir:

- Ana sayfa
- Kurulum sayfası
- Quick start
- Syntax reference
- CLI reference
- AI agents guide
- Örnek warehouse uygulaması
- Hata kodları sayfası
- `llms.txt`

### Uzun Vadeli Site Hedefi

BlackLang büyüdükçe site de dilin resmi hafızası olmalıdır.

Uzun vadede:

- Online playground
- `.black` kodunu tarayıcıda parse etme
- AST çıktısını gösterme
- Generated web karşılığını gösterme
- Versiyon seçici
- Migration guide
- Benchmark explorer
- Örnek uygulama galerisi

Bu yapı sayesinde BlackLang, sadece indirilen bir CLI değil; öğrenilebilir, aranabilir, güncel tutulabilir ve AI ajanlarının referans alabileceği tam bir dil ekosistemi haline gelir.

## Aşama 1: Minimal Dil Çekirdeği

Bu aşamada BlackLang sadece basit veri modeli ve sayfa tanımı yapabilir.

### Eklenecek Keywordler

- `app`
- `entity`
- `page`
- `source`
- `table`
- `columns`
- `form`
- `fields`
- `actions`
- `search`

### Desteklenecek Field Tipleri

- `text`
- `number`
- `integer`
- `decimal`
- `money`
- `email`
- `boolean`
- `date`
- `datetime`

### Desteklenecek Field Kuralları

- `required`
- `unique`
- `default`
- `optional`

### Örnek

```black
app Warehouse

entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
}

page Products {
  source Product

  table {
    columns sku, name, stock, price
    search sku, name
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete
}
```

### Üretilecek Web Karşılığı

- TypeScript entity tipleri
- Basit database schema
- CRUD API route
- React liste sayfası
- Arama inputu
- Create formu
- Edit davranışı
- Delete davranışı

### Bu Aşamada Bilerek Eklenmeyecekler

- Auth
- Role sistemi
- Complex relation
- Pagination
- File upload
- Test generation
- Deployment

## Aşama 2: Parser ve AST

Bu aşamada `.black` dosyaları okunabilir hale gelir.

### Eklenecekler

- Lexer
- Token stream
- Parser
- AST üretimi
- Syntax error raporu
- Satır ve kolon bilgisi
- JSON AST çıktısı

### Mevcut Durum

Draft v0.1 içinde parser artık ham satır bölme mantığına doğrudan bağlı değildir. Kaynak önce lexer tarafından token stream'e ayrılır; sonra `{`, `}`, virgül, operatör, quoted string, newline ve comment kuralları deterministik statement'lara dönüştürülür.

Bu sayede:

- Tırnak içindeki `#` ve `//` comment sayılmaz.
- Inline `{}` kullanımları statement olarak ayrılır.
- Kapanmamış string için `UNCLOSED_STRING` hatası üretilir.
- Parser ileride büyüyen syntax için daha güvenli bir tabana sahip olur.

### CLI Komutları

```bash
black parse app.black
black parse app.black --json
```

### AI İçin JSON Çıktısı

```json
{
  "success": true,
  "app": "Warehouse",
  "entities": ["Product"],
  "pages": ["Products"]
}
```

### Amaç

AI, dosyanın gerçekten parse edilip edilmediğini net biçimde görebilmelidir.

## Aşama 3: Validator

Parser syntax'ı okur; validator ise anlam hatalarını bulur.

### Kontrol Edilecekler

- Aynı entity iki kez tanımlanmış mı?
- Aynı field iki kez tanımlanmış mı?
- Page içinde belirtilen `source` var mı?
- Table columns gerçek fieldlara karşılık geliyor mu?
- Form fields gerçek fieldlara karşılık geliyor mu?
- `unique` sadece uygun fieldlarda mı?
- `default` değeri field tipiyle uyumlu mu?
- `search` alanları aranabilir tipte mi?
- `actions` desteklenen değerlerden mi oluşuyor?

### CLI Komutları

```bash
black validate
black validate --json
```

### AI İçin Hata Formatı

```json
{
  "success": false,
  "errors": [
    {
      "file": "app.black",
      "line": 17,
      "column": 13,
      "code": "UNKNOWN_FIELD",
      "message": "Page Products uses unknown field barcode.",
      "suggestion": "Add barcode to Product or remove it from columns."
    }
  ]
}
```

## Aşama 4: İlk Web Generator

Bu aşamada BlackLang ilk kez gerçek çalışan web uygulaması üretir.

### Hedef Stack

İlk generator tek bir stack hedefleyebilir:

- React
- TypeScript
- Express veya Fastify
- Prisma
- SQLite, PostgreSQL veya MySQL
- Vite

### Üretilecek Dosyalar

```text
generated/
├── .env.example
├── package.json
├── index.html
├── tsconfig.json
├── vite.config.ts
├── prisma/
│   └── schema.prisma
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── db.ts
│   ├── server.ts
│   ├── styles.css
│   ├── types.ts
│   ├── api/
│   │   └── product.ts
│   ├── routes/
│   │   └── product.ts
│   ├── pages/
│   │   └── ProductsPage.tsx
│   └── validation/
│       └── product.ts
└── README.md
```

### CLI Komutları

```bash
black build
black build --target web
npm run db:generate
npm run db:validate
npm run db:push
```

### Amaç

İlk gerçek kanıt burada çıkar:

> Tek bir `.black` tanımı çalışan web uygulamasına dönüşür.

Bu aşamada generator ayrıca veritabanı çalışma akışını standartlaştırır:

- `.env.example` üretir.
- Prisma schema dosyasını üretir.
- `db:generate`, `db:validate` ve `db:push` scriptlerini `package.json` içine ekler.
- `npm run build` komutu önce Prisma client üretir, sonra web uygulamasını derler.

## Aşama 5: CRUD Derinleştirme

Bu aşamada basit CRUD, gerçek uygulama seviyesine yaklaştırılır.

### Eklenecekler

- `list`
- `create`
- `read`
- `update`
- `delete`
- `bulkDelete`
- `duplicate`
- `archive`
- `restore`
- Soft delete
- Created/updated timestamp
- Empty state
- Loading state
- Error state

### Bu Aşamada Tamamlanan İlk Parça

- React sayfası artık generated API client üzerinden `list`, `create`, `update` ve `delete` çağırabilir.
- `src/api/<entity>.ts` client katmanı üretilir.
- `src/server.ts` Express server entry dosyası üretilir.
- `src/db.ts` Prisma Client singleton dosyası üretilir.
- `src/setup-db.ts` SQLite tablo hazırlama dosyası üretilir.
- `vite.config.ts` geliştirme sırasında `/api` isteklerini backend server'a yönlendirir.
- Sayfaya loading, saving ve error state eklenir.
- Henüz yeni BlackLang syntax'ı eklenmez; mevcut `actions create, edit, delete` davranışı güçlendirilir.

### Bu Aşamada Tamamlanan İkinci Parça

- API route'ları geçici memory array yerine Prisma Client kullanacak şekilde üretilir.
- `list`, `read`, `create`, `update` ve `delete` endpointleri database modeline bağlanır.
- SQLite hedefinde `decimal` ve `money` alanları MVP için `Float` olarak üretilir.
- Yerel Windows ortamında Prisma schema-engine `db push` adımında boş hata verdiği için MVP'de `db:push`, BlackLang'in ürettiği deterministik SQLite setup scriptine bağlanır.
- Native `prisma db push` tekrar değerlendirildi. Prisma 7.10 ile schema-engine boş hata vermiyor; ancak mevcut auth/audit tabloları Prisma schema dışında üretildiği için doğrudan `db push` veri kaybı uyarısıyla durabiliyor. Bu nedenle `db:push` güvenli BlackLang setup alias'ı olarak kalır, `db:push:native` ise bilinçli kontrol için ayrıca üretilir.
- Entity içinde `index field` ve `index fieldA, fieldB` syntax'ı desteklenir.
- Generator Prisma schema içinde deterministic `@@index([...], map: "..._idx")`, SQLite setup içinde `CREATE INDEX IF NOT EXISTS` üretir.
- Relation field index'leri generated `<field>Id` kolona maplenir; computed display field'lar database kolonu olmadığı için indexlenemez.
- `black inspect app.black --affected Entity.index --json`, index değişikliğinin generated database schema/setup etkisini raporlar.
- `black migrate plan old.black new.black --json|--ir`, iki `.black` kaynak arasındaki entity, field, relation, required/default/unique, index ve explicit rename farklarını read-only şekilde karşılaştırır.
- Migration plan JSON çıktısı `success`, `safe`, `destructive`, `summary`, `changes`, `steps` ve `changes[].risk` alanlarıyla safe/manual/destructive ayrımı yapar.
- Komut database'e bağlanmaz; explicit `migration Name { rename entity Old to New; rename field Entity.old to new }` deklarasyonları generated `db:setup`/`db:push` sırasında schema setup öncesi uygulanır. Rename tahmin edilmez.

### Bu Aşamada Tamamlanan Üçüncü Parça

- API client `get(id)` fonksiyonu üretir.
- Table action alanına `View` butonu eklenir.
- Seçili kayıt için detail panel üretilir.
- Detail panel `GET /api/<page>/<id>` endpointi üzerinden veriyi yeniden okur.
- Delete sonrası seçili kayıt temizlenir; update sonrası detail panel güncel kayıtla senkron kalır.

### Bu Aşamada Tamamlanan Dördüncü Parça

- `actions delete` varsa tablo seçim kutuları üretir.
- Görünen kayıtları tek seferde seçme davranışı üretir.
- API client `bulkDelete(ids)` fonksiyonu üretir.
- API route `DELETE /api/<page>` endpointiyle çoklu kayıt siler.
- Bulk delete sonrası tablo, seçili kayıtlar ve detail panel senkron temizlenir.

### Bu Aşamada Tamamlanan Beşinci Parça

- `archive` ve `restore` desteklenen action listesine eklenir.
- Prisma schema `archivedAt` soft-delete alanı üretir.
- SQLite setup scripti `archivedAt` kolonu üretir.
- API route `PATCH /api/<page>/<id>/archive` ve `PATCH /api/<page>/<id>/restore` endpointlerini üretir.
- Varsayılan liste arşivlenmiş kayıtları göstermez.
- UI `Show archived`, `Archive` ve `Restore` davranışlarını üretir.

### Örnek

```black
actions {
  create
  edit
  delete
  archive
  restore
}
```

### AI Açısından Değer

AI, CRUD davranışını yeniden yazmak zorunda kalmaz. Sadece hangi davranışların istendiğini belirtir.

## Aşama 6: Relation Sistemi

Web uygulamalarında entity ilişkileri kritik önemdedir.

### Eklenecek Relation Tipleri

- One-to-one
- One-to-many
- Many-to-one
- Many-to-many

### Örnek

```black
entity Customer {
  name text required
  email email unique
}

entity Order {
  customer Customer required
  total money
  status text default "draft"
}
```

### Üretilecekler

- Database foreign key
- Join query
- Form select input
- Detail page relation display
- API include parametreleri
- Validation

### Bu Aşamada Tamamlanan İlk Parça

- Entity field tipi olarak mevcut entity adları kabul edilir.
- `customer Customer required` gibi many-to-one ilişki temeli desteklenir.
- Validator bilinmeyen entity referanslarını hata olarak bırakır.
- Prisma schema relation field, foreign key field ve ters relation alanı üretir.
- SQLite setup scripti foreign key kolonunu üretir.
- Form select input ve relation display davranışları generated web çıktısında desteklenir.

### Bu Aşamada Tamamlanan İkinci Parça

- Relation field kullanılan form alanları artık metin input yerine select input olarak üretilir.
- Generated React sayfası relation hedefindeki kayıtları API client ile yükler.
- Form kaynağı BlackLang'de `customer` olarak okunabilir kalır; API tarafına gereken `customerId` payload'u generator üretir.
- Table ve detail panel relation alanlarında mümkün olduğunda ilişkili kaydın okunabilir etiketi gösterilir.
- Birden fazla `page` tanımı olduğunda generated React app sayfalar arası navigation üretir.

### Bu Aşamada Tamamlanan Üçüncü Parça

- Relation field'lar `load list`, `load detail`, `load query`, `load mutation` veya `load none` ile generated API response attach context'lerini deterministic seçebilir.
- `load` yazılmazsa backward-compatible varsayılan davranış korunur ve relation obje list/detail/query/mutation response'larında attach edilir.
- `load none`, response içinde yalnız generated relation ID field'ını bırakır.
- Generated list/query route'lar relation ID'lerini batch toplayıp her relation field için tek target lookup yapar; permission-aware build attached target record'ları sanitize eder.
- Relation load policy database schema'yı değiştirmez ve query filter/sort için relation join syntax'ı eklemez.
- OpenAPI operations `x-blacklang-relation-load-context` ve `x-blacklang-relation-load` metadata'sı üretir.

### Bu Aşamada Tamamlanan Dördüncü Parça

- Zorunlu relation field için hedef entity'de kayıt yoksa generated form submit butonunu devre dışı bırakır.
- Relation select alanı boş seçenek listesinde disabled hale gelir.
- Generated UI kullanıcıya hangi relation kaydının önce oluşturulması gerektiğini kısa mesajla söyler.
- Bu davranış yeni syntax eklemeden mevcut `required` bilgisinden türetilir.

### Bu Aşamada Tamamlanan Dördüncü Parça

- Zorunlu relation için hedef entity'nin kendi `page` tanımı varsa generated UI bu sayfaya geçiş butonu üretir.
- `Order.customer` örneğinde müşteri yoksa Orders formu `Open Customers` butonu gösterebilir.
- Generated `App` sayfa bileşenlerine navigation callback'i geçirir.
- Bu davranış relation ve page tanımlarından türetilir; yeni BlackLang syntax'ı eklenmez.

### Bu Aşamada Tamamlanan Beşinci Parça

- Relation field artık `table.search` içinde kullanılabilir.
- Generated table search relation alanında foreign key yerine ilişkili kaydın okunabilir etiketini arar.
- `search customer, status` gibi bir tanım Orders ekranında müşteri adı ve statü üzerinden filtreleme üretir.
- CLI docs `search` maddesi entity reference arama desteğini bildirir.

## Aşama 7: Form Sistemi

Formlar web uygulamalarının en yoğun tekrar alanlarından biridir.

### Eklenecekler

- Field label
- Placeholder
- Help text
- Required validation
- Min/max validation
- Pattern validation
- Select
- Multi-select
- Checkbox
- Radio
- Date picker
- File input
- Conditional fields
- Form sections
- Inline validation

### Örnek

```black
form ProductForm {
  source Product

  section "Temel Bilgiler" {
    sku label "SKU" required
    name label "Ürün Adı" required
  }

  section "Stok" {
    stock min 0
    price min 0
  }
}
```

### Bu Aşamada Tamamlanan İlk Parça

- Entity field satırlarında `label "Text"` modifier'ı desteklenir.
- Parser tırnak içindeki çok kelimeli label değerlerini tek değer olarak okur.
- Validator `label` modifier'ını tanır ve değeri eksikse hata üretir.
- Generated form label metinleri field adından değil, varsa `label` modifier'ından üretilir.
- Aynı label bilgisi table header ve detail alanlarında da kullanılır.

### Bu Aşamada Tamamlanan İkinci Parça

- Entity field satırlarında `placeholder "Text"` modifier'ı desteklenir.
- Parser tırnak içindeki çok kelimeli placeholder değerlerini tek değer olarak okur.
- Validator `placeholder` modifier'ını tanır ve değeri eksikse hata üretir.
- Generated form input alanları placeholder metnini kullanır.
- Relation select alanlarında placeholder boş option metni olarak kullanılır.

### Bu Aşamada Tamamlanan Üçüncü Parça

- Entity field satırlarında `help "Text"` modifier'ı desteklenir.
- Parser tırnak içindeki çok kelimeli help değerlerini tek değer olarak okur.
- Validator `help` modifier'ını tanır ve değeri eksikse hata üretir.
- Generated form alanları help metnini input/select altında kalıcı açıklama olarak gösterir.
- Bu metadata AI'nin alan amacını okumasını da kolaylaştırır.

### Bu Aşamada Tamamlanan Dördüncü Parça

- Generated React formları artık alan bazlı inline validation mesajları üretir.
- Mesajlar `required`, `email`, `number`/`money` ve zorunlu relation kurallarından gelir.
- Form invalid olduğunda API çağrısı yapılmadan kullanıcıya hangi alanın düzeltilmesi gerektiği gösterilir.
- Bu davranış backend validation ile aynı `.black` kaynak bilgisinden üretildiği için AI ajanı kuralı tek yerde okur.

## Aşama 8: Table ve Listeleme Sistemi

Tablolar, admin panelleri ve iş uygulamaları için ana yüzeydir.

### Eklenecekler

- Columns
- Sorting
- Filtering
- Search
- Pagination
- Column visibility
- Row actions
- Bulk actions
- Export CSV
- Empty state
- Cell formatting
- Status badge
- Relation column

### Örnek

```black
table ProductTable {
  source Product

  columns sku, name, stock, price
  search sku, name
  sort createdAt desc
  paginate 25

  rowActions edit, delete
  bulkActions delete, export
}
```

### Bu Aşamada Tamamlanan İlk Parça

- `table` blokları içinde `sort field asc|desc` syntax'ı desteklenir.
- Parser sort field ve direction bilgisini AST'ye ekler.
- Validator sort alanının source entity içinde var olduğunu ve yönün `asc` ya da `desc` olduğunu kontrol eder.
- Generated React listeleri önce search ile filtreler, sonra default sort sırasını uygular.
- Relation alanlarında sort, foreign key yerine okunabilir relation etiketi üzerinden yapılır.

### Bu Aşamada Tamamlanan İkinci Parça

- `table` blokları içinde `paginate number` syntax'ı desteklenir.
- Parser pagination değerini pozitif tam sayı olarak okur.
- Generated React listeleri search ve sort işleminden sonra kayıtları sayfalara böler.
- UI, Previous/Next butonları ve mevcut sayfa bilgisini üretir.
- Arama ya da archive filtresi değiştiğinde liste ilk sayfaya döner.

### Bu Aşamada Tamamlanan Üçüncü Parça

- Generated tablolar artık kolon görünürlüğü kontrolleri üretir.
- Her `columns` alanı varsayılan olarak görünür başlar.
- Kullanıcı generated UI içinde kolonları açıp kapatabilir.
- Bu davranış yeni syntax eklemeden mevcut `columns` listesinden türetilir.

### Bu Aşamada Tamamlanan Dördüncü Parça

- `table` blokları içinde `filter field...` syntax'ı desteklenir.
- Parser filter alanlarını AST'ye ekler.
- Validator filter alanlarının source entity içinde var olduğunu kontrol eder.
- Generated React listeleri global search sonrası field bazlı filtre uygular.
- Relation filtreleri foreign key yerine okunabilir relation etiketi üzerinden çalışır.

## Aşama 9: Sayfa ve Layout Sistemi

Bu aşamada uygulama sadece tek sayfadan çıkıp gerçek bir web app haline gelir.

### Eklenecekler

- Layout
- Sidebar
- Topbar
- Navigation
- Breadcrumb
- Tabs
- Detail page
- Dashboard page
- Settings page
- Modal
- Drawer
- Responsive davranış

### Örnek

```black
layout AdminLayout {
  sidebar {
    item Dashboard
    item Products
    item Customers
    item Orders
  }
}

page Products {
  layout AdminLayout
  source Product
  table ProductTable
}
```

### Bu Aşamada Tamamlanan İlk Parça

- Generated app artık page listesinden ortak bir application shell üretir.
- Shell içinde sidebar navigation, topbar ve breadcrumb yer alır.
- Aktif page state'i shell seviyesinde tutulur.
- Bu adım henüz yeni `layout` syntax'ı eklemez; mevcut `page` tanımlarından otomatik türetilir.

### Bu Aşamada Tamamlanan İkinci Parça

- Top-level `layout Name { ... }` syntax'ı desteklenir.
- `sidebar { item PageName }` ile sidebar navigation sırası tanımlanabilir.
- `page` blokları içinde `layout LayoutName` referansı kullanılabilir.
- Validator bilinmeyen layout, bilinmeyen sidebar page ve tekrar eden sidebar item hatalarını yakalar.
- Generated App sidebar sırasını explicit layout içindeki item sırasından üretir.

### Bu Aşamada Tamamlanan Üçüncü Parça

- Generated app shell küçük ekranlarda responsive drawer navigation üretir.
- Desktop görünümde sidebar sabit kalır.
- Mobil görünümde topbar içinde Menu butonu çıkar.
- Menü açıldığında sidebar drawer olarak görünür ve backdrop ile kapanabilir.
- Kullanıcı bir sayfaya geçtiğinde drawer otomatik kapanır.

## Aşama 10: API Sistemi

BlackLang sadece UI değil, API davranışlarını da tanımlayabilmelidir.

### Eklenecekler

- REST endpoint üretimi
- Query params
- Path params
- Request body
- Response shape
- Error shape
- Rate limit
- Public/private endpoint
- Webhook endpoint
- OpenAPI üretimi

### Şu An Eklenen İlk Parça

Draft v0.1 artık generated web çıktısına `openapi.json` dosyası ekler.

Bu dosya mevcut `entity`, `page` ve `actions` bloklarından otomatik üretilir:

- Sayfa kaynağı API şemasını belirler.
- Sayfa adı REST path adını belirler.
- `actions` listesi hangi create/edit/delete/archive/restore endpoint'lerinin sözleşmede görüneceğini belirler.
- Relation alanları request body içinde `customerId` gibi ID alanları olarak gösterilir.

Generated Express server bu sözleşmeyi şu adresten sunar:

```text
/openapi.json
```

Draft v0.2 ayrıca explicit API bloklarını generated declared-runtime route olarak okuyabilir:

```black
api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body tenantId text required
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku {
    value incomingStock = body.stock
    if incomingStock >= 0 and tenantId == body.tenantId
      set stock = incomingStock
    else
      set stock = stock
  }
  respond accepted
  webhook
  public
}
```

Bu bloklar declared-runtime çalışır ve POST/PUT/PATCH için bounded update handler taşıyabilir:

- Parse JSON çıktısında görünür.
- BlackIR çıktısında görünür.
- `black inspect` ve `black inspect --affected` çıktısında özetlenir.
- `generated/openapi.json` içine path, query param, path param, typed body schema, public/private metadata, `x-blacklang-runtime: declared`, `x-blacklang-handler` ve webhook metadata olarak yazılır.
- Generated Express server path/query/body parametrelerini doğrulayan deterministic route üretir.
- `update Entity where field == value set field = value` veya block-form `update Entity where field == value { value name = expression; if ... }` handler'ı `id` veya stored `unique` field ile tek bounded row seçer ve stored primitive non-policy field'ları günceller.
- Block-form update handler ordered local `value`, indentation-based `if`/`else` ve nested deterministic branch statement'larını destekler.
- Webhook route'ları `202 accepted` döner.
- `private` route'lar projede `auth` varsa auth/CSRF middleware arkasında çalışır.

### Örnek

```black
api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}
```

## Aşama 11: Auth ve Kullanıcı Sistemi

Gerçek web uygulamaları için authentication zorunludur.

Auth syntax'ına geçmeden önce generated API server güvenli varsayılanlarla başlamalıdır.

### Şu An Eklenen Güvenlik Zemini

Draft v0.1 generated Express server artık şu temel korumaları otomatik üretir:

- `X-Powered-By` header'ını kapatma
- Temel browser security header'ları
- `100kb` JSON request body limiti
- Basit IP bazlı rate limit

Bu parçalar ileride `auth`, `role`, `permission` ve audit log sisteminin üzerine oturacağı güvenlik tabanıdır.

### Şu An Eklenen Auth Dili

Draft v0.1 artık `auth` bloğunu parse ve validate eder:

```black
auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}
```

Compiler artık authentication niyetini kaynak dilin parçası olarak okuyabilir, doğrulayabilir ve AI-readable çıktılarda gösterebilir.

Draft v0.1 bu auth niyetinden temel login/register UI shell üretir.

Draft v0.1 ayrıca register, login, logout ve current-user API endpoint'leri üretir. Parolalar hashlenir ve session bilgisi cookie üzerinden saklanır.

Draft v0.1 generated CRUD API route'larını cookie session ile korur. Frontend açılışta `/api/auth/me` çağırarak mevcut session'ı kontrol eder ve logout davranışı üretir.

Draft v0.1 cookie auth için CSRF koruması üretir. Password reset ve OAuth sonraki auth aşamalarında eklenecektir.

### Eklenecekler

- User entity
- Login
- Register
- Logout
- Session
- Password reset
- Email verification
- OAuth provider
- API token
- Two-factor auth

### Örnek

```black
auth {
  strategy emailPassword
  session cookie

  user {
    name text
    email email unique
  }
}
```

## Aşama 12: Permission ve Role Sistemi

AI ajanlarının en çok hata yapabileceği alanlardan biri yetkilendirmedir. Bu yüzden BlackLang'de permission çok açık olmalıdır.

### Eklenecekler

- Role
- Permission
- Resource permission
- Field-level permission
- Page access
- Action access
- Ownership rule
- Tenant rule

### Örnek

```black
role Admin {
  allow all
}

role WarehouseWorker {
  allow read Product
  allow update Product.stock
  deny delete Product
}

page Products {
  source Product
  access Admin, WarehouseWorker
}
```

### Mevcut Durum

Bu aşamanın ilk parçası eklendi. BlackLang artık `role` bloklarını ve `page` içindeki `access` satırlarını okuyup doğrulayabilir.

Şu anda desteklenen permission action değerleri:

- `all`
- `manage`
- `read`
- `create`
- `update`
- `delete`

Mevcut v0.1 davranışı:

- Roller top-level tanımlanır.
- Sayfalar `access Admin, Worker` şeklinde rollere bağlanır.
- Permission resource değeri mevcut entity adını göstermelidir.
- `authenticated` access kullanılabilir ama bunun için `auth` bloğu gerekir.
- JSON ve BlackIR çıktıları bu bilgiyi taşır.

Runtime enforcement tarafının ilk parçası da eklendi:

- `BlackUser` tablosu primary `role` değeri ve `roles` seti saklar.
- Yeni kayıt olan kullanıcıya ilk tanımlı rol atanır.
- `/api/auth/me` primary rolü ve atanmış rol setini döndürür.
- `access` tanımlı sayfaların API route'ları role göre korunur.
- Yetkisiz role sahip kullanıcı `403 Forbidden` cevabı alır.
- Roller varsa generated uygulama temel bir Users ekranı üretir.
- İlk tanımlı rol kullanıcıları listeleyebilir ve bir kullanıcıya bir veya birden fazla rol atayabilir.
- Tenant policy varsa aynı Users ekranı kullanıcı `tenantId` değerini düzenleyebilir.
- Permission action tarafı çalışır:
  - `read` list/detail endpointlerini korur.
  - `create` create endpointini korur.
  - `update` edit/archive/restore endpointlerini korur.
  - `delete` single ve bulk delete endpointlerini korur.
- Generated React sayfaları kullanıcının hiçbir rolü izin vermiyorsa create/edit/archive/restore/delete kontrollerini gizler.
- Permission kontrolü kullanıcının tüm rol setini değerlendirir; atanmış rollerden herhangi birindeki `deny` eşleşen `allow` kuralını ezer, aksi halde herhangi bir roldeki `allow` erişim verir.
- Field-level read hiding çalışır:
  - `deny read Product price` Product kayıtlarını okunabilir bırakır.
  - API response içinden `price` alanını çıkarır.
  - React table/detail/form görünümünden `price` alanını gizler.
- Field-level mutation enforcement çalışır:
  - `allow update Product stock` role update endpointine izin verir.
  - Generated API sadece izinli field değerlerini veritabanına yazar.
  - Yetkisiz gönderilen field değerleri yok sayılır.
  - Generated React edit formunda sadece güncellenebilir field alanları görünür.
- Audit log desteği çalışır:
  - Auth ve role sistemi varsa `BlackAuditLog` tablosu üretilir.
  - Create, update, archive, restore, delete, bulk delete, register ve role update işlemleri audit kaydı yazar.
  - İlk tanımlı rol generated Audit ekranından son aktiviteleri görebilir.
- CSRF/session koruması çalışır:
  - Cookie auth, HttpOnly session cookie ile okunabilir CSRF cookie'sini birlikte üretir.
  - Generated frontend yazma isteklerinde `X-CSRF-Token` header'ı gönderir.
  - Generated API, state-changing authenticated isteklerde cookie ve header eşleşmiyorsa `403` döndürür.

Gelişmiş yetki yönetimi hâlâ büyüyecektir. Owner/tenant entity policy, multi-role kullanıcı atama, tenant admin UI, secret reference manifest, read-only secret manager provider preflight execution ve signed release trust MVP tamamlandı; provider-owned runtime value injection sonraki adapter katmanında kalır.

## Aşama 13: Workflow Sistemi

İş uygulamalarında süreçler sadece CRUD değildir. Sipariş, onay, teslimat, ödeme gibi akışlar gerekir.

### Mevcut Durum

Bu aşamanın UI parçası da eklendi. BlackLang artık top-level `workflow` bloklarını parse ve validate eder; authenticated web çıktısında transition API route'ları ve satır action butonları üretir.

Şu anda desteklenen workflow parçaları:

- `source Entity`
- `states draft, picking, verified`
- `transition Name { from State to State allow Role }`

Compiler şu kontrolleri yapar:

- Workflow source entity var mı?
- Workflow source entity içinde `status text` var mı?
- State listesi boş mu?
- Aynı state tekrar ediyor mu?
- Aynı transition tekrar ediyor mu?
- Transition `from` ve `to` değerleri state listesinde var mı?
- `allow` içindeki roller mevcut mu?
- `allow` kullanılıyorsa auth bloğu var mı?

Runtime generator artık transition endpoint'i ve tablo satırı workflow butonları üretir. Generated button sadece satırın mevcut `status` değeri transition `from` değeriyle eşleştiğinde görünür. Generated route update permission kontrolü yapar, transition `allow` rollerini kontrol eder, mevcut `status` değerinin `from` ile eşleşmesini ister, `status` alanını `to` değerine günceller ve audit log'a `workflow.<transition>` kaydı yazar.

### Eklenecekler

- Workflow
- Step
- Transition
- Status
- Approval
- Rejection
- Side effect
- Notification
- Audit log

### Örnek

```black
workflow OrderPreparation {
  source Order

  states draft, picking, verified, packaged, shipped

  transition startPicking {
    from draft
    to picking
    allow WarehouseWorker
  }

  transition ship {
    from packaged
    to shipped
    allow Admin
  }
}
```

## Aşama 14: State ve Client Davranışları

Her şey backend modeli değildir. Web uygulamasında client state de gerekir.

### Mevcut Durum

Bu aşamanın ilk parçası eklendi. BlackLang artık top-level `state` bloklarını parse ve validate eder.

Şu anda desteklenen state parçaları:

- Primitive state field: `activeFilter text`
- Entity list state: `selectedOrders Order[]`
- Modal state: `modal createOrder closed`

Compiler şu kontrolleri yapar:

- Aynı state tekrar ediyor mu?
- Aynı state field tekrar ediyor mu?
- State field tipi primitive mi veya mevcut entity mi?
- Aynı modal tekrar ediyor mu?
- Modal default değeri `open` veya `closed` mu?

Explicit state declaration artık generated React state'e ilk seviyede bağlanır. `OrdersPageState` veya `OrdersState`, `page Orders` ile eşleşir. State field'ları `useState` hook'u üretir. `modal createOrder closed` gibi modal tanımları open/close helper'ları üretir ve ilgili create formunun görünürlüğünü kontrol edebilir.

### Eklenecekler

- Local UI state
- Filter state
- Selected rows
- Wizard step
- Modal open/close
- Optimistic update
- Cache invalidation
- Realtime refresh

### Örnek

```black
state ProductPageState {
  selectedProducts Product[]
  activeFilter text
  modal createProduct closed
}
```

## Aşama 15: Component Sistemi

BlackLang başlangıçta component yazdırmamalı, ama tekrar kullanılabilir UI parçalarını tanımlayabilmelidir.

### Mevcut Durum

Bu aşamanın ilk parçası eklendi. BlackLang artık top-level `component` bloklarını parse ve validate eder.

Şu anda desteklenen component parçaları:

- Input: `input stock number`
- Entity/list input: `input products Product[]`
- Variant: `variant low when stock < 10`

Compiler şu kontrolleri yapar:

- Aynı component tekrar ediyor mu?
- Aynı component input tekrar ediyor mu?
- Component input tipi primitive mi veya mevcut entity mi?
- Aynı variant tekrar ediyor mu?
- Variant satırında `when` koşulu var mı?

Component declaration artık standalone React component dosyasına dönüşür. Variant koşulları `.black` içinde deterministik niyet olarak korunur; `stock < 10` gibi basit `input operator literal` koşulları runtime class seçimine çevrilir. Tek input'u entity field adı ve tipiyle eşleşen component'ler generated table/detail rendering alanlarına otomatik bağlanır. Aynı eşleşme generated form alanlarında canlı component önizlemesi olarak da kullanılır. Page `view` içinde `section StockSummary component StockBadge bind selected` gibi satırlarla declared component artık reusable inline page section olarak da yerleştirilebilir.

### Eklenecekler

- Collection-backed component section
- Nested component slots
- Variant
- Reusable form
- Reusable table
- Card
- Status display
- Metric display

### Örnek

```black
component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}
```

## Aşama 16: Validation Sistemi

Validation hem backend hem frontend tarafında aynı kaynaktan üretilmelidir.

### Mevcut Durum

Bu aşamanın ilk parçası eklendi. BlackLang artık field modifier olarak `min`, `max` ve `length min..max` okuyabilir.

Şu anda desteklenen validation parçaları:

- Number-like field için `min`
- Number-like field için `max`
- Text/email field için `length 3..40`
- Text/email field için `regex "pattern"`
- Text field için `url`
- Field bazlı özel validation mesajı için `message "Text"`
- Entity içinde cross-field validation için `validate left <= right message "Text"`
- Entity içinde conditional validation için `validate field required when otherField == value message "Text"`
- Frontend inline validation mesajları
- Backend/API validation mesajları
- Generated HTML input attribute'ları

Örnek:

```black
entity Product {
  sku text required unique length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  stock number min 0
  price money min 0
  website text optional url
}

entity Order {
  total money min 0
  discount money min 0
  status text default draft
  trackingNumber text optional
  validate discount <= total message "Discount cannot exceed total"
  validate trackingNumber required when status == shipped message "Tracking number is required when shipped"
}
```

### Tamamlananlar

- Required
- Min/max
- Length
- Regex
- Email
- URL
- Unique
- Custom message
- Cross-field validation
- Conditional validation

### Örnek

```black
entity Product {
  sku text required unique length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  stock number min 0
  price money min 0
  website text optional url
}

entity Order {
  total money min 0
  discount money min 0
  status text default draft
  trackingNumber text optional
  validate discount <= total message "Discount cannot exceed total"
  validate trackingNumber required when status == shipped message "Tracking number is required when shipped"
}
```

## Aşama 17: Query, Action ve Data Logic

Gerçek uygulamalarda basit CRUD dışında özel sorgular gerekir.

### Mevcut Durum: Custom Query MVP

Bu aşama, düşük stok gibi özel listeleri `.black` kaynağında tanımlayıp generated sayfaya bağlamayı sağlar. Filtre, sıralama ve kayıt sınırı server tarafında uygulanır.

```black
query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

page LowStock {
  source Product
  query LowStockProducts

  table {
    columns sku, name, stock
    search sku, name
    paginate 10
  }
}
```

İlk kapsam:

- Top-level `query Name` tek bir `source Entity` bildirir; page aynı source ile `query Name` kullanır.
- `where` yalnızca saklanan primitive field'ları ve tipe uygun literal değerleri kabul eder; birden fazla satır AND ile birleşir.
- Text/email/date/datetime literal'ları tırnaklıdır. Numeric ve boolean değerler tırnaksızdır.
- `==` ve `!=` tüm desteklenen tiplerde; `<`, `<=`, `>`, `>=` numeric/date/datetime tiplerinde kullanılabilir.
- Opsiyonel `sort field asc|desc` tek field alır; eşit değerler `id asc` ile sıralanır. Sort yoksa varsayılan `id asc` olur.
- `limit` 1..1000 arası tam sayıdır; varsayılan 100'dür.
- Page için `GET /api/lowstock/query` ve `queryList` client metodu üretilir. Sayfaya bağlanmamış query bağımsız endpoint üretmez.
- Mevcut entity listesi ve CRUD davranışı korunur; relation seçenekleri normal entity listesinden yüklenir.
- Table search/filter/sort/pagination, query'nin döndürdüğü sınırlı liste üzerinde çalışır. Bağlı sayfadaki mutation sonrası liste yeniden sorgulanır.
- Auth, page access, entity read ve response field hiding kuralları uygulanır. Koşul veya sort alanında read izni yoksa query endpoint'i 403 döndürür.
- Query bir liste seçim kuralıdır; ownership/tenant authorization deklarasyonu değildir. Entity row policy varsa query, detail ve mutation endpoint'leri aynı owner/tenant scope üzerinden çalışır.
- JSON/BlackIR, inspect/affected, diagnostics, `docs query --json`, `explain query --json` ve generated OpenAPI desteği vardır.

Warehouse örneğindeki ayrı, salt okunur LowStock sayfası `stock < 10` olan en fazla 50 ürünü gösterir. Aynı query artık background query job tarafından da tüketilebilir. Ayrıntılı referans `docs/query.md` içindedir.

### Mevcut Durum: Background Query Jobs MVP

Bu aşama, generated web uygulamasında public endpoint açmadan periyodik, salt okunur query işlerini çalıştırmayı sağlar. İlk sürüm queue provider, retry sistemi veya external scheduler bağlamaz; `.black` kaynağı deterministic worker niyetini tarif eder.

```black
job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
}
```

İlk kapsam:

- Top-level `job Name` PascalCase kullanır ve diğer top-level sembollerle çakışamaz.
- Tek schedule syntax'ı `schedule every <integer> minutes|hours|days` şeklindedir.
- Desteklenen aralıklar `1..1440 minutes`, `1..168 hours` ve `1..365 days` olarak sınırlıdır.
- Tek run mode `run query QueryName` şeklindedir; query mevcut olmalıdır.
- Generated worker query'nin stored-field `where`, deterministic `sort` ve `limit` kurallarını uygular.
- Worker yalnızca kayıt ID'lerini seçer ve job adı, schedule, query, source, count, limit ve timestamp içeren kompakt JSON log üretir.
- `jobs/manifest.json`, `src/worker.ts`, `jobs:run`, `jobs:loop` ve OpenAPI root `x-blacklang-jobs` metadata üretilir.
- JSON/BlackIR, inspect/affected, diagnostics, `docs job --json`, `explain job --json`, IDE metadata, Warehouse örneği ve docs sitesi desteği vardır.
- Queue provider, retry, delayed queue, mutating job, external call, provider scheduler config, cron expression syntax ve arbitrary code bu MVP dışındadır.

Warehouse örneğinde `LowStockMonitor`, `LowStockProducts` query'sini 15 dakikalık deterministic schedule ile worker'a bağlar. Ayrıntılı referans `docs/job.md` içindedir.

### Mevcut Durum: Custom Action MVP

Bu aşama, CRUD dışında satır bazlı domain işlemlerini `.black` kaynağında tanımlamayı sağlar. Örneğin Warehouse içinde `RestockProduct`, seçilen ürünün stok alanını doğrulanmış bir input ile artırır.

```black
action RestockProduct {
  source Product
  input quantity number required min 1 label "Quantity"
  value restockValue = quantity
  if restockValue > 0 and stock >= 0
    set stock = stock + restockValue
  else
    set stock = stock
  allow Admin, Worker
  success "Stock updated"
}

transaction RestockAtomic {
  action RestockProduct
}

page Products {
  source Product
  actions edit, RestockProduct
}
```

İlk kapsam:

- Top-level `action Name` tek bir `source Entity` bildirir; page aynı source ile `actions ..., Name` kullanır.
- Action input'ları primitive field type ve validation modifier kullanır.
- Input adı kaynak entity field adıyla çakışamaz; expression identifier'ları deterministik kalır.
- `value`, action içinde ordered local expression adı üretir.
- Indentation-based `if`/`else` branch'leri `value`, `set` ve nested `if` statement'larını taşıyabilir.
- Condition comparison'ları `and`, `or`, `not` ve parantezle bağlanabilir; precedence sırası `not`, sonra `and`, sonra `or` şeklindedir.
- `set` yalnızca saklanan primitive source field'larını değiştirir.
- Expression değerleri typed literal, action input, source field veya local value olabilir.
- Numeric expression için `+`, `-`, `*`, `/`, parentheses ve deterministic precedence desteklenir.
- Top-level `transaction Name { action ActionName }` bloğu generated route içinde source row lookup, update ve auth audit log yazımını tek Prisma transaction olarak çalıştırır.
- Computed display field, relation field, entity policy field, join, aggregate, raw SQL, arbitrary JS ve multi-row mutation bu MVP kapsamı dışındadır.
- Page'e bağlanmış action için POST route, API client metodu, backend validator, React row button/form panel, OpenAPI schema ve auth/role varsa audit log üretilir.
- Auth, page access, entity update permission, field-level update permission, entity row policy ve opsiyonel action `allow` kuralı birlikte uygulanır.
- Query-bound page üzerinde action başarılı olunca liste server query'den yeniden yüklenir.
- JSON/BlackIR, inspect/affected, diagnostics, `docs action --json`, `docs transaction --json`, `explain action --json`, `explain transaction --json` ve generated OpenAPI desteği vardır.

Warehouse örneğinde `RestockProduct`, Products ve LowStock sayfalarında kullanılabilir; `RestockAtomic` transaction bloğu hem bu action route’larını hem `StockWebhook` update handler’ını atomic hale getirir. Ayrıntılı referanslar `docs/action.md` ve `docs/transaction.md` içindedir.

### Mevcut Durum: Service Block MVP

Bu aşama, page kaynaklarına bağlı olmayan explicit API’leri ayrı bir service/module niyeti altında toplamayı sağlar. Endpoint hâlâ top-level `api` bloğunda tanımlanır; `service` bloğu yalnızca generated service metadata ve OpenAPI grouping üretir.

```black
service InventoryIntegration {
  api StockWebhook
}
```

İlk kapsam:

- Top-level `service Name` PascalCase kullanır.
- Her hedef satır `api APIName` şeklindedir.
- Service yalnızca mevcut explicit API deklarasyonlarını hedefler.
- Bir API en fazla bir service bloğuna bağlanabilir.
- Service bloğu route, handler veya mutation tanımlamaz.
- Generated output `services/manifest.json` ve `src/services/<service>.ts` üretir.
- OpenAPI root `tags`, root `x-blacklang-services` ve operation-level `x-blacklang-service` metadata üretir.
- Contract testleri service manifest ve OpenAPI service metadata'sını doğrular.
- JSON/BlackIR, inspect/affected, diagnostics, `docs service --json`, `explain service --json`, IDE metadata, Warehouse örneği ve site desteği vardır.

Warehouse örneğinde `InventoryIntegration`, `StockWebhook` API’sini service metadata altında gruplar; `RestockAtomic` ise aynı API’nin atomic runtime boundary’sini ayrı tutar. Ayrıntılı referans `docs/service.md` içindedir.

### Mevcut Durum: Owner/Tenant Entity Policy MVP

Bu aşama, generated web route'larında satır kapsamını `.black` kaynağında açıkça tanımlamayı sağlar:

```black
entity Order {
  tenantId text required default "default"
  ownerId text required default "system"
  total money default 0
  status text default draft
  policy tenant tenantId
  policy owner ownerId
}
```

İlk kapsam:

- `policy owner field` ve `policy tenant field` satırları entity içinde yazılır.
- Policy kullanımı `auth` gerektirir.
- Policy field'ı aynı entity üzerinde saklanan `text required` field olmalıdır ve `unique` olamaz.
- Generated form field listelerinde, API client input type'larında ve OpenAPI input schema'larında policy field'ları yoktur.
- Create/update route'ları policy field'larını current authenticated user context'ten damgalar.
- List, query, detail, archive, restore, delete, workflow transition ve custom action route'ları aynı row scope ile çalışır.
- Custom action `set` satırları policy field yazamaz.
- JSON/BlackIR, inspect/affected, diagnostics, `docs policy --json`, `explain policy --json`, Warehouse örneği ve docs sitesi desteği vardır.

Ayrıntılı referans `docs/policy.md` içindedir.

### Mevcut Durum: Ops Runtime Signals MVP

Bu aşama, generated web uygulamasının deploy edildikten sonra standart altyapı kontrolleriyle izlenebilmesini sağlar:

```black
ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

İlk kapsam:

- Top-level `ops` bloğu health, readiness, metrics, request logging ve external observability hook/exporter niyetini taşır.
- `health path` generated public GET endpoint üretir; process uptime, app adı, CLI version ve start time döner.
- `readiness path` database bağlantısını `SELECT 1` ile kontrol eder; hata varsa `503` döner.
- `metrics path` process-local request, error, status code, uptime ve start time değerlerini döner.
- `logging requests` her gözlenen istek bitince structured JSON log yazar.
- `observe webhook endpoint env NAME`, env set edilmişse her request sonunda non-blocking structured event POST eder.
- `observe otlp endpoint env NAME`, env set edilmişse her request sonunda non-blocking OTLP HTTP JSON trace payload POST eder.
- Observe middleware W3C `traceparent` context üretir veya gelen context'i taşır ve response `traceparent` header'ını yazar.
- Observability endpoint ve provider değerleri `.black` içinde literal secret/URL olarak tutulmaz; environment üzerinden okunur.
- Ops path'leri root-level public path olmalıdır; `/api`, `/openapi.json`, query string, fragment, brace, whitespace ve unsafe karakterler reddedilir.
- Ops route'ları API auth ve CSRF middleware'inden önce üretilir.
- OpenAPI içine `x-blacklang-ops` ve `x-blacklang-public` metadata yazılır.
- OpenAPI içine observe hook varsa `x-blacklang-observability` metadata yazılır.
- `ops/observability.json` provider, endpoint env, signal, protocol, trace context, delivery ve exporter metadata'sını kaydeder.
- Generated contract/API smoke testleri ops path, metadata, trace context ve local observability hook/exporter delivery kontrollerini çalıştırır.
- `deploy { target docker }` ve health path varsa Dockerfile ve docker-compose app healthcheck'i aynı health path'i probe eder.
- JSON/BlackIR, inspect/affected, diagnostics, `docs ops --json`, `explain ops --json`, IDE metadata, Warehouse örneği ve docs sitesi desteği vardır.

Ayrıntılı referans `docs/ops.md` içindedir.

### Mevcut Durum: Generated Contract/API/Frontend/Browser-Check/E2E/Matrix Test MVP

Bu aşama, `black build` ile üretilen web veya API-only uygulamasının kendi sözleşmesini ve temel HTTP davranışını doğrulamasını sağlar. `target web` ayrıca React render yüzeyini, declared browser-check beklentilerini, gerçek browser e2e akışını ve cross-browser matrix runner'ını doğrular. Generated proje `npm test` komutuyla OpenAPI, validation, job metadata, API server ve target'a uygun test yüzeylerini hızlıca smoke test eder; web target'ta `npm run test:e2e` tek selected browser üzerinden, `npm run test:e2e:plan` read-only matrix availability üzerinden, `npm run test:e2e:matrix` ise available Chrome/Edge/Chromium hedefleri üzerinden auth, navigation, text/page/action görünürlüğünü kontrol eder.

İlk kapsam:

- `target web` package script'i `npm run db:generate && tsx src/blacklang.contract.test.ts && tsx src/blacklang.api.test.ts && tsx src/blacklang.frontend.test.tsx` olarak üretilir; source içinde `test` varsa buna `tsx src/blacklang.browser.test.tsx` eklenir.
- `target api` package script'i yalnızca `npm run db:generate && tsx src/blacklang.contract.test.ts && tsx src/blacklang.api.test.ts` çalıştırır; React, frontend smoke, browser-check, e2e ve matrix test dosyaları üretilmez.
- Web source içinde `test` varsa `test:e2e`, `test:e2e:plan`, `test:e2e:matrix` ve `test:all` script'leri de üretilir; `test:e2e`, `db:generate` sonrası `tsx src/blacklang.e2e.test.ts` çalıştırır.
- Test dosyası `openapi.json` içindeki ops path/metadata, observability metadata, entity schema, custom action schema, CRUD route, bound query route, background job metadata, explicit API route ve custom action metadata kayıtlarını doğrular.
- Test dosyası generated validation fonksiyonlarını import eder ve temsilî geçerli/geçersiz entity/action payload'larını kontrol eder.
- API smoke test generated Express app'i random localhost portunda açar; `/openapi.json`, explicit API runtime route'ları, ops endpoint'leri, local observability hook delivery, anonymous API status, auth yoksa JSON 404 ve varsa CORS allow/deny davranışını kontrol eder.
- Frontend smoke test `target web` içinde generated React `App` bileşenini `react-dom/server` ile render eder.
- Browser-check smoke test `target web` içinde generated page metadata, target page action listesi ve generated text/render catalog üzerinden declared `expect text`, `expect page` ve `expect action` beklentilerini doğrular.
- Browser e2e test `target web` içinde ephemeral SQLite database kullanır, setup/seed modüllerini çalıştırır, generated API server ve Vite dev server'ı random localhost portlarında açar, auth varsa deterministic kullanıcı kaydı yapar, page navigation ve action button görünürlüğünü gerçek DOM'da doğrular.
- Browser matrix plan `target web` içinde `tests/browser-matrix.json` ve `src/blacklang.e2e.matrix.ts` üzerinden custom, Chrome, Edge ve Chromium executable availability bilgisini read-only JSON olarak raporlar.
- Browser matrix run aynı generated e2e intent'ini her available hedef için çalıştırır; missing hedefleri skip eder ve stdout/stderr tail alanlarıyla compact failure triage üretir.
- `number` ve `integer` alanlarında kesirli değerlerin reddedildiği smoke test edilir.
- Bu testler generated web/API çıktısı için deterministic contract/API/frontend/browser-check/e2e/matrix test katmanıdır. `test` declaration yine arbitrary browser step DSL değildir; generated auth, setup/seed, navigation, text/page/action kontrolleriyle sınırlıdır ve `target api` içinde unsupported diagnostic verir.

### Mevcut Durum: Benchmark Command MVP

Bu aşama, BlackLang kaynak boyutu, generated output boyutu ve AI task/token farkını tahmine bırakmadan ölçmeyi sağlar.

İlk kapsam:

- `black benchmark [file] --out <dir> --json|--ir` komutu vardır.
- `black benchmark tasks [file] --out <dir> --json|--ir` komutu vardır.
- `black benchmark eval [file] --out <dir> --json|--ir` komutu vardır.
- `black benchmark eval-history [--history <file>] --json|--ir` komutu vardır.
- Komut projeyi parse/validate eder ve generated output'u geçici dizinde üretip ölçer.
- Configured generated output dizinini değiştirmez.
- JSON çıktı `source`, `generated`, `ratios`, `sourceFiles`, `generatedFiles` ve `generatedKinds` alanlarını içerir.
- Task benchmark JSON çıktısı `baseline`, `scenarios` ve `totals` alanlarını içerir.
- Eval corpus JSON çıktısı `suite`, `cases` ve `totals` alanlarını içerir; case prompt, expected evidence, required commands ve 100 puanlık scoring rubric taşır.
- Eval history JSON çıktısı `history`, `summary` ve `runs` alanlarını içerir; evidence path, validation/model source, pass/fail total ve scoring policy taşır.
- İlk task benchmark seti 8 deterministic scenario üretir: query report, custom action, seed fixture, browser-check, row policy, explicit API handler, ops probes ve local preview/rollback metadata.
- Token estimate raporu current source/generated byte-line oranlarından ve fixed scenario context boyutlarından türetilen planning signal'dır; billed-token ölçümü değildir.
- Eval corpus read-only metadata'dır; AI model çağırmaz, configured output dizinini değiştirmez, commit/push/deploy yapmaz ve secret saklamaz.
- Eval history published-local manifest'i `benchmarks/eval-history.blackdir` içinden okunur; local validation entry'leri billed model benchmark skoru gibi sunulamaz.
- Compiler testleri Warehouse için kompakt golden manifest kullanır; dosya path/kind/satır/byte/SHA-256 değişimi review gerektirir.
- Warehouse ölçümünde collection-backed component section aşaması sonunda 770 `.black` kaynak satırı 65 generated dosyada 14503 satır üretmektedir; generated/black-source line ratio 18.84'tür.
- Warehouse task benchmark raporu 8 scenario için estimated BlackLang token total `10519`, conventional token total `90872`, estimated savings `%88` sinyali üretmektedir.

### Mevcut Durum: Seed / Fixture MVP

Bu aşama, demo ve test verisini `.black` içinde deterministic source intent olarak tutmayı sağlar. Generated setup schema işlemini bitirdikten sonra `db:seed` çalıştırır ve row key'leri stable id olarak upsert eder.

```black
seed DemoProducts {
  source Product

  row DemoProductLow {
    tenantId "default"
    sku "LOW-001"
    name "Low Stock Widget"
    stock 3
    price 19.99
  }
}
```

İlk kapsam:

- Top-level `seed Name` tek bir `source Entity` kullanır.
- `row RowKey` stable generated `id` olur.
- Scalar stored field değerleri typed literal olarak yazılır.
- Relation field değerleri `ref OtherRowKey` kullanır.
- Computed/system field, raw SQL, arbitrary code ve secret değerleri seed içinde desteklenmez.
- Generated web output `src/seed.ts`, `db:seed` ve `db:setup` wiring üretir.
- AI ajanları `black docs seed --json`, `black explain seed --json` ve `black inspect --affected DemoProducts --json` ile etkiyi okuyabilir.

### Mevcut Durum: Browser Test Declaration MVP

Bu aşama, generated web uygulamasının beklenen page/action/text yüzeyini `.black` içinde deterministic test niyeti olarak tutmayı sağlar. Amaç, AI ajanının “hangi sayfa ve aksiyon görünür olmalı?” sorusunu generated dosyaları elle okuyarak değil, source intent, JSON tooling, hızlı browser-check ve gerçek browser e2e koşumu üzerinden doğrulamasıdır.

```black
test WarehouseBrowserSmoke {
  page Products
  expect text "Depo"
  expect page LowStock
  expect action RestockProduct
}
```

İlk kapsam:

- Top-level `test Name` tek bir `page PageName` hedefler.
- `expect text`, `npm test` içinde generated React render çıktısı veya text catalog içinde literal arar; `npm run test:e2e` ve `npm run test:e2e:matrix` içinde gerçek DOM metnini bekler.
- `expect page`, `npm test` içinde declared page metadata varlığını; `npm run test:e2e` ve `npm run test:e2e:matrix` içinde gerçek navigation/page label görünürlüğünü kontrol eder.
- `expect action`, `npm test` içinde hedef page'in CRUD veya custom action listesinde action bulunduğunu; `npm run test:e2e` ve `npm run test:e2e:matrix` içinde action button görünürlüğünü doğrular.
- Generated web output `src/blacklang.browser.test.tsx` üretir ve `npm test` zincirine ekler; ayrıca `src/blacklang.e2e.test.ts`, `src/blacklang.e2e.matrix.ts`, `tests/browser-matrix.json`, `test:e2e`, `test:e2e:plan`, `test:e2e:matrix` ve `test:all` üretir.
- AI ajanları `black docs test --json`, `black explain test --json` ve `black inspect --affected WarehouseBrowserSmoke --json` ile etkiyi okuyabilir.
- Browser e2e execution gerçek Chrome/Chromium/Edge executable gerektirir; path otomatik bulunamazsa `BLACKLANG_E2E_BROWSER_PATH` kullanılabilir. Matrix hedefleri için `BLACKLANG_E2E_CHROME_PATH`, `BLACKLANG_E2E_EDGE_PATH` ve `BLACKLANG_E2E_CHROMIUM_PATH` kullanılabilir.

### Sonraki Genişleme

- Query parametreleri ve OR ifadeleri
- Aggregate, count, sum ve group by
- Join ve relation koşulları
- Computed field filtreleme/sıralama
- Server pagination ve total count
- Multi-step backend service/command orchestration
- External multi-model AI eval score history after independent harness runs
- Multi-step backend service/command syntax
- Cache ve realtime refresh

## Aşama 18: Dashboard ve Raporlama

Web iş uygulamalarında dashboard ve rapor ekranları sık görülür.

### Eklenecekler

- Metric
- Chart
- Table report
- Date range
- Aggregation
- Export
- Scheduled report

### Örnek

```black
dashboard WarehouseDashboard {
  metric TotalProducts count Product
  metric LowStock count Product where stock < 10

  chart StockByCategory {
    type bar
    from Product
    group category
    value stock sum
  }
}
```

## Aşama 19: Dosya ve Medya Yönetimi

Birçok web uygulaması dosya yükleme ihtiyacı duyar.

### Eklenecekler

- File upload
- Image upload
- File validation
- Storage provider
- Public/private file
- Image resize
- Attachment relation

### Örnek

```black
entity Product {
  name text required
  image file image optional
}

storage {
  provider local
  maxFileSize 5mb
}
```

## Aşama 20: Notification Sistemi

Uygulama içi ve dışı bildirimler ayrı bir declarative sistemle tanımlanmalıdır.

### Eklenecekler

- In-app notification
- Email notification
- SMS notification
- Web push
- Notification template
- Trigger
- Recipient rule

### Örnek

```black
notification LowStockAlert {
  when Product.stock < 10
  send email to Admin
  message "Stok seviyesi düştü: {{Product.name}}"
}
```

## Aşama 21: Event ve Automation Sistemi

Bu aşama iş mantığını daha güçlü hale getirir.

### Eklenecekler

- Event
- Trigger
- Scheduled job
- Background job
- Queue
- Retry
- Webhook call
- Audit event

### Örnek

```black
event ProductCreated {
  when Product created
  do create AuditLog
}

automation DailyStockCheck {
  schedule daily at "09:00"
  run LowStockReport
}
```

## Aşama 22: Realtime Özellikler

Bazı web uygulamaları canlı veri ister.

### Eklenecekler

- Realtime table refresh
- WebSocket channel
- Presence
- Live notification
- Live dashboard metric

### Örnek

```black
realtime {
  watch Product
  update pages Products, WarehouseDashboard
}
```

## Aşama 23: Error Handling ve Observability

AI'nin hata ayıklaması için uygulama kendi davranışını anlaşılır raporlamalıdır.

### Eklenecekler

- Error boundary
- API error format
- Logging
- Audit log
- Request tracing
- Metrics
- Health check
- Debug report

### Örnek

```black
ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

## Aşama 24: Test Sistemi

BlackLang'in en değerli taraflarından biri testleri de kaynaktan üretebilmesi olabilir.

### Eklenecekler

- Unit test
- API test
- Component test
- Browser-check test (ilk generated page/action/text MVP var)
- E2E test (gerçek browser automation eksik)
- Fixture (ilk seed syntax/runtime var)
- Seed data (ilk deterministic seed/runtime var)
- Permission test
- Workflow test

### Örnek

```black
test ProductCrud {
  create Product with sku "A-100", name "Keyboard"
  expect Product count 1
  update Product.stock to 5
  expect Product.stock equals 5
}
```

## Aşama 25: Migration ve Data Evolution

Uygulama büyüdükçe veri modeli değişir. Bu alan çok dikkatli tasarlanmalıdır.

### Eklenecekler

- Schema migration
- Rename field
- Rename entity
- Default backfill
- Required field migration
- Data transform
- Rollback
- Migration warning

### Mevcut MVP

Read-only schema migration plan, explicit rename runtime ve generated online migration runner tamamlandı:

```bash
black migrate plan old.black new.black --json
black migrate plan old.black new.black --ir
```

Bu komut parser/validator sonrası iki kaynak arasındaki generated database shape farklarını safe, manual ve destructive risklere ayırır. Database'e bağlanmaz. Entity/field rename niyeti current source içindeki migration bloklarıyla açık yazılır ve generated setup runtime tarafından schema setup öncesi idempotent uygulanır. Rollback ve data transform hâlâ gelecek genişletmelerdir.

Generated app içinde migration blokları varsa şu komutlar da üretilir:

```bash
npm run db:migrate:plan
npm run db:migrate
npm run db:setup
```

`db:migrate:plan` target database'e read-only bakar; SQLite'ta eksik DB dosyasını oluşturmaz, PostgreSQL connection string değerini JSON'a basmadan unreachable/ready bilgisini döndürür. Generated rename migration runner şimdilik SQLite ve PostgreSQL ile sınırlıdır; `database mysql` ile migration block birlikte kullanılırsa validator `UNSUPPORTED_TARGET_DATABASE_MIGRATION` raporlar. `db:migrate` yalnızca declared rename migration'larını uygular ve `BlackMigration` ledger'ına işler. Sonrasında `db:setup` normal generated setup ve seed yolunu çalıştırır.

### Örnek

```black
migration RenameSkuToBarcode {
  rename field Product.sku to barcode
}
```

## Aşama 26: Theming ve Design System

BlackLang görsel tasarımı da belirli kurallarla anlatabilmelidir, ama ilk sürümlerde aşırı serbest CSS yazdırmamalıdır.

### Eklenecekler

- Theme
- Color tokens
- Spacing tokens
- Typography tokens
- Component variants
- Dark mode
- Density mode
- Responsive breakpoint

### Örnek

```black
theme AdminTheme {
  color primary "#2563eb"
  color danger "#dc2626"
  radius 6
  density compact
}
```

### İleride Uygulanacak Karar: Generator UI Order Profile

AI'nin daha az token harcaması için component UI özellikleri uzun CSS benzeri anahtarlarla değil, sabit sıralı kısa satırla yazılabilir.

Örnek `.black` kullanımı:

```black
form LoginForm {
  fields email, password
  ui black 1 solid 8 8 5 5 6 center
}
```

Varsayılan sıra:

```text
ui <color> <width> <style> <pt> <pr> <pb> <pl> <radius> <place>
```

Generator bu sırayı soldan sağa okur:

```text
color  = black
width  = 1
style  = solid
pt     = 8
pr     = 8
pb     = 5
pl     = 5
radius = 6
place  = center
```

Proje bazlı özelleştirme için ileride `blackdir` içinde sıralama profili tanımlanabilir:

```blackdir
ui = color width style pt pr pb pl radius place;
```

Profil satırında `=` ile `;` arasındaki tokenlar soldan sağa yorumlanır.

### UI Profile Uyumluluk Kuralları

Pozisyonel UI syntax geriye dönük uyumluluk için sıkı kurallarla yönetilmelidir:

1. UI slot sırası proje başında belirlenir.
2. İlk gerçek kullanımdan sonra profil kilitlenir.
3. Var olan slot taşınamaz veya anlamı değiştirilemez.
4. Yeni slot sadece sona eklenebilir.
5. Sondaki eksik değerler default kullanır.
6. Araya ekleme veya yeniden sıralama gerekiyorsa `black theme migrate` bunu unsafe olarak raporlar; otomatik source rewrite ayrı bir gelecek refactor aşamasıdır.
7. IDE ve AI aynı profile metadata'sını kullanır.

Temel kural:

```text
Existing positional slots are immutable. New slots are append-only.
```

Örnek güvenli genişleme:

```blackdir
ui.version = 1
ui = color width style pt pr pb pl radius place;
```

Sonraki sürüm:

```blackdir
ui.version = 2
ui = color width style pt pr pb pl radius place shadow;
```

Eski kullanım bozulmaz:

```black
ui black 1 solid 8 8 5 5 6 center
```

Güncel CLI kontrolü:

```bash
black theme migrate theme-v1.blackthm theme-v2.blackthm --json
```

`success` ve `safe` birlikte `true` değilse mevcut `.black` kaynak theme değişikliğinden önce düzeltilmelidir.

Çünkü `shadow` verilmemiştir ve generator default değer kullanır.

Yasak genişleme:

```blackdir
ui = color width shadow style pt pr pb pl radius place;
```

Bu yasaktır çünkü eski `ui` satırlarının slot anlamlarını kaydırır.

IDE desteği profile metadata'sını okuyarak sıradaki slotu gösterebilir:

```text
ui [color] [width] [style] [pt] [pr] [pb] [pl] [radius] [place]
```

Kullanıcı `ui black` yazdığında IDE sıradaki alanın `width` olduğunu gösterebilir. AI ajanı da `black inspect --json` veya `blackdir` üzerinden aynı sırayı bir kez okuyup proje boyunca kullanabilir.

Bu yaklaşımın amacı:

- AI'nin component stilini tek satırda yazması
- Generator'ın CSS class/id çıktısını deterministik üretmesi
- Büyük CSS blokları yerine kısa, ezberlenebilir UI sırası kullanılması
- İnsanların sıralamayı dokümantasyon sitesinden veya `blackdir` profilinden kopyalayıp kullanabilmesi
- İleride tüm CSS özelliklerinin değil, güvenli ve desteklenen UI özelliklerinin bu profile kontrollü şekilde eklenmesi

### Mevcut v0.2 Uygulaması: `.blackthm` Lock Baseline

v0.2 içinde bu kararın ilk çalışan karşılığı `.blackthm` dosyasıdır. Profil kilitliyse current `mode` satırlarının eski sırayı bozmadığını anlamak için `baseline` satırları kullanılır.
Web UI profilleri ayrıca standart `box`, `text`, `table` ve `button` mode gruplarını içermelidir. Böylece compiler, IDE ve AI ajanı bir UI satırının container, typography, table veya action control alanına ait olduğunu tahmin etmeden anlayabilir.

```blackthm
blackthm WarehouseTheme {
  version 2
  locked true

  profile UICompact {
    version 2
    baseline box color width style pt pr pb pl radius place
    baseline text color size weight align
    baseline table color width style density zebra
    baseline button bg color radius size variant

    ui box = color width style pt pr pb pl radius place shadow;
    ui text = color size weight align;
    ui table = color width style density zebra;
    ui button = bg color radius size variant;
  }
}
```

Compiler kuralı:

```text
baseline slotları current mode satırının birebir başlangıcı olmalıdır.
Yeni slotlar sadece baseline sonrasına eklenebilir.
```

Bu yüzden aşağıdaki kullanım hatalıdır:

```blackthm
ui box = color width shadow style pt pr pb pl radius place;
```

Çünkü `shadow` araya eklenmiştir ve eski UI satırlarının anlamını kaydırabilir. Compiler bu durumda `NON_APPEND_ONLY_UI_SLOT` hatası verir.

`ui <mode> = <slot...>;` satırı generator'ın o mode için değerleri hangi sırayla okuyacağını belirler. `black theme inspect --json` çıktısındaki `profile.modeGroups`, bu standart grupların ne işe yaradığını, hangi elementlere uygulanacağını ve default slot sırasını gösterir. Standart gruplardan biri eksikse compiler `MISSING_STANDARD_UI_MODE` hatası verir.

Theme değiştirirken `black theme migrate old.blackthm new.blackthm --json` kullanılmalıdır. Komut read-only çalışır; eski ve yeni profilin aynı theme/profile kimliğiyle devam ettiğini, sürümlerin geriye gitmediğini, eski mode'ların silinmediğini ve eski slot listesinin yeni slot listesinin exact prefix'i olduğunu kontrol eder. Araya slot ekleme, reorder veya silme `UI_SLOT_MIGRATION_BREAK` döndürür.

### Mevcut v0.2 Uygulaması: Inline UI Intent

`.black` içinde field, form, table ve action button yanına kompakt `ui` niyeti yazılabilir. Bu, CSS yazmadan görsel niyeti kaynak dosyada ilgili öğenin yanında tutar.

```black
entity Product {
  name text required ui text "#172026" 14 semibold left
}

page Products {
  source Product

  table {
    id ProductsTable
    class inventoryTable
    columns name
    ui table border 1 solid compact true
  }

  form {
    id ProductForm
    class inventoryForm
    fields name
    ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
  }

  actions create
  action create id CreateProductButton
  action create class primaryAction
  action create ui button primary white 6 md solid
}
```

Compiler bu aşamada UI intent bilgisini parse eder, validate eder, JSON/BlackIR çıktısına taşır ve web target için stable `.bl-ui-*` class kurallarıyla CSS üretir. `black build`, `blacklang.toml` içinde `theme = ...` varsa generator okuma sırasını `.blackthm` profilinden alır.

Üretilen class düzeni:

```text
table   .bl-ui-table-<page>
form    .bl-ui-form-<page>
field   .bl-ui-field-<entity>-<field>
action  .bl-ui-action-<page>-<action>
```

Mevcut v0.2 uygulamasında table/form/action için explicit UI identity de desteklenir:

```black
table {
  id ProductsTable
  class inventoryTable
}

form {
  id ProductForm
  class inventoryForm
}

action create id CreateProductButton
action create class primaryAction
```

Generator `id` ve `class` değerlerini kebab-case HTML çıktısına çevirir. Row action butonlarında aynı `id` tekrar etmesin diye aksiyon id'leri kullanım yerine göre `-open`, `-submit`, `-bulk` veya `-item-<recordId>` suffix'i alır.

Geçerli bağlam kuralı:

```text
field   box, text
form    box, text, button
table   box, text, table
button  button
```

## Aşama 26.1: Page View Order

Bu aşama, bir sayfa içindeki generated parçaların sırasını `.black` kaynak dosyasından değiştirmeyi sağlar.

Amaç; formu, tabloyu veya detail panelini taşımak için generated React/CSS dosyalarına elle dokunmadan, sayfa niyetini BlackLang içinde tutmaktır.

### Mevcut Durum

BlackLang artık page içinde `view` bloğunu okuyabilir:

```black
page Products {
  source Product

  view {
    order form, table, detail
  }

  table {
    columns sku, name, stock, price
  }

  form {
    fields sku, name, stock, price
  }
}
```

Bu örnekte form önce, table sonra, detail paneli en sonda görünür.

İlk kapsam:

- `view` bloğu page içinde kullanılır.
- `order` satırı `table`, `detail`, `form` section adlarını kabul eder.
- Yazılan section'lar önce gelir.
- Eksik bırakılan destekli section'lar varsayılan `table, detail, form` sırasıyla sona eklenir.
- `compose grid`, `compose stack` ve `compose tabs` desteklenir.
- `compose grid ... stackAt sm|md|lg|none` canonical responsive syntax'tır; generator deterministic lg/md/sm media query ladder'ı üretir.
- Responsive kolon sayısından büyük section span'leri breakpoint içinde clamp edilir.
- `section detail display drawer side right title "Order Details"` gibi satırlar generated detail/form panellerini overlay olarak açabilir.
- `section form display modal title "Order Form"` create/edit formunu modal panel içinde açtırır.
- `group CustomerWorkspace sections detail, form compose stack gap md span 1 title "Customer Workspace"` gibi satırlar contiguous inline section'ları tek generated wrapper altında toplar.
- Group wrapper kendi `compose stack|grid`, `columns`, `gap`, `span` ve `title` niyetini taşıyabilir.
- `tab <Name> sections table, detail` satırı tabs modunda generated tab control üretir.
- Tabs modunda her ordered section tam bir tab içinde yer almalıdır; modal/drawer display ve group tabs ile bu MVP'de birleşmez.
- `trigger detail on rowSelect`, `trigger form on createStart|editStart`, `trigger detail|table on saveSuccess` ve `trigger table on close` generated tab/overlay akışını deterministik hale getirir.
- `section StockSummary component StockBadge bind selected span 1 title "Stock Summary"` declared component'i selected row'a bağlı reusable inline page section olarak üretir.
- `bind first` ilk yüklenen listedeki ilk kaydı component'e geçirir.
- Component section input'ları source entity'deki stored veya computed field'larla isim ve tip olarak eşleşmelidir.
- Component section'lar bu MVP'de inline render edilir; modal/drawer component section ilerideki daha zengin composition modeline bırakılmıştır.
- Duplicate veya desteklenmeyen section adları validator hatası üretir.
- Generator table/detail/form için stable `.bl-view-section-*`, `.bl-view-display-*`, `.bl-view-side-*`, `.view-group`, `.bl-view-group-*` ve `.bl-view-has-triggers` class'ları üretir.
- `view` intent varsa `src/styles.css` içinde deterministik order/composition/tab/overlay/group kuralları üretilir; trigger davranışı generated React state akışına bağlanır.
- `black audit accessibility --json`, modal/drawer section title'ı ve çoklu section group title'ı için read-only policy diagnostics üretir.

Sonraki genişleme:

- coordinate/fixed placement gibi kontrollü pozisyonlama

## Aşama 27: Internationalization

Çok dilli uygulamalar için metinler kaynak koddan ayrılmalıdır.

### Eklenecekler

- Locale
- Translation key
- Date format
- Currency format
- Number format
- RTL support

### Örnek

```black
i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}
```

### Mevcut Durum

Bu aşamanın genişletilmiş web kapsamı eklendi. BlackLang artık top-level `i18n` bloğunu, entity field hedefli `label Entity.field { ... }` translation bloklarını, generated UI copy hedefli `label app/page/action/table/status.* { ... }` bloklarını ve stored field hedefli `placeholder/help/message Entity.field { ... }` bloklarını okuyabilir:

```black
i18n {
  default tr
  locales tr, en
}

label Product.stock {
  tr "Stok Adedi"
  en "Stock Count"
}

label app.title {
  tr "Depo"
  en "Warehouse"
}

label page.Products {
  tr "Ürünler"
  en "Products"
}

label action.create {
  tr "Oluştur"
  en "Create"
}

placeholder Product.stock {
  tr "Stok adedini gir"
  en "Enter stock count"
}

help Product.stock {
  tr "Satılabilir mevcut ürün adedi"
  en "Current available quantity"
}

message Product.stock {
  tr "Geçerli bir stok adedi gir"
  en "Enter a valid stock count"
}
```

Compiler bu bilgiyi parse eder, validate eder, JSON/BlackIR çıktısına taşır ve web generator çok locale varsa runtime language selector üretir. Generated App aktif locale'i sayfalara geçirir ve `lang`/`dir` attribute'larını üretir. App chrome, navigation page name, table tool, status label, CRUD/custom action button, workflow transition button, table, detail, form, filter ve column visibility field label'ları; form placeholder/help text'leri ve field-level frontend validation message'ları kullanıcı dil değiştirdiğinde yeniden render olur. Table/detail number, integer, decimal, money, date ve datetime değerleri aktif locale ile `Intl` üzerinden formatlanır. Translation yoksa önce default locale çevirisi, sonra ilgili inline field modifier'ı veya deterministic fallback kullanılır.

İlk kapsam:

- `i18n.default`
- `i18n.locales`
- `label Entity.field`
- `label Entity.computedField`
- `label app.*`
- `label page.<Page>`
- `label action.<key|CustomAction|workflowTransition>`
- `label action.new/create/edit/view.<Entity>`
- `label table.*`
- `label status.*`
- `placeholder Entity.field`
- `help Entity.field`
- `message Entity.field`
- runtime field label switching
- runtime app chrome/action/table/status copy switching
- runtime field placeholder/help/message switching
- locale-aware value formatting
- basic RTL direction support
- locale ve target validation
- JSON/BlackIR visibility

## Aşama 28: Security Katmanı

Security sonradan eklenen bir süs değil, dil seviyesinde temsil edilen bir alan olmalıdır.

### BlackLang Source Security Kararı

BlackLang source dosyaları yüksek değerli kaynak varlık olarak kabul edilmelidir. Çünkü olgun bir projede birkaç bin satırlık `.black` dosyası, çok daha büyük bir web uygulamasının ana niyetini ve üretim reçetesini taşıyabilir.

Temel kural:

```text
.black source = protected source of truth
generated code = yeniden üretilebilir çıktı
production server = mümkünse sadece production artifact
```

Bu yüzden secret, password, API key, token, private key ve gerçek bağlantı bilgileri doğrudan `.black` dosyasına yazılmamalıdır.

Doğru hedef kullanım:

```black
database {
  url env DATABASE_URL
}

security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Kaçınılması gereken kullanım:

```black
database {
  url "postgres://user:password@example.com/app"
}
```

Eklenen BlackLang'e özgü korumalar:

- `database { url env DATABASE_URL }` parse/validate desteği
- `security { cors { origins env CORS_ORIGINS } }` parse/validate desteği
- JSON/BlackIR içinde database env referansının görünmesi
- JSON/BlackIR içinde security/cors env referansının görünmesi
- generated Express server içinde CORS middleware üretimi
- `black security scan --json`
- hardcoded secret tespiti
- production package üretirken `.black` ve `.black.enc` kaynaklarını dışarıda bırakma
- encrypted source mode: `app.black.enc`
- `black security encrypt <file>`
- `black security decrypt <file.black.enc> --stdout`
- `.black.enc` source'u parse/lint/validate/inspect/benchmark/security scan/build sırasında bellekte okuma

Kalan ileri seviye korumalar:

- `black build --secure`
- CI/CD secret ayrımı

Bu kararın amacı BlackLang'i korkarak kısıtlamak değil; tam tersine, kaynak dosyanın değeri arttıkça onu profesyonel source repository gibi korumaktır.

### Mevcut Durum

Bu aşamanın ilk çalışan parçaları eklendi. BlackLang artık top-level `database` ve `security` bloklarını okuyabilir:

```black
database {
  url env DATABASE_URL
}

security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Literal database URL yazımı parser tarafından reddedilir. Böylece connection string, password veya token gibi değerlerin `.black` source içine gömülmesi yerine environment üzerinden referans verilmesi temel kural haline gelir.

`security.cors`, tarayıcıdan API'ye hangi origin'lerin erişebileceğini `.black` içinde niyet olarak tanımlar ama gerçek domain listesini environment içinde bırakır. Generated Express server `CORS_ORIGINS` değerini virgülle ayrılmış liste olarak okur, listede olmayan browser origin'lerini reddeder, `OPTIONS` preflight yanıtı üretir ve `credentials true` ile cookie-auth uyumlu header'ları ekler.

CLI tarafında source security için şu komutlar da vardır:

```bash
black security scan --json
black security encrypted-source --json
black security encrypt app.black --out app.black.enc --json
black security decrypt app.black.enc --stdout
black package --production
```

`black security scan --json`, `.black` source içinde olası hardcoded database URL, private key, API key, token, secret ve password değerlerini raporlar.

`black security encrypted-source --json`, `.black.enc` protected source modunun durumunu ve üretim kurallarını AI/CI araçlarının okuyabileceği şekilde raporlar. `black security encrypt`, key değerini yalnızca environment'tan alarak AES-GCM encrypted source üretir. `black security decrypt --stdout`, plaintext'i dosyaya yazmadan açıkça stdout'a basar.

`parse`, `lint`, `validate`, `inspect`, `benchmark`, `security scan` ve `build` komutları key environment variable set edildiğinde `.black.enc` source'u bellekte okuyabilir. `black format`, `.black.enc` dosyasını rewrite etmez; plaintext source güvenilir workspace'te formatlanıp tekrar encrypt edilmelidir.

Generated web/API output artık `security/secrets.json`, `scripts/secrets-plan.mjs` ve `scripts/secrets-provider.mjs` üretir. Manifest `DATABASE_URL`, `CORS_ORIGINS`, cloud app/region env değerleri ve observability endpoint gibi referansları source/kind/required/sensitive metadata ile listeler ama değer saklamaz. `npm run security:secrets:plan`, environment readiness bilgisini JSON olarak raporlar. `npm run security:secrets:preflight`, environment veya secret manager handoff öncesi provider label, prefix ve CLI availability bilgisini JSON olarak raporlar; secret değerlerini fetch etmez veya yazdırmaz.

Release trust için repository artık `scripts/verify-release-trust.mjs` içerir. `black ecosystem --json` içinde `release.trust` alanı Ed25519 algoritmasını, `<artifact>.sig` pattern'ini, `BLACKLANG_RELEASE_PUBLIC_KEY` / `BLACKLANG_RELEASE_PUBLIC_KEY_FILE` public key referanslarını, strict verify command'ini, `transparencyLog` metadata'sını ve `keyRotation` policy metadata'sını gösterir. `packages/registry/release-transparency.blackdir`, append-only `transparency.blackdir` entry shape ve hash-chain alanlarını tanımlar. `packages/registry/key-rotation-policy.blackdir`, sha256-public-key-spki-prefix key id formatını, 30 günlük overlap kuralını ve `key-revocations.blackdir` revocation manifestini tanımlar. Npm wrapper download akışı checksum doğrulamasından sonra detached signature doğrulaması yapmadan binary extract etmez. Private signing key `.black`, manifest veya package metadata içinde tutulmaz.

`black package --production`, generated output'tan production artifact üretir ve `.black`, `.black.enc`, `.env`, local database, `node_modules` ve generated Prisma client output gibi taşınmaması gereken dosyaları pakete dahil etmez.

### Eklenecekler

- Auth enforcement
- Permission enforcement
- Input sanitization
- CSRF
- CORS
- Rate limit
- Secret management
- Secure headers
- SQL injection prevention
- XSS prevention

### Örnek

```black
security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

## Aşama 29: Deployment ve Environment

BlackLang sadece kod üretmekle kalmayıp uygulamanın nasıl çalıştırılacağını da tarif edebilir.

### Eklenecekler

- Environment variables
- Dockerfile
- Docker Compose
- Ops healthcheck
- Build script
- Start script
- Production config
- Preview deployment
- Database migration command

### Örnek

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}
```

### Mevcut Durum

Bu aşamanın ilk çalışan parçası eklendi. BlackLang artık top-level `deploy` bloğunu okuyabilir:

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
}
```

Bu blok şunları üretir:

- `Dockerfile`
- `.dockerignore`
- `docker-compose.yml`
- `docker-compose.preview.yml`
- `.env.example` içinde `PORT`
- `.env.example` içinde `BLACKLANG_PREVIEW_PORT`
- generated `package.json` içinde `start`
- generated `package.json` içinde `deploy:preview`, `deploy:preview:down` ve `deploy:rollback:plan`
- `deploy/manifest.json`
- `deploy/rollback.json`
- `deploy/cloud.json`
- `scripts/rollback-plan.mjs`
- `scripts/cloud-plan.mjs`
- `scripts/cloud-exec.mjs`
- generated `src/server.ts` içinde `process.env["PORT"]`
- generated Express server üzerinden `dist` frontend servis etme
- `ops.health` varsa Dockerfile ve docker-compose içinde app healthcheck

Şimdilik deploy target olarak yalnızca `docker` desteklenir. `preview local`, normal compose stack'ten ayrı host port ve preview database default'ları üretir. `rollback keep N`, altyapıyı değiştirmeyen read-only rollback plan metadata'sı üretir. `cloud fly|render|railway app env NAME [region env NAME]`, `deploy/cloud.json`, `deploy/manifest.json` cloud metadata'sı, read-only `deploy:cloud:plan`/`deploy:cloud:preflight` çıktısı ve açık apply gerektiren `deploy:cloud:exec` provider CLI runner'ı üretir. `database sqlite` için Docker Compose local SQLite verisini `/app/data` altında tutar; `database postgres` için PostgreSQL service, healthcheck ve app `DATABASE_URL` fallback'i üretir; `database mysql` için MySQL 8.4 service, healthcheck, `MYSQL_*` local env default'ları ve app `DATABASE_URL` fallback'i üretir. Runtime probe ve observability hook syntax'ı ayrı `ops` bloğudur.

## Aşama 30: Plugin ve Target Sistemi

BlackLang büyüdükçe farklı frontend/backend hedefleri gerekebilir.

### Eklenecekler

- Web target
- API-only target
- Admin target
- Mobile target hazırlığı
- Desktop target hazırlığı
- Generator plugin sistemi
- Adapter sistemi

### Örnek

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

### Mevcut Durum

Bu aşamanın web ve API-only çalışan parçaları eklendi. BlackLang artık top-level `target` bloğunu iki generated uygulama hedefiyle okuyabilir:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

```black
target api {
  backend node
  database sqlite
}
```

Şimdilik bilinçli olarak `web + react + node + sqlite|postgres|mysql` ve `api + node + sqlite|postgres|mysql` desteklenir. Bunun sebebi generator'ın gerçekten üretebildiği çalışan stack'leri syntax olarak açmasıdır. Admin, mobile, desktop veya farklı backend targetları syntax olarak açılmadan önce generator adapter desteği eklenmelidir.

`target api`, React/Vite/frontend smoke/browser-check/e2e/matrix dosyalarını üretmez. Page deklarasyonları yine entity kaynaklı API route/client/validation/OpenAPI yüzeyini tarif eder; browser `test` deklarasyonları ise yalnızca `target web` içinde desteklenir.

Bu aşama sayesinde:

- `.black` kaynağı hangi stack için yazıldığını açıkça söyler.
- JSON ve BlackIR çıktıları target bilgisini taşır.
- Validator desteklenmeyen targetları erken yakalar.
- Generated README üretilen stack'i raporlar.
- `black inspect --affected target --json` target değişirse kontrol edilecek dosyaları gösterir.
- `black docs target --json` ve `black explain target --json` AI ajanına kısa öğrenme paketi verir.

## Aşama 31: AI Agent Tooling

Bu aşama BlackLang'i gerçekten AI-native yapan katmandır.

### Eklenecek Komutlar

```bash
black inspect --json
black validate --json
black diff --json
black plan --json
black explain Product --json
black inspect app.black --affected Product.stock --json
```

### Amaç

AI bir değişikliğin etkisini tahmin etmek yerine compiler'dan öğrenir.

### Örnek

```json
{
  "change": "Product.stock",
  "affected": {
    "entities": ["Product"],
    "pages": ["Products", "WarehouseDashboard"],
    "api": ["ProductApi"],
    "tests": ["ProductCrud", "LowStockReport"]
  }
}
```

### Mevcut v0.2 Uygulaması: AI Agent Contract

Bu aşama, dış AI ajanlarının BlackLang'i yanlış yorumlamasını engellemek için resmi çalışma sınırını açık hale getirir.

Özellikle Cursor testinde görülen risk şudur:

```text
AI ajanı siteyi okur
BlackLang'in henüz yapamadığı bir işi ister
kendi uydurduğu runtime syntax ile çözüm üretir
bunu resmi BlackLang gibi gösterebilir
```

Bu yüzden repo ve dokümantasyon sitesine şu kural eklenmiştir:

```text
official BlackLang path:
.black source -> black CLI -> generated web/API output
```

Desteklenmeyen kullanım:

```html
<script type="text/black">
```

Bu kullanım, resmi browser runtime olmadığı sürece BlackLang syntax'ı sayılmaz.

Yeni referans dosyası:

```text
docs/ai-agent-contract.md
```

CLI kısa öğrenme komutu:

```bash
black docs agent-contract --json
```

Bu sözleşme şunu netleştirir:

- AI ajanı `.black` ve `.blackthm` kaynaklarını değiştirmelidir.
- Generated çıktıyı normal çözüm olarak elle değiştirmemelidir.
- Desteklenmeyen syntax uydurmamalıdır.
- Calculator/game/custom frontend event gibi işler bugün resmi BlackLang kapsamı değildir.
- Böyle bir istek gelirse ya normal web prototipi açıkça prototip olarak etiketlenmeli ya da önce compiler'a gerekli BlackLang özelliği eklenmelidir.

## Aşama 32: Refactor ve Kod Sağlığı

## Aşama 31.1: BlackIR Ara Temsil Formatı

JSON dış araçlar için güçlü ve standarttır, ama büyük projelerde çok satır ve token tüketebilir.

BlackLang bu yüzden kendi kısa ara temsil formatına sahip olmalıdır:

```text
.blackir
```

Roller:

```text
.black    → insan ve AI tarafından yazılan kaynak dil
.blackir  → BlackLang'in kısa, AI-readable ara temsili
.json     → dış araçlar ve entegrasyonlar için standart çıktı
```

Örnek:

```blackir
blackir 0.1

app Warehouse

entity Product
  sku text required unique
  name text required
  stock number default 0
  price money

page Products source Product
  table sku name stock price
  search sku name
  form sku name stock price
  actions create edit delete
```

Bu formatın amacı:

- JSON'a göre daha az satır kullanmak
- AI'nin daha az tokenla proje özetini anlamasını sağlamak
- İç compiler/generator akışında BlackLang'e ait bir temsil oluşturmak
- Yine de JSON desteğini koruyarak dış araçlara açık kalmak

Planlanan komutlar:

```bash
black parse app.black --ir
black validate app.black --ir
black inspect --ir
black build app.black --ir
```

## Aşama 32: Refactor ve Kod Sağlığı

AI ajanları için güvenli refactor komutları çok değerlidir.

### Eklenecekler

- Rename entity
- Rename field
- Move page
- Extract component
- Split module
- Detect unused entity
- Detect unused page
- Detect broken relation

### Örnek

```bash
black refactor rename-field Product.sku barcode
```

## Aşama 33: Benchmark ve Ölçüm

Bu aşamada projenin iddiası ölçülebilir hale gelir.

### Ölçülecekler

- BlackLang satır sayısı
- Üretilen kod satır sayısı
- Normal stack karşılığı satır sayısı
- AI input token miktarı
- AI output token miktarı
- Değişen dosya sayısı
- Hata sayısı
- Build süresi
- Test başarısı

### Karşılaştırma Tablosu

```text
Metric                 Normal Stack   BlackLang
Source lines                    TBD         TBD
Generated lines                 TBD         TBD
Files edited                    TBD         TBD
Input tokens                    TBD         TBD
Output tokens                   TBD         TBD
Validation errors               TBD         TBD
Build time                      TBD         TBD
```

## Aşama 34: Documentation ve Learning Pack

Yeni bir dili AI'nin hızlı öğrenmesi için dokümantasyon özel hazırlanmalıdır.

### Eklenecekler

- `SPEC.md`
- `BLACKLANG.md`
- `AGENTS.md`
- Quick examples
- Anti-pattern examples
- Error code reference
- Generated file policy
- Migration guide
- Best practices

### AI İçin Kural

```md
Before changing a BlackLang project:

1. Read `BLACKLANG.md`.
2. Run `black inspect --json`.
3. Change only `.black` source files.
4. Run `black validate --json`.
5. Run `black build`.
6. Do not manually edit generated files.
```

## Aşama 35: Gerçek Uygulama Şablonları

Bu aşamada BlackLang sadece oyuncak örnek değil, gerçek uygulama türleri üretebilir.

### Şablonlar

- CRM
- Warehouse management
- Inventory management
- Order management
- Admin dashboard
- Helpdesk
- Appointment system
- Invoice system
- Project management panel
- B2B customer portal

### Amaç

AI ajanı sıfırdan her şeyi kurmak yerine bilinen uygulama desenlerini BlackLang üzerinden hızlı üretebilir.

## Önceliklendirilmiş İlk 10 Adım

1. `SPEC.md` oluştur.
2. İlk syntax kararlarını yaz.
3. `examples/warehouse.black` oluştur.
4. Parser prototipi yaz.
5. JSON AST çıktısı üret.
6. Validator prototipi yaz.
7. `black validate --json` komutunu ekle.
8. React + Node + Prisma generator yaz.
9. İlk çalışan Products CRUD uygulamasını üret.
10. Normal stack karşılığıyla satır/token karşılaştırması yap.

## MVP Kapsamı

İlk MVP sadece şunları içermelidir:

- `app`
- `entity`
- `page`
- `table`
- `form`
- `actions`
- `search`
- Parser
- Validator
- Web generator
- JSON hata çıktısı

MVP'nin hedefi şudur:

> Tek bir `.black` dosyasından çalışan bir CRUD web uygulaması üretmek.

## Uzun Vadeli Hedef

BlackLang web tarafında olgunlaştıktan sonra aynı core temsil korunarak başka platformlara genişleyebilir.

```text
entity Product
page Products
workflow OrderPreparation
permission WarehouseWorker
```

Bu tanımlar ileride farklı targetlara çevrilebilir:

```text
Web        → React / Node
Mobile     → React Native
Desktop    → Tauri / Electron
API-only   → Fastify / Express
Automation → Background workers
```

## Sonuç

BlackLang'in web geliştirme yol haritası bir anda her şeyi yapmaya çalışmamalıdır. Bunun yerine önce AI'nin en çok zorlandığı ve en çok tekrar ettiği alanlardan başlanmalıdır:

- Entity
- CRUD
- Form
- Table
- Validation
- API
- Permission
- Workflow

Bu temel oturduktan sonra auth, relation, dashboard, notification, testing, deployment ve plugin sistemi eklenebilir.

Projenin ana iddiası her aşamada korunmalıdır:

> BlackLang, web uygulamalarını insanlar için daha kısa yazdıran bir dil değil; AI ajanlarının uygulamayı daha az enerjiyle okuyup anlayacağı, güvenli değiştireceği ve doğrulayacağı bir kaynak temsilidir.

## v0.2 Geçiş Notu

v0.1 roadmap tamamlandıktan sonra sonraki çalışma alanı `ROADMAP-v0.2.md` dosyasında takip edilir.

v0.2'nin amacı MVP'yi büyütmek değil, onu daha kullanılabilir hale getirmektir:

- dağıtılabilir CLI
- AI ve insan için daha iyi edit/inspect komutları
- BlackLang-native UI/theme dili
- gerçek app şablonları
- benchmark ve test sistemi
- protected/encrypted source tooling
- Python/W3Schools tarzı dokümantasyon sitesi
