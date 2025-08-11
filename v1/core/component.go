//go:build js && wasm

package core

type Component interface {
	Render() string
	Mount()
	Unmount()
	GetName() string
	GetID() string
}
