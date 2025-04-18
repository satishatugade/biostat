package router

import (
	"biostat/constant"
	"biostat/controller"
	"biostat/repository"
	"biostat/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitializeRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	log.Println("Inside InitializeRoutes.....")
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

	var roleRepo = repository.NewRoleRepository(db)
	var roleService = service.NewRoleService(roleRepo)

	var patientRepo = repository.NewPatientRepository(db)
	var patientService = service.NewPatientService(patientRepo)

	var userRepo = repository.NewTblUserGtokenRepository(db)
	var userService = service.NewTblUserGtokenService(userRepo)

	var supportGrpRepo = repository.NewSupportGroupRepository(db)
	var supportGrpService = service.NewSupportGroupService(supportGrpRepo)

	var hospitalRepo = repository.NewHospitalRepository(db)
	var hospitalService = service.NewHospitalService(hospitalRepo)

	var appointmentRepo = repository.NewAppointmentRepository(db)
	var appointmentService = service.NewAppointmentService(appointmentRepo)

	var patientController = controller.NewPatientController(patientService, dietService, allergyService, medicalRecordService, medicationService, appointmentService, diagnosticService, userService)

	var emailService = service.NewEmailService()
	var masterController = controller.NewMasterController(allergyService, diseaseService, causeService, symptomService, medicationService, dietService, exerciseService, diagnosticService, roleService, supportGrpService, hospitalService)
	MasterRoutes(apiGroup, masterController, patientController)
	PatientRoutes(apiGroup, patientController)

	var userController = controller.NewUserController(patientService, roleService, userService, emailService)
	UserRoutes(apiGroup, userController)

	var gmailRecordsController = controller.NewGmailSyncController(medicalRecordService, userService)

	GmailSyncRoutes(apiGroup, gmailRecordsController)

}

