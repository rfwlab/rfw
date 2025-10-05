package devtools

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
)

type storeInspector interface {
	Snapshot() map[string]any
	Module() string
	Name() string
}

var (
	htmlComponentType    = reflect.TypeOf(core.HTMLComponent{})
	htmlComponentPtrType = reflect.TypeOf((*core.HTMLComponent)(nil))
)

func captureTree(c core.Component) {
	resetTree()
	walk(c, "")
}

func walk(c core.Component, parentID string) {
	if c == nil {
		return
	}
	id := c.GetID()
	kind := componentKind(c)
	name := c.GetName()
	n := addComponent(id, kind, name, parentID)
	populateMetadata(n, c)
	for _, child := range extractDependencies(c) {
		walk(child, id)
	}
}

func componentKind(c core.Component) string {
	v := reflect.ValueOf(c)
	if !v.IsValid() {
		return ""
	}
	t := v.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

func extractDependencies(c core.Component) []core.Component {
	v := reflect.ValueOf(c)
	if !v.IsValid() {
		return nil
	}
	v = reflect.Indirect(v)
	if !v.IsValid() {
		return nil
	}
	if deps := mapOfComponents(v.FieldByName("Dependencies")); len(deps) > 0 {
		return deps
	}
	if hc := unwrapHTMLComponentValue(c); hc.IsValid() {
		if deps := mapOfComponents(reflect.Indirect(hc).FieldByName("Dependencies")); len(deps) > 0 {
			return deps
		}
	}
	return nil
}

func mapOfComponents(field reflect.Value) []core.Component {
	if !field.IsValid() || field.IsNil() {
		return nil
	}
	if !field.CanInterface() {
		return nil
	}
	if deps, ok := field.Interface().(map[string]core.Component); ok {
		list := make([]core.Component, 0, len(deps))
		for _, child := range deps {
			if child != nil {
				list = append(list, child)
			}
		}
		sort.Slice(list, func(i, j int) bool { return list[i].GetName() < list[j].GetName() })
		return list
	}
	return nil
}

func populateMetadata(n *node, c core.Component) {
	if n == nil || c == nil {
		return
	}
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	html := unwrapHTMLComponentValue(c)
	if html.IsValid() {
		assignMaps(n, html)
		assignStore(n, html)
		if host := extractString(html, "HostComponent"); host != "" {
			n.Host = host
		}
	}
	assignMaps(n, v)
	assignStore(n, v)
	if n.Host == "" {
		if host := extractString(v, "HostComponent"); host != "" {
			n.Host = host
		}
	}
	if updates := extractInt(v, "Updates"); updates > 0 {
		n.Updates = updates
	}
	if stats, ok := statsFromComponent(c); ok {
		applyStats(n, c.GetID(), stats)
	}
	if owner := extractString(v, "Owner"); owner != "" {
		n.Owner = owner
	}
	if signals := dom.SnapshotComponentSignals(c.GetID()); len(signals) > 0 {
		sanitized := make(map[string]any, len(signals))
		keys := make([]string, 0, len(signals))
		for k := range signals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sanitized[key] = sanitizeValue(signals[key])
		}
		n.Signals = sanitized
	} else if n.Signals == nil {
		if m := extractMap(v, "Signals"); len(m) > 0 {
			n.Signals = m
		}
	}
	applyStoreBindings(n, c.GetID())
}

type statsProvider interface {
	Stats() core.ComponentStats
}

func statsFromComponent(c core.Component) (core.ComponentStats, bool) {
	if c == nil {
		return core.ComponentStats{}, false
	}
	if provider, ok := any(c).(statsProvider); ok {
		return provider.Stats(), true
	}
	if html := unwrapHTMLComponentValue(c); html.IsValid() {
		return statsFromValue(html)
	}
	return core.ComponentStats{}, false
}

