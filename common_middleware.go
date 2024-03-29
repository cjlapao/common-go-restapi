package restapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	execution_context "github.com/cjlapao/common-go-execution-context"
	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/google/uuid"
)

const (
	REQUEST_ID_KEY = "requestId"
)

func RequestIdMiddlewareAdapter() controllers.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			r.Header.Add("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), REQUEST_ID_KEY, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JsonContentMiddlewareAdapter() controllers.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

func CorrelationMiddlewareAdapter(logHealthCheck bool) controllers.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := execution_context.Get()
			logger := ctx.Services.Logger
			ctx.Refresh()
			shouldLog := true
			if strings.ContainsAny(r.URL.Path, "health") && !logHealthCheck {
				shouldLog = false
			}

			if shouldLog {
				logger.Info("Http request with correlation %v", ctx.CorrelationId)
			}

			r.Header.Add("X-Correlation-Id", ctx.CorrelationId)
			next.ServeHTTP(w, r)
		})
	}
}

func LoggerMiddlewareAdapter(logHealthCheck bool) controllers.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shouldLog := true
			rMatch := regexp.MustCompile("health")
			if rMatch.MatchString(r.URL.Path) && !logHealthCheck {
				shouldLog = false
			}

			if shouldLog {
				id := GetRequestId(r)
				globalHttpListener.Logger.Info("[%s] [%v] %v from %v", id, r.Method, r.URL.Path, r.Host)
			}

			next.ServeHTTP(w, r)
		})
	}
}
