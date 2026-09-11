# BlackLang Web Completion Worklog

## Phase 21 / Custom Actions MVP

Bu aşama CRUD dışındaki güvenli domain işlemlerini mümkün kılıyor: `.black` içinde bir entity'ye bağlı row-level action tanımlanabiliyor, page `actions` listesiyle UI/API'ye açılabiliyor.

Tamamlananlar:

- Top-level `action Name { ... }` syntax'ı eklendi.
- Action source, primitive input, deterministic `set`, optional `allow`, optional `success` desteklendi.
- Parser/AST/validator action declaration'ı ve page action binding'i anlıyor.
- Validator computed/relation field set etmeyi, bilinmeyen operandları, type mismatch'i, input-field çakışmasını, source mismatch'i ve allow role hatalarını yakalıyor.
- Generated web output action route, API client method, backend input validator, React row button/form panel, OpenAPI schema ve auth/role varsa audit log üretiyor.
- Query-bound sayfalarda custom action sonrası liste server query'den yeniden yükleniyor.
- AI yüzeyleri güncellendi: `docs action --json`, `explain action --json`, BlackIR, inspect, affected graph, diagnostics.
- Warehouse örneğine `RestockProduct` eklendi; Products ve LowStock sayfalarından kullanılabiliyor.
- Yerel dokümanlar ve website kaynakları action MVP'yi anlatacak şekilde güncellendi.
- Generated numeric validation tutarlılığı iyileştirildi: mevcut SQLite/Prisma runtime'da `number`/`integer` whole-number, `decimal`/`money` finite fractional değer olarak doğrulanıyor.

Doğrulamalar:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs action --json` geçti.
- `go run ./cmd/black explain action --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `npm run build` generated web çıktısında geçti.

Notlar:

- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.

## Phase 22 - Generated Contract Tests MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web output artık kendi sözleşmesini hızlıca test edebiliyor. `black build` sonrasında üretilen uygulama içinde `npm test`, OpenAPI yolları/schemaları ile generated validation yüzeyini kontrol ediyor; böylece AI ajanları ve geliştiriciler build'in sadece derlenmesini değil, contract yüzeyinin de tutarlı kalmasını ölçebiliyor.

Yapılanlar:
- Generator `src/blacklang.contract.test.ts` dosyasını üretiyor.
- Generated `package.json` içine `"test": "tsx src/blacklang.contract.test.ts"` script'i eklendi.
- Contract test `openapi.json` içinden entity schema, action schema, CRUD path, bound query path, bound custom action path, explicit API path ve custom action metadata kayıtlarını doğruluyor.
- Contract test generated validation fonksiyonlarını import edip temsilî geçerli/geçersiz entity ve action payload'larını denetliyor.
- `number` ve `integer` alanlarında kesirli değerlerin reddedilmesi smoke test kapsamına alındı.
- Compiler testleri yeni generated test dosyasını ve custom action contract test beklentilerini kontrol edecek şekilde güncellendi.
- `docs generated-test --json` ve `explain generated-test --json` için yeni CLI docs keyword'ü eklendi.
- `docs/generated-test.md`, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, agent contract, llms dosyaları, sitemap ve website source güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs generated-test --json` geçti.
- `go run ./cmd/black explain generated-test --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.

Not:
Bu hâlâ gelecekteki BlackLang test dili değildir. Bu aşama generated web app için deterministic contract smoke test tabanı sağlar. Frontend component testleri, gerçek API route integration testleri, fixture/seed ve e2e test dili sonraki test aşamalarında açık kalır.

## Phase 22 Addendum - API/Frontend Smoke, Benchmark Command, and Golden Manifest

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web output artık sadece TypeScript build ile değil, contract, API server ve React render smoke testleriyle de doğrulanıyor. Ayrıca `black benchmark --json|--ir`, BlackLang kaynak boyutu ile generated web output boyutunu deterministic şekilde ölçebiliyor.

Yapılanlar:
- Generated server `createApp()` export edecek şekilde import edilebilir hale getirildi; `listen` sadece `src/server.ts` direkt çalıştırıldığında başlıyor.
- `src/blacklang.api.test.ts` üretildi. Test random localhost portunda generated Express app'i açıyor, `/openapi.json`, auth varsa anonymous 401, auth yoksa JSON 404 ve varsa CORS allow/deny davranışını kontrol ediyor.
- `src/blacklang.frontend.test.tsx` üretildi. Test generated React `App` bileşenini `react-dom/server` ile render edip temel render zincirini doğruluyor.
- Generated `npm test` zinciri artık contract + API + frontend smoke testlerini çalıştırıyor.
- `black benchmark [file] --out <dir> --json|--ir` eklendi. Komut projeyi parse/validate ediyor, output'u geçici dizinde üretip source/generated file/line/byte/ratio/kind metriklerini raporluyor ve configured output dizinini değiştirmiyor.
- `black agent startup --json` artık generated `npm test` ve `black benchmark --json` komutlarını öneriyor.
- Warehouse generated output için kompakt golden manifest testi eklendi: path, kind, satır, byte ve SHA-256 drift'i yakalanıyor.
- README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, docs/generated-test.md, docs/benchmark.md, docs/README.md, agent contract, llms dosyaları, sitemap ve website source güncellendi.

Ölçüm:
- Warehouse explicit `.black` source: 257 satır.
- Warehouse source + theme: 284 satır.
- Generated files: 43.
- Generated lines: 6772.
- Generated/black-source line ratio: 26.35.
- Generated/source+theme line ratio: 23.85.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs generated-test --json` geçti.
- `go run ./cmd/black docs benchmark --json` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.

Not:
Generated smoke testleri gelecekteki BlackLang test syntax'ının yerine geçmez. Bunlar bugünkü generated web app contract/API/frontend sağlığını yakalayan ilk güvenlik ağıdır.

## Phase 23 - Protected Source Mode MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` source artık production paketinden ayrı tutulabilen `.black.enc` protected source formatına çevrilebiliyor. Compiler, gerekli key environment variable set edildiğinde encrypted source'u bellekte decrypt edip parse/lint/validate/inspect/benchmark/security scan/build çalıştırabiliyor; plaintext source dosyaya yazılmıyor.

Yapılanlar:
- `.black.enc` header formatı eklendi: `BLACKLANG-ENC v1`, `AES-256-GCM`, `SHA256-ENV`, `keyEnv`, `nonce`, `---`, base64 ciphertext.
- Key kuralı netleştirildi: key değeri yalnızca environment'tan okunur; `.black` veya `.black.enc` içine yazılmaz.
- `black security encrypt <file> [--out <file.black.enc>] [--key-env <ENV>] --json|--ir` eklendi.
- `black security decrypt <file.black.enc> --stdout [--key-env <ENV>] --json|--ir` eklendi; `--stdout` yoksa `MISSING_DECRYPT_STDOUT` döner.
- `ReadBlackSource` merkezi kaynak okuyucusu eklendi; normal `.black` dosyayı doğrudan, `.black.enc` dosyayı bellekte decrypt ederek okuyor.
- `parse`, `lint`, `validate`, `inspect`, `benchmark`, `security scan` ve `build` encrypted source'u aynı merkezi okuyucu üzerinden kullanıyor.
- `black format`, `.black.enc` dosyaları rewrite etmeyi reddediyor ve `UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE` döndürüyor.
- Production package exclusion `.black.enc` için testle sabitlendi.
- Security IR çıktıları encrypt/decrypt/scan error code'larını taşıyacak şekilde genişletildi.
- `docs/protected-source.md` eklendi.
- `docs security --json`, `explain security --json`, diagnostics referansı, agent contract, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source, llms dosyaları ve sitemap güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs security --json` geçti.
- `go run ./cmd/black explain security --json` geçti.
- `go run ./cmd/black security encrypted-source --json` geçti.
- `go run ./cmd/black security encrypt ../../examples/warehouse/app.black --out <tmp>/warehouse.app.black.enc --json` geçti.
- `go run ./cmd/black security decrypt <tmp>/warehouse.app.black.enc --stdout` geçti.
- `go run ./cmd/black parse <tmp>/warehouse.app.black.enc --json` geçti.
- `go run ./cmd/black lint <tmp>/warehouse.app.black.enc --json` geçti.
- `go run ./cmd/black validate <tmp>/warehouse.app.black.enc --json` geçti.
- `go run ./cmd/black inspect <tmp>/warehouse.app.black.enc --affected Product.stock --json` geçti.
- `go run ./cmd/black security scan <tmp>/warehouse.app.black.enc --json` geçti.
- `go run ./cmd/black benchmark <tmp>/warehouse.app.black.enc --out <tmp>/generated-from-enc --json` geçti.
- `go run ./cmd/black build <tmp>/warehouse.app.black.enc --out <tmp>/generated-from-enc --json` geçti.
- Generated encrypted-source build output içinde `.black` veya `.black.enc` dosyası bulunmadı.
- Negatif CLI kontrolünde `security decrypt` `--stdout` olmadan `MISSING_DECRYPT_STDOUT` döndürdü.
- Negatif CLI kontrolünde `black format <file.black.enc> --check --json` `UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE` döndürdü.
- Normal Warehouse doğrulaması geçti: format, lint, validate, build.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Geçici encrypted-source test klasörleri `.tmp/` altında kaldı; `.tmp/` git tarafından ignored durumda.

## Phase 27 - Web Coverage Report MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang web hedefinin ne kadar tamamlandığını yorumla değil, compiler tarafından üretilen makine-okunur coverage matrisiyle takip etmeyi mümkün kılıyor. `black benchmark coverage --json|--ir`, web hedefini ağırlıklı alanlara, stable issue ID'lerine ve milestone'lara bölüyor.

Yapılanlar:
- `CoverageResult`, `CoverageArea`, `CoverageIssue` ve `CoverageMilestone` JSON şemaları eklendi.
- `black benchmark coverage --json|--ir` eklendi.
- Current weighted web coverage signal %50 olarak raporlanıyor.
- Coverage areas toplam weight 100 olacak şekilde tanımlandı: data-model-crud, frontend-ui-layout, backend-api-actions, auth-permissions-security, database-runtime, deployment-operations, testing-benchmarks, docs-ai-ergonomics, ecosystem-integrations.
- Roadmap gap'leri stable issue ID'lerine çevrildi: `WEB-UI-001`, `WEB-LAYOUT-001`, `WEB-I18N-001`, `WEB-I18N-002`, `WEB-DATA-001`, `WEB-DATA-002`, `WEB-SEC-001`, `WEB-OPS-001`, `WEB-TEST-001`, `WEB-TEST-002`, `WEB-DOCS-001`, `WEB-ECO-001`.
- `black agent startup --json` içine `black benchmark coverage --json` komutu eklendi.
- `black docs benchmark --json`, `black docs coverage --json` ve `black explain coverage --json` desteklendi.
- `docs/benchmark.md`, `docs/README.md`, `docs/llms.txt`, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source ve website llms index güncellendi.
- Website source'a `#coverage` bölümü eklendi; coverage matrix, current signal ve issue ID yaklaşımı dokümante edildi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti.
- `go run ./cmd/black benchmark coverage --ir` geçti.
- `go run ./cmd/black docs coverage --json` geçti.
- `go run ./cmd/black explain coverage --json` geçti.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Coverage percent bilinçli olarak planlama sinyalidir; full web coverage iddiası değildir.

## Phase 28 - Page View Composition MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Page içindeki generated `table`, `detail` ve `form` bölümlerinin yalnızca sırasını değil, ilk layout composition niyetini de `.black` source içinde ifade etmeyi mümkün kılıyor. Bu sayede generated React/CSS elle düzenlenmeden grid/stack layout, gap, responsive stack breakpoint ve section span üretilebiliyor.

Yapılanlar:
- `view` AST genişletildi: `compose` ve `section` intent eklendi.
- Parser şu syntax'ı destekliyor: `compose stack|grid [columns 1..4] [gap sm|md|lg] [stackAt sm|md|lg|none]`.
- Parser şu syntax'ı destekliyor: `section table|detail|form span 1..4`.
- Validator unsupported compose mode, columns, gap, stackAt, section name, duplicate section ve span hatalarını diagnostic code ile raporluyor.
- Empty `view {}` için mevcut `MISSING_VIEW_ORDER` davranışı korundu.
- BlackIR ve inspect IR output'ları `view-compose` ve `view-section` satırlarını yazıyor.
- Generated React page root artık gerektiğinde `bl-view-compose-grid` veya `bl-view-compose-stack` class'ı alıyor.
- Generated table/detail/form panelleri gerektiğinde `bl-view-span-*` class'ı alıyor.
- Generated CSS order kurallarını koruyup stack/grid/gap/responsive stackAt ve section span kuralları üretiyor.
- Warehouse Products page view'i composition örneğine çevrildi: table span 2, detail/form span 1, md altında tek kolon.
- Coverage signal layout MVP sonrası %51'e güncellendi; `WEB-LAYOUT-001` daha dürüst şekilde nested/tabs/modal/drawer/richer interactive composition için açık kaldı.
- `docs/view.md`, `docs/diagnostics.md`, `docs/README.md`, `docs/llms.txt`, `black docs view --json`, `BLACKLANG.md`, `SPEC.md`, `ROADMAP-v0.2.md`, web yol haritası, website source ve website llms index güncellendi.
- Warehouse golden manifest yeni generated output hash/line count ile güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Products --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Root config üzerinden Warehouse benchmark güncel ölçümü: `.black` 261 satır, source+theme 288 satır, generated 43 dosya, generated 6801 satır, generated/black-source ratio 26.06, generated/source+theme ratio 23.61.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Tabs, modal, drawer ve nested interactive composition bu MVP'de bilerek desteklenmedi; gerçek runtime state/control üretimi gerektirdiği için `WEB-LAYOUT-001` altında açık kaldı.

## Phase 20 - UI Profile Migration Rules MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.blackthm` profil slotları değişirken eski `.black` inline UI değerlerinin yanlış slotlara kaymasını engellemeyi mümkün kılıyor. `black theme migrate <old.blackthm> <new.blackthm> --json|--ir`, eski ve yeni theme dosyasını read-only karşılaştırıyor ve CI/AI ajanı için `safe` sonucunu, değişiklikleri ve stable diagnostics listesini veriyor.

Yapılanlar:
- `ThemeMigrationResult`, `ThemeMigrationSummary` ve `ThemeMigrationChange` JSON şemaları eklendi.
- `black theme migrate <old.blackthm> <new.blackthm> --json|--ir` komutu eklendi.
- Migration güvenlik kuralları eklendi: theme name/target stabil kalır, theme/profile version geriye gitmez, locked profile unlock edilemez, profile name stabil kalır, eski mode'lar silinemez ve eski slot listesi yeni slot listesinin exact prefix'i olmak zorundadır.
- Güvenli değişiklikler `theme-version-advanced`, `profile-version-advanced`, `profile-locked`, `mode-added` ve `slot-appended` change kayıtlarıyla raporlanıyor.
- Güvensiz değişiklikler `MISSING_THEME_MIGRATION_FILES`, `THEME_MIGRATION_NAME_CHANGED`, `THEME_MIGRATION_TARGET_CHANGED`, `THEME_VERSION_REGRESSION`, `UI_PROFILE_MIGRATION_NAME_CHANGED`, `UI_PROFILE_VERSION_REGRESSION`, `UI_PROFILE_UNLOCKED`, `UI_MODE_REMOVED` ve `UI_SLOT_MIGRATION_BREAK` diagnostic kodlarıyla raporlanıyor.
- `black docs theme-migration --json` ve `black explain theme-migration --json` eklendi.
- `docs/theme-migration.md` eklendi; theme/ui profile docs, diagnostics referansı, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source, sitemap ve llms dosyaları güncellendi.
- Coverage raporunda `WEB-UI-001` done oldu; frontend-ui-layout score 55, docs-ai-ergonomics score 72 ve current weighted web coverage signal %52 oldu.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs theme-migration --json` geçti.
- `go run ./cmd/black explain theme-migration --json` geçti.
- Safe migration CLI kontrolü `safe: true` ve `slot-appended` döndürdü.
- Unsafe migration CLI kontrolü beklenen şekilde non-zero çıktı ve `UI_PROFILE_UNLOCKED` + `UI_SLOT_MIGRATION_BREAK` döndürdü.
- `go run ./cmd/black docs --all --json` `count: 51` ve `theme-migration` döndürdü.
- `go run ./cmd/black benchmark coverage --json` `completionPercent: 52` ve `WEB-UI-001` status `done` döndürdü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Bu MVP intentional reorder için otomatik source rewrite yapmaz; reorder/silme/araya ekleme durumunu unsafe raporlar.

## Phase 28 - Page View Tabs Composition MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated page içindeki `table`, `detail` ve `form` bölümlerini `.black` kaynakta tab gruplarına bağlamayı mümkün kılıyor. Kullanıcı veya AI ajanı React state, tab button ve conditional render kodunu elle yazmadan `compose tabs` ve `tab <Name> sections ...` ile sekmeli web UI üretebiliyor.

Yapılanlar:
- `PageViewDecl` AST içine `tabs` eklendi; her tab `name`, `sections` ve `position` taşıyor.
- Parser `view` içinde `tab <Name> sections <section...>` satırını destekliyor.
- `compose tabs` desteklenen view compose mode oldu.
- Validator şu kuralları ekledi: tabs yalnızca `compose tabs` ile kullanılır; `compose tabs` en az bir `tab` satırı ister; tabs modunda her ordered section tam bir tab içinde yer alır; duplicate tab adı, duplicate tab section ve unsupported tab section diagnostic üretir; tabs `columns` ve `stackAt` kabul etmez.
- BlackIR ve inspect IR artık `view-tab <Name> sections ...` satırlarını üretiyor.
- Web generator `activeViewTab` state'i, `nav.view-tabs` tab controls ve `view-tab-panel` conditional section wrapper'ları üretiyor.
- Detail tab aktif değilken row-level custom action panelleri de detail ile beraber gizleniyor.
- Base generated CSS içine `.view-tabs`, `.view-tabs button.active` ve `.view-tab-panel` kuralları eklendi.
- Warehouse Products page tabs örneğine geçirildi: `List` tab table, `Record` tab detail+form.
- `docs/view.md`, `black docs view --json`, `black explain view --json`, diagnostics referansı, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source ve llms dosyaları güncellendi.
- Coverage raporunda frontend-ui-layout score 60 oldu; `WEB-LAYOUT-001` açık kalmaya devam ediyor ama tabs kapsamı başlıktan çıkarıldı. Current weighted web coverage signal %52 olarak kaldı.
- Warehouse golden manifest generator üzerinden güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Products --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Root config üzerinden Phase 28 Warehouse benchmark ölçümü: `.black` 260 satır, source+theme 287 satır, generated 43 dosya, generated 6819 satır, generated/black-source ratio 26.23, generated/source+theme ratio 23.76.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Nested sections, modal, drawer ve daha derin interactive composition hâlâ `WEB-LAYOUT-001` altında açık.

## Phase 26 - Runtime Language Switching for Field Labels MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`i18n` ve `label Entity.field` translation bloklarını generated React uygulamasında runtime language selector'a bağlamayı mümkün kılıyor. Kullanıcı dili değiştirdiğinde table, detail, form, filter ve column visibility field label'ları yeniden render oluyor; computed display field label'ları da aynı mekanizmaya dahil.

Yapılanlar:
- Generated App içine multi-locale projelerde `locales`, `defaultLocale`, `activeLocale` state'i ve topbar language selector eklendi.
- Generated page component props içine `locale?: string` eklendi ve App page render ederken `locale={activeLocale}` geçiriyor.
- Page generator entity başına deterministic `fieldLabels` dictionary ve `fieldLabel(field, fallback, locale)` helper'ı üretiyor.
- Stored field ve computed display field label render noktaları runtime helper'a bağlandı: filter label, column visibility label, table header, detail label ve form label.
- Placeholder, help text, frontend validation message, date/number/currency format ve RTL kapsamı bilinçli olarak ayrı açık issue altında bırakıldı.
- `docs i18n --json`, `docs label --json`, `explain i18n --json` ve `explain label --json` metinleri runtime davranışı ve computed label hedefleriyle güncellendi.
- `docs/i18n.md`, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source ve llms dosyaları güncellendi.
- Coverage raporunda `WEB-I18N-001` done oldu, frontend-ui-layout score 62'ye çıktı ve current weighted web coverage signal %53 oldu.
- Warehouse generated output ve golden manifest generator üzerinden yenilendi.

Doğrulama:
- Odaklı generator/coverage testleri geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs i18n --json` geçti.
- `go run ./cmd/black explain i18n --json` geçti.
- `go run ./cmd/black docs label --json` geçti.
- `go run ./cmd/black explain label --json` geçti.
- `go run ./cmd/black benchmark coverage --json` ve `--ir` geçti; `completionPercent: 53` döndü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.inventoryValue --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Warehouse benchmark güncel ölçümü: `.black` 260 satır, source+theme 287 satır, generated 43 dosya, generated 6890 satır, generated/black-source ratio 26.50, generated/source+theme ratio 24.01.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- `WEB-I18N-002` hâlâ açık: placeholder/help/message/date/number/currency/RTL kapsamı ayrıca tasarlanmalı.

## Phase 28 - Responsive Grid Breakpoint Ladder MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Mevcut `compose grid ... stackAt sm|md|lg|none` syntax'ını canonical responsive layout niyeti olarak güçlendirir. Generated CSS artık grid'i tek breakpoint'te bir anda 1 kolona düşürmek yerine deterministic lg/md/sm media query ladder'ı üretebilir; geniş section span'leri dar breakpointlerde taşmasın diye clamp edilir.

Yapılanlar:
- Yeni syntax eklenmedi; mevcut `stackAt` davranışı genişletildi.
- `pageViewComposeMediaBlocks` generator akışı eklendi ve grid media query üretimi tek bloktan çoklu responsive ladder'a taşındı.
- `stackAt sm` için 4 kolonlu grid lg'de 3, md'de 2, sm'de 1 kolona düşecek şekilde deterministic CSS üretir.
- `stackAt md` ve `stackAt lg` daha erken collapse davranışını korur; `stackAt none` responsive grid breakpoint üretimini kapatır.
- Responsive kolon sayısından büyük `section ... span` değerleri ilgili breakpointte clamp edilir.
- Generator testi `TestBuildWebGeneratesResponsiveGridBreakpointLadder` eklendi.
- `docs/view.md`, `black docs view --json`, `black explain view --json`, README, BLACKLANG.md, SPEC.md, web yol haritası, website source ve llms dosyaları güncellendi.
- Coverage raporunda frontend-ui-layout score 64'e çıktı; current weighted web coverage signal %53 olarak kaldı. `WEB-LAYOUT-001` hâlâ nested sections, modal/drawer ve accessibility policy checks için açık.

Doğrulama:
- `go test ./cmd/black -run "TestBuildWebGeneratesResponsiveGridBreakpointLadder|TestBuildWeb"` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; frontend-ui-layout score `64`, completionPercent `53` döndü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.

Notlar:
- Git commit/push yapılmadı.
- Canlı siteye yayın yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Warehouse şu an `compose tabs` kullandığı için responsive grid ladder kaynak örnekte değil, generator test fixture'ında doğrulandı.

## Phase 21/Database Runtime — Entity Index Declarations MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor? `.black` entity tanımında performans ve veri erişim niyetini açıkça tutmayı mümkün kılar; generator bu niyetten Prisma ve SQLite için deterministik index üretir, AI ajanları da docs/explain/inspect üzerinden hangi dosyaların etkilendiğini görebilir.

Yapılanlar:

- Entity içinde tek syntax ile `index field` ve `index fieldA, fieldB` desteği eklendi.
- Index alanları stored field ve relation field ile sınırlandı; computed display field indexlenemez hale getirildi.
- Validator tarafında boş index, bilinmeyen alan, computed alan, tekrarlı alan, tekrarlı index ve dört alandan uzun index için diagnostics eklendi.
- Generator Prisma `@@index(..., map: "...")` ve SQLite `CREATE INDEX IF NOT EXISTS ...` üretir hale geldi.
- `inspect --affected Product.index --json` desteği eklendi; etki alanı DB schema/setup ile sınırlı raporlanıyor.
- `docs index --json`, `explain index --json`, docs, AI contract, README, SPEC, roadmap, llms ve website güncellendi.
- Warehouse örneğine `Product.index stock` ve `Order.index customer, status` niyetleri eklendi; generated output generator üzerinden yenilendi.
- Web coverage database-runtime skoru 40'a, weighted signal %54'e çıktı.

Doğrulama:

- `go test -count=1 ./...`
- `go run ./cmd/black --help`
- `go run ./cmd/black docs --all --json`
- `go run ./cmd/black docs index --json`
- `go run ./cmd/black explain index --json`
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black`
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json`
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json`
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.index --json`
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json`
- `npm run build`
- `npm test`
- `black benchmark coverage --json`
- `black benchmark --json`

Benchmark:

- Warehouse `.black` kaynak: 262 satır.
- Warehouse source+theme: 289 satır.
- Generated output: 43 dosya, 6900 satır.
- Generated/source+theme ratio: 23.88.

Not:

- Git commit/push ve VPS yayını bu turda yapılmadı; son inceleme sonrası birlikte yapılacak.

## Database Runtime — Read-only Schema Migration Plan MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor? Eski ve yeni `.black` source dosyaları arasında generated database shape farkını deployment öncesi görmeyi sağlar; tablo/kolon/index değişiklikleri safe, manual ve destructive olarak ayrılır, ama gerçek database'e bağlanılmaz ve migration uygulanmaz.

Yapılanlar:

- Yeni CLI komutu eklendi: `black migrate plan <old.black> <new.black> --json|--ir`.
- Migration plan parser/validator sonrası çalışır; source okunamaz, parse edilemez veya validate edilemezse `success=false` döner.
- Geçerli planlarda `safe`, `destructive`, `summary`, `changes`, `steps`, `generatedFiles` ve `agentNotes` alanları üretildi.
- Entity/table ekleme-silme, stored field/column ekleme-silme, tip değişimi, generated column name değişimi, required/optional geçişi, unique/default değişimi, relation target değişimi ve entity index ekleme-silme karşılaştırılıyor.
- Risk sınıfları deterministik hale getirildi: `safe`, `manual`, `destructive`.
- Rename tahmini bilinçli olarak yapılmadı; first-class rename syntax gelene kadar rename add/drop olarak raporlanıyor.
- `docs migrate --json`, `explain migrate --json`, diagnostics, agent policy, affected database notu, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source, llms ve sitemap güncellendi.
- Coverage database-runtime score 50'ye, weighted signal %55'e çıktı. `WEB-DATA-001` done; migration apply/runtime ve rename/refactor yeni açık iş olarak ayrıldı.

Doğrulama:

- `go test -count=1 ./...`
- `go run ./cmd/black --help`
- `go run ./cmd/black docs --all --json`
- `go run ./cmd/black docs migrate --json`
- `go run ./cmd/black explain migrate --json`
- `go run ./cmd/black migrate plan ../../.tmp/migration-old.black ../../.tmp/migration-new.black --json`
- `go run ./cmd/black migrate plan ../../.tmp/migration-old.black ../../.tmp/migration-new.black --ir`
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected database --json`
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black`
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json`
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json`
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json`
- `npm run build`
- `npm test`
- rebuilt `.tmp/cli-bin/black.exe`
- `black benchmark coverage --json`
- `black benchmark --json`

Benchmark:

- Warehouse source+theme: 289 satır.
- Generated output: 43 dosya, 6900 satır.
- Generated/source+theme ratio: 23.88.

Not:

- Git commit/push ve VPS yayını yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Bu MVP migration plan üretir; applied migration runner değildir.

## Phase 28 — View Section Modal/Drawer Composition MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor? Generated table/detail/form sayfa parçalarını sadece inline grid/stack/tabs içinde değil, form ve detail için modal veya drawer overlay olarak açmayı mümkün kılar. Admin-style web akışlarında liste görünür kalırken detay veya form panelini kaynak `.black` niyetiyle yönetir.

Yapılanlar:

- Mevcut `view section` syntax'ı genişletildi; yeni top-level syntax eklenmedi.
- Canonical syntax: `section <name> [span 1..4] [display inline|modal|drawer] [side left|right] [title "Text"]`.
- `display modal` ve `display drawer` detail/form section'ları için desteklendi; table overlay bu MVP'de validator tarafından reddediliyor.
- `side left|right` sadece drawer display ile geçerli hale getirildi; drawer side default'u `right`.
- `compose tabs` ile modal/drawer section display'i bu MVP'de birlikte kullanılmıyor; validator deterministik hata üretiyor.
- Parser ve AST `display`, `side`, `title` metadata'sını JSON/BlackIR içinde korur hale geldi.
- Validator yeni hata kodları üretiyor: `INVALID_VIEW_SECTION_OPTION`, `DUPLICATE_VIEW_SECTION_OPTION`, `UNSUPPORTED_VIEW_SECTION_DISPLAY`, `UNSUPPORTED_VIEW_SECTION_SIDE`.
- Generator detail overlay için selected/read akışına bağlı drawer/modal wrapper üretir.
- Generator form overlay için create/edit butonlarından açılan panel state'i, close/reset davranışı ve stable CSS class'ları üretir.
- Generated CSS'e `.section-overlay`, `.section-overlay-modal`, `.section-overlay-drawer-*`, `.panel-header`, `.bl-view-display-*` ve `.bl-view-side-*` kuralları eklendi.
- Warehouse `Orders` sayfası `detail display drawer side right title "Order Details"` ve `form display modal title "Order Form"` kullanacak şekilde güncellendi.
- `docs/view.md`, `black docs view --json`, `black explain view --json`, diagnostics, README, BLACKLANG.md, SPEC.md, ROADMAP-v0.2.md, web yol haritası, website source ve llms dosyaları güncellendi.
- Coverage frontend-ui-layout score 71'e, weighted signal %56'ya çıktı. `WEB-LAYOUT-001` done; nested/richer composition `WEB-LAYOUT-002` olarak açık kaldı.

Doğrulama:

- `go test ./cmd/black -run "TestParsePageViewComposition|TestValidatePageViewCompositionErrors|TestBuildWebGeneratesViewSectionModalAndDrawer"` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Orders --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Warehouse golden manifest compiler çıktısından güncellendi.
- rebuilt `.tmp/cli-bin/black.exe`
- `black benchmark coverage --json` geçti.
- `black benchmark --json` geçti.

Benchmark:

- Warehouse `.black` kaynak: 270 satır.
- Warehouse source+theme: 297 satır.
- Generated output: 43 dosya, 7012 satır.
- Generated/source+theme ratio: 23.61.

Not:

- Git commit/push ve VPS yayını yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Overlay trigger DSL eklenmedi; detail selected/read akışına, form create/edit akışına bağlıdır.

## Phase 28 — View Nested Section Groups MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor? Generated page section'larını yalnızca düz kardeşler olarak değil, kaynak `.black` içinde tanımlanan nested layout wrapper altında birlikte göstermeyi sağlar. Böylece liste + record workspace gibi admin UI düzenleri generated React/CSS elle düzenlenmeden kurulabilir.

