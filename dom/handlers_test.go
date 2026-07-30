//go:build js && wasm

package dom

import (
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/js"
)

func eventOptions() js.Dict {
	opts := js.NewDict()
	opts.Set("bubbles", true)
	opts.Set("cancelable", true)
	return opts
}

func mountHandlerRoot(t *testing.T, html string) Element {
	t.Helper()
	root := Doc().CreateElement("div")
	root.SetHTML(html)
	Doc().Body().AppendChild(root)
	t.Cleanup(func() { root.Call("remove") })
	return root
}

func TestComponentHandlersAreScoped(t *testing.T) {
	firstRoot := mountHandlerRoot(t, `<button data-on-click="save">one</button>`)
	secondRoot := mountHandlerRoot(t, `<button data-on-click="save">two</button>`)
	var first, second int

	RegisterComponentHandlerFunc("first", "save", func() { first++ })
	RegisterComponentHandlerFunc("second", "save", func() { second++ })
	DelegateEvents("first", firstRoot.Value)
	DelegateEvents("second", secondRoot.Value)
	t.Cleanup(func() {
		RemoveDelegatedEvents("first", firstRoot.Value)
		RemoveDelegatedEvents("second", secondRoot.Value)
		ReleaseComponentHandlers("first")
		ReleaseComponentHandlers("second")
	})

	firstRoot.Query("button").Call("click")
	secondRoot.Query("button").Call("click")
	if first != 1 || second != 1 {
		t.Fatalf("scoped handler counts = %d, %d", first, second)
	}
}

func TestDelegatedEventModifiers(t *testing.T) {
	root := mountHandlerRoot(t, `<button data-on-click="save" data-on-click-modifiers="prevent,once">save</button>`)
	var calls int
	RegisterComponentHandlerFunc("modifiers", "save", func() { calls++ })
	DelegateEvents("modifiers", root.Value)
	t.Cleanup(func() {
		RemoveDelegatedEvents("modifiers", root.Value)
		ReleaseComponentHandlers("modifiers")
	})

	button := root.Query("button")
	first := js.Get("MouseEvent").New("click", eventOptions().Value)
	if allowed := button.Call("dispatchEvent", first).Bool(); allowed {
		t.Fatal("prevent modifier did not cancel the event")
	}
	button.Call("dispatchEvent", js.Get("MouseEvent").New("click", eventOptions().Value))
	if calls != 1 {
		t.Fatalf("once handler calls = %d", calls)
	}
}

func TestDelegatedKeyAndTimingModifiers(t *testing.T) {
	root := mountHandlerRoot(t, `
		<input data-on-keydown="submit" data-on-keydown-modifiers="enter">
		<button id="debounce" data-on-click="search" data-on-click-modifiers="debounce,10">search</button>
		<button id="throttle" data-on-click="refresh" data-on-click-modifiers="throttle,10">refresh</button>
	`)
	var submit, search, refresh int
	RegisterComponentHandlerFunc("timing", "submit", func() { submit++ })
	RegisterComponentHandlerFunc("timing", "search", func() { search++ })
	RegisterComponentHandlerFunc("timing", "refresh", func() { refresh++ })
	DelegateEvents("timing", root.Value)
	t.Cleanup(func() {
		RemoveDelegatedEvents("timing", root.Value)
		ReleaseComponentHandlers("timing")
	})

	input := root.Query("input")
	escape := eventOptions()
	escape.Set("key", "Escape")
	input.Call("dispatchEvent", js.Get("KeyboardEvent").New("keydown", escape.Value))
	enter := eventOptions()
	enter.Set("key", "Enter")
	input.Call("dispatchEvent", js.Get("KeyboardEvent").New("keydown", enter.Value))
	if submit != 1 {
		t.Fatalf("enter handler calls = %d", submit)
	}

	debounce := root.Query("#debounce")
	debounce.Call("click")
	debounce.Call("click")
	throttle := root.Query("#throttle")
	throttle.Call("click")
	throttle.Call("click")
	time.Sleep(30 * time.Millisecond)
	throttle.Call("click")

	if search != 1 {
		t.Fatalf("debounce handler calls = %d", search)
	}
	if refresh != 2 {
		t.Fatalf("throttle handler calls = %d", refresh)
	}
}

func TestDelegatedFocusHandlerUsesCaptureListener(t *testing.T) {
	root := mountHandlerRoot(t, `<input data-on-focus="focus">`)
	var calls int
	RegisterComponentHandlerFunc("focus", "focus", func() { calls++ })
	DelegateEvents("focus", root.Value)
	t.Cleanup(func() {
		RemoveDelegatedEvents("focus", root.Value)
		ReleaseComponentHandlers("focus")
	})

	options := js.NewDict()
	options.Set("bubbles", false)
	root.Query("input").Call("dispatchEvent", js.Get("FocusEvent").New("focus", options.Value))
	if calls != 1 {
		t.Fatalf("focus handler calls = %d", calls)
	}
}
