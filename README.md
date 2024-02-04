# tfectl
[![GitHub license](https://img.shields.io/github/license/AGLEnergyPublic/tfectl.svg)](https://github.com/AGLEnergyPublic/tfectl/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/AGLEnergyPublic/tfectl?status.svg)](https://godoc.org/github.com/AGLEnergyPublic/tfectl)
[![Go Report Card](https://goreportcard.com/badge/github.com/AGLEnergyPublic/tfectl)](https://goreportcard.com/report/github.com/AGLEnergyPublic/tfectl)
[![GitHub issues](https://img.shields.io/github/issues/AGLEnergyPublic/tfectl.svg)](https://github.com/AGLEnergyPublic/tfectl/issues)

* CLI Utility to query/manage TFE inspired by [tfe-cli](https://github.com/rgreinho/tfe-cli)

## Setup
* Copy the binary (either Windows or Linux) to a path on your machine. Add the `.exe` extension if using it on Windows
  ```ps
  PS> .\tfectl.exe
  Query TFE from the command line.

  Usage:
    tfectl [command]

  Available Commands:
    admin        Manage TFE admin operations
    completion   Generate the autocompletion script for the specified shell
    help         Help about any command
    policy       Query TFE policies
    policy-check Manage policy check workflows of a TFE run
    policy-set   Query TFE policy sets
    run          Manage TFE runs
    tag          Query TFE tags
    team         Manage TFE teams
    variable     Manage TFE workspace variables
    workspace    Manage TFE workspaces


  Flags:
    -h, --help                  help for tfectl
    -l, --log string            log level (debug, info, warn, error, fatal, panic)
    -o, --organization string   terraform organization or set TFE_ORG
    -q, --query string          JQ compatible query to parse JSON output
    -t, --token string          terraform token or set TFE_TOKEN
    -v, --version               version for tfectl

  Use "tfectl [command] --help" for more information about a command.
  ```

## Initialization
* `TFE_ADDRESS`: TFE URL defaults to `https://app.terraform.io/`
* `TFE_ORG`: TFE Organization
* `TFE_TOKEN`: token with read access to Organization specified in `TFE_ORG`
* Additionally `TFE_ORG` and `TFE_TOKEN` variables can be passed via CLI

## Usage
* To see available options
```bash
# /sbin/tfectl --help
Query TFE from the command line.

Usage:
  tfectl [command]

Available Commands:
  admin        Manage TFE admin operations
  completion   Generate the autocompletion script for the specified shell
  help         Help about any command
  policy       Query TFE policies
  policy-check Manage policy check workflows of a TFE run
  policy-set   Query TFE policy sets
  run          Manage TFE runs
  tag          Query TFE tags
  team         Manage TFE teams
  variable     Manage TFE workspace variables
  workspace    Manage TFE workspaces


Flags:
  -h, --help                  help for tfectl
  -l, --log string            log level (debug, info, warn, error, fatal, panic)
  -o, --organization string   terraform organization or set TFE_ORG
  -q, --query string          JQ compatible query to parse JSON output
  -t, --token string          terraform token or set TFE_TOKEN
  -v, --version               version for tfectl

Use "tfectl [command] --help" for more information about a command.
```

### Workspace
<details>
    <summary>Workspace Operations</summary>

* #### List
  * Run with no arguments to return the following for all workspaces in the Org

    | **Field**         | **Description**                                             | **Type** |
    |-------------------|-------------------------------------------------------------|----------|
    | name              | Name of the workspace                                       | string   |
    | id                | ID of the workspace                                         | string   |
    | locked            | Status of the workspace                                     | bool     |
    | execution_mode    | Whether the workspace runs remotely, locally or on an agent | string   |
    | terraform_version | Version of Terraform CLI running in the workspace           | string   |
    | tags              | List of tags against workspace                              | list     |

  * Run with `--filter`, which takes a workspace name or a substring of a name to get a filtered list of workspaces
 
  ```bash
    $ tfectl workspace list --filter workspace-1
    [
      {
        "name": "workspace-1",
        "id": "ws-RZP914jsX1Hmc9Yo"
        "locked": false,
        "execution_mode": "remote",
        "terraform_version": "1.3.0"
        "tags": [
            "tag:1",
            "tag:2"
        ]
      }
    ]
  ```

  * The `--filter` flag supports filtering by workspace tags using a prefix of `tags|`

  ```bash
    $ tfectl workspace list --filter "tags|tag:1,tag:2"
    [
      {
        "name": "workspace-1",
        "id": "ws-RZP914jsX1Hmc9Yo"
        "locked": false,
        "execution_mode": "remote",
        "terraform_version": "1.3.0"
        "tags": [
            "tag:1",
            "tag:2"
        ]
      },
      {
        "name": "workspace-2",
        "id": "ws-eLcff9y8r8bRBYfj"
        "locked": false,
        "execution_mode": "remote",
        "terraform_version": "1.3.7"
        "tags": [
            "tag:1",
            "tag:2"
        ]
      }
    ]
  ```

  * Run with the `--detail` flag to return the following details
    NOTE: This task takes a long time, it rate-limited and it is recommended to run it with the `--filter` argument

    | **Field**                  | **Description**                                                     | **Type** |
    |----------------------------|---------------------------------------------------------------------|----------|
    | name                       | Name of the workspace                                               | string   |
    | id                         | ID of the workspace                                                 | string   |
    | locked                     | Status of the workspace                                             | bool     |
    | execution_mode             | Whether the workspace runs remotely, locally or on an agent         | string   |
    | terraform_version          | Version of Terraform CLI running in the workspace                   | string   |
    | tags                       | List of tags against workspace                                      | list     |
    | created_days_ago           | How many days ago this workspace was created                        | string   |
    | updated_days_ago           | How many days ago this workspace was updated                        | string   |
    | last_remote_run_days_ago   | How many days ago was a remote run performed in this workspace      | string   |
    | last_state_update_days_ago | How many days ago was the terraform state updated in this workspace | string   |

 
  ```bash 
    $ tfectl workspace list --filter workspace-1 --detail
    [
      {
        "name": "workspace-1",
        "id": "ws-RZP914jsX1Hmc9Yo",
        "locked": false,
        "terraform_version": "1.3.0",
        "tags": [
            "tag:1",
            "tag:2"
        ]
        "created_days_ago": "819.167082",
        "updated_days_ago": "2.279692",
        "last_remote_run_days_ago": "2.281231",
        "last_state_update_days_ago": "30.174812"
      }
    ]
  ```
* #### Lock/Unlock
  * Run with a comma-separated string of workspaceIDs or a workspaceName filter (mutually exclusive)

  ```bash
    $ tfectl workspace lock --ids ws-SxWNNcYPkLD48ZC7
    [
      {
        "id": "ws-SxWNNcYPkLD48ZC7",
        "locked": true,
        "name": "test-workspace-1"
      }
    ] 
  ```

  * Operation can be run against a workspace that is already locked
  ```bash
    $ tfectl workspace lock --filter dev-workspace
    [
      {
        "id": "ws-5xUNCXVKrryoPcEp",
        "locked": true,
        "name": "dev-workspace"
      }
    ]
  ```

  * Optionally the `lock` operation takes a `--reason` argument
* #### Lock All/ Unlock All
  * Locks/Unlocks all workspaces in the specified org

  ```bash
    $ tfectl workspace lockall
    [
      {
        "id": "ws-SxWNNcYPkLD48ZC7",
        "locked": true,
        "name": "test-workspace-1"
      },
      {
        "id": "ws-LXkPCWnJKJ1FSgjs",
        "locked": true,
        "name": "uat-workspace"
      },
      {
        "id": "ws-E9o8VitHDAvCp3wj",
        "locked": true,
        "name": "uat-2-workspace"
      },
      {
        "id": "ws-5xUNCXVKrryoPcEp",
        "locked": true,
        "name": "dev-workspace"
      }
    ]
  ```
</details>

### Runs
<details>
    <summary>Run Operations</summary>

* `run` sub-command lets you manage runs against one or more workspaces
* #### List run
  * List runs in workspace specified by workspaceID
  * `--status` refers to valid [Run.Status](https://developer.hashicorp.com/terraform/enterprise/api-docs/run#run-states) attributes 
  * `--operation` refers to valid [Run.Operation](https://developer.hashicorp.com/terraform/enterprise/api-docs/run#run-operations) attributes

  ```bash
    $ tfectl run list --workspace-id ws-NMH66XMnUeF8duTx --status "policy_checked"
    [
        {
            "id": "run-zQFc5h2uPhEWW9Sr",
            "status": "policy_checked",
            "workspace_id": "ws-NMH66XMnUeF8duTx",
            "workspace_name": "tfc-infra-workspace"
        }
    ]
  ```

* #### Bulk Queue
  * Bulk queue plans against one or many workspaces
 
  ```bash
    $ tfectl run queue --filter workspace-sandbox
    [
      {
        "id": "run-pX9Lrq5KCrsgCYFH",
        "workspace_id": "ws-DpeRu7KpazXEWKoJ",
        "workspace_name": "workspace-sandbox",
        "status": "pending"
      }
    ]
  ```

* #### Apply runs
  * Apply pending plans - takes a comma-separated-string of runIDs

  ```bash
    $ tfectl run apply --ids run-UowKQd1cF7bgNfCp
    [
      {
        "id": "run-UowKQd1cF7bgNfCp",
        "workspace_id": "ws-N2qoyJxF1TkfeRYy",
        "workspace_name": "test-workspace-2",
        "status": "applying"
      }
    ]
  ```

* #### Query runs
  * Query/Get run-details from runIDs

  ```bash
    $ tfectl run get --ids run-UowKQd1cF7bgNfCp
    [
      {
        "id": "run-UowKQd1cF7bgNfCp",
        "workspace_id": "ws-N2qoyJxF1TkfeRYy",
        "workspace_name": "test-workspace-2",
        "status": "applied"
      }
    ]
  ```
</details>

### Variables
<details>
    <summary>Variable Operations</summary>

* CRUD operations on workspace variables
* #### Query/List workspace variables
  ```bash
    $ tfectl variable list --workspace-filter workspace-sandbox
    [
      {
        "workspace_id": "ws-DpeRu7KpazXEWKoJ",
        "workspace_name": "workspace-sandbox",
        "variables": [
          {
            "id": "var-RH7Q9pyD8gtgabtz",
            "key": "WORKSPACE_VAR_1",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": false
          },
          {
            "id": "var-wQutb5uQeSb4SwRn",
            "key": "workspace_tf_var",
            "value": "",
            "description": "",
            "category": "terraform",
            "hcl": false,
            "sensitive": true
          },
          {
            "id": "var-cSB5E11TRewuyfd9",
            "key": "WORKSPACE_VAR_2",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": false
          },
          {
            "id": "var-SP4Lcue83mCKVvHW",
            "key": "WORKSPACE_SECRET_VAR",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": true
          }
        ]
      }
    ]
  ```

* #### Create new workspace variable
  ```bash
    $ tfectl variable create --workspace-id ws-DpeRu7KpazXEWKoJ --description "test" --key "testCLI" --value "testCLI value" --sensitive true --type terraform --hcl
    {
      "id": "var-uCgZrzkPhis6qXTS",
      "key": "testCLI",
      "value": "",
      "description": "test",
      "category": "terraform",
      "hcl": true,
      "sensitive": true
    }
  ```
* #### Update existing workspace variable
  ```bash
    $ tfectl variable update --variable-id var-uCgZrzkPhis6qXTS --workspace-id ws-DpeRu7KpazXEWKoJ --value "test CLI Value 2" --key "testCLI" --hcl --sensitive true
    {
      "id": "var-uCgZrzkPhis6qXTS",
      "key": "testCLI",
      "value": "",
      "description": "Variable Updated by tfectl",
      "category": "terraform",
      "hcl": true,
      "sensitive": true
    }
  ```
* #### Delete existing workspace variable
  ```bash
    $ tfectl variable delete --variable-id var-uCgZrzkPhis6qXTS --workspace-id ws-DpeRu7KpazXEWKoJ
    # Returns current variables (similar to variable list)
    [
      {
        "workspace_id": "ws-DpeRu7KpazXEWKoJ",
        "workspace_name": "workspace-sandbox",
        "variables": [
          {
            "id": "var-RH7Q9pyD8gtgabtz",
            "key": "WORKSPACE_VAR_1",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": false
          },
          {
            "id": "var-wQutb5uQeSb4SwRn",
            "key": "workspace_tf_var",
            "value": "",
            "description": "",
            "category": "terraform",
            "hcl": false,
            "sensitive": true
          },
          {
            "id": "var-cSB5E11TRewuyfd9",
            "key": "WORKSPACE_VAR_2",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": false
          },
          {
            "id": "var-SP4Lcue83mCKVvHW",
            "key": "WORKSPACE_SECRET_VAR",
            "value": "",
            "description": "",
            "category": "env",
            "hcl": false,
            "sensitive": true
          }
        ]
      }
    ]
  ```

* #### Create variables from file
  ```bash
    $ tfectl variable create from-file --file variables.json --workspace-id ws-DpeRu7KpazXEWKoJ
    [
      {
        "id": "var-oDNV14eJf9ijjcc2",
        "key": "test1",
        "value": "value1",
        "description": "Test Variable 1",
        "category": "env",
        "hcl": false,
        "sensitive": false
      },
      {
        "id": "var-e1vFqg3ooToLi5xR",
        "key": "test2",
        "value": "",
        "description": "Test Variable 2 - sensitive",
        "category": "env",
        "hcl": false,
        "sensitive": true
      }
    ]
  ```
</details>

### Admin
<details>
    <summary>Admin Operations - TFE ONLY</summary>

* Perform Admin operations supported by the TFE Admin API.
* NOTE: Admin settings are only available in Terraform Enterprise.

* #### Runs
  * #### List - Lists Runs filtered on run status - querying the `admin/runs` endpoint
  ```bash
    $ tfectl admin run list --filter "plan_queued" --query '.[] | .id'
    [
        "run-4LuSKSss9KH2NAPN",
        "run-HCL7LVz67hVHEgsx",
        "run-ozEfahr1YrDQNokG",
        "run-hqWdU7BMuQPpqFrE",
        "run-BstJ5RJKFGmYnCni",
        "run-7WCMcDf8GZxYGqjN",
        "run-ZtzW7Xb5k6cfmgNK",
        "run-WC4q9Ec3vernx7Sc",
        "run-q9Lak8i1rzS5mXFU",
        "run-gbJyJAT89tzC2ziz",
        "run-MLMzcUuoSZL8Tz8C",
        "run-nbSKBf9CLRjPbj1q",
        "run-faqeyLU2VMBcHPJQ",
        "run-hB6RqJtY1SuGWsHF",
        "run-6HaUc4T31yZsENmC",
        "run-vPvYHNrjBCD6Y3ke",
        "run-ENdFcVpEp2AMLxNr",
        "run-kxgmgdReVzrVopVG"
    ]
  ```
  * #### Force-Cancel - Force cancels runIDs
  ```bash
    $ tfectl admin run force-cancel --ids run-UFaNv3rz5XnzPhCh
    [
        {
            "id": "run-UFaNv3rz5XnzPhCh",
            "workspace_id": "ws-ojAyfT3ar4oXt3eA",
            "workspace_name": "workspace-infrastructure-production",
            "status": "cancelling"
        }
    ]
  ```
</details>

### Policy
<details>
    <summary>Policy Operations</summary>

* Query policies in TFE/TFC

* #### List
  ```bash
    $ tfectl policy list --filter "production-tagging"
    [
      {
        "id": "pol-5Qgo4h2mp2z68u3N",
        "name": "production-tagging",
        "kind": "sentinel",
        "enforce": "hard-mandatory",
        "policy_set_count": 1
      }
    ]
  ```
</details>

### Tag
<details>
    <summary>Tag Operations</summary>

* Query Organization tag information in TFE/TFC

* #### List
  * The `--filter` flag takes a comma separated list of workspaceIds, and returns a list of all organization tags excluding the tags associated with these workspaces
  ```bash
    $ tfectl tag list --filter ws-ojAyfT3ar4oXt3eA
	[
		{
			"name": "tag:infrastructure",
			"id": "tag-kuyrvHJPWUNY6BCG",
			"instance_count": 1
		},
		{
			"name": "tag:application1",
			"id": "tag-X8oXEEMsNoU61D99",
			"instance_count": 2
		},
		{
			"name": "tag:application2",
			"id": "tag-49e9MLKrGFyLS9aT",
			"instance_count": 2
		}
	]
  ```
  * The `--search` flag returns details of the specified organization tag
  ```bash
	$ tfectl tag list --search "tag:infrastructure"
	[
		{
			"name": "tag:infrastructure",
			"id": "tag-kuyrvHJPWUNY6BCG",
			"instance_count": 1
		}
	]
  ```
</details>

### Policy Set
<details>
    <summary>Policy Set Operations</summary>
* Query policy sets in TFE/TFC

* #### 1. List
  * Lists all policy sets
  ```bash
    $ tfectl policy-set list
    [
        {
            "id": "polset-7586a2UeKeNgPD3s",
            "name": "dev-policy-set",
            "kind": "sentinel",
            "global": false,
            "workspaces": null,
            "workspace_count": 5,
            "workspace_exclusions": null,
            "projects": [
                "prj-LsSPiJnMYl7tSMZ"
            ],
            "project_count": 1,
            "policies": [
                "pol-B3pWfMyAzR2VtQI"
            ],
            "policy_count": 1
        },
        {
            "id": "polset-Q8zN9Q6TfMVs8mu",
            "name": "prod-policy-set",
            "kind": "sentinel",
            "global": false,
            "workspaces": null,
            "workspace_count": 10,
            "workspace_exclusions": null,
            "projects": [
                "prj-yOtqzR2msFUFCDx"
            ],
            "project_count": 1,
            "policies": [
                "pol-Lm0WgxPdwUm2zGE",
                "pol-crBeEEB5b8EZtaB"
            ],
            "policy_count": 2
        }
    ]
  ```
</details>

### Policy Check
<details>
    <summary>Policy Check Operations</summary>

* Examine the details of a policy check performed against a given RunID

* #### 1. Show
  * Generates the details of a policy check performed against a RunID
  ```bash
    $ tfectl policy-check show --run-id run-A8PuL0GnIeldng1
    {
        "id": "polchk-ndVuh5Y2abygp5fu",
        "result": {
            "advisory_failed": 2,
            "hard_failed": 0,
            "passed": 46,
            "result": true,
            "soft_failed": 0,
            "total_failed": 2,
            "sentinel": {
                "data": {
                    "policy-set-01": {
                        "error": null,
                        "policies": [
                            {
                                "error": null
                                # OUTPUT TRUNCATED
                            } # OUTPUT TRUNCATED
                        ] # OUTPUT TRUNCATED
                    } # OUTPUT TRUNCATED
                }
            }
        }
    }
  ```

  * To query only those checks which have failed
  ```bash
    $ tfectl policy-check show --run-id run-Wxk42edRCCLB5fMi --query '.result.sentinel.data | to_entries | .[].value.policies | .[] | select(.result|not) | .policy'
    [
        {
            "enforcement-level": "advisory",
            "name": "policy-set-01/deploy-to-approved-regions"
        },
        {
            "enforcement-level": "advisory",
            "name": "policy-set-02/iaas-allowed-vm-skus"
        }
    ]
  ```
</details>

### Registry Modules
<details>
    <summary>Private Registry Module Operations</summary>

* Query Private Modules in the Organization registry

* #### 1. List
  * List all available Modules in the Organization registry
  ```bash
    $ tfectl registry-module list --query '.[] | select(.provider == "azurerm")'
    [
        {
            "id": "mod-DHAq8Casdas32uC",
            "module_latest_version": "2.0.4",
            "name": "windows-instance",
            "namespace": "MyNamespace",
            "provider": "azurerm",
            "publishing_mechanism": "git_tag",
            "registry_name": "private",
            "status": "setup_complete",
            "test_config": true,
            "vcs_repo": "MyGHOrg/terraform-azurerm-windows-instance"
        }
    ]
  ```
</details>

### Registry Providers
<details>
    <summary>Private Provider Registry Operations</summary>

* Query Private Providers in Organization Registry

* #### 1. List
  * List all available Private Providers in the Organization Registry
  ```bash
    $ tfectl registry-provider list
    [
        {
            "id": "prov-5fws9JKkNQZDz2Gf",
            "name": "aws",
            "namespace": "MyTFCOrg",
            "registry_name": "private"
        },
        {
            "id": "prov-bGhLiwy6APQ9r4dZ",
            "name": "azure",
            "namespace": "MyTFCOrg",
            "registry_name": "private"
        }
    ]
  ```

* #### 2. Get
  * Get details of given Private provider
  ```bash
    $ tfectl registry-provider get --name aws
    {
        "id": "prov-5fws9JKkNQZDz2Gf",
        "name": "aws",
        "namespace": "MyTFCOrg",
        "registry_name": "private",
        "provider_latest_version": "5.32.2",
        "provider_platforms": [
            {
                "id": "provpltfrm-wCCMzzy91Rfdj6PW",
                "os": "linux",
                "arch": "amd64",
                "filename": "terraform-provider-awx_5.32.2_linux_amd64.zip"
            },
            {
                "id": "provpltfrm-c9jhJ2tmwEbbwuTV",
                "os": "windows",
                "arch": "amd64",
                "filename": "terraform-provider-awx_5.32.2_windows_amd64.zip"
            }
        ]
    }
  ```
</details>
    

### Build
GoReleaser is used to produce binaries for multiple platforms (Windows, Mac, Linux).

To build all binaries locally:

- Install GoReleaser https://goreleaser.com/install/
- Run the build target command:

```bash
  $ make build
```
- Binaries will be built and output to the `/dist` folder.

## Contributing
* see `CONTRIBUTING.md`
