package main

import (
	"net/http"
	"socialv2/internal/store"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
	PostID  int64  `json:"post_id" validate:"required"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload

	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	userId := int64(1)
	createComment := &store.Comment{
		Content: payload.Content,
		UserID:  userId,
		PostID:  payload.PostID,
	}
	if err := app.store.Comment.Create(r.Context(), createComment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

//func (app *application) postContextMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		postID := chi.URLParam(r, "postID")
//		id, err := strconv.ParseInt(postID, 10, 64)
//		if err != nil {
//			app.badRequestError(w, r, fmt.Errorf("invalid post ID"))
//			return
//		}
//
//		ctx := r.Context()
//
//		post, err := app.store.Posts.GetByID(ctx, id)
//		if err != nil {
//			if errors.Is(err, store.ErrNotFound) {
//				app.notFoundError(w, r, err)
//				return
//			}
//			app.internalServerError(w, r, err)
//			return
//		}
//
//		ctx = context.WithValue(r.Context(), postCtx, post)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}
//
//func getPostFromCtx(r *http.Request) *store.Post {
//	post, _ := r.Context().Value(postCtx).(*store.Post)
//	return post
//}
