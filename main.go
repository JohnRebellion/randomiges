package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"randomiges/envRouting"
	"randomiges/image"
	"randomiges/imageRepo"
	"randomiges/mailService"
	"randomiges/user"

	"github.com/JohnRebellion/go-utils/database"
	fiberUtils "github.com/JohnRebellion/go-utils/fiber"
	"github.com/JohnRebellion/go-utils/passwordHashing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	envRouting.LoadEnv()
	databaseConnect()
	mailService.Init(mailService.Options{
		From:           mailService.MailUser{Name: "Randomiges", Email: envRouting.MailEmail},
		SendGridAPIKey: envRouting.SendGridAPIKey,
	})
	err := imageRepo.Init()

	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(logger.New())
	setupRoutes(app)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", envRouting.Port)))
}

func databaseConnect() {
	database.SQLiteConnect(envRouting.SQLiteFilename)
	database.DBConn.AutoMigrate(&user.User{}, &image.Image{})
	var existingUser user.User
	database.DBConn.First(&existingUser, 1)
	password, _ := passwordHashing.HashPassword("12345678")

	if existingUser.ID == 0 {
		database.DBConn.Create(&user.User{
			Username:      "johnnecirrebellion@gmail.com",
			Password:      password,
			Name:          "John Necir Rebellion",
			Role:          "Admin",
			ContactNumber: "+639292946498",
			Email:         "johnnecirrebellion@gmail.com",
		})
	}

	if database.Err != nil {
		panic(database.Err)
	}
}

func setupRoutes(app *fiber.App) {
	app.Static("/", envRouting.StaticWebLocation)

	apiEndpoint := app.Group("/api")
	v1Endpoint := apiEndpoint.Group("/v1")

	userEndpoint := v1Endpoint.Group("/user")
	userEndpoint.Post("/", user.NewUser)

	userAuthEndpoint := userEndpoint.Group("/auth")
	userAuthEndpoint.Post("/", user.Authenticate)
	userAuthEndpoint.Post("/sendOTPBy/:by", user.SendOTPBy)
	userAuthEndpoint.Post("/2fa", user.AuthenticatePIN)
	userAuthEndpoint.Post("/sendResetRequest/:by", user.SendResetRequest)
	userAuthEndpoint.Get("/approveRequest", user.ResetPasswordApprove)
	userAuthEndpoint.Post("/resetPassword", user.ResetPassword)

	app.Use(fiberUtils.AuthenticationMiddleware(fiberUtils.JWTConfig{
		Duration:     24 * time.Hour,
		CookieMaxAge: 24 * 60 * 60,
		SetCookies:   true,
		SecretKey:    []byte(envRouting.SecretKey),
	}))
	userEndpoint.Get("/", user.GetUsers)
	userEndpoint.Put("/", user.UpdateUser)

	userEndpoint.Delete("/:id", user.DeleteUser)
	userEndpoint.Get("/:id", user.GetUser)

	imagesEndpoint := v1Endpoint.Group("/images")
	imagesEndpoint.Get("/", image.GetImages)
	imagesEndpoint.Get("/all", image.AllImages)
	imagesEndpoint.Post("/", image.NewImage)

	imagesEndpoint.Patch("/:id", image.UpdateImage)
	imagesEndpoint.Delete("/:id", image.DeleteImage)
	imagesEndpoint.Get("/:id", image.GetImage)
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModeDir|0755)
	}

	return nil
}
