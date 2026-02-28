package codebase

// CodeFile represents a parsed Go file
type CodeFile struct {
	Path   string
	Blocks []Block
}

// Block represents a logical block within a Go file
type Block struct {
	Name      string
	Content   string
	StartLine int
	EndLine   int
}

// Parser interface for parsing Go files
type Parser interface {
	Parse(filePath string) (*CodeFile, error)
}
