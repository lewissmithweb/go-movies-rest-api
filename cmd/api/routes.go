package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Get("/", app.Home)
	mux.Get("/movies", app.AllMovies)
	mux.Get("/movies/{id}", app.GetMovie)
	mux.Get("/genres", app.AllGenres)
	mux.Get("/movies/genres/{id}", app.AllMoviesByGenre)

	mux.Post("/graph", app.MoviesGraphql)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequired)

		mux.Get("/movies", app.MovieCatalog)
		mux.Get("/movies/{id}", app.MovieForEdit)
		mux.Put("/movies", app.InsertMovie)
		mux.Patch("/movies/{id}", app.UpdateMovie)
		mux.Delete("/movies/{id}", app.DeleteMovie)
	})

	return mux
}
