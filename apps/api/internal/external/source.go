package external

// Source identifies an external integration provider.
//
// Examples:
//   - "reddit"
//   - "future-provider"
//
// Source values are intentionally strings so adding a new provider
// does not require modifying this package's type definition.
type Source string

const (
	SourceReddit Source = "reddit"
)
