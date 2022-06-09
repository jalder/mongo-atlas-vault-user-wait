# MongoDB Atlas x Vault Database User initContainer

This initContainer is designed to run after Atlas database user credentials have been requested from Vault but before the application connects.

There is a delay between database credential request and rollout completion.  This container waits until the changes complete and then verifies the connection before exiting.

## Examples

### Environment Variables

Considering the following vault injector annotation examples, note the `agent-init-first` option.  This is required to place the vault injector initContainer **before** the wait and validate initContainer in this project executes.

We proceed to gather the required environment variables to communicate with the Atlas API from Vault as well.

```
metadata:
  annotations:
    vault.hashicorp.com/agent-inject: 'true'
    vault.hashicorp.com/role: 'devweb-app'
    vault.hashicorp.com/agent-inject-secret-atlas: 'database/creds/my-role'
    vault.hashicorp.com/agent-inject-template-atlas: |
      {{ with secret "database/creds/my-role" -}}
        export MONGODB_URI="mongodb+srv://{{ .Data.username }}:{{ .Data.password }}@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority"
      {{- end }}    
    vault.hashicorp.com/agent-inject-secret-atlasapi: 'kv/atlas/api-readonly'
    vault.hashicorp.com/agent-inject-template-atlasapi: |
      {{ with secret "kv/atlas/api-readonly" -}}
        export MONGODB_ATLAS_PUBLIC_API_KEY="{{ .Data.MONGODB_ATLAS_PUBLIC_API_KEY }}"
        export MONGODB_ATLAS_PRIVATE_API_KEY="{{ .Data.MONGODB_ATLAS_PRIVATE_API_KEY }}"
        export MONGODB_ATLAS_PROJECT_ID="{{ .Data.MONGODB_ATLAS_PROJECT_ID }}"
        export MONGODB_ATLAS_CLUSTER_NAME="{{ .Data.MONGODB_ATLAS_CLUSTER_NAME }}"
      {{- end }}
    vault.hashicorp.com/agent-init-first: 'true'
```

In examples/bash-style.yaml, you will find a proof of concept implementation.  This is not intended for production usage.
Reasons not to use in prod:
- the image is excessive in size
- mediocre security
- it can end up in an infinite loop
- it's ridiculous 

### Config File

```
metadata:
  annotations:
    vault.hashicorp.com/agent-inject: 'true'
    vault.hashicorp.com/role: 'devweb-app'
    vault.hashicorp.com/agent-inject-secret-atlas: 'database/creds/my-role'
    vault.hashicorp.com/agent-inject-template-atlas: |
      {{ with secret "database/creds/my-role" -}}
        mongodb+srv://{{ .Data.username }}:{{ .Data.password }}@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
      {{- end }}    
    vault.hashicorp.com/agent-inject-secret-atlasapi: 'kv/atlas/api-readonly'
    vault.hashicorp.com/agent-inject-template-atlasapi: |
      {{ with secret "kv/atlas/api-readonly" -}}
        {
          "publicKey":"{{ .Data.data.MONGODB_ATLAS_PUBLIC_API_KEY }}",
          "privateKey":"{{ .Data.data.MONGODB_ATLAS_PRIVATE_API_KEY }}",
          "projectId":"{{ .Data.data.MONGODB_ATLAS_PROJECT_ID }}",
          "clusterName":"{{ .Data.data.MONGODB_ATLAS_CLUSTER_NAME }}"
        }
      {{- end }}
    vault.hashicorp.com/agent-init-first: 'true'
```

The example main.go app uses this approach.


## Build

```
docker build -t atlas-wait-ready
```

## Disclaimer

Not supported by MongoDB.