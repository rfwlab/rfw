package animation

import "testing"

// TestKeyFrameMap verifies basic operations on KeyFrameMap.
func TestKeyFrameMap(t *testing.T) {
	k := NewKeyFrame()
	if len(k) != 0 {
		t.Fatalf("expected empty keyframe map, got %d entries", len(k))
	}

	k.Add("opacity", 0.5).Add("transform", "scale(2)")
	if v, ok := k["opacity"]; !ok || v != 0.5 {
		t.Fatalf("expected opacity 0.5, got %v", v)
	}

	k.Delete("opacity")
	if _, ok := k["opacity"]; ok {
		t.Fatalf("expected opacity key removed")
	}
}
