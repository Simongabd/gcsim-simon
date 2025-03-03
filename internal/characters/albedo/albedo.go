package albedo

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/character/profile"
)

func init() {
	core.RegisterCharFunc(keys.Albedo, NewChar)
}

type char struct {
	*tmpl.Character
	lastConstruct int
	bloomSnapshot combat.Snapshot
	//tracking skill information
	skillActive     bool
	skillAttackInfo combat.AttackInfo
	skillSnapshot   combat.Snapshot
	//c2 tracking
	c2stacks int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ profile.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 40
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.skillHook()
	return nil
}

func (c *char) Condition(k string) int64 {
	switch k {
	case "skill":
		fallthrough
	case "elevator":
		if c.skillActive {
			return 1
		}
		return 0
	case "c2stacks":
		return int64(c.c2stacks)
	default:
		return 0
	}
}
