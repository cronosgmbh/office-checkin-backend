package main

import "net/http"

var ErrorNotImplemented = ErrorResponse{
	Code:   http.StatusNotImplemented,
	Errors: []string{"this endpoint is not implemented right now. please check back later"},
}

var ErrorTokenInvalidOrNotFound = ErrorResponse{
	Code:   http.StatusUnauthorized,
	Errors: []string{"the token is invalid or not found in your request"},
}

var ErrorInternalError = ErrorResponse{
	Code:   http.StatusInternalServerError,
	Errors: []string{"internal server error. please contact your administrator for further assistance"},
}

var ErrorForbidden = ErrorResponse{
	Code:   http.StatusForbidden,
	Errors: []string{"you are not permitted to perform this request"},
}