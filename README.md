# terraform-provider-openai

Terraform provider for OpenAI platform admin API — projects, API keys, and service accounts.

## Usage

```hcl
terraform {
  required_providers {
    openai = {
      source  = "stark256-spec/openai"
      version = "~> 1.0"
    }
  }
}

provider "openai" {
  api_key = var.openai_api_key
}

resource "openai_project" "eng" {
  name = "engineering"
}

resource "openai_service_account" "deployer" {
  project_id = openai_project.eng.id
  name       = "deploy-bot"
}

resource "openai_api_key" "ci" {
  project_id = openai_project.eng.id
  name       = "ci-pipeline"
}
```

## Authentication

Set your API key via the `api_key` argument or the environment variable shown in the provider schema.

## Resources

| Resource | Description |
|----------|-------------|
| `openai_workspace` / `openai_project` / `openai_team` | Isolated environment |
| `openai_api_key` | API key scoped to a workspace/project |

## License

Apache 2.0
