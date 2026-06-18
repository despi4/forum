package main

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/handler"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/middleware"
	authsvc "01.tomorrow-school.ai/git/amadiuly/forum/internal/service/auth"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/repository/sqlite"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/db"
	"01.tomorrow-school.ai/git/amadiuly/forum/utils"
)

const pattern = "templates/*.html"

func main() {
	if err := utils.Loadenv(".env"); err != nil {
		log.Fatal(err)
	}

	dsn := os.Getenv("DB_DSN")
	port := os.Getenv("PORT")

	tmpl := template.Must(template.ParseGlob(pattern))

	conn, err := db.NewConnDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("executing migrations in db...")

	if err := utils.RunMigrator(conn.GetDB(), "./migrations"); err != nil {
		log.Fatal(err)
	}

	userRepo := sqlite.NewUserRepo(conn)
	sessionRepo := sqlite.NewSessionRepo(conn)

	authSvc := authsvc.NewAuthService(sessionRepo, userRepo)

	authHandler := handler.NewAuthHandler(authSvc, tmpl)
	handler := handler.NewHandler(tmpl)

	router := http.NewServeMux()

	router.Handle("GET /auth/register", middleware.GuestMiddleware(authSvc, http.HandlerFunc(authHandler.RegisterPage)))
	router.HandleFunc("POST /auth/register", authHandler.Register)
	router.Handle("GET /auth/login", middleware.GuestMiddleware(authSvc, http.HandlerFunc(authHandler.LoginPage)))
	router.HandleFunc("POST /auth/login", authHandler.Login)
	router.Handle("GET /home", middleware.AuthMiddleware(authSvc, http.HandlerFunc(handler.HomePage), true))
	router.Handle("POST /auth/logout", middleware.AuthMiddleware(authSvc, http.HandlerFunc(authHandler.Logout), false))
	router.Handle("GET /profile", middleware.AuthMiddleware(authSvc, http.HandlerFunc(handler.ProfilePage), false))
	router.Handle("PUT /profile/change-password", middleware.AuthMiddleware(authSvc, http.HandlerFunc(authHandler.ChangePassword), false))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	mux := middleware.Logger(logger, router)

	log.Printf("Server started on %s\n", port)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
