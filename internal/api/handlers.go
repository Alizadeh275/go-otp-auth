package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/example/go-otp-auth/internal/auth"
	"github.com/example/go-otp-auth/internal/config"
	pg "github.com/example/go-otp-auth/internal/storage"
	"github.com/example/go-otp-auth/internal/util"
)

type Handler struct {
	pg  *pg.Postgres
	rd  *pg.Redis
	cfg *config.Config
}

func NewHandler(pg *pg.Postgres, rd *pg.Redis, cfg *config.Config) *Handler {
	return &Handler{pg: pg, rd: rd, cfg: cfg}
}

// UserResponse used for Swagger documentation
type UserResponse struct {
	ID           int64  `json:"id"`
	Phone        string `json:"phone"`
	RegisteredAt string `json:"registered_at"`
}

type reqPhone struct {
	Phone string `json:"phone"`
}

type reqVerify struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// RequestOTP generates an OTP for login/registration
// @Summary Generate OTP
// @Description Generate an OTP for the given phone number
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body reqPhone true "Phone Number"
// @Success 200 {object} map[string]string "otp_generated"
// @Failure 400 {string} string "invalid request"
// @Failure 429 {string} string "rate limit exceeded"
// @Failure 500 {string} string "internal"
// @Router /otp/request [post]
func (h *Handler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req reqPhone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	allowed, err := h.rd.AllowOTPRequest(ctx, req.Phone, h.cfg.RateLimitMax, time.Duration(h.cfg.RateLimitWindowSeconds)*time.Second)
	if err != nil {
		log.Error().Err(err).Msg("redis error")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	otp, err := util.GenerateOTP()
	if err != nil {
		http.Error(w, "failed to generate otp", http.StatusInternalServerError)
		return
	}
	if err := h.rd.SaveOTP(ctx, req.Phone, otp, time.Duration(h.cfg.OTPTTLSeconds)*time.Second); err != nil {
		log.Error().Err(err).Msg("redis save otp")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	log.Info().Str("phone", req.Phone).Str("otp", otp).Msg("generated otp")
	writeJSON(w, map[string]string{"status": "otp_generated"})
}

// VerifyOTP verifies the OTP and returns JWT token
// @Summary Verify OTP
// @Description Verify OTP and login or register the user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body reqVerify true "Phone and OTP"
// @Success 200 {object} map[string]interface{} "token and user"
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "invalid or expired otp"
// @Failure 500 {string} string "internal"
// @Router /otp/verify [post]
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req reqVerify
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.OTP == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ok, err := h.rd.VerifyAndDeleteOTP(ctx, req.Phone, req.OTP)
	if err != nil {
		log.Error().Err(err).Msg("redis verify otp")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "invalid or expired otp", http.StatusUnauthorized)
		return
	}
	user, err := h.pg.FindUserByPhone(ctx, req.Phone)
	if err != nil {
		u, cerr := h.pg.CreateUser(ctx, req.Phone)
		if cerr != nil {
			log.Error().Err(cerr).Msg("create user")
			http.Error(w, "internal", http.StatusInternalServerError)
			return
		}
		user = u
	}
	tok, err := auth.CreateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("create token")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"token": tok, "user": user})
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve a single user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	u, err := h.pg.GetUserByID(ctx, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, u)
}

// ListUsers godoc
// @Summary List users
// @Description List users with pagination and search
// @Tags users
// @Produce json
// @Param search query string false "Search by phone"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "internal"
// @Security BearerAuth
// @Router /users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Optional search query
	search := r.URL.Query().Get("search")

	// Pagination defaults
	page := 1
	size := 10

	// Parse page if provided
	if p := r.URL.Query().Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}

	// Parse size if provided, enforce max 100
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
	ctx := r.Context()

	users, total, err := h.pg.ListUsers(ctx, search, offset, size)
	if err != nil {
		log.Error().Err(err).Msg("list users")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"total": total,
		"page":  page,
		"size":  size,
		"data":  users,
	})
}
