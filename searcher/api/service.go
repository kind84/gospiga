package api

type GospigaService struct {
	app App
}

func NewService(app App) *GospigaService {
	return &GospigaService{app: app}
}
