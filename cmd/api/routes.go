package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create a router mux
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Get("/", app.Home)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	mux.Get("/movies", app.AllMovies)
	mux.Get("/movies/{id}", app.PublicGetMovie)
	mux.Get("/genres", app.AllGenres)
	mux.Get("/movies/genres/{id}", app.GetMovieByGenre)

	mux.Post("/graph", app.moviesGraphQL)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequried)

		mux.Get("/movies", app.MovieCatelog)
		mux.Get("/movies/{id}", app.AdminGetMovie)
		mux.Put("/movies/0", app.CreateMovie)
		mux.Patch("/movies/{id}", app.UpdateMovie)
		mux.Delete("/movies/{id}", app.DeleteMovieById)
	})
	return mux
}
