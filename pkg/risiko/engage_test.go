package risiko

import (
	"fmt"
	"testing"
)

type loadedDices struct {
	throws []int
}

func getLoadedDices(want []int) Dices {
	return &loadedDices{throws: want}
}

func (l *loadedDices) Count() int {
	return len(l.throws)
}

func (l *loadedDices) Roll() []int {
	return l.throws
}

func TestMaxAttackersStrategy(t *testing.T) {
	strategy := &maxAttackers{genDices: FairDicesGen}
	testCases := []struct {
		state     BattleState
		wantDices int
		wantErr   bool
	}{
		{
			state:     BattleState{AttackerUnits: 1},
			wantDices: 0,
			wantErr:   true,
		},
		{
			state:     BattleState{AttackerUnits: 2},
			wantDices: 1,
		},
		{
			state:     BattleState{AttackerUnits: 3},
			wantDices: 2,
		},
		{
			state:     BattleState{AttackerUnits: 4},
			wantDices: 3,
		},
		{
			state:     BattleState{AttackerUnits: 1000},
			wantDices: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d attackers", tc.state.AttackerUnits), func(t *testing.T) {
			strategy.UpdateState(tc.state)
			dices, err := strategy.GetDices()
			if err != nil && tc.wantErr {
				return
			} else if err != nil {
				t.Errorf("Expected no error")
			}
			if dices.Count() != tc.wantDices {
				t.Errorf("Expected %d dices but got %d", tc.wantDices, dices.Count())
			}
		})
	}
}

func TestMaxDefendersStrategy(t *testing.T) {
	strategy := &maxDefenders{genDices: FairDicesGen}
	testCases := []struct {
		state     BattleState
		wantDices int
	}{
		{
			state: BattleState{
				DefenderUnits: 1,
			},
			wantDices: 1,
		},
		{
			state: BattleState{
				DefenderUnits: 2,
			},
			wantDices: 2,
		},
		{
			state: BattleState{
				DefenderUnits: 3,
			},
			wantDices: 3,
		},
		{
			state: BattleState{
				DefenderUnits: 4,
			},
			wantDices: 3,
		},
		{
			state: BattleState{
				DefenderUnits: 1000,
			},
			wantDices: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d defenders", tc.state.DefenderUnits), func(t *testing.T) {
			strategy.UpdateState(tc.state)
			dices, err := strategy.GetDices()
			if err != nil {
				t.Errorf("Expected no error")
			}
			if dices.Count() != tc.wantDices {
				t.Errorf("Expected %d dices but got %d", tc.wantDices, dices.Count())
			}
		})
	}
}

func TestEngage(t *testing.T) {
	testCases := []struct {
		name      string
		attackers Dices
		defenders Dices
		want      []int
	}{
		{
			name:      "same throws",
			attackers: getLoadedDices([]int{6, 6, 6}),
			defenders: getLoadedDices([]int{6, 6, 6}),
			want:      []int{3, 0},
		},
		{
			name:      "same throws different dices",
			attackers: getLoadedDices([]int{6, 6, 6}),
			defenders: getLoadedDices([]int{6}),
			want:      []int{1, 0},
		},
		{
			name:      "6 4 1 vs 5 5 1",
			attackers: getLoadedDices([]int{6, 4, 1}),
			defenders: getLoadedDices([]int{5, 5, 1}),
			want:      []int{2, 1},
		},
		{
			name:      "4",
			attackers: getLoadedDices([]int{4}),
			defenders: getLoadedDices([]int{4}),
			want:      []int{1, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attackerLoss, defenderLoss := engage(tc.attackers, tc.defenders)
			if tc.want[0] != attackerLoss {
				t.Errorf("Expected attacker loss %d but got %d", attackerLoss, tc.want[0])
			}
			if tc.want[1] != defenderLoss {
				t.Errorf("Expected defender loss %d but got %d", defenderLoss, tc.want[1])
			}
		})
	}
}
