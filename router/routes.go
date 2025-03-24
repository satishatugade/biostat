package router

import (
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

func (r routes) Patient(g *gin.RouterGroup) {
	patient := g.Group("/patient")
	for _, patientRoute := range patientRoutes {
		switch patientRoute.Method {
		case http.MethodPost:
			patient.POST(patientRoute.Path, patientRoute.HandleFunc)
		case http.MethodPut:
			patient.PUT(patientRoute.Path, patientRoute.HandleFunc)
		}
	}
}

func (r routes) Disease(g *gin.RouterGroup) {
	disease := g.Group("/disease")
	for _, diseaseRoute := range diseaseRoutes {
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

func (r routes) Diagnostic(g *gin.RouterGroup) {
	diagnostic := g.Group("/diagnostic")
	for _, diagnosticRoute := range diagnosticRoutes {
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

func (r routes) Medication(g *gin.RouterGroup) {
	medication := g.Group("/medication")
	for _, medicationRoute := range medicationRoutes {
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

func (r routes) Exercise(g *gin.RouterGroup) {
	exercise := g.Group("/exercise")
	for _, exerciseRoute := range exerciseRoutes {
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

func (r routes) Diet(g *gin.RouterGroup) {
	diet := g.Group("/diet")
	for _, dietRoute := range dietRoutes {
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
	r.Patient(apiGroup)
	r.Disease(apiGroup)
	r.Diagnostic(apiGroup)
	r.Medication(apiGroup)
	r.Exercise(apiGroup)
	r.Diet(apiGroup)
	r.router.Run(":" + os.Getenv("GO_SERVER_PORT"))
}
