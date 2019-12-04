package gatekeeper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Extract a key from an incoming request.
type KeyExtractor func(*http.Request) (Key, error)

type Middleware struct {
	Backend             Keeper
	Extract             KeyExtractor
	ErrorHandler        func(error) http.Handler
	LimitReachedHandler func(Stats) http.Handler
	KeeperHeaderPrefix  string
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, err := m.Extract(r)
		if err != nil {
			m.ErrorHandler(err).ServeHTTP(w, r)
		}
		allowed, stats, err := m.Backend.Allow(r.Context(), key)
		if err != nil {
			m.ErrorHandler(err).ServeHTTP(w, r)
		}

		if !allowed {
			m.LimitReachedHandler(stats).ServeHTTP(w, r)
		}

		w.Header().Add(fmt.Sprintf("%s-Limit", m.KeeperHeaderPrefix), strconv.FormatInt(stats.Limit, 10))
		w.Header().Add(fmt.Sprintf("%s-Remaining", m.KeeperHeaderPrefix), strconv.FormatInt(stats.Remaining, 10))

		next.ServeHTTP(w, enrichRequest(r, stats))
	})
}

var (
	ErrKeyNotFound      = errors.New("missing API key")
	ErrUnableToCheckKey = errors.New("unable to check key")
)

// Builds a middleware from a Keeper, with a default configuration:
// * api key is read from APIKEY HTTP header, missing key will return a 401
// * in case of error, will return a 500
// * header prefix is X-API-
func FromKeeper(k Keeper) *Middleware {
	return &Middleware{
		Backend: k,
		Extract: KeyExtractor(func(r *http.Request) (Key, error) {
			rawvalue := r.Header.Get("APIKEY")
			if rawvalue == "" {
				return "", ErrKeyNotFound
			} else {
				return Key(rawvalue), nil
			}
		}),
		ErrorHandler: func(e error) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if e == ErrKeyNotFound {
					http.Error(w, "missing API key", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			})
		},
		LimitReachedHandler: func(s Stats) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "limit reached", http.StatusTooManyRequests)
			})
		},
		KeeperHeaderPrefix: "X-API-",
	}
}

// used to include data in requests
type ctxKey int

const (
	statsKey ctxKey = 1
)

func enrichRequest(r *http.Request, stats Stats) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), statsKey, stats))
}

func GetStats(r *http.Request) (Stats, error) {
	if s, ok := r.Context().Value(statsKey).(Stats); !ok {
		return Stats{}, errors.New("not stats in this context")
	} else {
		return s, nil
	}
}
