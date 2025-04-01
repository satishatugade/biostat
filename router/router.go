package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/repository"
	"biostat/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitializeRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {

	var allergyRepo = repository.NewAllergyRepository(db)
	var allergyService = service.NewAllergyService(allergyRepo)

	var diseaseRepo = repository.NewDiseaseRepository(db)
	var diseaseService = service.NewDiseaseService(diseaseRepo)

	var causeRepo = repository.NewCauseRepository(db)
	var causeService = service.NewCauseService(causeRepo)

	var symptomRepo = repository.NewSymptomRepository(db)
	var symptomService = service.NewSymptomService(symptomRepo)

	var medicationRepo = repository.NewMedicationRepository(db)
	var medicationService = service.NewMedicationService(medicationRepo)

	var dietRepo = repository.NewDietRepository(db)
	var dietService = service.NewDietService(dietRepo)

	var exerciseRepo = repository.NewExerciseRepository(db)
	var exerciseService = service.NewExerciseService(exerciseRepo)

	var diagnosticRepo = repository.NewDiagnosticRepository(db)
	var diagnosticService = service.NewDiagnosticService(diagnosticRepo)

	var medicalRecordsRepo = repository.NewTblMedicalRecordRepository(db)
	var medicalRecordService = service.NewTblMedicalRecordService(medicalRecordsRepo)

	var patientRepo = repository.NewPatientRepository(db)
	var patientService = service.NewPatientService(patientRepo)

	var patientController = controller.NewPatientController(patientService, dietService, allergyService, medicalRecordService)

	var masterController = controller.NewMasterController(allergyService, diseaseService, causeService, symptomService, medicationService, dietService, exerciseService, diagnosticService)
	MasterRoutes(apiGroup, masterController, patientController)
	PatientRoutes(apiGroup, patientController)

	var userController = controller.NewUserController()
	UserRoutes(apiGroup, userController)

	var tblUserGtokenRepo = repository.NewTblUserGtokenRepository(db)
	var tblUserGtokenService = service.NewTblUserGtokenService(tblUserGtokenRepo)
	var gmailRecordsController = controller.NewGmailSyncController(medicalRecordService, tblUserGtokenService)

	GmailSyncRoutes(apiGroup, gmailRecordsController)

}

