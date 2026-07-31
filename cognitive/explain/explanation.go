package explain

import (
	"fmt"
	"strings"
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
)

// DecisionExplanation holds human-readable explanations and factor breakdowns of brain decisions.
type DecisionExplanation struct {
	Decision    string   `json:"decision"`    // "SILENT", "TOAST", "SPEAK", "EXECUTE"
	Summary     string   `json:"summary"`     // Natural language explanation in Indonesian/English
	Factors     []string `json:"factors"`     // Contributing factors (e.g. "VS Code active", "Typing speed high")
	SocialScore float64  `json:"social_score"`// Attention / stress score
	Autonomy    string   `json:"autonomy"`    // Current autonomy level
	Confidence  float64  `json:"confidence"`  // Decision confidence
	Timestamp   time.Time`json:"timestamp"`
}

// DecisionExplainer converts raw decision evaluation states into natural language explanations.
type DecisionExplainer struct{}

func NewDecisionExplainer() *DecisionExplainer {
	return &DecisionExplainer{}
}

// Explain analyzes the context, social state, autonomy, rules, and winning decision to generate explanation.
func (e *DecisionExplainer) Explain(
	ctx cognitive.CognitiveContext,
	social cognitive.SocialState,
	autonomyLevel autonomy.AutonomyLevel,
	candidateRuleResults []cognitive.CognitiveRuleResult,
	winningDecision string,
	winningConfidence float64,
) DecisionExplanation {

	var factors []string
	summary := "NEXA sedang memantau latar belakang secara pasif."

	// 1. Analyze Context Factors
	if ctx.FocusedApp != "" {
		factors = append(factors, fmt.Sprintf("Aplikasi aktif: %s", ctx.FocusedApp))
	}
	if ctx.Activity != "" {
		factors = append(factors, fmt.Sprintf("Aktivitas terdeteksi: %s", ctx.Activity))
	}

	// 2. Analyze Social & Attention Factors
	if social.AttentionScore > 0.8 {
		factors = append(factors, fmt.Sprintf("Skor Perhatian Tinggi (%.2f): Pengguna sedang fokus intensif", social.AttentionScore))
	}
	if !social.CanInterrupt {
		factors = append(factors, "Pembatas Sosial: Gangguan dinonaktifkan sementara")
	}

	// 3. Analyze Autonomy Level
	autonomyName := "Suggest"
	switch autonomyLevel {
	case autonomy.AutonomyPassive:
		autonomyName = "Passive"
		factors = append(factors, "Otonomi: Passive (hanya merespon perintah langsung)")
	case autonomy.AutonomySuggest:
		autonomyName = "Suggest"
		factors = append(factors, "Otonomi: Suggest (saran berbasis toast)")
	case autonomy.AutonomyAssist:
		autonomyName = "Assist"
		factors = append(factors, "Otonomi: Assist (tindakan non-destruktif diizinkan)")
	case autonomy.AutonomyAgent:
		autonomyName = "Agent"
		factors = append(factors, "Otonomi: Full Agent")
	}

	// 4. Analyze Evaluated Rules
	var matchedRules []string
	for _, r := range candidateRuleResults {
		if r.Action != "SILENT" && r.Confidence > 0.3 {
			matchedRules = append(matchedRules, fmt.Sprintf("%s (Action: %s, Confidence: %.2f)", r.RuleName, r.Action, r.Confidence))
		}
	}

	if len(matchedRules) > 0 {
		factors = append(factors, fmt.Sprintf("Rule aktif: %s", strings.Join(matchedRules, ", ")))
	}

	// 5. Build Human-Readable Natural Language Summary
	switch winningDecision {
	case "SILENT", "IGNORE", "OBSERVE", "":
		if social.AttentionScore > 0.8 {
			summary = fmt.Sprintf("NEXA memilih diam karena Anda sedang fokus mengetik di %s (Perhatian: %.0f%%).", ctx.FocusedApp, social.AttentionScore*100)
		} else if autonomyLevel == autonomy.AutonomyPassive {
			summary = "NEXA tetap diam karena mode otonomi saat ini diatur ke Passive."
		} else if len(candidateRuleResults) == 0 {
			summary = "NEXA tetap diam karena tidak ada masalah atau pemicu yang memerlukan intervensi."
		} else {
			summary = "NEXA memilih silent observation untuk menjaga fokus pengguna."
		}

	case "TOAST", "SUGGEST":
		summary = "NEXA menampilkan saran toast proaktif karena mendeteksi pemicu bantuan."

	case "SPEAK":
		summary = "NEXA memberikan tanggapan suara secara langsung."

	case "EXECUTE":
		summary = "NEXA menjalankan tindakan otomatis yang disetujui."
	}

	return DecisionExplanation{
		Decision:    winningDecision,
		Summary:     summary,
		Factors:     factors,
		SocialScore: social.AttentionScore,
		Autonomy:    autonomyName,
		Confidence:  winningConfidence,
		Timestamp:   time.Now(),
	}
}
