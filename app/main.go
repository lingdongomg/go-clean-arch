package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	mysqlRepo "github.com/bxcodec/go-clean-arch/internal/repository/mysql"

	"github.com/bxcodec/go-clean-arch/article"
	"github.com/bxcodec/go-clean-arch/internal/handler"
	"github.com/bxcodec/go-clean-arch/internal/handler/middleware"
	log "github.com/lingdongomg/g-lib/logger"
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
		// 在日志系统初始化之前，使用标准库
		fmt.Printf("Error reading config file: %v\n", err)
		panic(err)
	}
}

func main() {
	// 示例1：没有进行任何初始化，直接引用包名进行打印，打印输出到当前default.log文件中
	log.Info("应用启动中...")

	// 示例2：通过文件进行配置实例化，实例化后可以使用返回值logger打印，也可以直接使用包名进行打印（则可以忽略返回值logger）
	// 规范建议是统一使用包名log.XXX进行日志输出，另外任何框架都必须包括如下的日志配置文件，配置文件名不能随意更改
	_, err := log.NewZapLogger("./configs/log.conf.yaml")
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		log.Warn("日志配置文件不存在，使用默认配置:", err)
	}

	log.Info("日志系统初始化完成")

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
		log.Fatal("failed to ping database", err)
	}

	log.Info("数据库连接成功")

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("got error when closing the DB connection", err)
		}
	}()

	// 准备Gin引擎
	r := gin.New()

	// 注册中间件
	r.Use(gin.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.ErrorMiddleware())
	r.Use(middleware.CORS())

	// 设置超时中间件
	timeout := viper.GetInt("context.timeout")
	if timeout == 0 {
		log.Warn("timeout not configured, using default timeout")
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

	log.Infof("服务器启动在端口 %s", address)
	if err := r.Run(address); err != nil {
		log.Error("服务器启动失败:", err)
	}
}
