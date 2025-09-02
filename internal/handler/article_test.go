package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/internal/handler"
	"github.com/bxcodec/go-clean-arch/internal/handler/mocks"
	"github.com/gin-gonic/gin"
	faker "github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestFetch(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockUCase := new(mocks.ArticleService)
	mockListArticle := make([]domain.Article, 0)
	mockListArticle = append(mockListArticle, mockArticle)
	num := 1
	cursor := "2"
	mockUCase.On("Fetch", mock.Anything, cursor, int64(num)).Return(mockListArticle, "10", nil)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles?num=1&cursor="+cursor, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	responseCursor := w.Header().Get("X-Cursor")
	assert.Equal(t, "10", responseCursor)
	assert.Equal(t, http.StatusOK, w.Code)
	mockUCase.AssertExpectations(t)
}

func TestFetchError(t *testing.T) {
	mockUCase := new(mocks.ArticleService)
	num := 1
	cursor := "2"
	mockUCase.On("Fetch", mock.Anything, cursor, int64(num)).Return(nil, "", domain.ErrInternalServerError)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles?num=1&cursor="+cursor, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	responseCursor := w.Header().Get("X-Cursor")
	assert.Equal(t, "", responseCursor)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUCase.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockUCase := new(mocks.ArticleService)
	num := int(mockArticle.ID)
	mockUCase.On("GetByID", mock.Anything, int64(num)).Return(mockArticle, nil)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles/"+strconv.Itoa(num), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUCase.AssertExpectations(t)
}

func TestGetByIDInvalidID(t *testing.T) {
	mockUCase := new(mocks.ArticleService)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStore(t *testing.T) {
	mockArticle := domain.Article{
		Title:     "Title",
		Content:   "Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tempMockArticle := mockArticle
	tempMockArticle.ID = 0
	mockUCase := new(mocks.ArticleService)

	j, err := json.Marshal(tempMockArticle)
	assert.NoError(t, err)

	mockUCase.On("Store", mock.Anything, mock.AnythingOfType("*domain.Article")).Return(nil)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUCase.AssertExpectations(t)
}

func TestStoreInvalidJSON(t *testing.T) {
	mockUCase := new(mocks.ArticleService)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDelete(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockUCase := new(mocks.ArticleService)
	num := int(mockArticle.ID)
	mockUCase.On("Delete", mock.Anything, int64(num)).Return(nil)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/articles/"+strconv.Itoa(num), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUCase.AssertExpectations(t)
}

func TestDeleteInvalidID(t *testing.T) {
	mockUCase := new(mocks.ArticleService)

	r := setupRouter()
	handler.NewArticleHandler(r, mockUCase)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/articles/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
