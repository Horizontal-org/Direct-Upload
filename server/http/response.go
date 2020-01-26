package http

import "net/http"

func ok(w http.ResponseWriter) {
	send(w, http.StatusOK)
}

func errorValidation(w http.ResponseWriter) {
	send(w, http.StatusBadRequest)
}

func errorConflict(w http.ResponseWriter) {
	send(w, http.StatusConflict)
}

func errorForbidden(w http.ResponseWriter) {
	send(w, http.StatusForbidden)
}

func errorInternal(w http.ResponseWriter) {
	send(w, http.StatusInternalServerError)
}

func send(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
