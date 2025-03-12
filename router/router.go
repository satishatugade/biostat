package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/database"
	"biostat/repository"
	"biostat/service"
	"net/http"
)

var patientRepo = repository.NewPatientRepository(database.GetDBConn())
var patientService = service.NewPatientService(patientRepo)
var patientController = controller.NewPatientController(patientService)
var patientRoutes = Routes{
	Route{"patient", http.MethodPost, constant.PatientInfo, patientController.GetPatientInfo},
	Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
	// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},
}

var diseaseRepo = repository.NewDiseaseRepository(database.GetDBConn())
var diseaseService = service.NewDiseaseService(diseaseRepo)

var causeRepo = repository.NewCauseRepository(database.GetDBConn())
var causeService = service.NewCauseService(causeRepo)

var diseaseController = controller.NewDiseaseController(diseaseService, causeService)
var diseaseRoutes = Routes{
	Route{"disease", http.MethodPost, constant.Disease, diseaseController.GetDiseaseInfo},
	Route{"disease", http.MethodPost, constant.DiseaseProfile, diseaseController.GetDiseaseProfile},
	Route{"disease", http.MethodPost, constant.Cause, diseaseController.GetAllCauses},
	Route{"disease", http.MethodPost, constant.AddCause, diseaseController.AddDiseaseCause},
	Route{"disease", http.MethodPut, constant.UpdateCause, diseaseController.UpdateDiseaseCause},
}

var diagnosticRepo = repository.NewDiagnosticRepository(database.GetDBConn())
var diagnosticService = service.NewDiagnosticService(diagnosticRepo)
var diagnosticController = controller.NewDiagnosticController(diagnosticService)
var diagnosticRoutes = Routes{
	// Diagnostic Test Routes
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTests, diagnosticController.GetDiagnosticTests},
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTest, diagnosticController.CreateDiagnosticTest},
	Route{"diagnostic", http.MethodPut, constant.DiagnosticTest, diagnosticController.UpdateDiagnosticTest},
	Route{"diagnostic", http.MethodGet, constant.SingleDiagnosticTest, diagnosticController.GetSingleDiagnosticTest},
	Route{"diagnostic", http.MethodDelete, constant.SingleDiagnosticTest, diagnosticController.DeleteDiagnosticTest},
	// Diagnostic Component Routes
	Route{"diagnostic", http.MethodPost, constant.DiagnosticComponents, diagnosticController.GetAllDiagnosticComponents},
	Route{"diagnostic", http.MethodPost, constant.DiagnosticComponent, diagnosticController.CreateDiagnosticComponent},
	Route{"diagnostic", http.MethodPut, constant.DiagnosticComponent, diagnosticController.UpdateDiagnosticComponent},
	Route{"diagnostic", http.MethodGet, constant.SingleDiagnosticComponent, diagnosticController.GetSingleDiagnosticComponent},

	// Diagnostic Test Component Mapping Routes
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTestComponentMapping, diagnosticController.CreateDiagnosticTestComponentMapping},
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTestComponentMappings, diagnosticController.GetAllDiagnosticTestComponentMappings},
	Route{"diagnostic", http.MethodPut, constant.DiagnosticTestComponentMapping, diagnosticController.UpdateDiagnosticTestComponentMapping},
}

var medicationRepo = repository.NewMedicationRepository(database.GetDBConn())
var medicationService = service.NewMedicationService(medicationRepo)
var medicationController = controller.NewMedicationController(medicationService)
var medicationRoutes = Routes{
	Route{"Medication", http.MethodPost, constant.Medication, medicationController.GetAllMedication},
	Route{"Medication", http.MethodPost, constant.AddMedication, medicationController.AddMedication},
	Route{"Medication", http.MethodPut, constant.UpdateMedication, medicationController.UpdateMedication},
}
