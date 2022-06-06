# tf-cloud-opa-server

This example app acts as a webhook server that listens to Terraform Cloud [Run Tasks](https://www.terraform.io/cloud-docs/integrations/run-tasks) and performs OPA evaluation using [open-policy-agent Go API](https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api).

> *NOTE*: This is a demo-purpose app and is not intended to be used in any production environment.

## Custom Policies

This app assumes that you have one Rego module per TF workspace. The [`./rego`](./rego) folder contains your custom Rego modules and they are loaded automatically.

```go
workspace := strings.ReplaceAll(req.WorkspaceName, "-", "_")
q := rego.New(
    rego.Query("data.workspace."+workspace),
    rego.Load([]string{"./rego"}, nil),
    rego.Input(planJSON),
)
```

This way, you could have shared modules that other teams may manage, and you can selectively import which module to enforce for a given TF workspace.

## Development

```sh
$ git clone https://github.com/smaeda-ks/tf-cloud-opa-server tf-cloud-opa-server && cd tf-cloud-opa-server
$ go install

$ export PORT=8080
$ export TFC_RUN_TASK_HMAC_KEY=your_hmac_key
# This is optional. If you want to post debug output in the associated Run comment, set this env variable.
# see: https://github.com/smaeda-ks/tf-cloud-opa-server/blob/93fe849647dce99cc26b7e8d1fbfd538ec7ebb89/main.go#L127-L143
$ export TFC_API_KEY=your_tfc_api_key

$ go run main.go
```

For local development, recommend using [ngrok](https://ngrok.com/).

## Limitations

The [`Run Task Callback`](https://www.terraform.io/cloud-docs/api-docs/run-tasks-integration#run-task-callback) API can only accept a maximum of 512 characters for the `message` field. Therefore, you can't provide much information about your decision logs in response to the callback API.

A proper way of avoiding this limit is to generate a unique URL for each run and let users access there to view detailed logs. For a simple use, this requires far more effort.

Another workaround is to use the [`Comments API`](https://www.terraform.io/cloud-docs/api-docs/comments#create-comment) and post logs in the associated run's comment instead. But this requires a separate Terraform Cloud API token that is different from the sent `access_token` in the recieved payload since it doesn't have permission to create a comment.
