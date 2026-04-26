package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"socialv2/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
	//UserID  int64    `json:"user_id"`
}

type UpdatePostPayload struct {
	Title   *string   `json:"title" validate:"omitempty,max=100"`
	Content *string   `json:"content" validate:"omitempty,max=1000"`
	Tags    *[]string `json:"tags"`
	//UserID  int64    `json:"user_id"`
}

func (app *application) listPostHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *application) detailPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	comments, err := app.store.Comment.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
	}
	post.Comments = comments

	_ = app.jsonResponse(w, http.StatusOK, post)
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = *payload.Tags
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, post)
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	err := app.store.Posts.Delete(r.Context(), post)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusNoContent, nil)
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := app.getUserFromContext(r)
	createPost := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}
	if err := app.store.Posts.Create(r.Context(), createPost); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(postID, 10, 64)
		if err != nil {
			app.badRequestError(w, r, fmt.Errorf("invalid post ID"))
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				app.notFoundError(w, r, err)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		ctx = context.WithValue(r.Context(), postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
