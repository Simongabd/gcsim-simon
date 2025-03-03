package viridescent

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/weapon"
)

func init() {
	core.RegisterWeaponFunc(keys.TheViridescentHunt, NewWeapon)

}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p weapon.WeaponProfile) (weapon.Weapon, error) {
	//Upon hit, Normal and Charged Attacks have a 50% chance to generate a Cyclone, which will continuously
	//attract surrounding opponents, dealing 40% of ATK as DMG to these opponents every 0.5s for 4s. This
	//effect can only occur once every 14s.
	w := &Weapon{}
	r := p.Refine

	const icdKey = "viridescent-hunt-icd"
	cd := 900 - r*60
	mult := 0.3 + float64(r)*0.1

	c.Events.Subscribe(event.OnDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		trg := args[0].(combat.Target)

		//only proc on normal and charge attack
		switch atk.Info.AttackTag {
		case combat.AttackTagNormal:
		case combat.AttackTagExtra:
		default:
			return false
		}

		if char.StatusIsActive(icdKey) {
			return false
		}

		if c.Rand.Float64() > 0.5 {
			return false
		}

		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Viridescent",
			AttackTag:  combat.AttackTagWeaponSkill,
			ICDTag:     combat.ICDTagNone,
			ICDGroup:   combat.ICDGroupDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       mult,
		}

		for i := 0; i <= 240; i += 30 {
			c.QueueAttack(ai, combat.NewCircleHit(trg, 3, false, combat.TargettableEnemy), 0, i+1)
		}

		char.AddStatus(icdKey, cd, true)

		return false
	}, fmt.Sprintf("veridescent-%v", char.Base.Key.String()))

	return w, nil
}
