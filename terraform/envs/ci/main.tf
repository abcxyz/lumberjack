/**
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

module "e2e" {
  source        = "../../modules/e2e"
  folder_parent = "folders/316290568068"
  top_folder_id = "github-ci"

  // The billing account 'Gong Test'.
  billing_account = "016242-61A3FB-F92462"

  tag              = var.tag
  renew_random_tag = var.renew_random_tag
}
