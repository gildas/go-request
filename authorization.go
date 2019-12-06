package request

import "encoding/base64"

// BasicAuthorization builds a basic authorization string
func BasicAuthorization(user, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
}

// TokenAuthorization builds a Token authorization string
func TokenAuthorization(token string) string {
	return "token " + token
}