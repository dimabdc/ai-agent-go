package rag

// RAGService handles retrieval augmented generation
type RAGService struct {
	// dependencies
}

func NewRAGService() *RAGService {
	return &RAGService{}
}

func (s *RAGService) Retrieve(query string) ([]string, error) {
	return nil, nil
}

func (s *RAGService) Generate(prompt string) (string, error) {
	return "", nil
}
