package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sukrati192/greenlight/internal/data"
	"github.com/Sukrati192/greenlight/internal/validator"
	"github.com/gin-gonic/gin"
)

func (app *application) createMoviesHandler(c *gin.Context) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	if err := app.readJSON(c, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	if err := app.models.Movies.Insert(movie); err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	if err := app.writeJSON(c, http.StatusOK, envelope{"movie": movie}, headers); err != nil {
		app.serverErrorResponse(c, err)
	}
}

func (app *application) showMoviesHandler(c *gin.Context) {
	id, err := app.readIDParam(c)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if err := app.writeJSON(c, http.StatusOK, envelope{"movie": movie}, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}

func (app *application) updateMoviesHandler(c *gin.Context) {
	id, err := app.readIDParam(c)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if expectedVersion := c.Request.Header.Get("X-Expected-Version"); expectedVersion != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != expectedVersion {
			app.editConflictResponse(c)
			return
		}
	}
	var input struct {
		Title   *string       `json:"title,omitempty"`
		Year    *int32        `json:"year,omitempty"`
		Runtime *data.Runtime `json:"runtime,omitempty"`
		Genres  []string      `json:"genres,omitempty"`
	}
	if err := app.readJSON(c, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	if err := app.models.Movies.Update(movie); err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if err := app.writeJSON(c, http.StatusOK, envelope{"movie": movie}, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}

func (app *application) deleteMoviesHandler(c *gin.Context) {
	id, err := app.readIDParam(c)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}
	if err := app.models.Movies.Delete(id); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if err := app.writeJSON(c, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil); err != nil {
		app.serverErrorResponse(c, err)
		return
	}
}

func (app *application) listMoviesHandler(c *gin.Context) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()
	qs := c.Request.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.writeJSON(c, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
}
