# BlackLang State Declarations

`state` declares client-side UI state intent for generated pages.

It is for deterministic page state such as selected records, active filters, and generated modal open/close state. It is not a general-purpose browser expression runtime.

## Syntax

```black
state OrdersPageState {
  selectedOrders Order[]
  activeFilter text
  modal createOrder closed
}
```

## Rules

- Declare `state` only at the top level.
- Use `PageNameState` or `PageNamePageState` to bind a state block to a generated page.
- State field types may be primitive field types or existing entity names.
- Use `Entity[]` for list state such as `selectedOrders Order[]`.
- Modal declarations use `modal <name> open|closed`.
- State field names and modal names must be unique inside one state block.
- State declarations do not define computed expressions, arbitrary click handlers, loops, or browser-side BlackLang execution.

## Generated Web Output

For a matching generated page, the web generator emits:

- React `useState` hooks for state fields
- Modal `useState` hooks for modal declarations
- `open<Name>` and `close<Name>` helpers for modals
- Generated create-form visibility wiring when a modal name matches the page source create modal pattern

For example:

```black
state OrdersPageState {
  modal createOrder closed
}
```

can control the generated create form modal on `page Orders`.

## AI Agent Notes

Use `state` when the requested behavior fits generated page UI state. Do not use it as a replacement for JavaScript variables, calculator logic, custom event handlers, or arbitrary frontend functions.

Before changing state declarations, run:

```bash
black docs state --json
black inspect app.black --affected OrdersPageState --json
```

After changing state declarations, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```

For calculator-style local expression state, add the missing compiler feature first or clearly label a normal web prototype as non-BlackLang output.

## Diagnostics

```text
INVALID_STATE_DECLARATION
DUPLICATE_STATE
DUPLICATE_STATE_FIELD
UNSUPPORTED_STATE_FIELD_TYPE
INVALID_STATE_MODAL
DUPLICATE_STATE_MODAL
UNSUPPORTED_STATE_MODAL_DEFAULT
```
