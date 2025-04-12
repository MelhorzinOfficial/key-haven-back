package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"key-haven-back/pkg/secret"
	"testing"
	"time"
)

func Test_checkToken(t *testing.T) {
	token := getToken()

	app := fiber.New()

	tests := map[string]struct {
		setup   func(c fiber.Ctx)
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		"valid": {
			setup: func(c fiber.Ctx) {
				c.Request().Header.SetCookie("token", token)
			},
			want:    token,
			wantErr: assert.NoError,
		},
		"token invalid": {
			setup: func(c fiber.Ctx) {
				c.Request().Header.SetCookie("token", token[1:])
			},
			wantErr: assert.Error,
		},
		"empty token": {
			setup: func(c fiber.Ctx) {
				c.Request().Header.SetCookie("token", "")
			},
			wantErr: assert.Error,
		},
		"nil cookie": {
			setup:   func(c fiber.Ctx) {},
			wantErr: assert.Error,
		},
	}
	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			tt.setup(c)

			got, err := checkToken(c)
			if !tt.wantErr(t, err, fmt.Sprintf("checkToken(%v)", c)) {
				return
			}

			assert.Equalf(t, tt.want, got, "checkToken(%v)", c)
		})
	}
}

func Test_checkCookieToken(t *testing.T) {
	token := getToken()

	tests := map[string]struct {
		tokenJwt string
		want     string
		wantErr  assert.ErrorAssertionFunc
	}{
		"valid": {
			tokenJwt: token,
			want:     token,
			wantErr:  assert.NoError,
		},
		"invalid": {
			tokenJwt: token[1:],
			want:     "",
			wantErr:  assert.Error,
		},
	}
	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			got, err := checkCookieToken(tt.tokenJwt)
			if !tt.wantErr(t, err, fmt.Sprintf("checkCookieToken(%v)", tt.tokenJwt)) {
				return
			}

			assert.Equalf(t, tt.want, got, "checkCookieToken(%v)", tt.tokenJwt)
		})
	}
}

func Test_checkAuthToken(t *testing.T) {
	token := getToken()

	tests := map[string]struct {
		authHeader string
		want       string
		wantErr    assert.ErrorAssertionFunc
	}{
		"valid": {
			authHeader: fmt.Sprintf("Bearer %s", token),
			want:       token,
			wantErr:    assert.NoError,
		},
		"token invalid": {
			authHeader: fmt.Sprintf("Bearer %s", token[1:]),
			wantErr:    assert.Error,
		},
		"invalid format": {
			authHeader: "Bearer",
			wantErr:    assert.Error,
		},
		"invalid header": {
			authHeader: "Basic " + token,
			wantErr:    assert.Error,
		},
	}
	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			got, err := checkAuthToken(tt.authHeader)
			if !tt.wantErr(t, err, fmt.Sprintf("checkAuthToken(%v)", tt.authHeader)) {
				return
			}
			assert.Equalf(t, tt.want, got, "checkAuthToken(%v)", tt.authHeader)
		})
	}
}

func Test_validateTokenFormat(t *testing.T) {
	tests := map[string]struct {
		token string
		want  bool
	}{
		"valid": {
			token: getToken(),
			want:  true,
		},
		"invalid": {
			token: getToken()[1:],
			want:  false,
		},
	}
	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			got := validateTokenFormat(tt.token)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getToken() string {
	s, _ := secret.GenerateToken("0", "test@example.com", time.Hour)
	return s
}
