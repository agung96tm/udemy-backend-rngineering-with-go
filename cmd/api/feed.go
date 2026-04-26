package main

import (
	"net/http"
	"socialv2/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 0,
		Sort:   "desc",
	}
	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	posts, err := app.store.Posts.GetUserFeeds(ctx, int64(107), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
