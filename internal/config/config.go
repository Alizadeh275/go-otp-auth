package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    Port                     int
    DatabaseURL              string
    RedisAddr                string
    JWTSecret                string
    OTPTTLSeconds            int
    RateLimitMax             int
    RateLimitWindowSeconds   int
}

func LoadFromEnv() (*Config, error) {
    port := 8080
    if p := os.Getenv("PORT"); p != "" {
        if pi, err := strconv.Atoi(p); err == nil {
            port = pi
        }
    }
    db := os.Getenv("DATABASE_URL")
    if db == "" {
        return nil, fmt.Errorf("DATABASE_URL required")
    }
    r := os.Getenv("REDIS_ADDR")
    if r == "" {
        return nil, fmt.Errorf("REDIS_ADDR required")
    }
    jwt := os.Getenv("JWT_SECRET")
    if jwt == "" {
        return nil, fmt.Errorf("JWT_SECRET required")
    }
    otpTTLS := 120
    if v := os.Getenv("OTP_TTL_SECONDS"); v != "" {
        if vi, err := strconv.Atoi(v); err == nil {
            otpTTLS = vi
        }
    }
    rlMax := 3
    if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
        if vi, err := strconv.Atoi(v); err == nil {
            rlMax = vi
        }
    }
    rlWindow := 600
    if v := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); v != "" {
        if vi, err := strconv.Atoi(v); err == nil {
            rlWindow = vi
        }
    }
    return &Config{
        Port: port,
        DatabaseURL: db,
        RedisAddr: r,
        JWTSecret: jwt,
        OTPTTLSeconds: otpTTLS,
        RateLimitMax: rlMax,
        RateLimitWindowSeconds: rlWindow,
    }, nil
}