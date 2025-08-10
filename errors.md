## main

400 Bad Request
	•	Unsupported Content-Type
	•	The request can be accessed only by guests.
	•	The request can be accessed only by authenticated admins.
	•	The request can be accessed only by authenticated records.
	•	The request requires valid admin authorization token.
	•	The request requires valid record authorization token.
	•	Missing or invalid collection context.
	// "Missing required '...'" errors are dynamically generated, not in static list
	•	Invalid "sort" parameter format.
	•	Invalid "expand" parameter format.
	•	Invalid "filter" parameter format.
	•	Invalid "fields" parameter format.
	•	Invalid "page" parameter format.
	•	Invalid "perPage" parameter format.
	•	Invalid "skipTotal" parameter format.
	•	Invalid request payload.
	•	Invalid or missing file.
	•	Invalid request body. // (mainly when multipart parsing fails)
	•	Invalid or missing field "<field name>" // dynamically generated
	•	Invalid or missing record id.
	•	Invalid or missing password reset token.
	•	Invalid or missing verification token.
	•	Invalid or missing file token.
	•	Invalid or missing OAuth2 state parameter.
	•	Invalid or missing redirect URL.
	•	Invalid redirect status code.
	•	Failed to authenticate. // occurs when admin or record login fails

⸻

401 Unauthorized
	•	Missing or invalid authentication token. // [corrected] "Missing or invalid authentication." doesn't exist, this message is used
	•	Missing or invalid admin authorization token.
	•	Missing or invalid record authorization token.

⸻

403 Forbidden
	•	You are not allowed to perform this request.
	•	The authorized record is not allowed to perform this action.
	•	You are not allowed to perform this action. // [corrected] same as before, this is used instead of "The authorized admin is not allowed to perform this action."

⸻

404 Not Found
	•	The requested resource wasn't found.
	•	File not found.
	•	Collection not found.
	•	Record not found.

⸻

413 Request Entity Too Large
	•	Request entity too large.

⸻

429 Too Many Requests
	•	Too Many Requests.

⸻

500 Internal Server Error
	•	Something went wrong while processing your request.

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
	•	Invalid or missing field "<field name>" // dynamically generated
	•	Invalid or missing record id.
	•	Invalid or missing password reset token.
	•	Invalid or missing verification token.
	•	Invalid or missing file token.
	•	Invalid or missing OAuth2 state parameter.
	•	Invalid or missing redirect URL.
	•	Invalid redirect status code.
	•	Failed to authenticate.
	•	The new email address is already in use.
	•	The provided old password is not valid.

// Error codes defined in core/validation.go
// General validation (ozzo-validation library defaults)
	•	validation_is_required          // when required field is empty
	•	validation_nil_or_not_empty     // when field that should be empty has a value
	•	validation_in_invalid           // when value is not in allowed list
	•	validation_not_in_invalid       // when value is in disallowed list
	•	validation_length_too_short     // when shorter than minimum length
	•	validation_length_too_long      // when longer than maximum length
	•	validation_length_invalid       // when length is not in valid range
	•	validation_min_greater_equal    // when less than minimum value
	•	validation_min_invalid          // when less than minimum value (internally same as above)
	•	validation_max_less_equal       // when greater than maximum value
	•	validation_max_invalid          // when greater than maximum value (internally same as above)
	•	validation_match_invalid        // when doesn't match regex pattern
	•	validation_invalid_email        // when not a valid email format
	•	validation_invalid_url          // when not a valid URL format
	•	validation_invalid_ip           // when not a valid IP address format
	•	validation_invalid_ipv4         // when not a valid IPv4 address format
	•	validation_invalid_ipv6         // when not a valid IPv6 address format
	•	validation_date_invalid         // when not a valid date/time format (RFC3339)
	•	validation_date_too_early       // when earlier than specified date
	•	validation_date_too_late        // when later than specified date

// PocketBase custom validation
	•	validation_invalid_alphanumeric   // when not composed only of letters and numbers
	•	validation_invalid_json           // when not valid JSON format
	•	validation_invalid_slug           // when not valid slug format (lowercase, numbers, hyphens)
	•	validation_not_slug               // when field name is "slug" (collection field name rule)
	•	validation_invalid_system_collection_name // when collection name starts with "pb_" prefix
	•	validation_unique                 // when value already exists (e.g., duplicate email)
	•	validation_values_exist           // when relation field has non-existent ID as value

⸻

401 Unauthorized
	•	Missing or invalid authentication token.
	•	Missing or invalid admin authorization token.
	•	Missing or invalid record authorization token.

⸻

403 Forbidden
	•	You are not allowed to perform this request.
	•	The authorized record is not allowed to perform this action.
	•	You are not allowed to perform this action.

⸻

404 Not Found
	•	The requested resource wasn't found.
	•	File not found.
	•	Collection not found.
	•	Record not found.

⸻

413 Request Entity Too Large
	•	Request entity too large.