Yapılanlar:

- `view` içine canonical `group <Name> sections <section...> [compose stack|grid] [columns 1..4] [gap sm|md|lg] [span 1..4] [title "Text"]` syntax'ı eklendi.
- Group yalnızca mevcut `table`, `detail`, `form` section'larını sarar; yeni bağımsız section türü icat edilmedi.
- Group section'ları generated render order içinde contiguous olmak zorunda; `table, form` gibi araya section alan kombinasyonlar validator tarafından reddediliyor.
- Group, `compose stack` ve `compose grid` destekliyor; grid için columns/gap, dış grid için span, başlık için title metadata'sı eklendi.
- Group tabs ile bu MVP'de birleşmiyor; tabs zaten kendi section grouping davranışına sahip.
- Modal/drawer display kullanan section'lar group içine alınmıyor; grouped section'lar inline kalıyor.
- AST, parser, validator, BlackIR, docs/explain output, diagnostics, generator ve tests güncellendi.
- Generator stable `.view-group`, `.view-group-title`, `.bl-view-group-*`, `.bl-view-group-compose-*` ve group span/order CSS kuralları üretir hale geldi.
- Warehouse `Customers` sayfasına `CustomerWorkspace` group örneği eklendi.
- Coverage frontend-ui-layout score 78'e, weighted signal %57'ye çıktı. `WEB-LAYOUT-002` done; `WEB-LAYOUT-003` accessibility policy checks ve advanced interactive triggers için açık kaldı.

Doğrulama:

- `go test ./cmd/black -run "TestFindViewDoc|TestParsePageViewComposition|TestValidatePageViewCompositionErrors|TestValidatePageViewGroupRejectsTabsAndOverlaySections|TestBuildWebGeneratesViewSectionGroups|TestBuildWebGeneratesViewSectionModalAndDrawer|TestFormatBlackIR|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR"` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Customers --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Warehouse golden manifest compiler çıktısından güncellendi.
- rebuilt `.tmp/cli-bin/black.exe`
- `black benchmark coverage --json` geçti.
- `black benchmark --json` geçti.

Benchmark:

- Warehouse `.black` kaynak: 277 satır.
- Warehouse source+theme: 304 satır.
- Generated output: 43 dosya, 7080 satır.
- Generated/source+theme ratio: 23.29.

Not:

- Git commit/push ve VPS yayını yapılmadı.
- Generated dosyalar elle düzenlenmedi; `generated/` çıktısı compiler ile yeniden üretildi.
- Group MVP advanced trigger/event sistemi değildir; sadece contiguous inline section'lar için layout wrapper üretir.

## Accessibility Audit Policy Checks MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web UI composition artık read-only bir accessibility policy komutuyla denetlenebiliyor. AI ajanları modal/drawer section ve nested view group gibi yapıların erişilebilir isimlere sahip olup olmadığını JSON/IR üzerinden anlayabiliyor.

Yapılanlar:
- `black audit accessibility [file] [--json|--ir]` komutu eklendi.
- Overlay `detail`/`form` section için `title` yoksa `ACCESSIBILITY_MISSING_OVERLAY_TITLE` üretiliyor.
- Birden fazla section saran view group için `title` yoksa `ACCESSIBILITY_MISSING_GROUP_TITLE` üretiliyor.
- Parser/validator hataları önce dönüyor; source hatalıysa accessibility bulguları source üstüne bindirilmiyor.
- `docs accessibility --json`, `explain accessibility --json`, diagnostics, AI agent contract, README, SPEC, BLACKLANG, llms.txt ve site kaynakları güncellendi.
- Coverage matrisi frontend-ui-layout tarafında accessibility audit’i tamamlanmış madde olarak işaretledi; toplam web coverage `%58` oldu.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs accessibility --json` geçti.
- `go run ./cmd/black explain accessibility --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti; bulgu yok.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --ir` geçti; `findings 0`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `npm run build` geçti.
- `npm test` geçti.
- `.tmp/cli-bin/black.exe benchmark coverage --json` geçti; frontend-ui-layout score `84`, toplam coverage `%58`.

Notlar:
- Bu faz generated dosyaları elle değiştirmedi; Warehouse output compiler üzerinden yeniden üretildi.
- Commit/push/deploy yapılmadı; kullanıcıyla toplu incelemeden sonra yapılacak.

## Page View Interaction Triggers MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web sayfalarında tab/overlay akışını `.black` kaynakta deterministik olarak tarif etmeyi mümkün kılar. AI ajanları artık “row seçilince detail tabına geç”, “create/edit başlayınca form bölümünü aç”, “save sonrası detail veya table bölümüne dön” gibi davranışları generated React'e elle dokunmadan ifade edebilir.

Yapılanlar:
- `view` bloğuna `trigger <section> on <event>` syntax'ı eklendi.
- Desteklenen event'ler: `rowSelect`, `createStart`, `editStart`, `saveSuccess`, `close`.
- Desteklenen kombinasyonlar:
  - `trigger detail on rowSelect`
  - `trigger form on rowSelect`
  - `trigger form on createStart`
  - `trigger form on editStart`
  - `trigger detail on saveSuccess`
  - `trigger table on saveSuccess`
  - `trigger table on close`
- Aynı page içinde aynı trigger event'i bir kez kullanılacak şekilde validator kuralı eklendi.
- Trigger'ın ihtiyaç duyduğu CRUD action yoksa `UNSUPPORTED_VIEW_TRIGGER_ACTION` üretiliyor.
- Parser, AST, validator, BlackIR, generator ve tests güncellendi.
- Generated React, tabbed sayfalarda `activeViewTab` değiştiriyor; overlay sayfalarda mevcut form/detail panel state akışını kullanıyor.
- Generated page root'a `bl-view-has-triggers` class'ı eklendi.
- Warehouse örneğinde Products page tab trigger'ları ve Orders page overlay trigger'ları eklendi.
- `docs view --json`, `explain view --json`, diagnostics, README, SPEC, BLACKLANG, ai-agent-contract, llms.txt, roadmap ve web sitesi kaynakları güncellendi.
- Coverage matrisi `WEB-LAYOUT-004` maddesini done yaptı; frontend-ui-layout score `92`, toplam web coverage `%59` oldu.
- Golden generated manifest compiler çıktısı değiştiği için `UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black -run TestWarehouseGeneratedGoldenManifest` ile güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Products --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `npm run build` geçti.
- `npm test` geçti.
- `.tmp/cli-bin/black.exe benchmark --json` geçti.
- `.tmp/cli-bin/black.exe benchmark coverage --json` geçti; completion `%59`.

Benchmark:
- Warehouse `.black` kaynak: 287 satır.
- Warehouse source+theme: 314 satır.
- Generated output: 43 dosya, 7098 satır.
- Generated/source+theme ratio: 22.61.

Notlar:
- Trigger MVP arbitrary frontend event handler değildir; sadece compiler'ın bildiği generated table/detail/form state geçişlerine bağlanır.
- Generated dosyalar elle düzenlenmedi; output compiler ile üretildi.
- Commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Generated i18n Coverage Genişletmesi MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web UI artık sadece field label çevirisiyle sınırlı değil. Stored form field placeholder/help/message metinleri, field-level frontend validation mesajları, table/detail number-money-date formatting ve temel RTL direction desteği `.black` kaynaktan deterministic üretilebiliyor.

Yapılanlar:
- Top-level `placeholder Entity.field { locale "Text" }`, `help Entity.field { ... }` ve `message Entity.field { ... }` syntax'ı eklendi.
- Yeni AST JSON alanları eklendi: `placeholders`, `helpTexts`, `messages`.
- Validator bu blokların `i18n` gerektirdiğini, locale/default locale uyumunu, duplicate hedefleri ve duplicate locale satırlarını denetliyor.
- `placeholder/help/message` translation hedefleri stored field ile sınırlandı; computed display fields için yalnızca `label Entity.computedField` geçerli kalıyor.
- Yeni diagnostic kodları eklendi: `INVALID_*_DECLARATION`, `INVALID_*_TRANSLATION`, `UNCLOSED_*`, `DUPLICATE_*_TARGET`, `INVALID_*_TARGET`, `UNKNOWN_*_TARGET`, `UNSUPPORTED_*_TARGET`, `MISSING_*_TRANSLATION`, `UNKNOWN_*_LOCALE`, `DUPLICATE_*_LOCALE`, `MISSING_DEFAULT_*_TRANSLATION`.
- BlackIR top-level `placeholder`, `help` ve `message` translation bloklarını yazıyor.
- `inspect --affected i18n --json` desteği eklendi; App shell ve generated page dosya etkileri gösteriliyor.
- Generated React App runtime locale için `lang` ve deterministic `dir` attribute'ları üretiyor. `ar`, `fa`, `he`, `ur` ve region varyantları RTL sayılıyor.
- Generated pages runtime field text map'leri üretiyor: `fieldLabels`, `fieldPlaceholders`, `fieldHelpTexts`, `fieldMessages`.
- Generated form input/select placeholder, help text ve field-level validation mesajları aktif locale ile re-render oluyor.
- Generated table/detail `number`, `integer`, `decimal`, `money`, `date` ve `datetime` değerleri aktif locale ile `Intl` üzerinden formatlanıyor.
- Computed display field render'ları `money` gibi tip bilgisini formatter'a geçiriyor.
- Warehouse örneğine Product field placeholder/help/message translation blokları eklendi.
- `docs i18n/placeholder/help/message --json`, `explain i18n/placeholder/help/message --json`, diagnostics, README, SPEC, BLACKLANG, AI contract, llms.txt, roadmap ve web sitesi kaynakları güncellendi.
- Coverage matrisi `WEB-I18N-002` maddesini done yaptı; frontend-ui-layout score `97`, toplam web coverage `%60` oldu.
- Golden generated manifest compiler çıktısı değiştiği için `UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black -run TestWarehouseGeneratedGoldenManifest` ile güncellendi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs i18n --json` geçti.
- `go run ./cmd/black docs placeholder --json` geçti.
- `go run ./cmd/black docs help --json` geçti.
- `go run ./cmd/black docs message --json` geçti.
- `go run ./cmd/black explain i18n --json` geçti.
- `go run ./cmd/black explain placeholder --json` geçti.
- `go run ./cmd/black explain help --json` geçti.
- `go run ./cmd/black explain message --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected i18n --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- `.tmp/cli-bin/black.exe benchmark --json` geçti.
- `.tmp/cli-bin/black.exe benchmark coverage --json` geçti; completion `%60`.

Benchmark:
- Warehouse `.black` kaynak: 342 satır.
- Warehouse source+theme: 369 satır.
- Generated output: 43 dosya, 7328 satır.
- Generated/source+theme ratio: 19.86.

Notlar:
- `message Entity.field` frontend form field-level validation mesajlarını localize eder; entity-level `validate ... message "Text"` metinleri current draft'ta açık validation metadata olarak kalır.
- Full app chrome/action copy translation catalog hâlâ ayrı bir sonraki frontend i18n genişleme alanıdır.
- Generated dosyalar elle düzenlenmedi; output compiler ile üretildi.
- Commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## PostgreSQL Runtime Target MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang web generator artık default SQLite yanında `database postgres` hedefinden PostgreSQL runtime çıktısı üretebiliyor. Böylece production'a daha yakın veritabanı hedefi, `.black` içinde tek deterministic target seçimiyle Prisma provider, adapter, setup ve Docker Compose seviyesine taşınıyor.

Yapılanlar:
- Validator `target web { frontend react backend node database postgres }` kullanımını resmi destekli database hedefi olarak kabul ediyor; `postgresql` alias'ı eklenmedi.
- Parser `INVALID_TARGET_DATABASE` yardım metni `database sqlite` ve `database postgres` örneklerini gösteriyor.
- Generator database hedefi için merkezi helper'lar ekledi: target database, database env adı, default database URL, Prisma provider ve Compose database URL.
- `database { url env NAME }` artık generated runtime tarafından okunuyor; NAME verilmezse eski default `DATABASE_URL` davranışı korunuyor.
- PostgreSQL target için generated `package.json` `@prisma/adapter-pg`, `pg` ve `@types/pg` kullanıyor; SQLite target mevcut `better-sqlite3` yolunda kalıyor.
- PostgreSQL target için `prisma/schema.prisma` `provider = "postgresql"` üretiyor.
- Auth kullanılan PostgreSQL output'ta `BlackUser`, `BlackSession` ve `BlackAuditLog` Prisma modelleri schema'ya ekleniyor.
- PostgreSQL target için generated `src/db.ts` PrismaPg adapter ile `connectionString` kullanıyor.
- PostgreSQL target için generated `src/setup-db.ts` env default'u kurup `npm run db:push:native` üzerinden Prisma schema push çalıştırıyor.
- PostgreSQL target için generated `src/routes/auth.ts` SQLite direct SQL yerine Prisma client ile register/login/logout/session/users/audit/role update üretiyor.
- Docker Compose, PostgreSQL target'ta ayrı `postgres` service, healthcheck, postgres volume ve app database URL fallback'i üretiyor; SQLite target'taki `/app/data` volume davranışı korunuyor.
- `docs target --json`, `explain target --json`, affected database metni, SPEC, BLACKLANG, ROADMAP, web yol haritası, llms özetleri ve docs site kaynakları güncellendi.
- Coverage matrisi `WEB-DATA-002` maddesini done yaptı; database-runtime score `62`, toplam web coverage `%61` oldu.
- Generator testlerine `TestBuildWebSupportsPostgresRuntime` eklendi; Postgres output package/schema/db/setup/auth/compose dosyaları fixture seviyesinde denetleniyor.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestValidateTargetErrors|TestValidateTargetAllowsPostgresDatabase|TestBuildWebWritesExpectedFiles"` geçti.
- `go test -count=1 ./cmd/black -run "TestValidateTargetErrors|TestValidateTargetAllowsPostgresDatabase|TestBuildWebWritesExpectedFiles|TestBuildWebSupportsPostgresRuntime"` geçti.
- `go test -count=1 ./cmd/black` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs target --json` geçti.
- `go run ./cmd/black docs database --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain target --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected database --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%61`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Warehouse generated app içinde `npm run build` geçti.
- Warehouse generated app içinde `npm test` geçti.
- Temp `database postgres` SupportDesk app üretildi; `npm install`, `npm run build` ve `npm test` geçti.

Benchmark / coverage:
- Weighted web coverage signal: `%61`.
- database-runtime score: `62`.
- `WEB-DATA-002` status: `done`.

Notlar:
- PostgreSQL teknik output'u Prisma'nın resmi `postgresql` provider adı ve `@prisma/adapter-pg` PrismaPg driver adapter yolunu kullanır; BlackLang syntax ise kısa ve tek kalır: `database postgres`.
- Temp testte `.tmp` altındaki ilk deneme Express `sendFile` dot-directory güvenliği nedeniyle `/openapi.json` 404 verdi; aynı generated Postgres app noktasız temp klasörde başarıyla geçti. Bu generator runtime hatası değildi.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler ile yeniden üretildi.
- Commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Applied Migration Runtime and Rename/Refactor MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Plan seviyesindeki schema migration bilgisi artık generated web runtime tarafından uygulanabilen açık refactor niyetine dönüştü. `.black` içinde entity ve stored field rename kararları tek canonical syntax ile yazılıyor; generator bunlardan manifest, SQL önizlemesi ve SQLite/PostgreSQL setup runtime üretip mevcut veriyi güvenli ad taşıma akışına bağlıyor.

Yapılanlar:
- Top-level `migration <Name> { ... }` syntax'ı eklendi.
- Desteklenen MVP rename syntax'ları:
  - `rename entity OldEntity to NewEntity`
  - `rename field Entity.oldField to newField`
- Parser, AST, formatter, validator, BlackIR, docs/explain, diagnostics ve inspect/affected output'ları migration deklarasyonlarını anlayacak şekilde güncellendi.
- Validator migration adlarını, boş migration bloklarını, duplicate migration'ları, duplicate rename hedeflerini, mevcut current-source rename kaynaklarını ve computed field rename girişimlerini deterministik diagnostic'lerle reddediyor.
- `black migrate plan old.black new.black --json` artık explicit rename deklarasyonlarını `rename-table` ve `rename-column` safe change olarak raporluyor.
- Entity rename sırasında relation field hedefi de normalize edildi; örneğin eski `ProductItem` entity'si yeni `Product` entity'sine rename edildiyse `product ProductItem` -> `product Product` yapay destructive type/relation change üretmiyor.
- Generator `migrations/manifest.json` ve deterministic `migrations/NNN_name.sql` dosyaları üretiyor.
- Generated SQLite setup runtime `BlackMigration` tablosu ile migration idempotency kaydı tutuyor ve rename işlemlerini schema setup öncesinde transaction içinde uyguluyor.
- Generated PostgreSQL setup runtime aynı manifest yaklaşımıyla `pg` client üzerinden rename migration'larını Prisma schema push öncesinde uyguluyor.
- Prisma schema migration kullanılan projelerde `BlackMigration` modelini üretiyor.
- Warehouse örneğine `RenameWarehouseProduct` migration'ı eklendi: `WarehouseProduct -> Product` ve `quantity -> stock`.
- README, SPEC, BLACKLANG, ROADMAP, ROADMAP-v0.2, AI agent contract, diagnostics, migration docs, llms özetleri ve web sitesi kaynakları güncellendi.
- Coverage matrisi `WEB-DATA-003` maddesini done yaptı; database-runtime score `72`, weighted web coverage signal `%62` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black explain migration --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected migration --json` geçti.
- `go run ./cmd/black migrate plan ../../examples/warehouse/app.black ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- SQLite gerçek migration smoke test geçti: eski `WarehouseProduct.quantity` satırı generated setup sonrası `Product.stock` olarak korundu ve `BlackMigration` kaydı oluştu.
- Temp PostgreSQL target app için generated migration runtime TypeScript build'i geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%62`.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --json` geçti; Warehouse source `347` satır, generated output `45` dosya / `7484` satır.

Notlar:
- Bu MVP arbitrary data migration dili değildir; yalnızca stored table/column rename intent'ini destekler.
- Rename runtime source/target çakışmasında fail-fast davranır; veri kaybını sessizce düzeltmeye çalışmaz.
- Generated dosyalar elle düzenlenmedi; output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Owner/Tenant Entity Policy MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web uygulamalarında satır kapsamını `.black` kaynak seviyesine taşıyor. Artık entity bazında owner ve tenant policy yazılabiliyor; list/query/detail/mutation/action route'ları aynı auth context üzerinden deterministik şekilde scope ediliyor.

Yapılanlar:
- Entity içi canonical syntax eklendi: `policy owner ownerId` ve `policy tenant tenantId`.
- Normal `owner Customer required` gibi field/relation satırlarıyla karışmaması için policy deklarasyonu yalnızca `policy` keyword'üyle kabul ediliyor.
- Parser, AST, validator, BlackIR, formatter, docs/explain, diagnostics ve inspect/affected çıktıları policy deklarasyonlarını tanır hale getirildi.
- Validator policy için auth zorunluluğu, duplicate owner/tenant, duplicate policy field, unknown field, computed/relation/non-text field, missing `required` ve `unique` kullanımını stabil diagnostic kodlarıyla reddediyor.
- Page form alanlarında policy field kullanımı `UNSUPPORTED_POLICY_FORM_FIELD` ile engellendi.
- Custom action `set` satırlarının policy field yazması `UNSUPPORTED_ACTION_POLICY_FIELD` ile engellendi.
- Generator owner/tenant policy helper'larıyla list, query, detail, create, update, archive, restore, delete, workflow transition ve custom action route'larını row-scope uygular hale getirdi.
- Create/update input'ları current authenticated user context'ten policy field damgalıyor; kullanıcı payload'ındaki policy field değerleri writable input'tan ayrılıyor.
- Generated API client input type'ları ve OpenAPI input schema'ları policy field'ları dışarı açmıyor.
- Tenant policy kullanılan projelerde generated auth runtime `BlackUser.tenantId` alanını, SQLite/PostgreSQL schema/setup output'unu ve auth JSON output'unu üretiyor.
- Warehouse örneğine `tenantId` ve `ownerId` policy alanları eklendi; form alanları policy field içermeyecek şekilde korundu.
- `docs/policy.md`, docs/llms, website/llms, README, SPEC, BLACKLANG, ROADMAP-v0.2, web yol haritası ve site kaynakları güncellendi.
- Coverage matrisi `WEB-SEC-001` maddesini done yaptı; auth-permissions-security score `70`, weighted web coverage signal `%65` oldu.
- Warehouse benchmark güncellendi: 358 `.black` kaynak satırı, 45 generated dosya, 7738 generated satır, generated/source ratio `21.61`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestParseEntityPolicies|TestValidateEntityPolicies|TestValidateEntityPolicyDiagnostics|TestBuildWebGeneratesEntityPolicyRuntime" -v` geçti.
- `go test -count=1 ./cmd/black` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs policy --json` geçti.
- `go run ./cmd/black explain policy --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected policy --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%65`.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --json` geçti.

Notlar:
- Bu MVP tenant yönetim UI'ı veya multi-role user modeli değildir; policy field'ları current auth context ile deterministic route scope sağlar.
- Tenant değeri generated auth runtime'da şimdilik `default` olarak başlar; daha zengin tenant assignment ayrı bir sonraki güvenlik/admin aşamasıdır.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Runtime Ops Signals MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web uygulamasını deploy sonrası standart altyapı kontrollerine hazır hale getirir. `.black` kaynakta tek `ops` bloğuyla public health/readiness/metrics endpoint'leri, structured request logs ve Docker healthcheck davranışı tanımlanır; generated dosyalar elle düzenlenmeden runtime gözlemlenebilirliği artar.

Yapılanlar:
- Top-level canonical `ops { ... }` syntax'ı eklendi.
- Desteklenen MVP syntax:
  - `health path "/healthz"`
  - `readiness path "/readyz"`
  - `metrics path "/metrics"`
  - `logging requests`
- Parser, AST, validator, BlackIR, formatter, docs/explain, diagnostics ve inspect/affected output'ları ops deklarasyonlarını anlayacak şekilde güncellendi.
- Validator boş ops bloğunu, duplicate ops sinyallerini, duplicate path kullanımını, desteklenmeyen logging modunu ve unsafe path değerlerini stabil diagnostic kodlarıyla reddediyor.
- Generated Express server public `/healthz`, `/readyz` ve `/metrics` route'larını API auth/CSRF middleware'inden önce üretiyor.
- Readiness endpoint generated Prisma adapter üzerinden `SELECT 1` kontrolü yapıyor ve DB hazır değilse `503` döndürüyor.
- Metrics endpoint process-local request, error, status-code, uptime ve startedAt sayaçlarını döndürüyor.
- `logging requests`, gözlenen her request sonunda structured JSON console log yazıyor.
- Generated OpenAPI output ops endpoint'lerine `x-blacklang-ops` ve `x-blacklang-public` metadata ekliyor.
- Generated contract test OpenAPI ops path/metadata kayıtlarını kontrol ediyor.
- Generated API smoke test health/readiness/metrics endpoint'lerini gerçek HTTP server üzerinden probe ediyor.
- `deploy { target docker }` ve `ops.health` beraber olduğunda Dockerfile ve docker-compose app healthcheck'i generated health path'i probe ediyor.
- Docker healthcheck komutu Node 22 built-in `fetch` kullanıyor; `curl`/`wget` bağımlılığı eklenmedi.
- Docker healthcheck JavaScript'i `=>` yerine `function` syntax'ı kullanacak şekilde düzeltildi; Go JSON escaping yüzünden `\u003e` oluşup Node `-e` içinde kırılmıyor.
- Warehouse örneğine ops bloğu eklendi.
- `docs/ops.md`, docs/llms, website/llms, README, SPEC, BLACKLANG, ROADMAP, ROADMAP-v0.2, web yol haritası ve site kaynakları güncellendi.
- Coverage matrisi `WEB-OPS-002` maddesini done yaptı; deployment-operations score `55`, weighted web coverage signal `%67` oldu.
- Warehouse benchmark güncellendi: 365 `.black` kaynak satırı, 45 generated dosya, 8016 generated satır, generated/source ratio `21.96`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestParseOpsDeclaration|TestValidateOpsDiagnostics|TestBuildWebGeneratesOpsRuntime|TestFindOpsDoc|TestExplainOpsKeyword|TestAnalyzeAffectedOps|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestFormatBlackIR" -v` geçti.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `60`.
- `go run ./cmd/black docs ops --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black explain ops --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected ops --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; ops endpoints gerçek HTTP server üzerinde probe edildi.
- Generated Dockerfile/docker-compose içinde `/healthz` healthcheck'i doğrulandı ve `\u003e` escaping kalmadığı kontrol edildi.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%67`.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP external observability provider, tracing, preview deployment, cloud adapter veya rollback sistemi değildir; bunlar `WEB-OPS-001` altında açık kaldı.
- Ops endpoints public root-level probe olarak tasarlandı; `/api` altında değildir ve auth/CSRF tarafından engellenmez.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Explicit API Declared Runtime MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` içindeki explicit `api` deklarasyonlarını sadece OpenAPI contract kaydı olmaktan çıkarıp generated web runtime'da çalışan deterministic route'lara dönüştürür. AI ajanları artık `docs api`, `explain api` ve `inspect --affected ApiName` çıktısından hem sözleşmeyi hem generated runtime/test etkisini görebilir.

Yapılanlar:
- Existing canonical `api Name { method/path/param/query/public|private/webhook }` syntax'ı korundu; aynı davranış için yeni syntax eklenmedi.
- Generated Express server explicit API route'larını üretmeye başladı.
- Public explicit API route'ları auth middleware'den önce mount ediliyor.
- Private explicit API route'ları projede `auth` varsa generated auth/CSRF middleware arkasında mount ediliyor; auth olmayan projelerde geriye dönük uyum için route üretiliyor, metadata private kalıyor.
- Generated runtime path ve query parametrelerini declared primitive type'a göre doğruluyor.
- Route response'u deterministic JSON dönüyor: `api`, `status`, `runtime: "declared"`, `method`, `path`, `access`, `webhook`, `params`, `query`, `body`.
- Webhook route'ları `202 accepted` ve `status: "accepted"` dönüyor.
- `POST`, `PUT` ve `PATCH` explicit API route'ları parsed JSON body bilgisini response'a dahil ediyor; `GET` ve `DELETE` için body `null`.
- OpenAPI explicit API operation metadata'sına `x-blacklang-runtime: declared` eklendi.
- Generated contract test explicit API path, API adı, runtime metadata ve webhook metadata kontrolünü yapıyor.
- Generated API smoke test canlı Express server üzerinde public webhook route'unu `202`, private route'u auth varsa `401` olarak probe ediyor.
- Validator explicit API path'lerini safe `/api/...` shape ile sınırladı; `/api/auth`, trailing slash, query string, fragment, malformed `{param}` ve unsafe route karakterleri reddediliyor.
- Validator explicit API route'larının generated page CRUD, auth, bound query, custom action ve workflow route'larıyla çakışmasını `DUPLICATE_API_ROUTE` ile yakalıyor.
- Aynı route shape'e gelen farklı param adları da duplicate kabul ediliyor.
- `inspect --affected ApiName --json` artık `src/server.ts`, `openapi.json`, `src/blacklang.contract.test.ts` ve `src/blacklang.api.test.ts` dosyalarını listeliyor.
- Ops request logging generated route mount'larından sonra path kaybetmemesi için `req.originalUrl` kullanacak şekilde düzeltildi.
- `docs/api.md` eklendi; README, SPEC, BLACKLANG, ROADMAP, ROADMAP-v0.2, web yol haritası, diagnostics, AI agent contract, llms dosyaları, benchmark raporları ve website kaynakları güncellendi.
- Coverage matrisi `WEB-API-001` maddesini done yaptı; `WEB-API-002` explicit API business handler syntax için açık kaldı. backend-api-actions score `62`, weighted web coverage signal `%68` oldu.
- Warehouse generated benchmark güncellendi: 365 `.black` kaynak satırı, 45 generated dosya, 8127 generated satır, generated/source ratio `22.27`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestValidateExplicitAPI|TestAnalyzeAffectedExplicitAPI|TestBuildWebWritesExpectedFiles|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR" -v` geçti.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs api --json` geçti.
- `go run ./cmd/black explain api --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected LowStockReport --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%68`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; `/api/webhooks/stock` `202`, `/api/reports/low-stock/sample?limit=7` auth arkasında `401`, ops endpointleri ve CORS davranışı gerçek HTTP server üzerinde doğrulandı.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP explicit API business handler dili değildir; side-effect gereken işler için mevcut yol hâlâ page-bound `action` bloklarıdır.
- Webhook için route ve deterministic acknowledgement var; provider signature verification, retry queue ve side-effect handler syntax sonraki API/job fazına kaldı.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Aggregate Query Summary MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Page-bound custom query'lerin sadece bounded liste döndürmesini değil, aynı deterministic seçim kümesi üzerinden count/sum/avg/min/max gibi güvenli summary değerleri de üretmesini sağlar. AI ajanları dashboard/report ihtiyacında raw SQL'e veya generated React/Express dosyalarını elle düzenlemeye düşmeden `.black` içinden aggregate niyeti okuyabilir.

Yapılanlar:
- Existing canonical `query Name { source/where/sort/limit }` syntax'ı korundu; aynı davranış için ayrı query DSL veya ikinci syntax eklenmedi.
- `query` bloğuna repeatable `aggregate` clause eklendi.
- Desteklenen şekiller:
  - `aggregate lowStockCount count`
  - `aggregate totalStock sum stock`
  - `aggregate averagePrice avg price`
  - `aggregate lowestStock min stock`
  - `aggregate highestPrice max price`
- `count` field almaz; `sum`, `avg`, `min`, `max` stored numeric field ister (`number`, `integer`, `decimal`, `money`).
- Aggregate field'ları computed display field, relation field ve generated system field kullanamaz.
- Aggregate output adları query içinde unique olmalı ve `__proto__`, `constructor`, `prototype` gibi JS prototype-reserved adlar reddediliyor.
- AST/JSON inspect output'una `aggregates` alanı eklendi.
- BlackIR query output'u aggregate satırlarını koruyor; inspect IR artık query başına `filters` ve `aggregates` sayısını gösteriyor.
- Validator yeni stable diagnostics üretiyor: `INVALID_QUERY_AGGREGATE`, `DUPLICATE_QUERY_AGGREGATE`, `UNSUPPORTED_QUERY_AGGREGATE`, `MISSING_QUERY_AGGREGATE_FIELD`, `UNSUPPORTED_QUERY_AGGREGATE_FIELD`.
- Generated Express route eklendi: `GET /api/<page>/query/summary`.
- Summary route query list ile aynı `source`, `where`, archive mode ve row policy kurallarını kullanıyor; `sort` ve `limit` summary'ye uygulanmıyor.
- Summary route permission guard artık filter/sort/aggregate field read access kontrollerini birlikte yapıyor.
- Generated API client aggregate bulunan query-bound page için `querySummary(includeArchived)` metodu üretiyor.
- Generated React page tablo üstünde deterministic summary card'ları render ediyor ve query mutation sonrası list + summary'yi yeniden yüklüyor.
- Generated OpenAPI'ye `/query/summary` path'i, `x-blacklang-query-summary`, `x-blacklang-aggregates` ve required read field metadata'sı eklendi.
- Generated contract test query summary path ve aggregate metadata'yı doğruluyor.
- Generated API smoke test auth'lu projelerde `/api/lowstock/query/summary` anonymous isteğinin 401 döndüğünü gerçek Express server üzerinde probe ediyor.
- Explicit API route collision validator yeni generated query summary route'larıyla çakışmayı `DUPLICATE_API_ROUTE` olarak yakalıyor.
- `inspect --affected QueryName --json` ve field affected graph metinleri aggregate summary etkisini gösterecek şekilde güncellendi.
- Warehouse örneğinde `LowStockProducts` query'sine `lowStockCount`, `totalStock`, `averagePrice` aggregate'leri eklendi.
- `docs/query.md`, diagnostics, docs/explain output, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, generated-test docs, roadmap ve website kaynakları güncellendi.
- Coverage matrisi `WEB-API-003` maddesini done yaptı; backend-api-actions score `67`, weighted web coverage signal `%69` oldu.
- Warehouse benchmark güncellendi: 368 `.black` kaynak satırı, 45 generated dosya, 8306 generated satır, generated/source ratio `22.57`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestQuery|TestGeneratedQuery|TestValidateExplicitAPIRouteConflicts|TestWebCoverageReport|TestFormatCoverageIR" -v` geçti.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `60`.
- `go run ./cmd/black docs query --json` geçti ve aggregate syntax/diagnostics JSON'da göründü.
- `go run ./cmd/black explain query --json` geçti ve aggregate summary agent step'leri göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected LowStockProducts --json` geçti; generated page/api/route/openapi etkileri listelendi.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.price --json` geçti; aggregate field bağımlılığı query affected graph içinde göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%69`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; `/api/lowstock/query/summary` gerçek HTTP server üzerinde auth arkasında `401` olarak probe edildi.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP group-by, joins, query parametreleri, OR ifadeleri veya raw SQL değildir.
- Summary değerleri declared query'nin database seçim kümesini temsil eder; client-side table search/filter/pagination sonucu değiştirmez.
- Empty match set için `count` 0, diğer aggregate değerleri `null` döner.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Custom Action Transaction MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Custom action route'larında source row lookup, deterministic update ve auth/role varsa audit log yazımını tek database transaction niyetiyle çalıştırmayı sağlar. AI ajanı action'ın atomik çalışması gerektiğini `.black`, JSON/BlackIR, docs/explain, inspect/affected ve OpenAPI metadata'sından okuyabilir; generated route'u elle değiştirmeye gerek kalmaz.

