package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/go-otp-auth/internal/auth"
	"github.com/example/go-otp-auth/internal/util"
	"github.com/rs/zerolog/log"
)

type reqPhone struct {
	Phone string `json:"phone"`
}

type reqVerify struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
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
	WriteJSON(w, map[string]string{"status": "otp_generated"})
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

	WriteJSON(w, map[string]interface{}{"token": tok, "user": user})
}
