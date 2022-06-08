# MongoDB Atlas x Vault Database User initContainer

This initContainer is designed to run after Atlas database user credentials have been requested from Vault but before the application connects.

There is a delay between database credential request and rollout completion.  This container waits until the changes complete and then verifies the connection before exiting.

Not supported by MongoDB.
