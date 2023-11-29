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

locals {
  # This is the key id under secret manager, where user need to
  # manully upload the github app private key to.
  gh_private_key_id="github_app_private_key"
}

module "jvs" {
  source = "git::https://github.com/abcxyz/jvs.git//terraform/e2e?ref=main" # this should be pinned to the SHA desired

  project_id = "YOUR_PROJECT_ID"

  # Specify who can access JVS.
  jvs_invoker_members = ["user:foo@example.com"]

  # Use your own domain.
  jvs_api_domain = "api.jvs.example.com"
  jvs_ui_domain  = "web.jvs.example.com"

  iap_support_email = "support@example.com"

  jvs_container_image = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvs-plugin-github:0.0.1"


  # Specify the Email address where alert notifications send to.
  notification_channel_email = "foo@example.com"

  jvs_prober_image = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvs-prober:0.0.5-amd64"

  # Specify the plugin environment variables. See the file below for details:
  # https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go
  plugin_envvars = {
    "GITHUB_APP_ID" : "YOUR_GITHUB_APP_ID",
    "GITHUB_APP_INSTALLATION_ID" : "YOUR_GITHUB_APP_INSTALLATION_ID",
    "GITHUB_PLUGIN_DISPLAY_NAME" : "jvs plugin github",
    "GITHUB_PLUGIN_HINT" : "jvs plugin github hint"
  }

    plugin_secret_envvars = {
      "GITHUB_SECRET": {
        name: local.gh_private_key_id,
        version: "latest",
    }
  }
}

module "github_plugin" {
  # Pin to proper version.
  source = "git::https://github.com/abcxyz/jvs-plugin-github.git//terraform/modules/secret_manager?ref=69fdda2fb914a28e89d352d86a501397f4ddcaad"

  project_id = module.qinhang_jvs_plugin_github_dev.project_id

  gh_private_key_secret_id = local.gh_private_key_id

  gh_pk_accessor_members_map = {
    api_sa = module.jvs.jvs_api_service_account_member,
    ui_sa = module.jvs.jvs_ui_service_account_member
  }
}
