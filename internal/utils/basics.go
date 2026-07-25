package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"strconv"
	"time"
)

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func EnvOr(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func Max[T any](sizeFunc func(T) int, values ...T) int {
	if len(values) == 0 {
		return 0
	}
	maxSize := math.MinInt
	for _, value := range values {
		size := sizeFunc(value)
		if size > maxSize {
			maxSize = size
		}
	}
	return maxSize
}

func Min[T any](sizeFunc func(T) int, values ...T) int {
	if len(values) == 0 {
		return 0
	}
	minSize := math.MaxInt
	for _, value := range values {
		size := sizeFunc(value)
		if size < minSize {
			minSize = size
		}
	}
	return minSize
}

func ArgMax[T any](sizeFunc func(T) int, values ...T) T {
	if len(values) == 0 {
		var zeroValue T
		return zeroValue
	}
	maxSize := math.MinInt
	var maxValue T
	for _, value := range values {
		size := sizeFunc(value)
		if size > maxSize {
			maxSize = size
			maxValue = value
		}
	}
	return maxValue
}

func ArgMin[T any](sizeFunc func(T) int, values ...T) T {
	if len(values) == 0 {
		var zeroValue T
		return zeroValue
	}
	minSize := math.MaxInt
	var minValue T
	for _, value := range values {
		size := sizeFunc(value)
		if size < minSize {
			minSize = size
			minValue = value
		}
	}
	return minValue
}

func Hostname() string {
	// Check if it's running in Kubernetes.
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}

	// Is it running inside docker? Then use the hostname.
	// Docker creates a .dockerenv file at the root of the directory tree inside the container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		if hostname, err := os.Hostname(); err == nil {
			return hostname
		}
	}

	// Otherwise, use 'localhost' as a hostname
	return "localhost"
}

func PID() string {
	return strconv.Itoa(os.Getpid())
}

func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func RandomString(length int) string {
	b := make([]byte, length/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func Sleep(ctx context.Context, d time.Duration) (ctxDone bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return true
	case <-t.C:
		return false
	}
}

func NextBackoff(factor float64, maximum, current time.Duration) time.Duration {
	next := time.Duration(factor * float64(current))
	if next > maximum {
		return maximum
	}
	return next
}

func IsContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func ContextWithCancelOnClose[T any](ctx context.Context, stopSig <-chan T) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-stopSig:
			return
		case <-newCtx.Done():
			return
		}
	}()
	return newCtx
}
