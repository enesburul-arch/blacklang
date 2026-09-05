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
- SQLite veya PostgreSQL
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
- Form select input ve relation display davranışları sonraki relation parçalarına bırakılır.

### Bu Aşamada Tamamlanan İkinci Parça

- Relation field kullanılan form alanları artık metin input yerine select input olarak üretilir.
- Generated React sayfası relation hedefindeki kayıtları API client ile yükler.
- Form kaynağı BlackLang'de `customer` olarak okunabilir kalır; API tarafına gereken `customerId` payload'u generator üretir.
- Table ve detail panel relation alanlarında mümkün olduğunda ilişkili kaydın okunabilir etiketi gösterilir.
- Birden fazla `page` tanımı olduğunda generated React app sayfalar arası navigation üretir.

### Bu Aşamada Tamamlanan Üçüncü Parça

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

Draft v0.1 ayrıca explicit API contract bloklarını da okuyabilir:

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
  webhook
  public
}
```

Bu bloklar şu anda contract-first çalışır:

- Parse JSON çıktısında görünür.
- BlackIR çıktısında görünür.
- `black inspect` çıktısında özetlenir.
- `generated/openapi.json` içine path, query param, path param, public/private metadata ve webhook metadata olarak yazılır.
- Runtime Express route üretimi sonraki API aşamasına bırakılır.

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

- `BlackUser` tablosu tek bir `role` değeri saklar.
- Yeni kayıt olan kullanıcıya ilk tanımlı rol atanır.
- `/api/auth/me` kullanıcı rolünü döndürür.
- `access` tanımlı sayfaların API route'ları role göre korunur.
- Yetkisiz role sahip kullanıcı `403 Forbidden` cevabı alır.
- Roller varsa generated uygulama temel bir Users ekranı üretir.
- İlk tanımlı rol kullanıcıları listeleyebilir ve rollerini değiştirebilir.
- Permission action tarafı çalışır:
  - `read` list/detail endpointlerini korur.
  - `create` create endpointini korur.
  - `update` edit/archive/restore endpointlerini korur.
  - `delete` single ve bulk delete endpointlerini korur.
- Generated React sayfaları yetkisiz create/edit/archive/restore/delete kontrollerini gizler.
- `deny` kuralı eşleşen `allow` kuralını ezer.
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

Henüz yapılmayan kısım gelişmiş yetki yönetimidir. Bir kullanıcıya birden fazla rol verme, ownership rule ve tenant rule sonraki güvenlik adımlarında eklenecektir.

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

Component declaration artık standalone React component dosyasına dönüşür. Variant koşulları `.black` içinde deterministik niyet olarak korunur; `stock < 10` gibi basit `input operator literal` koşulları runtime class seçimine çevrilir. Tek input'u entity field adı ve tipiyle eşleşen component'ler generated table/detail rendering alanlarına otomatik bağlanır. Aynı eşleşme generated form alanlarında canlı component önizlemesi olarak da kullanılır.

### Eklenecekler

- Component
- Props
- Slots
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

## Aşama 17: Query ve Data Fetching

Gerçek uygulamalarda basit CRUD dışında özel sorgular gerekir.

### Eklenecekler

- Query
- Filter
- Sort
- Aggregate
- Count
- Sum
- Group by
- Include relation
- Computed field

### Örnek

```black
query LowStockProducts {
  from Product
  where stock < 10
  sort stock asc
}
```

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
observability {
  logs structured
  errors json
  healthcheck "/health"
  audit Product create, update, delete
}
```

## Aşama 24: Test Sistemi

BlackLang'in en değerli taraflarından biri testleri de kaynaktan üretebilmesi olabilir.

### Eklenecekler

- Unit test
- API test
- Component test
- E2E test
- Fixture
- Seed data
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

### Örnek

```black
migration RenameSkuToBarcode {
  rename Product.sku to barcode
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
6. Araya ekleme veya yeniden sıralama gerekiyorsa compiler migration yapar.
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

    mode box color width style pt pr pb pl radius place shadow
    mode text color size weight align
    mode table color width style density zebra
    mode button bg color radius size variant
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
mode box color width shadow style pt pr pb pl radius place
```

