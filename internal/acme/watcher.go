package acme

type Watcher struct{}

func NewWatcher() *Watcher {
	return &Watcher{}
}

func (w *Watcher) Start() error {

	// TODO:
	// Watch acme.json

	return nil
}