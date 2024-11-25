package risiko

import (
	"fmt"
	"testing"
)

func TestFairDices_Count(t *testing.T) {
	for i := -10; i <= 10; i++ {
		t.Run(fmt.Sprintf("%d dices", i), func(t *testing.T) {
			dices, err := FairDicesGen(i)
			if i < 0 && err == nil {
				t.Errorf("Expected new dices to fail")
			}
			if err == nil && dices.Count() != i {
				t.Errorf("Expected %d dices but got %d", i, dices.Count())
			}
		})
	}
}

func TestFairDices_Roll(t *testing.T) {
	for i := 0; i <= 10; i++ {
		t.Run(fmt.Sprintf("%d dices", i), func(t *testing.T) {
			dices, err := FairDicesGen(i)
			if err != nil {
				t.Errorf("Expected no error")
			}

			throws := dices.Roll()
			if len(throws) != i {
				t.Errorf("Expected %d throws but got %d", i, len(throws))
			}
		})
	}
}
