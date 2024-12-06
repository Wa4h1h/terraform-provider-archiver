package random

import "context"

// make sure we conform to Randomizer
var _ Randomizer = &Random{}

func NewRandomizer() Randomizer {
	return &Random{}
}

func (r *Random) Generate(ctx context.Context, opt Options) ([]byte, error) {
	return nil, nil
}
