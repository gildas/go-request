package request_test

import "testing"

import "github.com/stretchr/testify/assert"

import "github.com/gildas/go-request"

func TestCanCreateBasicAuthorization(t *testing.T) {
	expected := "Basic dXNlcjpwYXNzd29yZA=="
	assert.Equal(t, expected, request.BasicAuthorization("user", "password"))
}

func TestCanCreateTokenAuthorization(t * testing.T) {
	expected := "Bearer mytoken"
	assert.Equal(t, expected, request.BearerAuthorization("mytoken"))
}