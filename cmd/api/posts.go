package main

import (
	"context"
	"errors"
	"net/http"
	"socialv3/internal/store"
)

type postKey string

const postCtx postKey = "post"

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required,max=255"`
	Tags    []string `json:"tags"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,max=200"`
	Content *string `json:"content" validate:"omitempty,max=255"`
}

// createPostHandler godoc
//
//	@Summary		Buat post baru
//	@Description	Membuat post baru. Memerlukan title, content, dan optional tags.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreatePostRequest	true	"Data post"
//	@Success		201		{object}	store.Post			"Post yang baru dibuat"
//	@Failure		400		{object}	object				"Request body tidak valid"
//	@Failure		500		{object}	object				"Server error"
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	ctx := r.Context()

	var payload CreatePostRequest
	if err := app.readJSON(w, r, &payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.jsonResponse(w, http.StatusCreated, post)
}

// getPostHandler godoc
//
//	@Summary		Ambil detail post
//	@Description	Mengambil satu post berdasarkan ID beserta daftar komentar.
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		int			true	"ID post"
//	@Success		200		{object}	store.Post	"Detail post dengan komentar"
//	@Failure		400		{object}	object		"ID tidak valid"
//	@Failure		404		{object}	object		"Post tidak ditemukan"
//	@Failure		500		{object}	object		"Server error"
//	@Router			/posts/{postID} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := app.getPostFromCtx(r)
	if err != nil {
		app.notFoundError(w, r, err)
		return
	}

	ctx := r.Context()
	comments, err := app.store.Comments.GetAllByPostID(ctx, post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	_ = app.jsonResponse(w, http.StatusOK, post)
}

// updatePostHandler godoc
//
//	@Summary		Update post
//	@Description	Mengubah title dan/atau content post. Field bersifat optional (partial update).
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"ID post"
//	@Param			body	body		UpdatePostRequest	true	"Data yang di-update (title, content)"
//	@Success		200		{object}	store.Post			"Post setelah di-update"
//	@Failure		400		{object}	object				"Request body tidak valid"
//	@Failure		404		{object}	object				"Post tidak ditemukan"
//	@Failure		500		{object}	object				"Server error"
//	@Router			/posts/{postID} [put]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := app.getPostFromCtx(r)
	if err != nil {
		app.notFoundError(w, r, err)
		return
	}

	var payload UpdatePostRequest
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	//
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	ctx := r.Context()
	err = app.store.Posts.Update(ctx, post)
	if err != nil {
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

// deletePostHandler godoc
//
//	@Summary		Hapus post
//	@Description	Menghapus post berdasarkan ID.
//	@Tags			posts
//	@Param			postID	path	int	true	"ID post"
//	@Success		204		"No content"
//	@Failure		400		{object}	object	"ID tidak valid"
//	@Failure		404		{object}	object	"Post tidak ditemukan"
//	@Failure		500		{object}	object	"Server error"
//	@Router			/posts/{postID} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := app.readID(r, "postID")
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	err = app.store.Posts.Delete(ctx, id)
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

func (app *application) getPostFromCtx(r *http.Request) (*store.Post, error) {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	if post == nil {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := app.readID(r, "postID")
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
