package router

import (
	"biostat/auth"
	"biostat/config"
	"biostat/constant"
	"biostat/controller"
	"biostat/repository"
	"biostat/service"
	"biostat/worker"
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

	var notificiationRepo = repository.NewUserNotificationRepository(db)
	var notificationService = service.NewNotificationService(notificiationRepo)

	var emailService = service.NewEmailService(notificiationRepo)
	var apiService = service.NewApiService()

	var userRepo = repository.NewTblUserTokenRepository(db)
	var userService = service.NewTblUserTokenService(userRepo)

	var medicalRecordsRepo = repository.NewTblMedicalRecordRepository(db)
	var patientRepo = repository.NewPatientRepository(db)

	var roleRepo = repository.NewRoleRepository(db)
	var roleService = service.NewRoleService(roleRepo)

	var patientService = service.NewPatientService(patientRepo, apiService, allergyService, medicalRecordsRepo, roleRepo, notificationService)

	var diagnosticRepo = repository.NewDiagnosticRepository(db)
	var diagnosticService = service.NewDiagnosticService(diagnosticRepo, emailService, patientService)
	var medicalRecordService = service.NewTblMedicalRecordService(medicalRecordsRepo, apiService, diagnosticService, patientService, userService, config.AsynqClient, config.RedisClient)

	var smsService = service.NewSmsService()

	var supportGrpRepo = repository.NewSupportGroupRepository(db)
	var supportGrpService = service.NewSupportGroupService(supportGrpRepo)

	var hospitalRepo = repository.NewHospitalRepository(db)
	var hospitalService = service.NewHospitalService(hospitalRepo)

	var appointmentRepo = repository.NewAppointmentRepository(db)
	var appointmentService = service.NewAppointmentService(appointmentRepo)

	var orderRepo = repository.NewOrderRepository(db)
	var orderService = service.NewOrderService(orderRepo)

	var patientController = controller.NewPatientController(patientService, dietService, allergyService, medicalRecordService, medicationService, appointmentService, diagnosticService, userService, apiService, diseaseService, smsService, emailService, orderService, notificationService)

	var masterController = controller.NewMasterController(allergyService, diseaseService, causeService, symptomService, medicationService, dietService, exerciseService, diagnosticService, roleService, supportGrpService, hospitalService, userService)
	MasterRoutes(apiGroup, masterController, patientController)
	PatientRoutes(apiGroup, patientController)

	var authService = auth.NewAuthService(userRepo, userService, emailService)
	var userController = controller.NewUserController(patientService, roleService, userService, emailService, authService)
	UserRoutes(apiGroup, userController)

	var gmailRecordsController = controller.NewGmailSyncController(medicalRecordService, userService)

	GmailSyncRoutes(apiGroup, gmailRecordsController)

	// Workers
	worker.StartAppointmentScheduler(appointmentService)
	go worker.InitAsynqWorker(apiService, patientService, diagnosticService, medicalRecordsRepo)

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

		Route{"Causes-Type", http.MethodPost, constant.CauseType, masterController.GetAllCauseTypes},
		Route{"Causes-Type", http.MethodPost, constant.AddCauseType, masterController.AddCauseType},
		Route{"Causes-Type", http.MethodPut, constant.UpdateCauseType, masterController.UpdateCauseType},
		Route{"Causes-Type", http.MethodPost, constant.DeleteCauseType, masterController.DeleteCauseType},
		Route{"Causes-Type-Audit", http.MethodPost, constant.CauseTypeAudit, masterController.GetCauseTypeAuditRecord},

		//symptoms master
		Route{"Symptom", http.MethodPost, constant.DSMapping, masterController.AddDiseaseSymptomMapping},
		Route{"Symptom", http.MethodPost, constant.Symptom, masterController.GetAllSymptom},
		Route{"Symptom", http.MethodPost, constant.AddSymptom, masterController.AddSymptom},
		Route{"Symptom", http.MethodPut, constant.UpdateSymptom, masterController.UpdateDiseaseSymptom},
		Route{"Symptom", http.MethodPost, constant.DeleteSymptom, masterController.DeleteSymptom},
		Route{"Symptom", http.MethodPost, constant.SymptomAudit, masterController.GetSymptomAuditRecord},

		//symptoms-Type
		Route{"Symptom-Type", http.MethodPost, constant.SymptomType, masterController.GetAllSymptomTypes},
		Route{"Symptom-Type", http.MethodPost, constant.AddSymptomType, masterController.AddSymptomType},
		Route{"Symptom-Type", http.MethodPut, constant.UpdateSymptomType, masterController.UpdateSymptomType},
		Route{"Symptom-Type", http.MethodPost, constant.DeleteSymptomType, masterController.DeleteSymptomType},
		Route{"Symptom-Type-Audit", http.MethodPost, constant.SymptomTypeAudit, masterController.GetSymptomTypeAuditRecord},

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
		Route{"Genders", http.MethodPost, constant.Gender, patientController.GetAllGender},
		Route{"Genders", http.MethodPost, constant.GenderId, patientController.GetGenderById},

		Route{"patient", http.MethodPost, constant.PatientInfo, patientController.GetPatientInfo},
		Route{"patient", http.MethodPut, constant.UpdatePatient, patientController.UpdatePatientInfoById},
		Route{"patient", http.MethodPost, constant.PatientRelative, patientController.AddPatientRelative},
		Route{"patient", http.MethodPost, constant.Relative, patientController.GetPatientRelativeList},
		Route{"patient", http.MethodPost, constant.PrimaryCaregiver, patientController.AssignPrimaryCaregiver},

		Route{"patient", http.MethodPost, constant.UserProfile, patientController.GetUserProfile},
		Route{"patient", http.MethodPost, constant.UserOnboardingStatus, patientController.GetUserOnBoardingStatus},

		Route{"D-LAB", http.MethodPost, constant.GetAllLab, patientController.GetAllLabs},
		Route{"D-LAB", http.MethodPost, constant.AddLab, patientController.AddLab},
		Route{"D-LAB", http.MethodPost, constant.GetPatientLabs, patientController.GetPatientDiagnosticLabs},

		// patient relatives
		Route{"patient", http.MethodPost, constant.GetRelative, patientController.GetPatientRelative},

		//all relatives list
		Route{"patient", http.MethodPost, constant.RelativeList, patientController.GetRelativeList},

		//patient caregiver list
		Route{"patient", http.MethodPost, constant.Caregiver, patientController.GetPatientCaregiverList},

		Route{"patient - caregiver", http.MethodPost, constant.RemoveCaregiver, patientController.SetCaregiverMappingDeletedStatus},

		//all caregiver
		Route{"patient", http.MethodPost, constant.CaregiverList, patientController.GetCaregiverList},

		//all doctor
		Route{"patient", http.MethodPost, constant.DoctorList, patientController.GetDoctorList},

		//all nurses
		Route{"patient", http.MethodPost, constant.NursesList, patientController.GetNursesList},

		Route{"patient", http.MethodPost, constant.ChemistList, patientController.GetPharmacistList},

		//patient list
		Route{"patient", http.MethodPost, constant.PatientList, patientController.GetPatientList},

		Route{"patient", http.MethodPost, constant.SingleRelative, patientController.GetPatientRelativeByRelativeId},
		Route{"patient disease condition", http.MethodPost, constant.PatientDiseaseCondition, patientController.GetPatientDiseaseProfiles},
		Route{"patient health Profile", http.MethodPost, constant.UserHealthDetails, patientController.SaveUserHealthProfile},
		Route{"patient health detail", http.MethodPost, constant.HealthDetail, patientController.GetPatientHealthProfileInfo},
		Route{"update health detail", http.MethodPost, constant.UpdateHealthDetail, patientController.UpdatePatientHealthDetail},

		Route{"Medication", http.MethodPost, constant.Medication, patientController.GetMedication},
		// diagnostic lab test result api
		Route{"patient disease condition", http.MethodPost, constant.PatientResultValue, patientController.GetPatientDiagnosticResultValues},
		Route{"patient disease condition", http.MethodPost, constant.HistoricalTrendAnalysis, patientController.GetPatientDiagnosticTrendValue},
		Route{"patient disease condition", http.MethodPost, constant.DisplayConfig, patientController.AddTestComponentDisplayConfig},
		Route{"patient disease condition", http.MethodPost, constant.GetResultValue, patientController.GetDiagnosticResults},
		Route{"patient disease condition", http.MethodPost, constant.GetReportResult, patientController.GetPatientDiagnosticReportResult},
		Route{"patient disease condition", http.MethodPost, constant.ExportReport, patientController.ExportDiagnosticResultsExcel},
		Route{"patient disease condition", http.MethodPost, constant.ExportPDFReport, patientController.ExportDiagnosticResultsPDF},

		Route{"patient diet", http.MethodPost, constant.PatientDietPlan, patientController.GetPatientDietPlan},
		Route{"patient prescription", http.MethodPost, constant.PatientPrescription, patientController.AddPrescription},
		Route{"patient prescription", http.MethodPost, constant.UpdatePrescription, patientController.UpdatePrescription},
		Route{"patient prescription", http.MethodPost, constant.PrescriptionByPatientId, patientController.GetPrescriptionByPatientId},
		Route{"patient prescription", http.MethodPost, constant.PrescriptionDetail, patientController.GetPrescriptionDetailByPatientId},
		Route{"patient prescription", http.MethodPost, constant.UserMedications, patientController.GetUserMedications},
		Route{"prescription explanation", http.MethodPost, constant.PrescriptionInfo, patientController.PrescriptionInfobyAIModel},
		Route{"Pharmacokinetics", http.MethodPost, constant.Pharmacokinetics, patientController.PharmacokineticsInfobyAIModel},
		Route{"SummarizeHistorybyAIModel", http.MethodPost, constant.SummarizeHistory, patientController.SummarizeHistorybyAIModel},

		Route{"Patient Allergy", http.MethodPost, constant.PatientAllergy, patientController.AddPatientAllergyRestriction},
		Route{"Patient Allgery", http.MethodPost, constant.Allergy, patientController.GetPatientAllergyRestriction},
		Route{"Update Patient Allgery", http.MethodPut, constant.UpdateAllergy, patientController.UpdatePatientAllergyRestriction},
		Route{"Add Custom Range", http.MethodPost, constant.CustomRange, patientController.AddPatientClinicalRange},
		// Route{"update  prescription", http.MethodPut, constant.UpdatePrescription, patientController.UpdatePrescription},

		Route{"medical records create", http.MethodPost, constant.UploadRecord, patientController.CreateTblMedicalRecord},
		Route{"medical records", http.MethodPost, constant.MedicalRecord, patientController.GetAllMedicalRecord},
		Route{"medical records get", http.MethodPost, "/medical_records/:user_id", patientController.GetUserMedicalRecords},
		Route{"medical records get single", http.MethodGet, "/medical_records/:id", patientController.GetSingleTblMedicalRecord},
		Route{"medical records update", http.MethodPut, "/medical_records/:id", patientController.UpdateTblMedicalRecord},
		Route{"medical records delete", http.MethodDelete, "/medical_records/:id", patientController.DeleteTblMedicalRecord},

		Route{"Appointments", http.MethodPost, constant.ScheduleAppointment, patientController.ScheduleAppointment},
		Route{"Appointments", http.MethodPost, constant.GetAppointments, patientController.GetUserAppointments},
		Route{"Appointments", http.MethodPut, constant.ScheduleAppointment, patientController.UpdateUserAppointment},

		Route{"Digi Locker", http.MethodPost, constant.SyncDigiLocker, patientController.DigiLockerSyncController},
		Route{"Digi Locker", http.MethodPost, constant.GetMedicalResource, patientController.ReadUserUploadedMedicalFile},

		Route{"patient diagnostic report save", http.MethodPost, constant.SaveReport, patientController.SaveReport},
		Route{"DigitizationStatus", http.MethodPost, constant.DigitizationStatus, patientController.GetDigitizationStatus},
		Route{"Merge component", http.MethodPost, constant.MergeComponent, patientController.AddMappingToMergeTestComponent},
		Route{"Add health stat", http.MethodPost, constant.HealthStats, patientController.AddHealthStats},
		Route{"patient diagnostic report save", http.MethodPost, constant.ReportArchive, patientController.ArchivePatientDiagnosticReport},

		Route{"AddReportNote", http.MethodPost, constant.AddNote, patientController.AddPatientReportNote},

		Route{"Disease profile", http.MethodPost, constant.DiseaseProfile, patientController.GetDiseaseProfiles},
		Route{"Disease profile", http.MethodPost, constant.AttachDiseaseProfile, patientController.AttachDiseaseProfileTOPatient},
		Route{"Disease profile", http.MethodPost, constant.UpdateDiseaseProfile, patientController.UpdateDiseaseProfile},

		Route{"User Oders", http.MethodPost, constant.AddOrder, patientController.CreateOrder},
		Route{"User Oders", http.MethodPost, constant.GetOrders, patientController.GetUserOrders},

		Route{"send-sms", http.MethodPost, constant.SendSMS, patientController.SendSMS},
		Route{"send-sms", http.MethodPost, constant.ShareReport, patientController.ShareReport},

		Route{"User Notifications", http.MethodPost, constant.Reminder, patientController.SetUserReminder},
		Route{"User Notifications", http.MethodPost, constant.Messages, patientController.GetUserMessages},

		Route{"User permissions", http.MethodPost, constant.Permission, patientController.AssignPermissionHandler},
		Route{"User SOS", http.MethodPost, constant.SOS, patientController.SendSOSHandler},
	}
}