Yapılanlar:
- Existing `action Name { source/input/set/allow/success }` syntax'ı korundu; aynı davranış için ikinci action DSL eklenmedi.
- `action` bloğuna tek canonical, standalone `transaction` clause eklendi.
- `transaction` ekstra token alamaz ve action başına bir kez yazılabilir.
- AST/parse JSON output'una `transaction: true` ve `transactionPosition` bilgisi eklendi.
- Parser stable diagnostics üretiyor: `INVALID_ACTION_TRANSACTION`, `DUPLICATE_ACTION_TRANSACTION`.
- BlackIR action output'u `transaction` satırını koruyor; inspect IR action satırı `transaction true/false` bilgisini gösteriyor.
- Generated Express custom action route'u transaction varsa `prisma.$transaction` ile üretiliyor.
- Transaction route içinde source row lookup, row policy-scoped lookup, deterministic `set` update'i ve audit insert aynı Prisma transaction callback'i içinde çalışıyor.
- SQLite/PostgreSQL ayrımında Prisma model property sorununa düşmemek için generated audit yazımı transaction client'ın parametreli `$executeRaw` tag'iyle mevcut `BlackAuditLog` tablosuna yapılıyor. Bu `.black` içinde raw SQL açmaz; sadece generator runtime tekniğidir.
- Division guard sadece `/` kullanan action'larda generated response status olarak üretiliyor; gereksiz `divisionByZero` union kontrolü çıkarıldı.
- OpenAPI custom action operation'ları transaction action'larda `x-blacklang-transaction: true` metadata'sı yayıyor.
- Generated contract test transaction metadata'sını doğruluyor.
- `black docs action --json`, `black explain action --json`, diagnostics docs, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, generated-test docs, roadmap ve docs sitesi kaynakları güncellendi.
- Warehouse `RestockProduct` action'ına `transaction` satırı eklendi; generated Warehouse output compiler üzerinden yeniden üretildi.
- Coverage matrisi `WEB-DATA-004` maddesini done yaptı; backend-api-actions score `72`, database-runtime score `78`, weighted web coverage signal `%70` oldu.
- Warehouse benchmark güncellendi: 369 `.black` kaynak satırı, 45 generated dosya, 8342 generated satır, generated/source ratio `22.61`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestParseCustomAction|TestValidateCustomAction|TestFormatCustomActionIRInspectAndAffected|TestBuildWebGeneratesCustomActionRuntimeAndUI|TestGeneratedCustomActionRouteExecutesValidatedMutation|TestWebCoverageReport|TestFormatCoverageIR" -v` geçti.
- İlk generated `npm run build` denemesi Prisma transaction client üzerinde `blackAuditLog` model property hatasını yakaladı; generator audit yazımı `$executeRaw` ile düzeltildi.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve reviewed golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs action --json` geçti; transaction syntax, örnek ve diagnostic kodları JSON'da göründü.
- `go run ./cmd/black explain action --json` geçti; agent step'leri transaction'ı biliyor.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti; agent note transaction etkisini gösterdi.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --ir` geçti; action satırı `transaction true` gösterdi.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%70`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; OpenAPI contract, API smoke, ops endpointleri, query summary auth probe ve frontend render smoke testleri geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP genel-purpose transaction block dili değildir; sadece custom action route'unun atomik runtime mode'udur.
- Validation, auth middleware, page access, action `allow` ve input parsing transaction başlamadan önce çalışır; database row lookup/update/audit kısmı transaction içindedir.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Localized App Chrome and Action Copy MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Mevcut `label ... { locale ... }` modelini field metinlerinin dışına taşıyıp generated web uygulamasındaki app chrome, page navigation, tablo/status metinleri, CRUD butonları, custom action butonları ve workflow transition butonları için de runtime dil seçiciyle deterministik yerelleştirme sağlar. AI ajanı UI metinlerini `.black` kaynaktan okuyabilir; generated React dosyalarına elle girilmez.

Yapılanlar:
- Yeni syntax eklenmedi; existing canonical `label <target> { locale ... }` yapısı korundu.
- Supported UI label target kataloğu eklendi:
  - `app.title`, `app.menu`, `app.language`, `app.logout`, `app.closeNavigation`, `app.primaryNavigation`, `app.checkingSession`, `app.noPages`, `app.source`, `app.recordForm`
  - `page.<Page>`
  - `table.search`, `table.showArchived`, `table.columns`, `table.filter`, `table.status`, `table.actions`, `table.selectVisibleRecords`, `table.selectRecord`, `table.empty`, `table.previous`, `table.next`, `table.page`, `table.of`
  - `status.active`, `status.archived`, `status.loadingRecords`, `status.loadingDetails`, `status.selectRecord`
  - `action.view`, `action.create`, `action.edit`, `action.delete`, `action.deleteSelected`, `action.archive`, `action.restore`, `action.cancel`, `action.close`, `action.saveChanges`, `action.saving`, `action.run`, `action.running`
  - `action.<CustomAction>`, `action.<workflowTransition>`, `action.new.<Entity>`, `action.create.<Entity>`, `action.edit.<Entity>`, `action.view.<Entity>`
- Validator UI label target'larını entity field target'larından ayrı ve deterministic biçimde doğruluyor; unsupported app/table/status/action key'leri stable diagnostics ile reddediliyor.
- `inspect --affected i18n --json` artık UI copy etkilerini app/page/generated file düzeyinde doğru raporluyor; `label app.title` gibi target'lar yanlışlıkla entity etkisi üretmiyor.
- Generated `App.tsx` runtime i18n varsa app title, auth app name, checking session, language selector, sidebar/topbar/nav/breadcrumb/no-page metinlerini `uiLabel(...)` üzerinden üretiyor.
- Generated page output runtime i18n varsa search/filter/columns/table headers/status badges/empty state/pagination/detail/form/modal/CRUD/custom action/workflow action metinlerini aynı UI label helper'ı üzerinden üretiyor.
- Generated frontend smoke test default locale UI copy'sini dikkate alacak şekilde güncellendi.
- Warehouse örneğine Türkçe/İngilizce UI label kataloğu eklendi; Product/Customer/Order page'leri, LowStock query page'i, generated Users/Audit page'leri, RestockProduct action'ı ve workflow transition'ları kapsandı.
- `docs/i18n.md`, `black docs i18n --json`, `black explain i18n --json`, `black explain label --json`, SPEC, BLACKLANG, README, AI agent contract, llms dosyaları ve website kaynakları güncellendi.
- Coverage matrisi `WEB-I18N-003` maddesini done yaptı; `frontend-ui-layout` score `98`, weighted web coverage signal `%70` kaldı.
- Warehouse benchmark güncellendi: 679 `.black` kaynak satırı, 563 code line, 45 generated dosya, 8690 generated satır, generated/source ratio `12.80`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestValidateI18NAndLabelTranslationErrors|TestValidateUILabelTranslationTargets|TestBuildWebUsesDefaultLocaleFieldLabels|TestAnalyzeAffectedI18N" -v` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs i18n --json` geçti.
- `go run ./cmd/black explain i18n --json` geçti.
- `go run ./cmd/black explain label --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected i18n --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP runtime app chrome/action/table/status copy kataloğudur; locale resource dosyası, pluralization veya date/number format dışında gelişmiş i18n kuralları değildir.
- UI label target'ları mevcut page/entity/action/workflow varlığına bağlı doğrulanır; typo'lar sessiz fallback'e bırakılmaz.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Explicit API Handler MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Explicit `api` deklarasyonlarını sadece çalışan acknowledgement route'u olmaktan çıkarıp typed request body doğrulayan ve tek güvenli stored-field update yapan deterministic backend intent seviyesine taşır. AI ajanı webhook veya dış sistem entegrasyonu için raw SQL/JS yazmadan `.black`, docs/explain, inspect/affected, OpenAPI metadata ve generated smoke testlerden davranışı okuyabilir.

Yapılanlar:
- Existing top-level `api Name { ... }` syntax'ı korundu; aynı davranış için ikinci handler DSL eklenmedi.
- `api` bloklarına `body`, `update` ve `respond` satırları eklendi.
- Body fields `POST`, `PUT` ve `PATCH` için desteklendi; `GET`/`DELETE` body handler kullanımını stable diagnostic ile reddediyor.
- Body modifiers canonical field/action input çizgisiyle sınırlı tutuldu: `required`, `optional`, `min`, `max`, `length`, `regex`, `url`, `message`.
- MVP handler şekli tek tutuldu: `update Entity where field == body.name set field = body.name`.
- Handler `where` alanı `id` veya stored `unique` field olmak zorunda; unbounded update reddediliyor.
- Handler set target'ları stored primitive non-policy field olmak zorunda; computed field, relation field, generated system field ve policy field update'i reddediliyor.
- Handler operand'ları `body.name`, `param.name`, source field, typed string/number/boolean literal ve basit numeric `+`, `-`, `*`, `/` ifadeleriyle sınırlandı.
- Division by zero için generated runtime guard eklendi.
- Public tenant-scoped handler için tenant policy field required body veya path param olarak zorunlu hale getirildi.
- Public owner-scoped handler reddediliyor; owner scope için private/auth yolu korunuyor.
- Auth+roles kullanan private handler'larda generated field permission ve audit log entegrasyonu eklendi.
- AST/JSON output `api.body`, `api.update`, `api.respond` alanlarını taşıyor.
- BlackIR explicit API output'u body/update/respond satırlarını koruyor; inspect IR `body N handler declared|update` özetini veriyor.
- Validator stable diagnostics üretiyor: `INVALID_API_BODY`, `UNSUPPORTED_API_BODY_METHOD`, `UNSUPPORTED_API_BODY_TYPE`, `UNSUPPORTED_API_BODY_MODIFIER`, `DUPLICATE_API_BODY`, `INVALID_API_UPDATE`, `UNSUPPORTED_API_HANDLER_METHOD`, `UNBOUNDED_API_UPDATE`, `MISSING_API_HANDLER_POLICY_SCOPE`, `UNSUPPORTED_API_HANDLER_POLICY_SCOPE`, `UNSUPPORTED_API_HANDLER_POLICY_FIELD`, `UNKNOWN_API_HANDLER_VALUE`, `API_HANDLER_VALUE_TYPE_MISMATCH`, `INVALID_API_RESPOND`, `UNSUPPORTED_API_RESPOND`.
- Generated Express server typed body parsing, bounded row lookup, policy scope, deterministic update ve `202/200` response üretiyor.
- Generated OpenAPI explicit API body schema, `x-blacklang-handler: update` ve `x-blacklang-update` metadata yayıyor.
- Generated contract test explicit API typed body schema ve update handler metadata kontrol ediyor.
- Generated API smoke test required body olmayan explicit routes için sample body, required body'li handler için 400 validation probe üretiyor; ayrıca gerçek SQLite runtime testinde webhook `202` ile `Product.stock` güncellendi.
- `docs api/openapi/generated-test`, `explain api`, diagnostics, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, roadmap ve website kaynakları güncellendi.
- Warehouse `StockWebhook` örneği typed body + bounded update handler ile güncellendi.
- Coverage matrisi `WEB-API-002` maddesini done yaptı; backend-api-actions score `78`, weighted web coverage signal `%71` oldu.
- Warehouse benchmark güncellendi: 684 `.black` kaynak satırı, 568 code line, 45 generated dosya, 8748 generated satır, generated/source ratio `12.79`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestParseExplicitAPIDeclaration|TestValidateExplicitAPI|TestBuildWebWritesExpectedFiles" -v` geçti.
- `go test -count=1 ./cmd/black -run "TestDocs|TestExplain|TestCoverage|TestParseExplicitAPIDeclaration|TestValidateExplicitAPI|TestBuildWebWritesExpectedFiles" -v` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `60`.
- `go run ./cmd/black docs api --json` geçti; body/update/respond syntax ve diagnostics JSON'da göründü.
- `go run ./cmd/black docs openapi --json` geçti; explicit API typed body/handler metadata notu göründü.
- `go run ./cmd/black docs generated-test --json` geçti; generated testlerin typed body/handler metadata kontrolü göründü.
- `go run ./cmd/black explain api --json` geçti; handler agent step'leri göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockWebhook --json` geçti; API'nin `Product.stock` yazdığı agent note olarak göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti; explicit API handler etkisi affected graph içinde göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --ir` geçti; `api StockWebhook ... body 3 handler update` göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%71`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; explicit API request validation gerçek HTTP server üzerinde probe edildi.
- Geçici SQLite runtime testinde `/api/webhooks/stock` `202 accepted` döndü ve seeded `Product.stock` değeri `1 -> 9` güncellendi.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP arbitrary backend service dili, raw SQL, provider-specific webhook signature verification, retry queue veya background job sistemi değildir.
- Handler tek entity üzerinde tek bounded update yapar; multi-step business flows için sıradaki mantıklı yön background jobs, seed/fixture/test syntax ve daha geniş service/transaction tasarımıdır.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Seed / Fixture Syntax MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Demo ve test verisini `.black` içinde deterministic source intent olarak tutmayı mümkün kılar. Generated web app artık schema setup sonrası aynı stable row key'lerle fixture verisini idempotent upsert edebilir; AI ajanı da seed değişikliğinin hangi entity, field ve generated dosyaları etkilediğini JSON/IR üzerinden okuyabilir.

Yapılanlar:
- Top-level `seed Name { source Entity ... }` declaration eklendi.
- Nested `row RowKey { ... }` syntax'ı eklendi; `RowKey` generated stable `id` olarak kullanılıyor.
- Scalar stored field değerleri için typed literal desteği eklendi: text/email/date/datetime string literal, numeric field'lar finite number literal, boolean field'lar `true`/`false`.
- Relation seed değerleri için tek canonical syntax eklendi: `relationField ref OtherRowKey`.
- Parser/AST JSON output `program.seeds[]` alanını taşıyor.
- Validator seed source, row key, duplicate row, duplicate field, stored/computed field ayrımı, scalar type, relation ref, required field, unique value, min/max/length/regex/url constraint ve secret-safe sınırı açısından stable diagnostics üretiyor.
- `docs seed --json`, `explain seed --json`, `inspect --affected SeedName --json`, `inspect --affected seed --json`, BlackIR ve affected graph desteği eklendi.
- Generated web output, seed varsa `src/seed.ts` dosyasını ve `package.json` içinde `db:seed` script'ini üretiyor.
- Generated `db:setup`, seed varsa schema setup sonrası `npm run db:seed` çalıştıracak şekilde bağlandı.
- SQLite ve PostgreSQL seed runtime output'u parameterized insert/upsert üretir; raw SQL `.black` içine alınmadı.
- Warehouse örneğine `DemoProducts`, `DemoCustomers` ve `DemoOrders` seed deklarasyonları eklendi.
- Warehouse relation seed örneğinde `Order.customer ref DemoCustomerAcme` değeri generated `customerId = "DemoCustomerAcme"` olarak uygulandı.
- `docs/seed.md`, diagnostics, AI agent contract, llms dosyaları, README, SPEC, BLACKLANG, roadmap, web yol haritası, sitemap ve site kaynakları güncellendi.
- Coverage matrisi `WEB-TEST-002` maddesini done yaptı; `WEB-TEST-003` browser e2e syntax için ayrı açık iş olarak ayrıldı.
- Coverage signal `%72` oldu; database-runtime score `82`, testing-benchmarks score `55`.
- Warehouse benchmark güncellendi: 728 `.black` kaynak satırı, 605 code line, 46 generated dosya, 8874 generated satır, generated/source ratio `12.19`.

Doğrulama:
- `gofmt -w cmd/black/seed.go cmd/black/generator_seed.go cmd/black/docs.go cmd/black/generator.go cmd/black/explain.go cmd/black/affected.go cmd/black/blackir.go cmd/black/parser.go cmd/black/validator.go cmd/black/ast.go cmd/black/query.go` geçti.
- `go test -count=1 ./cmd/black -run "Test.*Seed|TestDocs|TestExplain|TestCoverage|TestBuildWebWritesExpectedFiles" -v` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `61`.
- `go run ./cmd/black docs seed --json` geçti.
- `go run ./cmd/black docs generated-test --json` geçti; seed runtime wiring notu göründü.
- `go run ./cmd/black docs inspect --json` geçti; seed affected örnekleri göründü.
- `go run ./cmd/black explain seed --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected DemoProducts --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected DemoOrders --json` geçti; relation ref etkisi Customer entity olarak göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected seed --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.sku --json` geçti; `DemoProducts` seed etkisi göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --ir` geçti; `seeds 3` göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%72`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; output içinde `generated/src/seed.ts` kind `database-seed` göründü.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Geçici SQLite database ile generated app içinde `npm run db:setup` geçti; output `BlackLang seed data applied: 4 rows`.
- Geçici SQLite database okumasında `DemoProductLow`, `DemoProductFull`, `DemoCustomerAcme` ve `DemoOrderDraft` row'ları doğru yazıldı; `DemoOrderDraft.customerId` değeri `DemoCustomerAcme` olarak doğrulandı.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP browser e2e test dili, arbitrary fixture factory, raw SQL seed, random/generated fake data veya production data migration sistemi değildir.
- Seed runtime local/demo fixture verisi içindir; secret, token, API key veya production credential `.black` seed literal'larına yazılmamalıdır.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Browser Check Test Declaration MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web uygulamasının beklenen page/action/text yüzeyini `.black` içinde deterministic browser-check niyeti olarak tutmayı sağlar. AI ajanı artık “bu sayfa var mı, bu action expose ediliyor mu, bu generated UI metni bekleniyor mu?” sorularını generated dosyaları elle kurcalamadan source intent, docs/explain ve inspect/affected çıktılarıyla doğrulayabilir.

Yapılanlar:
- Top-level `test Name { ... }` declaration eklendi.
- Tek canonical MVP syntax belirlendi: `page PageName`, `expect text "..."`, `expect page PageName`, `expect action actionName`.
- Parser/AST JSON output `program.tests[]` alanını taşıyor.
- Validator test name, duplicate/collision, missing page, unknown page, missing expectation, duplicate expectation, empty/too-long text, unknown expected page/action ve unsupported expectation kind için stable diagnostics üretiyor.
- `docs test --json`, `explain test --json`, BlackIR, inspect IR ve affected graph desteği eklendi.
- `inspect --affected WarehouseBrowserSmoke --json` generated `src/blacklang.browser.test.tsx`, `src/App.tsx` ve `package.json` etkisini gösteriyor.
- `inspect --affected Products --json` Products page'i hedefleyen browser-check testlerini de listeliyor.
- Generated web output, source içinde test varsa `src/blacklang.browser.test.tsx` üretiyor.
- Generated `package.json` test script'i test declaration varsa `tsx src/blacklang.browser.test.tsx` adımını contract/API/frontend smoke testlerinin sonuna ekliyor.
- Generated browser-check smoke test React `App` render yüzeyini, generated page metadata'sını, page action listesini ve generated text catalog'u doğruluyor.
- Warehouse örneğine `WarehouseBrowserSmoke` eklendi: Products page, LowStock page ve RestockProduct action beklentileri doğrulanıyor.
- `docs/test.md`, `docs/generated-test.md`, diagnostics, AI agent contract, llms dosyaları, README, SPEC, BLACKLANG, roadmap, web yol haritası, sitemap ve site kaynakları güncellendi.
- Coverage matrisi `WEB-TEST-003` maddesini done yaptı; gerçek browser automation için `WEB-TEST-004` açık iş olarak ayrıldı.
- Coverage signal `%73` oldu; testing-benchmarks score `62`.
- Warehouse benchmark güncellendi: 735 `.black` kaynak satırı, 612 code line, 47 generated dosya, 9119 generated satır, generated/source ratio `12.41`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "Test.*Browser|TestFind.*Doc|TestAllDocs|TestExplain.*Keyword|TestWebCoverage|TestFormatCoverageIR|TestBuildWebGeneratesBrowserChecks" -v` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `62`.
- `go run ./cmd/black docs test --json` geçti.
- `go run ./cmd/black explain test --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected WarehouseBrowserSmoke --json` geçti; generated browser test, App ve package etkileri göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected test --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Products --json` geçti; `WarehouseBrowserSmoke` Products page etkisinde göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --ir` geçti; `tests 1` ve `test WarehouseBrowserSmoke page Products expectations 3` göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%73`, testing-benchmarks score `62`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; output içinde `generated/src/blacklang.browser.test.tsx` kind `test` göründü.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; `BlackLang generated browser checks passed` çıktısı alındı.
- Geçici SQLite database ile generated app içinde `npm run db:setup` geçti; seed data `products=2`, `customers=1`, `orders=1` olarak doğrulandı.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.

Notlar:
- Bu MVP tam browser e2e automation değildir; Playwright/browser binary, click, form submit, wait ve network step üretmez.
- `test` declaration deterministic generated page/action/text beklentileri içindir; arbitrary frontend event dili veya browser-side BlackLang runtime değildir.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## AI Task Benchmark and Token Estimate Reports MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang'in “AI ile web işi daha az dosya, daha az bağlam ve daha az tokenla tekrar edilebilir” iddiasını ölçülebilir planlama sinyaline dönüştürür. Komut AI modeli çağırmaz; mevcut source/generated ölçümlerinden deterministic scenario ve token estimate raporu üretir.

Yapılanlar:
- `black benchmark tasks [file] [--out <dir>] --json|--ir` alt komutu eklendi.
- Komut mevcut project config/source/out bilgisini kullanıyor, projeyi parse/validate ediyor ve existing `black benchmark` gibi generated output'u geçici dizinde ölçüyor.
- Configured generated output dizini mutate edilmiyor.
- JSON output `baseline`, `scenarios` ve `totals` alanlarını içeriyor.
- Baseline current source/generated file/line/byte ölçümlerini, generated/source line ratio değerini ve source/generated token-per-line katsayılarını taşıyor.
- İlk deterministic scenario seti 7 task içeriyor: `AI-TASK-QUERY-001`, `AI-TASK-ACTION-001`, `AI-TASK-SEED-001`, `AI-TASK-TEST-001`, `AI-TASK-POLICY-001`, `AI-TASK-API-001`, `AI-TASK-OPS-001`.
- Her scenario `purpose`, `trigger`, `projectEvidence`, `blacklangEdits`, `generatedImpact`, `commands` ve `estimatedTokens` alanlarını veriyor.
- Token estimate formülü deterministic tutuldu: current source/generated byte-line oranları, fixed scenario context line değerleri, fixed docs/context token değerleri ve command/edit text token estimate kullanılıyor.
- Estimates billed-token ölçümü olarak sunulmuyor; planning signal olarak etiketlendi.
- `black benchmark tasks --ir`, scenario ve toplam token farkını kompakt BlackIR olarak veriyor.
- `black agent startup --json`, `benchmark-tasks` komutunu startup command listesine ekliyor.
- `black docs benchmark --json`, `docs/benchmark.md`, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, roadmap, web yol haritası, Warehouse benchmark notu ve website kaynakları güncellendi.
- Coverage matrisi `WEB-TEST-001` maddesini done yaptı; testing-benchmarks score `70`, weighted web coverage signal `%74` oldu.
- Warehouse task benchmark sonucu: 7 scenario, estimated BlackLang tokens `9538`, estimated conventional tokens `92743`, estimated savings signal `%90`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestBenchmarkTasks|TestFormatAITaskBenchmarkIR" -v` geçti.
- `go test -count=1 ./cmd/black -run "TestAgent|TestFindBenchmarkDoc|TestBenchmarkTasks|TestFormatAITaskBenchmarkIR|TestWebCoverage|TestFormatCoverageIR" -v` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti; benchmark açıklaması source/generated size, AI task estimates ve web coverage olarak göründü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `62`.
- `go run ./cmd/black docs benchmark --json` geçti; syntax içinde `benchmark tasks` göründü.
- `go run ./cmd/black benchmark tasks ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black benchmark tasks ../../examples/warehouse/app.black --out ../../generated --ir` geçti; 7 scenario ve totals satırı göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%74`, testing-benchmarks score `70`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti; contract/API/frontend/browser-check zinciri temiz.

Notlar:
- Token estimate raporu gerçek model çalıştırması, billing ölçümü veya benchmark yarışması değildir; deterministic planning signal'dır.
- Scenario değerleri ileride gerçek AI eval koşumlarıyla kalibre edilebilir; bugünkü değerler current generated/source ölçümlerine bağlı sabit ve tekrar edilebilir katsayılardır.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Full Browser E2E Execution MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web uygulamasının `.black` içindeki deterministic `test` niyetini gerçek browser üzerinden doğrulamasını sağlar. Artık hızlı `npm test` browser-check metadata/render kontrolünü yaparken, explicit `npm run test:e2e` generated API server, Vite dev server, auth/register, seed data, navigation, visible text ve action button görünürlüğünü gerçek Chrome/Chromium/Edge DOM'u üzerinden koşar.

Yapılanlar:
- Mevcut top-level `test Name { ... }` syntax'ı korundu; yeni bir alternatif syntax eklenmedi.
- Source içinde test declaration varsa generator artık `src/blacklang.e2e.test.ts` üretir.
- Generated `package.json` içinde test declarations varken `test:e2e` ve `test:all` script'leri üretilir.
- `npm test` hızlı kaldı: contract/API/frontend/browser-check smoke zincirini çalıştırır.
- `npm run test:e2e`, `db:generate` sonrası generated E2E runtime'ını çalıştırır.
- Generated E2E runtime `playwright-core` ile local Chrome/Chromium/Edge executable bulur; `BLACKLANG_E2E_BROWSER_PATH` override desteklenir.
- SQLite target için default ephemeral `file:./blacklang-e2e.db` kullanılır; `BLACKLANG_E2E_DATABASE_URL` verilirse custom E2E database kullanılabilir.
- E2E runtime setup sırasını deterministic tuttu: env database URL set edilir, eski ephemeral DB temizlenir, `setup-db` import edilir, seed varsa `seed` import edilir, sonra server/Vite/browser başlatılır.
- CORS intent varsa generated E2E web origin'i CORS origins env değerine eklenir; Warehouse gibi CORS tanımlı projeler local E2E'de takılmaz.
- Auth varsa E2E deterministic bir kullanıcı register eder; sonra declared page/text/action beklentilerini gerçek DOM'da doğrular.
- Generated auth route, E2E sonunda SQLite bağlantısını kapatabilsin diye `closeAuthDatabase()` export eder; Prisma client da E2E cleanup sırasında disconnect edilir.
- Warehouse `WarehouseBrowserSmoke` beklentisi gerçek browser için stabil olan `expect text "Depo"` değerine taşındı.
- `inspect --affected WarehouseBrowserSmoke --json` ve `inspect --affected test --json` artık `src/blacklang.e2e.test.ts` etkisini gösterir.
- `docs test --json`, `docs generated-test --json`, `explain test --json`, local docs, README, SPEC, BLACKLANG, roadmap, Türkçe web yol haritası, website kaynakları, llms dosyaları ve benchmark notu güncellendi.
- Coverage matrisi `WEB-TEST-004` maddesini done yaptı; testing-benchmarks score `82`, weighted web coverage signal `%75` oldu.
- Warehouse benchmark güncellendi: 735 `.black` kaynak satırı, 48 generated dosya, 9480 generated satır, generated/source ratio `12.90`.
- Warehouse AI task benchmark sonucu güncellendi: 7 scenario, estimated BlackLang tokens `9551`, estimated conventional tokens `87915`, estimated savings signal `%89`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestBuildWebGeneratesBrowserChecks|TestBuildWebWritesExpectedFiles" -v` geçti.
- `go test -count=1 ./cmd/black -run "TestBuildWebGeneratesBrowserChecks|TestFindGeneratedTestDoc|TestFindBrowserTestDoc|TestExplainBrowserTestKeyword|TestWebCoverage|TestFormatCoverageIR|TestBenchmarkTasks" -v` geçti.
- `go test -count=1 ./...` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `62`.
- `go run ./cmd/black docs test --json` geçti; E2E dosyası ve `test:e2e` notları göründü.
- `go run ./cmd/black docs generated-test --json` geçti; `playwright-core` ve `test:e2e` notları göründü.
- `go run ./cmd/black explain test --json` geçti; `npm run test:e2e` agent step'i göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected WarehouseBrowserSmoke --json` geçti; `src/blacklang.browser.test.tsx`, `src/blacklang.e2e.test.ts` ve `package.json` etkileri göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected test --json` geçti; E2E test dosyası etkisi göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%75`, testing-benchmarks score `82`, `WEB-TEST-004=done`.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 48 generated file, 9480 generated line, ratio `12.90`.
- `go run ./cmd/black benchmark tasks ../../examples/warehouse/app.black --out ../../generated --json` geçti; totals `9551 / 87915 / %89`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 48 file üretildi ve `src/blacklang.e2e.test.ts` output'ta göründü.
- Generated app içinde `npm install` geçti; `playwright-core` dev dependency eklendi.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Generated app içinde `npm run test:e2e` geçti; API server/Vite/browser/auth/register/query/action akışı başarıyla koştu.
- Generated app içinde `npm run test:all` geçti.
- `test:e2e` ve `test:all` sonrası `blacklang-e2e.db` dosyasının geride kalmadığı doğrulandı.

