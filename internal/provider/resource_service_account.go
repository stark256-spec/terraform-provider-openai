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

var _ resource.Resource = &ServiceAccountResource{}

type ServiceAccountResource struct{ client *OpenAIClient }

type ServiceAccountResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewServiceAccountResource() resource.Resource { return &ServiceAccountResource{} }

func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI project service account (non-human machine identity).",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Service account ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "Project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":       schema.StringAttribute{Required: true, MarkdownDescription: "Service account name.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"role":       schema.StringAttribute{Computed: true, MarkdownDescription: "Role assigned to the service account."},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Unix creation timestamp."},
		},
	}
}

func (r *ServiceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sa, err := r.client.CreateServiceAccount(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create service account failed", err.Error())
		return
	}
	plan.ID = types.StringValue(sa.ID)
	plan.Role = types.StringValue(sa.Role)
	plan.CreatedAt = types.Int64Value(sa.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sa, err := r.client.GetServiceAccount(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Role = types.StringValue(sa.Role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceAccountResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Service accounts are immutable. Replace to update.")
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteServiceAccount(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete service account failed", err.Error())
	}
}
