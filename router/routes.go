package router

import (
	"biostat/auth"
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
	"/v1/diet":           {"admin", "patient"},
	constant.Medication:  {"admin", "doctor"},
	constant.PatientInfo: {"admin"},
}

func MasterRoutes(g *gin.RouterGroup, masterController *controller.MasterController, patientController *controller.PatientController) {
	master := g.Group("/master")
	for _, masterRoute := range getMasterRoutes(masterController) {
		switch masterRoute.Method {
		case http.MethodPost:
			master.POST(masterRoute.Path, masterRoute.HandleFunc)
		case http.MethodPut:
			master.PUT(masterRoute.Path, masterRoute.HandleFunc)
		}
	}
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
		handler := auth.ApplyMiddleware(diet.BasePath(), ProtectedRoutes, dietRoute.HandleFunc)
		switch dietRoute.Method {
		case http.MethodPost:
			diet.POST(dietRoute.Path, handler)
		case http.MethodGet:
			diet.GET(dietRoute.Path, dietRoute.HandleFunc)
		case http.MethodPut:
			diet.PUT(dietRoute.Path, dietRoute.HandleFunc)
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

func TblMedicalRecordsRoutes(g *gin.RouterGroup, tblMedicalRecordsController *controller.TblMedicalRecordController) {
	tblMedicalRecords := g.Group("/medical_records")
	for _, route := range getTblMedicalRecordsRoutes(tblMedicalRecordsController) {
		switch route.Method {
		case http.MethodPost:
			tblMedicalRecords.POST(route.Path, route.HandleFunc)
		case http.MethodGet:
			tblMedicalRecords.GET(route.Path, route.HandleFunc)
		case http.MethodPut:
			tblMedicalRecords.PUT(route.Path, route.HandleFunc)
		case http.MethodDelete:
			tblMedicalRecords.DELETE(route.Path, route.HandleFunc)
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
