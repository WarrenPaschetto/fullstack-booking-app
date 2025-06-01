![alt dynamic badge for workflow tests](https://github.com/WarrenPaschetto/fullstack-booking-app/actions/workflows/backend.yml/badge.svg?branch=main)
![alt dynamic badge for workflow tests](https://github.com/WarrenPaschetto/fullstack-booking-app/actions/workflows/frontend.yml/badge.svg?branch=main)

[![codecov-backend](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app/branch/main/graph/badge.svg?flag=backend)](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app)



# Fullstack Booking App

A fullstack scheduling and booking application built with:

- **Go** for the backend API
- **Supabase (PostgreSQL-powered )** for the database
- **Next.js + Tailwind CSS** for the frontend
- **Monorepo** structure (frontend and backend together)
- **CI/CD** with GitHub Actions
- **Codecov** for coverage reports of testing

---


# The Backend


## 🛠️ Prerequisites

A fullstack scheduling and booking application built with:

- Go 1.18+ (https://golang.org/dl)
- PostgreSQL (for the psql CLI)
- Goose CLI (https://github.com/pressly/goose)
- A Supabase account (https://supabase.com)




## ⚙️ Setup Supabase

1. Log in to Supabase and create a new project.
2. In the dashboard, go to Settings → Database → Connection string.
3. Copy the Connection string (libpq) URL, for example:
```
postgresql://postgres:<PASSWORD>@db.<project>.supabase.co:5432/postgres?sslmode=require
```


## 🌐 Environment Variables

In the backend/ directory, create a file named .env with:

```
DATABASE_URL=postgresql://postgres:<PASSWORD>@db.<project>.supabase.co:5432/postgres?sslmode=require
PORT=8080
JWT_SECRET=<your_jwt_secret_here>  # optional, for JWT auth
```



## 🪿 Install Goose

Install the Goose CLI for managing migrations:

```
# via Go modules
go install github.com/pressly/goose/v3/cmd/goose@latest

# or on macOS using Homebrew
brew install goose
```
Verify installation:
```
goose --version
```



## 🗄️ Database Migrations

1. Go into the backend directory:
   ```
   cd backend
   ```
   
2. In your shell, export your DATABASE_URL:
   ```
   export DATABASE_URL=postgresql://postgres:<PASSWORD>@db.<project>.supabase.co:5432/postgres?sslmode=require
   ```
   
3. Apply migrations:
   ```
   goose -dir sql/schema postgres "$DATABASE_URL" up
   ```

4. Verify the created tables:
   ```
   psql "&DATABASE_URL" -c '/dt'
   ```



## 🏃‍♂️ Run the Application

1. Fetch dependencies and run the server:
   ```
   go mod tidy
   go run ./cmd/main.go
   ```

2. You should see:
   ```
   ✅ Connected to Supabase Postgres
   Listening on :8080
   ```



## 🧪 Testing Endpoints

Open a new terminal and use curl to exercise your handlers:
- **Register a new user**
  ```
  curl -i -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"first_name":"John","last_name":"Doe","email":"john@example.com","password":"s3cret"}'
  ```

- **Log in to get a JWT**
  ```
  curl -i -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"s3cret"}'
  ```

- **Create a booking**
  ```
  TOKEN=<your_jwt_token>
  curl -i -X POST http://localhost:8080/api/bookings/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"appointment_start":"2025-06-01T09:00:00Z","duration_minutes":60}'
  ```

- **Delete a booking**
  ```
  curl -i -X DELETE http://localhost:8080/api/bookings/{id of booking} \
  -H "Authorization: Bearer $TOKEN" 
  ```

- **List bookings for user**
  ```
  curl -i -X GET http://localhost:8080/api/bookings \
  -H "Authorization: Bearer $TOKEN" 
  ```

- **Get booking by its id**
  ```
  curl -i -X GET http://localhost:8080/api/bookings/{id of booking} \
  -H "Authorization: Bearer $TOKEN" 
  ```

- **Reschedule a booking**
  ```
  curl -i -X PUT http://localhost:8080/api/bookings/{id of booking} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"appointment_start":"2025-06-01T08:00:00Z","duration_minutes":30}'
  ```



---


## 📜 License

MIT License