func getMasterRoutes(masterController *controller.MasterController, patientController *controller.PatientController) Routes {
	return Routes{

		//Roles
		Route{"Roles", http.MethodPost, constant.GetRole, masterController.GetRoleById},
		// Bulk upload master data
		Route{"Disease", http.MethodPost, constant.BulkUpload, masterController.UploadMasterData},

		//disease master

		Route{"DP", http.MethodPost, constant.CreateDP, masterController.CreateDiseaseProfile},
		Route{"Disease", http.MethodPost, constant.AddDisease, masterController.CreateDisease},
		Route{"Disease", http.MethodPost, constant.Disease, masterController.GetDiseaseInfo},
		Route{"Disease", http.MethodPost, constant.AllDisease, masterController.GetDiseaseInfo},
		Route{"Disease", http.MethodPut, constant.UpdateDisease, masterController.UpdateDiseaseInfo},
		Route{"Disease", http.MethodPost, constant.DeleteDisease, masterController.DeleteDisease},

		//DM AUDIT
		Route{"Disease", http.MethodPost, constant.DiseaseAudit, masterController.GetDiseaseAuditLogs},

		//disease condition(profile)
		Route{"Disease profile", http.MethodPost, constant.DiseaseProfile, masterController.GetDiseaseProfile},
		Route{"Disease profile", http.MethodPost, constant.SingleDiseaseProfile, masterController.GetDiseaseProfileById},

		//Causes master
		Route{"Causes", http.MethodPost, constant.DCMapping, masterController.AddDiseaseCauseMapping},
		Route{"Causes", http.MethodPost, constant.Cause, masterController.GetAllCauses},
		Route{"Causes", http.MethodPost, constant.AddCause, masterController.AddDiseaseCause},
		Route{"Causes", http.MethodPut, constant.UpdateCause, masterController.UpdateDiseaseCause},
		Route{"Causes", http.MethodPost, constant.DeleteCause, masterController.DeleteCause},
		Route{"Causes", http.MethodPost, constant.CauseAudit, masterController.GetCauseAuditRecord},

		//symptoms master
		Route{"Symptom", http.MethodPost, constant.DSMapping, masterController.AddDiseaseSymptomMapping},
		Route{"Symptom", http.MethodPost, constant.Symptom, masterController.GetAllSymptom},
		Route{"Symptom", http.MethodPost, constant.AddSymptom, masterController.AddSymptom},
		Route{"Symptom", http.MethodPut, constant.UpdateSymptom, masterController.UpdateDiseaseSymptom},
		Route{"Symptom", http.MethodPost, constant.DeleteSymptom, masterController.DeleteSymptom},
		Route{"Causes", http.MethodPost, constant.SymptomAudit, masterController.GetSymptomAuditRecord},

		// Allergy master
		Route{"Allergy", http.MethodPost, constant.AllergyMaster, masterController.GetAllergyRestrictions},

		//Medication master
		Route{"Medication", http.MethodPost, constant.DMMapping, masterController.AddDiseaseMedicationMapping},
		Route{"Medication", http.MethodPost, constant.Medication, masterController.GetAllMedication},
		Route{"Medication", http.MethodPost, constant.AddMedication, masterController.AddMedication},
		Route{"Medication", http.MethodPut, constant.UpdateMedication, masterController.UpdateMedication},
		Route{"Medication", http.MethodPost, constant.DeleteMedication, masterController.DeleteMedication},
		Route{"Medication", http.MethodPost, constant.MedicationAudit, masterController.GetMedicationAuditRecord},

		//exercise master
		Route{"Exercise", http.MethodPost, constant.DEMapping, masterController.AddDiseaseExerciseMapping},
		Route{"Exercise", http.MethodPost, constant.Exercise, masterController.AddExercise},
		Route{"Exercise", http.MethodPost, constant.AllExercise, masterController.GetAllExercises},
		Route{"Exercise", http.MethodPost, constant.SingleExercise, masterController.GetExerciseById},
		Route{"Exercise", http.MethodPut, constant.UpdateExercise, masterController.UpdateExercise},
		Route{"Exercise", http.MethodPost, constant.DeleteExercise, masterController.DeleteExercise},
		Route{"Exercise", http.MethodPost, constant.ExerciseAudit, masterController.GetExerciseAuditRecord},

		//diet master

		Route{"Diet", http.MethodPost, constant.DDMapping, masterController.AddDiseaseDietMapping},
		Route{"Diet", http.MethodPost, constant.Diet, masterController.AddDietPlanTemplate},
		Route{"Diet", http.MethodPost, constant.AllDietTemplate, masterController.GetAllDietPlanTemplates},
		Route{"Diet", http.MethodPost, constant.SingleDiet, masterController.GetDietPlanById},
		Route{"Diet", http.MethodPut, constant.UpdateDiet, masterController.UpdateDietPlanTemplate},

		// Diagnostic Test Routes

		Route{"DTM", http.MethodPost, constant.DDTMapping, masterController.AddDiseaseDiagnosticTestMapping},
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
		Route{"DTM", http.MethodPost, constant.DeleteDTComponent, masterController.DeleteDiagnosticTestComponent},

		// Diagnostic Test Component Mapping Routes
		Route{"DTM", http.MethodPost, constant.DiagnosticTestComponentMapping, masterController.CreateDiagnosticTestComponentMapping},
		Route{"DTM", http.MethodPost, constant.DiagnosticTestComponentMappings, masterController.GetAllDiagnosticTestComponentMappings},
		Route{"DTM", http.MethodPut, constant.DiagnosticTestComponentMapping, masterController.UpdateDiagnosticTestComponentMapping},
		Route{"DTM", http.MethodPost, constant.DeleteDiagnosticTestComponentMapping, masterController.DeleteDiagnosticTestComponentMapping},

		Route{"D-LAB", http.MethodPost, constant.DiagnosticLab, masterController.CreateLab},
		Route{"D-LAB", http.MethodPost, constant.GetLabById, masterController.GetLabById},
		Route{"D-LAB", http.MethodPost, constant.GetAllLab, masterController.GetAllLabs},
		Route{"D-LAB", http.MethodPut, constant.UpdateLabInfo, masterController.UpdateLab},
		Route{"D-LAB", http.MethodPost, constant.DeleteLab, masterController.DeleteLab},
		Route{"D-LAB", http.MethodPost, constant.AuditViewLab, masterController.GetDiagnosticLabAuditRecord},

		Route{"Support-Group", http.MethodPost, constant.AddGroup, masterController.AddSupportGroup},
		Route{"Support-Group", http.MethodPost, constant.GetAllGroup, masterController.GetAllSupportGroups},
		Route{"GET-SUPPORT-GROUP", http.MethodPost, constant.GetGroupById, masterController.GetSupportGroupById},
		Route{"UPDATE-SUPPORT-GROUP", http.MethodPut, constant.UpadteSupportGroup, masterController.UpdateSupportGroup},
		Route{"DELETE-SUPPORT-GROUP", http.MethodPost, constant.DeleteSupportGroup, masterController.DeleteSupportGroup},
		Route{"DELETE-SUPPORT-GROUP", http.MethodPost, constant.AuditSupportGroup, masterController.GetSupportGroupAuditRecord},

		// Hospital Routes
		Route{"Add-Hospital", http.MethodPost, constant.AddHospital, masterController.AddHospital},
		Route{"Update-Hospital", http.MethodPut, constant.UpdateHospital, masterController.UpdateHospital},
		Route{"Get-All-Hospitals", http.MethodPost, constant.GetAllHospitals, masterController.GetAllHospitals},
		Route{"Get-Hospital-By-Id", http.MethodPost, constant.GetHospitalById, masterController.GetHospitalById},
		Route{"Delete-Hospital", http.MethodPost, constant.DeleteHospital, masterController.DeleteHospital},
		Route{"Audit-hospital", http.MethodPost, constant.AuditHospital, masterController.GetHospitalAuditRecord},

		Route{"Audit-hospital", http.MethodPost, constant.MappedHospitalService, masterController.AddServiceMapping},

		Route{"Add-Service", http.MethodPost, constant.AddService, masterController.CreateService},
		Route{"Get-All-Services", http.MethodPost, constant.GetAllServices, masterController.GetAllServices},
		Route{"Get-Service-By-Id", http.MethodPost, constant.GetServiceById, masterController.GetServiceById},
		Route{"Update-Service", http.MethodPut, constant.UpdateService, masterController.UpdateService},
		Route{"Delete-Service", http.MethodPost, constant.DeleteService, masterController.DeleteService},
		Route{"Audit-Service", http.MethodPost, constant.AuditService, masterController.GetServiceAuditRecord},

		Route{"Test Reference Range", http.MethodPost, constant.AddRefRange, masterController.AddTestReferenceRange},
		Route{"Test Reference Range", http.MethodPut, constant.UpdateRefRange, masterController.UpdateTestReferenceRange},
		Route{"Test Reference Range", http.MethodPost, constant.DeleteRefRange, masterController.DeleteTestReferenceRange},
		Route{"Test Reference Range", http.MethodPost, constant.ViewRefRange, masterController.ViewTestReferenceRange},
		Route{"Test Reference Range", http.MethodPost, constant.ViewAllRefRange, masterController.GetAllTestReferenceRange},
		Route{"Test Reference Range", http.MethodPost, constant.ViewAuditRefRange, masterController.GetTestReferenceRangeAuditRecord},
	}
}
func getPatientRoutes(patientController *controller.PatientController) Routes {
	return Routes{
		//All relations
		Route{"Relations", http.MethodPost, constant.Relation, patientController.GetAllRelation},

		Route{"patient", http.MethodPost, constant.PatientInfo, patientController.GetPatientInfo},
		Route{"patient", http.MethodPost, constant.SinglePatient, patientController.GetPatientByID},
		Route{"patient", http.MethodPut, constant.UpdatePatient, patientController.UpdatePatientInfoById},
		Route{"patient", http.MethodPost, constant.PatientRelative, patientController.AddPatientRelative},
		Route{"patient", http.MethodPost, constant.Relative, patientController.GetPatientRelativeList},

		Route{"patient", http.MethodPost, constant.UserProfile, patientController.GetUserProfile},
		Route{"patient", http.MethodPost, constant.UserOnboardingStatus, patientController.GetUserOnBoardingStatus},

		Route{"D-LAB", http.MethodPost, constant.GetAllLab, patientController.GetAllLabs},

		// patient relatives
		Route{"patient", http.MethodPost, constant.GetRelative, patientController.GetPatientRelative},

		//all relatives list
		Route{"patient", http.MethodPost, constant.RelativeList, patientController.GetRelativeList},

		//patient caregiver list
		Route{"patient", http.MethodPost, constant.Caregiver, patientController.GetPatientCaregiverList},

		//all caregiver
		Route{"patient", http.MethodPost, constant.CaregiverList, patientController.GetCaregiverList},

		// patient doctor
		Route{"patient", http.MethodPost, constant.Doctor, patientController.GetPatientDoctorList},
		//all doctor
		Route{"patient", http.MethodPost, constant.DoctorList, patientController.GetDoctorList},

		//all nurses
		Route{"patient", http.MethodPost, constant.NursesList, patientController.GetNursesList},

		//patient list
		Route{"patient", http.MethodPost, constant.PatientList, patientController.GetPatientList},

		Route{"patient", http.MethodPost, constant.SingleRelative, patientController.GetPatientRelativeByRelativeId},
		Route{"patient", http.MethodPut, constant.UpdateRealtiveInfo, patientController.UpdatePatientRelative},
		Route{"patient disease condition", http.MethodPost, constant.PatientDiseaseCondition, patientController.GetPatientDiseaseProfiles},

		Route{"Medication", http.MethodPost, constant.Medication, patientController.GetMedication},
		// diagnostic lab test result api
		Route{"patient disease condition", http.MethodPost, constant.PatientResultValue, patientController.GetPatientDiagnosticResultValues},

		Route{"patient diet", http.MethodPost, constant.PatientDietPlan, patientController.GetPatientDietPlan},
		Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
		Route{"patient prescription", http.MethodPost, constant.PrescriptionByPatientId, patientController.GetPrescriptionByPatientId},
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

		{"Appointments", http.MethodPost, constant.ScheduleAppointment, patientController.ScheduleAppointment},
		{"Appointments", http.MethodPost, constant.GetAppointments, patientController.GetUserAppointments},

		{"Digi Locker", http.MethodPost, constant.SyncDigiLocker, patientController.DigiLockerSyncController},
	}
}

func getUserRoutes(userController *controller.UserController) Routes {
	return Routes{
		Route{"User", http.MethodPost, constant.RegisterUser, userController.RegisterUser},
		Route{"User", http.MethodPost, constant.UserRegistrationByPatient, userController.UserRegisterByPatient},
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
