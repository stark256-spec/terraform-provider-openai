terraform {
  required_providers {
    openai = {
      source  = "stark256-spec/openai"
      version = "~> 1.0"
    }
  }
}

provider "openai" {
  api_key = var.openai_api_key   # or OPENAI_API_KEY env var
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

output "ci_key" {
  value     = openai_api_key.ci.secret_key
  sensitive = true
}