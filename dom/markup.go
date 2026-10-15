package dom

import (
	"fmt"
	"regexp"
	"strings"
)

// reExpandEvent matches the @on:event:handler (and @event:handler) template
// directives, including dot modifiers, terminated by whitespace, a
// self-closing slash or the end of the tag.
var reExpandEvent = regexp.MustCompile(`@(on:)?(\w+(?:\.\w+)*):(\w+)([\s>/])`)

// ExpandEvents rewrites @on:event:handler directives into the data-on-*
// attributes event delegation resolves. Templates go through it
// automatically; call it on markup built at runtime so dynamic rows can use
// the same syntax as .rtml files:
//
//	rows += `<tr @on:click:openRow data-id="` + id + `">...</tr>`
//	el.SetHTML(dom.ExpandEvents(rows))
func ExpandEvents(markup string) string {
	// one execution per match: ReplaceAllStringFunc would run the pattern
	// again inside the callback just to reach the groups, and a row-heavy
	// document carries one directive per row
	locs := reExpandEvent.FindAllStringSubmatchIndex(markup, -1)
	if locs == nil {
		return markup
	}
	var sb strings.Builder
	sb.Grow(len(markup))
	last := 0
	for _, loc := range locs {
		fullEvent := markup[loc[4]:loc[5]]
		handler := markup[loc[6]:loc[7]]
		suffix := markup[loc[8]:loc[9]]
		event, modifiers := fullEvent, ""
		if dot := strings.IndexByte(fullEvent, '.'); dot >= 0 {
			event = fullEvent[:dot]
			modifiers = strings.ReplaceAll(fullEvent[dot+1:], ".", ",")
		}
		sb.WriteString(markup[last:loc[0]])
		fmt.Fprintf(&sb, "data-on-%s=%q", event, handler)
		if modifiers != "" {
			fmt.Fprintf(&sb, " data-on-%s-modifiers=%q", event, modifiers)
		}
		sb.WriteString(suffix)
		last = loc[1]
	}
	sb.WriteString(markup[last:])
	return sb.String()
}
