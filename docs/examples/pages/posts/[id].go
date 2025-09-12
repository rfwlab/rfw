//go:build js && wasm

package pages

import (
	core "github.com/rfwlab/rfw/v1/core"
)

// PostsId renders a post detail page using a dynamic segment.
func PostsId() core.Component {
	return core.NewComponent("PostsIdPage", []byte("<div>Post Page</div>"), nil)
}
