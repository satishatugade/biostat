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
	Route{"patient", http.MethodPost, constant.SinglePatient, patientController.GetPatientByID},
	Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
	Route{"patient prescription", http.MethodPost, constant.PrescriptionByPatientId, patientController.GetPrescriptionByPatientID},
	Route{"Get prescription", http.MethodPost, constant.AllPrescription, patientController.GetAllPrescription},
	// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},

}

var diseaseRepo = repository.NewDiseaseRepository(database.GetDBConn())
var diseaseService = service.NewDiseaseService(diseaseRepo)

var causeRepo = repository.NewCauseRepository(database.GetDBConn())
var causeService = service.NewCauseService(causeRepo)

var symptomRepo = repository.NewSymptomRepository(database.GetDBConn())
var symptomService = service.NewSymptomService(symptomRepo)

var diseaseController = controller.NewDiseaseController(diseaseService, causeService, symptomService)
var diseaseRoutes = Routes{
	Route{"disease", http.MethodPost, constant.AddDisease, diseaseController.CreateDisease},
	Route{"disease", http.MethodPost, constant.Disease, diseaseController.GetDiseaseInfo},
	Route{"disease", http.MethodPost, constant.AllDisease, diseaseController.GetDiseaseInfo},
	Route{"disease", http.MethodPost, constant.DiseaseProfile, diseaseController.GetDiseaseProfile},
	Route{"disease", http.MethodPost, constant.Cause, diseaseController.GetAllCauses},
	Route{"disease", http.MethodPost, constant.AddCause, diseaseController.AddDiseaseCause},
	Route{"disease", http.MethodPut, constant.UpdateCause, diseaseController.UpdateDiseaseCause},

	Route{"disease", http.MethodPost, constant.Symptom, diseaseController.GetAllSymptom},
	Route{"disease", http.MethodPost, constant.AddSymptom, diseaseController.AddSymptom},
	Route{"disease", http.MethodPut, constant.UpdateSymptom, diseaseController.UpdateDiseaseSymptom},
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

var exerciseRepo = repository.NewExerciseRepository(database.GetDBConn())
var exerciseService = service.NewExerciseService(exerciseRepo)
var exerciseController = controller.NewExerciseController(exerciseService)
var exerciseRoutes = Routes{
	Route{"exercise", http.MethodPost, constant.Exercise, exerciseController.AddExercise},
	Route{"exercise", http.MethodPost, constant.AllExercise, exerciseController.GetAllExercises},
	Route{"exercise", http.MethodPost, constant.SingleExercise, exerciseController.GetExerciseByID},
	Route{"exercise", http.MethodPut, constant.UpdateExercise, exerciseController.UpdateExercise},
}

var dietRepo = repository.NewDietRepository(database.GetDBConn())
var dietService = service.NewDietService(dietRepo)
var dietController = controller.NewDietController(dietService)
var dietRoutes = Routes{
	Route{"diet", http.MethodPost, constant.Diet, dietController.AddDietPlanTemplate},
	Route{"diet", http.MethodPost, constant.AllDietTemplate, dietController.GetAllDietPlanTemplates},
	Route{"diet", http.MethodPost, constant.SingleDiet, dietController.GetDietPlanById},
	Route{"diet", http.MethodPut, constant.UpdateDiet, dietController.UpdateDietPlanTemplate},
}
