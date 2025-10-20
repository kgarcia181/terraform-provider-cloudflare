# Terraform Provider Cloudflare Migration Tool v2

A modular migration tool for upgrading Terraform configurations and state files from Cloudflare Provider v4 to v5.

## Architecture

This tool implements a **Strategy Pattern** combined with **Chain of Responsibility Pattern** for modular, extensible migrations.

### Components

- **CLI Layer** (`cmd/main.go`): Command-line interface and file processing
- **Core Engine** (`pipeline.go`): Orchestrates the transformation pipeline
- **Handlers** (`handlers/`): Chain of responsibility for processing stages
- **Resources** (`resources/`): Resource-specific transformation strategies
- **External** (`external/`): Git submodules for provider versions

## Prerequisites

- Go 1.21+
- Git

## Setup

### First-Time Setup

After cloning the repository, you need to initialize the git submodules that contain the provider versions:

```bash
# Option 1: Run the setup script
cd cmd/migrate-v2
./setup.sh

# Option 2: Manual setup
git submodule init
git submodule update --init --recursive
```

### For Existing Clones

If you've already cloned the repository but the submodules are empty:

```bash
# From repository root
git submodule update --init --recursive

# Or use the setup script
cd cmd/migrate-v2
./setup.sh
```

## Building

```bash
cd cmd/migrate-v2
make build
```

## Usage

```bash
# Dry run - preview changes without modifying files
./migrate-v2 -dryrun -config /path/to/terraform/config -state terraform.tfstate

# Migrate configuration files only
./migrate-v2 -config /path/to/terraform/config

# Migrate state file only
./migrate-v2 -state terraform.tfstate

# Migrate specific resources only
./migrate-v2 -config /path/to/terraform/config -resource dns_record,zone_settings_override

# Enable verbose logging
./migrate-v2 -verbose -config /path/to/terraform/config
```

### Flags

- `-config <dir>`: Directory containing Terraform configuration files (.tf)
- `-state <file>`: Path to Terraform state file (.tfstate)
- `-dryrun`: Preview changes without modifying files
- `-resource <list>`: Comma-separated list of resource types to migrate
- `-verbose`: Enable debug logging

## Supported Resources

Currently supported resource migrations:

- `cloudflare_record` → `cloudflare_dns_record`
- `cloudflare_zone_settings_override` → Individual `cloudflare_zone_setting` resources

## Development

### Adding a New Resource Migration

1. Create a new transformer in `resources/<resource_name>/`
2. Implement the `ResourceTransformer` interface
3. Register it in `resources/factory.go`

### Testing

```bash
make test
```

### Test Data

Example configurations and state files are available in `testdata/`:

```bash
# Test with example data
./migrate-v2 -dryrun -config ./testdata/cloudflare_record
```

## Submodules

This tool uses git submodules to reference the actual provider implementations:

- `external/provider-v4`: Cloudflare Provider v4 (branch: v4)
- `external/provider-v5`: Cloudflare Provider v5 (branch: next)

To update submodules to latest commits:

```bash
git submodule update --remote --merge
```

## Architecture Details

### Transformation Pipeline

1. **PreprocessHandler**: String-level transformations (pre-parsing)
2. **ParseHandler**: Parse HCL to AST
3. **ResourceTransformHandler**: Apply resource-specific strategies
4. **FormatterHandler**: Format and output HCL

### Strategy Pattern

Each resource type implements its own transformation strategy:

```go
type ResourceTransformer interface {
    CanHandle(resourceType string) bool
    TransformConfig(block *hclwrite.Block) (*TransformResult, error)
    TransformState(json gjson.Result, path string) (string, error)
    Preprocess(content string) string
}
```

## License

See the main repository LICENSE file.