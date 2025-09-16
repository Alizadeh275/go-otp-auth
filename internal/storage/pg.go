package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/example/go-otp-auth/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	var db *sqlx.DB
	var err error

	maxAttempts := 10
	wait := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sqlx.Connect("pgx", dsn)
		if err == nil {
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(2)
			db.SetConnMaxLifetime(time.Hour)
			// verify with small ping context
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if pingErr := db.PingContext(ctx); pingErr == nil {
				log.Info().Msg("connected to postgres")
				return &Postgres{db: db}, nil
			} else {
				err = pingErr
			}
		}

		log.Warn().Err(err).Int("attempt", attempt).Msg("postgres not ready, retrying")
		time.Sleep(wait)
		wait *= 2 // exponential backoff (2s,4s,8s,...)
	}

	return nil, fmt.Errorf("unable to connect to postgres after %d attempts: %w", maxAttempts, err)
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	err := p.db.GetContext(ctx, &u, "SELECT id, phone, registered_at FROM users WHERE phone=$1", phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) CreateUser(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	err := p.db.GetContext(ctx, &u, "INSERT INTO users (phone) VALUES ($1) RETURNING id, phone, registered_at", phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := p.db.GetContext(ctx, &u, "SELECT id, phone, registered_at FROM users WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type User struct {
	ID           int64  `db:"id" json:"id"`
	Phone        string `db:"phone" json:"phone"`
	RegisteredAt string `db:"registered_at" json:"registered_at"`
}

// ListUsers returns users with optional search and pagination
func (pg *Postgres) ListUsers(ctx context.Context, search string, offset, limit int) ([]User, int, error) {
	users := []User{}
	var total int
	var err error

	if search == "" {
		err = pg.db.SelectContext(ctx, &users, `
			SELECT id, phone, registered_at
			FROM users
			ORDER BY id
			LIMIT $1 OFFSET $2`, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		err = pg.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`)
	} else {
		err = pg.db.SelectContext(ctx, &users, `
			SELECT id, phone, registered_at
			FROM users
			WHERE phone LIKE $1
			ORDER BY id
			LIMIT $2 OFFSET $3`, "%"+search+"%", limit, offset)
		if err != nil {
			return nil, 0, err
		}
		err = pg.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users WHERE phone LIKE $1`, "%"+search+"%")
	}

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
