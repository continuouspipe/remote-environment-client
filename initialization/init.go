package initialization

type InitState interface {
	Handle() (suggestion string, err error)
	Next() InitState
	Name() string
}
