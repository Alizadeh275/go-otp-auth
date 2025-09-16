package util

import (
    "crypto/rand"
    "fmt"
)

// Generate a secure 6-digit OTP
func GenerateOTP() (string, error) {
    const digits = "0123456789"
    b := make([]byte, 6)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    for i := 0; i < 6; i++ {
        b[i] = digits[int(b[i])%10]
    }
    return fmt.Sprintf("%s", string(b)), nil
}