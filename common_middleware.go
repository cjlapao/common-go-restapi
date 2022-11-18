package restapi

import (
	"net/http"
	"strings"

	"github.com/cjlapao/common-go/controllers"
	"github.com/cjlapao/common-go/execution_context"
)

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
			next.ServeHTTP(w, r)
		})
	}
}

func LoggerMiddlewareAdapter(logHealthCheck bool) controllers.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shouldLog := true
			if strings.ContainsAny(r.URL.Path, "health") && !logHealthCheck {
				shouldLog = false
			}

			if shouldLog {
				globalHttpListener.Logger.Info("[%v] %v from %v", r.Method, r.URL.Path, r.Host)
			}

			next.ServeHTTP(w, r)
		})
	}
}
