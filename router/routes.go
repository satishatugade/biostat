package router

import (
	"biostat/auth"
	"biostat/controller"
	"biostat/database"
	"log"
	"net/http"
	"os"
	"strings"

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
	"/v1/master":  {"admin"},
	"/v1/patient": {"admin", "patient"},
	"/v1/user":    {"admin", "patient", "relative", "caregiver"},
}

func MasterRoutes(g *gin.RouterGroup, masterController *controller.MasterController, patientController *controller.PatientController) {
	master := g.Group("/master")
	for _, masterRoute := range getMasterRoutes(masterController) {
		protectedHandler := auth.Authenticate(master.BasePath(), ProtectedRoutes, masterRoute.HandleFunc)
		switch masterRoute.Method {
		case http.MethodGet:
			master.GET(masterRoute.Path, protectedHandler)
		case http.MethodPost:
			master.POST(masterRoute.Path, protectedHandler)
		case http.MethodPut:
			master.PUT(masterRoute.Path, protectedHandler)
		case http.MethodDelete:
			master.DELETE(masterRoute.Path, protectedHandler)
		}
	}
}

func PatientRoutes(g *gin.RouterGroup, patientController *controller.PatientController) {

	patient := g.Group("/patient")
	for _, patientRoute := range getPatientRoutes(patientController) {
		protectedHandler := auth.Authenticate(patient.BasePath(), ProtectedRoutes, patientRoute.HandleFunc)
		switch patientRoute.Method {
		case http.MethodGet:
			patient.GET(patientRoute.Path, protectedHandler)
		case http.MethodPost:
			patient.POST(patientRoute.Path, protectedHandler)
		case http.MethodPut:
			patient.PUT(patientRoute.Path, protectedHandler)
		}
	}
}

func UserRoutes(g *gin.RouterGroup, userController *controller.UserController) {
	user := g.Group("/user")
	for _, userRoute := range getUserRoutes(userController) {
		// protectedHandler := auth.Authenticate(user.BasePath(), ProtectedRoutes, userRoute.HandleFunc)
		switch userRoute.Method {
		case http.MethodPost:
			user.POST(userRoute.Path, userRoute.HandleFunc)
		case http.MethodGet:
			user.GET(userRoute.Path, userRoute.HandleFunc)
		}
	}
}

func GmailSyncRoutes(g *gin.RouterGroup, gmailSyncController *controller.GmailSyncController) {
	gmail := g.Group("/mail")
	for _, gmailroute := range getMailSyncRoutes(gmailSyncController) {
		// protectedHandler := auth.Authenticate(gmail.BasePath(), ProtectedRoutes, gmailroute.HandleFunc)
		switch gmailroute.Method {
		case http.MethodPost:
			gmail.POST(gmailroute.Path, gmailroute.HandleFunc)
		case http.MethodGet:
			gmail.GET(gmailroute.Path, gmailroute.HandleFunc)
		case http.MethodPut:
			gmail.PUT(gmailroute.Path, gmailroute.HandleFunc)
		case http.MethodDelete:
			gmail.DELETE(gmailroute.Path, gmailroute.HandleFunc)
		}

	}
}

func Routing() {
	r := routes{
		router: gin.Default(),
	}
	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	log.Println("corsOrigins", corsOrigins)
	r.router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control"},
		AllowCredentials: true,
	}))
	r.router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Biostat server running..."})
	})
	apiGroup := r.router.Group(os.Getenv("ApiVersion"))
	db := database.GetDBConn()
	InitializeRoutes(apiGroup, db)
	r.router.Run(":" + os.Getenv("GO_SERVER_PORT"))
}
