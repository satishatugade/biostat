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

	AuthUser             = "/auth/login"
	LogoutUser           = "/logout"
	Disease              = "/get-disease/:disease_id"
	AllDisease           = "/get-disease"
	AddDisease           = "/add-disease"
	UpdateDisease        = "/update-disease"
	DeleteDisease        = "/delete-disease/:disease_id"
	DiseaseAudit         = "/disease-audit"
	Cause                = "/causes"
	AddCause             = "/add-cause"
	UpdateCause          = "/update-cause"
	DeleteCause          = "/delete-cause/:cause_id"
	CauseAudit           = "/cause-audit"
	Symptom              = "/symptom"
	AddSymptom           = "/add-symptom"
	UpdateSymptom        = "/update-symptom"
	DeleteSymptom        = "/delete-symptom/:symptom_id"
	SymptomAudit         = "symptom-audit"
	DiseaseProfile       = "/disease-profile"
	SingleDiseaseProfile = "/disease-profile/:disease_profile_id"
	DiagnosticTests      = "/diagnostic-tests"
	Medication           = "/get-medication"
	AddMedication        = "/add-medication"
	// UpdateMedication    = "/update-medication/:medication_id"
	UpdateMedication                = "/update-medication"
	DiagnosticTest                  = "/diagnostic-test"
	SingleDiagnosticTest            = "/diagnostic-test/:diagnosticTestId"
	DiagnosticComponents            = "/diagnostic-components"
	DiagnosticComponent             = "/diagnostic-component"
	SingleDiagnosticComponent       = "/diagnostic-component/:diagnosticComponentId"
	DiagnosticTestComponentMappings = "/diagnostic-test-component-mappings"
	DiagnosticTestComponentMapping  = "/diagnostic-test-component-mapping"

	Exercise       = "/add-exercise"
	AllExercise    = "/get-exercise"
	SingleExercise = "/exercise/:exercise_id"
	UpdateExercise = "/update-exercise/:exercise_id"

	Diet            = "/add-diet-template"
	AllDietTemplate = "/diet-template"
	SingleDiet      = "/diet-template/:diet_id"
	UpdateDiet      = "/update-diet/:diet_id"

	//Roles master
	GetRole = "/get-role/:role_id"

	//bulk upload master data
	BulkUpload = "/bulk-upload/:entity"
)

const (
	Success = "success"
	Failure = "failure"
)

const (
	UPDATE = "Update"
	DELETE = "Delete"
)
