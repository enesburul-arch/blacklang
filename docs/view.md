# BlackLang View Composition

`view` is a page-level block that controls generated section order, reusable component sections, and first layout composition inside one page.

It exists so humans and AI agents can move, size, tab, optionally overlay generated page parts, and place declared components as page panels from `.black` source instead of editing generated React or CSS.

## Syntax

```black
component StockBadge {
  input stock number
  variant low when stock < 10
  variant normal when stock >= 10
}

page Products {
  source Product

  view {
    order table, StockSummary, StockCards, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 2
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    section detail display drawer side right title "Product Details"
    section form display modal title "Product Form"
    trigger detail on rowSelect
    trigger form on createStart
    trigger form on editStart
    trigger detail on saveSuccess
    trigger table on close
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }
}
```

## Current Sections

Draft v0.2 supports these page sections:

```text
table
detail
form
declared component sections
```

Listed sections render first. Omitted supported sections are appended in the default order:

```text
table, detail, form
```

Declared component sections are appended after the built-in default order when omitted from `order`.

For example:

```black
view {
  order form
}
```

means:

```text
form, table, detail
```

For example:

```black
view {
  order table, StockSummary
  section StockSummary component StockBadge bind selected
}
```

means:

```text
table, StockSummary, detail, form
```

## Composition

`compose` declares the page section layout mode.

```black
compose stack gap md
compose grid columns 2 gap md stackAt md
compose tabs gap md
```

Supported MVP modes:

```text
stack   vertical flow
        optional gap sm|md|lg

grid    CSS grid flow
        optional columns 1..4, default 2
        optional gap sm|md|lg, default md
        optional stackAt sm|md|lg|none, default md
        stackAt controls generated responsive breakpoint media queries

tabs    generated React tab controls
        optional gap sm|md|lg, default md
        columns and stackAt are not supported
```

`section` declares per-section display options and component sections:

```black
section table span 2
section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
section StockCards component StockBadge bind each span 1 title "Stock Cards"
section detail display drawer side right title "Product Details"
section form display modal title "Product Form"
```

Supported section options:

```text
component <ComponentName>, only for custom component sections
bind     selected|first|each, required for component sections
span     1..4, useful only with compose grid
display  inline|modal|drawer, default inline
side     left|right, only valid with display drawer, default right
title    string, used as the generated modal/drawer panel heading
```

`display inline` keeps the section in normal page flow. `display drawer` is supported for `detail` and `form`; `display modal` is supported for `form` and `detail`. `table` stays inline in this MVP so list navigation remains visible and deterministic.

When a `detail` section uses `display drawer` or `display modal`, the generated page opens it when a record is selected or loaded through the read/detail flow. When a `form` section uses `display drawer` or `display modal`, generated create and edit controls open that panel, and cancel/reset closes it.

A component section uses a custom section name and renders one declared component as a generated panel:

```black
section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
section FirstStock component StockBadge bind first
section StockCards component StockBadge bind each span 1 title "Stock Cards"
```

`bind selected` passes the current selected detail record to the component. `bind first` passes the first loaded list record. `bind each` renders one component instance for every loaded list/query record on the page. Component inputs bind by matching the input name and type to stored or computed fields on the page source entity. Component section inputs are scalar primitive values in this MVP; list inputs and entity/object inputs are rejected for page component sections. Component sections render inline and are hidden when their required field permissions are hidden.

`group` declares a first nested layout wrapper around contiguous generated sections:

```black
group RecordWorkspace sections detail, form compose stack gap md span 1 title "Record Workspace"
group RecordGrid sections detail, form compose grid columns 2 gap sm span 2
```

Supported group options:

```text
sections  table|detail|form or declared component section names, contiguous in generated render order
compose   stack|grid, default stack
columns   1..4, only useful with compose grid
gap       sm|md|lg, default md
span      1..4, applies to the group in an outer grid
title     string, renders a generated group heading
```

Groups are valid on stack/grid pages. `compose tabs` already groups sections behind tab controls, so tabs and `group` are intentionally separate in this MVP. Sections that use `display modal` or `display drawer` cannot be placed inside a group yet; keep grouped sections inline.

