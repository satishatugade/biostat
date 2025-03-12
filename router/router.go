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
	Route{"diagnostic", http.MethodPost, constant.DiagnosticTests, diagnosticController.GetDiagnosticTests},
}

var medicationRepo = repository.NewMedicationRepository(database.GetDBConn())
var medicationService = service.NewMedicationService(medicationRepo)
var medicationController = controller.NewMedicationController(medicationService)
var medicationRoutes = Routes{
	Route{"Medication", http.MethodPost, constant.Medication, medicationController.GetAllMedication},
	Route{"Medication", http.MethodPost, constant.AddMedication, medicationController.AddMedication},
	Route{"Medication", http.MethodPut, constant.UpdateMedication, medicationController.UpdateMedication},
}
