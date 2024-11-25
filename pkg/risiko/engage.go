package risiko

import (
	"fmt"
	"slices"
)

const ENGAGE_RULE_MAX_UNITS = 3
const ENGAGE_RULE_MIN_ATTACK = 2

type EngageStrategy interface {
	UpdateState(BattleState)
	GetDices() (Dices, error)
}

///////////////////////////////////////////////////////////////////////////////
// Max attackers -> Always attack with maximum units, regardless of battle state
///////////////////////////////////////////////////////////////////////////////

type maxAttackers struct {
	genDices DicesGenerator
	state    BattleState
}

func (m *maxAttackers) UpdateState(state BattleState) {
	m.state = state
}

func (m *maxAttackers) GetDices() (Dices, error) {
	nUnits, err := getMaxAttackers(m.state.AttackerUnits)
	if err != nil {
		return nil, err
	}
	dices, err := m.genDices(nUnits)
	if err != nil {
		return nil, err
	}
	return dices, nil
}

func getMaxAttackers(units int) (int, error) {
	if units < ENGAGE_RULE_MIN_ATTACK {
		return 0, fmt.Errorf("cannot attack with 1 unit")
	} else if units > ENGAGE_RULE_MAX_UNITS {
		return ENGAGE_RULE_MAX_UNITS, nil
	} else {
		return units - 1, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// Max defenders -> Always defend with maximum units
///////////////////////////////////////////////////////////////////////////////

type maxDefenders struct {
	genDices DicesGenerator
	state    BattleState
}

func (m *maxDefenders) UpdateState(state BattleState) {
	m.state = state
}

func (m *maxDefenders) GetDices() (Dices, error) {
	nUnits, err := getMaxDefenders(m.state.DefenderUnits)
	if err != nil {
		return nil, err
	}
	dices, err := m.genDices(nUnits)
	if err != nil {
		return nil, err
	}
	return dices, nil
}

func getMaxDefenders(availableDefenders int) (int, error) {
	if availableDefenders <= 0 {
		return 0, fmt.Errorf("cannot defend with 0 units")
	} else if availableDefenders >= ENGAGE_RULE_MAX_UNITS {
		return ENGAGE_RULE_MAX_UNITS, nil
	} else {
		return availableDefenders, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// Engage function

// Rolls the dices used by the attacker and the defender and compares results to
// establish units lost per side. Returns attacker loss followed by defender
// loss.
func engage(attacker Dices, defender Dices) (int, int) {
	// Roll
	attackerThrows := attacker.Roll()
	defenderThrows := defender.Roll()

	// Prepare throws for comparison
	slices.Sort(attackerThrows)
	slices.Sort(defenderThrows)
	slices.Reverse(attackerThrows)
	slices.Reverse(defenderThrows)
	nCompare := defender.Count()
	if attacker.Count() < defender.Count() {
		nCompare = attacker.Count()
	}

	// Compare
	attackerLoss := 0
	defenderLoss := 0
	for i := range nCompare {
		attDice := attackerThrows[i]
		defDice := defenderThrows[i]
		if attDice > defDice {
			defenderLoss += 1
		} else {
			attackerLoss += 1
		}
	}
	return attackerLoss, defenderLoss
}
