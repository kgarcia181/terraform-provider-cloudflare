# Migration Models

This directory contains simplified model adapters for the struct_transform approach.

## Why Adapter Models?

The actual provider models in `external/` are complex:
- V4 uses SDK v2 with schema-based approach (no models)
- V5 uses Framework with complex types (`types.String`, `customfield.Set`, etc.)

These adapter models provide:
1. Simple Go structs for easier manipulation
2. Conversion functions to/from provider models
3. A clean interface for migration logic

## Structure

- `v4/` - Simplified v4 resource representations
- `v5/` - Simplified v5 resource representations
- `adapters/` - Conversion logic between simplified and provider models