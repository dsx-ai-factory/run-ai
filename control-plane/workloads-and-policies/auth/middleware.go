package auth

import (
	"encoding/json"
	"net/http"
)

var ForbiddenError = Error{
	Code:    http.StatusForbidden,
	Message: "not permitted to perform this operation",
}

func ExtractAuthInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isOk := ValidateAuthenticationHeaders(r)
		if !isOk {
			errText, jsonErr := json.Marshal(ForbiddenError)
			if jsonErr != nil {
				http.Error(w, ForbiddenError.Message, http.StatusForbidden)
			} else {
				http.Error(w, string(errText), http.StatusForbidden)
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
