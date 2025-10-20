package migrate_v2

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"

	handlers2 "github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/handlers"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/interfaces"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/registry"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/struct_transform"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/struct_transform/resources"
)

func BuildPipeline(reg *registry.StrategyRegistry) *Pipeline {
	return BuildPipelineWithOptions(reg, PipelineOptions{})
}

func BuildPipelineWithDryRun(reg *registry.StrategyRegistry, dryRun bool) *Pipeline {
	return BuildPipelineWithOptions(reg, PipelineOptions{DryRun: dryRun})
}

func BuildStatePipeline(reg *registry.StrategyRegistry, dryRun bool) *Pipeline {
	return BuildPipelineWithOptions(reg, PipelineOptions{
		DryRun:  dryRun,
		IsState: true,
	})
}

// PipelineOptions contains configuration options for the pipeline
type PipelineOptions struct {
	// UseStructMode enables the struct-based transformation approach
	UseStructMode bool
	// DryRun indicates whether this is a dry run
	DryRun bool
	// IsState indicates whether this pipeline is for state transformation
	IsState bool
}

// BuildPipelineWithOptions creates a pipeline with specified options
func BuildPipelineWithOptions(reg *registry.StrategyRegistry, opts PipelineOptions) *Pipeline {
	// State pipelines don't use the regular handler chain
	if opts.IsState {
		return &Pipeline{
			handler:  nil, // State transformation doesn't use handlers
			registry: reg,
			dryRun:   opts.DryRun,
			isState:  true,
		}
	}

	builder := NewPipelineBuilder().
		WithPreprocessing(reg).
		WithParsing()

	// Choose transformation approach based on options
	if opts.UseStructMode {
		// Create a registry with struct-based transformers
		structReg := registry.NewStrategyRegistry()
		// Register all struct-based transformers from factory
		resources.RegisterAllStructTransformers(structReg)

		// Use struct-based transformation handler with struct registry
		builder = builder.WithStructTransformation(structReg)
	} else {
		// Use traditional AST-based transformation
		builder = builder.WithResourceTransformation(reg)
	}

	pipeline := builder.
		// Cross-resource migrations would go here
		// WithHandler(handlers.NewCrossResourceHandler()).
		// Import generation for split resources would go here
		// WithHandler(handlers.NewImportGeneratorHandler()).
		// Validation would go here
		// WithHandler(handlers.NewValidationHandler()).
		WithFormatting().
		Build()

	pipeline.dryRun = opts.DryRun
	return pipeline
}

type Pipeline struct {
	handler  interfaces.TransformationHandler
	registry *registry.StrategyRegistry
	dryRun   bool
	isState  bool
}

func NewPipeline(reg *registry.StrategyRegistry) *Pipeline {
	pipeline := BuildPipeline(reg)
	return pipeline
}

func (p *Pipeline) Transform(content []byte, filename string) ([]byte, error) {
	ctx := &interfaces.TransformContext{
		Content:     content,
		Filename:    filename,
		Diagnostics: nil,
		Metadata:    make(map[string]interface{}),
		DryRun:      p.dryRun,
	}

	result, err := p.handler.Handle(ctx)
	if err != nil {
		return nil, err
	}

	if result.Diagnostics.HasErrors() {
		return nil, result.Diagnostics.Errs()[0]
	}

	return result.Content, nil
}

// TransformState transforms a Terraform state file
func (p *Pipeline) TransformState(content []byte, filename string) ([]byte, error) {
	if !p.isState {
		return nil, fmt.Errorf("pipeline not configured for state transformation")
	}

	jsonStr := string(content)
	result := jsonStr

	// Process each resource
	resources := gjson.Get(jsonStr, "resources")
	if !resources.Exists() {
		return content, nil
	}

	// Iterate backwards to avoid index shifting issues when deleting
	resourcesArray := resources.Array()
	for i := len(resourcesArray) - 1; i >= 0; i-- {
		resourcePath := fmt.Sprintf("resources.%d", i)
		resourceType := gjson.Get(result, resourcePath+".type").String()

		// Find the appropriate transformer
		transformer := p.registry.Find(resourceType)
		if transformer != nil {
			// Apply state transformation
			resourceJSON := gjson.Get(result, resourcePath)
			transformed, err := transformer.TransformState(resourceJSON, resourcePath)
			if err != nil {
				return nil, fmt.Errorf("failed to transform %s state: %w", resourceType, err)
			}

			// If transformer returns empty string, remove the resource
			if transformed == "" {
				result, _ = sjson.Delete(result, resourcePath)
			} else {
				// Parse the transformed JSON and update the result
				var transformedMap map[string]interface{}
				if err := json.Unmarshal([]byte(transformed), &transformedMap); err == nil {
					for key, value := range transformedMap {
						result, _ = sjson.Set(result, resourcePath+"."+key, value)
					}
				}
			}
		}
	}

	// Pretty format the JSON output
	formatted := pretty.Pretty([]byte(result))
	return formatted, nil
}

type PipelineBuilder struct {
	handlers []interfaces.TransformationHandler
	registry *registry.StrategyRegistry
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		handlers: make([]interfaces.TransformationHandler, 0),
		registry: nil,
	}
}

func (b *PipelineBuilder) WithPreprocessing(reg *registry.StrategyRegistry) *PipelineBuilder {
	b.handlers = append(b.handlers, handlers2.NewPreprocessHandler(reg))
	return b
}

func (b *PipelineBuilder) WithParsing() *PipelineBuilder {
	b.handlers = append(b.handlers, handlers2.NewParseHandler())
	return b
}

func (b *PipelineBuilder) WithResourceTransformation(reg *registry.StrategyRegistry) *PipelineBuilder {
	b.registry = reg
	b.handlers = append(b.handlers, handlers2.NewResourceTransformHandler(reg))
	return b
}

func (b *PipelineBuilder) WithStructTransformation(structReg *registry.StrategyRegistry) *PipelineBuilder {
	b.handlers = append(b.handlers, struct_transform.NewStructTransformHandler(structReg))
	return b
}

func (b *PipelineBuilder) WithFormatting() *PipelineBuilder {
	b.handlers = append(b.handlers, handlers2.NewFormatterHandler())
	return b
}

func (b *PipelineBuilder) WithHandler(handler interfaces.TransformationHandler) *PipelineBuilder {
	b.handlers = append(b.handlers, handler)
	return b
}

func (b *PipelineBuilder) Build() *Pipeline {
	if len(b.handlers) == 0 {
		return nil
	}
	for i := 0; i < len(b.handlers)-1; i++ {
		b.handlers[i].SetNext(b.handlers[i+1])
	}

	return &Pipeline{
		handler:  b.handlers[0],
		registry: b.registry,
	}
}
