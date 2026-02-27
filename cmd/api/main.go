package main

import (
	"context"
	"errors"
	"log"
	"movie-reservation-system/internal/database"
	"movie-reservation-system/internal/handlers"
	"movie-reservation-system/internal/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Connecting to the database...")
	dbPool, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established.")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	userHandler := handlers.UserHandler{
		DB:        dbPool,
		JWTSecret: []byte(jwtSecret),
	}

	movieHandler := handlers.MovieHandler{
		DB: dbPool,
	}

	theaterHandler := handlers.TheaterHandler{
		DB: dbPool,
	}

	screenHandler := handlers.ScreenHandler{
		DB: dbPool,
	}

	seatHandler := handlers.SeatHandler{
		DB: dbPool,
	}

	showHandler := handlers.ShowHandler{
		DB: dbPool,
	}

	bookingHandler := handlers.BookingHandler{
		DB: dbPool,
	}

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users/signup", userHandler.SignupHandler)
		r.Post("/users/login", userHandler.LoginHandler)

		r.Get("/movies", movieHandler.GetMoviesHandler)
		r.Get("/movies/{id}", movieHandler.GetMovieByIDHandler)

		r.Get("/theaters", theaterHandler.GetTheatersHandler)

		r.Get("/theaters/{theater_id}/screens", screenHandler.GetScreensHandler)

		r.Get("/shows", showHandler.GetShowsHandler)
		r.Get("/shows/{show_id}/seats", showHandler.GetShowSeatsHandler)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware([]byte(jwtSecret)))

			r.Post("/bookings/lock", bookingHandler.LockSeatsHandler)
			r.Post("/bookings/confirm", bookingHandler.ConfirmBookingHandler)
			r.Get("/bookings", bookingHandler.GetBookingsHandler)
			r.Get("/bookings/{id}", bookingHandler.GetBookingByIDHandler)
			r.Post("/bookings/{id}/cancel", bookingHandler.CancelBookingHandler)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminOnlyMiddleware)

				r.Post("/admin/movies", movieHandler.PostMovieHandler)
				r.Put("/admin/movies/{id}", movieHandler.PutMovieHandler)
				r.Delete("/admin/movies/{id}", movieHandler.DeleteMovieHandler)

				r.Post("/admin/theaters", theaterHandler.PostTheaterHandler)

				r.Post("/admin/theaters/{theater_id}/screens", screenHandler.PostScreenHandler)

				r.Post("/admin/screens/{screen_id}/seats", seatHandler.GenerateSeatsHandler)

				r.Post("/admin/shows", showHandler.PostShowHandler)
			})
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully")
}
