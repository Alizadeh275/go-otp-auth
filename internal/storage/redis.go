package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) (*Redis, error) {
	opt := &redis.Options{Addr: addr}
	client := redis.NewClient(opt)

	maxAttempts := 8
	wait := 1 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := client.Ping(ctx).Err()
		cancel()
		if err == nil {
			log.Info().Str("addr", addr).Msg("connected to redis")
			return &Redis{client: client}, nil
		}
		log.Warn().Err(err).Int("attempt", attempt).Msg("redis not ready, retrying")
		time.Sleep(wait)
		wait *= 2
	}

	return nil, fmt.Errorf("unable to connect to redis at %s after %d attempts", addr, maxAttempts)
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// OTP
func (r *Redis) SaveOTP(ctx context.Context, phone, code string, ttl time.Duration) error {
	key := fmt.Sprintf("otp:%s", phone)
	return r.client.Set(ctx, key, code, ttl).Err()
}

func (r *Redis) VerifyAndDeleteOTP(ctx context.Context, phone, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if val != code {
		return false, nil
	}
	// delete
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// Rate limiter: increment and return whether allowed
func (r *Redis) AllowOTPRequest(ctx context.Context, phone string, max int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rl:%s", phone)
	// INCR
	n, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		// set expire
		if err := r.client.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}
	return n <= int64(max), nil
}
