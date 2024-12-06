package random

// make sure we conform to Randomizer
var _ Randomizer = &Random{}

func NewRandomizer() Randomizer {
	return &Random{}
}

func (r *Random) Generate(opt Options) ([]byte, error) {
	return nil, nil
}
