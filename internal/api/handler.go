package api

import (
	"github.com/example/go-otp-auth/internal/config"
	pg "github.com/example/go-otp-auth/internal/storage"
)

type Handler struct {
	pg  *pg.Postgres
	rd  *pg.Redis
	cfg *config.Config
}

func NewHandler(pg *pg.Postgres, rd *pg.Redis, cfg *config.Config) *Handler {
	return &Handler{pg: pg, rd: rd, cfg: cfg}
}
