# MongoDB Atlas x Vault Database User initContainer

This initContainer is designed to run after Atlas database user credentials have been requested from Vault but before the application connects.

There is a delay between database credential request and rollout completion.  This container waits until the changes complete and then verifies the connection before exiting.

### Examples

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

Not supported by MongoDB.
