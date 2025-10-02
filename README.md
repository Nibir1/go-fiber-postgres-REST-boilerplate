# 🚀 Go Fiber + PostgreSQL REST Boilerplate

A lightweight boilerplate for building RESTful APIs with **Golang** ([Fiber](https://github.com/gofiber/fiber)) and **PostgreSQL**.  
This project provides a clean, modular backend setup with **SQLC, Paseto/JWT authentication, Docker, and migrations**.  

It’s designed as a **starting point** for rapid prototyping, learning, or small projects — without heavy cloud deployment overhead.

---

## 📂 Project Structure

```
.
├── .github/workflows/test.yml    # GitHub Actions CI
├── api/                          # API layer (handlers, middleware, tests)
│   ├── account.go / account_test.go
│   ├── middleware.go / middleware_test.go
│   ├── server.go
│   ├── transfer.go / transfer_test.go
│   ├── user.go / user_test.go
│   ├── validator.go
│
├── db/                           # Database layer
│   ├── migration/                # Migration files
│   ├── mock/                     # Mocks for testing
│   ├── query/                    # Custom SQL queries
│   ├── sqlc/                     # Auto-generated code (sqlc)
│   └── Simple_Bank.sql           # Schema
│
├── token/                        # Authentication (Paseto & JWT)
│   ├── maker.go
│   ├── jwt_maker.go / jwt_maker_test.go
│   ├── paseto_maker.go / paseto_maker_test.go
│   ├── payload.go
│
├── util/                         # Utilities
│   ├── config.go
│   ├── currency.go
│   ├── password.go / password_test.go
│   ├── random.go
│   └── role.go
│
├── app.env                       # Environment variables
├── Makefile                      # Dev workflow automation
├── sqlc.yaml                     # SQLC config
├── main.go                       # Application entrypoint
├── go.mod / go.sum               # Dependencies
└── README.md
```

---

## 📦 Dependencies

- **[Fiber v2](https://github.com/gofiber/fiber)** – Web framework (fast & minimal).  
- **[SQLC](https://github.com/kyleconroy/sqlc)** – Generate type-safe Go from SQL.  
- **[Paseto](https://github.com/o1egl/paseto)** & **[JWT-Go](https://github.com/dgrijalva/jwt-go)** – Secure authentication.  
- **[Viper](https://github.com/spf13/viper)** – Config management.  
- **[Go-Playground Validator](https://github.com/go-playground/validator)** – Request validation.  
- **[Lib/pq](https://github.com/lib/pq)** – PostgreSQL driver.  
- **[Golang Mock](https://github.com/golang/mock)** – Mocks for tests.  
- **[Testify](https://github.com/stretchr/testify)** – Assertions in tests.  
- **[x/crypto](https://pkg.go.dev/golang.org/x/crypto)** – Secure password hashing.  

---

## ⚡ Getting Started

### 1. Clone the repo
```bash
git clone https://github.com/nibir1/go-fiber-postgres-REST-boilerplate.git
cd go-fiber-postgres-REST-boilerplate
```

### 2. Setup environment variables
Edit **`app.env`**:

```env
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
```

### 3. Run PostgreSQL with Docker
```bash
make postgres
make createdb
```

### 4. Run migrations
```bash
make migrateup
```

### 5. Generate SQLC code
```bash
make sqlc
```

### 6. Start the server
```bash
make server
```

---

## 🧪 Running Tests

Run all unit tests with coverage:
```bash
make test
```

Mocks are auto-generated with:
```bash
make mock
```

---

## 🔑 Authentication

- Auth is handled with **Paseto tokens** (safer alternative to JWT).  
- Middleware checks `Authorization: Bearer <token>` headers.  
- On login, users receive a valid access token.  

---

## 📡 Example API Usage ( Can Be Tested Using Postman )

### Create User
```bash
curl -X POST http://localhost:8080/users   -H "Content-Type: application/json"   -d '{"username":"nahasat","password":"secret123","full_name":"Nahasat Nibir","email":"nahasat@example.com"}'
```

### Login
```bash
curl -X POST http://localhost:8080/users/login   -H "Content-Type: application/json"   -d '{"username":"nahasat","password":"secret123"}'
```

### Create Account (Authorized)
```bash
curl -X POST http://localhost:8080/accounts   -H "Authorization: Bearer <ACCESS_TOKEN>"   -H "Content-Type: application/json"   -d '{"owner": "nahasat","currency": "USD"}'
```

### Transfer Between Accounts
```bash
curl -X POST http://localhost:8080/transfers   -H "Authorization: Bearer <ACCESS_TOKEN>"   -H "Content-Type: application/json"   -d '{"from_account_id":1,"to_account_id":2,"amount":100,"currency":"USD"}'
```

### List Accounts
```bash
curl -X POST http://localhost:8080/accounts?page_id=1&page_size=5   -H "Content-Type: application/json"
```

---

## 🛠 Development Workflow

| Command | Description |
|---------|-------------|
| `make postgres` | Run PostgreSQL in Docker |
| `make createdb` | Create database |
| `make dropdb` | Drop database |
| `make psql` | Open psql shell |
| `make migratenew` | Create new migration file |
| `make migrateup` | Apply migrations |
| `make migratedown` | Rollback migrations |
| `make sqlc` | Generate Go code from SQL |
| `make mock` | Generate mocks |
| `make test` | Run tests |
| `make server` | Run the app |

---

## 📜 License

MIT License © 2025 [Nahasat Nibir](https://github.com/nibir1)  
