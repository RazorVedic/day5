# Domain-Driven Design (DDD)

## What is DDD?

Domain-Driven Design (DDD) is an approach to software development that emphasizes the importance of the domain (the problem space) in the design process. It focuses on creating a shared understanding of the domain among stakeholders and aligning the software model with the business domain.

To effectively create DDD-driven microservices, consider the following principles and steps:

1. **Understand the Domain**: Gain a deep understanding of the business domain and its requirements.
2. **Establish Ubiquitous Language**: Develop a common language shared by both developers and domain experts. This language should be used consistently in discussions, documentation, and code to ensure clear and precise communication.
3. **Define Bounded Contexts**: Identify distinct areas within the domain where specific models apply. Each bounded context should have its own ubiquitous language and rules, allowing for modular and manageable systems.
4. **Focus on the Core Domain**: Prioritize the core domain and its logic. Identify the most critical aspects of the business and design models that address these core challenges and opportunities directly.
5. **Manage Complexity**: Use DDD to manage complexity by breaking down the domain into bounded contexts and creating clear models.
6. **Design Microservices**: Apply DDD principles to design microservices with clear boundaries and responsibilities. Each microservice should correspond to a specific bounded context, allowing for independent evolution and scalability.
7. **Implement DDD Patterns**: Use DDD patterns such as Aggregates, Entities, Value Objects, Repositories, and Factories to structure your code. These patterns help maintain the integrity of your domain model and support the principles of DDD.

## Building Blocks of DDD

### Domain
A domain is the sphere of knowledge or activity around which the application logic revolves. It includes all related concepts, rules, and behaviors.

**Example**: The banking domain encompasses core concepts such as accounts, transactions, customers, loans, and regulatory compliance. These elements collectively define how banking operations are structured and executed within the software system.

### Subdomain
A subdomain is a specialized part of the overall domain that focuses on specific business capabilities or knowledge areas.

**Example**: Within the banking domain, distinct subdomains might include retail banking (handling individual customer accounts and transactions) and investment banking (managing complex financial instruments and portfolios). Each subdomain has its own unique set of rules, models, and workflows tailored to its specific business focus.

### Bounded Context
A bounded context defines the explicit boundaries within which a particular model or language applies. It isolates the domain models, ensuring clarity and separation of concerns.

**Example**: In a banking application, the retail banking bounded context may define its own models and terminology for concepts like savings accounts, checking accounts, and overdraft protection. Meanwhile, the investment banking bounded context may have distinct models and language for concepts like securities, trades, and risk management strategies. Each bounded context ensures that models are consistent and relevant within their defined scope, avoiding ambiguity and conflict.

### Aggregates
Aggregates serve as clusters that group together entities and value objects, treating them as a single unit. An aggregate has a root entity, known as an aggregate root, through which all interactions with the aggregate occur. Aggregates maintain the consistency and integrity of related objects, ensuring that changes to the cluster remain coherent and adhere to the domain’s rules and invariants.

**Example**: In a banking application, a common aggregate is an Account. The Account aggregate includes entities such as AccountHolder and Transaction. Additionally, the Account aggregate may include value objects like Money. The Account entity serves as the aggregate root, through which all interactions with the account and its related entities and value objects are managed. Changes to the account balance or transaction history are coordinated through the Account aggregate, ensuring that all related operations maintain consistency and adhere to banking rules and regulations.

### Entities
Entities are objects that have a distinct identity and are defined primarily by their identity rather than their attributes.

**Example**: A Transaction entity in banking is identified uniquely by a transaction ID. It encapsulates attributes such as the transaction amount, date, type (e.g., deposit, withdrawal), and related account ID. Entities like Transaction are critical in tracking and managing individual financial operations within the banking system.

### Value Objects
Value Objects are objects within the domain characterized primarily by their attributes rather than a unique identity. Value objects are typically immutable, meaning their state cannot be altered once created. They can be freely replaced with another instance having the same attributes without affecting the overall state of the system.

**Example**: Money value object represents a specific amount in a specific currency. Once instantiated with a particular amount and currency (e.g., 100 INR), a Money object's state cannot change. If another Money object with the same amount and currency is created, it represents the same monetary value and can be used interchangeably without impacting the system's overall state. This immutability ensures consistency in financial calculations and operations across the banking domain.

