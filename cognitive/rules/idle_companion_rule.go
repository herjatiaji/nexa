package rules

import (
	"time"

	"github.com/heraji/jarvis/cognitive"
)

// IdleCompanionRule suggests a friendly toast or summary check-in when user is idle.
type IdleCompanionRule struct{}

func NewIdleCompanionRule() *IdleCompanionRule {
	return &IdleCompanionRule{}
}

func (r *IdleCompanionRule) Name() string {
	return "idle_companion_rule"
}

func (r *IdleCompanionRule) Evaluate(ctx cognitive.CognitiveContext, memory cognitive.WorkingMemory) *cognitive.CognitiveRuleResult {
	if ctx.Activity == "idle" && ctx.IdleDuration > 10*time.Minute {
		return &cognitive.CognitiveRuleResult{
			Action:     "TOAST",
			Confidence: 0.40,
			Reason:     "User idle for > 10 minutes",
			Payload: map[string]string{
				"message": "NEXA is resting in background. Call me whenever you need help!",
			},
		}
	}
	return nil
}
