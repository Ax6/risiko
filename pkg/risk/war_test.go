package risk

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

// Returns a generator of dices that always throw the same given number
func createTestSingleSidedDicesGen(number int) func(int) (Dices, error) {
	gen := func(count int) (Dices, error) {
		throws := make([]int, count)
		for i := range throws {
			throws[i] = number
		}
		return getLoadedDices(throws), nil
	}
	return gen
}

// deterministic random dices gen
func testDeterministicDicesGen(count int) (Dices, error) {
	if count < 0 {
		return nil, fmt.Errorf("Dices cannot be a negative number")
	}
	return &fairDices{nDices: count, random: rand.New(rand.NewSource(1))}, nil
}

func TestWarEssentials(t *testing.T) {
	testCases := []struct {
		name     string
		attacker WarStrategy
		defender WarStrategy
		state    WarState
		want     WarState
	}{
		{
			name: "attacker cheater",
			attacker: func() BattleStrategy {
				return &maxAttackers{
					genDices: createTestSingleSidedDicesGen(6),
				}
			},
			defender: func() BattleStrategy {
				return &maxAttackers{
					genDices: createTestSingleSidedDicesGen(3),
				}
			},
			state: WarState{
				AttackerUnits: 10,
				DefenderUnits: 3,
			},
			want: WarState{
				AttackerUnits: 10,
				DefenderUnits: 0,
			},
		},
		{
			name: "defender cheater",
			attacker: func() BattleStrategy {
				return &maxAttackers{
					genDices: createTestSingleSidedDicesGen(1),
				}
			},
			defender: func() BattleStrategy {
				return &maxDefenders{
					genDices: createTestSingleSidedDicesGen(1),
				}
			},
			state: WarState{
				AttackerUnits: 10,
				DefenderUnits: 2,
			},
			want: WarState{
				AttackerUnits: 0,
				DefenderUnits: 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := War(tc.state, tc.attacker, tc.defender)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			if got.AttackerUnits != tc.want.AttackerUnits {
				t.Errorf("Unexpected final state. Want %d but got %d attackers", tc.want.AttackerUnits, got.AttackerUnits)
			}
			if got.DefenderUnits != tc.want.DefenderUnits {
				t.Errorf("Unexpected final state. Want %d but got %d defenders", tc.want.DefenderUnits, got.DefenderUnits)
			}

		})
	}
}

func TestWarResults(t *testing.T) {
	maxAttackers := 15
	maxDefenders := 15
	nWars := 500
	type TestCase = struct {
		name     string
		attacker WarStrategy
		defender WarStrategy
		state    WarState
	}
	testCases := []TestCase{}
	for a := 1; a <= maxAttackers; a++ {
		for d := 1; d <= maxDefenders; d++ {
			for i := 0; i < nWars; i++ {
				testCases = append(testCases, TestCase{
					name:     fmt.Sprintf("war a%d d%d run%d", a, d, i),
					attacker: NewMaxAttackersStrategy(testDeterministicDicesGen),
					defender: NewMaxDefendersStrategy(testDeterministicDicesGen),
					state: WarState{
						AttackerUnits: a,
						DefenderUnits: d,
					},
				})
			}
		}
	}

	var wg sync.WaitGroup
	for _, tc := range testCases {
		wg.Add(1)
		go func(tc TestCase) {
			defer wg.Done()
			t.Run(tc.name, func(t *testing.T) {
				got, err := War(tc.state, tc.attacker, tc.defender)
				if err != nil {
					t.Errorf("Unexpected error")
				}
				if got.AttackerUnits < 0 {
					t.Errorf("Unexpected final state. Want more than 0 but got %d attackers", got.AttackerUnits)
				}
				if got.DefenderUnits < 0 {
					t.Errorf("Unexpected final state. Want more than 0 but got %d defenders", got.DefenderUnits)
				}
				if got.AttackerUnits > tc.state.AttackerUnits {
					t.Errorf("Unexpected final state. Want less than %d but got %d attackers", tc.state.AttackerUnits, got.AttackerUnits)
				}
				if got.DefenderUnits > tc.state.DefenderUnits {
					t.Errorf("Unexpected final state. Want less than %d but got %d defenders", tc.state.DefenderUnits, got.DefenderUnits)
				}
			})
		}(tc)
	}

	wg.Wait()
}

func TestSimulate(t *testing.T) {
	rand.New(rand.NewSource(1))
	ctx := context.Background()
	testCases := []struct {
		name        string
		attacker    WarStrategy
		defender    WarStrategy
		nUnitsSweep int
		nRuns       int
	}{
		{
			name:        "normal cheater",
			attacker:    NewMaxAttackersStrategy(createTestSingleSidedDicesGen(6)),
			defender:    NewMaxDefendersStrategy(createTestSingleSidedDicesGen(6)),
			nUnitsSweep: 20,
			nRuns:       10000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Simulate(ctx, tc.nRuns, tc.nUnitsSweep, tc.attacker, tc.defender)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			for a := 1; a <= tc.nUnitsSweep; a++ {
				for d := 1; d <= tc.nUnitsSweep; d++ {
					if _, ok := result[a]; !ok {
						t.Errorf("unexpected empty object for %d attackers", a)
					} else if got, ok := result[a][d]; !ok {
						t.Errorf("unexpected empty object for %d defenders", d)
					} else if got.NRuns != tc.nRuns {
						t.Errorf("unexpected n of runs for %d attackers and %d defenders. Got %d and wanted %d", a, d, got.NRuns, tc.nRuns)
					}
					if result[a][d].NAttackerWon < 0 {
						t.Errorf("impossible that attackers won is less than 0, but got %d", result[a][d].NAttackerWon)
					}
					if result[a][d].TotalAttackerUnitsLeft < 0 {
						t.Errorf("impossible that total attacker units left are less than 0, but got %d", result[a][d].TotalAttackerUnitsLeft)
					}
				}
			}
		})
	}
}
