package worker

// Worker could scheduled by mio or customized scheduler
type Worker interface {
	Run() error
	Stop() error
}
