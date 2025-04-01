package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/database"
	"biostat/repository"
	"biostat/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitializeRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	var patientRepo = repository.NewPatientRepository(db)
	var patientService = service.NewPatientService(patientRepo)
	var dietRepo = repository.NewDietRepository(database.GetDBConn())
	var dietService = service.NewDietService(dietRepo)
	var patientController = controller.NewPatientController(patientService, dietService)

	PatientRoutes(apiGroup, patientController)

	var diseaseRepo = repository.NewDiseaseRepository(database.GetDBConn())
	var diseaseService = service.NewDiseaseService(diseaseRepo)
	var causeRepo = repository.NewCauseRepository(database.GetDBConn())
	var causeService = service.NewCauseService(causeRepo)
	var symptomRepo = repository.NewSymptomRepository(database.GetDBConn())
	var symptomService = service.NewSymptomService(symptomRepo)
	var diseaseController = controller.NewDiseaseController(diseaseService, causeService, symptomService)

	DiseaseRoutes(apiGroup, diseaseController)

	var diagnosticRepo = repository.NewDiagnosticRepository(database.GetDBConn())
	var diagnosticService = service.NewDiagnosticService(diagnosticRepo)
	var diagnosticController = controller.NewDiagnosticController(diagnosticService)

	DiagnosticRoutes(apiGroup, diagnosticController)

	var medicationRepo = repository.NewMedicationRepository(database.GetDBConn())
	var medicationService = service.NewMedicationService(medicationRepo)
	var medicationController = controller.NewMedicationController(medicationService)

	MedicationRoutes(apiGroup, medicationController)

	var exerciseRepo = repository.NewExerciseRepository(database.GetDBConn())
	var exerciseService = service.NewExerciseService(exerciseRepo)
	var exerciseController = controller.NewExerciseController(exerciseService)
	ExerciseRoutes(apiGroup, exerciseController)

	// var dietRepo = repository.NewDietRepository(database.GetDBConn())
	// var dietService = service.NewDietService(dietRepo)
	var dietController = controller.NewDietController(dietService)
	DietRoutes(apiGroup, dietController)

	var tblMedicalRecordsRepo = repository.NewTblMedicalRecordRepository(db)
	var tblMedicalRecordsService = service.NewTblMedicalRecordService(tblMedicalRecordsRepo)
	var tblMedicalRecordsController = controller.NewTblMedicalRecordController(tblMedicalRecordsService)

	TblMedicalRecordsRoutes(apiGroup, tblMedicalRecordsController)

	var tblUserGtokenRepo = repository.NewTblUserGtokenRepository(db)
	var tblUserGtokenService = service.NewTblUserGtokenService(tblUserGtokenRepo)
	var gmailRecordsController = controller.NewGmailSyncController(tblMedicalRecordsService, tblUserGtokenService)

	GmailSyncRoutes(apiGroup, gmailRecordsController)

}

func getPatientRoutes(patientController *controller.PatientController) Routes {
	return Routes{
		Route{"patient", http.MethodPost, constant.PatientInfo, patientController.GetPatientInfo},
		Route{"patient", http.MethodPost, constant.SinglePatient, patientController.GetPatientByID},
		Route{"patient", http.MethodPut, constant.UpdatePatient, patientController.UpdatePatientInfoById},
		Route{"patient", http.MethodPost, constant.PatientRelative, patientController.AddPatientRelative},
		Route{"patient", http.MethodPost, constant.GetRelative, patientController.GetPatientRelative},
		Route{"patient", http.MethodPut, constant.UpdateRealtiveInfo, patientController.UpdatePatientRelative},
		Route{"patient disease condition", http.MethodPost, constant.PatientDiseaseCondition, patientController.GetPatientDiseaseProfiles},
		Route{"patient diet", http.MethodPost, constant.PatientDietPlan, patientController.GetPatientDietPlan},
		Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
		Route{"patient prescription", http.MethodPost, constant.PrescriptionByPatientId, patientController.GetPrescriptionByPatientID},
		Route{"Get prescription", http.MethodPost, constant.AllPrescription, patientController.GetAllPrescription},
		// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},
	}
}

