//go:build js && wasm

package framework

type Component interface {
	Render() string
	Mount()
	Unmount()
	GetName() string
	GetID() string
}
