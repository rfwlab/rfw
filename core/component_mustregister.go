package core

// MustRegisterComponent registers a component constructor under the provided name
// and panics if the component is already registered.
func MustRegisterComponent(name string, ctor func() Component) {
	if err := RegisterComponent(name, ctor); err != nil {
		panic(err)
	}
}
