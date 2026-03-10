package chat

import (
	"context"
	"time"

	"github.com/gaebalai/secfarm/pkg/adapter"
	"github.com/gaebalai/secfarm/pkg/tool"
	"google.golang.org/genai"
)

// Plan represents an investigation plan with ordered steps
type Plan struct {
	Objective   string    `json:"objective"` // 조사의 목적
	Steps       []Step    `json:"steps"`     // 실행 단계
	GeneratedAt time.Time `json:"generated_at"`
}

// Step represents a single step in the investigation plan
type Step struct {
	ID          string     `json:"id"`          // 단계 ID (예: "step_1")
	Description string     `json:"description"` // 단계 설명
	Tools       []string   `json:"tools"`       // 예상 도구
	Expected    string     `json:"expected"`    // 기대되는 결과
	Status      StepStatus `json:"status"`
}

// StepStatus represents the execution status of a step
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusCanceled   StepStatus = "canceled"
)

// StepResult contains the result of executing a step
type StepResult struct {
	StepID     string     `json:"step_id"`
	Success    bool       `json:"success"`
	Findings   string     `json:"findings"`   // 발견 사항
	ToolCalls  []ToolCall `json:"tool_calls"` // 실행한 도구 호출
	ExecutedAt time.Time  `json:"executed_at"`
}

// ToolCall represents a single tool invocation
type ToolCall struct {
	Name   string         `json:"name"`
	Args   map[string]any `json:"args"`
	Result string         `json:"result"`
}

// Reflection contains the evaluation of a step's execution
type Reflection struct {
	StepID      string       `json:"step_id"`
	Achieved    bool         `json:"achieved"`     // 목적 달성
	Insights    []string     `json:"insights"`     // 새로운 통찰
	PlanUpdates []PlanUpdate `json:"plan_updates"` // 계획 갱신
	ReflectedAt time.Time    `json:"reflected_at"`
}

// PlanUpdate represents a modification to the plan
type PlanUpdate struct {
	Type   UpdateType `json:"type"`              // add_step, update_step, or cancel_step
	Step   Step       `json:"step,omitempty"`    // 단계 정보 (add_step/update_step의 경우)
	StepID string     `json:"step_id,omitempty"` // 단계 ID (cancel_step의 경우)
	Reason string     `json:"reason,omitempty"`  // 취소 사유 (cancel_step의 경우)
}

// UpdateType represents the type of plan update
type UpdateType string

const (
	UpdateTypeAddStep    UpdateType = "add_step"
	UpdateTypeUpdateStep UpdateType = "update_step"
	UpdateTypeCancelStep UpdateType = "cancel_step"
)

// Conclusion contains the final conclusion
type Conclusion struct {
	Content     string    `json:"content"` // 결론 내용 (마크다운 형식)
	GeneratedAt time.Time `json:"generated_at"`
}

// PlanExecuteResult contains the complete result of plan & execute mode
type PlanExecuteResult struct {
	Plan        *Plan
	Results     []*StepResult
	Reflections []*Reflection
	Conclusion  *Conclusion
}

// planGenerator generates investigation plans
type planGenerator struct {
	gemini   adapter.Gemini
	registry *tool.Registry
}

// conclusionGenerator generates final conclusions
type conclusionGenerator struct {
	gemini adapter.Gemini
}

// newPlanGenerator creates a new plan generator
func newPlanGenerator(gemini adapter.Gemini, registry *tool.Registry) *planGenerator {
	return &planGenerator{
		gemini:   gemini,
		registry: registry,
	}
}

// newConclusionGenerator creates a new conclusion generator
func newConclusionGenerator(gemini adapter.Gemini) *conclusionGenerator {
	return &conclusionGenerator{gemini: gemini}
}

// shouldUsePlanExecuteMode determines if plan & execute mode should be used
func shouldUsePlanExecuteMode(ctx context.Context, gemini adapter.Gemini, message string, history []*genai.Content) bool {
	// Use LLM to judge if the message requires systematic multi-step execution
	systemPrompt := `You are evaluating whether a user's request requires systematic multi-step execution (Plan & Execute mode) or can be handled with direct conversation.

Plan & Execute mode is needed when:
- Multi-step tasks or operations are required
- Complex tasks combining multiple tools or actions
- User requests deep or thorough work ("in detail", "thoroughly", "investigate", "analyze")
- Systematic data collection or processing is necessary

Plan & Execute mode is NOT needed when:
- Simple questions or confirmations
- Questions about already displayed information
- Simple viewing or checking
- Single-step operations
- Follow-up questions in an ongoing conversation

Respond with ONLY "yes" or "no".`

	thinkingBudget := int32(0)
	config := &genai.GenerateContentConfig{
		Temperature: ptrFloat32(0.0),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  &thinkingBudget,
		},
	}

	// Build contents with history + current message
	contents := make([]*genai.Content, 0, len(history)+2)
	contents = append(contents, genai.NewContentFromText(systemPrompt, genai.RoleUser))

	// Add conversation history if exists
	if len(history) > 0 {
		contents = append(contents, history...)
	}

	// Add current message
	contents = append(contents, genai.NewContentFromText(message, genai.RoleUser))

	resp, err := gemini.GenerateContent(ctx, contents, config)
	if err != nil {
		// On error, fall back to direct mode
		return false
	}

	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return false
	}

	answer := resp.Candidates[0].Content.Parts[0].Text
	return answer == "yes" || answer == "Yes" || answer == "YES"
}

func ptrFloat32(f float32) *float32 {
	return &f
}
