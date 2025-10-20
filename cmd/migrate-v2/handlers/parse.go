package handlers

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/interfaces"
)

// ParseHandler converts the preprocessed content into an HCL AST
type ParseHandler struct {
	interfaces.BaseHandler
}

// NewParseHandler creates a new parsing handler
func NewParseHandler() interfaces.TransformationHandler {
	return &ParseHandler{}
}

// Handle parses the HCL content into an AST
func (h *ParseHandler) Handle(ctx *interfaces.TransformContext) (*interfaces.TransformContext, error) {
	file, diags := hclwrite.ParseConfig(ctx.Content, ctx.Filename, hcl.InitialPos)

	ctx.Diagnostics = append(ctx.Diagnostics, diags...)
	if diags.HasErrors() {
		return ctx, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	ctx.AST = file

	return h.CallNext(ctx)
}
