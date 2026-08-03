package acme

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(path string) error {

	// TODO:
	// Parse acme.json

	return nil
}