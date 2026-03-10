package ingest

# AWS GuardDuty 알림의 속성 추출
alert contains {
	"title": sprintf("GuardDuty: %s", [input.type]),
	"description": sprintf("Severity %d alert in %s", [input.severity, input.region]),
	"attributes": extract_guardduty_attributes(input),
} if {
	input.service.serviceName == "guardduty"
	# 기각 조건: 테스트 알림이나 화이트리스트 IP는 제외
	not is_test_alert
	not is_whitelisted
}

# 헬퍼: 테스트 알림 판정
is_test_alert if {
	input.environment == "development"
	input.test == true
}

# 헬퍼: 화이트리스트 판정
is_whitelisted if {
	some ip in input.source_ips
	ip in data.whitelist.ips
}

# GuardDuty 속성 추출 헬퍼
extract_guardduty_attributes(alert_data) = attrs if {
	attrs := [
		{
			"key": "aws_account_id",
			"value": alert_data.accountId,
			"type": "string",
		},
		{
			"key": "resource_type",
			"value": alert_data.resource.resourceType,
			"type": "string",
		},
		{
			"key": "severity",
			"value": sprintf("%d", [alert_data.severity]),
			"type": "number",
		},
		{
			"key": "region",
			"value": alert_data.region,
			"type": "string",
		},
	]
}
