# Oddities and Notes

- it looks like it makes sense to just use `POST /api/v1/*` for the proxy endpoint, but the gateway will proxy all HTTP methods on `/api/v1/*` since limiting to POST is likely insufficient for real usage.
- I've created a source of truth for users/service accs in the db pushed via compose init. not the best but it works for our case.
- route matching is intentionally simple for now: exact path or `/*` prefix wildcard with boundary-safe prefix checks.
- local compose host ports for infra use non-default values to avoid local conflicts: postgres `5435`, redis `6389`.
