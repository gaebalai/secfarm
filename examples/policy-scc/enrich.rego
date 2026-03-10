package enrich

# BigQuery에서 로그를 검색하여 영향 분석 수행
# JSON 형식으로 {"result": "false_positive"} 또는 {"result": "confirmed"}를 반환하도록 지시

prompt contains {
	"id": "bigquery_impact_analysis",
	"content": sprintf(`You are a security analyst investigating whether this alert is a false positive or real threat.

Alert Title: %s
Category: %s
Resource: %s

CRITICAL INSTRUCTION - Evidence Priority:
Your verdict MUST be based PRIMARILY on BigQuery log search results, NOT on the alert description itself.

Analysis Steps:
1. Use BigQuery to search for actual log evidence of the reported activity
2. Search for logs related to the resource and timeframe mentioned in the alert
3. If BigQuery returns NO matching logs or 0 rows:
   - This strongly suggests FALSE POSITIVE (the activity didn't actually occur in logs)
   - The alert description may be misleading, test data, or simulation
4. If BigQuery returns matching logs:
   - Check if it's authorized testing (look for test/dev patterns)
   - Verify if the activity matches known legitimate operations
5. Base your final verdict on actual log evidence, not alert claims

Decision Criteria:
- NO BigQuery evidence found (0 rows) → "false_positive"
- BigQuery shows authorized test activity → "false_positive"
- BigQuery confirms suspicious activity with actual data → "confirmed"

IMPORTANT: If you searched BigQuery and got 0 rows, the verdict should be "false_positive" because there's no evidence in logs.

Return your analysis in JSON format:
{
  "result": "false_positive" or "confirmed",
  "reasoning": "explain what BigQuery search showed (especially if 0 rows) and why it led to this verdict",
  "evidence": ["specific log findings from BigQuery", "or note that 0 rows were found"]
}`, [
		input.title,
		get_category(input),
		get_resource_name(input),
	]),
	"format": "json",
} if {
	# Critical 또는 High severity 알림만 상세 분석
	some attr in input.attributes
	attr.key == "severity"
	attr.value in ["CRITICAL", "HIGH"]
}

# 카테고리 조회 헬퍼
get_category(alert) = category if {
	some attr in alert.attributes
	attr.key == "category"
	category := attr.value
}

get_category(alert) = "unknown" if {
	not [attr | some attr in alert.attributes; attr.key == "category"][0]
}

# 리소스명 조회 헬퍼
get_resource_name(alert) = resource if {
	some attr in alert.attributes
	attr.key == "resource_name"
	resource := attr.value
}

get_resource_name(alert) = "unknown" if {
	not [attr | some attr in alert.attributes; attr.key == "resource_name"][0]
}
