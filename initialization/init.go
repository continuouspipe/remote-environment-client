package initialization

type InitState interface {
	Handle() error
	Next() InitState
	Name() string
}
