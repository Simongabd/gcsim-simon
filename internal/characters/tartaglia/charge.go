package tartaglia

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

var (
	chargeFrames   []int
	chargeHitmarks = []int{14, 27}
)

func init() {
	chargeFrames = frames.InitAbilSlice(55)
	chargeFrames[action.ActionSkill] = 29
	chargeFrames[action.ActionBurst] = 29
	chargeFrames[action.ActionDash] = 14
	chargeFrames[action.ActionJump] = 15
	chargeFrames[action.ActionSwap] = 52
}

// since E is aoe, so this should be considered aoe too
// hitWeakPoint: tartaglia can proc Prototype Cresent's Passive on Geovishap's weakspots.
// Evidence: https://youtu.be/oOfeu5pW0oE
func (c *char) ChargeAttack(p map[string]int) action.ActionInfo {
	if !c.StatusIsActive(meleeKey) {
		c.Core.Log.NewEvent("charge called when not in melee stance", glog.LogActionEvent, c.Index).
			Write("action", action.ActionCharge)
		return action.ActionInfo{
			Frames:          func(action.Action) int { return 1200 },
			AnimationLength: 1200,
			CanQueueAfter:   1200,
			State:           action.Idle,
		}
	}

	hitWeakPoint, ok := p["hitWeakPoint"]
	if !ok {
		hitWeakPoint = 0
	}

	ai := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         "Charged Attack",
		AttackTag:    combat.AttackTagExtra,
		ICDTag:       combat.ICDTagExtraAttack,
		ICDGroup:     combat.ICDGroupDefault,
		StrikeType:   combat.StrikeTypeSlash,
		Element:      attributes.Hydro,
		Durability:   25,
		HitWeakPoint: hitWeakPoint != 0,
	}

	for i, mult := range eCharge {
		ai.Mult = mult[c.TalentLvlSkill()]
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(c.Core.Combat.Player(), 1, false, combat.TargettableEnemy),
			chargeHitmarks[i],
			chargeHitmarks[i],
			c.meleeApplyRiptide, // call back for applying riptide
			c.rtSlashCallback,   // call back for triggering slash
		)
	}

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionDash], // earliest cancel
		State:           action.ChargeAttackState,
	}
}