Notlar:
- Bu feature arbitrary browser automation DSL değildir; `.black` içinde click/wait/form step dili eklenmedi.
- E2E, mevcut deterministic `test` declaration niyetini generated auth, setup/seed, navigation, visible text/page/action kontrolleriyle sınırlar.
- Full browser E2E için local Chrome/Chromium/Edge executable gerekir; otomatik path bulunamazsa `BLACKLANG_E2E_BROWSER_PATH` kullanılmalıdır.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Local Preview Deploy and Rollback Metadata MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Generated web uygulamasının production deploy niyetini Dockerfile/compose seviyesinden bir adım ileri taşır: aynı `.black` deploy bloğundan izole local preview stack, deploy manifest metadata'sı ve altyapıyı değiştirmeyen rollback plan çıktısı üretir. Bu sayede AI ajanı deploy etkisini tahmin etmek yerine `docs/explain/inspect/build` çıktılarından okuyabilir.

Yapılanlar:
- Mevcut top-level `deploy { ... }` bloğu korundu; yeni ayrı deploy syntax ailesi açılmadı.
- `preview local` satırı eklendi.
- `rollback keep <count>` satırı eklendi; keep aralığı `1..20` ile sınırlandı.
- AST JSON çıktısı `deploy.preview` ve `deploy.rollback` alanlarını taşır hale geldi.
- Parser yeni satırları stable diagnostic'lerle okur: `INVALID_DEPLOY_PREVIEW`, `DUPLICATE_DEPLOY_PREVIEW`, `INVALID_DEPLOY_ROLLBACK`, `DUPLICATE_DEPLOY_ROLLBACK`.
- Validator yeni semantic sınırları kontrol eder: `UNSUPPORTED_DEPLOY_PREVIEW`, `UNSUPPORTED_DEPLOY_ROLLBACK`, `INVALID_DEPLOY_ROLLBACK_KEEP`.
- BlackIR deploy çıktısı `preview local` ve `rollback keep N` satırlarını üretir.
- Generator `preview local` için `docker-compose.preview.yml` üretir.
- Preview compose, normal Docker Compose stack'ten ayrı `BLACKLANG_PREVIEW_PORT` default'u kullanır; Warehouse için `4001`.
- Preview compose SQLite için `file:/app/data/preview.db`, PostgreSQL için `blacklang_preview` database default'u kullanacak şekilde üretildi.
- Generator `rollback keep N` için `deploy/manifest.json`, `deploy/rollback.json` ve `scripts/rollback-plan.mjs` üretir.
- Generated `package.json` içinde `deploy:preview`, `deploy:preview:down` ve `deploy:rollback:plan` script'leri yalnızca ilgili deploy satırları varsa eklenir.
- Generated rollback plan script'i read-only çalışır; `releases/` dizinini varsa okur, retained release listesini ve rollback candidate değerini JSON olarak basar, altyapıyı değiştirmez.
- Generated `.env.example` içinde preview local varsa `BLACKLANG_PREVIEW_PORT` üretilir.
- `inspect --affected deploy --json` artık preview compose, deploy manifest, rollback JSON ve rollback script etkilerini gösterir.
- `docs deploy --json`, `explain deploy --json`, local docs, README, SPEC, BLACKLANG, roadmap, Türkçe web yol haritası, website kaynakları, diagnostics reference ve llms dosyaları güncellendi.
- Warehouse örneğine `preview local` ve `rollback keep 3` eklendi.
- Coverage matrisi deployment-operations score'unu `55` -> `70` yaptı; weighted web coverage signal `%75` -> `%77` oldu.
- `WEB-OPS-001` local preview + rollback metadata olarak done yapıldı; cloud adapter ve external observability için `WEB-OPS-003` açık bırakıldı.
- AI task benchmark'a `AI-TASK-DEPLOY-001` senaryosu eklendi.
- Warehouse benchmark güncellendi: 737 `.black` kaynak satırı, 52 generated dosya, 9610 generated satır, generated/source ratio `13.04`.
- Warehouse AI task benchmark sonucu güncellendi: 8 scenario, estimated BlackLang tokens `10499`, estimated conventional tokens `90872`, estimated savings signal `%88`.

Doğrulama:
- `go test -count=1 ./cmd/black -run "TestParseDeploy|TestValidateDeployErrors|TestBuildWebWritesExpectedFiles|TestAnalyzeAffectedDeploy|TestFormatBlackIR|TestFindDeployDoc|TestExplainDeployKeyword|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR" -v` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti ve golden manifest güncellendi.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `62`.
- `go run ./cmd/black docs deploy --json` geçti; `preview local`, `rollback keep`, yeni generated files ve diagnostic codes göründü.
- `go run ./cmd/black explain deploy --json` geçti; `deploy:rollback:plan` agent step'i göründü.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected deploy --json` geçti; yeni preview/rollback artifact etkileri göründü.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%77`, deployment-operations score `70`, `WEB-OPS-001=done`, `WEB-OPS-003=open`.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 52 generated file, 9610 generated line, ratio `13.04`.
- `go run ./cmd/black benchmark tasks ../../examples/warehouse/app.black --out ../../generated --json` geçti; 8 scenario ve totals `10499 / 90872 / %88`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; `docker-compose.preview.yml`, `deploy/manifest.json`, `deploy/rollback.json`, `scripts/rollback-plan.mjs` output'ta göründü.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Generated app içinde `npm run test:e2e` geçti.
- Generated app içinde `npm run deploy:rollback:plan` geçti ve read-only JSON plan bastı.
- E2E sonrası `generated/blacklang-e2e.db` ve journal dosyasının geride kalmadığı doğrulandı.

Notlar:
- Bu faz cloud/provider deploy adapter açmadı; o kapsam `WEB-OPS-003` olarak bilinçli şekilde açık kaldı.
- `rollback keep N` gerçek altyapıda symlink/container/image mutation yapmaz; sadece deterministic manifest ve read-only plan üretir.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Tracked Issue Export MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Web completion coverage içindeki açık işleri AI/CI/GitHub issue hazırlığı için ayrı, küçük ve deterministik JSON/IR çıktısı olarak verir. Böylece bir ajan ya da otomasyon büyük coverage matrisini parse etmeden toplam/açık/tamamlanan issue sayılarını ve açık takip listesini okuyabilir.

Yapılanlar:
- Tek canonical komut olarak `black benchmark issues --json|--ir` eklendi.
- Coverage matrisi authoritative kaynak olarak korundu; issue export coverage raporundan türetiliyor.
- JSON sonuç `summary.total`, `summary.open`, `summary.done`, `issues` ve `openIssues` alanlarını döndürüyor.
- IR sonuç kompakt `benchmark issues ok`, `completion`, `summary`, `openIssues` ve `issue ...` satırları üretiyor.
- `agent startup --json` içine `coverage-issues` komutu eklendi.
- `docs benchmark --json`, `docs coverage --json`, `explain coverage --json`, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, roadmap ve website kaynakları güncellendi.
- Coverage docs-ai-ergonomics score'u `72` -> `82` oldu.
- `WEB-DOCS-002` tracked issue export olarak done yapıldı.
- `WEB-DOCS-003` IDE diagnostics/autocomplete olarak açık bırakıldı.
- Weighted web coverage signal `%77` -> `%78` oldu.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs benchmark --json` geçti.
- `go run ./cmd/black docs coverage --json` geçti.
- `go run ./cmd/black explain coverage --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%78`, docs-ai-ergonomics score `82`.
- `go run ./cmd/black benchmark issues --json` geçti; total `27`, open `3`, done `24`.
- `go run ./cmd/black benchmark issues --ir` geçti; açık issue listesi `WEB-OPS-003`, `WEB-DOCS-003`, `WEB-ECO-001`.

Notlar:
- Bu faz gerçek GitHub issue açmadı; kullanıcı talimatı gereği dış sistemlere yazma final incelemeye bırakıldı.
- Export stable issue ID'leriyle çalışıyor; issue source hâlâ tek yerden, coverage matrisinden geliyor.
- Generated dosyalar elle düzenlenmedi.

## IDE Metadata, Completion, Snippet, and Diagnostics MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang bilinmeyen bir dil olduğu için editor, LSP bridge ve AI ajanların `.black` syntax'ını tahmin etmeden compiler-owned metadata'dan öğrenmesini sağlar. Completion items, canonical snippets, diagnostic code catalog ve editor-friendly diagnostic ranges tek deterministic CLI contract'ı üzerinden çıkar.

Yapılanlar:
- Yeni canonical IDE contract komutu eklendi: `black ide --json|--ir`.
- Yeni live diagnostic komutu eklendi: `black ide diagnostics [file] --json|--ir`.
- `black ide --json` artık language id, `.black`/`.blackthm` extension bilgisi, comment prefix, capabilities, command list, completion items, snippets ve diagnostic code catalog döndürüyor.
- Completion items; top-level keyword, block keyword, stored field type, computed field type, field modifier, comparison operator, view compose/display value, deploy preview value ve rollback strategy context'leriyle deterministic sıralanıyor.
- Snippets; app/target, auth, entity, computed, query, page, action, explicit API update, seed, test, deploy docker ve ops blokları için canonical source templates veriyor.
- Diagnostic catalog local `docs` entries içindeki stable error code listesinden türetiliyor.
- `black ide diagnostics <file> --json` format, parse, validate ve source-security diagnostics'i source'u değiştirmeden çalıştırıyor.
- IDE diagnostics LSP uyumlu zero-based, end-exclusive range döndürüyor.
- Okunabilir ama hatalı source için `success: true`, `valid: false` davranışı tanımlandı; command-level read/decrypt failures `success: false` kalıyor.
- `agent startup --json` içine `ide` ve `ide-diagnostics` komutları eklendi.
- `docs ide --json`, `explain ide --json`, local `docs/ide.md`, diagnostics reference, AI agent contract, README, SPEC, BLACKLANG, llms dosyaları, roadmap, benchmark docs, Türkçe web yol haritası ve website kaynakları güncellendi.
- Coverage docs-ai-ergonomics score'u `82` -> `92` oldu.
- `WEB-DOCS-003` IDE diagnostics/autocomplete olarak done yapıldı.
- `WEB-DOCS-004` IDE refactor workflows ve packaged editor extension için yeni açık takip başlığı oldu.
- Weighted web coverage signal `%78` -> `%79` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black -run "IDE|AgentStartup|FindIDEDoc|ExplainIDE|WebCoverage|FormatCoverage|CoverageIssues|FindBenchmarkDoc|FindCoverageDoc" -v` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; docs count `63`.
- `go run ./cmd/black docs ide --json` geçti.
- `go run ./cmd/black explain ide --json` geçti.
- `go run ./cmd/black ide --json` geçti; completion items `99`, snippets `12`, diagnostic codes `512`.
- `go run ./cmd/black ide --ir` geçti.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --json` geçti; `valid=true`, diagnostic total `0`.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --ir` geçti.
- `go run ./cmd/black agent startup ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%79`, docs-ai-ergonomics score `92`.
- `go run ./cmd/black benchmark issues --json` geçti; total `28`, open `3`, done `25`; açık issue listesi `WEB-OPS-003`, `WEB-DOCS-004`, `WEB-ECO-001`.
- `go run ./cmd/black benchmark issues --ir` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated app içinde `npm run build` geçti.
- Generated app içinde `npm test` geçti.
- Generated app içinde `npm run test:e2e` geçti.
- Generated app içinde `npm run deploy:rollback:plan` geçti.
- E2E sonrası `generated/blacklang-e2e.db` dosyasının geride kalmadığı doğrulandı.

Notlar:
- Bu faz packaged VS Code extension yayımlamadı; compiler-owned IDE contract, o extension'ın tüketmesi gereken authoritative data source oldu.
- Refactor/code-action tarafı `WEB-DOCS-004` olarak açık kaldı.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler üzerinden tekrar üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Packageable VS Code IDE Bridge MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`black ide` compiler contract'ını gerçek bir editör tüketim katmanına bağlar. VS Code bridge, BlackLang parsing/validation/snippet bilgisini JavaScript içinde yeniden yazmadan compiler'dan okur; `.black` yazarken syntax highlighting, autocomplete, snippet, diagnostics, format quick fix ve affected-symbol inspection akışı başlatır.

Yapılanlar:
- `editors/vscode-blacklang` altında packageable VS Code extension source eklendi.
- Extension manifest `.black` ve `.blackthm` language id/extension contribution'larını tanımlıyor.
- `.black` ve `.blackthm` için başlangıç TextMate grammar dosyaları eklendi.
- Extension runtime `black ide --json` çağırarak completion items ve snippets sağlar.
- Extension runtime `black ide diagnostics <file> --json` çağırarak VS Code DiagnosticCollection üretir.
- Diagnostic ranges zero-based/end-exclusive contract'a göre VS Code range'e çevrilir.
- `FORMAT_REQUIRED` için `BlackLang: Format Source` quick fix ve source fix-all action eklendi.
- `BlackLang: Refresh Diagnostics`, `BlackLang: Show IDE Manifest`, `BlackLang: Inspect Affected Symbol` komutları eklendi.
- `Inspect Affected Symbol`, cursor altındaki symbol için `black inspect <file> --affected <symbol> --json` çıktısını JSON document olarak açar.
- `blacklang.cliPath` ayarı eklendi; CLI PATH'te değilse editor içinden yol verilebilir.
- `editors/vscode-blacklang/scripts/validate-extension.mjs` eklendi; manifest, grammar JSON, extension source contract ve JS syntax check doğrulanıyor.
- `docs/ide.md`, release artifact docs, README, SPEC, BLACKLANG, AI agent contract, llms dosyaları, roadmap, Türkçe web yol haritası ve website IDE/coverage bölümleri güncellendi.
- Coverage docs-ai-ergonomics score'u `92` -> `98` oldu.
- `WEB-DOCS-004` IDE refactor workflows ve packaged editor extension olarak done yapıldı.
- Weighted web coverage signal `%79` kaldı; score artışı yuvarlama eşiğini değiştirmedi.

Doğrulama:
- `npm test` içinde `editors/vscode-blacklang` extension validator geçti.
- `go test -count=1 ./cmd/black -run "IDE|AgentStartup|FindIDEDoc|ExplainIDE|WebCoverage|FormatCoverage|CoverageIssues" -v` geçti.
- `gofmt -w ...` repo kökünden geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs ide --json` geçti.
- `go run ./cmd/black explain ide --json` geçti.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%79`, docs-ai-ergonomics score `98`.
- `go run ./cmd/black benchmark issues --json` geçti; total `28`, open `2`, done `26`; açık issue listesi `WEB-OPS-003`, `WEB-ECO-001`.

Notlar:
- Extension, BlackLang'i JS içinde parse etmiyor; CLI output'u authoritative kaynak olarak kullanıyor.
- Marketplace publish yapılmadı; kullanıcı talimatı gereği dış sistemlere yazma final incelemeye bırakıldı.
- Generated web output elle düzenlenmedi.

## Ecosystem Discovery and npm Wrapper MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang CLI'nin npm/npx üzerinden ince bir wrapper olarak kullanılabilmesini ve AI ajanların release, package, editor extension ve adapter sınırlarını `black ecosystem --json|--ir` ile deterministik keşfetmesini sağlar. Provider-specific genişleme core syntax'ı şişirmeden adapter/extension katmanında büyüyebilir.

Yapılanlar:
- `packages/npm` altında packageable npm wrapper source eklendi.
- Wrapper `blacklang` ve `black` bin komutlarını aynı native BlackLang CLI binary'sine forward ediyor.
- `BLACKLANG_BINARY` trusted binary override desteği eklendi.
- Postinstall davranışı güvenli tutuldu: vendored/trusted binary yoksa otomatik indirme yapmaz; GitHub Release indirmesi yalnızca `BLACKLANG_ALLOW_DOWNLOAD=1` ile çalışır.
- Platform mapping, release artifact URL üretimi, checksum parse/verify helper'ları ve package validator eklendi.
- `black ecosystem --json` ve `black ecosystem --ir` komutları eklendi.
- Ecosystem output release manifest/checksum bilgisini, npm wrapper'ı, VS Code bridge'i, built-in web/deploy/observability/editor adapter'larını ve trust policy notlarını raporlar.
- `docs ecosystem --json`, `explain ecosystem --json`, agent startup checklist, AI agent contract, README, SPEC, BLACKLANG, llms dosyaları, website ve npm/release docs güncellendi.
- Coverage ecosystem-integrations score'u `10` -> `45` oldu.
- `WEB-ECO-001` installable release artifacts ve plugin/adapter discovery olarak done yapıldı.
- `WEB-ECO-002` package registry publishing ve provider adapter marketplace için yeni açık takip başlığı oldu.
- Weighted web coverage signal `%79` -> `%81` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black -run "Ecosystem|AgentStartup|FindEcosystemDoc|ExplainEcosystem|WebCoverage|FormatCoverage|CoverageIssues" -v` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black ecosystem --ir` geçti.
- `go run ./cmd/black docs ecosystem --json` geçti.
- `go run ./cmd/black explain ecosystem --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%81`.
- `go run ./cmd/black benchmark issues --json` geçti; total `29`, open `2`, done `27`; açık issue listesi `WEB-OPS-003`, `WEB-ECO-002`.
- `npm test` içinde `packages/npm` wrapper validator geçti.
- `npm test` içinde `editors/vscode-blacklang` extension validator geçti.

Notlar:
- npm paketi yayımlanmadı; kullanıcı talimatı gereği dış sistemlere yazma final incelemeye bırakıldı.
- VS Code extension marketplace publish yapılmadı; packageable source ve validator yerelde hazırlandı.
- Generated web output elle düzenlenmedi.

## Cloud Adapter Manifest and Observability Hooks MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` içinde cloud deploy adapter planı ve external observability hook niyeti env referanslarıyla tarif edilir; generator bunu manifest, env örneği, read-only plan scripti, OpenAPI metadata'sı, runtime hook ve smoke test coverage'a taşır.

Yapılanlar:
- `deploy { cloud fly|render|railway app env NAME [region env NAME] }` syntax'ı parser, AST, validator, formatter ve JSON explain/docs contract'ına eklendi.
- `ops { observe webhook endpoint env NAME }` syntax'ı parser, AST, validator, formatter ve JSON explain/docs contract'ına eklendi.
- Cloud provider değerleri güvenli ve deterministik tutuldu: provider allowlist var, app/region değerleri sadece env referansı olarak taşınıyor.
- Generator `deploy/cloud.json`, `deploy/manifest.json` cloud metadata'sı ve `scripts/cloud-plan.mjs` read-only readiness planı üretir hale geldi.
- Generator `.env.example`, Docker compose env wiring ve package scriptlerini cloud/observe env referanslarıyla genişletti.
- Generated server, `observe webhook` tanımı olduğunda istekleri bozmayacak şekilde non-blocking structured event POST hook'u üretir hale geldi.
- Generated OpenAPI çıktısına `x-blacklang-observability` metadata'sı eklendi.
- Generated contract/API smoke testleri observability metadata'sını ve lokal webhook receiver ile delivery yolunu doğrular hale geldi.
- `inspect --affected deploy|cloud|ops|observe --json`, `black ecosystem --json`, IDE completions/snippets ve diagnostics referansları güncellendi.
- Warehouse örneği cloud deploy adapter ve observability hook niyetiyle güncellendi.
- Docs, SPEC, BLACKLANG, README, AI agent contract, llms dosyaları, roadmap, Türkçe web yol haritası ve website coverage/deploy/ops/ecosystem bölümleri güncellendi.
- Coverage deployment-operations score'u `70` -> `88` oldu.
- `WEB-OPS-003` cloud adapters and external observability provider hooks olarak done yapıldı.
- Weighted web coverage signal `%81` -> `%83` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black -run "Deploy|Ops|Ecosystem|IDE|Coverage|FormatBlackIR|AnalyzeAffected|BuildWebGeneratesCloud|BuildWebWritesExpectedFiles|BuildWebGeneratesOpsRuntime|FindDeployDoc|FindOpsDoc" -v` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs deploy --json` geçti.
- `go run ./cmd/black docs ops --json` geçti.
- `go run ./cmd/black explain deploy --json` geçti.
- `go run ./cmd/black explain ops --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected deploy --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected ops --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected cloud --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected observe --json` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%83`.
- `go run ./cmd/black benchmark issues --json` geçti; total `29`, open `1`, done `28`; açık issue listesi `WEB-ECO-002`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated `npm run build` geçti.
- Generated `npm test` geçti.
- Generated `npm run test:e2e` geçti.
- Generated `npm run deploy:rollback:plan` geçti.
- Generated `npm run deploy:cloud:plan` geçti; env yokken `ready: false` ve eksik env listesiyle read-only rapor üretti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve `TestWarehouseGeneratedGoldenManifest` geçti.
- `go test -count=1 ./...` CLI modül kökünden geçti.

Notlar:
- Real provider deploy execution yapılmadı; bu aşama bilinçli olarak read-only adapter manifest/plan sınırında kaldı.
- Observability endpoint değeri `.black` içine yazılmıyor; yalnızca env adı tutuluyor.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.
- Generated web output elle düzenlenmedi; compiler üzerinden yeniden üretildi.

## Package Registry and Provider Adapter Marketplace Manifests MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
BlackLang'in package, editor bridge ve provider adapter ekosistemini core syntax'ı büyütmeden stable manifestlerle keşfedilebilir hale getirir. AI ajanları `black ecosystem --json|--ir`, kısa docs/explain çıktıları ve lokal validator üzerinden hangi paket/adaptörün hangi trust kurallarıyla kullanılacağını okuyabilir.

Yapılanlar:
- `packages/registry/package-index.blackdir` eklendi; `npm:blacklang` ve `vscode:blacklang-vscode` stable package ID'leri, source path'leri, docs, commands ve trust flag'leri kaydedildi.
- `packages/registry/trust-policy.blackdir` eklendi; release manifest, checksum, native CLI forwarding, no language reimplementation ve download opt-in kuralları tanımlandı.
- `adapters/marketplace/adapter-index.blackdir` eklendi; target, deploy, preview, cloud deploy, observability ve editor adapter ID'leri stable marketplace manifestinde toplandı.
- `adapters/marketplace/trust-policy.blackdir` eklendi; env-only secrets, read-only plan before mutation, stable adapter ID, docs-before-enable, JSON/IR discovery ve generated tests rules tanımlandı.
- `packages/registry/scripts/validate-registry.mjs` eklendi; package registry ve provider adapter marketplace manifestlerinin beklenen stable marker'larını doğruluyor.
- `black ecosystem --json|--ir` çıktısı `registries`, `marketplaces` ve `trustWorkflow` alanlarıyla genişletildi.
- Human-readable `black ecosystem` çıktısı registry/marketplace satırlarını gösterir hale geldi.
- `docs package-registry --json`, `docs adapter-marketplace --json`, `explain package-registry --json` ve `explain adapter-marketplace --json` learning contract'ları eklendi.
- Agent startup, docs, explain related graph, README, SPEC, BLACKLANG, docs/llms, website/llms, roadmap, Türkçe web yol haritası ve website ecosystem/coverage bölümleri güncellendi.
- Coverage ecosystem-integrations score'u `45` -> `82` oldu.
- `WEB-ECO-002` prepared package registries and provider adapter marketplace olarak done yapıldı.
- Weighted web coverage signal `%83` -> `%85` oldu.

Doğrulama:
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `go test -count=1 ./cmd/black -run "Ecosystem|PackageRegistry|AdapterMarketplace|ExplainEcosystem|ExplainPackageRegistry|ExplainAdapterMarketplace|Coverage|AgentStartup|FindPackageRegistryDoc|FindAdapterMarketplaceDoc|FindEcosystemDoc" -v` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black ecosystem --ir` geçti.
- `go run ./cmd/black docs package-registry --json` geçti.
- `go run ./cmd/black docs adapter-marketplace --json` geçti.
- `go run ./cmd/black explain package-registry --json` geçti.
- `go run ./cmd/black explain adapter-marketplace --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count `66`.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%85`.
- `go run ./cmd/black benchmark issues --json` geçti; total `29`, open `0`, done `29`.
- `cd packages/npm && npm test` geçti.
- `cd editors/vscode-blacklang && npm test` geçti.
- CLI modül kökünden `go test -count=1 ./...` geçti.

Notlar:
- npm, VS Code Marketplace, GitHub Releases veya dış adapter registry yayını yapılmadı; kullanıcı talimatı gereği dış sistemlere yazma final incelemeye bırakıldı.
- Bu aşama gerçek provider CLI execution eklemez; provider mutation öncesi read-only manifest/trust hazırlığı sağlar.
- Generated web output elle düzenlenmedi.

## File and Image Media Fields MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` içinde `file` ve `image` stored field'ları deterministic biçimde tanımlanabilir. Generated web form seçilen dosyayı local data URL olarak okur, normal JSON API validation zincirinden geçirir, string kolon olarak saklar ve tablo/detay/form preview UI'da güvenli şekilde gösterir. Bu MVP provider upload, bucket, token veya private endpoint davranışını core dile sokmaz.

Yapılanlar:
- `file` ve `image` entity field type'ları compiler validator support listesine eklendi.
- `accept "MIME hint"` tek canonical media hint syntax'ı olarak eklendi; yalnızca entity `file`/`image` field'larında geçerli tutuldu.
- `image` field'ları için `accept` değeri `image/` içermiyorsa diagnostic üretilir; `accept` değeri yoksa generated input default `image/*` kullanır.
- Custom action media input'ları ve auth user media field'ları bu MVP'de açıkça unsupported diagnostic ile kapatıldı.
- Parser, AST/JSON output, validator, formatter, docs, explain, diagnostics, IDE metadata/snippet, seed validation, query literal validation ve inspect/affected akışı medya field'larını tanır hale getirildi.
- Generator Prisma/SQLite/PostgreSQL schema tarafında media field'larını string kolon olarak üretir hale geldi.
- Generated React form medya input'ları seçilen dosyayı `FileReader.readAsDataURL` ile okur ve form state'e yazar hale geldi.
- Generated table/detail/form preview UI `image` değerlerini thumbnail, `file` değerlerini link olarak render eder hale geldi.
- Generated frontend ve backend validation `image` için `data:image/*` veya absolute `http(s)`, `file` için `data:*` veya absolute `http(s)` değerleri kabul eder hale geldi.
- Generated Express JSON body limit'i entity media field varsa `2mb` olur hale geldi.
- OpenAPI schema output'una `x-blacklang-media`, `x-blacklang-encoding: data-url` ve image için `contentMediaType: image/*` metadata'sı eklendi.
- Warehouse örneğine `Product.photo image optional accept "image/*"` eklendi ve generated output compiler üzerinden yeniden üretildi.
- `docs/media.md`, README, BLACKLANG, SPEC, AI agent contract, docs/llms, website/llms, roadmap, Türkçe web yol haritası, website ana sayfası ve sitemap medya desteğiyle güncellendi.
- Coverage `data-model-crud` score'u `85` -> `90` oldu; weighted web coverage signal yuvarlama nedeniyle `%85` kaldı.
- `black docs --all --json` doc count `67` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black -run "Media|Seed|BuildWebWritesExpectedFiles|IDESupport|Coverage|ParseMedia|ValidateMedia|ExplainMedia|FindMedia|ExplicitAPI" -v` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs media --json` geçti.
- `go run ./cmd/black explain media --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count `67`.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.photo --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%85`, `data-model-crud` score `90`.
- `go run ./cmd/black benchmark issues --json` geçti; total `29`, open `0`, done `29`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated `npm run build` geçti.
- Generated `npm test` geçti.
- Generated `npm run test:e2e` geçti.
- Generated `npm run deploy:rollback:plan` geçti.
- Generated `npm run deploy:cloud:plan` geçti; env yokken read-only `ready: false` raporu üretti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve `TestWarehouseGeneratedGoldenManifest` geçti.
- CLI modül kökünden `go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `cd packages/npm && npm test` geçti.
- `cd editors/vscode-blacklang && npm test` geçti.

Notlar:
- External media storage, image processing, signed download permission ve provider upload adapter'ları bilinçli olarak core MVP dışında bırakıldı.
- Provider secrets, bucket isimleri, upload token'ları veya private endpoint'ler `.black` içine yazılmadı.
- Generated web output elle düzenlenmedi; compiler üzerinden yeniden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Background Query Jobs MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` içinde public endpoint açmadan, declared query'leri generated worker içinde deterministik schedule ile çalıştırmayı sağlar. İlk sürüm read-only query worker olarak kalır; queue provider, retry, delayed task, mutating job, external call ve provider scheduler core dile eklenmez.

Yapılanlar:
- Top-level `job Name { ... }` syntax'ı eklendi.
- Tek canonical schedule syntax'ı `schedule every <integer> minutes|hours|days` olarak belirlendi.
- Tek MVP run mode `run query QueryName` olarak eklendi.
- Parser, AST/JSON output, validator, BlackIR, formatter-aware inspect output, diagnostics, docs, explain, IDE completion/snippet ve inspect/affected akışı job deklarasyonlarını tanır hale getirildi.
- Validator job adı, duplicate job, top-level sembol çakışması, duplicate/missing schedule, unsupported unit, bounded interval, duplicate/missing run ve unknown query durumlarını stable diagnostic kodlarıyla raporlar hale geldi.
- Generator `jobs/manifest.json` ve `src/worker.ts` üretir hale geldi.
- Generated worker `blackJobs`, `runJob(name)`, `runAllJobsOnce()`, `--once`, `--loop` ve opsiyonel `--job Name` desteği üretir hale geldi.
- Generated worker query'nin stored-field `where`, deterministic `sort` ve `limit` kurallarını uygular, yalnızca ID seçer ve job adı/schedule/query/source/count/limit/timestamp içeren kompakt JSON log üretir.
- Generated `package.json` içine `jobs:run` ve `jobs:loop` script'leri eklendi.
- Generated OpenAPI root metadata'sına `x-blacklang-jobs` eklendi.
- Generated contract test job metadata'sını doğrular hale getirildi.
- Warehouse örneğine `LowStockMonitor` eklendi ve `LowStockProducts` query'sine bağlandı.
- `docs/job.md`, README, BLACKLANG, SPEC, AI agent contract, docs/llms, website/llms, diagnostics, generated-test/api/query referansları, ROADMAP, ROADMAP-v0.2, Türkçe web yol haritası, website ana sayfası ve sitemap job desteğiyle güncellendi.
- Coverage `backend-api-actions` score'u `78` -> `86` oldu.
- Weighted web coverage signal `%85` -> `%86` oldu.
- `WEB-API-004` deterministic background query worker jobs olarak done yapıldı.
- `black docs --all --json` doc count `68` oldu.
- `black benchmark issues --json` total `30`, open `0`, done `30` oldu.

