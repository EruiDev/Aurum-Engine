package middleware

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiterConfig struct {
	Rate float64
	Burst int
}

type bucket struct {
	tokens    float64
	lastRefil time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	cfg     RateLimiterConfig
}

func newRateLimiter(cfg RateLimiterConfig) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		cfg:     cfg,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) (allowed bool, remaining int, resetAt time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &bucket{tokens: float64(rl.cfg.Burst), lastRefil: now}
		rl.buckets[ip] = b
	}

	elapsed := now.Sub(b.lastRefil).Seconds()
	b.tokens += elapsed * rl.cfg.Rate
	if b.tokens > float64(rl.cfg.Burst) {
		b.tokens = float64(rl.cfg.Burst)
	}
	b.lastRefil = now

	timeToFull := time.Duration((float64(rl.cfg.Burst)-b.tokens)/rl.cfg.Rate*1000) * time.Millisecond
	resetAt = now.Add(timeToFull)

	if b.tokens < 1 {
		return false, 0, resetAt
	}

	b.tokens--
	return true, int(b.tokens), resetAt
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.buckets {
			if time.Since(b.lastRefil) > 3*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimit(cfg RateLimiterConfig, exemptPaths ...string) func(http.Handler) http.Handler {
	rl := newRateLimiter(cfg)
	exempt := make(map[string]struct{}, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := exempt[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			allowed, remaining, resetAt := rl.allow(ip)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Burst))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				retryAfter := time.Until(resetAt).Seconds()
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					Code:    "rate_limit_exceeded",
					Message: "too many requests, please slow down",
				}); err != nil {
					slog.Error("ratelimit: failed to encode response", "err", err)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := len(xff); idx > 0 {
			for i := 0; i < idx; i++ {
				if xff[i] == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
