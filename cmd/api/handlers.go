package main

import (
	"backend/internal/graph"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var paylod = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go movies up and running",
		Version: "1.0.0",
	}

	_ = app.writeJson(w, http.StatusOK, paylod)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()

	if err != nil {
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}
	// validate user against database

	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}
	// check password
	valid, err := user.PassowrdMatches(requestPayload.Password)

	if err != nil || !valid {
		app.errorJson(w, errors.New("invalid Credentials"), http.StatusUnauthorized)
		return
	}
	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJson(w, err)
	}
	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	// fmt.Printf("Coockie %s", refreshCookie)
	http.SetCookie(w, refreshCookie)
	app.writeJson(w, http.StatusAccepted, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parse the token to get the client
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})

			if err != nil {
				app.errorJson(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user id from the token claims
			userID, err := strconv.Atoi(claims.Subject)
			fmt.Println("user id :", userID)
			if err != nil {
				fmt.Println("Here failed 1")
				app.errorJson(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserById(userID)
			fmt.Println("user id :", userID)

			if err != nil {
				// fmt.Println("not Refreshed")
				fmt.Println("Here failed 2")
				app.errorJson(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokens, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJson(w, err)
			}
			http.SetCookie(w, app.auth.GetRefreshCookie(tokens.RefreshToken))
			app.writeJson(w, http.StatusOK, tokens)

		}
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredTheRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application) MovieCatelog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()

	if err != nil {
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}

func (app *application) PublicGetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	movieID, err := strconv.Atoi(id)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	movie, err := app.DB.OneMovie(movieID)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movie)
}

func (app *application) AdminGetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	movieID, err := strconv.Atoi(id)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	movie, genres, err := app.DB.OneMovieForEdit(movieID)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		genres,
	}

	_ = app.writeJson(w, http.StatusOK, payload)

}

func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.AllGenres()

	if err != nil {
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, genres)
}

func (app *application) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie

	err := app.readJson(w, r, &movie)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	// try to get an image
	movie = app.getPoster(movie)

	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	newID, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	// now handle genres
	err = app.DB.UpdateMovieGenres(newID, movie.GenresArray)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJson(w, http.StatusAccepted, resp)

}

func (app *application) getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}
	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	req, err := http.NewRequest("GET", theUrl+"&query="+url.QueryEscape(movie.Title), nil)
	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return movie
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObj TheMovieDB

	json.Unmarshal(bodyBytes, &responseObj)

	if len(responseObj.Results) > 0 {
		movie.Image = responseObj.Results[0].PosterPath
	}

	return movie
}

func (app *application) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var payload models.Movie

	err := app.readJson(w, r, &payload)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	movie, err := app.DB.OneMovie(payload.ID)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MPAARating = payload.MPAARating
	movie.RunTime = payload.RunTime
	movie.UpdatedAt = time.Now()

	err = app.DB.UpdateMovie(*movie)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	err = app.DB.UpdateMovieGenres(movie.ID, payload.GenresArray)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJson(w, http.StatusAccepted, resp)
}

func (app *application) DeleteMovieById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		app.errorJson(w, err)
		return
	}

	err = app.DB.DeleteMovieById(id)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error:   false,
		Message: "movie deleted",
	}

	app.writeJson(w, http.StatusAccepted, resp)
}

func (app *application) GetMovieByGenre(w http.ResponseWriter, r *http.Request) {
	// all movies by genre

	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		app.errorJson(w, err)
		return
	}

	movies, err := app.DB.AllMovies(id)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusOK, movies)
}

func (app *application) moviesGraphQL(w http.ResponseWriter, r *http.Request) {
	// need to populate the our graph type with movies
	movies, _ := app.DB.AllMovies()

	// get the query from the request
	q, _ := io.ReadAll(r.Body)
	query := string(q)

	// create a new variable of tyupe *graph.Graph
	g := graph.New(movies)

	// set the query string on the variable
	g.QueryString = query

	// perform the query
	resp, err := g.Query()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	j, _ := json.MarshalIndent(resp, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
