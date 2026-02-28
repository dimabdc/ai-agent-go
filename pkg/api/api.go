package api

// API interface for all endpoints
type API interface {
	SetupRoutes() error
}
