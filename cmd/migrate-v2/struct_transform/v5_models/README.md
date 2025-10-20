# V5 Migration Models

These are simplified models used specifically for migration from v4 to v5.

## Why Simplified Models?

The actual v5 provider (`external/provider-v5/`) uses Terraform Framework pattern with complex types:
- `types.String` instead of `string`
- `customfield.Set[types.String]` for sets
- `timetypes.RFC3339` for timestamps
- `jsontypes.Normalized` for JSON fields

These simplified models:
1. Use basic Go types for easier manipulation during migration
2. Focus only on fields that change between v4 and v5
3. Provide a clean transformation target

## Relationship to External Provider

The v5 provider models can be found at:
- `external/provider-v5/internal/services/dns_record/model.go`

These simplified models are inspired by but simpler than the actual v5 models.

## Models

- `dns_record.go` - Simplified cloudflare_dns_record resource model

## Future Enhancement

Could potentially generate converters between these simplified models and the actual provider models if needed for state migration.