`trigger` declares deterministic generated section interactions:

```black
trigger detail on rowSelect
trigger form on createStart
trigger form on editStart
trigger detail on saveSuccess
trigger table on saveSuccess
trigger table on close
```

Supported trigger events:

```text
rowSelect     after a table row is loaded through the generated View action
createStart   when the generated New button starts a create flow
editStart     when the generated Edit button starts an edit flow
saveSuccess   after generated create or update succeeds
close         when generated cancel/reset or overlay close runs
```

Trigger targets are generated sections, not arbitrary frontend handlers. `trigger detail on rowSelect` selects or switches to the detail section after a row loads. `trigger form on createStart` and `trigger form on editStart` select or open the form section. `trigger detail on saveSuccess` keeps the saved record selected; `trigger table on saveSuccess` returns to the list section. `trigger table on close` returns tabbed or overlay flows to the list section.

For grid composition, `stackAt` is the canonical responsive syntax. The generator emits named breakpoint CSS from the declared column count:

```text
lg  1024px
md   768px
sm   640px
```

When `stackAt sm` is used with `columns 4`, generated CSS steps from 4 columns to 3 at `lg`, 2 at `md`, and 1 at `sm`. `stackAt md` collapses at `md`, `stackAt lg` collapses at `lg`, and `stackAt none` disables generated grid breakpoint rules. If a section span is wider than the responsive column count, the generator clamps that span at the breakpoint.

`tab` declares which generated sections appear under one tab:

```black
compose tabs gap md
tab List sections table
tab Record sections detail, form
```

Tabs are valid only with `compose tabs`. Every ordered section must appear in exactly one tab. The generator emits local React tab state and tab controls. Modal/drawer section display is intentionally not combined with `compose tabs` in this MVP; use inline tab sections or switch the page to stack/grid composition.

## Generated Web Output

The web generator adds stable classes:

```text
page       .page-view .page-view-products .bl-view-compose-grid
modal page .page-view .page-view-products .bl-view-has-overlay
tabs page  .page-view .page-view-products .bl-view-compose-tabs

table      .bl-view-section-table .bl-view-span-2
detail     .bl-view-section-detail .bl-view-display-drawer .bl-view-side-right
form       .bl-view-section-form .bl-view-display-modal
group      .view-group .bl-view-group-record-workspace .bl-view-group-compose-stack .bl-view-span-1
component  .bl-view-component-section .bl-view-component-stock-badge
triggers   .page-view .bl-view-has-triggers
```

Generated `src/pages/<Page>Page.tsx` renders sections in the effective `view.order`, including reusable component sections. This keeps keyboard, screen-reader, and test traversal aligned with the source intent.

Generated `src/styles.css` still receives deterministic order metadata, composition, tabs, and overlay rules:

```css
.page-view-products.bl-view-compose-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: start;
  gap: 16px;
}

.page-view-products .bl-view-section-table {
  order: 1;
  grid-column: span 2;
}

.page-view-products .bl-view-section-stock-summary {
  order: 2;
  grid-column: span 1;
}

.section-overlay-modal {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
}

.section-overlay-drawer-right {
  position: fixed;
  inset: 0 0 0 auto;
}

.page-view-products .bl-view-group-record-workspace {
  display: flex;
  flex-direction: column;
  gap: 16px;
  order: 2;
}
```

With `compose grid columns 4 gap md stackAt sm`, the generator emits a larger ladder:

```css
@media (max-width: 1024px) {
  .page-view-products.bl-view-compose-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .page-view-products.bl-view-compose-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .page-view-products.bl-view-compose-grid {
    grid-template-columns: 1fr;
  }

  .page-view-products.bl-view-compose-grid .panel {
    grid-column: 1 / -1;
  }
}
```

For `compose tabs`, generated `src/pages/<Page>Page.tsx` includes:

```tsx
const [activeViewTab, setActiveViewTab] = useState("List");

<nav className="view-tabs" role="tablist" aria-label="Products view sections">
  <button className={activeViewTab === "List" ? "active" : ""} type="button" role="tab" aria-selected={activeViewTab === "List"} onClick={() => setActiveViewTab("List")}>List</button>
</nav>
```

