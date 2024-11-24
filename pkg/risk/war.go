// Rules of risk:
package risk

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

func War(state WarState, attacker BattleStrategy, defender BattleStrategy) (WarState, error) {
	for state.AttackerUnits > 0 && state.DefenderUnits > 0 {
		attacker.UpdateState(state)
		defender.UpdateState(state)

		roundAttackers, err := attacker.GetDices()
		if err != nil {
			return WarState{}, fmt.Errorf("oh no %v", err)
		}

		roundDefenders, err := defender.GetDices()
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

func Simulate(ctx context.Context, nRuns int, nUnitsSweep int, attackerStrategy BattleStrategy, defenderStrategy BattleStrategy) (SimulationSweep, error) {
	simsCount := 0
	simResult := SimulationSweep{}
	ch := make(chan []WarState)
	chErr := make(chan error)
	defer close(ch)
	defer close(chErr)

	go func() {
		for nDefenders := 1; nDefenders <= nUnitsSweep; nDefenders++ {
			for nAttackers := 1; nAttackers <= nUnitsSweep; nAttackers++ {
				go func(nAttackers int, nDefenders int) {
					for i := 0; i < nRuns; i++ {
						initialState := WarState{
							AttackerUnits: nAttackers,
							DefenderUnits: nDefenders,
						}
						finalState, err := War(initialState, attackerStrategy, defenderStrategy)
						if err != nil {
							chErr <- err
						} else {
							ch <- []WarState{initialState, finalState}
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
			if finalState.AttackerUnits > 0 {
				attackerWon = 1
			}

			// Make sure object is mapped
			if sas, ok := simResult[initialState.AttackerUnits]; !ok {
				simResult[initialState.AttackerUnits] = map[int]SimulationResult{}
				if _, ok := sas[initialState.DefenderUnits]; !ok {
					simResult[initialState.AttackerUnits][initialState.DefenderUnits] = SimulationResult{}
				}
			}

			simBatch := simResult[initialState.AttackerUnits][initialState.DefenderUnits]
			simResult[initialState.AttackerUnits][initialState.DefenderUnits] = SimulationResult{
				NRuns:                  simBatch.NRuns + 1,
				NAttackerWon:           simBatch.NAttackerWon + attackerWon,
				TotalAttackerUnitsLeft: simBatch.TotalAttackerUnitsLeft + finalState.AttackerUnits,
			}

			simsCount++
			if simsCount == nRuns*nUnitsSweep*nUnitsSweep {
				return simResult, nil
			}
		}
	}
}
