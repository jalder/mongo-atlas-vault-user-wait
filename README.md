# MongoDB Atlas x Vault Database User initContainer

This initContainer is designed to run after Atlas database user credentials have been requested from Vault but before the application connects.

There is a delay between database credential request and rollout completion.  This container waits until the changes complete and then verifies the connection before exiting.

## Examples

### Environment Variables and Bash

The following vault injector annotation examples contain the `agent-init-first` and `agent-cache-enable` options.  The `agent-init-first` is required to place the vault injector initContainer **before** the wait and validate initContainer as this project executes.  The `agent-cache-enable` avoids firing the injector credential request for each initContainer and container, generating only one set of database credentials for the pod.

We then gather the required environment variables to communicate with the Atlas API from Vault as well.

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
    vault.hashicorp.com/agent-cache-enable: "true"
```

In examples/bash-style.yaml, you will find a proof of concept implementation.  This is not intended for production usage.
Reasons not to use in prod:
- the image is excessive in size
- mediocre security
- it can end up in an infinite loop
- it's ridiculous 


### Config File and Application (better)

The next example (`examples/image-style.yaml`) uses a json config file and a custom image (`main.go`).  It is a better solution.

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
    vault.hashicorp.com/agent-cache-enable: "true"
```

## Build

```
docker build -t atlas-wait-ready .
```

## Deploy

For anything other than PoC or sandbox, I recommend building and maintaining this image in an in-house registry.

Push to your registry of choice and update the initContainer's image path in the example patch snippet below.

```
  initContainers:
    - image: jalder/atlas-wait-ready:0.0.1
      name: wait-db-user-ready
      command: ["/atlas-wait-ready"]
      args: 
       - -apiKeyFile
       - "/vault/secrets/atlasapi" 
       - -uriFile
       - "/vault/secrets/atlas"
```

Successful runs should contain logs similar to:
```
$ kubectl logs -n mongodb devwebapp -c wait-db-user-ready
Starting MongoDB Atlas x Vault DB User Liveness Check
Checking Status of Cluster User Changes...
Atlas reports changeStatus: PENDING
Sleeping...
Checking Status of Cluster User Changes...
Atlas reports changeStatus: APPLIED
Confirming Vault Credentials and Atlas Access are Valid...
MongoDB Atlas Authentication Succeeded and Primary Pinged.
Exiting
```

The above test implies Atlas took ~10 seconds to apply and confirm the database user is ready for use.  The example image build is about 15MB in size (compressed).

## Disclaimer

Not supported by anyone.