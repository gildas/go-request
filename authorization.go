package request

import "encoding/base64"

// BasicAuthorization builds a basic authorization string
func BasicAuthorization(user, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
}

// BearerAuthorization builds a Token authorization string
func BearerAuthorization(token string) string {
	return "Bearer " + token
}