func getMasterRoutes(masterController *controller.MasterController) Routes {
	return Routes{
		//disease master
		Route{"Disease", http.MethodPost, constant.AddDisease, masterController.CreateDisease},
		Route{"Disease", http.MethodPost, constant.Disease, masterController.GetDiseaseInfo},
		Route{"Disease", http.MethodPost, constant.AllDisease, masterController.GetDiseaseInfo},
		Route{"Disease", http.MethodPost, constant.DiseaseProfile, masterController.GetDiseaseProfile},
		Route{"Disease", http.MethodPost, constant.SingleDiseaseProfile, masterController.GetDiseaseProfileById},

		//Causes master
		Route{"Causes", http.MethodPost, constant.Cause, masterController.GetAllCauses},
		Route{"Causes", http.MethodPost, constant.AddCause, masterController.AddDiseaseCause},
		Route{"Causes", http.MethodPut, constant.UpdateCause, masterController.UpdateDiseaseCause},

		//symptoms master
		Route{"Symptom", http.MethodPost, constant.Symptom, masterController.GetAllSymptom},
		Route{"Symptom", http.MethodPost, constant.AddSymptom, masterController.AddSymptom},
		Route{"Symptom", http.MethodPut, constant.UpdateSymptom, masterController.UpdateDiseaseSymptom},

		// Allergy master
		Route{"Allergy", http.MethodPost, constant.AllergyMaster, masterController.GetAllergyRestrictions},

		//Medication master
		Route{"Medication", http.MethodPost, constant.Medication, masterController.GetAllMedication},
		Route{"Medication", http.MethodPost, constant.AddMedication, masterController.AddMedication},
		Route{"Medication", http.MethodPut, constant.UpdateMedication, masterController.UpdateMedication},

		//exercise master
		Route{"Exercise", http.MethodPost, constant.Exercise, masterController.AddExercise},
		Route{"Exercise", http.MethodPost, constant.AllExercise, masterController.GetAllExercises},
		Route{"Exercise", http.MethodPost, constant.SingleExercise, masterController.GetExerciseByID},
		Route{"Exercise", http.MethodPut, constant.UpdateExercise, masterController.UpdateExercise},

		//diet master
		Route{"Diet", http.MethodPost, constant.Diet, masterController.AddDietPlanTemplate},
		Route{"Diet", http.MethodPost, constant.AllDietTemplate, masterController.GetAllDietPlanTemplates},
		Route{"Diet", http.MethodPost, constant.SingleDiet, masterController.GetDietPlanById},
		Route{"Diet", http.MethodPut, constant.UpdateDiet, masterController.UpdateDietPlanTemplate},

		// Diagnostic Test Routes
		Route{"DTM", http.MethodPost, constant.DiagnosticTests, masterController.GetDiagnosticTests},
		Route{"DTM", http.MethodPost, constant.DiagnosticTest, masterController.CreateDiagnosticTest},
		Route{"DTM", http.MethodPut, constant.DiagnosticTest, masterController.UpdateDiagnosticTest},
		Route{"DTM", http.MethodGet, constant.SingleDiagnosticTest, masterController.GetSingleDiagnosticTest},
		Route{"DTM", http.MethodDelete, constant.SingleDiagnosticTest, masterController.DeleteDiagnosticTest},
		// Diagnostic Component Routes
		Route{"DTM", http.MethodPost, constant.DiagnosticComponents, masterController.GetAllDiagnosticComponents},
		Route{"DTM", http.MethodPost, constant.DiagnosticComponent, masterController.CreateDiagnosticComponent},
		Route{"DTM", http.MethodPut, constant.DiagnosticComponent, masterController.UpdateDiagnosticComponent},
		Route{"DTM", http.MethodGet, constant.SingleDiagnosticComponent, masterController.GetSingleDiagnosticComponent},

		// Diagnostic Test Component Mapping Routes
		Route{"DTM", http.MethodPost, constant.DiagnosticTestComponentMapping, masterController.CreateDiagnosticTestComponentMapping},
		Route{"DTM", http.MethodPost, constant.DiagnosticTestComponentMappings, masterController.GetAllDiagnosticTestComponentMappings},
		Route{"DTM", http.MethodPut, constant.DiagnosticTestComponentMapping, masterController.UpdateDiagnosticTestComponentMapping},
	}
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
		Route{"Patient Allergy", http.MethodPost, constant.PatientAllergy, patientController.AddPatientAllergyRestriction},
		Route{"Patient Allgery", http.MethodPost, constant.Allergy, patientController.GetPatientAllergyRestriction},
		Route{"Update Patient Allgery", http.MethodPut, constant.UpdateAllergy, patientController.UpdatePatientAllergyRestriction},
		Route{"Add Custom Range", http.MethodPost, constant.CustomRange, patientController.AddPatientClinicalRange},
		// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},

		{"medical records create", http.MethodPost, "/medical_records", patientController.CreateTblMedicalRecord},
		{"medical records get", http.MethodGet, "/medical_records", patientController.GetAllTblMedicalRecords},
		{"medical records get", http.MethodPost, "/medical_records/:user_id", patientController.GetUserMedicalRecords},
		{"medical records get single", http.MethodGet, "/medical_records/:id", patientController.GetSingleTblMedicalRecord},
		{"medical records update", http.MethodPut, "/medical_records/:id", patientController.UpdateTblMedicalRecord},
		{"medical records delete", http.MethodDelete, "/medical_records/:id", patientController.DeleteTblMedicalRecord},
	}
}

func getUserRoutes(userController *controller.UserController) Routes {
	return Routes{
		Route{"User", http.MethodPost, constant.RegisterUser, userController.RegisterUser},
		Route{"User", http.MethodPost, constant.AuthUser, userController.LoginUser},
		Route{"User", http.MethodPost, constant.LogoutUser, userController.LogoutUser},
	}
}

func getMailSyncRoutes(gmailSyncController *controller.GmailSyncController) Routes {
	return Routes{
		{"gmail sync route", http.MethodGet, "/inbox/:user_id", gmailSyncController.FetchEmailsHandler},
		{"gmail sync route", http.MethodGet, "/oauth2callback", gmailSyncController.GmailCallbackHandler},
		{"gmail sync route", http.MethodGet, "/login", controller.GmailLoginHandler},
	}
}