Doğrulama:
- `go test -count=1 ./cmd/black -run "Job|BuildWebWritesExpectedFiles|IDESupport|Coverage|FormatCoverage|AnalyzeAffected|FindJob|ExplainJob|ParseJob|ValidateJob" -v` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs job --json` geçti.
- `go run ./cmd/black explain job --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count `68`.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected LowStockMonitor --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected LowStockProducts --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%86`, `backend-api-actions` score `86`.
- `go run ./cmd/black benchmark issues --json` geçti; total `30`, open `0`, done `30`.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated `npm run build` geçti.
- Generated `npm test` geçti.
- Generated `npm run test:e2e` geçti.
- Generated `npm run jobs:run` temiz test database URL'iyle geçti ve `LowStockMonitor` için compact JSON log üretti.
- Generated `npm run deploy:rollback:plan` geçti.
- Generated `npm run deploy:cloud:plan` geçti; env yokken read-only `ready: false` raporu üretti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve `TestWarehouseGeneratedGoldenManifest` geçti.
- CLI modül kökünden `go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `cd packages/npm && npm test` geçti.
- `cd editors/vscode-blacklang && npm test` geçti.

Notlar:
- Varsayılan generated `dev.db` bu workspace'te eski şemadan kalan lokal bir dosya olduğu için ilk çıplak `npm run jobs:run` denemesi `setup-db.ts` sırasında `tenantId` kolonuna takıldı. Kullanıcı verisini silmemek için `dev.db` temizlenmedi; worker doğrulaması yeni bir `DATABASE_URL=file:./blacklang-job-smoke-*.db` ile yapıldı.
- Job MVP public HTTP endpoint üretmez; route yüzeyi yerine generated worker, manifest, package script ve OpenAPI metadata üretir.
- Provider scheduler, queue/retry, delayed task, mutating job ve external integration davranışları sonraki adapter/runtime aşamalarına bırakıldı.
- Generated web output elle düzenlenmedi; compiler üzerinden yeniden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## API-only Target MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
`.black` kaynağından React/Vite UI üretmeden yalnızca deterministic Node/Express API, Prisma runtime, OpenAPI sözleşmesi, doğrulama, API testleri, job runner ve deploy metadata çıktısı alınmasını sağlar. Böylece BlackLang web tarafı sadece panel üreten bir araç değil, gerektiğinde servis/API projesi üreten bir hedef haline gelir.

Yapılanlar:
- `target api { backend node; database sqlite|postgres }` syntax'ı eklendi.
- `target web` için `frontend react` zorunluluğu korunurken, `target api` için frontend satırı yasaklandı.
- `target api` altında top-level browser `test` blokları `UNSUPPORTED_API_TARGET_TEST` diagnostic'iyle reddedilir hale getirildi.
- Parser duplicate/invalid target önerileri web ve API target örnekleriyle güncellendi.
- Generator API-only target'ta `index.html`, Vite config, React entry/app/page/component output'ları, frontend smoke testi, browser-check testi ve e2e browser testini üretmez hale getirildi.
- API-only output yine README, `.env.example`, package, Docker/deploy dosyaları, Prisma schema/config/client/setup/seed, OpenAPI, Express server, validation, route, API client, contract/API tests, job manifest ve worker üretir.
- API-only generated `package.json` script'leri React/Vite bağımlılığı olmadan `tsx src/server.ts`, `npm run db:generate && tsc`, contract/API testleri ve job/deploy scriptleriyle üretildi.
- Generated `/openapi.json` route'u `sendFile(process.cwd())` bağımlılığından çıkarıldı; server artık OpenAPI sözleşmesini gömülü deterministic JSON nesnesi olarak döndürüyor. Bu, Windows/Türkçe karakterli path etkisini de ortadan kaldırdı.
- Generated package script JSON değerleri `strconv.Quote` ile güvenli basılır hale getirildi; web target build script'indeki iç tırnaklar artık `package.json` formatını bozmuyor.
- `inspect --affected target --json` API-only target için frontend/Vite/browser dosyalarını dışarıda bırakır hale getirildi.
- `inspect --affected <page> --json` API-only target'ta page kaynaklı React page/App etkisini dışarıda bırakır hale getirildi.
- `docs target --json`, `docs generated-test --json`, `explain target --json`, diagnostics reference ve IDE metadata API-only davranışıyla güncellendi.
- IDE completion/snippet listesine `target api` ve `app-api` starter snippet'i eklendi.
- README, BLACKLANG, SPEC, AI agent contract, docs index/llms, website llms, live site source, ROADMAP, ROADMAP-v0.2 ve Türkçe web yol haritası API-only hedefiyle güncellendi.
- Coverage `backend-api-actions` score'u `86` -> `92` oldu.
- Weighted web coverage signal `%86` -> `%87` oldu.
- `WEB-API-005` API-only target generation olarak done yapıldı.

Doğrulama:
- `go test -count=1 ./cmd/black -run "BuildWebWritesExpectedFiles|APIOnly|Target" -v` geçti.
- `go test -count=1 ./cmd/black -run "BuildWebWritesExpectedFiles|APIOnly|BrowserCheck|GeneratedTest" -v` geçti.
- `go test -count=1 ./cmd/black -run "Target|APIOnly|Coverage|IDESupport|Docs|ExplainTarget|GeneratedTest|ValidateTarget|ParseTarget|AnalyzeAffected|BuildWebWritesExpectedFiles" -v` geçti.
- API-only smoke source format/lint/validate/build akışı geçti.
- API-only generated `npm run build` geçti.
- API-only generated `npm test` geçti; `/openapi.json` artık `200` döndü.
- API-only generated `npm run jobs:run` geçti.
- API-only generated `npm run deploy:rollback:plan` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs target --json` geçti.
- `go run ./cmd/black explain target --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completion `%87`.
- `go run ./cmd/black benchmark issues --json` geçti; completion `%87`.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run jobs:run` temiz test database URL'iyle geçti ve `LowStockMonitor` için compact JSON log üretti.
- Generated Warehouse `npm run deploy:rollback:plan` geçti.
- Generated Warehouse `npm run deploy:cloud:plan` geçti; env yokken read-only `ready: false` raporu üretti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve `TestWarehouseGeneratedGoldenManifest` geçti.
- CLI modül kökünden `go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `cd packages/npm && npm test` geçti.
- `cd editors/vscode-blacklang && npm test` geçti.

Notlar:
- API-only MVP yeni bir davranış için ikinci syntax üretmedi; mevcut `target` bloğunu deterministic şekilde genişletti.
- Browser testleri API-only hedefte desteklenmez; validator bunu erken ve JSON diagnostic ile raporlar.
- Generated OpenAPI route'unun path bağımsız hale getirilmesi web target için de daha sağlam runtime davranışı sağladı.
- Varsayılan generated `dev.db` bu workspace'te eski şemadan kalan lokal bir dosya olduğu için çıplak Warehouse `npm run jobs:run` denemesi `tenantId` kolonuna takıldı. Kullanıcı verisini silmemek için `dev.db` temizlenmedi; worker doğrulaması yeni `DATABASE_URL=file:./blacklang-job-smoke-api-target.db` ile yapıldı.
- Generated output elle düzenlenmedi; compiler üzerinden yeniden üretildi.
- Kullanıcı talimatı gereği commit/push ve VPS yayını yapılmadı; toplu inceleme sonrası yapılacak.

## Multi-role Auth Runtime MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated auth runtime artık bir kullanıcıya birden fazla rol atanmasını destekler. `.black` tarafında tek canonical syntax olarak mevcut `role` blokları korunur; generated runtime `roles: string[]` setiyle karar verir, eski `role` alanını da primary/backward-compatible değer olarak tutar.

Yapılanlar:

- Generated `BlackUser` modeli, SQLite setup/migration ve PostgreSQL Prisma şeması `roles` setini saklayacak şekilde genişletildi.
- `/api/auth/me` ve public user shape artık `role` yanında `roles` döndürür.
- Generated Users ekranı tekil role seçimi yerine bir kullanıcıya bir veya birden fazla rol atayan checkbox tabanlı yönetim üretir.
- Page access, action allow, field permission, custom query, custom action ve workflow guard kontrolleri kullanıcının tüm rol setini değerlendirir.
- Permission modeli deterministik tutuldu: atanmış rollerden herhangi birindeki explicit `deny` eşleşen `allow` kararını ezer; aksi halde herhangi bir roldeki `allow` erişim verir.
- Generated OpenAPI `AuthUser.roles` ve `AuthRoleUpdateInput.roles` array sözleşmesini yayınlar; generated contract tests bu sözleşmeyi kontrol eder.
- API-only target ile React/browser çıktısı olmayan generated servislerde multi-role değişikliği geriye dönük kırılma yaratmayacak şekilde doğrulandı.
- Coverage taxonomy güncellendi: `auth-permissions-security` score 70 -> 78, `WEB-SEC-002` done, weighted web coverage 87 -> 88.
- `README.md`, `SPEC.md`, `BLACKLANG.md`, `docs/ai-agent-contract.md`, `docs/benchmark.md`, `website/index.html`, `website/llms.txt` ve `blacklang-web-yol-haritasi.md` multi-role davranışını anlatacak şekilde güncellendi.

Doğrulama:

- `go test -count=1 ./cmd/black -run "BuildWebWritesExpectedFiles|Auth|Policy|Permission|Action|Query" -v` geçti.
- Warehouse generated çıktı: `npm run build`, `npm test`, `npm run test:e2e`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan` geçti.
- API-only smoke generated çıktı: `npm run build`, `npm test`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan` geçti.
- `go run ./cmd/black benchmark coverage --json` artık `completionPercent: 88` döndürüyor.

Notlar:

- Generated dosyalar elle düzenlenmedi; web/API çıktısı generator üzerinden yeniden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası yapılacak.
- Açık kalan security gap'leri: tenant administration UI, secret manager integrations, signed compiler/release verification.

## Tenant Administration UI MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

`policy tenant` kullanan generated uygulamalarda ilk tanımlı rol artık kullanıcıların tenant scope'unu yönetebilir. Böylece tenant row policy sadece runtime filtreleme değil, generated admin ekranından değiştirilebilir bir kullanıcı yönetimi davranışı olur.

Yapılanlar:

- Generated Users sayfası tenant policy varsa `tenantId` editörü üretir; tenant policy yoksa görünür tenant yönetimi çıkmaz.
- SQLite ve PostgreSQL auth runtime'larına `PUT /api/auth/users/:id/tenant` endpoint'i eklendi.
- Tenant update route'u `requireAuth`, CSRF ve first-role admin page access guard arkasında çalışır.
- Tenant ID değeri deterministic ve güvenli tutuldu: boş değer, 64 karakter üstü değer, harf/rakam/underscore/dash dışı karakterler reddedilir.
- Tenant güncellemesi audit log'a `tenant.update` olarak yazılır.
- OpenAPI'ye `/api/auth/users/{id}/tenant` path'i ve `AuthTenantUpdateInput` schema'sı eklendi.
- Generated contract tests tenant policy varsa tenant update path/schema sözleşmesini doğrular.
- `black docs policy --json`, `black explain policy --json`, SPEC, BLACKLANG, README, policy guide, llms dosyaları, roadmap ve site coverage bilgisi güncellendi.
- Coverage taxonomy güncellendi: `auth-permissions-security` score 78 -> 82, `WEB-SEC-003` done, weighted web coverage 88 -> 89.

Doğrulama:

- `go test -count=1 ./cmd/black -run "BuildWebWritesExpectedFiles|Policy|Auth|Contract|Tenant" -v` geçti.
- Warehouse generated çıktı: `npm run build`, `npm test`, `npm run test:e2e`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan` geçti.
- API-only smoke generated çıktı: `npm run build`, `npm test`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan` geçti.
- `go test -count=1 ./...` geçti.
- `black format --check --json`, `black lint --json`, `black validate --json`, `black docs --all --json`, `black explain policy --json` geçti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve doğrulandı.
- `go run ./cmd/black benchmark coverage --json` artık `completionPercent: 89` döndürüyor.

Notlar:

- Yeni `.black` syntax eklenmedi; mevcut `policy tenant` ve `role` sistemi üzerinden deterministic runtime tamamlandı.
- Generated dosyalar elle düzenlenmedi; output generator üzerinden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası yapılacak.
- Açık kalan security gap'leri: secret manager integrations ve signed compiler/release verification.

## Secret Reference Manifest MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated web/API uygulamaları artık secret değerlerini kaynak dosyaya veya manifestlere yazmadan, hangi environment referanslarının gerektiğini, hangilerinin hassas olduğunu ve deploy öncesi neyin eksik kaldığını makine-okunur şekilde raporlar. Bu, BlackLang'in deterministik ve AI-okunur deploy güvenliği çizgisini güçlendirir.

Yapılanlar:

- Generator `security/secrets.json` manifesti üretir; referans adı, kaynakları, türü, required/optional durumu ve sensitive bayrağı değer basmadan listelenir.
- Generator `scripts/secrets-plan.mjs` read-only kontrol script'i üretir; `npm run security:secrets:plan` env var varlığını doğrular ama secret değerlerini yazdırmaz.
- Generated `package.json`, README ve deploy file listesi secret plan akışını içerir.
- `DATABASE_URL`, deploy env referansları, cloud app/region, preview/runtime portları, CORS origin listesi ve observability endpoint referansları deterministic şekilde sınıflandırıldı.
- Aynı env adı birden fazla kaynaktan gelirse syntax çoğaltmadan tek referans altında sources listesiyle birleştirildi; semantic kind ilk güvenli sınıflandırmadan korunur.
- Generated contract tests `security/secrets.json` politikasını ve database URL hassasiyetini doğrular.
- `docs/secret-management.md`, README, docs index, SPEC, BLACKLANG, deployment docs, AI agent contract, llms dosyaları, roadmap, coverage benchmark ve canlı site kaynağı güncellendi.
- Coverage taxonomy güncellendi: `auth-permissions-security` score 82 -> 86, `WEB-SEC-004` done, weighted web coverage 89 -> 90.

Doğrulama:

- Warehouse generated çıktı: `npm run build`, `npm test`, `npm run test:e2e`, `npm run security:secrets:plan`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan` geçti.
- API-only smoke generated çıktı: `npm run build`, `npm test`, `npm run security:secrets:plan`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan` geçti.
- Warehouse golden manifest generator çıktısıyla güncellendi ve `TestWarehouseGeneratedGoldenManifest` geçti.
- CLI modül kökünden `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain entity --json` ve `go run ./cmd/black explain deploy --json` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `cd packages/npm && npm test` geçti.
- `cd editors/vscode-blacklang && npm test` geçti.
- `go run ./cmd/black benchmark coverage --json` artık `completionPercent: 90` döndürüyor.

Notlar:

- Secret değerleri `.black` içine, generated manifestlere veya test loglarına yazılmadı.
- Yeni `.black` syntax eklenmedi; mevcut deploy/database/security/ops referansları manifest olarak dışa verildi.
- Generated dosyalar elle düzenlenmedi; output generator üzerinden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası yapılacak.
- Açık kalan security gap: signed compiler/release verification.

## Signed Release Trust MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang public install ve package/adapter publish yolları artık sadece checksum'a değil, deterministic signed release trust contract'ına bağlanır. AI ajanı veya CI, private key görmeden `release.blackdir`, `checksums.sha256`, detached Ed25519 imzalar ve trusted public key durumunu JSON/IR ile doğrulayabilir.

Yapılanlar:

- Root seviyeye read-only `scripts/verify-release-trust.mjs` eklendi.
- Verifier `release.blackdir`, `checksums.sha256`, archive byte hash'i, `<artifact>.sig` detached Ed25519 imzası ve `BLACKLANG_RELEASE_PUBLIC_KEY` / `BLACKLANG_RELEASE_PUBLIC_KEY_FILE` public key referansını kontrol eder.
- Strict mod `ready:true` değilse non-zero exit verir; non-strict mod release hazırlığı sırasında eksikleri JSON olarak raporlar.
- `black ecosystem --json|--ir` çıktısına `release.trust` eklendi: algorithm, signature file pattern, public key env adları, verify command ve publish öncesi required listesi görünür oldu.
- Package registry ve adapter marketplace trust policy manifestleri signed release/package verification marker'larıyla genişletildi.
- `node packages/registry/scripts/validate-registry.mjs` yeni trust marker'larını ve root verifier script'ini doğrular.
- Npm wrapper download akışı checksum doğrulamasından sonra detached signature doğrulamadan release archive extract etmez.
- `packages/npm/lib/signature.mjs` eklendi; wrapper validation geçici Ed25519 keypair ile positive signature verification test eder.
- `black docs release-trust --json` ve `black explain release-trust --json` eklendi; agent guidance public install/publish güven koşullarını açıklar.
- README, SPEC, BLACKLANG, install/release/npm/ecosystem/package-registry/adapter-marketplace docs, llms dosyaları, roadmap ve site kaynağı güncellendi.
- Coverage taxonomy güncellendi: `auth-permissions-security` score 86 -> 90, `ecosystem-integrations` score 82 -> 90, `WEB-SEC-005` ve `WEB-ECO-003` done, weighted web coverage 90 -> 91.

Doğrulama:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black benchmark coverage --json` `completionPercent: 91` döndürdü.
- `go run ./cmd/black ecosystem --ir` release trust ve verify command satırlarını üretti.
- `go run ./cmd/black docs release-trust --json` geçti.
- `go run ./cmd/black explain release-trust --json` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `node --check scripts/verify-release-trust.mjs` geçti.
- Mevcut dev release klasörü için non-strict verifier `ready:false` ve eksik `.sig`/public key listesini değer sızdırmadan raporladı.
- Geçici imzalı test release klasörü için strict verifier `ready:true`, `signatureVerified:true`, `missing: []`, `errors: []` döndürdü.
- `cd packages/npm && npm test` geçti ve wrapper signature helper'ı geçici Ed25519 keypair ile doğrulandı.

Notlar:

- Bu faz private signing key üretmez veya saklamaz; signing release-owner/CI sorumluluğudur.
- Yeni `.black` syntax eklenmedi; release trust compiler/ecosystem tooling contract'ı olarak kaldı.
- Public GitHub Release, npm, editor marketplace veya provider marketplace yayını yapılmadı.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası yapılacak.
- Açık kalan security gap'leri: release transparency log, key rotation policy ve secret manager provider execution.

## General Transaction Block MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Custom action ve explicit API update handler'ları artık kaynak içinde top-level `transaction Name { ... }` bloğuna bağlanarak tek deterministik atomic runtime niyetiyle çalıştırılabilir. Transaction bloğu endpoint veya mutation tanımlamaz; var olan action/API update davranışının generated Prisma transaction sınırını ve OpenAPI metadata'sını belirler.

Yapılanlar:

- Parser ve AST top-level `transaction` declaration desteği kazandı; hedef satırları `action ActionName` ve `api APIName` olarak tek canonical syntax ile tutuldu.
- Eski action-level `transaction` satırı canonical syntax olmaktan çıkarıldı ve `UNSUPPORTED_ACTION_TRANSACTION` diagnostic'iyle top-level block önerisine yönlendirildi.
- Validator transaction blokları için PascalCase, duplicate, boş blok, bilinmeyen action/API, update handler'ı olmayan API ve aynı target'ın birden fazla transaction'a bağlanması kontrollerini ekledi.
- Custom action generated route'ları transaction hedefini artık action içi bayraktan değil top-level transaction binding'inden okuyor.
- Explicit API update handler'ları transaction'a bağlandığında bounded lookup, deterministic update ve private auth audit yazımı `prisma.$transaction` içinde üretiliyor.
- Generated OpenAPI ve contract tests targeted action/API route'larında `x-blacklang-transaction` ve `x-blacklang-transaction-name` metadata'sını doğruluyor.
- `inspect --affected` transaction symbol'larını, hedef action/API ilişkilerini ve etkilenen generated dosyaları döndürüyor.
- `black docs transaction --json`, `black explain transaction --json`, diagnostics, IDE completion/snippet ve BlackIR/inspect IR çıktıları eklendi.
- Warehouse örneğinde `RestockProduct` action'ı ve `StockWebhook` API'si `transaction RestockAtomic` bloğuna bağlandı.
- `docs/transaction.md`, `docs/action.md`, `docs/api.md`, SPEC, BLACKLANG, AI agent contract, llms dosyaları, roadmap, coverage benchmark ve site kaynağı yeni syntax'a göre güncellendi.
- Coverage taxonomy güncellendi: `backend-api-actions` score 92 -> 94, `database-runtime` score 90 -> 94, `WEB-DATA-005` done, weighted web coverage 91 -> 92.

Doğrulama:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs transaction --json` geçti.
- `go run ./cmd/black explain transaction --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockAtomic --json` geçti ve action/API/transaction affected grafiğini döndürdü.
- `go run ./cmd/black benchmark coverage --json` artık `completionPercent: 92` döndürüyor.
- Warehouse source: `format --check --json`, `lint --json`, `validate --json`, `build --out ../../generated --json` geçti.
- Warehouse generated çıktı: `npm run build`, `npm test`, `npm run test:e2e`, `npm run security:secrets:plan`, fresh SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan` geçti.
- API-only transaction smoke: temp `target api` source format/lint/validate/inspect geçti; generated çıktı `npm run build` ve `npm test` geçti.
- Eski action-level transaction dokümantasyon kalıntıları için repo taraması yapıldı; `INVALID_ACTION_TRANSACTION`, `DUPLICATE_ACTION_TRANSACTION`, nested/general transaction block phrasing kalmadı.

Notlar:

- Tek davranış için iki syntax bırakılmadı; atomic runtime niyeti sadece top-level `transaction` bloğuyla ifade ediliyor.
- Transaction blokları secret veya external provider bilgisi taşımaz.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden yeniden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan backend/database gap'leri: service/module APIs beyond page resources, online migration runners ve advanced database providers.

## Service Blocks MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Explicit API declarations artık top-level `service Name { api APIName }` bloğuyla deterministic bir servis/modül sınırına bağlanabilir. Service block endpoint, handler veya mutation yaratmaz; var olan explicit API'leri generated service manifest/module, OpenAPI tag metadata'sı ve compact AI affected grafiği altında gruplar.

Yapılanlar:

- Parser ve AST top-level `service` declaration desteği kazandı; tek canonical target satırı `api APIName` olarak tutuldu.
- Validator service blokları için PascalCase, duplicate, boş blok, bilinmeyen API, duplicate target ve aynı API'nin birden fazla service'e bağlanması kontrollerini ekledi.
- Generated web output `services/manifest.json` ve `src/services/<service>.ts` service module dosyalarını üretiyor.
- OpenAPI service-bound explicit API operation'larına `tags` ve `x-blacklang-service`, root spec'e `x-blacklang-services` metadata'sı ekleniyor.
- Generated contract tests service manifest, module-facing metadata, OpenAPI root service listesi, tags ve route-level service metadata'sını doğruluyor.
- `inspect --affected` service symbol'larını, bağlı explicit API'leri, transaction ilişkilerini, update entity etkilerini ve generated dosyaları JSON/BlackIR içinde döndürüyor.
- `black docs service --json`, `black explain service --json`, diagnostics, IDE completion/snippet ve BlackIR/inspect IR çıktıları eklendi.
- Warehouse örneğinde `StockWebhook` API'si `service InventoryIntegration` bloğuna bağlandı.
- `docs/service.md`, `docs/api.md`, diagnostics, generated-test docs, SPEC, BLACKLANG, README, AI agent contract, llms dosyaları, roadmap, web yol haritası ve site kaynağı güncellendi.
- Coverage taxonomy güncellendi: `backend-api-actions` score 94 -> 98, `WEB-API-006` done, weighted web coverage 92 -> 93.

Doğrulama:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs service --json` geçti.
- `go run ./cmd/black explain service --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected InventoryIntegration --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockWebhook --json` geçti.
- `go run ./cmd/black benchmark coverage --json` artık `completionPercent: 93` döndürüyor.
- Warehouse source: `format --check --json`, `lint --json`, `validate --json`, `build --out ../../generated --json` geçti.
- Warehouse generated çıktı: `npm run build`, `npm test`, `npm run test:e2e`, `npm run security:secrets:plan`, geçici SQLite ile `npm run jobs:run`, `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan` geçti.
- API-only service smoke: temp `target api` source format/lint/validate/inspect/build geçti; generated çıktı `npm install --silent`, `npm run build` ve `npm test` geçti.
- Static docs/site taraması yapıldı; service'in endpoint/handler üretmediği, tek syntax'ın `service Name { api APIName }` olduğu ve generated metadata çıktılarının tutarlı anlatıldığı doğrulandı.

Notlar:

- Service blocks secret veya external provider değeri taşımaz.
- Tek davranış için ikinci syntax eklenmedi; service grouping sadece top-level service bloğuyla ifade ediliyor.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden yeniden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan backend/API gap'i: cross-service command orchestration beyond explicit API grouping.

## Online Migration Runner MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Declared rename migration'ları artık generated web/API output içinde deploy öncesi ayrı bir read-only plan ve explicit apply kapısından geçebiliyor. Böylece AI veya deploy ajanı hedef veritabanını değer sızdırmadan kontrol eder, sadece kaynakta açıkça tanımlı rename operasyonlarını uygular ve ardından normal generated setup/seed yolunu çalıştırır.

Yapılanlar:

