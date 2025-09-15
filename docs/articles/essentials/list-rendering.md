# List Rendering

Rendering collections is essential for dynamic interfaces. RTML’s `@for` directive iterates over slices, ranges, or maps and keeps the DOM synchronized when data changes.

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

Provide a stable key to reorder lists efficiently:

```rtml
@for:todo in todos
  <li [key {todo.ID}]>{todo.Text}</li>
@endfor
```

Keys let rfw match DOM nodes with data items. Without them, items are recreated whenever the order changes.

---

## Ranges and Maps

The directive also supports ranges and key/value maps:

```rtml
@for:i in 0..count
  <span>{i}</span>
@endfor

@for:key,val in dict
  <p>{key}: {val}</p>
@endfor
```

* **Ranges**: expand from start to end (inclusive).
* **Maps**: iterate key/value pairs in Go maps.

---

## Nesting and Conditions

`@for` can be combined with conditionals and nested components:

```rtml
@for:todo in todos
  @if:todo.Done
    <li class="done">{todo.Text}</li>
  @else
    <li>{todo.Text}</li>
  @endif
@endfor
```

This makes it easy to express dynamic UIs with complex data structures.

---

## Summary

* Use `@for` to iterate slices, ranges, or maps.
* Add `[key {...}]` for stable, efficient updates.
* Combine with `@if` or component includes for flexibility.

List rendering ensures templates remain declarative while rfw keeps the DOM in sync with your data.
