While creating a new file, always use the root directory path: `github.com/razorpay/go-foundation-v2/` and then add the file name. Do not add any numbers or symbols in the file name other than `_` and chars.

1. `cmd/<aggregate-name>/main.go`

   - **Responsibility**: Entry point for the domain-specific service. Initializes the application, configures all necessary dependencies (publishers, consumers, servers, services, storage, etc.), and starts the server. Acts as a wiring unit for the domain to run.

2. `cmd/<aggregate-name>/config/config.go`

   - **Responsibility**: Defines the config struct for the service. Also exposes methods to load the config.

3. `cmd/<aggregate-name>-migration/main.go`

   - **Responsibility**: Entry point for the domain-specific migration service. Initializes the migration for the domain, runs migrations to create entities in the database, configures necessary dependencies, and starts the migration process.

4. `cmd/<aggregate-name>-migration/condfig/config.go`

   - **Responsibility**: Defines the config struct for the migration service. Also exposes methods to load the config.

5. `config/<aggregate-name>/default.toml`

   - **Responsibility**: Default configuration settings.

6. `config/<aggregate-name>/stage.toml`

   - **Responsibility**: Stage-specific configuration settings.

7. `config/<aggregate-name>-migration/default.toml`

   - **Responsibility**: Default configuration settings.

8. `config/<aggregate-name>-migration/stage.toml`

   - **Responsibility**: Stage-specific configuration settings.

9. `internal/<aggregate-name>/migrations/20240508180733_create_<model1>.go`

   - **Responsibility**: Migration script to create the first model.

10. `internal/<aggregate-name>/migrations/20240508180733_create_<model2>.go`

    - **Responsibility**: Migration script to create the second model.

11. `internal/<aggregate-name>/service/core/model/model.go`

    - **Responsibility**: Domain model definitions.

12. `internal/<aggregate-name>/service/core/repo/repo.go`

    - **Responsibility**: Repository interface and implementation to interact with storage.

13. `internal/<aggregate-name>/service/core/core.go`

    - **Responsibility**: Core service logic that uses models and repositories.

14. `internal/<aggregate-name>/service/service.go`

    - **Responsibility**: High-level service implementation that interacts with the core components.

15. `internal/<aggregate-name>/service/option.go`

    - **Responsibility**: Service options for creating service using options pattern.

16. `internal/<aggregate-name>/validator/validator.go`

    - **Responsibility**: Implements the validation logic for the domain.

17. `internal/<aggregate-name>/validator/validator_test.go`

    - **Responsibility**: Contains tests for the validation logic.

18. `internal/<aggregate-name>/server.go`

    - **Responsibility**: Implements the server logic for handling incoming requests. Validates the input request and then calls the service layer.

19. `internal/<aggregate-name>/server_test.go`

    - **Responsibility**: Contains tests for the server logic.

20. `internal/topics/<aggregate-name>.go`

    - **Responsibility**: Defines the topics and events that the domain will publish to or subscribe from.

21. `pkg/pubsub/pub/pub.go`

    - **Responsibility**: Implements the logic for publishing events to a message broker like Kafka.

22. `pkg/pubsub/sub/sub.go`

    - **Responsibility**: Implements the logic for subscribing to events from a message broker like Kafka.

23. `pkg/storage/storage.go`

    - **Responsibility**: Defines the storage interface and common storage utilities.

24. `pkg/storage/sql/gorm.go`

    - **Responsibility**: Implements SQL-based storage using GORM.

25. `proto/go-foundation-v2/v1/<aggregate-name>/<domain-name>.proto`

    - **Responsibility**: Protobuf definitions for the domain service and request responses.

26. `proto/go-foundation-v2/v1/<aggregate-name>/event/<event-name>.proto`
    - **Responsibility**: Protobuf definitions for domain events.
