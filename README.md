# Vault Plugin: Grafana Secrets Backend
[![Tests](https://github.com/Boostport/vault-plugin-secrets-grafana/actions/workflows/tests.yml/badge.svg)](https://github.com/Boostport/vault-plugin-secrets-grafana/actions/workflows/tests.yml)

This is a [HashiCorp Vault](https://www.github.com/hashicorp/vault) plugin that generates tokens for 
[Grafana Cloud](https://grafana.com/products/cloud/) and standalone [Grafana instances](https://grafana.com/grafana/).

## Download
Binary releases are available at https://github.com/Boostport/vault-plugin-secrets-grafana/releases.

## Verify Binaries
The checksum for the binaries are signed with cosign. To verify the binaries, download the following files (where
`${VERSION}` is the version of the release):
- `vault-plugin-secrets-grafana_${VERSION}_checksums.txt`
- `vault-plugin-secrets-grafana_${VERSION}_checksums.txt.pem`
- `vault-plugin-secrets-grafana_${VERSION}_checksums.txt.sig`

Then download the release binaries you need. Here, we just download the linux amd64 binary:
-  `vault-plugin-secrets-grafana_${VERSION}_linux_amd64`

Then run the following commands to verify the checksums and signature:
```sh
# Verify checksum signature
$ cosign verify-blob --signature vault-plugin-secrets-grafana_${VERSION}_checksums.txt.sig --certificate vault-plugin-secrets-grafana_${VERSION}_checksums.txt.pem vault-plugin-secrets-grafana_${VERSION}_checksums.txt --certificate-identity "https://github.com/Boostport/vault-plugin-secrets-grafana/.github/workflows/release.yml@refs/tags/v${VERSION}" --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

# Verify checksum with binaries
$ sha256sum -c vault-plugin-secrets-grafana_${VERSION}_checksums.txt
```

## Getting Started
### Grafana Cloud
1. Create an [Access Policy](https://grafana.com/docs/grafana-cloud/account-management/authentication-and-permissions/access-policies/)
   in Grafana Cloud with the following scopes: `accesspolicies:read`, `accesspolicies:write`, `accesspolicies:delete`, `stacks:read` and `stack-service-accounts:write`.
2. Generate a token for the Access Policy.
3. Configure the Grafana secrets backend:
```shell
vault write grafana/config type=cloud token=<token>
```
4. Create an Access Policy role:
```shell
vault write grafana/roles/my-access-policy-role type=cloud_access_policy region=us scopes="accesspolicies:read, accesspolicies:write" realms='[{"type": "org", "identifier": "<org_identifier>", "labelPolicies": []}]'
```
5. Generate a token for the Access Policy role:
```shell
vault read grafana/creds/my-access-policy-role
```
6. Create a Service Account role:
```shell
vault write grafana/roles/my-service-account-role type=grafana_service_account stack=mycompany role=Editor
```
7. Generate a token for the Service Account role:
```shell
vault read grafana/creds/my-service-account-role
```
### Grafana Instance
1. Create a Service Account in your Grafana instance with the `Admin`  basic role, or the following fixed roles: `Roles:Role writer`, `Service accounts:Service account writer`.
2. Generate a token for the Service Account.
3. Configure the Grafana secrets backend:
```shell
vault write grafana/config type=grafana token=<token> url=<instance_url>
```
4. Create a Service Account role:
```shell
vault write grafana/roles/my-service-account-role role=Editor
```
5. Generate a token for the Service Account role:
```shell
vault read grafana/creds/my-service-account-role
```

## Backend Configuration
### Grafana Cloud
#### Configuration Parameters
| Parameter | Description                                             | Required | Default               |
|-----------|---------------------------------------------------------|----------|-----------------------|
| `type`    | The Grafana installation type. Should be set to `cloud` | `yes`    | `none`                |
| `token`   | The Access Policy token.                                | `yes`    | `none`                |
| `url`     | The URL of the Grafana Cloud instance.                  | `no`     | `https://grafana.com` |

#### Required Scopes for Access Policy:
- `accesspolicies:read`
- `accesspolicies:write`
- `accesspolicies:delete`
- `stacks:read`
- `stack-service-accounts:write`

### Grafana Instance
#### Configuration Parameters
| Parameter | Description                                                                | Required | Default |
|-----------|----------------------------------------------------------------------------|----------|---------|
| `type`    | The Grafana installation type. Should be set to `grafana`                  | `yes`    | `none`  |
| `token`   | The Service Account token.                                                 | `yes`    | `none`  |
| `url`     | The URL of the Grafana instance, example: `https://myinstance.grafana.net` | `yes`    | `none`  |

#### Required Roles for Service Account:
- If using basic roles: `Admin`
- If using fixed roles:
  - `Roles:Role writer` 
  - `Service accounts:Service account writer`

## Role Configuration
### Grafana Cloud
For Grafana Cloud, roles can be created to generate either Access Policy tokens or Service Account tokens.
#### Access Policy Roles
| Parameter         | Description                                                                                                                                   | Required | Default | Example                                                                                                                        |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|----------|---------|--------------------------------------------------------------------------------------------------------------------------------|
| `type`            | The role type. Should be `cloud_access_policy`.                                                                                               | `yes`    | `none`  |                                                                                                                                |
| `region`          | The region the Grafana Cloud organization is in.                                                                                              | `yes`    | `none`  | `us`                                                                                                                           |
| `scopes`          | Comma separated list of scopes.                                                                                                               | `yes`    | `none`  | `accesspolicies:read, accesspolicies:wrte`                                                                                     |
| `realms`          | [JSON array string](https://grafana.com/docs/grafana-cloud/developer-resources/api-reference/cloud-api/#request-body) representing the realm. | `yes`    | `none`  | `[{"type": "org", "identifier": "123456", "labelPolicies": []}, {"type": "org", "identifier": "456789", "labelPolicies": []}]` |
| `allowed_subnets` | Comma separated list of allowed subnets.                                                                                                      | `no`     | `none`  | `192.168.0.10/32, 2001:db0:82a3:0:0:8a5e:370:1234/1238`                                                                        |

#### Service Account Roles
| Parameter    | Description                                                                                                                                                                                                                | Required | Default | Example                                                           |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------|-------------------------------------------------------------------|
| `type`       | The role type. Should be `grafana_service_account`.                                                                                                                                                                        | `yes`    | `none`  |                                                                   | 
| `stack`      | The stack slug for your Grafana Cloud instance                                                                                                                                                                             | `yes`    | `none`  | `mycompany`                                                       |
| `role`       | The basic role. Valid values are `Admin`, `Editor` or `Viewer`.                                                                                                                                                            | `no`     | `none`  | `Editor`                                                          |
| `rbac_roles` | Comma separated list of fixed or custom roles. Use the role's name, rather than it's id as the backend automatically looks up the id of each role and uses them. **Note**: use the name of the role, not the display name. | `no`     | `none`  | `fixed:roles:writer, fixed:alerting.rules:reader, my-custom-role` |

### Grafana Instance
For Grafana instances, roles can only generate Service Account tokens.

| Parameter    | Description                                                                                                                                                                                                                | Required | Default | Example                                                           |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------|-------------------------------------------------------------------|
| `role`       | The basic role. Valid values are `Admin`, `Editor` or `Viewer`.                                                                                                                                                            | `no`     | `none`  | `Editor`                                                          |
| `rbac_roles` | Comma separated list of fixed or custom roles. Use the role's name, rather than it's id as the backend automatically looks up the id of each role and uses them. **Note**: use the name of the role, not the display name. | `no`     | `none`  | `fixed:roles:writer, fixed:alerting.rules:reader, my-custom-role` |

## Troubleshooting
### Why do I get a 403 error when trying to generate a server account token for Grafana Cloud?

Grafana instances hosted on Grafana Cloud become dormant if they have not been used for a while. This results in service
account creation failing when the instance is dormant:
```shell
$ vault read grafana/creds/service-account-role
Error reading grafana/creds/service-account-role: Error making API request.

URL: GET http://127.0.0.1:8200/v1/grafana/creds/service-account-role
Code: 500. Errors:

* 1 error occurred:
  * error creating service account: error creating service account from cloud token: error response from server (403): {"code":"Forbidden","traceID":"4dXXXX","message":"operation not allowed"}
```
To resolve the issue, wake up the Grafana instance by logging into the Grafana Cloud control panel in the browser and
launching the Grafana instance.

## Developing
To run unit tests, run `go test -v ./...` from the root of the repository.

To run unit tests and acceptance tests, set the following environment variables and run `go test -v ./...` from the root of the repository:

| Environment Variable                | Description                                                                                                                                                                           | Example                                |
|-------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------|
| `VAULT_ACC`                         | Whether to run acceptance tests or not.                                                                                                                                               | `1`                                    |
| `TEST_GRAFANA_CLOUD_TOKEN`          | Grafana Cloud Access Policy token with the following scopes: `accesspolicies:read`, `accesspolicies:write`, `accesspolicies:delete`, `stacks:read` and `stack-service-accounts:write` | `glc_eyJvIjoiXXXXIjp7InIiOiJ1cyJ9fQ==` |
| `TEST_GRAFANA_CLOUD_STACK_SLUG`     | The Grafana Cloud instance stack slug.                                                                                                                                                | `mycompany`                            |
| `TEST_GRAFANA_CLOUD_REGION`         | The region the Grafana Cloud organization is in.                                                                                                                                      | `us`                                   |
| `TEST_GRAFANA_CLOUD_ORG_IDENTIFIER` | The id of the grafana cloud organization.                                                                                                                                             | `123456`                               |

**Notes**: 
- Running acceptance tests will create and delete Access Policies and Service Accounts in your Grafana Cloud account.
- To test the backend in Grafana instance mode, we create a test service account in the Grafana Cloud stack and use that to simulate a standalone Grafana instance.