### Domain Events
Domain Events are meaningful occurrences within the domain that have significance and trigger reactions or state changes in the system.

**Example**: A "TransactionCompleted" domain event in banking is raised when a transaction successfully updates the balance of an account. This event might trigger notifications to the customer, updates to transaction history, and adjustments to account summaries. Domain events capture important business milestones and facilitate communication between different parts of the banking system.

For more understanding refer to the following link: [Building Domain-Driven Microservices](https://medium.com/walmartglobaltech/building-domain-driven-microservices-af688aa1b1b8)

## Example Explaining DDD Architecture

The Switch Domain Design represents a sophisticated application architecture built on Domain-Driven Design (DDD) principles. It organizes the application into various bounded contexts, each encapsulating specific business logic and entities relevant to that particular domain. Below is a detailed explanation using DDD terminologies such as domains, bounded contexts, aggregates, entities, value objects, and domain events.

## Tenant
### Bounded Context: Tenant
**Aggregates and Entities**:
- **Tenant**: Aggregate root representing a tenant.

**Relationships and Interactions**:
- A **Tenant** has many **Customers**.

### Customer
#### Bounded Context: Customer
**Aggregates and Entities**:
- **Customer**: Aggregate root representing the customer.
    - **Customer**: Entity representing a customer.
    - **Merchant**: Entity representing a merchant.
    - **User**: Entity representing user information.
    - **Terminal**: Entity representing a terminal associated with a merchant.
- **Device**: Aggregate representing a customer's device.
    - **Device**: Entity representing a device.
    - **Device Token**: Entity representing a device token.
    - **Verification Token**: Entity representing a verification token.
- **Contact**: Aggregate representing contact information.
    - **Contact**: Entity representing a contact.
    - **User Contact**: Entity representing a user’s contact.

**Relationships and Interactions**:
- A **Customer** can have multiple **Devices**.
- A **Merchant** has many **Terminals**.
- A **Customer** can have multiple **Contacts**.
- A **Customer** can be a **Merchant** or a **User**.
- A **Customer** interacts with payment context to initiate or collect payments.
- A **Contact** interacts with **Fundsource** context.
- A **Tenant** interacts with **Customer** while onboarding **Merchants**.

### Payment Context
#### Bounded Context: Payment
**Aggregates and Entities**:
- **Payment**: Aggregate root representing a payment.
- **Refund**: Aggregate representing a refund transaction.
- **Payer**: Entity representing the entity initiating the payment.
- **Payee**: Entity representing the entity receiving the payment.
- **Amount**: Entity representing the amount involved in a payment.

**Value Objects**:
- **VPA**: Virtual Payment Address associated with a payment.
- **Fundsource**: Represents the source of funds (e.g., bank account).
- **Creds**: Credentials for the payment.

**Domain Events**: Payment initiation, payment completion, payment failure.

**Relationships and Interactions**:
- A **Payment** involves a **Payer** and multiple **Payees**.
- A **Payer** or **Payee** can use **VPA**, **Fundsource**, or **Creds**.
- A **Payment** has one **Amount**.

### FundSource Context
#### Bounded Context: FundSource
**Aggregates**:
- **Fundsource**: Aggregate root representing a source of funds.
- **UPI Number**: Aggregate unique identifier for the VPA.
- **VPA**: Aggregate Virtual Payment Address linked to a fund source.

**Value Objects**:
- **Balance**: Balance of a fund source.
- **Savings, Current, NRE**: Types of fund sources.

**Relationships and Interactions**:
- A **Fundsource** is of type **Savings, Current**, or **NRE**.
- A **Fundsource** can have one **Balance**.
- A **Fundsource** can be linked to a **VPA**.
- A **VPA** has a **UPI Number**.
- **Fundsource** interacts with **Fundsource Provider** context.

### FundSourceProvider Context
#### Bounded Context: FundSourceProvider
**Aggregates and Entities**:
- **Fundsource Provider**: Aggregate root representing banks or other financial institutions providing fund sources.

### Complaint Context
#### Bounded Context: Complaint
**Aggregates and Entities**:
- **Complaint**: Aggregate root representing a complaint lodged by a customer.

**Domain Events**: Complaint initiated, complaint resolved.

**Relationships and Interactions**:
- **Complaints** are related to payments, indicating an interaction with the **Payment Context**.