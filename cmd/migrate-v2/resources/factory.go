package resources

import (
	"log/slog"

	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/interfaces"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/registry"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/resources/cloudflare_record"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/resources/cloudflare_zone_settings_override"
)

// ResourceFactory is a function that creates a new resource transformer
type ResourceFactory func() interfaces.ResourceTransformer

// ResourceFactories maps resource names to their factory functions
var ResourceFactories = map[string]ResourceFactory{
	"dns_record":             cloudflare_record.NewDNSRecord,
	"zone_settings_override": cloudflare_zone_settings_override.NewZoneSettingsOverride,
}

// RegisterFromFactories registers resource transformers with the registry
func RegisterFromFactories(reg *registry.StrategyRegistry, names ...string) {
	if len(names) == 0 {
		slog.Info("Registering all available resources")
		for _, factory := range ResourceFactories {
			reg.Register(factory())
		}
		return
	}
	// Register only specified resources
	slog.Debug("Registering specific resources", "resources", names)
	for _, name := range names {
		if factory, ok := ResourceFactories[name]; ok {
			reg.Register(factory())
		}
	}
}

// RegisterAll is a convenience function that registers all available resources
func RegisterAll(reg *registry.StrategyRegistry) {
	RegisterFromFactories(reg)
}

// GetAvailableResources returns a list of all available resource names
func GetAvailableResources() []string {
	names := make([]string, 0, len(ResourceFactories))
	for name := range ResourceFactories {
		names = append(names, name)
	}
	return names
}