func statsFromValue(v reflect.Value) (core.ComponentStats, bool) {
	if !v.IsValid() {
		return core.ComponentStats{}, false
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return core.ComponentStats{}, false
	}
	if v.Kind() != reflect.Pointer {
		if v.CanAddr() {
			v = v.Addr()
		} else {
			return core.ComponentStats{}, false
		}
	}
	if !v.CanInterface() {
		return core.ComponentStats{}, false
	}
	if provider, ok := v.Interface().(statsProvider); ok {
		return provider.Stats(), true
	}
	return core.ComponentStats{}, false
}

func applyStats(n *node, id string, stats core.ComponentStats) {
	if n == nil {
		return
	}
	if stats.LastRender > 0 {
		n.Time = durationToMillis(stats.LastRender)
	}
	if stats.AverageRender > 0 {
		n.Average = durationToMillis(stats.AverageRender)
	}
	if stats.TotalRender > 0 {
		n.Total = durationToMillis(stats.TotalRender)
	}
	if stats.RenderCount > n.Updates {
		n.Updates = stats.RenderCount
	}
	timeline := combineTimelines(stats.Timeline, snapshotLifecycle(id))
	if len(timeline) > 0 {
		n.Timeline = timeline
	}
}

func applyStoreBindings(n *node, componentID string) {
	if n == nil || componentID == "" {
		return
	}
	if bindings := snapshotStoreBindings(componentID); len(bindings) > 0 {
		n.StoreBindings = bindings
	}
}

func durationToMillis(d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(d) / float64(time.Millisecond)
}

func combineTimelines(render []core.ComponentTimelineEntry, lifecycle []lifecycleEvent) []timelineEntry {
	total := len(render) + len(lifecycle)
	if total == 0 {
		return nil
	}
	merged := make([]timelineEntry, 0, total)
	for _, ev := range render {
		if ev.Kind == "" {
			continue
		}
		at := ev.Timestamp.UnixNano() / int64(time.Millisecond)
		merged = append(merged, timelineEntry{
			Kind:     ev.Kind,
			At:       at,
			Duration: durationToMillis(ev.Duration),
		})
	}
	for _, ev := range lifecycle {
		if ev.Kind == "" {
			continue
		}
		merged = append(merged, timelineEntry{
			Kind: ev.Kind,
			At:   ev.At.UnixNano() / int64(time.Millisecond),
		})
	}
	if len(merged) == 0 {
		return nil
	}
	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].At == merged[j].At {
			if merged[i].Duration == merged[j].Duration {
				return merged[i].Kind < merged[j].Kind
			}
			return merged[i].Duration < merged[j].Duration
		}
		return merged[i].At < merged[j].At
	})
	base := merged[0].At
	for i := range merged {
		merged[i].At -= base
		if merged[i].At < 0 {
			merged[i].At = 0
		}
	}
	return merged
}

func assignMaps(n *node, v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if n.Props == nil {
		if props := extractMap(v, "Props"); len(props) > 0 {
			n.Props = props
		}
	}
	if n.Slots == nil {
		if slots := extractMap(v, "Slots"); len(slots) > 0 {
			n.Slots = slots
		}
	}
}

func assignStore(n *node, v reflect.Value) {
	if !v.IsValid() || n.Store != nil {
		return
	}
	if snap := extractStore(v, "Store"); snap != nil {
		n.Store = snap
	}
}

func extractStore(v reflect.Value, field string) *storeSnapshot {
	f := fieldByName(v, field)
	if !f.IsValid() || !f.CanInterface() {
		return nil
	}
	if isNilable(f.Kind()) && f.IsNil() {
		return nil
	}
	if inspector, ok := f.Interface().(storeInspector); ok {
		state := sanitizeMap(inspector.Snapshot())
		return &storeSnapshot{
			Module: inspector.Module(),
			Name:   inspector.Name(),
			State:  state,
		}
	}
	return nil
}