Çünkü `shadow` araya eklenmiştir ve eski UI satırlarının anlamını kaydırabilir. Compiler bu durumda `NON_APPEND_ONLY_UI_SLOT` hatası verir.

`black theme inspect --json` çıktısındaki `profile.modeGroups`, bu standart grupların ne işe yaradığını, hangi elementlere uygulanacağını ve default slot sırasını gösterir. Standart gruplardan biri eksikse compiler `MISSING_STANDARD_UI_MODE` hatası verir.

### Mevcut v0.2 Uygulaması: Inline UI Intent

`.black` içinde field, form, table ve action button yanına kompakt `ui` niyeti yazılabilir. Bu, CSS yazmadan görsel niyeti kaynak dosyada ilgili öğenin yanında tutar.

```black
entity Product {
  name text required ui text "#172026" 14 semibold left
}

page Products {
  source Product

  table {
    columns name
    ui table border 1 solid compact true
  }

  form {
    fields name
    ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
  }

  actions create
  action create ui button primary white 6 md solid
}
```

Compiler bu aşamada UI intent bilgisini parse eder, validate eder, JSON/BlackIR çıktısına taşır ve web target için stable `.bl-ui-*` class kurallarıyla CSS üretir.

Üretilen class düzeni:

```text
table   .bl-ui-table-<page>
form    .bl-ui-form-<page>
field   .bl-ui-field-<entity>-<field>
action  .bl-ui-action-<page>-<action>
```

Geçerli bağlam kuralı:

```text
field   box, text
form    box, text, button
table   box, text, table
button  button
```

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
```

Kaçınılması gereken kullanım:

```black
database {
  url "postgres://user:password@example.com/app"
}
```

İleride eklenecek BlackLang'e özgü korumalar:

- `database { url env DATABASE_URL }` parse/validate desteği
- JSON/BlackIR içinde database env referansının görünmesi
- `black security scan --json`
- hardcoded secret tespiti
- production package üretirken `.black` kaynaklarını dışarıda bırakma
- signed compiler/release kontrolü
- encrypted source mode: `app.black.enc`
- `black build --secure`
- CI/CD secret ayrımı

Bu kararın amacı BlackLang'i korkarak kısıtlamak değil; tam tersine, kaynak dosyanın değeri arttıkça onu profesyonel source repository gibi korumaktır.

### Mevcut Durum

Bu aşamanın ilk çalışan parçaları eklendi. BlackLang artık top-level `database` bloğunu okuyabilir:

```black
database {
  url env DATABASE_URL
}
```

Literal database URL yazımı parser tarafından reddedilir. Böylece connection string, password veya token gibi değerlerin `.black` source içine gömülmesi yerine environment üzerinden referans verilmesi temel kural haline gelir.

CLI tarafında source security için şu komutlar da vardır:

```bash
black security scan --json
black security encrypted-source --json
black package --production
```

`black security scan --json`, `.black` source içinde olası hardcoded database URL, private key, API key, token, secret ve password değerlerini raporlar.

`black security encrypted-source --json`, `.black.enc` protected source modunun durumunu ve üretim kurallarını AI/CI araçlarının okuyabileceği şekilde raporlar. Draft v0.1 içinde bu mod planlıdır; production package `.black` ve `.black.enc` kaynaklarını dışarıda bırakır.

`black package --production`, generated output'tan production artifact üretir ve `.black` source, `.env`, local database, `node_modules` ve generated Prisma client output gibi taşınmaması gereken dosyaları pakete dahil etmez.

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
  csrf true
  cors sameOrigin
  rateLimit api 100 per minute
  secrets env
}
```

## Aşama 29: Deployment ve Environment

BlackLang sadece kod üretmekle kalmayıp uygulamanın nasıl çalıştırılacağını da tarif edebilir.

### Eklenecekler

- Environment variables
- Dockerfile
- Docker Compose
- Build script
- Start script
- Production config
- Preview deployment
- Database migration command

### Örnek

```black
deploy {
  target docker
  database postgres
  env DATABASE_URL required
}
```

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
  database postgres
}
```

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
