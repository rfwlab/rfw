//go:build js && wasm

package core

type Component interface {
	Render() string
	Mount()
	Unmount()
	OnMount()
	OnUnmount()
	GetName() string
	GetID() string
}
