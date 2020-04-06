package grpc

type App interface {
	AllRecipeTags() ([]string, error)
}
