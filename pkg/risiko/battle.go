// Rules of risk:
package risiko

import (
	"context"
	"fmt"
)

type SimulationResult struct {
	NRuns                  int
	NAttackerWon           int
	TotalAttackerUnitsLeft int
}

type SimulationSweep = map[int]map[int]SimulationResult

type BattleState struct {
	AttackerUnits int
	DefenderUnits int
}

type BattleStrategy = func() EngageStrategy

func NewMaxAttackersStrategy(gen DicesGenerator) BattleStrategy {
	return func() EngageStrategy {
		return &maxAttackers{genDices: gen}
	}
}

func NewMaxDefendersStrategy(gen DicesGenerator) BattleStrategy {
	return func() EngageStrategy {
		return &maxDefenders{genDices: gen}
	}
}

func Battle(state BattleState, attacker BattleStrategy, defender BattleStrategy) (BattleState, error) {
	att := attacker()
	def := defender()
	for state.AttackerUnits >= ENGAGE_RULE_MIN_ATTACK && state.DefenderUnits > 0 {
		att.UpdateState(state)
		def.UpdateState(state)

		attackerThrows, err := att.GetDices()
		if err != nil {
			return BattleState{}, fmt.Errorf("oh no %v", err)
		}

		defenderThrows, err := def.GetDices()
		if err != nil {
			return BattleState{}, fmt.Errorf("oh no %v", err)
		}

		attackerLoss, defenderLoss := engage(attackerThrows, defenderThrows)
		state = BattleState{
			AttackerUnits: state.AttackerUnits - attackerLoss,
			DefenderUnits: state.DefenderUnits - defenderLoss,
		}
	}
	return state, nil
}

func Simulate(ctx context.Context, nRuns int, nUnitsSweep int, attackerStrategy BattleStrategy, defenderStrategy BattleStrategy) (SimulationSweep, error) {
	simsCount := 0
	simResult := SimulationSweep{}
	ch := make(chan []*BattleState)
	chErr := make(chan error)
	defer close(ch)
	defer close(chErr)

	go func() {
		for nDefenders := 1; nDefenders <= nUnitsSweep; nDefenders++ {
			for nAttackers := ENGAGE_RULE_MIN_ATTACK; nAttackers <= nUnitsSweep; nAttackers++ {
				go func(nAtt int, nDef int) {
					for i := 0; i < nRuns; i++ {
						initialState := BattleState{
							AttackerUnits: nAtt,
							DefenderUnits: nDef,
						}
						finalState, err := Battle(initialState, attackerStrategy, defenderStrategy)
						if finalState.AttackerUnits < 0 {
							fmt.Printf("WOWOWOWO %v", finalState)
						}
						if err != nil {
							chErr <- err
						} else {
							ch <- []*BattleState{&initialState, &finalState}
						}
					}
				}(nAttackers, nDefenders)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return simResult, nil
		case err := <-chErr:
			return nil, err
		case simRun := <-ch:
			// Prepare metrics
			initialState := simRun[0]
			finalState := simRun[1]
			attackerWon := 0
			if finalState.AttackerUnits >= ENGAGE_RULE_MIN_ATTACK {
				// If the attacker is left with less units than what's needed to
				// attack it means they ran out of attacks
				attackerWon = 1
			}

			// Make sure object is mapped
			if sas, ok := simResult[initialState.AttackerUnits]; !ok {
				simResult[initialState.AttackerUnits] = map[int]SimulationResult{}
				if _, ok := sas[initialState.DefenderUnits]; !ok {
					simResult[initialState.AttackerUnits][initialState.DefenderUnits] = SimulationResult{
						NRuns:                  0,
						NAttackerWon:           0,
						TotalAttackerUnitsLeft: 0,
					}
				}
			}

			simBatch := simResult[initialState.AttackerUnits][initialState.DefenderUnits]
			simResult[initialState.AttackerUnits][initialState.DefenderUnits] = SimulationResult{
				NRuns:                  simBatch.NRuns + 1,
				NAttackerWon:           simBatch.NAttackerWon + attackerWon,
				TotalAttackerUnitsLeft: simBatch.TotalAttackerUnitsLeft + finalState.AttackerUnits,
			}

			if finalState.AttackerUnits < 0 {
				fmt.Printf("Whats going on %v -> %v", initialState, finalState)
			}

			simsCount++
			if simsCount == nRuns*(nUnitsSweep-(ENGAGE_RULE_MIN_ATTACK-1))*nUnitsSweep {
				return simResult, nil
			}
		}
	}
}
