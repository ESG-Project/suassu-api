package response

import (
	"encoding/json"
	"net/http"
)

type Envelope[T any] struct {
	Data T   `json:"data"`
	Meta any `json:"meta,omitempty"`
}

type MetaCursor struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"nextCursor,omitempty"`
	PrevCursor *string `json:"prevCursor,omitempty"`
	HasMore    bool    `json:"hasMore"`
}

func JSON[T any](w http.ResponseWriter, status int, data T, meta any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope[T]{Data: data, Meta: meta})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
