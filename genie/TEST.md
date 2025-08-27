# Test
## Instructions

1. Use the following libraries:
    - `github.com/smartystreets/goconvey/convey`
    - `github.com/stretchr/testify/suite`

2. Utilize `convey.Convey` to group and execute tests.

3. Write test functions:
    - Create an array of test cases with different scenarios.
    ```go
    tests := []struct {
        name         string
        req          *rpc.FunctionRequest
        expectedErr  error
    }
    ```
    - Generate various request scenarios for the function under test and their expected responses.
    - Include scenarios such as valid requests and invalid cases (e.g., missing fields, invalid lengths).

4. Iterate through each test case and execute assertions using `convey.So`.
5. Use the files `internal/complaint/validator/valdiator_test.go` and 
`internal/complaint/server/server_test.go` for creating tests.