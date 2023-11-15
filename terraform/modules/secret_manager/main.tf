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
  for_each = toset([
    "secretmanager.googleapis.com"
  ])

  project = var.project_id

  service                    = each.key
  disable_on_destroy         = false
  disable_dependent_services = false
}

# Secret Manager secrets for github plugin to use.
resource "google_secret_manager_secret" "gh_app_private_key" {
  project = var.project_id

  secret_id = var.gh_private_key_secret_id
  labels    = var.labels

  replication {
    auto {}
  }

  depends_on = [
    google_project_service.services
  ]
}

resource "google_secret_manager_secret_version" "gh_app_private_key_version" {
  secret = google_secret_manager_secret.gh_app_private_key.id

  # default value used for initial revision to allow cloud run to map the secret
  # to manage this value and versions, use the google cloud web application
  secret_data = "DEFAULT_VALUE"

  lifecycle {
    ignore_changes = [
      enabled,

      # Ignore secret data so Terraform won't reset the secret.
      # Operator will need to put the github_app credential into the secret manually.
      secret_data,
    ]
  }
}

// grant service account to access the secret, so the services can use that
// for auth with github.
resource "google_secret_manager_secret_iam_member" "gh_app_pk_accessor" {
  for_each = var.gh_pk_accessor_members_map

  project = var.project_id

  secret_id = google_secret_manager_secret.gh_app_private_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = each.value
}
