package request

import (
	"encoding/json"
	"net/http"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

type cookie http.Cookie

func (c cookie) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Name     string        `json:"name"`
		Value    string        `json:"value"`
		Path     string        `json:"path,omitempty"`
		Domain   string        `json:"domain,omitempty"`
		Expires  core.Time     `json:"expires,omitempty"`
		MaxAge   int           `json:"maxAge,omitempty"`
		Secure   bool          `json:"secure,omitempty"`
		HttpOnly bool          `json:"httpOnly,omitempty"`
		SameSite http.SameSite `json:"sameSite,omitempty"`
	}{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  core.Time(c.Expires),
		MaxAge:   c.MaxAge,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
		SameSite: c.SameSite,
	})

	return data, errors.JSONMarshalError.Wrap(err)
}
