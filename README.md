# StoreFlow API

Production-oriented commerce backend built in Go using GraphQL, PostgreSQL and Redis.

StoreFlow demonstrates backend engineering patterns used in SaaS systems: transactional order processing, JWT authentication, Redis-backed caching with invalidation, store-level authorization, Docker-based development, and clean service boundaries.

## Highlights

- GraphQL API with gqlgen
- PostgreSQL schema for users, stores, products, orders and order items
- Transactional order creation with inventory decrement
- Redis product listing cache with TTL and invalidation
- JWT authentication and store ownership checks
- Docker Compose development environment
- GitHub Actions CI for tests, formatting, migrations and build
- Graceful HTTP server shutdown
## Tech Stack

- Go
- GraphQL with gqlgen
- PostgreSQL
- Redis
- Docker Compose
- JWT authentication
- bcrypt password hashing
- PostgreSQL transactions
- Clean modular architecture

## Architecture

The project follows a clean service-oriented structure:

```txt
GraphQL Resolver
  -> Service Layer
    -> Repository Interface
      -> PostgreSQL / Redis

Main folders:

cmd/api                 Application entrypoint
internal/auth           JWT, password hashing, auth middleware
internal/user           User domain
internal/store          Store domain
internal/product        Product domain + Redis product cache
internal/order          Orders, order items, transactional checkout logic
internal/graphql        gqlgen schema, resolvers, generated GraphQL code
internal/db             PostgreSQL connection
internal/cache          Redis connection
migrations              SQL migrations
Current Features
Users
Register user
Login user
JWT generation
Authenticated me query
Password hashing with bcrypt
Stores
Create store
List stores owned by authenticated user
Store ownership checks
Products
Create products
Public product listing by store
Product price stored as integer cents
Product inventory tracking
Redis cache for product listings
Cache TTL
Cache invalidation on product creation
Orders
Create order
Order items
Order status tracking:
PENDING
PAID
SHIPPED
CANCELLED
Transactional order creation
Inventory decrement inside transaction
Historical unit price copied into order item
Owner-only order listing
Owner-only order status update
Redis product cache invalidated after order creation
Why This Project Matters

This project demonstrates backend concepts expected in production SaaS systems:

Clean architecture boundaries
No database logic inside GraphQL resolvers
Context-aware database and cache operations
Proper money handling with integer cents
JWT authentication
Store-level authorization checks
PostgreSQL transactions for multi-step business operations
Redis used for a real performance case, not decorative caching
Cache invalidation after writes
SQL migrations
Meaningful unit tests
Requirements
Go
Docker Desktop
Docker Compose
Local Setup

Start PostgreSQL and Redis:

docker compose up -d postgres redis

Run database migrations:

Get-ChildItem .\migrations\*.up.sql | Sort-Object Name | ForEach-Object {
  Write-Host "Running migration:" $_.Name
  Get-Content $_.FullName | docker exec -i storeflow-postgres psql -v ON_ERROR_STOP=1 -U storeflow -d storeflow
}

Run the API:

go run .\cmd\api\main.go

Health check:

curl.exe http://localhost:8080/health

Expected response:

OK

GraphQL endpoint:

http://localhost:8080/graphql

GraphQL Playground:

http://localhost:8080/playground
Example GraphQL Requests
Register
mutation Register($input: RegisterInput!) {
  register(input: $input) {
    token
    user {
      id
      email
      name
      createdAt
      updatedAt
    }
  }
}

Variables:

{
  "input": {
    "email": "jan@example.com",
    "password": "Password123!",
    "name": "Jan Test"
  }
}
Create Store

Requires:

Authorization: Bearer <token>
mutation CreateStore($input: CreateStoreInput!) {
  createStore(input: $input) {
    id
    ownerId
    name
    slug
    createdAt
    updatedAt
  }
}

Variables:

{
  "input": {
    "name": "Test Store",
    "slug": "test-store"
  }
}
Create Product

Requires:

Authorization: Bearer <token>
mutation CreateProduct($input: CreateProductInput!) {
  createProduct(input: $input) {
    id
    storeId
    name
    priceCents
    currency
    inventoryCount
    active
  }
}

Variables:

{
  "input": {
    "storeId": "STORE_ID",
    "name": "Test Cap",
    "description": "Black cotton cap",
    "priceCents": 3900,
    "currency": "USD",
    "inventoryCount": 50
  }
}
Create Order

Public checkout-style operation.

mutation CreateOrder($input: CreateOrderInput!) {
  createOrder(input: $input) {
    id
    storeId
    customerEmail
    status
    totalCents
    currency
    items {
      id
      productId
      quantity
      unitPriceCents
      totalCents
      product {
        id
        name
        inventoryCount
      }
    }
    createdAt
    updatedAt
  }
}

Variables:

{
  "input": {
    "storeId": "STORE_ID",
    "customerEmail": "customer@example.com",
    "items": [
      {
        "productId": "PRODUCT_ID",
        "quantity": 2
      }
    ]
  }
}
Update Order Status

Requires:

Authorization: Bearer <token>
mutation UpdateOrderStatus($id: ID!, $status: OrderStatus!) {
  updateOrderStatus(id: $id, status: $status) {
    id
    status
    updatedAt
  }
}

Variables:

{
  "id": "ORDER_ID",
  "status": "PAID"
}
Redis Cache

Product listings are cached with this key format:

store:{storeId}:products:active:{true|false}

Example:

store:789adbc2-b8f5-4380-9095-2727017410c5:products:active:true

The cache uses a 60-second TTL.

Cache is invalidated when:

a product is created
an order is created and product inventory changes
Tests

Run all tests:

go test ./...

Current test coverage includes:

password hashing
password verification
JWT generation and validation
product cache key behavior
order status validation
order service validation for invalid inputs
Current Limitations

Not implemented yet:

payments
Stripe integration
refunds
shipping
tax calculation
pagination
dataloaders
role-based access control
production deployment pipeline
Next Planned Improvements
Add pagination for products and orders
Add integration tests with PostgreSQL
Add dataloaders for GraphQL nested product resolution
Add structured logging
Add request IDs
Add rate limiting
Add CI pipeline with GitHub Actions
