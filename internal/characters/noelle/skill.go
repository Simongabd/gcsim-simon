package noelle

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
)

var skillFrames []int

const skillHitmark = 15

func init() {
	skillFrames = frames.InitAbilSlice(78)
	skillFrames[action.ActionAttack] = 12
	skillFrames[action.ActionSkill] = 14 // uses burst frames
	skillFrames[action.ActionBurst] = 14
	skillFrames[action.ActionDash] = 11
	skillFrames[action.ActionJump] = 11
	skillFrames[action.ActionWalk] = 43
}

func (c *char) Skill(p map[string]int) action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Breastplate",
		AttackTag:          combat.AttackTagElementalArt,
		ICDTag:             combat.ICDTagElementalArt,
		ICDGroup:           combat.ICDGroupDefault,
		StrikeType:         combat.StrikeTypeBlunt,
		Element:            attributes.Geo,
		Durability:         50,
		Mult:               shieldDmg[c.TalentLvlSkill()],
		UseDef:             true,
		CanBeDefenseHalted: true,
	}
	snap := c.Snapshot(&ai)

	//add shield first
	defFactor := snap.BaseDef*(1+snap.Stats[attributes.DEFP]) + snap.Stats[attributes.DEF]
	shieldhp := shieldFlat[c.TalentLvlSkill()] + shieldDef[c.TalentLvlSkill()]*defFactor
	c.Core.Player.Shields.Add(c.newShield(shieldhp, shield.ShieldNoelleSkill, 720))

	//activate shield timer, on expiry explode
	c.shieldTimer = c.Core.F + 720 //12 seconds

	c.a4Counter = 0

	// center on player
	// use char queue for this just to be safe in case of C4
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy),
			0,
			0,
		)
	}, skillHitmark)

	// handle C4
	if c.Base.Cons >= 4 {
		c.Core.Tasks.Add(func() {
			if c.shieldTimer == c.Core.F {
				//deal damage
				c.explodeShield()
			}
		}, 720)
	}

	c.SetCDWithDelay(action.ActionSkill, 24*60, 6)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

// C4:
// When Breastplate's duration expires or it is destroyed by DMG, it will deal 400% ATK of Geo DMG to surrounding opponents.
func (c *char) explodeShield() {
	c.shieldTimer = 0
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Breastplate",
		AttackTag:          combat.AttackTagElementalArt,
		ICDTag:             combat.ICDTagElementalArt,
		ICDGroup:           combat.ICDGroupDefault,
		StrikeType:         combat.StrikeTypeBlunt,
		Element:            attributes.Geo,
		Durability:         50,
		Mult:               4,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.15 * 60,
		CanBeDefenseHalted: true,
	}

	//center on player
	c.Core.QueueAttack(ai, combat.NewCircleHit(c.Core.Combat.Player(), 4, false, combat.TargettableEnemy), 0, 0)
}
