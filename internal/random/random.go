package random

// make sure we conform to Randomizer
var _ Randomizer = &Random{}

type Randomizer interface{}

type Random struct{}

func NewRandomizer() Randomizer {
	return &Random{}
}
