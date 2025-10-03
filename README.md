# Internal Transfers System

A Go-based internal transfers application that facilitates financial transactions between accounts. This system provides HTTP endpoints for creating accounts, querying account balances, and processing transfers between accounts.

## Prerequisites


- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **Make**: Usually pre-installed on Unix systems

## Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Nauman-S/Internal-Transfers-System
   cd Internal-Transfers-System
   ```

## Running the Application

### Option 1: Run with Docker (Recommended)
```bash
# Start all services (database + application)
make run

# View application logs
make logs

# Stop all services
make stop
```

The application will be available at `http://localhost:8080`

### Option 2: Run locally
```bash
# Install Go dependencies
go mod download

# Start only database services
make db-up

# Run the application locally
make run-local
```

The application will be available at `http://localhost:8080`

## API Endpoints

### Account Creation
**POST** `/accounts`

Creates a new account with the specified ID and initial balance.

**Request Body:**
```json
{
  "account_id": 123,
  "initial_balance": "100.23344"
}
```

**Response:**
- **Success**: Empty response (200 OK)
- **Error**: Error message with appropriate HTTP status code

### Account Query
**GET** `/accounts/{account_id}`

Retrieves account information including current balance.

**Response:**
```json
{
  "account_id": 123,
  "balance": "100.23344"
}
```

### Transaction Submission
**POST** `/transactions`

Processes a transfer between two accounts.

**Request Body:**
```json
{
  "source_account_id": 123,
  "destination_account_id": 456,
  "amount": "100.12345"
}
```

**Response:**
```json
{
  "transaction_id": 1,
  "status": "COMPLETED",
  "source_balance": "0.00000000",
  "destination_balance": "100.12345",
  "amount": "100.12345",
  "created_at": "2025-01-03T10:30:00Z"
}
```

## Database Access

### pgAdmin (Web Interface)
- **URL**: http://localhost:8081
- **Email**: admin@example.com
- **Password**: password
- **Server**: PostgreSQL 18 (auto-configured)

### Command Line
```bash
# Connect to PostgreSQL
docker exec transfers_postgres psql -U postgres -d transfers_db

# List tables
\dt

# Query accounts
SELECT * FROM accounts;

# Query transactions
SELECT * FROM transactions;

# Exit
\q
```

### External Database Client
- **Host**: localhost
- **Port**: 5432
- **Database**: transfers_db
- **Username**: postgres
- **Password**: password

## Testing the API

### Create an Account
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 123, "initial_balance": "100.50"}'
```

### Get Account Balance
```bash
curl http://localhost:8080/accounts/123
```

### Create a Transfer
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 123, "destination_account_id": 456, "amount": "25.75"}'
```

## Assumptions

1. **Single Currency**: All accounts use the same currency
2. **No Authentication**: No authentication or authorization is required
3. **Account IDs**: Account IDs are positive integers
4. **Amounts**: All amounts are positive decimal values
5. **Precision**: Decimal amounts support up to 8 decimal places
6. **Atomic Operations**: All transactions are processed atomically
7. **No Pending State**: Transactions are either completed or failed (no pending status)

## Error Handling

The system handles various error scenarios:

- **Invalid Account ID**: 400 Bad Request
- **Account Not Found**: 404 Not Found
- **Account Already Exists**: 409 Conflict
- **Insufficient Funds**: 400 Bad Request
- **Same Account Transfer**: 400 Bad Request
- **Invalid Amount**: 400 Bad Request
- **System Errors**: 500 Internal Server Error

## Architecture

```
├── cmd/server/          # Application entry point
├── api/                 # HTTP routing and middleware
├── service/             # Business logic layer
│   ├── account/         # Account management
│   └── transactions/    # Transaction processing
├── storage/             # Data access layer
├── models/              # Domain models
├── codes/               # Error codes and messages
├── rest_handler/        # HTTP response handling
├── config/              # Configuration management
└── migrations/          # Database migrations
```

## Development

### Building
```bash
make build
```

### Running Tests
```bash
# Run integration tests (requires database)
make db-up
make test-integration
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

## Database Schema

The system uses PostgreSQL with two main tables:

### `accounts` Table
Stores account information and balances.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique account identifier |
| `balance` | DECIMAL(20,8) | Account balance with 8 decimal precision |
| `created_at` | TIMESTAMP WITH TIME ZONE | Account creation timestamp |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Last update timestamp |

### `transactions` Table
Audit trail for all transfer transactions.

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL PRIMARY KEY | Auto-incrementing transaction ID |
| `source_account_id` | INTEGER | Source account ID (FK to accounts.id) |
| `destination_account_id` | INTEGER | Destination account ID (FK to accounts.id) |
| `amount` | DECIMAL(20,8) | Transfer amount with 8 decimal precision |
| `created_at` | TIMESTAMP WITH TIME ZONE | Transaction timestamp |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Last update timestamp |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_NAME` | transfers_db | Database name |
| `DB_USER` | postgres | Database username |
| `DB_PASSWORD` | password | Database password |
| `DB_SSL_MODE` | disable | SSL mode for database connection |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.