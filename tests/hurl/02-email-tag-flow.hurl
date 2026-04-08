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

# Add a lightweight smoke tag
POST {{base_url}}/api/v1/email/{{message_id}}/tags/hurl-smoke-{{run_id}}
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.message_id" == "{{message_id}}"
jsonpath "$.tags" contains "hurl-smoke-{{run_id}}"
