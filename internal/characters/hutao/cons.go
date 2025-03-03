package hutao

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c6ICDKey = "hutao-c6-icd"
)

func (c *char) c6() {
	c.c6buff = make([]float64, attributes.EndStatType)
	c.c6buff[attributes.CR] = 1
	c.Core.Events.Subscribe(event.OnCharacterHurt, func(_ ...interface{}) bool {
		c.checkc6()
		return false
	}, "hutao-c6")
}

func (c *char) checkc6() {
	if c.StatusIsActive(c6ICDKey) {
		return
	}
	//check if hp less than 25%
	if c.HPCurrent/c.MaxHP() > .25 {
		return
	}
	//if dead, revive back to 1 hp
	if c.HPCurrent <= -1 {
		c.HPCurrent = 1
	}

	//increase crit rate to 100%
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("hutao-c6", 600),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return c.c6buff, true
		},
	})

	c.AddStatus(c6ICDKey, 3600, true)
}

//Upon defeating an enemy affected by a Blood Blossom that Hu Tao applied
//herself, all nearby allies in the party (excluding Hu Tao herself) will have
//their CRIT Rate increased by 12% for 15s.
func (c *char) c4() {
	c.c4buff = make([]float64, attributes.EndStatType)
	c.c4buff[attributes.CR] = 0.12
	c.Core.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		//do nothing if not an enemy
		if !ok {
			return false
		}
		if !t.StatusIsActive(bbDebuff) {
			return false
		}
		for i, char := range c.Core.Player.Chars() {
			//does not affect hutao
			if c.Index == i {
				continue
			}
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("hutao-c4", 900),
				AffectedStat: attributes.CR,
				Amount: func() ([]float64, bool) {
					return c.c4buff, true
				},
			})
		}

		return false
	}, "hutao-c4")
}
