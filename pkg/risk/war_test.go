package risk

import (
	"context"
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

func TestWar(t *testing.T) {
	testCases := []struct {
		name     string
		attacker BattleStrategy
		defender BattleStrategy
		state    WarState
		want     WarState
	}{
		{
			name: "attacker cheater",
			attacker: &maxAttackers{
				genDices: createTestSingleSidedDicesGen(6),
			},
			defender: &maxAttackers{
				genDices: createTestSingleSidedDicesGen(3),
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
			attacker: &maxAttackers{
				genDices: createTestSingleSidedDicesGen(1),
			},
			defender: &maxDefenders{
				genDices: createTestSingleSidedDicesGen(1),
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

func TestSimulate(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name        string
		attacker    BattleStrategy
		defender    BattleStrategy
		nUnitsSweep int
		nRuns       int
	}{
		{
			name:        "normal cheater",
			attacker:    NewMaxAttackersStrategy(FairDicesGen),
			defender:    NewMaxDefendersStrategy(FairDicesGen),
			nUnitsSweep: 3,
			nRuns:       5,
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
				}
			}
		})
	}
}
