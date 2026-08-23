# Security / Sensitive Data Audit

The final production hardening audit verifies that sensitive values are not intentionally emitted through application logs, Prometheus labels, or HTTP error responses.

## Never log

- passwords
- authorization headers
- bearer tokens
- secrets
- connection strings / DSNs
- raw SQL
- request bodies

## Prometheus label policy

Community metrics use only bounded labels:

- operation/status where defined by the metric contract
- validation `field`

The following must never become labels:

- `request_id`
- `community_id`
- `name`
- `slug`
- raw error messages

## HTTP errors

Unexpected repository/database errors are mapped to the stable `internal_error` response. Validation errors may expose a safe validation message, but raw database/SQL details are not exposed.

## Automated checks

The production verification script searches logging code for sensitive terms and verifies the known metric-label contract through the integration suite.
