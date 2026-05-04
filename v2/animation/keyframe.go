package animation

// KeyFrameMap simplifies building keyframe properties without
// dealing with raw map structures.
type KeyFrameMap map[string]any

// NewKeyFrame returns an empty KeyFrameMap.
func NewKeyFrame() KeyFrameMap {
	return KeyFrameMap{}
}

// Add inserts or updates a property in the keyframe map and returns the map
// for chaining.
func (k KeyFrameMap) Add(prop string, value any) KeyFrameMap {
	k[prop] = value
	return k
}

// Delete removes a property from the keyframe map if present and returns the map
// for chaining.
func (k KeyFrameMap) Delete(prop string) KeyFrameMap {
	delete(k, prop)
	return k
}
