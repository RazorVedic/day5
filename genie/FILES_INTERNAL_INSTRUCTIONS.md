## STEPS TO BE FOLLOWED IN ORDER TO GENERATE INTERNAL CODE

**ONLY GENERATE `internal` FILES. DO NOT GENERATE `proto` again. Also, do not generate `cmd` files**
 
1. Define Migration Scripts: Think of a logical DB schema and write migration scripts for database schema changes.
internal/<aggregate-name>/migrations/20240508180733_create_<model1>.go
internal/<aggregate-name>/migrations/20240508180733_create_<model2>.go

2. Implement model.go following the DB schema defined in the migration files.

3. Implement repo.go using pkg/store storage dependency

4. Implement core.go using the repo.go as dependency

5. All business logic should be part of the service.go. 

6. Follow `internal/complaint/service/service.go` to write the service.go for this service.

7. Implement service functions based on the defined protobufs and required functionality from given requirements. 


8. Implement validator for validating the requests using golang ozzo-validation pkgFollow `internal/complaint/validator/valdiator.go` 
   internal/<aggregate-name>/validator/validator.go
    internal/<aggregate-name>/validator/validator_test.go   

9. Implement the server logic for validating and handling requests. It basically use validator to perform request validations use service to perform the business tasks.
internal/<aggregate-name>/server.go
internal/<aggregate-name>/server_test.go