package chroma

// ChromaClient interface for vector database operations
type ChromaClient interface {
	StoreVectors() error
	QueryVectors() ([]float32, error)
}
