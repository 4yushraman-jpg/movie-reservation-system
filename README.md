# Movie Reservation System

A RESTful API for movie ticket booking built with Go. Features include user authentication, movie and theater management, seat selection with real-time locking, and booking management.

## Features

- **User Authentication**: JWT-based signup and login with role-based access (user/admin)
- **Movie Management**: CRUD operations for movies (admin only)
- **Theater & Screen Management**: Multi-theater support with multiple screens per theater
- **Seat Management**: Configurable seating layouts with row/seat numbering and seat types
- **Show Scheduling**: Schedule movies on specific screens with pricing
- **Seat Locking**: 5-minute seat lock during checkout to prevent double booking
- **Booking System**: Complete booking flow with confirmation and cancellation

## Tech Stack

- **Language**: Go 1.25+
- **Router**: [Chi](https://github.com/go-chi/chi)
- **Database**: PostgreSQL
- **Database Driver**: [pgx](https://github.com/jackc/pgx)
- **Authentication**: JWT ([golang-jwt](https://github.com/golang-jwt/jwt))
- **Password Hashing**: bcrypt

## Project Structure

```
movie-reservation-system/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── database/
│   │   ├── db.go             # Database connection
│   │   └── migrations/
│   │       └── 001_init.sql  # Database schema
│   ├── handlers/
│   │   ├── booking.go        # Booking endpoints
│   │   ├── movie.go          # Movie endpoints
│   │   ├── screen.go         # Screen endpoints
│   │   ├── seat.go           # Seat endpoints
│   │   ├── show.go           # Show endpoints
│   │   ├── theater.go        # Theater endpoints
│   │   └── user.go           # User endpoints
│   ├── middleware/
│   │   └── auth.go           # JWT authentication middleware
│   └── models/
│       ├── booking.go
│       ├── movie.go
│       ├── screen.go
│       ├── seat.go
│       ├── show.go
│       ├── theater.go
│       └── user.go
├── .env                      # Environment variables (not in repo)
├── .gitignore
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 14 or higher

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/4yushraman-jpg/movie-reservation-system.git
   cd movie-reservation-system
   ```

2. **Set up the database**
   ```bash
   createdb movie_reservation
   psql -d movie_reservation -f internal/database/migrations/001_init.sql
   ```

3. **Configure environment variables**
   
   Create a `.env` file in the project root:
   ```env
   DATABASE_URL=postgres://username:password@localhost:5432/movie_reservation?sslmode=disable
   JWT_SECRET=your-secret-key-here
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

The server will start on `http://localhost:8080`.

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users/signup` | Register a new user |
| POST | `/api/v1/users/login` | Login and get JWT token |
| GET | `/api/v1/movies` | List all movies |
| GET | `/api/v1/movies/{id}` | Get movie details |
| GET | `/api/v1/theaters` | List all theaters |
| GET | `/api/v1/theaters/{theater_id}/screens` | List screens in a theater |
| GET | `/api/v1/shows` | List all shows |
| GET | `/api/v1/shows/{show_id}/seats` | Get available seats for a show |

### Protected Endpoints (Require Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/bookings/lock` | Lock seats for checkout (5 min) |
| POST | `/api/v1/bookings/confirm` | Confirm booking and payment |
| GET | `/api/v1/bookings` | Get user's booking history |
| GET | `/api/v1/bookings/{id}` | Get booking details |
| POST | `/api/v1/bookings/{id}/cancel` | Cancel a booking |

### Admin Endpoints (Require Admin Role)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/admin/movies` | Create a new movie |
| PUT | `/api/v1/admin/movies/{id}` | Update a movie |
| DELETE | `/api/v1/admin/movies/{id}` | Delete a movie |
| POST | `/api/v1/admin/theaters` | Create a new theater |
| POST | `/api/v1/admin/theaters/{theater_id}/screens` | Add a screen to a theater |
| POST | `/api/v1/admin/screens/{screen_id}/seats` | Generate seats for a screen |
| POST | `/api/v1/admin/shows` | Schedule a new show |

## Usage Examples

### Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/users/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepassword"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepassword"}'
```

### Lock Seats for Booking

```bash
curl -X POST http://localhost:8080/api/v1/bookings/lock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"show_id": 1, "show_seat_ids": [1, 2, 3]}'
```

### Confirm Booking

```bash
curl -X POST http://localhost:8080/api/v1/bookings/confirm \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"booking_id": 1}'
```

## Database Schema

The system uses the following main tables:

- **users**: User accounts with role-based access
- **movies**: Film information (title, description, duration, release date)
- **theaters**: Cinema locations
- **screens**: Individual halls within theaters
- **seats**: Physical seats with row/number and type (standard/premium/VIP)
- **shows**: Movie screenings with timing and pricing
- **bookings**: User reservations with status tracking
- **show_seats**: Per-show seat availability with locking mechanism

## Booking Flow

1. User browses available shows and selects seats
2. System locks selected seats for 5 minutes
3. User completes payment (confirm booking)
4. Seats are marked as booked, or released if timer expires

## Tech stack

- **Router** — [Chi](https://github.com/go-chi/chi)
- **Database** — [pgx](https://github.com/jackc/pgx/v5)
- **JWT** — [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- **Env** — [godotenv](https://github.com/joho/godotenv)
- **Project URL** — [project URL](https://roadmap.sh/projects/movie-reservation-system)
