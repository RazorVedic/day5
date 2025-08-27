## Linters Configuration

**MUST FOLLOW THESE LINT RULES**

The following lint rules must be followed in the project:

- Wrap the line if it exceeds 100 characters.
- Ensures all files are ASCII encoded.
- Errors must not be unchecked in the code.
- There must not be unnecessary type assertions.
- There must not be assignments that are ineffective.
- There must not be unused variables, functions, and imports.
- Group standard library imports, external imports, and project-specific imports separately.
  - Ensure import paths are ordered alphabetically within their groups.
  - Example:
    ```go
    import (
        "fmt"
        "os"

        "github.com/pkg/errors"
    )
    ```

- Ensure comments are properly formatted.
  - Comments for exported functions, types, and variables must start with the name of the element being described.
  - Example:
    ```go
    // Add adds two integers and returns the result.
    func Add(x, y int) int {
        return x + y
    }
    ```

- There must not be trailing whitespace from lines.
- Ensure no extra blank lines at the end of a file.

- Enforce that each import path is declared on a separate line.
- Example:
  ```go
  import (
      "fmt"
      "os"
  )
  ```

- There must not be redundant aliases where the alias name is the same as the package name.
- Example:
  ```go
  import (
      "encoding/json"  // instead of json "encoding/json"
  )
  ```

- Ensure that no import path is repeated within the import block.