func getUserRoutes(userController *controller.UserController) Routes {
	return Routes{
		Route{"User", http.MethodPost, constant.RegisterUser, userController.RegisterUser},
		Route{"User", http.MethodPost, constant.UserRegistrationByPatient, userController.UserRegisterByPatient},
		Route{"User", http.MethodPost, constant.AuthUser, userController.LoginUser},
		Route{"User", http.MethodPost, constant.RefreshToken, userController.RefreshToken},
		Route{"User", http.MethodPost, constant.LogoutUser, userController.LogoutUser},
		Route{"redirect", http.MethodGet, constant.RedirectURL, userController.UserRedirect},
		Route{"User", http.MethodPost, constant.ValidateUserEmailMobile, userController.CheckUserEmailMobileExist},
		Route{"User", http.MethodPost, constant.ResetPassword, userController.ResetUserPassword},
		Route{"User", http.MethodPost, constant.SentLink, userController.SendResetPasswordLink},
		Route{"User", http.MethodPost, constant.MapUserToPatient, userController.AddRelationHandler},
		// Route{"User", http.MethodPost, constant.SentOTP, userController.SendOTP},
		// Route{"User", http.MethodPost, constant.VerifyOTP, userController.VerifyOTP},

		//postal code
		Route{"Postalcode", http.MethodPost, constant.Postalcode, userController.FetchAddressByPincode},
	}
}

func getMailSyncRoutes(gmailSyncController *controller.GmailSyncController) Routes {
	return Routes{
		{"gmail sync route", http.MethodPost, "/app-sync", gmailSyncController.FetchEmailsHandler},
		{"gmail sync route", http.MethodGet, "/oauth2callback", gmailSyncController.GmailCallbackHandler},
		{"gmail sync route", http.MethodGet, "/login", controller.GmailLoginHandler},
	}
}