With `trigger detail on rowSelect`, generated row View actions can switch the active tab after loading the record:

```tsx
const item = await productApi.get(id);
setSelectedItem(item);
setActiveViewTab("Record");
```

With a component section, generated `src/pages/<Page>Page.tsx` imports the component and binds it to the declared source record context:

```tsx
const stockSummaryComponentItem = selectedItem;
const stockCardsComponentItems = items;

<section className="panel bl-view-section-stock-summary bl-view-component-section bl-view-component-stock-badge bl-view-span-1">
  <h2>Stock Summary</h2>
  {permissions.fields.stock !== false ? (
    stockSummaryComponentItem ? (
      <div className="component-section-body">
        <StockBadge stock={Number(stockSummaryComponentItem.stock ?? 0)} />
      </div>
    ) : (
      <p className="muted">Select a record to render StockBadge.</p>
    )
  ) : (
    <p className="muted">Component hidden by field permissions.</p>
  )}
</section>

<section className="panel bl-view-section-stock-cards bl-view-component-section bl-view-component-stock-badge bl-view-span-1">
  <h2>Stock Cards</h2>
  {permissions.fields.stock !== false ? (
    stockCardsComponentItems.length > 0 ? (
      <div className="component-section-list" role="list" aria-label="Stock Cards">
        {stockCardsComponentItems.map((item) => (
          <div className="component-section-list-item" role="listitem" key={String(item.id)}>
            <StockBadge stock={Number(item.stock ?? 0)} />
          </div>
        ))}
      </div>
    ) : (
      <p className="muted">Load records to render StockBadge.</p>
    )
  ) : (
    <p className="muted">Component hidden by field permissions.</p>
  )}
</section>
```

## Rules

- Use `view` only inside a `page` block.
- Use one `view` block per page.
- Use one `order` line inside `view`.
- Use one `compose` line inside `view`.
- Use `section <name> span <columns>` for per-section span intent.
- Use `section <Name> component <Component> bind selected|first|each` to place a declared component as a generated page section.
- Use `section <name> display inline|modal|drawer` for generated section display intent.
- Use `side left|right` only with `display drawer`.
- Use `title "Text"` when the generated modal/drawer heading should differ from the default panel heading.
- Use `group <Name> sections <section...>` to wrap contiguous inline sections in one nested generated layout container.
- Use `compose stack|grid`, `columns`, `gap`, `span`, and `title` as group options when the nested wrapper needs local layout.
- Use `tab <Name> sections <section...>` for tabs intent when `compose tabs` is active.
- Use `trigger <section> on <event>` for supported generated section interactions.
- Current supported section names are `table`, `detail`, `form`, and declared component section names.
- Component sections must use custom section names, not `table`, `detail`, or `form`.
- Component sections require `bind selected`, `bind first`, or `bind each`.
- Component section inputs must be scalar primitive component inputs matching stored or computed fields on the page source entity by name and type.
- Component sections render inline in this MVP.
- Duplicate section names in `order` or `section` are validation errors.
- Duplicate section options are parser errors.
- Duplicate group names and duplicate grouped sections are validation errors.
- Duplicate tab names and duplicate tab sections are validation errors.
- Duplicate trigger events are validation errors; keep one target section per event.
- Unsupported section names are validation errors.
- Unknown component names, unsupported component bind values, missing input fields, input type mismatches, and list/entity component inputs in page sections are validation errors.
- Unsupported trigger sections, events, or section/event combinations are validation errors.
- Trigger events that require missing CRUD actions are validation errors.
- `compose` supports `stack`, `grid`, and `tabs` in this MVP.
- `gap` supports `sm`, `md`, and `lg`.
- `stackAt` supports `sm`, `md`, `lg`, and `none`; it emits deterministic responsive grid breakpoints.
- `tabs` accepts `gap` and rejects `columns`, `stackAt`, and modal/drawer section display.
- `tabs` rejects `group` in this MVP.
- `columns` and `span` support 1..4.
- `table` sections cannot use modal or drawer display in this MVP.
- Grouped sections must be contiguous in generated render order and inline in this MVP.
- Generated JSX/DOM section order follows the effective `view.order`; CSS `order` rules are still emitted as stable metadata and a layout backstop.

