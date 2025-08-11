package state

type ReactiveVar struct {
	value     string
	listeners []func(string)
}

func NewReactiveVar(initial string) *ReactiveVar {
	return &ReactiveVar{
		value: initial,
	}
}

func (rv *ReactiveVar) Set(newValue string) {
	rv.value = newValue
	for _, listener := range rv.listeners {
		listener(newValue)
	}
}

func (rv *ReactiveVar) Get() string {
	return rv.value
}

func (rv *ReactiveVar) OnChange(listener func(string)) {
	rv.listeners = append(rv.listeners, listener)
}
