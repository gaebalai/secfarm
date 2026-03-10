package triage

# 기본값은 일반 알림으로 수리
default action = "accept"

default severity = "medium"

default note = ""

# 알려진 오탐 패턴은 기각 (최우선)
action = "discard" if {
	contains(input.alert.title, "scheduled maintenance")
}

severity = "info" if {
	contains(input.alert.title, "scheduled maintenance")
}

note = "Known false positive pattern" if {
	contains(input.alert.title, "scheduled maintenance")
}

# Enrich 실행 결과에서 악의를 감지한 경우 critical
severity = "critical" if {
	some exec in input.enrich
	contains(exec.result, "malicious")
}

note = "Malicious activity detected" if {
	some exec in input.enrich
	contains(exec.result, "malicious")
}

# Enrich 실행 결과에서 의심스러운 활동을 감지한 경우 high
severity = "high" if {
	some exec in input.enrich
	contains(exec.result, "suspicious")
	not contains(exec.result, "malicious")
}

# 특정 enrich 결과 ID에 기반한 판정도 가능
severity = "critical" if {
	some exec in input.enrich
	exec.id == "ip_threat_intel"
	contains(exec.result, "malicious")
}

# 원본 알림의 severity도 참고
severity = "critical" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	to_number(attr.value) >= 8
}

severity = "high" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	sev := to_number(attr.value)
	sev >= 5
	sev < 8
}

severity = "low" if {
	some attr in input.alert.attributes
	attr.key == "severity"
	to_number(attr.value) < 2
}
