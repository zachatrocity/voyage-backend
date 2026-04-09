# Find one message to use in subsequent requests
GET {{base_url}}/api/v1/search?q={{search_query}}&limit=1
X-API-Key: {{api_key}}
HTTP 200
[Captures]
message_id: jsonpath "$.results[0].message_id"
[Asserts]
jsonpath "$.results[0].message_id" exists

# Read the message by id
GET {{base_url}}/api/v1/email/{{message_id}}
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.message_id" == "{{message_id}}"

# Capture an existing tag and re-apply it (minimal mutation / idempotent)
GET {{base_url}}/api/v1/email/{{message_id}}
X-API-Key: {{api_key}}
HTTP 200
[Captures]
existing_tag: jsonpath "$.tags[0]"
[Asserts]
jsonpath "$.tags[0]" exists

POST {{base_url}}/api/v1/email/{{message_id}}/tags/{{existing_tag}}
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.message_id" == "{{message_id}}"
jsonpath "$.tags" contains "{{existing_tag}}"
