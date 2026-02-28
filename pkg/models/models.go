package models

// CodeBlock represents a block of code for vectorization
type CodeBlock struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// QueryRequest represents an analysis query
type QueryRequest struct {
	Query   string `json:"query"`
	Context string `json:"context,omitempty"`
}

// QueryResponse represents the response to a query
type QueryResponse struct {
	Result string      `json:"result"`
	Blocks []CodeBlock `json:"blocks,omitempty"`
}
