package combat

import (
	"fmt"
	"log"

	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

func willAttackLand(a *AttackEvent, t Target, index int) (bool, string) {
	//shape shouldn't be nil; panic here
	if a.Pattern.Shape == nil {
		panic("unexpected nil shape")
	}
	//shape can't be nil now, check if type matches
	if !a.Pattern.Targets[t.Type()] {
		return false, "wrong type"
	}
	//skip if self harm is false and dmg src == i
	if !a.Pattern.SelfHarm && a.Info.DamageSrc == index {
		return false, "no self harm"
	}

	//check if shape matches
	switch v := a.Pattern.Shape.(type) {
	case *Circle:
		return t.Shape().IntersectCircle(*v), "intersect circle"
	case *Rectangle:
		return t.Shape().IntersectRectangle(*v), "intersect rectangle"
	case *SingleTarget:
		//only true if
		return v.Target == index, "target"
	default:
		return false, "unknown shape"
	}
}

func (h *Handler) ApplyAttack(a *AttackEvent) float64 {
	// died := false
	var total float64
	var landed bool
	for i, t := range h.targets {
		//skip nil targets; we don't want to reindex...
		if t == nil {
			continue
		}

		if !t.IsAlive() {
			//this should always evaluate to true if DamageMode is not set since we check for DamageMode before killing
			continue
		}

		willHit, reason := willAttackLand(a, t, i)
		if !willHit {
			// Move target logs into the "Sim" event log to avoid cluttering main display for stuff like Guoba
			// And obvious things like "Fischl A4 is single target so it didn't hit targets 2-4"
			// TODO: Maybe want to add a separate set of log events for this?
			//don't log this for target 0
			if i > 0 {
				h.Log.NewEventBuildMsg(glog.LogDebugEvent, a.Info.ActorIndex, "skipped ", a.Info.Abil, " ", reason).
					Write("attack_tag", a.Info.AttackTag).
					Write("applied_ele", a.Info.Element).
					Write("dur", a.Info.Durability).
					Write("target", i).
					Write("shape", a.Pattern.Shape.String())
			}
			continue
		}

		//make a copy first
		cpy := *a

		//at this point attack will land
		h.Events.Emit(event.OnAttackWillLand, t, &cpy)

		//check to make sure it's not cancelled for w/e reason
		if a.Cancelled {
			continue
		}
		landed = true

		var amp string
		var dmg float64
		var crit bool

		evt := h.Log.NewEvent(cpy.Info.Abil, glog.LogDamageEvent, cpy.Info.ActorIndex).
			Write("target", i).
			Write("attack-tag", cpy.Info.AttackTag).
			Write("ele", cpy.Info.Element.String()).
			Write("damage", &dmg).
			Write("crit", &crit).
			Write("amp", &amp).
			Write("abil", cpy.Info.Abil).
			Write("source_frame", cpy.SourceFrame)
		evt.WriteBuildMsg(cpy.Snapshot.Logs...)

		if !cpy.Info.SourceIsSim {
			if cpy.Info.ActorIndex < 0 {
				log.Println(cpy)
			}
			preDmgModDebug := h.Team.CombatByIndex(cpy.Info.ActorIndex).ApplyAttackMods(&cpy, t)
			evt.Write("pre_damage_mods", preDmgModDebug)
		}

		dmg, crit = t.Attack(&cpy, evt)
		total += dmg

		h.Events.Emit(event.OnDamage, t, &cpy, dmg, crit)

		//callbacks
		cb := AttackCB{
			Target:      t,
			AttackEvent: &cpy,
			Damage:      dmg,
			IsCrit:      crit,
		}
		for _, f := range cpy.Callbacks {
			f(cb)
		}

		//check if target is dead; skip this for i = 0 since we don't want to
		//delete the player by accident
		if h.DamageMode && t.HP() <= 0 {
			t.Kill()
			h.Events.Emit(event.OnTargetDied, t, cpy)
			// h.targets[i] = nil
		}

		// this works because string in golang is a slice underneath, so the &amp points to the slice info
		// that's why when the underlying string in amp changes (has to be reallocated) the pointer doesn't
		// change since it's just pointing to the slice "header"
		if cpy.Info.Amped {
			amp = string(cpy.Info.AmpType)
		}

	}
	//add hitlag to actor but ignore if this is deployable
	if h.EnableHitlag && landed && !a.Info.IsDeployable {
		dur := a.Info.HitlagHaltFrames
		if h.DefHalt && a.Info.CanBeDefenseHalted {
			dur += 3.6 //0.06
		}
		if dur > 0 {
			h.Team.ApplyHitlag(a.Info.ActorIndex, a.Info.HitlagFactor, dur)
			if h.Debug {
				h.Log.NewEvent(fmt.Sprintf("%v applying hitlag: %.3f", a.Info.Abil, dur), glog.LogHitlagEvent, a.Info.ActorIndex).
					Write("duration", dur).
					Write("factor", a.Info.HitlagFactor)
			}
		}
	}
	h.TotalDamage += total
	return total
}
