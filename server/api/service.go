package api

// GospigaService wraps app implementation.
type GospigaService struct {
	app App
}

// NewService returns a new instance of GospigaService.
func NewService(app App) *GospigaService {
	return &GospigaService{app: app}
}
