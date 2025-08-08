## main

400 Bad Request
	•	Unsupported Content-Type
	•	The request can be accessed only by guests.
	•	The request can be accessed only by authenticated admins.
	•	The request can be accessed only by authenticated records.
	•	The request requires valid admin authorization token.
	•	The request requires valid record authorization token.
	•	Missing or invalid collection context.
	•	Missing required "expand" parameter.
	•	Missing required "filter" parameter.
	•	Missing required "fields" parameter.
	•	Missing required "sort" parameter.
	•	Invalid "sort" parameter format.
	•	Invalid "expand" parameter format.
	•	Invalid "filter" parameter format.
	•	Invalid "fields" parameter format.
	•	Invalid "page" parameter format.
	•	Invalid "perPage" parameter format.
	•	Invalid "skipTotal" parameter format.
	•	Invalid request payload.
	•	Invalid or missing file.
	•	Invalid request body. (multipart fail)
	•	Invalid or missing field "<field name>"
	•	Invalid or missing record id.
	•	Invalid or missing password reset token.
	•	Invalid or missing verification token.
	•	Invalid or missing file token.
	•	Invalid or missing OAuth2 state parameter.
	•	Invalid or missing redirect URL.
	•	Invalid redirect status code.
	•	Failed to fetch admins info.
	•	Failed to fetch records info.
	•	Failed to fetch collection schema.

⸻

401 Unauthorized
	•	Missing or invalid authentication.
	•	Missing or invalid authentication token.
	•	Missing or invalid admin authorization token.
	•	Missing or invalid record authorization token.

⸻

403 Forbidden
	•	You are not allowed to perform this request.
	•	The authorized record is not allowed to perform this action.
	•	The authorized admin is not allowed to perform this action.

⸻

404 Not Found
	•	The requested resource wasn't found.
	•	File not found.
	•	Collection not found.
	•	Record not found.

⸻

413 Request Entity Too Large
	•	Request entity too large
(BodyLimit)

⸻

429 Too Many Requests
	•	Too Many Requests. (RateLimit)

⸻

500 Internal Server Error
	•	Something went wrong while processing your request.
(panic, db errer etc)

⸻

## v0.22.x (v0.22.34)

400 Bad Request
	•	Unsupported Content-Type
	•	The request can be accessed only by guests.
	•	The request can be accessed only by authenticated admins.
	•	The request can be accessed only by authenticated records.
	•	The request requires valid admin authorization token.
	•	The request requires valid record authorization token.
	•	Missing or invalid collection context.
	•	Missing required "expand" parameter.
	•	Missing required "filter" parameter.
	•	Missing required "fields" parameter.
	•	Missing required "sort" parameter.
	•	Invalid "sort" parameter format.
	•	Invalid "expand" parameter format.
	•	Invalid "filter" parameter format.
	•	Invalid "fields" parameter format.
	•	Invalid "page" parameter format.
	•	Invalid "perPage" parameter format.
	•	Invalid "skipTotal" parameter format.
	•	Invalid request payload.
	•	Invalid or missing file.
	•	Invalid request body.
	•	Invalid or missing field "<field name>"
	•	Invalid or missing record id.
	•	Invalid or missing password reset token.
	•	Invalid or missing verification token.
	•	Invalid or missing file token.
	•	Invalid or missing OAuth2 state parameter.
	•	Invalid or missing redirect URL.
	•	Invalid redirect status code.

⸻

401 Unauthorized
	•	Missing or invalid authentication token.
	•	Missing or invalid admin authorization token.
	•	Missing or invalid record authorization token.

⸻

403 Forbidden
	•	You are not allowed to perform this request.
	•	The authorized record is not allowed to perform this action.
	•	The authorized admin is not allowed to perform this action.

⸻

404 Not Found
	•	The requested resource wasn't found.
	•	File not found.
	•	Collection not found.
	•	Record not found.

⸻

413 Request Entity Too Large
	•	Request entity too large

⸻

500 Internal Server Error
	•	Something went wrong while processing your request.
(panic, db errer etc)
