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

type WarState struct {
	AttackerUnits int
	DefenderUnits int
}

type WarStrategy = func() BattleStrategy

func NewMaxAttackersStrategy(gen DicesGenerator) WarStrategy {
	return func() BattleStrategy {
		return &maxAttackers{genDices: gen}
	}
}

func NewMaxDefendersStrategy(gen DicesGenerator) WarStrategy {
	return func() BattleStrategy {
		return &maxDefenders{genDices: gen}
	}
}

func War(state WarState, attacker WarStrategy, defender WarStrategy) (WarState, error) {
	att := attacker()
	def := defender()
	for state.AttackerUnits >= BATTLE_RULE_MIN_ATTACK && state.DefenderUnits > 0 {
		att.UpdateState(state)
		def.UpdateState(state)

		roundAttackers, err := att.GetDices()
		if err != nil {
			return WarState{}, fmt.Errorf("oh no %v", err)
		}

		roundDefenders, err := def.GetDices()
		if err != nil {
			return WarState{}, fmt.Errorf("oh no %v", err)
		}

		attackerLoss, defenderLoss := battle(roundAttackers, roundDefenders)
		state = WarState{
			AttackerUnits: state.AttackerUnits - attackerLoss,
			DefenderUnits: state.DefenderUnits - defenderLoss,
		}
	}
	return state, nil
}

func Simulate(ctx context.Context, nRuns int, nUnitsSweep int, attackerStrategy WarStrategy, defenderStrategy WarStrategy) (SimulationSweep, error) {
	simsCount := 0
	simResult := SimulationSweep{}
	ch := make(chan []*WarState)
	chErr := make(chan error)
	defer close(ch)
	defer close(chErr)

	go func() {
		for nDefenders := 1; nDefenders <= nUnitsSweep; nDefenders++ {
			for nAttackers := 1; nAttackers <= nUnitsSweep; nAttackers++ {
				go func(nAtt int, nDef int) {
					for i := 0; i < nRuns; i++ {
						initialState := WarState{
							AttackerUnits: nAtt,
							DefenderUnits: nDef,
						}
						finalState, err := War(initialState, attackerStrategy, defenderStrategy)
						if finalState.AttackerUnits < 0 {
							fmt.Printf("WOWOWOWO %v", finalState)
						}
						if err != nil {
							chErr <- err
						} else {
							ch <- []*WarState{&initialState, &finalState}
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
			if finalState.AttackerUnits >= BATTLE_RULE_MIN_ATTACK {
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
			if simsCount == nRuns*nUnitsSweep*nUnitsSweep {
				return simResult, nil
			}
		}
	}
}
