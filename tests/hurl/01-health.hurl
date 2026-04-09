GET {{base_url}}/health
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.status" == "up"
jsonpath "$.database" exists
jsonpath "$.timestamp" exists
