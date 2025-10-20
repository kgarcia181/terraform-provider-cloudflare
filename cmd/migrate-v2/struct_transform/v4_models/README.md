# V4 Migration Models

These are simplified models used specifically for migration from v4 to v5.

## Why Simplified Models?

The actual v4 provider (`external/provider-v4/`) uses SDK v2 pattern which is schema-based, not model-based. These simplified models:

1. Provide a struct representation of v4 resources for easier manipulation
2. Map directly to the HCL attributes users write
3. Include only the fields relevant for migration

## Relationship to External Provider

The v4 provider schema can be found at:
- `external/provider-v4/internal/sdkv2provider/schema_cloudflare_record.go`

These models are derived from that schema but simplified for migration purposes.

## Models

- `dns_record.go` - Simplified cloudflare_record resource model