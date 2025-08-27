## STEPS TO BE FOLLOWED IN ORDER TO GENERATE CMD CODE

**ONLY GENERATE `cmd` FILES. DO NOT GENERATE `proto` and `internal` code again.**

1. Implement cmd files: Implement the main.go files for initialising the dependency services like 
pub, storage, repo, service, server and handle graceful shutdown pass these dependencies to the services required these.
   cmd/<aggregate-name>/main.go
   cmd/<aggregate-name>/config/config.go

2. Implement the migration cmd file to run it for migrating the database.
      cmd/<aggregate-name>-migration/main.go
      cmd/<aggregate-name>-migration/config/config.go


