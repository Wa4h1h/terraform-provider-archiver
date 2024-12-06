package random

type Randomizer interface {
	Generate(opt Options) ([]byte, error)
}

type Options struct {
	SpecialString string
	RegexPattern  string
	MinLength     int
	Secret        bool
	Special       bool
	Numbers       bool
	UpperCase     bool
}

type Random struct{}
