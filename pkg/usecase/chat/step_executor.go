package chat

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/gaebalai/goerr/v2"
	"github.com/gaebalai/secfarm/pkg/adapter"
	"github.com/gaebalai/secfarm/pkg/tool"
	"google.golang.org/genai"
)

//go:embed prompt/execute.md
var executePromptRaw string

var executePromptTmpl = template.Must(template.New("execute").Parse(executePromptRaw))

// runStepExecution executes a single step in the investigation plan
func runStepExecution(ctx context.Context, gemini adapter.Gemini, registry *tool.Registry, step *Step, plan *Plan, previousResults []*StepResult) (*StepResult, error) {
	// Tool call limit
	const maxIterations = 8

	// Build prompt using template
	toolList := buildToolList(registry)

	var buf bytes.Buffer
	if err := executePromptTmpl.Execute(&buf, map[string]any{
		"Objective":       plan.Objective,
		"StepID":          step.ID,
		"StepDescription": step.Description,
		"StepExpected":    step.Expected,
		"PreviousResults": previousResults,
		"ToolList":        toolList,
		"MaxIterations":   maxIterations,
	}); err != nil {
		return nil, goerr.Wrap(err, "failed to execute execute prompt template")
	}

	// Build config with tools
	thinkingBudget := int32(0)
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(buf.String(), ""),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  &thinkingBudget,
		},
	}

	if registry != nil {
		config.Tools = registry.Specs()
	}

	// Create initial user message
	userMessage := fmt.Sprintf("Execute this step: %s", step.Description)
	contents := []*genai.Content{
		genai.NewContentFromText(userMessage, genai.RoleUser),
	}

	// Tool Call loop
	var findings strings.Builder
	toolCalls := make([]ToolCall, 0)

	for i := 0; i < maxIterations; i++ {
		// Update system instruction with current iteration count
		currentPrompt := fmt.Sprintf("%s\n\n**Current Status**: Tool call iteration %d/%d", buf.String(), i+1, maxIterations)
		config.SystemInstruction = genai.NewContentFromText(currentPrompt, "")
		resp, err := gemini.GenerateContent(ctx, contents, config)
		if err != nil {
			return &StepResult{
				StepID:     step.ID,
				Success:    false,
				Findings:   fmt.Sprintf("Error: %v", err),
				ToolCalls:  toolCalls,
				ExecutedAt: time.Now(),
			}, goerr.Wrap(err, "failed to generate content")
		}

		// Check if response contains function calls
		hasFunctionCall := false
		var functionResponses []*genai.Part

		for _, candidate := range resp.Candidates {
			if candidate.Content == nil {
				continue
			}

			// Add assistant response to history
			contents = append(contents, candidate.Content)

			for _, part := range candidate.Content.Parts {
				// Collect text responses
				if part.Text != "" {
					// Display LLM's thinking process to user (subtle format)
					fmt.Printf("   💭 %s\n", part.Text)
					findings.WriteString(part.Text)
					findings.WriteString("\n")
				}

				if part.FunctionCall != nil {
					hasFunctionCall = true
					// Execute the tool
					funcResp, execErr := executeTool(ctx, registry, *part.FunctionCall)

					// Record tool call
					resultStr := ""
					if execErr != nil {
						resultStr = fmt.Sprintf("Error: %v", execErr)
						// Display error to user
						fmt.Printf("   ❌ 도구 에러 (%s): %v\n", part.FunctionCall.Name, execErr)
					} else if result, ok := funcResp.Response["result"].(string); ok {
						resultStr = result
					}

					toolCalls = append(toolCalls, ToolCall{
						Name:   part.FunctionCall.Name,
						Args:   part.FunctionCall.Args,
						Result: resultStr,
					})

					if execErr != nil {
						// Create error response
						funcResp = &genai.FunctionResponse{
							Name:     part.FunctionCall.Name,
							Response: map[string]any{"error": execErr.Error()},
						}
					}

					// Collect function response (will be added as single Content later)
					functionResponses = append(functionResponses, &genai.Part{FunctionResponse: funcResp})
				}
			}
		}

		// Add all function responses as a single Content
		if len(functionResponses) > 0 {
			funcRespContent := &genai.Content{
				Role:  genai.RoleUser,
				Parts: functionResponses,
			}
			contents = append(contents, funcRespContent)
		}

		// If no function call, we're done
		if !hasFunctionCall {
			break
		}
	}

	return &StepResult{
		StepID:     step.ID,
		Success:    true,
		Findings:   findings.String(),
		ToolCalls:  toolCalls,
		ExecutedAt: time.Now(),
	}, nil
}

// executeTool executes a tool via registry
func executeTool(ctx context.Context, registry *tool.Registry, funcCall genai.FunctionCall) (*genai.FunctionResponse, error) {
	if registry == nil {
		return nil, goerr.New("tool registry not available")
	}

	// Execute the tool via registry
	resp, err := registry.Execute(ctx, funcCall)
	if err != nil {
		return nil, goerr.Wrap(err, "tool execution failed")
	}

	return resp, nil
}

// buildToolList creates a formatted list of available tools
func buildToolList(registry *tool.Registry) string {
	if registry == nil {
		return "사용 가능한 도구가 없습니다"
	}

	enabledTools := registry.EnabledTools()
	if len(enabledTools) == 0 {
		return "사용 가능한 도구가 없습니다"
	}

	var list strings.Builder
	list.WriteString("다음 도구를 사용할 수 있습니다:\n")
	for _, name := range enabledTools {
		list.WriteString(fmt.Sprintf("- %s\n", name))
	}

	return list.String()
}
