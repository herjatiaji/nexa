package brain

import (
	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
)

// Arbiter evaluates multiple candidate Decisions and selects the winning decision.
type Arbiter struct{}

func NewArbiter() *Arbiter {
	return &Arbiter{}
}

// Select picks the decision with the highest final weighted score.
// FinalScore = Confidence × SocialPermission × AutonomyPermission.
func (a *Arbiter) Select(
	decisions []Decision,
	social cognitive.SocialState,
	autonomyCtrl *autonomy.AutonomyController,
) *Decision {
	if len(decisions) == 0 {
		return nil
	}

	var winner *Decision
	maxScore := 0.0

	for i := range decisions {
		d := &decisions[i]
		if d.Action == "SILENT" || d.Action == "OBSERVE" {
			continue
		}

		socialPerm := social.PermissionFor(d.Action)
		autonomyPerm := autonomyCtrl.PermissionFor(d.Action)

		finalScore := d.Confidence * socialPerm * autonomyPerm

		if finalScore > maxScore {
			maxScore = finalScore
			winner = d
		}
	}

	// Threshold guard: do not act if overall confidence score is below 0.35
	if maxScore < 0.35 {
		return nil
	}

	return winner
}
