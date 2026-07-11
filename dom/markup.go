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
	return reExpandEvent.ReplaceAllStringFunc(markup, func(match string) string {
		parts := reExpandEvent.FindStringSubmatch(match)
		if len(parts) != 5 {
			return match
		}
		fullEvent := parts[2]
		handler := parts[3]
		suffix := parts[4]
		eventParts := strings.Split(fullEvent, ".")
		event := eventParts[0]
		attr := fmt.Sprintf("data-on-%s=%q", event, handler)
		if len(eventParts) > 1 {
			attr += fmt.Sprintf(" data-on-%s-modifiers=%q", event, strings.Join(eventParts[1:], ","))
		}
		return attr + suffix
	})
}
