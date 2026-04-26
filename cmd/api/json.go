package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

func (app *application) writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Message string `json:"message"`
	}

	return app.writeJSON(w, status, envelope{
		message,
	})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return app.writeJSON(w, status, &envelope{data})
}
