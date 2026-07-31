package cognitive

// CognitiveRule defines a policy rule that evaluates context & working memory to propose a Decision.
type CognitiveRule interface {
	Name() string
	Evaluate(ctx CognitiveContext, memory WorkingMemory) *CognitiveRuleResult
}

// CognitiveRuleResult holds decision proposed by a CognitiveRule.
type CognitiveRuleResult struct {
	Action     string      `json:"action"`     // "SUGGEST", "TOAST", "SPEAK", "EXECUTE", "SILENT"
	Confidence float64     `json:"confidence"` // 0.0 – 1.0
	Reason     string      `json:"reason"`
	RuleName   string      `json:"rule_name"`
	Payload    interface{} `json:"payload"`
}

// PolicyEngine evaluates registered CognitiveRules during the Think phase.
type PolicyEngine struct {
	rules []CognitiveRule
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		rules: make([]CognitiveRule, 0),
	}
}

// RegisterRule adds a rule to the policy engine.
func (pe *PolicyEngine) RegisterRule(rule CognitiveRule) {
	pe.rules = append(pe.rules, rule)
}

// EvaluateAll runs all rules and returns non-nil proposed decisions.
func (pe *PolicyEngine) EvaluateAll(ctx CognitiveContext, memory WorkingMemory) []CognitiveRuleResult {
	var results []CognitiveRuleResult
	for _, r := range pe.rules {
		if res := r.Evaluate(ctx, memory); res != nil {
			res.RuleName = r.Name()
			results = append(results, *res)
		}
	}
	return results
}
