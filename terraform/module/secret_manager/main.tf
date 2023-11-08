# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

resource "google_project_service" "services" {
  project = var.project_id

  service                    = "secretmanager.googleapis.com"
  disable_on_destroy         = false
  disable_dependent_services = false
}

# Secret Manager secrets for github plugin to use.
resource "google_secret_manager_secret" "github_app_pem" {
  project = var.project_id

  secret_id = var.secret_id
  labels    = var.labels
  replication {
    auto {}
  }

  depends_on = [
    google_project_service.services
  ]
}

resource "google_secret_manager_secret_version" "jira_api_token_version" {
  secret = google_secret_manager_secret.jira_api_token.id

  # default value used for initial revision to allow cloud run to map the secret
  # to manage this value and versions, use the google cloud web application
  secret_data = "DEFAULT_VALUE"

  lifecycle {
    ignore_changes = [
      enabled,

      # Ignore secret data so Terraform won't reset the secret.
      # Operator will need to put the JIRA credential into the secret manually.
      secret_data,
    ]
  }
}