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

variable "project_id" {
  description = "GCP project ID where the secrets belongs."
  type        = string
}

variable "labels" {
  description = "The labels assigned to the secret."
  type        = map(string)
  default     = {}
}

variable "gh_pk_accessor_members" {
  description = <<EOT
    The service accounts that need roles/secretmanager.secretAccessor role 
    for accessing keys stored in secret manager. Normally it will be jvs's
    api and ui service account.
  EOT
  type = list(string)
}

variable "gh_private_key_secret_id" {
  description = <<EOT
    ID of the secret. This must be unique within the project. This is also
    the secret's name.
  EOT
  type        = string
}
