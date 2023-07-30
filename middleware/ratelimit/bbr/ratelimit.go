package bbr

// DoneFunc is done function.
type DoneFunc func(DoneInfo)

// DoneInfo is done info.
type DoneInfo struct {
	Err error
}

// Limiter is a rate limiter.
type Limiter interface {
	Allow() (DoneFunc, error)
}
