package database

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// DB 是 MySQL 数据库连接
	DB *gorm.DB

	// RedisClient 是 Redis 客户端
	RedisClient *redis.Client

	// Config 是 Viper 配置
	Config *viper.Viper
)

func Init() {
	// 初始化 Viper 配置
	initConfig()

	// 初始化 MySQL 数据库连接
	initMySQL()
	if DB == nil {
		log.Fatal("数据库连接初始化失败")
	}
	// 初始化 Redis 客户端
	initRedis()
}

func initConfig() {
	// 初始化 Viper
	Config = viper.New()
	Config.SetConfigFile("./global/config/database.yaml")

	// 读取配置文件
	if err := Config.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// 设置默认值
	Config.SetDefault("mysql.dsn", "user:password@tcp(127.0.0.1:3306)/dbname")
	Config.SetDefault("redis.addr", "localhost:6379")
	Config.SetDefault("redis.password", "")
	Config.SetDefault("redis.db", 0)
}

func initMySQL() {
	// 从配置文件中获取 MySQL DSN
	dsn := Config.GetString("mysql.dsn")

	// 连接 MySQL 数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %s", err)
	}

	// 获取底层的数据库连接并测试连接
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error getting underlying database connection: %s", err)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Error pinging MySQL: %s", err)
	}

	log.Println("MySQL connected successfully")
	DB.AutoMigrate(&User{})
}
func initRedis() {
	// 从配置文件中获取 Redis 配置
	redisAddr := Config.GetString("redis.addr")
	redisPassword := Config.GetString("redis.password")
	redisDB := Config.GetInt("redis.db")

	// 初始化 Redis 客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %s", err)
	}

	log.Println("Redis connected successfully")
}