func getDiseaseRoutes(diseaseController *controller.DiseaseController) Routes {
	return Routes{
		Route{"disease", http.MethodPost, constant.AddDisease, diseaseController.CreateDisease},
		Route{"disease", http.MethodPost, constant.Disease, diseaseController.GetDiseaseInfo},
		Route{"disease", http.MethodPost, constant.AllDisease, diseaseController.GetDiseaseInfo},
		Route{"disease", http.MethodPost, constant.DiseaseProfile, diseaseController.GetDiseaseProfile},
		Route{"disease", http.MethodPost, constant.SingleDiseaseProfile, diseaseController.GetDiseaseProfileById},
		Route{"disease", http.MethodPost, constant.Cause, diseaseController.GetAllCauses},
		Route{"disease", http.MethodPost, constant.AddCause, diseaseController.AddDiseaseCause},
		Route{"disease", http.MethodPut, constant.UpdateCause, diseaseController.UpdateDiseaseCause},

		Route{"disease", http.MethodPost, constant.Symptom, diseaseController.GetAllSymptom},
		Route{"disease", http.MethodPost, constant.AddSymptom, diseaseController.AddSymptom},
		Route{"disease", http.MethodPut, constant.UpdateSymptom, diseaseController.UpdateDiseaseSymptom},
	}
}

func getDiagnosticRoutes(diagnosticController *controller.DiagnosticController) Routes {
	return Routes{
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
}

func getMedicationRoutes(medicationController *controller.MedicationController) Routes {
	return Routes{
		Route{"Medication", http.MethodPost, constant.Medication, medicationController.GetAllMedication},
		Route{"Medication", http.MethodPost, constant.AddMedication, medicationController.AddMedication},
		Route{"Medication", http.MethodPut, constant.UpdateMedication, medicationController.UpdateMedication},
	}
}

func getExerciseRoutes(exerciseController *controller.ExerciseController) Routes {
	return Routes{
		Route{"exercise", http.MethodPost, constant.Exercise, exerciseController.AddExercise},
		Route{"exercise", http.MethodPost, constant.AllExercise, exerciseController.GetAllExercises},
		Route{"exercise", http.MethodPost, constant.SingleExercise, exerciseController.GetExerciseByID},
		Route{"exercise", http.MethodPut, constant.UpdateExercise, exerciseController.UpdateExercise},
	}
}

func getDietRoutes(dietController *controller.DietController) Routes {
	return Routes{
		Route{"diet", http.MethodPost, constant.Diet, dietController.AddDietPlanTemplate},
		Route{"diet", http.MethodPost, constant.AllDietTemplate, dietController.GetAllDietPlanTemplates},
		Route{"diet", http.MethodPost, constant.SingleDiet, dietController.GetDietPlanById},
		Route{"diet", http.MethodPut, constant.UpdateDiet, dietController.UpdateDietPlanTemplate},
	}
}

func getTblMedicalRecordsRoutes(tblMedicalRecordsController *controller.TblMedicalRecordController) Routes {
	return Routes{
		{"medical records create", http.MethodPost, "/medical_records", tblMedicalRecordsController.CreateTblMedicalRecord},
		{"medical records get", http.MethodGet, "/medical_records", tblMedicalRecordsController.GetAllTblMedicalRecords},
		{"medical records get", http.MethodPost, "/medical_records/:user_id", tblMedicalRecordsController.GetUserMedicalRecords},
		{"medical records get single", http.MethodGet, "/medical_records/:id", tblMedicalRecordsController.GetSingleTblMedicalRecord},
		{"medical records update", http.MethodPut, "/medical_records/:id", tblMedicalRecordsController.UpdateTblMedicalRecord},
		{"medical records delete", http.MethodDelete, "/medical_records/:id", tblMedicalRecordsController.DeleteTblMedicalRecord},
	}
}

func getMailSyncRoutes(gmailSyncController *controller.GmailSyncController) Routes {
	return Routes{
		{"gmail sync route", http.MethodGet, "/inbox/:user_id", gmailSyncController.FetchEmailsHandler},
		{"gmail sync route", http.MethodGet, "/oauth2callback", gmailSyncController.GmailCallbackHandler},
		{"gmail sync route", http.MethodGet, "/login", controller.GmailLoginHandler},
	}
}
