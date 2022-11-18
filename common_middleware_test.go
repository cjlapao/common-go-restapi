package restapi

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestCorrelationMiddlewareAdapter(t *testing.T) {
// 	// create a handler to use as "next" which will verify the request
// 	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		val := r.Header.Get("X-Correlation-Id")
// 		if val == "" {
// 			t.Error("X-Correlation-Id header not present")
// 		}

// 		if val != "1234" {
// 			t.Error("wrong reqId")
// 		}
// 	})

// 	handlerTest := CorrelationMiddlewareAdapter(false)
// 	req := httptest.NewRequest("GET", "http://testing", nil)

// 	handlerTest.ServeHTTP(httptest.NewRecorder(), req)
// }
