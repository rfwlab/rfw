# List Rendering

Rendering collections is essential for dynamic interfaces. RTML's `@for:` directive iterates over slices, maps, and ranges, and keeps the DOM synchronized when data changes.

---

## Iterating Slices

Loop through a slice with `@for:item in items`:

```rtml
<ul>
  @for:todo in todos
    <li>{todo.Text}</li>
  @endfor
</ul>
```

When items are added or removed, rfw patches only the affected `<li>` elements instead of re-rendering the whole list.

---

## Keyed Updates

Provide a stable key with `[key {expr}]` for efficient reordering:

```rtml
@for:todo in todos
  <li [key {todo.ID}]>{todo.Text}</li>
@endfor
```

Keys let rfw match DOM nodes with data items. Without them, items are recreated whenever the order changes.

---

## Iterating Maps

Use `@for:key,val in obj` to iterate key/value pairs in a Go map:

```rtml
@for:name,age in ages
  <p>{name}: {age}</p>
@endfor
```

Map iteration order follows Go's map iteration semantics.

---

## Signals in Loops

When iterating a signal-backed collection, use the `signal:` prefix:

```rtml
@for:item in signal:Items
  <li [key {item.ID}]>{item.Name}</li>
@endfor
```

Changes to the `Items` signal patch only the affected DOM nodes.

---

## Nesting and Conditions

`@for` can be combined with `@if:` and nested components:

```rtml
@for:todo in signal:Todos
  @if:todo.Done
    <li class="done">{todo.Text}</li>
  @else
    <li>{todo.Text}</li>
  @endif
@endfor
```

---

## @endforeach

The `@endforeach` directive is an alias for `@endfor`, supporting the alternative `@foreach:` syntax:

```rtml
@foreach:items as item
  <li>{item}</li>
@endforeach
```

This is equivalent to `@for:item in items ... @endfor`.

---

## Summary

| Syntax                        | Purpose                          |
| ----------------------------- | -------------------------------- |
| `@for:item in items`          | Iterate a slice                  |
| `@for:key,val in obj`        | Iterate a map's key/value pairs  |
| `[key {expr}]`                | Stable key for efficient updates |
| `@for:item in signal:Items`  | Iterate a signal-backed collection |
| `@foreach:items as item`      | Alternative loop syntax          |
| `@endfor` / `@endforeach`     | Close the loop block              |

List rendering ensures templates remain declarative while rfw keeps the DOM in sync with your data.