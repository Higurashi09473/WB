package internal

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

type App struct {
	httpServer *fiber.App
	database   *mongo.Database

	quizService *service.QuizService
	netService  *service.NetService
}

func (a *App) Init() {
	a.setupDb()
	a.setupServices()
	a.setupHttp()

	log.Fatal(a.httpServer.Listen(":3000"))
}

func (a *App) setupHttp() {
	app := fiber.New()
	app.Use(cors.New())

	quizController := controller.Quiz(a.quizService)
	app.Get("/api/quizzes", quizController.GetQuizzes)
	

	wsController := controller.Ws(a.netService)
	app.Get("/ws", websocket.New(wsController.Ws))

	a.httpServer = app
}


func (a *App) setupDb() {
	// client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://root:Star09473@localhost:27017/?authSource=admin"))

	if err != nil {
		panic(err)
	}

	a.database = client.Database("quiz")
}