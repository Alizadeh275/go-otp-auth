# -------- BUILD STAGE --------
    FROM golang:1.24-bullseye AS build

    WORKDIR /app
    
    # کپی go.mod و go.sum
    COPY go.mod go.sum ./
    
    # دانلود dependencyها (اگر دسترسی اینترنت باشد)
    RUN go mod download
    
    # کپی کل سورس
    COPY . .
    
    # Build برنامه
    RUN go build -o main ./cmd/server
    
    # -------- FINAL STAGE --------
    FROM debian:bullseye-slim
    
    WORKDIR /app
    
    RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
    
    COPY --from=build /app/main .
    
    EXPOSE 8080
    CMD ["./main"]
    