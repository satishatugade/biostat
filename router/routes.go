package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/database"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Name       string
	Method     string
	Path       string
	HandleFunc func(*gin.Context)
}

type routes struct {
	router *gin.Engine
}

type Routes []Route

var ProtectedRoutes = map[string][]string{
	"/v1/master":         {"admin"},
	constant.Medication:  {"admin", "doctor"},
	constant.PatientInfo: {"patient", "doctor"},
}

func MasterRoutes(g *gin.RouterGroup, masterController *controller.MasterController, patientController *controller.PatientController) {
	master := g.Group("/master")
	for _, masterRoute := range getMasterRoutes(masterController) {
		switch masterRoute.Method {
		case http.MethodGet:
			master.GET(masterRoute.Path, masterRoute.HandleFunc)
		case http.MethodPost:
			master.POST(masterRoute.Path, masterRoute.HandleFunc)
		case http.MethodPut:
			master.PUT(masterRoute.Path, masterRoute.HandleFunc)
		case http.MethodDelete:
			master.DELETE(masterRoute.Path, masterRoute.HandleFunc)
		}
	}
}

func PatientRoutes(g *gin.RouterGroup, patientController *controller.PatientController) {

	patient := g.Group("/patient")
	for _, patientRoute := range getPatientRoutes(patientController) {
		switch patientRoute.Method {
		case http.MethodPost:
			patient.POST(patientRoute.Path, patientRoute.HandleFunc)
		case http.MethodPut:
			patient.PUT(patientRoute.Path, patientRoute.HandleFunc)
		}
	}
}

func UserRoutes(g *gin.RouterGroup, userController *controller.UserController) {
	user := g.Group("/user")
	for _, userRoute := range getUserRoutes(userController) {
		switch userRoute.Method {
		case http.MethodPost:
			user.POST(userRoute.Path, userRoute.HandleFunc)
		case http.MethodGet:
			user.GET(userRoute.Path, userRoute.HandleFunc)
		}
	}
}

func GmailSyncRoutes(g *gin.RouterGroup, gmailSyncController *controller.GmailSyncController) {
	gmailRoutGroup := g.Group("/mail")
	for _, route := range getMailSyncRoutes(gmailSyncController) {
		switch route.Method {
		case http.MethodPost:
			gmailRoutGroup.POST(route.Path, route.HandleFunc)
		case http.MethodGet:
			gmailRoutGroup.GET(route.Path, route.HandleFunc)
		case http.MethodPut:
			gmailRoutGroup.PUT(route.Path, route.HandleFunc)
		case http.MethodDelete:
			gmailRoutGroup.DELETE(route.Path, route.HandleFunc)
		}

	}
}

func Routing() {
	r := routes{
		router: gin.Default(),
	}
	r.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control"},
		AllowCredentials: true,
	}))
	apiGroup := r.router.Group(os.Getenv("ApiVersion"))
	db := database.GetDBConn()
	InitializeRoutes(apiGroup, db)
	r.router.Run(":" + os.Getenv("GO_SERVER_PORT"))
}
