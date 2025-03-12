package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/database"
	"biostat/repository"
	"biostat/service"
	"net/http"
)

var patientRepo = repository.NewPatientRepository(database.InitDB())
var patientService = service.NewPatientService(patientRepo)
var patientController = controller.NewPatientController(patientService)
var patientRoutes = Routes{
	Route{"patient", http.MethodPost, constant.PatientInfo, patientController.GetPatientInfo},
	Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
	// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},
}

var diseaseRepo = repository.NewDiseaseRepository(database.InitDB())
var diseaseService = service.NewDiseaseService(diseaseRepo)
var diseaseController = controller.NewDiseaseController(diseaseService)
var diseaseRoutes = Routes{
	Route{"disease", http.MethodPost, constant.Disease, diseaseController.GetDiseaseInfo},
	Route{"disease", http.MethodPost, constant.DiseaseProfile, diseaseController.GetDiseaseProfile},
}

var diagnosticRepo = repository.NewDiagnosticRepository(database.InitDB())
var diagnosticService = service.NewDiagnosticService(diagnosticRepo)
var diagnosticController = controller.NewDiagnosticController(diagnosticService)
var diagnosticRoutes = Routes{
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTests, diagnosticController.GetDiagnosticTests},	
}