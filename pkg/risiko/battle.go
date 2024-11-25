package risiko

import "fmt"

const BATTLE_RULE_MAX_UNITS = 3
const BATTLE_RULE_MIN_ATTACK = 2

type BattleStrategy interface {
	UpdateState(WarState)
	GetDices() (Dices, error)
}

///////////////////////////////////////////////////////////////////////////////
// Max attackers -> Always attack with maximum units, regardless of war state
///////////////////////////////////////////////////////////////////////////////

type maxAttackers struct {
	genDices DicesGenerator
	state    WarState
}

func (m *maxAttackers) UpdateState(state WarState) {
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
	if units < BATTLE_RULE_MIN_ATTACK {
		return 0, fmt.Errorf("cannot attack with 1 unit")
	} else if units > BATTLE_RULE_MAX_UNITS {
		return BATTLE_RULE_MAX_UNITS, nil
	} else {
		return units - 1, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// Max defenders -> Always defend with maximum units
///////////////////////////////////////////////////////////////////////////////

type maxDefenders struct {
	genDices DicesGenerator
	state    WarState
}

func (m *maxDefenders) UpdateState(state WarState) {
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
	} else if availableDefenders >= BATTLE_RULE_MAX_UNITS {
		return BATTLE_RULE_MAX_UNITS, nil
	} else {
		return availableDefenders, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// Battle function

func battle(attacker Dices, defender Dices) (int, int) {
	attackerThrows := attacker.Roll()
	defenderThrows := defender.Roll()

	nCompare := defender.Count()
	if attacker.Count() < defender.Count() {
		nCompare = attacker.Count()
	}

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
