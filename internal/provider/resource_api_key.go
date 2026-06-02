package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIKeyResource{}

type APIKeyResource struct{ client *OpenAIClient }

type APIKeyResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	RedactedValue types.String `tfsdk:"redacted_value"`
	SecretKey     types.String `tfsdk:"secret_key"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}

func NewAPIKeyResource() resource.Resource { return &APIKeyResource{} }

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI project API key.\n\n> The full `secret_key` is only available at creation and stored in Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Key ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project_id":     schema.StringAttribute{Required: true, MarkdownDescription: "Project to scope this key to.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":           schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable key name."},
			"redacted_value": schema.StringAttribute{Computed: true, MarkdownDescription: "Redacted key value for identification.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"secret_key":     schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Full key value — only available at creation.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp of creation."},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	k, err := r.client.CreateAPIKey(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create API key failed", err.Error())
		return
	}
	plan.ID = types.StringValue(k.ID)
	plan.RedactedValue = types.StringValue(k.RedactedValue)
	plan.CreatedAt = types.Int64Value(k.CreatedAt)
	if k.Value != nil {
		plan.SecretKey = types.StringValue(*k.Value)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	k, err := r.client.GetAPIKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Name = types.StringValue(k.Name)
	state.RedactedValue = types.StringValue(k.RedactedValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// OpenAI API keys are immutable except through delete+recreate
	resp.Diagnostics.AddError("Update not supported", "OpenAI API keys cannot be updated in-place. Use lifecycle { create_before_destroy = true } to rotate.")
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete API key failed", err.Error())
	}
}
