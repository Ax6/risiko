package risk

import (
	"fmt"
	"math/rand"
	"slices"
)

type Dices interface {
	Count() int
	Roll() []int
}

type fairDices struct {
	random *rand.Rand
	nDices int
}

type DicesGenerator = func(int) (Dices, error)

func FairDicesGen(count int) (Dices, error) {
	if count < 0 {
		return nil, fmt.Errorf("Dices cannot be a negative number")
	}
	return &fairDices{nDices: count, random: rand.New(rand.NewSource(int64(rand.Int())))}, nil
}

// Returns Count() dices throws sorted in descending order
func (f *fairDices) Roll() []int {
	res := []int{}
	for i := 0; i < f.nDices; i++ {
		res = append(res, f.random.Intn(6)+1)
	}
	slices.Sort(res)
	slices.Reverse(res)
	return res
}

// How many dices?
func (f *fairDices) Count() int {
	return f.nDices
}
