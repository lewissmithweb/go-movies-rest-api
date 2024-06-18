package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = 8080

type application struct {
	DSN          string
	Domain       string
	DB           repository.DatabaseRepo
	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

func main() {
	// config
	var app application

	// read command line
	flag.StringVar(
		&app.DSN,
		"dsn",
		"host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5",
		"Postgres connection string",
	)
	flag.StringVar(&app.JWTSecret, "jwt-secret", "top-secret", "JWT secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", "example.com", "JWT issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", "example.com", "Audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie Domain")
	flag.StringVar(&app.Domain, "domain", "example.com", "Domain")
	flag.Parse()

	// db connect
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	defer app.DB.Connection().Close()

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "__Host-refresh_token",
		CookieDomain:  app.CookieDomain,
	}

	// start webserver
	log.Println("Starting application on port:", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
