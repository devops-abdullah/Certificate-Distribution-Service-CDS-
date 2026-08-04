package certs

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Export() error {

	// TODO:
	// Export fullchain.pem
	// Export privkey.pem

	return nil
}
