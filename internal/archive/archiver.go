package archive

// make sure we conform to Archiver
var _ Archiver = &Archive{}

type Archiver interface{}

type Archive struct{}

func NewArchiver() Archiver {
	return &Archive{}
}
