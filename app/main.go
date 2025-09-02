package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	mysqlRepo "github.com/bxcodec/go-clean-arch/internal/repository/mysql"

	"github.com/bxcodec/go-clean-arch/article"
	"github.com/bxcodec/go-clean-arch/internal/handler"
	"github.com/bxcodec/go-clean-arch/internal/handler/middleware"
)

const (
	defaultTimeout = 30
	defaultAddress = ":9090"
)

func init() {
	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file:", err)
	}
}

func main() {
	// 设置Gin模式
	if !viper.GetBool("debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	// 准备数据库连接
	dbHost := viper.GetString("database.host")
	dbPort := viper.GetString("database.port")
	dbUser := viper.GetString("database.user")
	dbPass := viper.GetString("database.password")
	dbName := viper.GetString("database.name")
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil {
		log.Fatal("failed to open connection to database", err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal("failed to ping database ", err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("got error when closing the DB connection", err)
		}
	}()

	// 准备Gin引擎
	r := gin.New()

	// 创建日志实例
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// 注册中间件
	r.Use(gin.Logger())
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.ErrorMiddleware(logger))
	r.Use(middleware.CORS())

	// 设置超时中间件
	timeout := viper.GetInt("context.timeout")
	if timeout == 0 {
		log.Println("timeout not configured, using default timeout")
		timeout = defaultTimeout
	}
	timeoutContext := time.Duration(timeout) * time.Second
	r.Use(middleware.SetRequestContextWithTimeout(timeoutContext))

	// 准备Repository
	authorRepo := mysqlRepo.NewAuthorRepository(dbConn)
	articleRepo := mysqlRepo.NewArticleRepository(dbConn)

	// 构建Service层
	svc := article.NewService(articleRepo, authorRepo)
	handler.NewArticleHandler(r, svc)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 启动服务器
	address := viper.GetString("server.address")
	if address == "" {
		address = defaultAddress
	}

	log.Printf("Server starting on %s", address)
	log.Fatal(r.Run(address))
}
