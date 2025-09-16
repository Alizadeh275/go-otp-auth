package api

import (
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

// GetUserMe godoc
// @Summary Get current user
// @Description Retrieve details of the authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Security BearerAuth
// @Router /users/me [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.pg.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	WriteJSON(w, u)
}

// ListUsers godoc
// @Summary List users
// @Description List users with optional search and pagination
// @Tags users
// @Produce json
// @Param search query string false "Search by phone (optional)"
// @Param page query int false "Page number (optional, default 1)" default(1)
// @Param size query int false "Page size (optional, default 10)" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "internal server error"
// @Router /users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page := 1
	size := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}
	if s := r.URL.Query().Get("size"); s != "" {
		if si, err := strconv.Atoi(s); err == nil && si > 0 {
			if si > 100 {
				size = 100
			} else {
				size = si
			}
		}
	}

	offset := (page - 1) * size
	users, total, err := h.pg.ListUsers(r.Context(), search, offset, size)
	if err != nil {
		log.Error().Err(err).Msg("list users")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"total": total,
		"page":  page,
		"size":  size,
		"data":  users,
	})
}