- Migration block içeren generated output artık src/migrate.ts dosyasını üretiyor.
- package.json içine db:migrate:plan ve db:migrate scriptleri eklendi.
- migrations/manifest.json runner metadata'sı planCommand, applyCommand, modes ve policy alanlarıyla genişletildi.
- SQLite plan modu missing database dosyasını oluşturmadan databaseReachable:false ve ready:false raporu veriyor.
- PostgreSQL plan modu unreachable database durumunu connection string değerini basmadan JSON olarak raporluyor.
- Apply modu yalnızca declared rename migration'larını çalıştırıyor ve başarılı migration'ları BlackMigration ledger'ına kaydediyor.
- Entity rename ve aynı migration içindeki field rename sırası deterministic kontrol ediliyor; WarehouseProduct -> Product ardından Product.quantity -> stock gibi zincirler plan modunda doğru source table üzerinden inceleniyor.
- Generated contract testleri package scripts, migration manifest runner metadata'sı ve src/migrate.ts varlığını doğruluyor.
- Deploy manifest affected files listesine migrations/manifest.json, migrations/*.sql ve src/migrate.ts eklendi.
- inspect --affected migration çıktısı generated migration runner ve package script etkilerini raporluyor.
- docs/migration.md, docs/deployment.md, docs/ai-agent-contract.md, README, SPEC, BLACKLANG, llms dosyaları, roadmap, web yol haritası ve website migration/coverage bölümleri güncellendi.
- Coverage taxonomy güncellendi: database-runtime score 94 -> 98, WEB-DATA-006 done, weighted web coverage 93 seviyesinde kaldı.

Doğrulama:

- go test -count=1 ./cmd/black -run 'Migration|Coverage|Contract|Deploy|Affected|BuildWebWritesMigrationRuntime' -v geçti.
- go test -count=1 ./cmd/black -run 'Docs|Explain|Migration|Coverage|BuildWebWritesMigrationRuntime|Golden' -v geçti.
- Warehouse golden manifest generator üzerinden güncellendi.
- cd packages/cli && go test -count=1 ./... geçti.
- CLI final zinciri geçti: help, docs migration, explain migrate, docs --all, explain entity, affected migration, benchmark coverage, format, lint, validate, build.
- Warehouse generated output: npm run build geçti.
- Warehouse generated output: npm test geçti.
- Warehouse generated output: npm run test:e2e geçti.
- Warehouse generated output: npm run security:secrets:plan geçti ve secret değerlerini basmadan eksik env listesini raporladı.
- Fresh SQLite ile npm run db:setup ve npm run jobs:run geçti.
- npm run deploy:rollback:plan geçti.
- npm run deploy:cloud:plan geçti.
- Migration runner smoke geçti: legacy WarehouseProduct.quantity veritabanında db:migrate:plan ready:true döndürdü, db:migrate RenameWarehouseProduct migration'ını uyguladı, ikinci db:migrate:plan pending:0 döndürdü.
- API-only migration smoke daha önce temp target api kaynakla format, lint, validate, inspect, build, npm build, npm test ve plan/apply/plan zincirinden geçti.
- Static docs/site taraması temiz: eski 91/92 coverage sinyali, online migration runners ifadesi ve schema prepared gibi fazla iddialı migration metinleri kalmadı.

Notlar:

- Yeni .black syntax eklenmedi; mevcut migration Name { rename ... } syntax'ı üzerinden generated runtime güvenliği artırıldı.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden üretildi.
- db:migrate:plan read-only kalacak şekilde tasarlandı; apply mutasyonu ayrı db:migrate scriptiyle sınırlı.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan database gap'i advanced database providers tarafında.

## Cross-Browser E2E Matrix MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated browser test intent'i artık tek browser koşusuyla sınırlı kalmadan read-only matrix planı ve available browser hedefleri üzerinde deterministic matrix run üretebiliyor. AI/CI ajanı hangi browser hedeflerinin bulunduğunu JSON olarak görebilir, ardından aynı generated e2e intent'ini Chrome, Edge, Chromium veya custom executable üzerinden tekrar çalıştırabilir.

Yapılanlar:

- Yeni .black syntax eklenmedi; mevcut top-level test Name { page ... expect ... } syntax'ı korunarak generated test yüzeyi genişletildi.
- target web ve test declaration olduğunda generator artık src/blacklang.e2e.matrix.ts üretir.
- target web ve test declaration olduğunda generator artık tests/browser-matrix.json üretir.
- package.json içine test:e2e:plan ve test:e2e:matrix scriptleri eklendi.
- test:all script'i npm test ardından test:e2e:matrix çalıştıracak şekilde güncellendi; test:e2e tek-browser geriye uyumlu komut olarak korundu.
- Browser matrix plan modu read-only JSON döndürür; browser veya app server başlatmaz, database mutasyonu yapmaz.
- Browser matrix target listesi custom, chrome, edge ve chromium olarak deterministic tutuldu.
- Matrix runner BLACKLANG_E2E_BROWSER_PATH, BLACKLANG_E2E_CHROME_PATH, BLACKLANG_E2E_EDGE_PATH ve BLACKLANG_E2E_CHROMIUM_PATH environment override'larını destekler.
- Matrix run mevcut src/blacklang.e2e.test.ts dosyasını child process ile yeniden kullanır; e2e davranışı iki farklı runtime'a bölünmedi.
- Matrix result JSON summary, targets, results, skipped, stdoutTail ve stderrTail alanlarıyla compact failure triage üretir.
- Generated contract tests package scriptleri, tests/browser-matrix.json manifestini ve src/blacklang.e2e.matrix.ts varlığını doğrular.
- inspect --affected test çıktısı matrix runner, matrix manifest ve yeni package script etkilerini raporlar.
- docs/test.md, docs/generated-test.md, docs/target.md, docs/README.md, docs/ai-agent-contract.md, docs/llms.txt, website/llms.txt, README, SPEC, BLACKLANG, ROADMAP-v0.2.md, blacklang-web-yol-haritasi.md ve website/index.html güncellendi.
- Coverage taxonomy güncellendi: testing-benchmarks score 82 -> 88, WEB-TEST-005 done, weighted web coverage 93 -> 94.

Doğrulama:

- go test -count=1 ./cmd/black -run 'BuildWebGeneratesBrowserChecks|Contract|Affected|Coverage' -v geçti.
- go test -count=1 ./cmd/black -run 'Coverage|Docs|Explain|BenchmarkTasks|BuildWebGeneratesBrowserChecks|Affected' -v geçti.
- Warehouse golden manifest generator üzerinden güncellendi.
- cd packages/cli && go test -count=1 ./... geçti.
- go run ./cmd/black benchmark coverage --json completionPercent 94 döndürdü.
- CLI final zinciri geçti: help, docs generated-test, docs test, explain test, docs --all, explain entity, inspect --affected WarehouseBrowserSmoke, benchmark coverage, format, lint, validate, build.
- Warehouse generated output compiler üzerinden yeniden üretildi ve yeni src/blacklang.e2e.matrix.ts ile tests/browser-matrix.json dosyalarını içerdi.
- Warehouse generated npm run build geçti.
- Warehouse generated npm test geçti; generated contract test matrix manifest ve script metadata'sını doğruladı.
- Warehouse generated npm run test:e2e:plan geçti; custom ve chromium skipped, Chrome ve Edge available olarak read-only JSON raporladı.
- Warehouse generated npm run test:e2e:matrix geçti; Chrome ve Edge hedeflerinde 2/2 passed, 0 failed, 2 skipped raporladı.
- Warehouse generated npm run test:e2e geriye uyumluluk için ayrıca geçti.
- Warehouse generated npm run test:all geçti; smoke test + matrix zinciri birlikte çalıştı.
- npm run security:secrets:plan, npm run deploy:rollback:plan, npm run deploy:cloud:plan geçti.
- Fresh SQLite ile npm run db:setup ve npm run jobs:run geçti.
- Static docs/site/generated taraması temiz: eski %93 coverage sinyali, testing score 82 ve cross-browser e2e matrix eksikliği kalmadı.

Notlar:

- Matrix runner Chromium family üzerinde çalışır; Firefox/WebKit ayrı provider/test runner tasarımı gerektirir ve bu MVP'nin dışında bırakıldı.
- Browser executable path değerleri secret değildir; matrix JSON path availability bilgisini gösterir ama secret/env değerlerini basmaz.
- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden üretildi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan testing gap'i long-running AI eval corpus tarafında.

## Long-running AI Eval Corpus MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang'in AI için daha düşük token maliyeti ve daha güvenli edit döngüsü iddiasını tek seferlik benchmark notundan çıkarıp tekrar koşulabilir, JSON/BlackIR ile okunabilir bir eval corpus'una bağlar. AI/CI harness'ları aynı case setini tekrar tekrar çalıştırabilir; case prompt, expected evidence, required commands ve 100 puanlık scoring rubric tek komuttan alınabilir.

Yapılanlar:

- Yeni .black syntax eklenmedi; mevcut source/generated benchmark ve AI task scenario altyapısı yeniden kullanıldı.
- `black benchmark eval [file] [--out <dir>] --json|--ir` komutu eklendi.
- Eval corpus read-only metadata olarak tasarlandı; AI model çağırmaz, configured output dizinini değiştirmez, commit/push/deploy yapmaz ve secret saklamaz.
- Eval suite metadata'sı `BLACKLANG-WEB-AI-EVAL-v0.2`, `mode: long-running`, `repeat: 3`, `timeoutMinutes: 120` ve discovery command listesiyle döner.
- Her AI task scenario için bir eval case üretilir; Warehouse şu an 8 case ve 24 total run planı döndürür.
- Her eval case reusable prompt, expected evidence, required validation commands, deterministic token estimate ve 100 puanlık scoring rubric taşır.
- Scoring rubric dört deterministic başlıktan oluşur: source-intent 30, agent-learning 20, validation 30, generated-boundary 20.
- `black benchmark eval --ir` compact BlackIR çıktısı suite id, mode, repeat, timeout, baseline, case count, total runs, max score, estimated minutes ve token totals bilgilerini verir.
- `black docs benchmark --json` eval komutunu ve agent kullanım notlarını içerir.
- `black agent startup --json` içine `benchmark-eval` komutu ve checklist adımı eklendi.
- Coverage taxonomy güncellendi: testing-benchmarks score 88 -> 98, WEB-TEST-006 done, weighted web coverage 94 -> 95.
- Docs/site güncellendi: README, BLACKLANG, SPEC, docs/benchmark.md, docs/README.md, docs/ai-agent-contract.md, docs/llms.txt, website/llms.txt, website/index.html, ROADMAP.md, ROADMAP-v0.2.md, blacklang-web-yol-haritasi.md ve benchmarks/warehouse-v0.1.md.
- Warehouse benchmark notundaki ölçümler güncel CLI sonucuna çekildi: 768 `.black` kaynak satırı, 63 generated dosya, 13756 generated satır, generated/source ratio 17.91, task token totals 10519 / 90872 / %88.
- Kullanıcının JavaScript parity sorusu sonrası roadmap'e stratejik ayrım eklendi: mevcut web completion hedefi deterministic web capability parity'dir; post-web Core Program Logic v1 ayrı future track olarak izlenecektir.
- Core Program Logic v1 future track başlıkları kaydedildi: literals, variables, assignment, full expression grammar, operator precedence, boolean logic, conditionals, functions, parameters, return, scope, collections, loops, errors ve bounded async operations.
- Ayrı core eval adayları kaydedildi: CORE-EVAL-001 Calculator, CORE-EVAL-002 Todo with local state, CORE-EVAL-003 Conditional pricing, CORE-EVAL-004 Loop over collection, CORE-EVAL-005 Bounded async API call.

Doğrulama:

- go test -count=1 ./cmd/black -run 'BenchmarkEval|Coverage|AgentStartup|Docs' -v geçti.
- cd packages/cli && go test -count=1 ./... geçti.
- go run ./cmd/black --help geçti ve benchmark açıklaması AI task/eval estimates olarak güncellendi.
- go run ./cmd/black benchmark --help geçti ve `black benchmark eval [file] [--out <dir>] [--json|--ir]` kullanımını gösterdi.
- go run ./cmd/black docs benchmark --json geçti ve syntax/agentNotes içinde benchmark eval bilgisi vardı.
- go run ./cmd/black agent startup ../../examples/warehouse/app.black --out ../../generated --json geçti ve commands içinde `benchmark-eval` döndürdü.
- go run ./cmd/black benchmark eval ../../examples/warehouse/app.black --out ../../generated --json geçti; 8 case, repeat 3, totalRuns 24, maxScoreAcrossRepeats 2400, estimatedSavingsPercent 88 döndürdü.
- go run ./cmd/black benchmark eval ../../examples/warehouse/app.black --out ../../generated --ir geçti; `benchmark eval ok` ve `cases 8 total_runs 24` içerdi.
- go run ./cmd/black benchmark coverage --json geçti; completionPercent 95 döndürdü.
- go run ./cmd/black benchmark issues --json geçti; summary total 41 open 0 done 41 ve WEB-TEST-006 done döndürdü.
- go run ./cmd/black benchmark issues --ir geçti; completion 95 ve summary total 41 open 0 done 41 döndürdü.
- go run ./cmd/black format --check --json ../../examples/warehouse/app.black geçti.
- go run ./cmd/black lint ../../examples/warehouse/app.black --json geçti.
- go run ./cmd/black validate ../../examples/warehouse/app.black --json geçti.
- go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json geçti.
- Warehouse generated npm run build geçti.
- Warehouse generated npm test geçti.
- Warehouse generated npm run test:e2e:plan geçti; Chrome ve Edge available, custom ve Chromium skipped raporlandı.
- Warehouse generated npm run test:e2e:matrix geçti; Chrome ve Edge hedeflerinde 2/2 passed, 0 failed, 2 skipped raporlandı.
- Warehouse generated npm run security:secrets:plan, npm run deploy:rollback:plan ve npm run deploy:cloud:plan geçti.
- `npm run jobs:run` default `generated/dev.db` ile stale schema yüzünden `tenantId` kolonuna takıldı; bu mevcut eski runtime DB kalıntısı olarak ayrıştırıldı.
- Fresh SQLite DATABASE_URL ile `npm run jobs:run` geçti; db:setup, db:seed ve LowStockMonitor worker query başarılı çalıştı.

Notlar:

- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden üretildi.
- Eval corpus gerçek model çalıştıran bir harness değildir; model-independent, deterministic plan ve scoring metadata üretir.
- `benchmark issues --ir` sadece compact total/open/done bilgisini ve open issue listesini verir; done issue ID'leri JSON export içinde görünür.
- Web completion çizgisi JavaScript runtime parity değildir; full JS/Python benzeri programlama kapasitesi için Core Program Logic v1 ayrı post-web roadmap track olarak tutuldu.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan testing gap'i published multi-model eval result history tarafında.

## Release Transparency + Key Rotation Policy MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Release imza doğrulamasını yalnızca checksum/signature eşleşmesinden çıkarıp public release güven zincirine bağlar. AI/CI artık yayınlanacak archive için transparency log entry formatını, entry hash zincirini, kullanılan public key identity formatını ve key rotation/revocation kurallarını deterministic JSON/IR metadata olarak okuyabilir.

Yapılanlar:

- Yeni `.black` runtime syntax eklenmedi; release trust/registry tarafında deterministic policy metadata genişletildi.
- `ReleaseTrust` metadata'sına `transparencyLog` ve `keyRotation` alanları eklendi.
- `black ecosystem --json` release trust altında transparency policy file, release log file, hash algorithm, required entry fields, key id format, minimum overlap days ve revocation file bilgilerini döndürür hale geldi.
- `black ecosystem --ir` compact output içine release transparency ve release key-rotation satırları eklendi.
- Human-readable `black ecosystem` çıktısı release transparency ve key rotation özetini gösterir hale geldi.
- Registry trust workflow policy files listesine `packages/registry/release-transparency.blackdir` ve `packages/registry/key-rotation-policy.blackdir` eklendi.
- Yeni registry manifestleri eklendi: `release-transparency.blackdir`, `key-rotation-policy.blackdir`, `key-revocations.blackdir`.
- Registry ve marketplace trust policy manifestleri release transparency policy ve key rotation policy gereksinimlerini içerir hale getirildi.
- Package index ve adapter index trust metadata'sına `release-transparency-log true` ve `release-key-rotation true` işaretleri eklendi.
- `packages/registry/scripts/validate-registry.mjs` yeni policy dosyalarını ve trust marker'larını doğrular hale geldi.
- `black docs release-trust --json`, `black docs ecosystem --json` ve package registry doküman notları transparency/key rotation bilgisini AI ajanları için açık hale getirdi.
- Coverage taxonomy güncellendi: auth-permissions-security score 90 -> 98, WEB-SEC-006 done, weighted web coverage 95 -> 96.
- Docs/site güncellendi: README, BLACKLANG, SPEC, docs/ecosystem.md, docs/release-trust.md, docs/package-registry.md, docs/README.md, docs/ai-agent-contract.md, docs/install.md, docs/release-artifacts.md, docs/llms.txt, website/llms.txt, website/index.html, ROADMAP-v0.2.md ve blacklang-web-yol-haritasi.md.
- Eski security gap ifadesi güncellendi: release transparency/key rotation artık prepared local policy seviyesinde tamamlandı; kalan gap secret manager provider execution.

Doğrulama:

- cd packages/cli && go test -count=1 ./... geçti.
- node packages/registry/scripts/validate-registry.mjs geçti.
- go run ./cmd/black ecosystem --json geçti; `release.trust.transparencyLog.policyFile` ve `release.trust.keyRotation.minimumOverlapDays` beklenen değerleri döndürdü.
- go run ./cmd/black ecosystem --ir geçti; release transparency ve release key-rotation satırları vardı.
- go run ./cmd/black docs release-trust --json geçti; transparency metadata ve key rotation agent notes içinde yer aldı.
- go run ./cmd/black docs ecosystem --json geçti.
- go run ./cmd/black benchmark coverage --json geçti; completionPercent 96 döndürdü.
- go run ./cmd/black benchmark issues --json geçti; summary total 42 open 0 done 42 ve WEB-SEC-006 done döndürdü.
- go run ./cmd/black benchmark issues --ir geçti; completion 96 ve summary total 42 open 0 done 42 döndürdü.
- Static docs/site taraması temiz: eski %95/%41 sinyali yalnızca önceki fazın tarihsel worklog kaydında kaldı.
- Root `package.json` bulunmadığı için root `npm run build` uygulanmadı; bu faz generated web output değiştirmedi.

Notlar:

- Private signing key, token veya secret üretilmedi ve `.black`/manifest dosyalarına secret yazılmadı.
- Bu faz gerçek public release publish işlemi yapmaz; public publishing öncesi gereken trust metadata/policy sınırını hazırlar.
- Generated dosyalar elle düzenlenmedi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan security gap'i secret manager provider execution tarafında.

## Deployment Provider CLI Adapter Execution MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

`deploy { cloud ... }` niyetini sadece Docker image ve plan metadata seviyesinden çıkarıp provider CLI adapter sınırına taşır. AI/CI artık cloud deploy için manifest, plan, read-only preflight ve mutation yapan explicit apply komutlarını ayırt edebilir; gerçek provider deploy yalnızca açık `deploy:cloud:exec` komutuyla çalışır.

Yapılanlar:

- Yeni `.black` runtime syntax eklenmedi; mevcut `deploy cloud` niyeti compiler/generator metadata ve generated script tarafında genişletildi.
- Cloud deploy manifesti `adapter: provider-cli`, `execution.mode: explicit-apply`, runner file, plan file ve preflight file bilgilerini taşır hale geldi.
- Generated package scriptleri üçe ayrıldı: `deploy:cloud:plan`, `deploy:cloud:preflight`, `deploy:cloud:exec`.
- `scripts/cloud-plan.mjs` auth env, optional env, present env ve execution metadata bilgilerini JSON olarak raporlar hale geldi.
- Yeni generated `scripts/cloud-exec.mjs` eklendi: default preflight read-only çalışır, `--apply` olmadan provider mutation yapmaz, eksik env/CLI durumunda JSON blockReasons döndürür.
- Apply runner provider CLI komutunu yalnızca gerekli env ve CLI hazırsa çalıştırır; stdout/stderr içinde configured env değerlerini redact eder ve uzun çıktıları truncation marker ile sınırlar.
- Fly için `fly deploy . --app ${FLY_APP_NAME} --yes`, Render için `render deploys create ${RENDER_SERVICE_ID} --wait --confirm -o json`, Railway için `railway up . --project ${RAILWAY_PROJECT_ID} --ci --yes --json` adapter argümanları üretildi.
- Cloud provider CLI metadata'sı AI-readable hale geldi: primary binary, alternative binaries, version args, auth env, optional env, required env ve apply args manifestte yer alıyor.
- `black inspect --affected deploy --json` artık cloud runner scriptini ve package script etkilerini döndürüyor.
- `black docs deploy --json`, `black explain deploy --json`, adapter marketplace docs, IDE snippet açıklaması ve ecosystem adapter metadata'sı provider CLI preflight/apply sınırını anlatır hale geldi.
- Adapter marketplace manifesti `scripts/cloud-exec.mjs`, read-only preflight ve explicit apply mutation trust marker'larını içerir hale getirildi.
- Coverage taxonomy güncellendi: deployment-operations score 88 -> 96, WEB-OPS-004 done, weighted web coverage 96 -> 97.
- Docs/site güncellendi: README, BLACKLANG, SPEC, docs/deployment.md, docs/ecosystem.md, docs/adapter-marketplace.md, docs/diagnostics.md, docs/ide.md, docs/ai-agent-contract.md, docs/llms.txt, website/llms.txt, website/index.html, ROADMAP.md, ROADMAP-v0.2.md, blacklang-web-yol-haritasi.md ve benchmarks/warehouse-v0.1.md.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated manifest golden test resmi update flag ile güncellendi.
- Warehouse benchmark ölçümü güncellendi: 768 `.black` kaynak satırı, 64 generated dosya, 14013 generated satır, generated/source ratio 18.25, task token totals 10519 / 90872 / %88.

Doğrulama:

- go test -count=1 ./cmd/black -run 'BuildWeb|Coverage|Affected|Docs|Explain|Ecosystem|IDE|Golden' -v geçti.
- UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v geçti ve golden manifest compiler çıktısına göre güncellendi.
- cd packages/cli && go test -count=1 ./... geçti.
- node packages/registry/scripts/validate-registry.mjs geçti.
- go run ./cmd/black --help geçti.
- go run ./cmd/black docs deploy --json geçti; cloud runner/preflight/apply notları vardı.
- go run ./cmd/black explain deploy --json geçti.
- go run ./cmd/black docs adapter-marketplace --json geçti.
- go run ./cmd/black ecosystem --json ve --ir geçti; deploy-cloud adapter capabilities provider CLI runner bilgilerini içerdi.
- go run ./cmd/black inspect ../../examples/warehouse/app.black --affected deploy --json geçti; `scripts/cloud-exec.mjs` affected files içindeydi.
- go run ./cmd/black benchmark coverage --json geçti; completionPercent 97 döndürdü.
- go run ./cmd/black benchmark issues --json geçti; summary total 43 open 0 done 43 ve WEB-OPS-004 done döndürdü.
- go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json geçti; generated files 64 ve ratio 18.25 döndürdü.
- go run ./cmd/black benchmark tasks ../../examples/warehouse/app.black --out ../../generated --json geçti; token savings %88 kaldı.
- go run ./cmd/black benchmark eval ../../examples/warehouse/app.black --out ../../generated --json geçti.
- go run ./cmd/black format --check --json ../../examples/warehouse/app.black geçti.
- go run ./cmd/black lint ../../examples/warehouse/app.black --json geçti.
- go run ./cmd/black validate ../../examples/warehouse/app.black --json geçti.
- go run ./cmd/black security scan ../../examples/warehouse/app.black --json geçti.
- go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json geçti.
- Warehouse generated npm run build geçti.
- Warehouse generated npm test geçti.
- Warehouse generated npm run test:e2e:plan geçti.
- Warehouse generated npm run test:e2e:matrix geçti; Chrome ve Edge passed, custom/Chromium skipped raporlandı.
- Warehouse generated npm run security:secrets:plan geçti.
- Warehouse generated npm run deploy:rollback:plan geçti.
- Warehouse generated npm run deploy:cloud:plan geçti.
- Warehouse generated npm run deploy:cloud:preflight geçti; read-only JSON olarak eksik env ve PATH'te provider CLI bulunmamasını blocked state ile raporladı.
- Fresh SQLite DATABASE_URL ile Warehouse generated npm run jobs:run geçti.

Notlar:

- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden üretildi.
- `deploy:cloud:exec` gerçek provider mutation yapabileceği için lokal doğrulamada çalıştırılmadı; bu komut bilinçli olarak explicit apply sınırıyla bırakıldı.
- Provider CLI komutları resmi Fly, Render ve Railway CLI dokümantasyonundaki güncel deploy komutları baz alınarak seçildi.
- Fresh jobs testinden kalan geçici SQLite DB git status içinde görünmedi; güvenlik filtresi lokal `Remove-Item` temizliğini reddettiği için silme tekrar zorlanmadı.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.
- Açık kalan deployment gap'i distributed tracing/exporter plugins tarafında.

## Advanced Relation Loading Policies MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?
Relation field'larının generated API response içinde hangi context'lerde relation object include edeceğini `.black` kaynakta deterministic seçilebilir hale getirir. Böylece AI veya geliştirici generated JS/TS'ye dokunmadan list/detail/query/mutation response davranışını tek kaynak niyetle kontrol eder.

Tamamlananlar:
- Entity relation field'ları için tek canonical syntax eklendi: `customer Customer required load list detail query mutation` veya `load none`.
- Parser/AST `RelationLoad` bilgisini JSON output'a taşır; BlackIR `load list detail query` formunu virgülsüz korur.
- Validator `load` modifier'ını sadece relation field'larda kabul eder; eksik scope, unsupported scope, duplicate scope, duplicate `load`, `load none` conflict ve auth user field kullanımı için stable diagnostics üretir.
- Generator list, detail, query ve mutation route context'lerinde Prisma `include` üretimini relation load policy'ye bağladı.
- Generated OpenAPI operations artık `x-blacklang-relation-load-context` ve `x-blacklang-relation-load` metadata'sı yayımlar.
- `inspect --affected Entity.relationField --json` relation field'ları ayrı `relation-field` kind olarak raporlar ve relation load policy etkisini generated files/agent notes içinde gösterir.
- IDE manifest `load` modifier'ı, relation-load scope completion'ları ve `relation-load` snippet'i içerir.
- `docs relation-load --json`, `explain relation-load --json`, diagnostics docs, query docs, AI agent contract, llms, README, SPEC, BLACKLANG, roadmap, Türkçe web yol haritası ve statik website source güncellendi.
- Warehouse örneğinde `Order.customer` explicit `load list detail query mutation` policy ile güncellendi.
- Coverage report'ta `WEB-DATA-007` done olarak eklendi; data-model-crud score 99 ve weighted web coverage signal 98 oldu.
- Generated SQLite setup runtime'da text/date/datetime default literal'larının çift tırnak yerine single-quoted SQL literal olarak üretilmesi düzeltildi; bu, fresh generated `jobs:run` doğrulamasında yakalanan gerçek runtime hatasını giderdi.

Doğrulama:
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain entity --json`, `explain query --json`, `explain relation-load --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Order.customer --json` geçti ve `kind: relation-field` döndürdü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; Warehouse source 768 lines, generated 64 files, generated 14253 lines, ratio 18.56.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 98, data-model-crud score 99.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 44 open 0 done 44.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:plan` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.
- Generated Warehouse fresh SQLite `DATABASE_URL=file:./.tmp-relation-load-jobs-final-20260906.db npm run jobs:run` geçti; seed 4 rows ve LowStockMonitor count 1 döndürdü.
- Generated Warehouse `npm run deploy:rollback:plan`, `npm run deploy:cloud:plan`, `npm run deploy:cloud:preflight`, `npm run security:secrets:plan` geçti. Cloud preflight expected read-only blocked state döndürdü çünkü FLY_APP_NAME/FLY_REGION env ve provider CLI yok.

Notlar:
- `load` yazılmazsa backward-compatible default olarak list/detail/query/mutation response'larında relation object include edilir.
- `load none`, relation object include etmez; generated UI relation ID fallback ile deterministik kalır.
- Relation load policy database schema'yı değiştirmez ve custom query filter/sort için relation join syntax'ı eklemez.
- Bu aşamada GitHub commit/push veya VPS publish yapılmadı; kullanıcının istediği gibi sonuç iş günlüğüne kaydedildi.
- Default `generated/dev.db` eski local schema artığı olduğu için çıplak `npm run jobs:run` hâlâ eski DB'ye takılabilir; doğrulama fresh `DATABASE_URL` ile yapıldı. SQLite default literal bug'ı generator'da düzeltildi.

## Advanced Reusable Component Composition MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Declared `component` bloklarının sadece table/detail/form field renderer olarak değil, page `view` içinde reusable inline panel olarak yerleşmesini sağlar. AI veya geliştirici artık generated React/CSS dosyalarına dokunmadan, `.black` içinde bir component'i selected row ya da ilk listedeki kayıt üzerinden sayfa section'ı olarak gösterebilir.

Tamamlananlar:

- Tek canonical syntax eklendi: `section StockSummary component StockBadge bind selected span 1 title "Stock Summary"`.
- `bind selected` current selected/detail record'u, `bind first` loaded list'in ilk kaydını component'e geçirir.
- Parser/AST `ViewSectionDecl` içine `component` ve `bind` bilgisini ekledi; JSON output ve BlackIR `view-section ... component ... bind ...` olarak taşır.
- Validator component section adlarını `order`, `group` ve `tab` içinde destekler; unknown component, missing/unsupported bind, built-in section conflict, invalid component section name, unsupported modal/drawer/side, list/entity input, missing source field ve type mismatch için stable diagnostics üretir.
- Component section input'ları source entity üzerindeki stored veya computed field'larla isim ve tip eşleşmesiyle deterministic bağlanır.
- Generator page component section'ları için component import'u, selected/first item state binding'i, permission guard, empty-state message, stable CSS class'lar ve inline panel markup üretir.
- Component section'lar table/detail/form fiziksel render slotlarına deterministic şekilde yerleşir; order'da belirtilmeyen component section'lar default built-in section'lardan sonra append edilir.
- `compose grid`, `compose tabs`, contiguous `group` ve view order mekanizması component section adlarını anlayacak şekilde güncellendi.
- `inspect --affected` component section kullanımını sayfa ve generated component/page dosyalarına bağlar; field affected analizinde component section input kullanımı hesaba katılır.
- IDE metadata `component`, `bind`, `selected`, `first` completion'ları ve `view-component-section` snippet'i içerir.
- `docs view --json`, `docs component --json`, `explain view --json`, diagnostics docs, AI agent contract, llms, README, SPEC, BLACKLANG, roadmap, Türkçe web yol haritası ve statik website source güncellendi.
- Warehouse örneğinde Products page artık Record tab içinde `StockSummary` component section'ı render eder.
- Coverage taxonomy güncellendi: `WEB-LAYOUT-005` done; `frontend-ui-layout` score 98 -> 99; weighted web coverage signal yuvarlama nedeniyle %98 kaldı.
- Warehouse benchmark ölçümü güncellendi: 769 `.black` kaynak satırı, 64 generated dosya, 14289 generated satır, generated/source ratio 18.58.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated manifest golden test resmi update flag ile güncellendi.

Doğrulama:

- `cd packages/cli && go test -count=1 ./...` geçti.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black` geçti ve golden manifest güncellendi.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black docs component --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockBadge --json` geçti; Products page ve `src/pages/ProductsPage.tsx` component section etkisiyle raporlandı.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti; component section field kullanımını hesaba kattı.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti; findings boş.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 98, frontend-ui-layout score 99.
- `go run ./cmd/black benchmark coverage --ir` geçti; WEB-LAYOUT-005 done göründü.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 45 open 0 done 45.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 45 open 0 done 45.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; generated/source ratio 18.58.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; Warehouse output compiler/generator üzerinden üretildi.
- Component section'lar bu MVP'de inline render edilir; modal/drawer component panels, collection-backed component sections ve nested component slots Core Program Logic/collection support sonrası büyütülecek alandır.
- Component section input binding bilinçli olarak isim+tip eşleşmesine indirildi; arbitrary prop mapping syntax'ı eklenmedi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.


## DOM-order View Rendering MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated page section'larının gerçek JSX/DOM sırasını `.black` içindeki effective `view.order` ile aynı hale getirir. Böylece görsel sıra, klavye/screen-reader akışı, test seçicileri ve gelecekteki daha zengin composition modeli aynı deterministic kaynak niyetine dayanır.

Tamamlananlar:

- Generator page render akışı refactor edildi: `pageViewOrder(page)` artık table/detail/form ve declared component section'ları doğrudan JSX/DOM sırasıyla render ediyor.
- Table, detail ve form JSX üretimi ayrı helper'lara taşındı; component section rendering aynı order pipeline'ına bağlandı.
- Eski slot tabanlı component section render helper'ları kaldırıldı; component placement için tek model effective page view order oldu.
- Products generated page sırası doğrulandı: `bl-view-section-table`, `bl-view-section-stock-summary`, `bl-view-section-detail`, `bl-view-section-form`.
- Generator testi ProductsPage içinde DOM marker sırasını assert edecek şekilde genişletildi.
- CSS order expectation'ları component section sırasını da kapsayacak şekilde güncellendi; CSS order artık stable metadata/layout backstop olarak dokümante edildi.
- Coverage taxonomy güncellendi: `WEB-LAYOUT-006` done; `frontend-ui-layout` score 99 ve weighted web coverage signal %98 olarak kaldı.
- `docs view --json`, `docs/README.md`, `docs/ai-agent-contract.md`, `docs/llms.txt`, `README.md`, `SPEC.md`, `BLACKLANG.md`, `ROADMAP-v0.2.md`, Türkçe web yol haritası, website source ve Warehouse benchmark notu DOM-order davranışını anlatacak şekilde güncellendi.
- Warehouse benchmark ölçümü güncellendi: 769 `.black` kaynak satırı, 64 generated dosya, 14288 generated satır, generated/source ratio 18.58.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated manifest golden test resmi update flag ile güncellendi.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black` beklenen şekilde yalnızca golden manifest değişikliğiyle durdu.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black` geçti ve golden manifest güncellendi.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockBadge --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti.
- `go run ./cmd/black benchmark coverage --ir` geçti; `WEB-LAYOUT-006` done göründü.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 46 open 0 done 46.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 46 open 0 done 46.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; generated/source ratio 18.58.
- Generated ProductsPage DOM marker doğrulaması geçti: table line 603, StockSummary line 692, detail line 710, form line 755.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; output compiler/generator üzerinden üretildi.
- Bu aşama yeni `.black` syntax eklemedi; mevcut `view.order` davranışını DOM seviyesinde tamamladı.
- CSS `order` kuralları geriye dönük stabil sınıf/layout metadata olarak kalıyor, ama generated JSX/DOM artık kaynak order ile hizalı.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.


## Single-hop Relation Batch Prefetch and Nested Sanitization MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated API response'larında relation object attach davranışını daha ölçekli ve güvenli hale getirir. List/query response'ları relation ID'lerini tekilleştirip her relation field için tek target lookup yapar; permission-aware build'lerde attached target records da hedef entity read policy'sine göre sanitize edilir.

Tamamlananlar:

- Yeni `.black` syntax eklenmedi; mevcut relation `load list|detail|query|mutation|none` intent'i daha iyi runtime'a çevrildi.
- Generator route output'u için `attach<Entity>Relations(req, items, context)` helper'ı eklendi.
- List ve query route'ları relation ID'lerini `Set` ile tekilleştirip her relation field için tek `findMany` target lookup yapıyor.
- Detail, create, update, archive, restore, workflow transition ve custom action response'ları aynı attach helper'ı tek item array'i ile kullanıyor.
- Eski Prisma `include: { relation: true }` response loading yolu ana CRUD/query/action route'larından kaldırıldı; relation placement için tek deterministic attach modeli kaldı.
- Permission-aware route'larda source relation field okunamıyorsa relation object attach edilmiyor.
- Attached target records için `sanitize<Target>Relation(record, roles)` helper'ı üretildi; hedef entity field read izinleri nested relation response içinde de uygulanıyor.
- Target entity owner/tenant policy varsa batch relation lookup target row policy helper'ından geçiyor.
- `routeNeedsCurrentUser` relation target policy kullanan route'larda current user helper üretimini hesaba katacak şekilde genişletildi.
- Relation-load generator testi batch attach, context gating ve nested sanitizer output'unu doğrulayacak şekilde güncellendi.
- Coverage taxonomy güncellendi: `WEB-DATA-008` done; `data-model-crud` score 100, `auth-permissions-security` score 99, weighted web coverage signal %98.
- `docs relation-load --json`, `docs/relation-load.md`, `docs/llms.txt`, `website/llms.txt`, README, SPEC, BLACKLANG, ROADMAP-v0.2, Türkçe web yol haritası, website source ve Warehouse benchmark notu güncellendi.
- Warehouse benchmark ölçümü güncellendi: 769 `.black` kaynak satırı, 64 generated dosya, 14325 generated satır, generated/source ratio 18.63.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run TestBuildWebGeneratesRelationLoadPolicies` geçti.
- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBuildWebGeneratesRelationLoadPolicies|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR|TestFindDocs|TestExplain"` geçti.
- `cd packages/cli && go test -count=1 ./cmd/black` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs relation-load --json` geçti.
- `go run ./cmd/black explain relation-load --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Order.customer --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; data-model-crud score 100, auth-permissions-security score 99, completionPercent 98.
- `go run ./cmd/black benchmark coverage --ir` geçti.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 47 open 0 done 47.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 47 open 0 done 47.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; generated/source ratio 18.63.
- Generated Order route marker kontrolü geçti: `attachOrderRelations`, `customerIds`, `customerRecords`, `customerMap`, `sanitizeCustomerRelation`, `rowPolicyWhereForCustomerRelation`; eski `include: { customer: true }` yok.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; output compiler/generator üzerinden üretildi.
- Bu MVP single-hop belongs-to relation attachment içindir. Multi-hop relation graph loading explicit relation graph syntax sonrası değerlendirilmelidir.
- Relation load policy hâlâ custom query predicate join syntax'ı eklemez; query where/sort/aggregate stored primitive alanlarla sınırlıdır.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.

## Secret Manager Provider Preflight Execution MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated web/API output secret değerlerini `.black`, manifest veya log içine yazmadan provider handoff sınırını doğrulayabilir. Deployment agent'ı artık environment readiness ile birlikte seçilen secret provider label/prefix ve provider CLI availability bilgisini JSON olarak kontrol eder.

Tamamlananlar:

- Yeni `.black` syntax eklenmedi; mevcut env/secret manifest intent'i daha güçlü generated runtime artefaktına çevrildi.
- Generated output'a `scripts/secrets-provider.mjs` eklendi.
- Generated `package.json` içine `security:secrets:preflight` script'i eklendi: `node scripts/secrets-provider.mjs --preflight`.
- `security/secrets.json` manifest metadata'sına `providerFile: "scripts/secrets-provider.mjs"` eklendi.
- Provider preflight sadece read-only çalışır; `--preflight` olmadan çalışmayı reddeder.
- Supported provider labels aynı tutuldu: `environment`, `1password`, `vault`, `doppler`, `aws-secrets-manager`.
- Default `environment` provider için required env presence raporlanır.
- External provider labels için provider CLI availability kontrol edilir; secret value fetch/inject/write/log yapılmaz.
- Provider CLI check environment'ı sınırlı tutuldu; secret env değerleri child process env'ine aktarılmıyor.
- Generated contract testleri secret provider script metadata'sını, package script'ini ve dosya varlığını doğrulayacak şekilde genişletildi.
- `inspect --affected security` artık `security/secrets.json`, `scripts/secrets-plan.mjs`, `scripts/secrets-provider.mjs` ve package script etkilerini gösteriyor.
- Coverage taxonomy güncellendi: `WEB-SEC-007` done; `auth-permissions-security` score 100 oldu, weighted web coverage signal yuvarlamayla %98 kaldı.
- Docs ve site kaynakları güncellendi: `docs/secret-management.md`, `docs/deployment.md`, `docs/ai-agent-contract.md`, `docs/llms.txt`, `website/llms.txt`, README, SPEC, BLACKLANG, ROADMAP-v0.2, Türkçe web yol haritası ve website source.
- Warehouse benchmark ölçümü güncellendi: 769 `.black` kaynak satırı, 65 generated dosya, 14461 generated satır, generated/source ratio 18.80.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated manifest golden test resmi update flag ile güncellendi.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBuildWebWritesExpectedFiles|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR|TestFindDocs|TestExplain|TestInspect"` geçti.
- `cd packages/cli && go test -count=1 ./...` önce beklenen golden manifest değişikliğiyle durdu.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black` geçti ve golden manifest güncellendi.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs security --json` geçti.
- `go run ./cmd/black docs deploy --json` geçti.
- `go run ./cmd/black explain security --json` geçti.
- `go run ./cmd/black explain deploy --json` geçti.
- `go run ./cmd/black inspect --affected security ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect --affected deploy ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 98, auth-permissions-security score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 48 open 0 done 48.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 48 open 0 done 48.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti.
- Generated Warehouse `npm run security:secrets:preflight --silent` geçti; command `security:secrets:preflight`, provider `environment`, references 7, output içinde `value` field yok.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; output compiler/generator üzerinden üretildi.
- Bu aşama provider-owned runtime value injection değildir. Bilinçli sınır: secret manager'dan değer çekme/enjekte etme policy'si provider adapter katmanında ayrıca tasarlanmalı.
- Secret değerleri `.black`, manifest, generated preflight output veya provider CLI child env içine konmadı.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.

## Collection-backed Component Sections MVP

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Reusable page component section'ları sadece selected veya first record'a bağlı kalmaz; aynı component loaded list/query collection içindeki her row için deterministic olarak render edilebilir. Bu, Core Program Logic v1'e geçmeden önce web UI tarafında güvenli collection rendering ihtiyacını kapatır.

Tamamlananlar:

- Mevcut `section <Name> component <Component> bind ...` syntax'ı genişletildi; canonical yeni bind modu `each` oldu.
- Parser hata mesajı `bind selected|first|each` bilgisini verecek şekilde güncellendi.
- Validator `bind each` değerini destekliyor; unsupported bind negatif testi `bind all` değerine taşındı.
- Generator `bind each` için `const <section>ComponentItems = items;` state alias'ı üretir.
- Generated JSX `component-section-list` ve `component-section-list-item` wrapper'larıyla her loaded item için component instance render eder.
- Permission gating selected/first ile aynı kaldı; component input alanları hidden ise collection section da hidden olur.
- Component props stored veya computed source fields üzerinden aynı deterministic name/type binding kuralıyla üretilir.
- Generated CSS'e collection component list stilleri eklendi.
- `inspect --affected` artık custom view section symbol'larını tanıyor: `StockCards` ve qualified `Products.StockCards` kind `view-section` olarak döner.
- IDE completion/snippet metadata `selected`, `first`, `each` bind modlarını anlatacak şekilde güncellendi.
- Parser, validator, generator, BlackIR, affected ve coverage testleri yeni behavior'u kapsayacak şekilde genişletildi.
- Warehouse örneğine `section StockCards component StockBadge bind each span 1 title "Stock Cards"` eklendi.
- Coverage taxonomy güncellendi: `WEB-LAYOUT-007` done; `frontend-ui-layout` score 100 oldu, weighted web coverage signal yuvarlamayla %98 kaldı.
- Docs ve site kaynakları güncellendi: `docs/view.md`, `docs/diagnostics.md`, `docs/README.md`, `docs/ai-agent-contract.md`, `docs/llms.txt`, `website/llms.txt`, README, SPEC, BLACKLANG, ROADMAP-v0.2, Türkçe web yol haritası ve website source.
- Warehouse benchmark ölçümü güncellendi: 770 `.black` kaynak satırı, 65 generated dosya, 14503 generated satır, generated/source ratio 18.84.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated manifest golden test resmi update flag ile güncellendi.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run "TestParsePageViewComponentSection|TestValidatePageViewComponentSection|TestBuildWebWritesExpectedFiles|TestFormatBlackIR|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR|TestFindDocs|TestExplain|TestIDE"` geçti.
- `cd packages/cli && go test -count=1 ./...` önce beklenen golden manifest değişikliğiyle durdu.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black` geçti ve golden manifest güncellendi.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs view --json` geçti.
- `go run ./cmd/black explain view --json` geçti.
- `go run ./cmd/black ide --json` geçti.
- `go run ./cmd/black ide diagnostics ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect --affected StockCards ../../examples/warehouse/app.black --json` geçti; affected kind `view-section`, found `true`.
- `go run ./cmd/black inspect --affected Products.StockCards ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 98, frontend-ui-layout score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 49 open 0 done 49.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 49 open 0 done 49.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 770 source lines, 65 generated files, 14503 generated lines, ratio 18.84.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- Generated ProductsPage marker doğrulaması geçti: `const stockCardsComponentItems = items;`, `bl-view-section-stock-cards`, `component-section-list`, `stockCardsComponentItems.map((item) => (`.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; output compiler/generator üzerinden üretildi.
- `bind each` arbitrary loop syntax değildir; page'in mevcut loaded list/query collection'ını kullanır. Genel collection iteration Core Program Logic v1 track'inde ayrıca tasarlanmalıdır.
- Component section input binding hâlâ isim+tip eşleşmesiyle sınırlıdır; arbitrary prop mapping eklenmedi.
- GitHub commit/push ve VPS/site yayını yapılmadı; kullanıcı final incelemesi sonrası toplu yapılacak.

## Resume Baseline Validation - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Context reset sonrası repo gerçek durumunu yeniden kurar ve yeni web-completion işi başlamadan önce compiler ile generated Warehouse çıktısının hâlâ yeşil olduğunu kanıtlar.

Kontrol edilenler:

- `AGENTS.md`, `SPEC.md`, `BLACKLANG.md`, `ROADMAP.md`, `ROADMAP-v0.2.md`, `blacklang-web-yol-haritasi.md`, `blacklang.toml`, mevcut `WEB-COMPLETION-WORKLOG.md`, git status, generated output dizini ve test notu izleri okundu.
- Git çalışma ağacında önceki web completion fazlarından kalan büyük uncommitted değişiklik seti olduğu doğrulandı; commit/push/deploy yapılmadı.
- Coverage raporu `completionPercent: 98`, `issues` summary `total: 49 open: 0 done: 49` döndürdü.
- Açık kalan en yüksek etkili web boşluğu deployment/ops tarafında `distributed tracing/exporter plugins` olarak ayrıştırıldı.
- Warehouse generated output compiler üzerinden yeniden üretildi.

Doğrulama:

- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; `count: 72`.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected ops --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; deployment-operations score 96 ve distributed tracing/exporter plugins missing göründü.
- `go run ./cmd/black benchmark issues --json` geçti; open issue yok.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 65 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; output compiler/generator üzerinden üretildi.
- Root `.tmp_*.txt` dosyaları ve generated içindeki eski local SQLite test DB artıkları mevcut; doğrulama açısından bloklayıcı değiller.
- Bundan sonraki aday aşama, ops observability modeline deterministic tracing/exporter manifest/preflight desteği eklemek.

## OTLP Trace Exporter / Distributed Tracing MVP - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Generated web/API uygulamalarının mevcut `observe PROVIDER endpoint env NAME` syntax'ı üzerinden distributed trace context üretmesini, OTLP HTTP JSON trace payload export edebilmesini ve bu davranışı AI ajanlarının OpenAPI/manifest/test çıktılarından deterministik şekilde anlayabilmesini sağlar.

Tamamlananlar:

- `observe` provider listesine `otlp` eklendi; eski `webhook` provider davranışı korundu.
- Generated server observe varsa W3C `traceparent` context üretir veya gelen context'i taşır, response `traceparent` header'ını yazar.
- `observe webhook endpoint env NAME` non-blocking structured request event göndermeye devam eder.
- `observe otlp endpoint env NAME` non-blocking OTLP HTTP JSON trace payload gönderir.
- CORS kullanılan generated app'lerde observe varsa `traceparent` request/response header yüzeyi izinli hale getirildi.
- Generated output'a `ops/observability.json` eklendi; provider, endpoint env, protocol, mode, signal, trace context, delivery ve exporter metadata'sı kaydediliyor.
- OpenAPI `x-blacklang-observability`, generated contract test ve generated API smoke test OTLP/trace context metadata'sını doğrulayacak şekilde güncellendi.
- Generated API smoke test local receiver ile OTLP `resourceSpans` payload'ını ve response `traceparent` header'ını doğruluyor; non-blocking delivery kapanış gürültüsünü önlemek için final drain beklemesi eklendi.
- Warehouse örneği `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT` kullanacak şekilde güncellendi.
- `black inspect --affected ops --json` artık `ops/observability.json` ve exporter metadata etkisini gösteriyor.
- `black ecosystem --json|--ir` ve adapter marketplace metadata içine `observability:otlp` eklendi.
- Coverage taxonomy güncellendi: `WEB-OPS-005` done; `deployment-operations` score 100; weighted web coverage signal %99; issues summary `total: 50 open: 0 done: 50`.
- Docs, SPEC, BLACKLANG, roadmap, AI agent contract, IDE docs/snippet, docs site source ve `llms.txt` OTLP/trace context davranışıyla güncellendi.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated dosyalar elle düzenlenmedi.
- Warehouse golden manifest resmi `UPDATE_BLACKLANG_GOLDEN=1` test yolu ile güncellendi ve flag olmadan tekrar doğrulandı.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBuildWebGeneratesOpsRuntime|TestBuildWebGeneratesOTLPTraceExporter|TestValidateOpsDiagnostics|TestAnalyzeAffectedOps|TestFindOpsDoc|TestExplainOpsKeyword|TestIDESupportIncludesCompletionSnippetsAndDiagnostics|TestEcosystem|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR" -v` geçti.
- `cd packages/cli && go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` önce beklenen golden değişikliğiyle durdu.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest -v` geçti.
- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBuildWebGeneratesOpsRuntime|TestBuildWebGeneratesOTLPTraceExporter|TestWarehouseGeneratedGoldenManifest" -v` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti; adapter listesinde `observability:otlp` var.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs ops --json` geçti; syntax `observe webhook|otlp endpoint env <ENV_NAME>`.
- `go run ./cmd/black docs --all --json` geçti; `count: 72`.
- `go run ./cmd/black explain ops --json` geçti.
- `go run ./cmd/black explain entity --json` geçti.
- `go run ./cmd/black ecosystem --json` ve `go run ./cmd/black ecosystem --ir` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black security scan ../../examples/warehouse/app.black --json` geçti; finding yok.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected ops --json` geçti; affected graph `ops/observability.json` içeriyor.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 99, deployment-operations score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 50 open 0 done 50.
- `go run ./cmd/black benchmark coverage --ir` geçti; completion 99.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 50 open 0 done 50.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 770 source lines, 66 generated files, 14628 generated lines, generated/source ratio 19.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- `generated/ops/observability.json` doğrulandı: provider `otlp`, endpointEnv `BLACKLANG_OTLP_ENDPOINT`, protocol `otlp-http-json`, signal `traces`, traceContext `w3c-traceparent`, delivery `non-blocking`.
- `cd generated && npm run security:secrets:plan --silent` exit 0; `BLACKLANG_OTLP_ENDPOINT` optional sensitive observability endpoint olarak listelendi, secret değeri yazdırılmadı. `ready: false` beklenen şekilde sadece required deployment env'leri (`DATABASE_URL`, `FLY_APP_NAME`, `FLY_REGION`) set olmadığı için.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Commit, push veya VPS/canlı site deploy yapılmadı.
- Kullanıcının son talimatına göre bu mevcut iş burada tamamlandı ve yeni faza geçilmedi.

## MySQL Target / Advanced Database Provider MVP - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang'in `target database` yüzeyini SQLite/PostgreSQL dışına taşıyıp MySQL'i deterministic generator, deploy metadata, seed runtime, IDE/AI docs ve registry/marketplace akışına ekler. Böylece kısa vadeli hedef olan deterministic web capability parity tarafında advanced database provider boşluğu repo içinde test edilebilir hale gelir.

Tamamlananlar:

- Validator `database mysql` hedefini kabul edecek şekilde güncellendi; unsupported provider önerileri `sqlite`, `postgres`, `mysql` ile hizalandı.
- MySQL + `migration` blokları için bilinçli sınır kondu: generated rename migration runner henüz MySQL DDL güvenlikleri tamamlanmadan desteklenmiyor ve `UNSUPPORTED_TARGET_DATABASE_MIGRATION` diagnostic'i üretiyor.
- Generator MySQL için Prisma provider `mysql`, `@prisma/adapter-mariadb` / `PrismaMariaDb` runtime, default `mysql://blacklang:blacklang@localhost:3306/blacklang` URL, `.env.example` MySQL değişkenleri ve Prisma config default'u üretiyor.
- Docker Compose ve preview compose SQL servis üretimi ortaklaştırıldı; MySQL hedefinde `mysql:8.4`, deterministic env defaults, healthcheck ve ayrı preview database URL'i üretiliyor.
- MySQL seed runtime Prisma `upsert` tabanlı üretildi; seed satırlarına deterministic `model` ve `types` metadata'sı eklendi, date/datetime değerleri provider-safe şekilde coerce ediliyor.
- MySQL transactional audit yollarında provider uyumlu Prisma create kullanıldı; SQLite/PostgreSQL mevcut raw SQL davranışı korundu.
- Generated API smoke testleri SQLite dışı provider'larda DB env değerini zorla memory SQLite'a çevirmeyecek şekilde düzeltildi.
- IDE completion listesi compiler-owned `supportedTargetDatabases` kaynağından `target-database` completion'ları üretiyor.
- Ecosystem discovery ve adapter marketplace metadata'sına `target:web-react-node-mysql` eklendi.
- Coverage taxonomy güncellendi: `WEB-DATA-009` done; `database-runtime` score 100; issues summary `total: 51 open: 0 done: 51`.
- Docs, SPEC, BLACKLANG, roadmap, AI agent contract, docs site source, `llms.txt`, diagnostics, seed, target, deployment ve adapter marketplace dokümanları MySQL hedefi ve migration sınırı ile güncellendi.
- Warehouse generated output compiler üzerinden yeniden üretildi; generated dosyalar elle düzenlenmedi.
- Warehouse golden manifest resmi `UPDATE_BLACKLANG_GOLDEN=1` test yolu ile güncellendi ve flag olmadan tekrar doğrulandı.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBuildWebSupportsMySQLRuntime|TestValidateTargetAllowsMySQLDatabase|TestValidateTargetMySQLRejectsMigrations|TestEcosystemStatusIncludesRegistryAndMarketplaceMetadata|TestFormatEcosystemIR|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestFormatCoverageIssuesIR|TestIDESupportIncludesCompletionSnippetsAndDiagnostics"` geçti.
- `cd packages/cli && UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti; adapter listesinde `target:web-react-node-mysql` var.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black explain target --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 99, `database-runtime` score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 51 open 0 done 51.
- `go run ./cmd/black ecosystem --json` geçti; MySQL built-in target adapter ve marketplace provider metadata'sı göründü.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Warehouse örneği bilinçli olarak `database sqlite` kalmaya devam ediyor; MySQL runtime compiler testinde ayrı bir deterministic app üzerinden doğrulandı.
- MySQL rename migration runner bilinçli olarak sonraki DDL-safety track'ine bırakıldı; dilde yeni syntax eklenmedi.
- Commit, push veya VPS/canlı site deploy yapılmadı.

## Eval History / Published Results Metadata - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang'in uzun koşulu AI eval corpus metadata'sını gerçek sonuç geçmişinden ayırır ve `benchmarks/eval-history.blackdir` üzerinden evidence-backed, compiler-owned JSON/BlackIR history export'u sağlar. Böylece ajanlar "hangi eval koşuları, hangi kanıtlarla, hangi pass/fail toplamlarıyla kayda geçmiş?" sorusunu az tokenle ve deterministik biçimde yanıtlayabilir.

Tamamlananlar:

- Yeni `black benchmark eval-history [--history <file>] --json|--ir` alt komutu eklendi.
- Varsayılan manifest keşfi çalışma dizininden yukarı doğru `benchmarks/eval-history.blackdir` arayacak şekilde yapıldı; `--history <file>` ile harici manifest okunabiliyor.
- `benchmarks/eval-history.blackdir` eklendi; ilk kayıt `published-local` statüsünde, Warehouse web completion doğrulamalarına ve worklog/benchmark/docs kanıtlarına bağlı.
- JSON result tipi `history`, `summary`, `runs`, `models`, `evidence`, `notes` alanlarıyla eklendi.
- BlackIR formatter eklendi; compact history identity, summary, run, model/source ve evidence satırları üretiyor.
- Local validation entry'lerinin billed model benchmark skoru gibi sunulmaması policy olarak manifest, docs ve tests içine işlendi.
- Coverage taxonomy güncellendi: `WEB-TEST-007` done; `testing-benchmarks` score 100 ve status `mature`.
- Docs, SPEC, BLACKLANG, README, AI agent contract, `llms.txt`, docs site source ve Türkçe web yol haritası eval-history komutu ve policy sınırıyla güncellendi.
- Website benchmark sayıları güncel generated ölçüme çekildi: 770 source lines, 66 generated files, 14665 generated lines, ratio 19.05.

Doğrulama:

- `cd packages/cli && go test -count=1 ./cmd/black -run "TestBenchmarkEvalHistory|TestFormatAIEvalHistoryIR|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR|TestFindBenchmarkDoc|TestDocsAll"` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black docs benchmark --json` geçti.
- `go run ./cmd/black explain target --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 770 source lines, 66 generated files, 14665 generated lines, ratio 19.05.
- `go run ./cmd/black benchmark eval ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black benchmark eval-history --json` geçti; summary runCount 1, modelCount 1, totalRuns 24, passedRuns 24, failedRuns 0.
- `go run ./cmd/black benchmark eval-history --ir` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 99, `testing-benchmarks` score 100.
- `go run ./cmd/black benchmark coverage --ir` geçti; issues 52.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 52 open 0 done 52.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Gerçek dış model skoru uydurulmadı; ilk manifest entry'si `provider local`, `source compiler-generated-checks` olarak açıkça işaretlendi.
- External multi-model scored eval board bağımsız harness koşuları ve release review sonrasına bırakıldı.
- Commit, push veya VPS/canlı site deploy yapılmadı.

## Multi-Editor Marketplace Metadata - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Compiler-owned editor bridge'in sadece VS Code dosyası olarak kalmasını engeller; VS Code Marketplace, Open VSX ve Cursor-compatible VSIX kanallarını aynı kaynak, aynı trust workflow ve aynı registry/marketplace metadata'sı üzerinden AI-readable hale getirir. Dış marketplace'e publish yapmadan, publish öncesi hangi package/adaptor ID'lerinin ve hangi kontrollerin geçerli olduğu deterministikleşti.

Tamamlananlar:

- `packages/registry/package-index.blackdir` multi-editor channel metadata'sıyla güncellendi: `vscode:blacklang-vscode`, `openvsx:blacklang-vscode`, `cursor:blacklang-vscode`.
- `adapters/marketplace/adapter-index.blackdir` editor adapter kanallarıyla güncellendi: `editor:vscode`, `editor:open-vsx`, `editor:cursor-compatible`.
- `black ecosystem --json|--ir` package, adapter, registry, marketplace ve trustWorkflow çıktıları yeni editor channel ID'lerini ve `npm run package:vsix` kontrolünü gösterecek şekilde güncellendi.
- `packages/registry/scripts/validate-registry.mjs` yeni package/adaptor marker'larını ve `VSIX package reuse` capability'sini doğrulayacak hale getirildi.
- Yeni `docs/editor-marketplace.md` eklendi; external publication'ın release-owner step olduğu, publish token/signing key/secrets'in `.black`, generated output, manifest veya package metadata içine yazılmaması gerektiği açıklandı.
- `black docs editor-marketplace --json` ve `black explain editor-marketplace --json` learning surfaces eklendi; package registry, adapter marketplace, release trust ve IDE docs ile ilişkilendirildi.
- README, `docs/README.md`, `docs/ide.md`, `docs/package-registry.md`, `docs/adapter-marketplace.md`, `docs/release-artifacts.md`, `docs/ai-agent-contract.md`, `docs/benchmark.md`, `docs/llms.txt`, `website/llms.txt`, `website/index.html`, ROADMAP dosyaları ve Türkçe web yol haritası multi-editor marketplace metadata'sıyla güncellendi.
- Coverage taxonomy güncellendi: `docs-ai-ergonomics` score 100/status `mature`; yeni `WEB-DOCS-005` issue'su done; issues summary total 53 open 0 done 53.
- `editors/vscode-blacklang/package.json` marketplace packaging kalitesi için repository metadata'sı ve paket içi `LICENSE` dosyasıyla güncellendi.
- `editors/vscode-blacklang/.gitignore` eklendi; lokal üretilen `.vsix` artefact'i kaynak kontrolüne karışmayacak.

Doğrulama:

- `gofmt -w` ilgili Go dosyalarında çalıştırıldı.
- `cd packages/cli && go test -count=1 ./cmd/black -run "TestEcosystem|TestFormatEcosystemIR|TestEcosystemManifestFilesAreReadable|TestFindEditorMarketplaceDoc|TestFindPackageRegistryDoc|TestFindAdapterMarketplaceDoc|TestExplainEditorMarketplaceKeyword|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestFormatCoverageIssuesIR|TestAllDocsReturnsSortedEntries"` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti; output package listesinde `openvsx:blacklang-vscode` ve `cursor:blacklang-vscode`, adapter listesinde `editor:open-vsx` ve `editor:cursor-compatible` var.
- `cd editors/vscode-blacklang && npm test` geçti.
- `cd editors/vscode-blacklang && npm run package:vsix` geçti; `blacklang-vscode-0.1.0.vsix` lokal üretildi ve `.gitignore` nedeniyle ignored.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black docs editor-marketplace --json` geçti.
- `go run ./cmd/black explain editor-marketplace --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black ecosystem --ir` geçti; packages 4, adapters 12.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 99, `docs-ai-ergonomics` score 100.
- `go run ./cmd/black benchmark coverage --ir` geçti; issues 53.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 53 open 0 done 53.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 53 open 0 done 53.
- `go run ./cmd/black benchmark eval-history --json` geçti; summary runCount 1, modelCount 1, totalRuns 24, passedRuns 24, failedRuns 0.
- `go run ./cmd/black benchmark eval-history --ir` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; `ready:false` beklenen env eksiklerini değer yazdırmadan raporladı.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Dış VS Code Marketplace/Open VSX/Cursor publish yapılmadı; bu faz sadece prepared-local metadata, docs ve packageability doğrulaması yaptı.
- Marketplace token'ı, signing key veya secret değeri `.black`, generated output, manifest ya da package metadata içine eklenmedi.
- Commit, push veya VPS/canlı site deploy yapılmadı.

## Public Ecosystem Index Source - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Package/provider adapter discovery için gelecekte host edilecek index'in deploy öncesi kaynağını hazırlar. Dış npm/editor registry yayını veya VPS/canlı site deploy yapmadan, `black ecosystem --json` tarafından üretilen `publicIndex` metadata'sı, `packages/registry/public-index.blackdir` manifesti ve `website/ecosystem-index.json` static snapshot'ı üzerinden public/hosted index yolu deterministikleşti.

Tamamlananlar:

- `EcosystemResult` içine `publicIndex` alanı eklendi.
- `black ecosystem --json` artık `public-index:blacklang-ecosystem` ID'sini, `website/ecosystem-index.json` path'ini, kaynak manifestleri, package/adaptor ID'lerini, trust notlarını ve validator komutunu raporluyor.
- `black ecosystem --ir` compact `public-index` bölümü üretmeye başladı; format/source/package/adapter/command satırları eklendi.
- Yeni `packages/registry/public-index.blackdir` manifesti eklendi; index path, formatlar, source manifestler, package/adaptor ID'leri, no-secret policy ve release-owner sonrası publish boundary açıklandı.
- `website/ecosystem-index.json`, `go run ./cmd/black ecosystem --json` çıktısından üretildi; elle yazılmış static kopya değil.
- `packages/registry/scripts/validate-registry.mjs` public-index manifestini ve `website/ecosystem-index.json` snapshot'ını doğrulayacak şekilde güncellendi; output'a `indexes: ["public-index:blacklang-ecosystem"]` eklendi.
- Ecosystem docs, package registry docs, adapter marketplace docs, AI agent contract, `llms.txt`, website `llms.txt`, README, roadmap, Türkçe web yol haritası ve docs site source public index metadata'sıyla güncellendi.
- Coverage taxonomy güncellendi: `ecosystem-integrations` score 95; yeni `WEB-ECO-004` done; weighted `completionPercent` 100; issues summary total 54 open 0 done 54.

Doğrulama:

- `gofmt -w` ilgili Go dosyalarında çalıştırıldı.
- `cd packages/cli && go test -count=1 ./cmd/black -run "TestEcosystem|TestFormatEcosystemIR|TestEcosystemManifestFilesAreReadable|TestWebCoverageReportIsWeightedAndActionable|TestFormatCoverageIR|TestWebCoverageIssuesReport|TestFormatCoverageIssuesIR|TestFindPackageRegistryDoc|TestFindAdapterMarketplaceDoc|TestExplainEcosystemKeyword|TestExplainPackageRegistryKeyword|TestExplainAdapterMarketplaceKeyword"` geçti.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti; manifests listesinde `packages/registry/public-index.blackdir` ve `website/ecosystem-index.json` var.
- `go run ./cmd/black --help` geçti; ecosystem help satırı public index metadata'sını anıyor.
- `go run ./cmd/black ecosystem --json` geçti; `publicIndex.id` = `public-index:blacklang-ecosystem`.
- `go run ./cmd/black ecosystem --ir` geçti; compact `public-index` bölümü üretildi.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100, `ecosystem-integrations` score 95.
- `go run ./cmd/black benchmark coverage --ir` geçti; completion 100, issues 54.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain ecosystem --json` geçti; `publicIndex.path` agent steps içinde göründü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- `completionPercent 100` weighted planning signal'dır; backend/API tarafındaki future cross-service orchestration ve dış registry/hosted deployment adımları hâlâ explicit review/release-owner sonrası track olarak görünür.
- Public index snapshot secret-free metadata içerir; provider credential, marketplace token, DSN, app name/region veya private signing key eklenmedi.
- Dış registry publish, VPS/canlı site deploy, commit veya push yapılmadı.

## Completion Audit / Stop Point - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Son kod/dokümantasyon fazlarından sonra repo durumunun bırakılabilir olup olmadığını kontrol eder. Amaç yeni syntax eklemek değil; stale coverage yüzdeleri, generated artefact durumu, public index snapshot doğruluğu ve son doğrulama sonuçlarını tek yerde kaydetmekti.

Audit sonucu:

- `black benchmark coverage --json` artık weighted `completionPercent` 100 raporluyor.
- `black benchmark issues --json` artık summary total 54 open 0 done 54 raporluyor.
- `docs-ai-ergonomics` score 100/status `mature`.
- `ecosystem-integrations` score 95/status `strong-mvp`; remaining items dış npm/editor registry publication ve hosted provider adapter index deploy, yani release-owner/review sonrası adımlar.
- `website/index.html` coverage kartı 100%, stable issue count 54 ve `WEB-ECO-004` ile hizalandı.
- `website/ecosystem-index.json` `black ecosystem --json` çıktısından yeniden üretildi ve validator tarafından okundu.
- Stale marker taramasında eski `%99` / `total 53` ifadeleri yalnızca tarihsel worklog kayıtlarında ve negatif test guard'larında kaldı; aktif docs/site/CLI output'ta eski state görünmüyor.
- `generated/` ignored durumda; `editors/vscode-blacklang/blacklang-vscode-0.1.0.vsix` lokal paket artefact'i ignored durumda.

Son doğrulama seti:

- `cd packages/cli && go test -count=1 ./...` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black ecosystem --json` geçti; `publicIndex.id` = `public-index:blacklang-ecosystem`.
- `go run ./cmd/black ecosystem --ir` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100.
- `go run ./cmd/black benchmark coverage --ir` geçti; completion 100, issues 54.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black benchmark issues --ir` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain ecosystem --json` geçti; `publicIndex.path` agent steps içinde göründü.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected target --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.
- Editor bridge `npm test` geçti.
- Editor bridge `npm run package:vsix` geçti; `.vsix` artefact'i ignored.

Stop point:

- Commit, push, VPS/canlı site deploy veya external marketplace/registry publish yapılmadı.
- Mevcut kurallarla bu nokta bırakılabilir: compiler/docs/registry/generated doğrulama temiz, weighted web coverage signal 100, tracked open issue 0.

## Core Program Logic Expression Foundation - 2026-09-07

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang web tarafında ilk Core Program Logic temelini atar: `computed`, custom `action set` ve explicit API update ifadeleri artık ortak, deterministic bir arithmetic expression tree üzerinden `+`, `-`, `*`, `/`, parantez ve öncelik mantığını anlayabilir. Bu, ileride Python-benzeri bloklu `if/else` ve named local value katmanının üstüne kurulacağı güvenli expression zeminidir.

Yapılanlar:

- `packages/cli/cmd/black/expression.go` ile ortak expression parser/formatter/JavaScript renderer eklendi.
- Lexer `(` ve `)` sembollerini tanıyacak şekilde genişletildi.
- AST tarafında computed field, action expression ve API expression için expression tree alanı eklendi.
- Parser tarafında computed/action/API numeric expression okuması recursive hale getirildi; eski tek operandlı ifade davranışı korunarak precedence ve parentheses desteği eklendi.
- Validator tarafında numeric-only arithmetic, unknown field/type diagnostics ve literal divide-by-zero kontrolleri expression tree üzerinde recursive çalışır hale getirildi.
- Generator tarafında computed alanlar, custom actions ve explicit API runtime update ifadeleri ortak renderer ile üretildi; nested division guard kontrolleri recursive hale getirildi.
- BlackIR ve affected analysis expression tree-aware hale getirildi.
- Docs/site/AI contract yüzeyi güncellendi: `docs/computed.md`, `docs/action.md`, `docs/api.md`, `docs/ai-agent-contract.md`, `BLACKLANG.md`, `SPEC.md`, roadmap dosyaları, `docs/llms.txt`, `website/llms.txt`, `website/index.html`.
- Yeni tests: computed precedence parse, nested custom action runtime generation ve nested explicit API runtime generation.

Doğrulama:

- `gofmt -w` ilgili Go dosyalarında çalıştırıldı.
- `cd packages/cli && go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti; `changed=false`.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti; hata yok.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti; hata yok.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain computed --json` geçti.
- `go run ./cmd/black explain action --json` geçti.
- `go run ./cmd/black explain api --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.inventoryValue --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı, eksik deployment env referansları deterministic şekilde listelendi.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.

Notlar:

- Generated dosyalar elle düzenlenmedi; çıktı generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- `if/else` henüz eklenmedi; kullanıcı tercihi Python-benzeri bloklu yapı yönünde netleşti. Sonraki Core Program Logic aşaması için önerilen sıra: named local values + Python-benzeri deterministic `if/else` blokları.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Core Program Logic v2 - Named Values And If/Else - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Custom action ve explicit API update handler içinde ordered local `value` declarations ve Python-like indentation-based `if`/`else` branch mantığını resmi, deterministic web capability olarak tamamlar. Tek syntax ile ara değer, koşullu set, branch-local scope, JSON/BlackIR/OpenAPI/affected visibility ve generated TypeScript runtime üretimi sağlar.

Yapılanlar:

- Action/API AST, parser, validator, formatter, BlackIR, affected analysis ve generator statement-tree aware hale getirildi.
- `value`, `if`, `else`, nested branch statements; action/API branch-local duplicate set validation; numeric/equality condition validation eklendi.
- Runtime generator local values, branch sets, division guards, transaction/non-transaction routes, API metadata ve field permission metadata ile çalışıyor.
- API/action direct numeric local values TypeScript için `Number(... ?? 0)` daraltması alıyor; nullable numeric source/body değerlerinden gelen ordered comparisons artık TS build'i bozmaz.
- Warehouse örneği `RestockProduct` ve `StockWebhook` ile yeni syntax'ı generator üzerinden gösteriyor.
- Docs/site/llms/roadmap/agent contract/diagnostics/coverage metadata güncellendi; `backend-api-actions` score 100 oldu.
- Reserved local value names case-insensitive validate ediliyor.

Doğrulama:

- `gofmt -w cmd/black/generator_action.go cmd/black/generator_api_runtime.go cmd/black/expression_test.go` çalıştırıldı.
- `go test -count=1 ./cmd/black -run "TestBuildWebNarrowsActionSourceNumericLocalValue|TestBuildWebNarrowsAPIDirectNumericLocalValue|TestBuildWebGeneratesAPIValueIfElseRuntime|TestBuildWebGeneratesActionValueIfElseRuntime|TestValidateRejectsReservedLogicValueNamesCaseInsensitive"` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti; `changed=false`.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti; hata yok.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti; hata yok.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain action --json` geçti.
- `go run ./cmd/black explain api --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockWebhook --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100, `backend-api-actions` score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 781 BlackLang source lines -> 14,680 generated lines, 66 files, 18.8x generated/source line ratio.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı, eksik required env listesi deterministic çıktı.
- Generated Warehouse `npm run security:secrets:preflight --silent` geçti; read-only provider preflight secret değerleri fetch/log/write etmedi.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:plan` geçti; Chrome ve Edge available, custom/Chromium optional skipped.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.
- Generated Warehouse izole `DATABASE_URL=file:./blacklang-jobs-smoke.db npm run jobs:run` geçti; `LowStockMonitor` query job count 1 döndü.
- `git diff --check` geçti; sadece mevcut repo line-ending uyarıları göründü.

Notlar:

- Generated dosyalar elle düzenlenmedi; çıktı generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## License And Trademark Hygiene - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

MIT lisansının kod kullanım izni verdiğini, fakat BlackLang adını/markasını otomatik olarak devretmediğini repo içinde netleştirir. Copyright holder olarak tüzel kişi olmayan `BlackLang` yerine gerçek kişi adı kullanılır; böylece GitHub üzerindeki lisans bildirimi daha temiz ve anlaşılır hale gelir.

Yapılanlar:

- Root `LICENSE` dosyasında copyright holder `Muhammet Enes Burul` olarak güncellendi.
- `editors/vscode-blacklang/LICENSE` dosyasında aynı holder düzeltmesi yapıldı.
- Root `TRADEMARKS.md` eklendi; MIT kod lisansı ile BlackLang ad/marka kullanımı ayrımı kısa ve ajan-okunur şekilde belgelendi.
- `README.md` içine `License And Brand` bölümü ve `TRADEMARKS.md` bağlantısı eklendi.

Doğrulama:

- `rg -n "Copyright \(c\) 2026 BlackLang" .` taraması sonucunda eski holder kalmadı.
- `rg -n "Copyright \(c\) 2026|Trademark And Brand Notice|BlackLang source code is licensed" LICENSE README.md TRADEMARKS.md editors/vscode-blacklang/LICENSE` beklenen yeni holder ve README/TRADEMARKS bağlantılarını gösterdi.
- `git diff --check` geçti; sadece mevcut repo line-ending uyarıları göründü.

Notlar:

- MIT lisans türü değiştirilmedi.
- Hukuki tavsiye veya marka tescili iddiası eklenmedi; sadece repo içi kullanım sınırı dokümante edildi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## State Documentation Closure - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Mevcut `state` capability'sini ayrı, kısa ve ajan-okunur bir referans sayfasına taşır. Böylece UI state'in ne olduğu, ne ürettiği ve ne olmadığı (`calculator` tarzı local expression state veya arbitrary browser runtime değil) netleşir.

Yapılanlar:

- `docs/state.md` eklendi; syntax, validation rules, generated React output, AI agent notes ve diagnostics listelendi.
- `README.md` dokümantasyon listesine `State Declarations` bağlantısı eklendi.
- `docs/README.md` plan/structure listesine `state.md` ve client state açıklaması eklendi.
- `docs/llms.txt` içine kısa `state` kontratı eklendi.
- `website/llms.txt` içine `/state.md` linki ve state boundary notu eklendi; numaralı agent checklist 1..54 aralığında çakışmasız hale getirildi.
- `black docs state --json` agent notes içine state'in computed expressions, arbitrary click handlers, loops veya browser-side BlackLang execution tanımlamadığı belirtildi.

Doğrulama:

- `gofmt -w cmd/black/docs.go` çalıştırıldı.
- `go test -count=1 ./cmd/black -run "TestDocs|TestExplain|TestAgentContract|TestBlackIRIncludesState|TestValidateStateDeclaration"` geçti.
- `go run ./cmd/black docs state --json` geçti ve yeni boundary notunu gösterdi.
- `website/llms.txt` numara kontrolü geçti; count 54, first 1, last 54, duplicate yok.
- `git diff --check` geçti; sadece mevcut repo line-ending uyarıları göründü.

Notlar:

- Yeni `state` syntax'ı eklenmedi; yalnızca mevcut capability belgelendi ve CLI docs sınırı netleştirildi.
- Generated dosyalar elle düzenlenmedi.
- JSON output modları korundu.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Final Local Readiness Audit - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Son docs/metadata dokunuşlarından sonra commit/push/deploy öncesi yerel çalışma ağacının test ve doğrulama kanıtını tazeler.

Doğrulama:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti; `changed=false`.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `git status --short` alındı; önceki aşamalardan kalan geniş modified/untracked çalışma ağacı duruyor.
- `git diff --stat` alındı; tracked diff 67 files, 24,324 insertions, 2,057 deletions seviyesinde.

Notlar:

- Bu audit yeni build/deploy üretmedi; generated full suite Core Program Logic v3 kaydında zaten taze geçti.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Release Candidate Preparation - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Büyük web-completion değişiklik setini commit/push/deploy öncesi incelenebilir hale getirir. Geçici dosyaları ayıklar, agent startup ergonomisini düzeltir, registry/editor/package yüzeylerini doğrular ve önerilen commit gruplarını çıkarır.

Yapılanlar:

- `git status --short`, `git diff --name-status`, `git diff --stat` ve untracked dosya listesi alındı.
- `.tmp_openapi.txt`, `.tmp_postgresauth.txt`, `.tmp_sqliteauth.txt`, `.tmp_userspage.txt` dosyaları eski scratch çıktıları olarak doğrulandı ve kaldırıldı.
- `.gitignore` içine `.tmp_*` kuralı eklendi; kökte kalan scratch dosyaların RC listesine dönmesi engellendi.
- `agent startup` read-first path çözümlemesinde alt klasörden çalıştırma ergonomi hatası düzeltildi. `go -C packages/cli run ./cmd/black agent startup ../../examples/warehouse/app.black --json` artık root `AGENTS.md`, `blacklang.toml`, `BLACKLANG.md`, `SPEC.md` ve `docs/diagnostics.md` dosyalarını `exists=true` gösteriyor.
- `agent startup` için subdirectory regression testi eklendi.
- Untracked dosyalar sınıflandırıldı: `adapters` 3, `benchmarks` 1, `docs` 26, `editors` 10, `packages` 70, `scripts` 1, `website` 1, ayrıca `TRADEMARKS.md` ve `WEB-COMPLETION-WORKLOG.md`.
- Önerilen commit grupları çıkarıldı:
  - Core compiler/runtime: parser, validator, AST, BlackIR, affected graph, generator, action/API/expression/query/security/theme/ops/seed/service/test modules.
  - Generated web capability examples and golden: `examples/warehouse/app.black`, benchmark updates, golden manifest.
  - Docs and website: `README.md`, `SPEC.md`, `BLACKLANG.md`, `docs/*`, `website/*`, roadmap files.
  - Ecosystem/release/editor/package: `packages/npm`, `packages/registry`, `adapters/marketplace`, `editors/vscode-blacklang`, release trust script.
  - License/brand hygiene: `LICENSE`, `editors/vscode-blacklang/LICENSE`, `TRADEMARKS.md`.
  - RC hygiene: `.gitignore`, `WEB-COMPLETION-WORKLOG.md`.

Doğrulama:

- `gofmt -w cmd/black/agent.go cmd/black/agent_test.go` çalıştırıldı.
- `go test -count=1 ./cmd/black -run "TestAgentStartupChecklist"` geçti.
- `go -C packages/cli run ./cmd/black agent startup ../../examples/warehouse/app.black --json` geçti; readFirst root dosyaları `exists=true`.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black agent startup ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti; `changed=false`.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti; format/parse/validate/security bulgusu 0.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain action --json` geçti.
- `go run ./cmd/black explain api --json` geçti.
- `go run ./cmd/black explain state --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti; findings 0.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest` geçti.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `node packages/npm/scripts/validate-package.mjs` geçti.
- `node editors/vscode-blacklang/scripts/validate-extension.mjs` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti; Prisma 7.10.0 -> 8.0.0-rc.13 update notice gösterdi, test sonucu başarılı.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; gerçek deploy env olmadığı için `ready=false`, secret value yazdırmadı.
- Generated Warehouse `npm run security:secrets:preflight --silent` geçti; read-only, secret value fetch/log/write yok.
- Generated Warehouse `npm run test:e2e:plan` geçti; Chrome ve Edge available, custom/Chromium optional skipped.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, failed 0.
- Generated Warehouse izole `DATABASE_URL=file:./blacklang-jobs-smoke.db npm run jobs:run` geçti; `LowStockMonitor` count 1.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100.
- `go run ./cmd/black benchmark issues --json` geçti; total 54, open 0, done 54.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 781 BlackLang source lines -> 14,680 generated lines, 66 files, 18.8x generated/source line ratio.
- `git diff --check` geçti; sadece mevcut line-ending uyarıları göründü.
- `git ls-files --others --exclude-standard | Select-String -Pattern '^\.tmp'` boş döndü.

Notlar:

- Generated dosyalar elle düzenlenmedi; warehouse output generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Core Program Logic v3 - Compound Conditions - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Custom action ve explicit API update handler içindeki `if` koşullarını tek karşılaştırmadan çıkarıp `and`, `or`, `not` ve parantez destekli deterministic condition tree haline getirir. Böylece gereksiz nested `if` yazmadan daha okunaklı iş kuralı ifade edilir; generated TypeScript tarafında `&&`, `||`, `!` ve doğru parantezleme ile güvenli runtime üretilir.

Yapılanlar:

- Action/API `if` parser katmanı shared `ConditionExpressionDecl` tree ile güncellendi; comparison, `and`, `or`, `not` ve boolean parantezleri destekleniyor.
- `not > and > or` precedence kuralı resmi hale getirildi; arithmetic expression parantezleri ile boolean grouping ayrıştırıldı.
- Action/API validator compound condition tree içinde tüm karşılaştırmaları recursive doğruluyor; ordered comparisons yine numeric değerlere bağlı.
- Runtime generator compound condition ifadelerini TypeScript boolean expression olarak üretiyor; division guard toplama logic'i condition tree üzerinden çalışıyor.
- BlackIR condition formatter compound yapıyı tekrar BlackLang syntax'ına deterministic basıyor.
- API expression BlackIR prefix bug'ı düzeltildi; `body.stock` gibi operandlar artık `stock` diye kısalmıyor.
- `inspect --affected` action/API condition tree içindeki alan referanslarını görecek şekilde güncellendi.
- Warehouse örneğinde `RestockProduct` ve `StockWebhook` compound condition kullanacak şekilde source üzerinden güncellendi; generated çıktı generator ile yenilendi.
- Docs/site/llms/roadmap/agent contract/diagnostics/coverage metadata yeni `if condition` sözleşmesini ve `and`/`or`/`not` kullanımını anlatacak şekilde güncellendi.
- Kullanıcı kararıyla MIT lisans copyright holder satırı `Copyright (c) 2026 Muhammet Enes Burul` olarak düzeltildi.

Doğrulama:

- `gofmt -w cmd/black/expression.go cmd/black/action.go cmd/black/api_handler.go cmd/black/generator_action.go cmd/black/generator_api_runtime.go cmd/black/blackir.go cmd/black/affected.go cmd/black/expression_test.go cmd/black/coverage.go cmd/black/coverage_test.go cmd/black/docs.go` çalıştırıldı.
- Stale docs/syntax taraması yapıldı; kalan tek `if incomingStock >= 0` kullanımı bilinçli tek-koşullu regression test örneği.
- Focused compound/value/coverage testleri geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti; 66 generated file raporlandı.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest` geçti.
- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti; `changed=false`.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti; format/parse/validate/security bulgusu 0.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain action --json` geçti.
- `go run ./cmd/black explain api --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockWebhook --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected Product.stock --json` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100, `backend-api-actions` score 100.
- `go run ./cmd/black benchmark issues --json` geçti; summary total 54 open 0 done 54.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 781 BlackLang source lines -> 14,680 generated lines, 66 files, 18.8x generated/source line ratio.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti; secret değerleri yazdırılmadı, eksik required env listesi deterministic çıktı.
- Generated Warehouse `npm run security:secrets:preflight --silent` geçti; read-only provider preflight secret değerleri fetch/log/write etmedi.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:plan` geçti; Chrome ve Edge available, custom/Chromium optional skipped.
- Generated Warehouse `npm run test:e2e:matrix` geçti; Chrome ve Edge passed, custom/Chromium optional skipped.
- Generated Warehouse izole `DATABASE_URL=file:./blacklang-jobs-smoke.db npm run jobs:run` geçti; `LowStockMonitor` query job count 1 döndü.
- `git diff --check` geçti; sadece mevcut repo line-ending uyarıları göründü.

Notlar:

- Generated dosyalar elle düzenlenmedi; çıktı generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Website Documentation Index Closure - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Canlı dokümantasyon sitesi commit/push/deploy öncesinde bütün yeni Markdown dokümanları sitemap ve AI başlangıç indekslerinde görünür hale getirir. Böylece insan arama motorları, botlar ve AI ajanları yeni web capability belgelerini eksiksiz keşfedebilir.

Yapılanlar:

- `website/sitemap.xml` gerçek `docs/*.md` yayın yüzeyiyle karşılaştırıldı.
- Sitemap içinde eksik kalan 13 doküman route'u eklendi: adapter marketplace, ecosystem, editor marketplace, GitHub publish, IDE, install, npm wrapper, package registry, relation load, release artifacts, release trust, secret management ve state.
- `docs/llms.txt` core local files listesi 43 Markdown dokümanı kapsayacak şekilde tamamlandı.
- `website/llms.txt` kök Markdown route listesi 43 public docs sayfasını kapsayacak şekilde tamamlandı.
- `website/llms.txt` içinde `release-trust` canlı sitedeki diğer Markdown sayfalarıyla aynı kök path biçimine çekildi.

Doğrulama:

- Node tabanlı statik indeks kontrolü geçti: docsCount 43, sitemapDocCount 43, sitemap missing/stale/duplicate 0, docs llms missing/stale 0, website llms stale 0.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black docs agent-contract --json` geçti.
- `git diff --check` geçti; sadece mevcut line-ending uyarıları göründü.

Notlar:

- Generated dosyalar elle düzenlenmedi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## Final RC Green Sweep - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Release candidate çalışma ağacının son statik site indeks düzeltmelerinden sonra hâlâ compiler, generated web, e2e, benchmark ve paket/registry yüzeylerinde yeşil olduğunu kanıtlar. Bu aşama commit/push/deploy yapmadan, review öncesi son yerel güven noktasını oluşturur.

Yapılanlar:

- Compiler testleri yeniden çalıştırıldı.
- Static docs index kontrolü yeniden çalıştırıldı.
- Registry, npm wrapper ve VS Code extension manifest validator kontrolleri yeniden çalıştırıldı.
- CLI agent startup, format, lint, validate, docs, explain, inspect, accessibility audit, ecosystem, build ve golden manifest akışı yeniden çalıştırıldı.
- Generated Warehouse uygulamasında build, test, secret plan/preflight, e2e plan, e2e, e2e matrix ve job runner yeniden çalıştırıldı.
- Coverage, issue export ve measured source/generated benchmark yeniden alındı.
- Geçici `.tmp_*` dosya kalmadığı doğrulandı.

Doğrulama:

- `go test -count=1 ./...` geçti.
- Static docs index kontrolü geçti: 43 docs, 43 sitemap doc route, missing/stale/duplicate 0; `docs/llms.txt` ve `website/llms.txt` 43 public docs yüzeyiyle uyumlu.
- `node packages/registry/scripts/validate-registry.mjs` geçti.
- `node packages/npm/scripts/validate-package.mjs` geçti.
- `node editors/vscode-blacklang/scripts/validate-extension.mjs` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black agent startup ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black docs --all --json` geçti; count 73.
- `go run ./cmd/black explain action --json`, `explain api --json`, `explain state --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected RestockProduct --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --affected StockWebhook --json` geçti.
- `go run ./cmd/black audit accessibility ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black ecosystem --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `UPDATE_BLACKLANG_GOLDEN=1 go test -count=1 ./cmd/black -run TestWarehouseGeneratedGoldenManifest` geçti.
- Generated Warehouse `npm run build` geçti.
- Generated Warehouse `npm test` geçti.
- Generated Warehouse `npm run security:secrets:plan --silent` geçti.
- Generated Warehouse `npm run security:secrets:preflight --silent` geçti.
- Generated Warehouse `npm run test:e2e:plan` geçti.
- Generated Warehouse `npm run test:e2e` geçti.
- Generated Warehouse `npm run test:e2e:matrix` geçti.
- Generated Warehouse izole `DATABASE_URL=file:./blacklang-jobs-smoke.db npm run jobs:run` geçti.
- `go run ./cmd/black benchmark coverage --json` geçti; completionPercent 100.
- `go run ./cmd/black benchmark issues --json` geçti; total 54, open 0, done 54.
- `go run ./cmd/black benchmark ../../examples/warehouse/app.black --out ../../generated --json` geçti; 781 BlackLang source lines -> 14,680 generated lines, 66 files, 18.8x generated/source line ratio.
- `git diff --check` geçti; sadece mevcut line-ending uyarıları göründü.
- `git ls-files --others --exclude-standard | Select-String -Pattern '^\.tmp'` boş döndü.

Notlar:

- Generated dosyalar elle düzenlenmedi; warehouse output generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push, VPS/canlı site deploy veya external publish yapılmadı.

## GitHub Main Update - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

Release candidate değişiklik setini yerel çalışma ağacından GitHub `main` branch'ine taşır. Böylece compiler, docs, website source, registry/editor/npm metadata ve worklog tek resmi repo geçmişinde görünür hale gelir.

Yapılanlar:

- `git fetch origin` çalıştırıldı.
- `main...origin/main` ahead/behind durumu kontrol edildi; sonuç `0 0`.
- `generated/` altında tracked dosya olmadığı doğrulandı.
- Secret pattern taraması yapıldı; sadece generator içindeki runtime değişken adları ve roadmap token hesap satırları eşleşti, gerçek secret bulunmadı.
- Tüm release candidate değişiklikleri stage edildi.
- `docs/api.md` ve `docs/migration.md` dosyalarındaki fazla EOF boş satırları temizlendi.
- Tek ana commit oluşturuldu: `bd28674 Complete BlackLang web capability RC`.
- Commit `origin/main` üzerine push edildi.

Doğrulama:

- `git diff --check` stage öncesi ve stage sonrası geçti; sadece line-ending uyarıları kaldı.
- `git push origin main` geçti; `a68e1be..bd28674 main -> main`.

Notlar:

- Generated dosyalar commit edilmedi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Bu worklog kaydı commit sonrası eklendi; canlı site deploy kaydıyla beraber ayrı küçük takip commit'i olarak kaydedildi.

## Live Documentation Site Deploy - 2026-09-08

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

GitHub'a taşınan release candidate dokümantasyon sitesini VPS üzerindeki canlı `https://black.muenspeak.com` yayınına geçirir. Böylece insanlar ve AI ajanları yeni web capability dokümanlarına, sitemap'e, `llms.txt` dosyasına ve public ecosystem index'e canlı domain üzerinden erişebilir.

Yapılanlar:

- Local staging paketi hazırlandı: `website/` içeriği ve `docs/*.md` public dokümanları site kökü formatında birleştirildi.
- Staging paketi 51 root file, 43 Markdown dokümanı ve 4 asset dosyası içerdi.
- Paket `/tmp/blacklang-site-20260908193717.tar.gz` olarak VPS'ye yüklendi.
- Mevcut canlı site `/tmp/blacklang-site-backup-20260908193717` altında yedeklendi.
- Yeni paket `/var/www/black.muenspeak.com` içine kopyalandı.
- Site dosya izinleri public static serve için düzeltildi: dizinler `755`, dosyalar `644`.
- Nginx site config'ine Markdown route'ları için `text/markdown` MIME davranışı eklendi.
- `nginx -t` geçti ve Nginx reload edildi.

Doğrulama:

- Remote site root sayımı geçti: 51 root file, 43 Markdown file, 4 asset file.
- `curl -I https://black.muenspeak.com/` 200 döndü.
- `https://black.muenspeak.com/sitemap.xml` içinde `state.md`, `relation-load.md`, `release-trust.md` ve `secret-management.md` göründü.
- `https://black.muenspeak.com/llms.txt` içinde aynı yeni docs route'ları göründü.
- `curl -I https://black.muenspeak.com/state.md` 200 ve `Content-Type: text/markdown` döndü.
- `curl -I https://black.muenspeak.com/relation-load.md` 200 ve `Content-Type: text/markdown` döndü.
- `curl -I https://black.muenspeak.com/diagnostics.md` 200 ve `Content-Type: text/markdown` döndü.
- `curl -I https://black.muenspeak.com/ecosystem-index.json` 200 ve `Content-Type: application/json` döndü.
- `curl -I https://black.muenspeak.com/assets/favicon.png` 200 ve `Content-Type: image/png` döndü.
- `https://black.muenspeak.com/no-such-doc.md` beklenen şekilde 404 döndü; Markdown route'ları artık SPA fallback'e düşmüyor.

Notlar:

- Canlı site static source üzerinden güncellendi; generated app output elle düzenlenmedi.
- VPS'ye secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Bu deploy kaydı ve önceki GitHub publish kaydı ayrı takip commit'i olarak kaydedildi.

## Public Bootstrap / External Agent Onboarding - 2026-09-11

Bu aşama ne işe yarıyor / neyi mümkün kılıyor?

BlackLang kurulu olmayan bir bilgisayarda veya sadece canlı dokümantasyon sitesini okuyan dış AI ajanında ilk adımı netleştirir. Ajan artık syntax tahmin etmek yerine GitHub reposunu klonlayıp `packages/cli` içinden `go run ./cmd/black ...` ile resmi compiler/docs/validate/build akışını çalıştırması gerektiğini görür.

Yapılanlar:

- `docs/install.md` içine güncel public source bootstrap akışı eklendi.
- `docs/ai-agent-contract.md` içine no-installed-CLI bootstrap kuralı eklendi.
- Calculator öğrenme sınırı netleştirildi: entity `computed` alanlarıyla aritmetik örnek yapılabilir, fakat button-driven browser calculator ve `validate rightValue != 0` tarzı literal-comparison entity validation henüz resmi syntax değildir.
- `packages/cli/cmd/black/docs.go` içindeki `black docs agent-contract --json` çıktısı aynı bootstrap ve calculator sınırını anlatacak şekilde güncellendi.
- `packages/cli/cmd/black/docs_test.go` agent-contract docs testleri source bootstrap ve unsupported literal validation uyarısını da arayacak şekilde genişletildi.
- `website/index.html` kurulum bölümü bugünkü güvenilir yolu GitHub repo + Go source bootstrap olarak gösterecek şekilde güncellendi.
- `website/index.html` Agent Contract bölümü, `black` yoksa ajanların syntax uydurmadan source bootstrap yolunu kullanması gerektiğini anlatacak şekilde güncellendi.
- `website/llms.txt`, `docs/llms.txt` ve `README.md` dış AI ajanlarının hızlı okuyacağı bootstrap notuyla güncellendi.

Doğrulama:

- `go test -count=1 ./...` geçti.
- `go run ./cmd/black --help` geçti.
- `go run ./cmd/black docs agent-contract --json` geçti ve source bootstrap + calculator validation uyarılarını içerdi.
- `go run ./cmd/black docs --all --json` geçti.
- `go run ./cmd/black format --check --json ../../examples/warehouse/app.black` geçti.
- `go run ./cmd/black lint ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black validate ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black inspect ../../examples/warehouse/app.black --json` geçti.
- `go run ./cmd/black build ../../examples/warehouse/app.black --out ../../generated --json` geçti.
- `npm test` geçti.
- `npm run build` geçti.
- `npm run test:e2e` ilk paralel matrix denemesiyle aynı SQLite dosyasına çarpıştığı için başarısız göründü; tek başına tekrar çalıştırıldığında geçti.
- `npm run test:e2e:matrix` ilk paralel e2e denemesiyle aynı SQLite dosyasına çarpıştığı için başarısız göründü; tek başına tekrar çalıştırıldığında Chrome ve Edge için geçti, optional custom/chromium hedefleri beklendiği gibi skip edildi.
- `git diff --check` geçti; yalnızca mevcut line-ending uyarıları göründü.

Notlar:

- Generated dosyalar elle düzenlenmedi; warehouse output generator üzerinden yenilendi.
- `.black` içine secret yazılmadı.
- JSON output modları korundu.
- Tek davranış için alternatif syntax eklenmedi.
- Commit, push veya canlı site deploy henüz yapılmadı.