func extractMap(v reflect.Value, field string) map[string]any {
	f := fieldByName(v, field)
	if !f.IsValid() {
		return nil
	}
	if isNilable(f.Kind()) && f.IsNil() {
		return nil
	}
	if !f.CanInterface() {
		return nil
	}
	switch val := f.Interface().(type) {
	case map[string]any:
		return sanitizeMap(val)
	default:
		if f.Kind() == reflect.Map {
			iter := f.MapRange()
			tmp := make(map[string]any, f.Len())
			for iter.Next() {
				key := keyToString(iter.Key())
				tmp[key] = sanitizeValue(iter.Value().Interface())
			}
			return sanitizeMap(tmp)
		}
	}
	return nil
}

func extractString(v reflect.Value, field string) string {
	f := fieldByName(v, field)
	if !f.IsValid() || !f.CanInterface() {
		return ""
	}
	if s, ok := f.Interface().(string); ok {
		return s
	}
	if f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}

func extractInt(v reflect.Value, field string) int {
	f := fieldByName(v, field)
	if !f.IsValid() || !f.CanInterface() {
		return 0
	}
	if i, ok := f.Interface().(int); ok {
		return i
	}
	if f.Kind() == reflect.Int || f.Kind() == reflect.Int64 || f.Kind() == reflect.Int32 {
		return int(f.Int())
	}
	return 0
}

func fieldByName(v reflect.Value, name string) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return reflect.Value{}
	}
	return v.FieldByName(name)
}

func unwrapHTMLComponentValue(c core.Component) reflect.Value {
	val := reflect.ValueOf(c)
	if !val.IsValid() {
		return reflect.Value{}
	}
	if val.Type() == htmlComponentPtrType {
		return val
	}
	if val.Kind() == reflect.Pointer && !val.IsNil() {
		if val.Elem().Type() == htmlComponentType {
			return val
		}
		val = val.Elem()
	}
	if !val.IsValid() {
		return reflect.Value{}
	}
	if val.Type() == htmlComponentType {
		if val.CanAddr() {
			return val.Addr()
		}
		return reflect.Value{}
	}
	field := val.FieldByName("HTMLComponent")
	if !field.IsValid() {
		return reflect.Value{}
	}
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return reflect.Value{}
		}
		return field
	}
	if field.CanAddr() {
		return field.Addr()
	}
	return reflect.Value{}
}

func sanitizeMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out[k] = sanitizeValue(in[k])
	}
	return out
}

func sanitizeSlice(in []any) []any {
	if len(in) == 0 {
		return nil
	}
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = sanitizeValue(v)
	}
	return out
}

func sanitizeValue(v any) any {
	switch val := v.(type) {
	case nil:
		return nil
	case string, bool:
		return val
	case json.Number:
		return val.String()
	case fmt.Stringer:
		return val.String()
	case json.Marshaler:
		if b, err := val.MarshalJSON(); err == nil {
			var out any
			if err := json.Unmarshal(b, &out); err == nil {
				return out
			}
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	case []byte:
		return string(val)
	case map[string]any:
		return sanitizeMap(val)
	case []any:
		return sanitizeSlice(val)
	}
	if reader, ok := v.(interface{ Read() any }); ok {
		return sanitizeValue(reader.Read())
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return nil
		}
		return sanitizeValue(rv.Elem().Interface())
	case reflect.Slice, reflect.Array:
		arr := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			arr[i] = sanitizeValue(rv.Index(i).Interface())
		}
		return arr
	case reflect.Map:
		result := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := keyToString(iter.Key())
			result[key] = sanitizeValue(iter.Value().Interface())
		}
		return sanitizeMap(result)
	case reflect.Struct:
		return fmt.Sprintf("%v", v)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return rv.Int()
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	}
	return fmt.Sprintf("%v", v)
}

func keyToString(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	return fmt.Sprintf("%v", v)
}

func isNilable(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	default:
		return false
	}
}
