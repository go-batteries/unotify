
### Requirements:

- go v1.21+
- goose
- protoc
- protoc-gen-go
- protobuf
- postgres(optional) or supabase url
- terraform (for deployment)

**External Depenencies**:

- aws account
- atlassian api key [from here](https://id.atlassian.com/manage-profile/security/api-tokens)


### Deployment:

```shell
AWS_ACCOUNT="" ./infraa/build/ecs.sh
```

```shell
cd infraa/deploy && AWS_ACCOUNT="" make run
```


### Considerations

- Storage layer is redis for now
- This helps to have a publish subscribe with redis
- But some parts of the information needs to be stored in persistent storage,
  like hook provider configs
- The endpoint url for different projects can be like,
  `/webhook/provider/:project/payload`, that way for github, the base
  functionality can remain same.
- You could also provide different paths, but then you have to add them to
  router. in `cmd/server/main.go`

Each `Hook` has to implement a `Validate` method. Wether to validate or not,
depends on the provider.

**Example**:

Github allows setting a secret key when the webhook is created from their UI.
(For this requirement we only need `Releases`). So github uses this secret to
generate a `X-Hub-Signature-256`, which is sent in header. 

This value can be validated, by calculating the request body's signature, as
[mentioned by github](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries#examples).


In most cases, you will have _a single webhook endpoint per repository, served
by a single endpoint_ . Reason being it gets harder/tricker to validate secrets
if a single endpoint had multiple hooks, encrypted with different keys. We don't
want that sort of complexity.

