//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	js "github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
)

//go:embed templates/todo_component.rtml
var todoTpl []byte

type todo struct {
	ID    int
	Title string
	Done  bool
}

var (
	todos  []todo
	nextID int
	filter = "all" // all | active | done

	// store drives the reactive bindings in the template: the active
	// counter and the live-update counter.
	store = state.NewStore("todos", state.WithModule("app"))
)

type TodoComponent struct {
	*core.HTMLComponent
}

func NewTodoComponent() *TodoComponent {
	store.Set("active", 0)
	store.Set("live", 0)

	c := &TodoComponent{
		HTMLComponent: core.NewHTMLComponent("TodoComponent", todoTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("addTodo", addTodo)

	dom.RegisterHandlerEvent("newTodoKeydown", func(evt js.Value) {
		if evt.Truthy() && evt.Get("key").String() == "Enter" {
			addTodo()
		}
	})

	dom.RegisterHandlerElem("toggleTodo", func(el dom.Element, _ dom.Event) {
		id, err := strconv.Atoi(el.Data("id"))
		if err != nil {
			return
		}
		for i := range todos {
			if todos[i].ID == id {
				todos[i].Done = !todos[i].Done
				break
			}
		}
		renderTodos()
	})

	dom.RegisterHandlerElem("deleteTodo", func(el dom.Element, _ dom.Event) {
		id, err := strconv.Atoi(el.Data("id"))
		if err != nil {
			return
		}
		for i := range todos {
			if todos[i].ID == id {
				todos = append(todos[:i], todos[i+1:]...)
				break
			}
		}
		renderTodos()
	})

	dom.RegisterHandlerElem("setFilter", func(el dom.Element, _ dom.Event) {
		f := el.Data("filter")
		if f == "all" || f == "active" || f == "done" {
			filter = f
			renderTodos()
		}
	})

	// Live-update scenario: tick a store-bound counter every 100ms,
	// 30 times (3 seconds), exercising fine-grained store bindings.
	dom.RegisterHandlerFunc("startLive", func() {
		store.Set("live", 0)
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for i := 1; i <= 30; i++ {
				<-ticker.C
				store.Set("live", i)
			}
		}()
	})

	c.SetOnMount(func(*core.HTMLComponent) { renderTodos() })

	c.Init(nil)
	return c
}

func addTodo() {
	input := dom.Query("#new-todo")
	title := input.Val()
	if title == "" {
		return
	}
	nextID++
	todos = append(todos, todo{ID: nextID, Title: title})
	input.SetValue("")
	renderTodos()
}

func renderTodos() {
	rows := ""
	active := 0
	for _, t := range todos {
		if !t.Done {
			active++
		}
		if (filter == "active" && t.Done) || (filter == "done" && !t.Done) {
			continue
		}
		checked := ""
		class := ""
		if t.Done {
			checked = " checked"
			class = ` class="done"`
		}
		rows += fmt.Sprintf(
			`<li%s data-id="%d"><input type="checkbox" @on:change:toggleTodo data-id="%d"%s><span>%s</span><button class="destroy" @on:click:deleteTodo data-id="%d">x</button></li>`,
			class, t.ID, t.ID, checked, html.EscapeString(t.Title), t.ID)
	}
	dom.Query("#todo-list").SetHTML(dom.ExpandEvents(rows))
	store.Set("active", active)
}
