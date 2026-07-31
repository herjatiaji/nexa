package rules

import (
	"github.com/heraji/jarvis/cognitive"
)

// CodingFocusRule suppresses non-critical notifications during active IDE coding.
type CodingFocusRule struct{}

func NewCodingFocusRule() *CodingFocusRule {
	return &CodingFocusRule{}
}

func (r *CodingFocusRule) Name() string {
	return "coding_focus_rule"
}

func (r *CodingFocusRule) Evaluate(ctx cognitive.CognitiveContext, memory cognitive.WorkingMemory) *cognitive.CognitiveRuleResult {
	if ctx.Activity == "coding_activity" && ctx.Confidence >= 0.85 {
		return &cognitive.CognitiveRuleResult{
			Action:     "SILENT",
			Confidence: 0.90,
			Reason:     "User is actively coding in IDE; suppressing interruptions",
		}
	}
	return nil
}