## AI Agent Notes

Use `black docs view --json` or `black explain view --json` before editing page section order or composition.

Prefer changing:

```black
view {
  order table, StockSummary, StockCards, detail, form
  compose grid columns 2 gap md stackAt md
  section table span 2
  section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
  section StockCards component StockBadge bind each span 1 title "Stock Cards"
  section detail display drawer side right title "Product Details"
  section form display modal title "Product Form"
  trigger detail on rowSelect
  trigger form on createStart
  trigger form on editStart
}
```

instead of editing generated React or CSS.

After changing view composition, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```

## Diagnostics

```text
INVALID_VIEW_DECLARATION
INVALID_VIEW_ORDER
INVALID_VIEW_COMPOSE
INVALID_VIEW_COMPOSE_COLUMNS
INVALID_VIEW_COMPOSE_OPTION
INVALID_VIEW_SECTION
INVALID_VIEW_SECTION_OPTION
INVALID_VIEW_SECTION_SPAN
INVALID_VIEW_TAB
INVALID_VIEW_GROUP
INVALID_VIEW_GROUP_COLUMNS
INVALID_VIEW_GROUP_OPTION
INVALID_VIEW_GROUP_SPAN
INVALID_VIEW_COMPONENT_SECTION
DUPLICATE_VIEW
DUPLICATE_VIEW_ORDER
DUPLICATE_VIEW_COMPOSE
DUPLICATE_VIEW_COMPOSE_OPTION
DUPLICATE_VIEW_SECTION_OPTION
CONFLICTING_VIEW_COMPONENT_SECTION
DUPLICATE_VIEW_GROUP
DUPLICATE_VIEW_GROUP_OPTION
DUPLICATE_VIEW_GROUP_SECTION
MISSING_VIEW_ORDER
MISSING_VIEW_COMPONENT_BIND
MISSING_VIEW_TABS
MISSING_VIEW_GROUP_SECTION
UNKNOWN_VIEW_COMPONENT
UNKNOWN_VIEW_COMPONENT_INPUT_FIELD
UNSUPPORTED_VIEW_SECTION
UNSUPPORTED_VIEW_SECTION_BIND
UNSUPPORTED_VIEW_COMPONENT_BIND
UNSUPPORTED_VIEW_COMPONENT_DISPLAY
UNSUPPORTED_VIEW_COMPONENT_INPUT
UNSUPPORTED_VIEW_COMPONENT_SIDE
DUPLICATE_VIEW_SECTION
UNSUPPORTED_VIEW_COMPOSE_MODE
UNSUPPORTED_VIEW_COMPOSE_COLUMNS
UNSUPPORTED_VIEW_GAP
UNSUPPORTED_VIEW_STACK_AT
UNSUPPORTED_VIEW_SECTION_SPAN
UNSUPPORTED_VIEW_SECTION_DISPLAY
UNSUPPORTED_VIEW_SECTION_SIDE
UNSUPPORTED_VIEW_GROUP
UNSUPPORTED_VIEW_GROUP_COMPOSE_MODE
UNSUPPORTED_VIEW_GROUP_COMPOSE_COLUMNS
UNSUPPORTED_VIEW_GROUP_GAP
UNSUPPORTED_VIEW_GROUP_SECTION
UNSUPPORTED_VIEW_GROUP_SECTION_DISPLAY
UNSUPPORTED_VIEW_GROUP_SPAN
UNSUPPORTED_VIEW_TAB
UNSUPPORTED_VIEW_TAB_SECTION
DUPLICATE_VIEW_TAB
DUPLICATE_VIEW_TAB_SECTION
MISSING_VIEW_TAB_SECTION
VIEW_COMPONENT_INPUT_TYPE_MISMATCH
UNCLOSED_VIEW
UNEXPECTED_VIEW_TOKEN
```
