package router

import (
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

func DiseaseRoutes(g *gin.RouterGroup, diseaseController *controller.DiseaseController) {
	disease := g.Group("/disease")
	for _, diseaseRoute := range getDiseaseRoutes(diseaseController) {
		switch diseaseRoute.Method {
		case http.MethodPost:
			disease.POST(diseaseRoute.Path, diseaseRoute.HandleFunc)
		case http.MethodGet:
			disease.GET(diseaseRoute.Path, diseaseRoute.HandleFunc)
		case http.MethodPut:
			disease.PUT(diseaseRoute.Path, diseaseRoute.HandleFunc)
		}
	}
}

func DiagnosticRoutes(g *gin.RouterGroup, diagnosticController *controller.DiagnosticController) {
	diagnostic := g.Group("/diagnostic")
	for _, diagnosticRoute := range getDiagnosticRoutes(diagnosticController) {
		switch diagnosticRoute.Method {
		case http.MethodPost:
			diagnostic.POST(diagnosticRoute.Path, diagnosticRoute.HandleFunc)
		case http.MethodPut:
			diagnostic.PUT(diagnosticRoute.Path, diagnosticRoute.HandleFunc)
		case http.MethodGet:
			diagnostic.GET(diagnosticRoute.Path, diagnosticRoute.HandleFunc)
		case http.MethodDelete:
			diagnostic.DELETE(diagnosticRoute.Path, diagnosticRoute.HandleFunc)
		}
	}
}

func MedicationRoutes(g *gin.RouterGroup, medicationController *controller.MedicationController) {
	medication := g.Group("/medication")
	for _, medicationRoute := range getMedicationRoutes(medicationController) {
		switch medicationRoute.Method {
		case http.MethodPost:
			medication.POST(medicationRoute.Path, medicationRoute.HandleFunc)
		case http.MethodGet:
			medication.GET(medicationRoute.Path, medicationRoute.HandleFunc)
		case http.MethodPut:
			medication.PUT(medicationRoute.Path, medicationRoute.HandleFunc)
		}
	}
}

func ExerciseRoutes(g *gin.RouterGroup, exerciseController *controller.ExerciseController) {
	exercise := g.Group("/exercise")
	for _, exerciseRoute := range getExerciseRoutes(exerciseController) {
		switch exerciseRoute.Method {
		case http.MethodPost:
			exercise.POST(exerciseRoute.Path, exerciseRoute.HandleFunc)
		case http.MethodGet:
			exercise.GET(exerciseRoute.Path, exerciseRoute.HandleFunc)
		case http.MethodPut:
			exercise.PUT(exerciseRoute.Path, exerciseRoute.HandleFunc)
		}
	}
}

func DietRoutes(g *gin.RouterGroup, dietController *controller.DietController) {
	diet := g.Group("/diet")
	for _, dietRoute := range getDietRoutes(dietController) {
		switch dietRoute.Method {
		case http.MethodPost:
			diet.POST(dietRoute.Path, dietRoute.HandleFunc)
		case http.MethodGet:
			diet.GET(dietRoute.Path, dietRoute.HandleFunc)
		case http.MethodPut:
			diet.PUT(dietRoute.Path, dietRoute.HandleFunc)
		}
	}
}

func Routing() {
	r := routes{
		router: gin.Default(),
	}
	r.router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control"},
	}))
	apiGroup := r.router.Group(os.Getenv("ApiVersion"))
	db := database.GetDBConn()
	InitializeRoutes(apiGroup, db)
	r.router.Run(":" + os.Getenv("GO_SERVER_PORT"))
}
