package handlers

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/interfaces"
)

// FormatterHandler formats the transformed AST back into HCL text
type FormatterHandler struct {
	interfaces.BaseHandler
}

// NewFormatterHandler creates a new formatting handler
func NewFormatterHandler() interfaces.TransformationHandler {
	return &FormatterHandler{}
}

// Handle formats the AST and converts it back to bytes
func (h *FormatterHandler) Handle(ctx *interfaces.TransformContext) (*interfaces.TransformContext, error) {
	if ctx.AST == nil {
		return ctx, fmt.Errorf("AST is nil - cannot format")
	}

	bytes := ctx.AST.Bytes()
	formatted := hclwrite.Format(bytes)
	ctx.Content = formatted

	return h.CallNext(ctx)
}
