package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/models"
	sessionRepository "github.com/zahartd/social-network/src/services/user-service/internal/repository/session/inmemory"
	userRepository "github.com/zahartd/social-network/src/services/user-service/internal/repository/user/inmemory"
	"github.com/zahartd/social-network/src/services/user-service/internal/service"
)

func newDummyKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP("127.0.0.1:9999"),
		Topic:                  "test-topic",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
}

func newTestRouter() *gin.Engine {
	ur := userRepository.NewInMemoryUserRepo()
	sr := sessionRepository.NewInMemorySessionRepo()

	auth.SetSessionRepo(sessionRepo)
	userService := service.NewUserService(ur, sr, newDummyKafkaWriter())

	auth.InitJWT()
	r := gin.Default()
	// initCustomValidators()
	SetupRouter(r, userService)
	return r
}

func createTestUser(repo *InMemoryUserRepo, login, password, firstname, surname, email string) (*models.User, string) {
	// Генерим UUID:
	id := uuid.NewString()
	// Хэшируем пароль:
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &models.User{
		ID:           id,
		Login:        login,
		Firstname:    firstname,
		Surname:      surname,
		Email:        email,
		PasswordHash: string(hash),
	}
	// Передать созданного пользователя в репозиторий:
	_ = repo.Create(user)
	return user, string(hash)
}

