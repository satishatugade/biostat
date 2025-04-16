package constant

const (
	Allergy              = "/allergy/:patient_id"
	UpdateAllergy        = "/update-allergy"
	AllergyMaster        = "allergy"
	CustomRange          = "/custom-range"
	PatientAllergy       = "/add-allergy"
	Patient              = "/registration"
	PatientList          = "/patient-list"
	PatientInfo          = "/patient-info"
	SinglePatient        = "/patient-info/:patient_id"
	UpdatePatient        = "/update-patient-info"
	PatientRelative      = "/patient-relative"
	GetRelative          = "/patient-relative/:patient_id"
	Relative             = "/get-relative/:patient_id"
	RelativeList         = "/relative-list"
	SingleRelative       = "/relative/:relative_id"
	UpdateRealtiveInfo   = "/patient-relative/:relative_id"
	UserProfile          = "/user-profile"
	UserOnboardingStatus = "/user-onboarding-status"

	Caregiver     = "/get-caregiver/:patient_id"
	CaregiverList = "/caregiver-list"

	Doctor     = "/get-doctor/:patient_id"
	DoctorList = "/doctor-list"

	NursesList = "/nurse-list"

	// PatientDiseaseCondition = "/patient-disease-condition/:patient_disease_profile_id"
	PatientResultValue        = "/diagnostic-value"
	PatientDiseaseCondition   = "/patient-disease-condition/:patient_id"
	PatientDietPlan           = "/patient-diet-plan/:patient_id"
	PatientRegistration       = "/patient-registration"
	PatientPrescription       = "/patient-prescription"
	PrescriptionByPatientId   = "/patient-prescription/:patient_id"
	AllPrescription           = "/get-prescription"
	UpdatePrescription        = "/patient-prescription/:prescription_id"
	User                      = "/getuser"
	RegisterUser              = "/register"
	UserRegistrationByPatient = "/create-by-patient/:user_id"

	AuthUser                             = "/auth/login"
	LogoutUser                           = "/logout"
	Disease                              = "/get-disease/:disease_id"
	AllDisease                           = "/get-disease"
	AddDisease                           = "/add-disease"
	CreateDP                             = "/create-dp"
	UpdateDisease                        = "/update-disease"
	DeleteDisease                        = "/delete-disease/:disease_id"
	DiseaseAudit                         = "/disease-audit"
	Cause                                = "/causes"
	DCMapping                            = "/dc-mapping"
	AddCause                             = "/add-cause"
	UpdateCause                          = "/update-cause"
	DeleteCause                          = "/delete-cause/:cause_id"
	CauseAudit                           = "/cause-audit"
	Symptom                              = "/symptom"
	DSMapping                            = "/ds-mapping"
	AddSymptom                           = "/add-symptom"
	UpdateSymptom                        = "/update-symptom"
	DeleteSymptom                        = "/delete-symptom/:symptom_id"
	SymptomAudit                         = "symptom-audit"
	DiseaseProfile                       = "/disease-profile"
	SingleDiseaseProfile                 = "/disease-profile/:disease_profile_id"
	DiagnosticTests                      = "/diagnostic-tests"
	DDTMapping                           = "/ddt-mapping"
	Medication                           = "/get-medication"
	DMMapping                            = "/dm-mapping"
	AddMedication                        = "/add-medication"
	UpdateMedication                     = "/update-medication"
	DeleteMedication                     = "/delete-medication/:medication_id"
	MedicationAudit                      = "/medication-audit"
	DiagnosticTest                       = "/diagnostic-test"
	SingleDiagnosticTest                 = "/diagnostic-test/:diagnosticTestId"
	DiagnosticComponents                 = "/diagnostic-components"
	DiagnosticComponent                  = "/diagnostic-component"
	SingleDiagnosticComponent            = "/diagnostic-component/:diagnosticComponentId"
	DiagnosticTestComponentMappings      = "/diagnostic-test-component-mappings"
	DiagnosticTestComponentMapping       = "/diagnostic-test-component-mapping"
	DeleteDiagnosticTestComponentMapping = "/delete-diagnostic-test-component-mapping"
	DeleteDTComponent                    = "/delete-dt-component/:diagnostic_test_component_id"

	DEMapping      = "de-mapping"
	Exercise       = "/add-exercise"
	AllExercise    = "/get-exercise"
	SingleExercise = "/exercise/:exercise_id"
	UpdateExercise = "/update-exercise"
	DeleteExercise = "/delete-exercise/:exercise_id"
	ExerciseAudit  = "/exercise-audit"

	Diet            = "/add-diet-template"
	AllDietTemplate = "/diet-template"
	SingleDiet      = "/diet-template/:diet_id"
	UpdateDiet      = "/update-diet/:diet_id"
	DDMapping       = "/dd-mapping"

	//Roles master
	GetRole  = "/get-role/:role_id"
	Relation = "/all-relation"

	//bulk upload master data
	BulkUpload = "/bulk-upload/:entity"

	DiagnosticLab = "diagnostic-lab"
	GetAllLab     = "get-diagnostic-lab"
	GetLabById    = "diagnostic-lab/:lab_id"
	UpdateLabInfo = "update-lab-info"
	DeleteLab     = "delete-lab/:lab_id"
	AuditViewLab  = "lab-audit-view"

	AddGroup           = "add-support-group"
	GetAllGroup        = "get-support-group"
	GetGroupById       = "support-group/:support_group_id"
	DeleteSupportGroup = "delete-support-group/:support_group_id"
	UpadteSupportGroup = "upadate-support-group"
	AuditSupportGroup  = "support-group"

	//hospital
	AddHospital     = "add-hospital"
	UpdateHospital  = "update-hospital"
	GetAllHospitals = "get-all-hospitals"
	GetHospitalById = "get-hospital-by-id"
	DeleteHospital  = "delete-hospital/:hospital_id"
	AuditHospital   = "hospital-audit"

	MappedHospitalService = "map-hospital-service"
	//services
	AddService     = "add-service"
	GetAllServices = "get-all-service"
	GetServiceById = "get-service-by-id/:service_id"
	UpdateService  = "update-service"
	DeleteService  = "delete-service/:service_id"
	AuditService   = "service-audit"

	// Appointments
	ScheduleAppointment = "/schedule-appointment"
	GetAppointments     = "/get-appointments"
)

const (
	Success = "success"
	Failure = "failure"
)

const (
	CREATE = "Create"
	UPDATE = "Update"
	DELETE = "Delete"
)
const (
	AuditErrorMessage   = "Unable to show a history of this record. It has not been changed since it was created"
	AuditSuccessMessage = "Audit records fetched successfully"
)
