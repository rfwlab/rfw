# List Rendering

Rendering arrays or maps is common in dynamic interfaces. RTML's `@for` directive iterates over collections and keeps DOM elements in sync when the underlying data changes.

## Iterating Arrays

Use `@for:item in items` to loop through a slice:

```rtml
<ul>
@for:todo in todos
  <li>{todo.Text}</li>
@endfor
</ul>
```

When items are added or removed, RFW patches only the affected `<li>` elements.

## Keyed Updates

Provide a stable key with the `[key {expr}]` constructor to reorder lists efficiently:

```rtml
@for:todo in todos
  <li [key {todo.ID}]>{todo.Text}</li>
@endfor
```

Without keys, elements are recreated when their order changes.

## Ranges and Maps

`@for` also supports numeric ranges and key/value pairs:

```rtml
@for:i in 0..count
  <span>{i}</span>
@endfor

@for:key,val in dict
  <p>{key}: {val}</p>
@endfor
```

List rendering works seamlessly with conditionals and nested components to display complex data structures.
