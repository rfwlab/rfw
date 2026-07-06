package types

import "testing"

func TestTypedSignals(t *testing.T) {
	count := NewInt(1)
	if count.Get() != 1 || count.Read() != 1 {
		t.Fatalf("unexpected initial int signal value")
	}

	var seen int
	sub := count.OnChange(func(v int) { seen = v })
	count.Set(2)
	if seen != 2 {
		t.Fatalf("expected OnChange value 2, got %d", seen)
	}
	sub.Stop()
	count.Set(3)
	if seen != 2 {
		t.Fatalf("expected stopped subscription to remain 2, got %d", seen)
	}
}

func TestSignalChannelAndHostConversion(t *testing.T) {
	count := NewInt(0)
	ch := count.Channel()
	count.SetFromHost(float64(7))
	if count.Get() != 7 {
		t.Fatalf("expected host float64 converted to int 7, got %d", count.Get())
	}
	select {
	case got := <-ch:
		if got != 7 {
			t.Fatalf("expected channel value 7, got %d", got)
		}
	default:
		t.Fatalf("expected channel update")
	}
}

func TestCollectionsPropsAndInject(t *testing.T) {
	items := NewSlice([]string{"a"})
	items.Set(append(items.Get(), "b"))
	if got := items.Get(); len(got) != 2 || got[1] != "b" {
		t.Fatalf("unexpected slice value: %v", got)
	}

	m := NewMap(map[string]int{"a": 1})
	if m.Get()["a"] != 1 {
		t.Fatalf("unexpected map value: %v", m.Get())
	}

	prop := NewProp("old")
	prop.Set("new")
	if prop.Get() != "new" {
		t.Fatalf("unexpected prop value: %s", prop.Get())
	}

	inject := Inject[int]{Value: 42}
	if inject.Value != 42 {
		t.Fatalf("unexpected inject value: %d", inject.Value)
	}
}
