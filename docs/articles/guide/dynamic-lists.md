# Dynamic lists and events

Every real app has the same page somewhere: fetch a slice from an API,
render it as rows, react to clicks on each row. This guide shows the
idiomatic rfw way, which works both for template markup and for markup you
build at runtime.

## The key idea: event delegation

rfw components do not attach a listener per element. The component root
listens for bubbling events and resolves `data-on-<event>` attributes to
handlers registered from Go. Two consequences:

- markup injected at runtime with `SetHTML` is automatically live: no
  re-binding, no listener bookkeeping;
- one handler serves any number of rows, and receives the element that
  declared it, so per-row data travels in plain `data-*` attributes.

## Rendering a list

Fetch, build rows, inject. Runtime markup can use the same `@on:` syntax as
`.rtml` templates by passing it through `dom.ExpandEvents`:

```go
//go:embed templates/users.rtml
var usersTpl []byte

func NewUsersPage() *core.HTMLComponent {
	c := core.NewHTMLComponent("UsersPage", usersTpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	dom.RegisterHandlerElem("openUser", func(el dom.Element, _ dom.Event) {
		router.Navigate("/users/" + el.Data("id"))
	})

	c.SetOnMount(func(*core.HTMLComponent) { loadUsers() })
	return c
}

func loadUsers() {
	http.Request("/api/users", http.RequestOptions{}, func(status int, body string) {
		var users []User
		json.Unmarshal([]byte(body), &users)

		rows := ""
		for _, u := range users {
			rows += fmt.Sprintf(
				`<tr @on:click:openUser data-id="%s"><td>%s</td><td>%s</td></tr>`,
				u.ID, html.EscapeString(u.Name), html.EscapeString(u.Email))
		}
		dom.Query("#users-rows").SetHTML(dom.ExpandEvents(rows))
	})
}
```

The template only needs the container:

```html
<root>
  <table>
    <thead><tr><th>Name</th><th>Email</th></tr></thead>
    <tbody id="users-rows"></tbody>
  </table>
</root>
```

That is the whole pattern. No listener is ever attached to a row; replacing
the tbody's HTML never breaks anything.

## Handler flavors

Pick the registration matching what the handler needs:

```go
dom.RegisterHandlerFunc("save", func())                        // no arguments
dom.RegisterHandlerEvent("keyed", func(evt js.Value))          // the raw event
dom.RegisterHandlerElem("open", func(el Element, evt Event))   // the element that
                                                               // declared data-on-*,
                                                               // plus the event
```

`RegisterHandlerElem` is the one you want for lists: `el` is the row (or
button) carrying the attribute, even when the actual click landed on an icon
inside it, and `el.Data("id")` reads `data-id`.

## Reading inputs

```go
name := dom.Query("#form-name").Val()
dom.Query("#form-name").SetValue("")
agreed := dom.Query("#form-tos").Checked()
```

## Binary responses

`http.Request` decodes the response as text, which corrupts binary payloads.
Use `RequestBytes` for anything that is not text:

```go
http.RequestBytes("/api/avatar.png", http.RequestOptions{}, func(status int, body []byte) {
	// body is the exact bytes
})
```

## Store-driven lists

When the list lives in a store, `@for` in the template re-renders it
reactively without any of the above; this guide covers the imperative case,
where data arrives from an API call you control.
