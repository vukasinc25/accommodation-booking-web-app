package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/vukasinc25/fst-airbnb/token"

	"github.com/gorilla/mux"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a Gorilla middleware for authorization
func AuthMiddleware(tokenMaker token.Maker) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the authorization header from the request
			authorizationHeader := r.Header.Get(authorizationHeaderKey)

			if len(authorizationHeader) == 0 {
				// If authorization header is not provided, return an error
				err := errors.New("authorization header is not provided")
				writeError(w, http.StatusUnauthorized, err)
				return
			}

			// Split the authorization header into fields
			fields := strings.Fields(authorizationHeader)
			if len(fields) < 2 {
				// If the authorization header format is invalid, return an error
				err := errors.New("invalid authorization header format")
				writeError(w, http.StatusUnauthorized, err)
				return
			}

			// Extract the authorization type
			authorizationType := strings.ToLower(fields[0])
			if authorizationType != authorizationTypeBearer {
				// If the authorization type is not supported, return an error
				err := fmt.Errorf("unsupported authorization type %s", authorizationType)
				writeError(w, http.StatusUnauthorized, err)
				return
			}

			// Extract the access token
			accessToken := fields[1]
			payload, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				// If the token verification fails, return an error
				writeError(w, http.StatusUnauthorized, err)
				return
			}

			// Store the payload in the request context
			r = r.WithContext(context.WithValue(r.Context(), AuthorizationPayloadKey, payload))

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}

// writeError writes an error response to the client
func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}
