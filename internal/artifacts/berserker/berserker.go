﻿package berserker

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/artifact"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.Berserker, NewSet)
}

type Set struct {
	Index int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (artifact.Set, error) {
	s := Set{}

	// 2 Piece: CRIT Rate +12%
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = 0.12
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("berserker-2pc", -1),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	// 4 Piece: When HP is below 70%, CRIT Rate increases by an additional 24%.
	if count >= 4 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = 0.24
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("berserker-4pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				maxhp := char.MaxHP()
				if char.HPCurrent > 0.70*maxhp {
					return nil, false
				}
				return m, true
			},
		})
	}

	return &s, nil
}
