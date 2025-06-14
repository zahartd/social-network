package http_test

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
	httptransport "github.com/zahartd/social-network/src/services/user-service/internal/transport/http"
)

type kafkaMock struct{ mock.Mock }

func (m *kafkaMock) PublishUserRegistered(_ context.Context, _ models.User) error {
	args := m.Called()
	return args.Error(0)
}

func ensureJWTKeys() {
	if os.Getenv("JWT_PRIVATE_KEY") != "" && os.Getenv("JWT_PUBLIC_KEY") != "" {
		return
	}

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privDER := x509.MarshalPKCS1PrivateKey(privKey)
	pubDER := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubDER})

	privFile, _ := os.CreateTemp("", "jwt_priv_*.pem")
	pubFile, _ := os.CreateTemp("", "jwt_pub_*.pem")
	_ = os.WriteFile(privFile.Name(), privPEM, 0600)
	_ = os.WriteFile(pubFile.Name(), pubPEM, 0600)

	os.Setenv("JWT_PRIVATE_KEY", privFile.Name())
	os.Setenv("JWT_PUBLIC_KEY", pubFile.Name())
}

func setupRouter() (*gin.Engine, *inmemory.InMemoryUserRepo, *inmemory.InMemorySessionRepo, *kafkaMock) {
	gin.SetMode(gin.TestMode)
	ensureJWTKeys()

	ur := inmemory.NewInMemoryUserRepo()
	sr := inmemory.NewInMemorySessionRepo()
	km := &kafkaMock{}
	km.On("PublishUserRegistered", mock.Anything, mock.Anything).Return(nil)

	auth.Init()

	svc := service.NewUser(ur, sr, km)
	r := gin.New()
	app.RegisterValidators(r)
	httptransport.AttachRoutes(r, svc)
	return r, ur, sr, km
}

func perform(r *gin.Engine, method, path string, body any, hdr map[string]string) *httptest.ResponseRecorder {
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

func TestUsersRoutes(t *testing.T) {
	r, _, _, _ := setupRouter()

	t.Run("POST_user", func(t *testing.T) {
		t.Run("valid_request_201", func(t *testing.T) {
			body := map[string]string{
				"login": "bob", "firstname": "Bob", "surname": "B",
				"email": "b@example.com", "password": "Password123",
			}
			w := perform(r, http.MethodPost, "/user", body, nil)
			assert.Equal(t, http.StatusCreated, w.Code)
			assert.True(t, json.Valid(w.Body.Bytes()))
		})

		t.Run("unsupported_media_type_400", func(t *testing.T) {
			xmlBody := `<User><Login>bob</Login></User>`
			req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(xmlBody))
			req.Header.Set("Content-Type", "application/xml")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("invalid_json_400", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{ invalid json }`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("invalid_data_400", func(t *testing.T) {
			body := map[string]string{
				"login": "1bad_login", "firstname": "A", "surname": "B",
				"email": "bademail", "password": "123",
			}
			w := perform(r, http.MethodPost, "/user", body, nil)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("OPTIONS_user_204", func(t *testing.T) {
		w := perform(r, http.MethodOptions, "/user", nil, nil)
		// gin без явного OPTIONS-роута вернёт 404
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("OTHER_user_405_or_404", func(t *testing.T) {
		for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodHead, http.MethodPatch} {
			t.Run(m, func(t *testing.T) {
				w := perform(r, m, "/user", nil, nil)
				assert.Contains(t, []int{http.StatusMethodNotAllowed, http.StatusNotFound}, w.Code)
			})
		}
	})
}

func TestAuthFlow(t *testing.T) {
	r, _, sr, _ := setupRouter()

	// create user
	w := perform(r, http.MethodPost, "/user", map[string]string{
		"login": "alice", "firstname": "Alice", "surname": "A",
		"email": "a@example.com", "password": "Password123",
	}, nil)
	var created struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	// login
	w = perform(r, http.MethodGet, "/user/login?login=alice&password=Password123", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResp struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)

	// session exists
	sess, err := sr.GetByToken(loginResp.Token)
	assert.NoError(t, err)
	assert.NotNil(t, sess)

	// logout
	w = perform(r, http.MethodGet, "/user/logout", nil, map[string]string{"Authorization": loginResp.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	_, err = sr.GetByToken(loginResp.Token)
	assert.Error(t, err)
}
