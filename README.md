![alt dynamic badge for workflow tests](https://github.com/WarrenPaschetto/fullstack-booking-app/actions/workflows/backend.yml/badge.svg?branch=main)
[![codecov-backend](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app/branch/main/graph/badge.svg?flag=backend)](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app)

![alt dynamic badge for workflow tests](https://github.com/WarrenPaschetto/fullstack-booking-app/actions/workflows/frontend.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app/branch/main/graph/badge.svg)](https://codecov.io/gh/WarrenPaschetto/fullstack-booking-app)


# Fullstack Booking App

A fullstack scheduling and booking application built with:

- **Go** for the backend API
- **Turso (SQLite-based)** for the database
- **Next.js + Tailwind CSS** for the frontend
- **Monorepo** structure (frontend and backend together)
- **CI/CD** with GitHub Actions

---

## 🧠 Project Goals

Build a fully functional, real-world booking platform with features like:

- User registration and login
- View available time slots
- Book, cancel, and reschedule appointments
- Admin dashboard to manage availability
- JWT-based authentication
- CI/CD pipeline for automated build, test, and deploy

---

## 📁 Project Structure

```
fullstack-booking-app/
├── backend/ # Go backend with Turso (SQLite)
│ ├── cmd/
│ ├── internal/
│ ├── go.mod
│ └── schema.sql
│
├── frontend/ # Next.js + Tailwind frontend
│ ├── pages/
│ ├── components/
│ ├── styles/
│ ├── tailwind.config.js
│ └── next.config.js
│
├── .github/workflows/ # GitHub Actions CI/CD
│ ├── backend.yml
│ └── frontend.yml
│
├── .gitignore
├── README.md
└── docker-compose.yml # (optional)
```

---

## 🛠️ Setup Steps

### 🔁 General

- [x] Create project folder: `fullstack-booking-app/`
- [x] Create `.gitignore` in root
- [x] Create GitHub repo and push initial commit

---

### 🧱 Backend (Go + Turso)

- [x] Create `backend/` folder
- [x] `cd backend && go mod init github.com/YOUR_USERNAME/fullstack-booking-app/backend`
- [x] Add Turso database and schema.sql
- [x] Create folder structure:
  - `/cmd/main.go`
  - `/internal/auth`
  - `/internal/db`
  - `/internal/handlers`
  - `/internal/models`
- [ ] Add JWT authentication
- [ ] Add booking logic and conflict detection
- [ ] Add admin-only routes

---

### 💅 Frontend (Next.js + Tailwind CSS)

- [x] Create `frontend/` folder
- [x] `cd frontend && npx create-next-app . --ts`
- [x] `npm install -D tailwindcss postcss autoprefixer`
- [x] `npx tailwindcss init -p`
- [ ] Configure `tailwind.config.js` and `globals.css`
- [ ] Build signup/login UI
- [ ] Build booking dashboard
- [ ] Add admin panel UI

---

### ⚙️ CI/CD

- [x] Add GitHub Actions workflow for backend (`backend.yml`)
- [x] Add GitHub Actions workflow for frontend (`frontend.yml`)
- [ ] Deploy backend (Railway, Fly.io, etc.)
- [ ] Deploy frontend (Vercel)

---

## 🚀 Future Enhancements

- Email or SMS reminders
- Google Calendar sync
- Recurring bookings
- Multi-provider support
- Analytics dashboard for admins

---

## 📜 License

MIT License
