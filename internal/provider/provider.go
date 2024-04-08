package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ProxmoxVEKubeadmProvider{}

// ProxmoxVEKubeadmProvider defines the provider implementation.
type ProxmoxVEKubeadmProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ProxmoxVEKubeadmProviderModel describes the provider data model.
type ProxmoxVEKubeadmProviderModel struct {
	APIEndpoint   types.String `tfsdk:"api_endpoint"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
	Nodes         types.List   `tfsdk:"nodes"`
	Auth          struct {
		API struct {
			Username types.String `tfsdk:"username"`
			Password types.String `tfsdk:"password"`
		} ` tfsdk:"api"`
		SSH struct {
			Username   types.String `tfsdk:"username"`
			Password   types.String `tfsdk:"password"`
			PrivateKey types.String `tfsdk:"private_key"`
		} `tfsdk:"ssh"`
	} `tfsdk:"auth"`
}

func (p *ProxmoxVEKubeadmProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kubeadm"
	resp.Version = p.version
}

func (p *ProxmoxVEKubeadmProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "Hostname for your Proxmox Web UI/API server. If not running on port `8006`, include `:port` in the hostname",
				Required:            true,
			},
			"skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Set to true if using a self signed TLS certificate",
				Optional:            true,
			},
			"nodes": schema.ListAttribute{
				MarkdownDescription: "The node names where VMs can be placed. If omitted, VMs will be placed across all available nodes",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"auth": schema.SingleNestedAttribute{
				MarkdownDescription: "Required provider authentication information",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"api": schema.SingleNestedAttribute{
						Required:            true,
						MarkdownDescription: "Authentication information for the Proxmox API",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Proxmox API/GUI user with all necessary permmissions. Required with `password` if `token` is not set",
							},
							"password": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "Proxmox API/GUI user password. Required with `username` if `token` is not set",
							},
							"token": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "API token for Proxmox user in form `user@realm!token-name=token-secret. Required if username and password are not set, and takes precedence if they are",
							},
						},
						Validators: nil, //TODO
					},
					"ssh": schema.SingleNestedAttribute{
						Required:            true,
						MarkdownDescription: "SSH authentication information for the Proxmox nodes",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "username on the node vms with passwordless sudo (should be defined in the template)",
							},
							"password": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "Password for the user defined in `username`. Required unless `private_key` is set",
							},
							"private_key": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "PEM encoded _private_ key for the user defined in `username`. Required unless `password` is set. The public key must already exist on the template VM",
							},
						},
						Validators: nil, //TODO
					},
				},
			},
		},
	}
}

func (p *ProxmoxVEKubeadmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProxmoxVEKubeadmProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ProxmoxVEKubeadmProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *ProxmoxVEKubeadmProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

// func (p *ProxmoxVEKubeadmProvider) Functions(ctx context.Context) []func() function.Function {
// 	return nil
// }

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProxmoxVEKubeadmProvider{
			version: version,
		}
	}
}
