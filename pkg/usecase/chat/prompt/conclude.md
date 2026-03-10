# Conclusion

You have completed all steps of the task. Now synthesize the findings into a comprehensive conclusion.

## Objective

{{.Objective}}

## Steps Executed

{{range $index, $step := .Steps}}
### Step {{add $index 1}}: {{$step.ID}}
**Description**: {{$step.Description}}
**Status**: {{$step.Status}}
{{end}}

## Step Results

{{range .Results}}
### {{.StepID}}
**Success**: {{.Success}}
**Findings**:
{{.Findings}}

{{end}}

## Reflections

{{range .Reflections}}
### {{.StepID}}
**Achieved**: {{.Achieved}}
**Insights**:
{{range .Insights}}
- {{.}}
{{end}}
{{end}}

## Your Task

Generate a comprehensive conclusion that synthesizes all findings in markdown format.

**Guidelines**:
- Write in Korean
- Structure the content appropriately based on the task objective
- Be specific and evidence-based
- Clearly separate facts from interpretation
- Acknowledge limitations and gaps when relevant

**Typical Structure** (adapt as needed):
- Summary: Concise overview
- Key findings: Important discoveries with evidence
- Assessment: Interpretation of findings
- Recommendations: Concrete next steps (if applicable)
- Uncertainties: What remains unclear (if any)

**Flexibility**: Adapt the structure to match the task. For simple tasks, a brief summary may suffice. For complex investigations, provide detailed analysis.

## Response Format

Respond directly with markdown-formatted text in Korean. Do not wrap it in JSON.

**Example**:
```
# 결론

## 요약
이 알림은...

## 주요 발견
- 발견 1
- 발견 2

## 평가
...
```

**IMPORTANT**:
- Write in Korean
- Use markdown formatting
- Structure the content appropriately for the task
- Include specific evidence and details
- Make the conclusion actionable and clear
