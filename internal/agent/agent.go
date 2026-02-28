package agent

// Agent represents an analysis agent
type Agent interface {
	Analyze() error
}
