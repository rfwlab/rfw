package rendertrace

// Cause identifies why a component render was requested.
type Cause struct {
	Kind   string
	Module string
	Store  string
	Key    string
	Signal string
}

// Record is one phase of a render trace. Emit assigns the schema version,
// sequence and timestamp so callers only provide render-specific facts.
type Record struct {
	Event             string
	BatchID           uint64
	RenderID          uint64
	ComponentID       string
	ComponentName     string
	ParentComponentID string
	Depth             int
	Cause             Cause
	Causes            []Cause
	QueueDepth        int
	CoalescedCount    int
	TemplateMS        float64
	DOMMS             float64
	TotalMS           float64
	Outcome           string
	Reason            string
	SupersededBy      uint64
}

// NormalizeCause gives callers that have no more specific provenance a stable
// explicit cause instead of emitting an empty object.
func NormalizeCause(cause Cause) Cause {
	if cause.Kind == "" {
		cause.Kind = "explicit"
	}
	return cause
}

// AppendCause adds one cause unless an identical cause is already present.
func AppendCause(causes []Cause, cause Cause) []Cause {
	cause = NormalizeCause(cause)
	for _, existing := range causes {
		if existing == cause {
			return causes
		}
	}
	return append(causes, cause)
}
