package qiqi

import (
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// Implements event hook and incoming healing bonus function for A1
// TODO: Could possibly change this so the AddIncHealBonus occurs at start, then event subscription occurs upon using Qiqi skill?
// TODO: Likely more efficient to not maintain event subscription always, but grouping the two for clarity currently
// When a character under the effects of Adeptus Art: Herald of Frost triggers an Elemental Reaction, their Incoming Healing Bonus is increased by 20% for 8s.
func (c *char) a1() {
	a1Hook := func(args ...interface{}) bool {
		if c.StatusIsActive(skillBuffKey) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)

		// Active char is the only one under the effects of Qiqi skill
		active := c.Core.Player.ActiveChar()
		if atk.Info.ActorIndex != active.Index {
			return false
		}

		active.AddHealBonusMod(character.HealBonusMod{
			Base: modifier.NewBaseWithHitlag("qiqi-a1", 8*60),
			Amount: func() (float64, bool) {
				return .2, false
			},
		})

		return false
	}

	for i := event.Event(event.ReactionEventStartDelim + 1); i < event.ReactionEventEndDelim; i++ {
		c.Core.Events.Subscribe(i, a1Hook, "qiqi-a1")
	}
}

const a4ICDKey = "qiqi-a4-icd"
