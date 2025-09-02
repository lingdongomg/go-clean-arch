package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/internal/handler/middleware"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// ArticleService represent the article's usecases
//
//go:generate mockery --name ArticleService
type ArticleService interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]domain.Article, string, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	Update(ctx context.Context, ar *domain.Article) error
	GetByTitle(ctx context.Context, title string) (domain.Article, error)
	Store(context.Context, *domain.Article) error
	Delete(ctx context.Context, id int64) error
}

// ArticleHandler  represent the httphandler for article
type ArticleHandler struct {
	Service   ArticleService
	validator *validator.Validate
}

const defaultNum = 10

// NewArticleHandler will initialize the articles/ resources endpoint
func NewArticleHandler(r *gin.Engine, svc ArticleService) {
	handler := &ArticleHandler{
		Service:   svc,
		validator: validator.New(),
	}

	// 注册路由
	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", handler.FetchArticle)
		v1.POST("/articles", handler.Store)
		v1.GET("/articles/:id", handler.GetByID)
		v1.DELETE("/articles/:id", handler.Delete)
	}
}

// FetchArticle will fetch the article based on given params
func (a *ArticleHandler) FetchArticle(c *gin.Context) {
	numS := c.DefaultQuery("num", "10")
	num, err := strconv.Atoi(numS)
	if err != nil || num == 0 {
		num = defaultNum
	}

	cursor := c.Query("cursor")
	ctx := c.Request.Context()

	listAr, nextCursor, err := a.Service.Fetch(ctx, cursor, int64(num))
	if err != nil {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(getStatusCode(err), "获取文章列表失败", err))
		return
	}

	c.Header("X-Cursor", nextCursor)
	c.JSON(http.StatusOK, listAr)
}

// GetByID will get article by given id
func (a *ArticleHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	idP, err := strconv.Atoi(idParam)
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	id := int64(idP)
	ctx := c.Request.Context()

	art, err := a.Service.GetByID(ctx, id)
	if err != nil {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(getStatusCode(err), "获取文章失败", err))
		return
	}

	c.JSON(http.StatusOK, art)
}

func (a *ArticleHandler) isRequestValid(m *domain.Article) (bool, error) {
	err := a.validator.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Store will store the article by given request body
func (a *ArticleHandler) Store(c *gin.Context) {
	var article domain.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(http.StatusBadRequest, "请求参数错误", err))
		return
	}

	var ok bool
	var err error
	if ok, err = a.isRequestValid(&article); !ok {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(http.StatusBadRequest, "参数验证失败", err))
		return
	}

	ctx := c.Request.Context()
	err = a.Service.Store(ctx, &article)
	if err != nil {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(getStatusCode(err), "创建文章失败", err))
		return
	}

	c.JSON(http.StatusCreated, article)
}

// Delete will delete article by given param
func (a *ArticleHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	idP, err := strconv.Atoi(idParam)
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	id := int64(idP)
	ctx := c.Request.Context()

	err = a.Service.Delete(ctx, id)
	if err != nil {
		middleware.HandleError(c, middleware.NewAppErrorWithErr(getStatusCode(err), "删除文章失败", err))
		return
	}

	c.Status(http.StatusNoContent)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