func doRequest(r *gin.Engine, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ----------------------------
// ТЕСТЫ
// ----------------------------
func TestCreateUser_Success(t *testing.T) {
	r, userRepo, _ := setupRouter()

	// Проверим, что репозиторий пуст.
	assert.Len(t, userRepo.usersByID, 0)

	// Собираем тело запроса:
	payload := map[string]string{
		"login":     "testuser1",
		"firstname": "Иван",
		"surname":   "Иванов",
		"email":     "iv@example.com",
		"password":  "Password123",
	}
	b, _ := json.Marshal(payload)

	w := doRequest(r, http.MethodPost, "/user", b, map[string]string{
		"Content-Type": "application/json",
	})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp struct {
		User  models.User `json:"user"`
		Token string      `json:"token"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Убедимся, что в ответе есть ID и поля совпадают:
	assert.NotEmpty(t, resp.User.ID)
	assert.Equal(t, "testuser1", resp.User.Login)
	assert.Equal(t, "Иван", resp.User.Firstname)
	assert.Equal(t, "Иванов", resp.User.Surname)
	assert.Equal(t, "iv@example.com", resp.User.Email)
	assert.NotEmpty(t, resp.Token)

	// Проверьте, что пользователь реально сохранился в InMemoryUserRepo:
	_, err = userRepo.GetByID(resp.User.ID)
	assert.NoError(t, err)
}

func TestCreateUser_DuplicateLogin(t *testing.T) {
	r, userRepo, _ := setupRouter()

	// Сначала создаём пользователя напрямую через InMemoryUserRepo:
	_, _ = createTestUser(userRepo, "duplicate", "Password123", "A", "B", "dup@example.com")

	// Пробуем сделать тот же login ещё раз через HTTP:
	payload := map[string]string{
		"login":     "duplicate",
		"firstname": "Пётр",
		"surname":   "Петров",
		"email":     "pt@example.com",
		"password":  "Password123",
	}
	b, _ := json.Marshal(payload)
	w := doRequest(r, http.MethodPost, "/user", b, map[string]string{
		"Content-Type": "application/json",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Сообщение об ошибке должно содержать «user with this login already exists»
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "user with this login already exists")
}

func TestLogin_SuccessAndGetUser_SelfAndOther(t *testing.T) {
	r, userRepo, sessionRepo := setupRouter()

	// Создаём двух пользователей:
	u1, _ := createTestUser(userRepo, "alice", "Secret123", "Alice", "Wonder", "alice@example.com")
	u2, _ := createTestUser(userRepo, "bob", "Secret123", "Bob", "Builder", "bob@example.com")

	// 1) Пытаемся залогиниться за «alice»:
	wLogin := doRequest(r, http.MethodGet, "/user/login?login=alice&password=Secret123", nil, nil)
	assert.Equal(t, http.StatusOK, wLogin.Code)

	var loginResp struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)

	// Проверим, что сессия действительно сохранилась:
	_, err = sessionRepo.GetSessionByToken(loginResp.Token)
	assert.NoError(t, err)

	// 2) GET /user/:identifier (self):
	selfPath := "/user/" + u1.ID
	wGetSelf := doRequest(r, http.MethodGet, selfPath, nil, map[string]string{
		"Authorization": "Bearer " + loginResp.Token,
	})
	assert.Equal(t, http.StatusOK, wGetSelf.Code)

	var userFull models.User
	err = json.Unmarshal(wGetSelf.Body.Bytes(), &userFull)
	assert.NoError(t, err)
	assert.Equal(t, u1.Login, userFull.Login)
	assert.Equal(t, u1.Email, userFull.Email)
	assert.Equal(t, u1.Firstname, userFull.Firstname)
	assert.Equal(t, u1.Surname, userFull.Surname)

	// 3) GET /user/:identifier (other):
	otherPath := "/user/" + u2.Login
	wGetOther := doRequest(r, http.MethodGet, otherPath, nil, map[string]string{
		"Authorization": "Bearer " + loginResp.Token,
	})
	assert.Equal(t, http.StatusOK, wGetOther.Code)

	// Ожидаем, что придёт не полный объект, а summary (map без полей ID/CreatedAt/PasswordHash/...):
	var summary map[string]interface{}
	err = json.Unmarshal(wGetOther.Body.Bytes(), &summary)
	assert.NoError(t, err)
	assert.Contains(t, summary, "login")
	assert.Contains(t, summary, "email")
	assert.Contains(t, summary, "firstname")
	assert.Contains(t, summary, "surname")
	assert.NotContains(t, summary, "passwordHash")
	assert.NotContains(t, summary, "id")
}

func TestUpdateUser_SuccessAndUnauthorized(t *testing.T) {
	r, userRepo, sessionRepo := setupRouter()

	// Создаём пользователя «charlie»:
	u, _ := createTestUser(userRepo, "charlie", "Secret123", "Charlie", "Brown", "charlie@example.com")

	// Логинимся за «charlie»:
	wLogin := doRequest(r, http.MethodGet, "/user/login?login=charlie&password=Secret123", nil, nil)
	assert.Equal(t, http.StatusOK, wLogin.Code)
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)

	// Убедимся, что сессия есть:
	_, err := sessionRepo.GetSessionByToken(loginResp.Token)
	assert.NoError(t, err)

	// 1) Пробуем обновить свой профиль (успех):
	updatePayload := map[string]string{
		"email":     "newcharlie@example.com",
		"firstname": "Charles",
		"surname":   "Brownie",
		"phone":     "+71234567890",
		"bio":       "Hello world!",
	}
	updBody, _ := json.Marshal(updatePayload)
	path := "/user/" + u.Login
	wUpdate := doRequest(r, http.MethodPut, path, updBody, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + loginResp.Token,
	})
	assert.Equal(t, http.StatusOK, wUpdate.Code)

	var updatedUser models.User
	_ = json.Unmarshal(wUpdate.Body.Bytes(), &updatedUser)
	assert.Equal(t, "newcharlie@example.com", updatedUser.Email)
	assert.Equal(t, "Charles", updatedUser.Firstname)
	assert.Equal(t, "Brownie", updatedUser.Surname)
	assert.Equal(t, "+71234567890", updatedUser.Phone)
	assert.Equal(t, "Hello world!", updatedUser.Bio)

	// 2) Пробуем обновить чужой профиль (unauthorized):
	// Сначала создадим «dave»:
	_, _ = createTestUser(userRepo, "dave", "Secret123", "Dave", "Smith", "dave@example.com")

	// Логинимся за «dave»:
	wLogin2 := doRequest(r, http.MethodGet, "/user/login?login=dave&password=Secret123", nil, nil)
	assert.Equal(t, http.StatusOK, wLogin2.Code)
	var loginResp2 struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wLogin2.Body.Bytes(), &loginResp2)

	// Теперь «dave» пытается обновить «charlie»:
	wUpdate2 := doRequest(r, http.MethodPut, "/user/"+u.ID, updBody, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + loginResp2.Token,
	})
	assert.Equal(t, http.StatusUnauthorized, wUpdate2.Code)
}

// func TestDeleteUser_SuccessAndUnauthorized(t *testing.T) {
// 	r, userRepo, sessionRepo := SetupRouter()

// 	// Создаём «ellen»:
// 	u, _ := createTestUser(userRepo, "ellen", "Secret123", "Ellen", "Ripley", "ellen@example.com")

// 	// Логинимся за «ellen»:
// 	wLogin := doRequest(r, http.MethodGet, "/user/login?login=ellen&password=Secret123", nil, nil)
// 	assert.Equal(t, http.StatusOK, wLogin.Code)
// 	var loginResp struct {
// 		Token string `json:"token"`
// 	}
// 	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)

// 	// 1) Успешное удаление собственного аккаунта:
// 	wDelete := doRequest(r, http.MethodDelete, "/user/"+u.Login, nil, map[string]string{
// 		"Authorization": "Bearer " + loginResp.Token,
// 	})
// 	assert.Equal(t, http.StatusOK, wDelete.Code)

// 	// Проверим, что теперь пользователь действительно удалён:
// 	_, err := userRepo.GetByID(u.ID)
// 	assert.Error(t, err)

// 	// 2) Попытка удалить вовсе несуществующего или чужого (текущего у нас уже нет):
// 	wDelete2 := doRequest(r, http.MethodDelete, "/user/"+u.ID, nil, map[string]string{
// 		"Authorization": "Bearer " + loginResp.Token,
// 	})
// 	// Поскольку сессия ellen всё ещё есть (мы не удаляем сессию при Logout в тестах),
// 	// но репозиторий не найдёт пользователя — сервис выдаст 404 (Not Found) или 500. По факту DeleteUser возвращает ошибку из repo.Delete.
// 	// В данном случае в handler DeleteUser блок "if err != nil" выдаст 500.
// 	assert.Equal(t, http.StatusInternalServerError, wDelete2.Code)
// }

func TestLogout_SuccessAndMissingToken(t *testing.T) {
	r, userRepo, _ := setupRouter()

	// Сначала создаём «frank» и логинимся:
	_, _ = createTestUser(userRepo, "frank", "Secret123", "Frank", "Castle", "frank@example.com")
	wLogin := doRequest(r, http.MethodGet, "/user/login?login=frank&password=Secret123", nil, nil)
	assert.Equal(t, http.StatusOK, wLogin.Code)
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)

	// Выполняем logout корректно:
	wLogout := doRequest(r, http.MethodGet, "/user/logout", nil, map[string]string{
		"Authorization": loginResp.Token,
	})
	assert.Equal(t, http.StatusOK, wLogout.Code)

	// Попытка logout без токена:
	wLogout2 := doRequest(r, http.MethodGet, "/user/logout", nil, nil)
	assert.Equal(t, http.StatusBadRequest, wLogout2.Code)
}
