# Payment Microservice  
## Event-Driven Backend Architecture

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Architecture](https://img.shields.io/badge/Event--Driven-Architecture-blue?style=for-the-badge)
![DDD](https://img.shields.io/badge/DDD-Domain%20Driven%20Design-green?style=for-the-badge)
![Observability](https://img.shields.io/badge/Observability-Structured%20Logs-orange?style=for-the-badge)

This project is a **Payment Processing API built in Go**, designed to simulate real-world backend challenges such as:

- Event-driven communication
- Retry with exponential backoff
- Idempotency guarantees
- Concurrency safety
- Structured logging & metrics
- Clean Architecture (DDD-oriented)

It was designed as a **production-grade backend case study**, not just a CRUD API.

---

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Architecture Overview](#architecture-overview)
- [Core Concepts](#core-concepts)
- [API Endpoints](#api-endpoints)
- [Testing Strategy](#testing-strategy)

---

## Installation

1. Clone the repository

```bash
git clone https://github.com/your-username/payment-microservice
```

2. Install dependencies

```bash
go mod tidy
```

3. Run the application

```bash
go run cmd/server/main.go
```

---

## Configuration

Currently the system runs **in-memory** for simplicity and testing.

No external dependencies are required.

Future roadmap includes:
- SQLite persistence
- Outbox pattern
- Prometheus metrics export
- Distributed tracing support

---

## Usage

1. Start the application:

```bash
go run cmd/server/main.go
```

2. The API will be accessible at:

```
http://localhost:8080
```

---

## Architecture Overview

The system follows a layered architecture:

```
HTTP Layer
   ↓
Application Layer (Use Cases)
   ↓
Domain Layer (Business Rules)
   ↓
Infrastructure Layer (Persistence, EventBus, Metrics, Logging)
```

### Key Architectural Principles

- Domain-first design
- Explicit business invariants
- Event-driven communication
- Idempotent processing
- Concurrency-safe operations
- Observability as a first-class concern

---

## Core Concepts

### Invoice

Represents a business obligation.

Possible states:

- `PENDING`
- `PROCESSING`
- `PAID`
- `FAILED`

---

### Payment

Represents a **technical attempt** to fulfill an invoice.

An invoice may have multiple payments due to:

- Retry after failure
- Ambiguous gateway error
- Manual retry
- Concurrent event processing

This models real-world payment systems.

---

### Event-Driven Flow

Main domain events:

- `PaymentRequested`
- `PaymentSucceeded`
- `PaymentFailed`

Payments are processed asynchronously via an in-memory EventBus.

Retry logic publishes new events instead of mutating state silently.

---

### Idempotency

The system guarantees:

- No duplicate payment creation
- Safe handling of re-delivered events
- Concurrency-safe processing

Implemented using:

- Idempotency keys
- Repository safeguards
- Deterministic processing logic

---

### Retry Strategy

- Exponential backoff
- Configurable max attempts
- Retry emitted as new domain event
- No implicit retry loops

---

### Observability

The system includes:

- Structured logs
- Explicit metrics interface
- Correlation-ready architecture
- Clear separation between business logic and monitoring

Metrics tracked:

- Payments processed
- Payments succeeded
- Payments failed

---

## API Endpoints

The API provides the following endpoints:

### Create Invoice

```markdown
POST /invoices
```

**Body**

```json
{
  "id": "inv-123",
  "amount": 1000
}
```

---

### Request Payment

```markdown
POST /invoices/{id}/pay
```

Triggers the asynchronous payment flow.

---

### List Invoices

```markdown
GET /invoices
```

Returns all invoices and their current status.

---

## Testing Strategy

The project includes tests covering:

- Successful payment flow
- Failure and retry flow
- Idempotency guarantees
- Concurrency scenarios
- Metrics validation

Tests validate behavior, not implementation details.

---

## Why This Project Matters

This project demonstrates understanding of:

- Distributed system realities
- Failure handling
- Concurrency issues
- Domain modeling
- Clean separation of concerns
- Production-grade backend thinking

It reflects backend engineering focused on reliability and correctness rather than surface-level functionality.
