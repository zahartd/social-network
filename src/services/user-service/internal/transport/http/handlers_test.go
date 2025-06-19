package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zahartd/social-network/src/services/user-service/internal/app"
	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/inmemory"
)

type kafkaMock struct{ mock.Mock }

func (m *kafkaMock) PublishUserRegistered(_ context.Context, _ models.User) error {
	return m.Called().Error(0)
}

func ensureJWT() {
	if os.Getenv("JWT_PRIVATE_KEY") != "" {
		return
	}
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	priv := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pub := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})
	pf, _ := os.CreateTemp("", "priv*.pem")
	qf, _ := os.CreateTemp("", "pub*.pem")
	_ = os.WriteFile(pf.Name(), priv, 0600)
	_ = os.WriteFile(qf.Name(), pub, 0600)
	os.Setenv("JWT_PRIVATE_KEY", pf.Name())
	os.Setenv("JWT_PUBLIC_KEY", qf.Name())
}

func newRouter() (*gin.Engine, *inmemory.InMemoryUserRepo, *inmemory.InMemorySessionRepo) {
	gin.SetMode(gin.TestMode)
	ensureJWT()

	ur := inmemory.NewInMemoryUserRepo()
	sr := inmemory.NewInMemorySessionRepo()
	km := &kafkaMock{}
	km.On("PublishUserRegistered", mock.Anything, mock.Anything).Return(nil)

	auth.Init()

	svc := service.NewUser(ur, sr, km)

	r := gin.New()
	app.RegisterValidators(r)
	AttachRoutes(r, svc)
	return r, ur, sr
}

func call(r *gin.Engine, method, path string, body any, hdr map[string]string) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUserAPI(t *testing.T) {
	r, ur, sr := newRouter()

	valid := map[string]string{
		"login": "john", "firstname": "J", "surname": "D",
		"email": "j@example.com", "password": "Password123",
	}

	t.Run("POST /user", func(t *testing.T) {
		t.Run("201 created", func(t *testing.T) {
			w := call(r, http.MethodPost, "/user", valid, nil)
			assert.Equal(t, http.StatusCreated, w.Code)
		})
		t.Run("400 duplicate", func(t *testing.T) {
			w := call(r, http.MethodPost, "/user", valid, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("400 wrong content-type", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`<xml/>`))
			req.Header.Set("Content-Type", "application/xml")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("400 bad json", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{ bad }`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("400 invalid fields", func(t *testing.T) {
			invalid := map[string]string{
				"login": "1bad", "firstname": "", "surname": "",
				"email": "bad", "password": "123",
			}
			w := call(r, http.MethodPost, "/user", invalid, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("GET /user/login", func(t *testing.T) {
		t.Run("200 ok", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/login?login=john&password=Password123", nil, nil)
			assert.Equal(t, http.StatusOK, w.Code)
		})
		t.Run("400 wrong password", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/login?login=john&password=wrong", nil, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("400 missing params", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/login", nil, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	login := func() string {
		w := call(r, http.MethodGet, "/user/login?login=john&password=Password123", nil, nil)
		var resp struct{ Token string }
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return resp.Token
	}

	token := login()
	authHdr := map[string]string{"Authorization": "Bearer " + token}

	t.Run("GET /user/logout", func(t *testing.T) {
		t.Run("200 ok", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/logout", nil, authHdr)
			assert.Equal(t, http.StatusOK, w.Code)
			_, err := sr.GetByToken(token)
			assert.Error(t, err)
		})
		t.Run("400 missing token", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/logout", nil, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	token = login()
	authHdr = map[string]string{"Authorization": "Bearer " + token}
	johnID := func() string { u, _ := ur.GetByLogin("john"); return u.ID }()

	wCreateBob := call(r, http.MethodPost, "/user", map[string]string{
		"login": "bob", "firstname": "B", "surname": "B",
		"email": "b@ex.com", "password": "Password123",
	}, nil)
	assert.Equal(t, http.StatusCreated, wCreateBob.Code)

	t.Run("GET /user/:identifier", func(t *testing.T) {
		t.Run("200 self by id", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/"+johnID, nil, authHdr)
			assert.Equal(t, http.StatusOK, w.Code)
		})
		t.Run("200 self by login", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/john", nil, authHdr)
			assert.Equal(t, http.StatusOK, w.Code)
		})
		t.Run("400 invalid identifier", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/BadLogin", nil, authHdr)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("404 not found", func(t *testing.T) {
			w := call(r, http.MethodGet, "/user/ghost", nil, authHdr)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	update := map[string]string{
		"email": "new@example.com", "firstname": "J", "surname": "D",
		"phone": "+1234567890", "bio": "bio",
	}

	t.Run("PUT /user/:identifier", func(t *testing.T) {
		t.Run("200 ok", func(t *testing.T) {
			w := call(r, http.MethodPut, "/user/john", update, authHdr)
			assert.Equal(t, http.StatusOK, w.Code)
		})
		t.Run("400 bad body", func(t *testing.T) {
			w := call(r, http.MethodPut, "/user/john",
				map[string]string{"email": "bad"}, authHdr)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("401 unauthorized other", func(t *testing.T) {
			w := call(r, http.MethodPut, "/user/bob", update, authHdr)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("DELETE /user/:identifier", func(t *testing.T) {
		t.Run("401 unauthorized", func(t *testing.T) {
			w := call(r, http.MethodDelete, "/user/john", nil, nil)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
		t.Run("404 not found", func(t *testing.T) {
			w := call(r, http.MethodDelete, "/user/ghost", nil, authHdr)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
		t.Run("200 ok", func(t *testing.T) {
			w := call(r, http.MethodDelete, "/user/john", nil, authHdr)
			assert.Equal(t, http.StatusOK, w.Code)
			_, err := ur.GetByLogin("john")
			assert.Error(t, err)
		})
	})
}
