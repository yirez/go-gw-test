# Oddities and Notes

- it looks like it makes sense to just use `POST /api/v1/*` for the proxy endpoint, but the gateway will proxy all HTTP methods on `/api/v1/*` since limiting to POST is likely insufficient for real usage.

