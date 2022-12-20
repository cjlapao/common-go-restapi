package controllers

import (
	"net/http"
	"strconv"
)

func SetContentType(contentType string, w http.ResponseWriter) {
	w.Header().Del("content-type")
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", contentType)
}

func SetContentLength(contentLength int, w http.ResponseWriter) {
	w.Header().Del("content-length")
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Length", strconv.Itoa(contentLength))
}
