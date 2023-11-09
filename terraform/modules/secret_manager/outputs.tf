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

output "gh_private_key_secret_id" {
  value = google_secret_manager_secret.gh_app_private_key.id
}

output "gh_private_key_secret_name" {
  value = google_secret_manager_secret.gh_app_private_key.name
}

output "gh_private_key_secret_version_name" {
  value = google_secret_manager_secret_version.gh_app_private_key_version.name
}
