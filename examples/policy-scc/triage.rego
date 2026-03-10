package triage

# 기본값은 수리
default action = "accept"

default severity = "high"

default note = ""

# Enrich 결과가 false_positive인 경우 기각
action = "discard" if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "false_positive"
}

severity = "info" if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "false_positive"
}

note = reasoning if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "false_positive"
	reasoning := sprintf("False positive: %s", [analysis.reasoning])
}

# 확인된 위협의 경우 critical
severity = "critical" if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "confirmed"
}

note = reasoning if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "confirmed"
	reasoning := sprintf("Confirmed threat: %s", [analysis.reasoning])
}

# 원래의 SCC severity도 고려
severity = "critical" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	attr.value == "CRITICAL"
	# Enrich 결과가 없거나 false_positive가 아닌 경우
	not is_false_positive
}

severity = "high" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	attr.value == "HIGH"
	not is_false_positive
}

severity = "medium" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	attr.value == "MEDIUM"
	not is_false_positive
}

# false_positive 판정 헬퍼
is_false_positive if {
	some exec in input.enrich
	exec.id == "bigquery_impact_analysis"
	analysis := json.unmarshal(exec.result)
	analysis.result == "false_positive"
}
