package restapi

import "net/http"

func GetRequestId(r *http.Request) string {
	id := r.Context().Value(REQUEST_ID_KEY).(string)

	return id
}
