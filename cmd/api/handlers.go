package main

import (
	"backend/internal/graph"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go Movies up and running",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
	}

	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusUnauthorized)
		return
	}

	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.Token)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("invalid token"), http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.Refresh))
			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application) MovieCatalog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieId, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
	}

	movie, err := app.DB.MovieByID(movieId)
	if err != nil {
		app.errorJSON(w, err)
	}

	_ = app.writeJSON(w, http.StatusOK, movie)
}

func (app *application) MovieForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieId, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
	}

	movie, genres, err := app.DB.MovieByIDForEdit(movieId)
	if err != nil {
		app.errorJSON(w, err)
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		Movie:  movie,
		Genres: genres,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.AllGenres()
	if err != nil {
		app.errorJSON(w, err)
	}

	_ = app.writeJSON(w, http.StatusOK, genres)
}

func (app *application) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie

	err := app.readJSON(w, r, &movie)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	movie.Image = ""
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	newID, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.UpdateMovieGenres(newID, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
	}

	response := JSONResponse{
		Error:   false,
		Message: "Movie successfully inserted",
	}

	app.writeJSON(w, http.StatusOK, response)
}

func (app *application) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var payload models.Movie
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	movie, err := app.DB.MovieByID(payload.ID)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MPAARating = payload.MPAARating
	movie.Runtime = payload.Runtime
	movie.UpdatedAt = time.Now()

	err = app.DB.UpdateMovie(*movie)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.UpdateMovieGenres(movie.ID, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	response := JSONResponse{
		Error:   false,
		Message: "Movie successfully updated",
	}

	app.writeJSON(w, http.StatusOK, response)
}

func (app *application) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	response := JSONResponse{
		Error:   false,
		Message: "Movie successfully deleted",
	}

	app.writeJSON(w, http.StatusOK, response)
}

func (app *application) AllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	movies, err := app.DB.AllMovies(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) MoviesGraphql(w http.ResponseWriter, r *http.Request) {
	movies, _ := app.DB.AllMovies()

	q, _ := io.ReadAll(r.Body)
	query := string(q)

	g := graph.New(movies)
	g.QueryString = query

	response, err := g.Query()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	j, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
