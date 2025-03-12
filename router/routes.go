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
		}
	}
}

func (r routes) Diagnostic(g *gin.RouterGroup) {
	diagnostic := g.Group("/diagnostic")
	for _, diagnosticRoute := range diagnosticRoutes {
		switch diagnosticRoute.Method {
		case http.MethodPost:
			diagnostic.POST(diagnosticRoute.Path, diagnosticRoute.HandleFunc)
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
	r.router.Run(":" + os.Getenv("GO_SERVER_PORT"))
}
