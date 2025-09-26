package constant

const (
	Version              = "/api/version"
	Allergy              = "/allergy"
	UpdateAllergy        = "/update-allergy"
	AllergyMaster        = "allergy"
	CustomRange          = "/custom-range"
	PatientAllergy       = "/add-allergy"
	PatientList          = "/patient-list"
	PatientInfo          = "/patient-info"
	SinglePatient        = "/patient-info/:patient_id"
	UpdatePatient        = "/update-patient-info"
	UpdateRelative       = "/update-relative-info"
	GetRelative          = "/patient-relative/:patient_id"
	RelativeInfo         = "/get-relative/:patient_id"
	PrimaryCaregiver     = "/primary-caregiver"
	RelativeList         = "/relative-list"
	SingleRelative       = "/relative/:relative_id"
	UserProfile          = "/user-profile"
	UserOnboardingStatus = "/user-onboarding-status"
	UserHealthDetails    = "/user-health-details"
	HealthDetail         = "/health-detail"
	UpdateHealthDetail   = "/update-health-detail"

	GetCaregiver    = "/get-caregiver"
	AssignedPatient = "/caregiver/assigned-patient"
	RemoveMapping   = "/remove/user-mapping"
	CaregiverList   = "/caregiver-list"

	DoctorList = "/doctor-list/:user"

	NursesList  = "/nurse-list"
	ChemistList = "/pharmacist-list"

	PatientResultValue        = "/diagnostic-value"
	HistoricalTrendAnalysis   = "/historical-trend-analysis"
	DisplayConfig             = "/diagnostic-component/configuration"
	GetResultValue            = "/get-result-value"
	GetReportResult           = "/report/diagnostic-trend"
	ExportReport              = "/export-report-data"
	ExportPDFReport           = "/export-pdf"
	PatientDiseaseCondition   = "/patient-disease-condition"
	PatientDietPlan           = "/patient-diet-plan/:patient_id"
	PatientRegistration       = "/patient-registration"
	PatientPrescription       = "/patient-prescription"
	ArchivePrescription       = "/archive-prescription"
	UpdatePrescription        = "/update-prescription"
	PrescriptionByPatientId   = "/get-prescription"
	PrescriptionDetail        = "/prescription-detail"
	PrescriptionInfo          = "/prescription-info"
	UserMedications           = "/user-medications"
	Pharmacokinetics          = "/api/drug/pharmacokinetics"
	SummarizeHistory          = "/api/summerize-history"
	User                      = "/getuser"
	RegisterUser              = "/register"
	UserRegistrationByPatient = "/create-by-patient"
	MapUserToPatient          = "/map-user-to-patient"

	AuthUser                             = "/auth/login"
	RefreshToken                         = "/api//auth/refresh-token"
	LogoutUser                           = "/logout"
	Disease                              = "/get-disease/:disease_id"
	AllDisease                           = "/get-disease"
	AddDisease                           = "/add-disease"
	CreateDP                             = "/create-dp"
	UpdateDisease                        = "/update-disease"
	DeleteDisease                        = "/delete-disease/:disease_id"
	DiseaseAudit                         = "/disease-audit"
	Cause                                = "/causes"
	CauseType                            = "/causes-type"
	DCMapping                            = "/dc-mapping"
	AddCause                             = "/add-cause"
	AddCauseType                         = "/add-cause-type"
	UpdateCause                          = "/update-cause"
	UpdateCauseType                      = "/update-cause-type"
	DeleteCause                          = "/delete-cause/:cause_id"
	DeleteCauseType                      = "/delete-cause-type/:cause_type_id"
	CauseTypeAudit                       = "/cause-type-audit"
	CauseAudit                           = "/cause-audit"
	Symptom                              = "/symptom"
	SymptomType                          = "/symptom-type"
	DSMapping                            = "/ds-mapping"
	AddSymptom                           = "/add-symptom"
	AddSymptomType                       = "/add-symptom-type"
	UpdateSymptom                        = "/update-symptom"
	UpdateSymptomType                    = "/update-symptom-type"
	DeleteSymptom                        = "/delete-symptom/:symptom_id"
	DeleteSymptomType                    = "/delete-symptom-type/:symptom_type_id"
	SymptomAudit                         = "/symptom-audit"
	SymptomTypeAudit                     = "/symptom-type-audit"
	DiseaseProfile                       = "/disease-profile"
	AttachDiseaseProfile                 = "/attach-disease-profile"
	UpdateDiseaseProfile                 = "/update-disease-profile"
	SingleDiseaseProfile                 = "/disease-profile/:disease_profile_id"
	DiagnosticTests                      = "/diagnostic-tests"
	DDTMapping                           = "/ddt-mapping"
	Medication                           = "/get-medication"
	Sources                              = "/get-sources"
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
	Gender   = "/all-gender"
	GenderId = "/gender/:gender_id"

	//bulk upload master data
	BulkUpload = "/bulk-upload/:entity"

	DiagnosticLab  = "/diagnostic-lab"
	GetAllLab      = "/get-diagnostic-lab"
	AddLab         = "/add-diagnostic-lab"
	GetPatientLabs = "/diagnostic-lab"
	GetLabById     = "/diagnostic-lab/:lab_id"
	UpdateLabInfo  = "/update-lab-info"
	DeleteLab      = "/delete-lab/:lab_id"
	AuditViewLab   = "/lab-audit-view"

	AddGroup           = "/add-support-group"
	GetAllGroup        = "/get-support-group"
	GetGroupById       = "/support-group/:support_group_id"
	DeleteSupportGroup = "/delete-support-group/:support_group_id"
	UpadteSupportGroup = "/upadate-support-group"
	AuditSupportGroup  = "/support-group"

	//hospital
	AddHospital     = "/add-hospital"
	UpdateHospital  = "/update-hospital"
	GetAllHospitals = "/get-all-hospitals"
	GetHospitalById = "/get-hospital-by-id"
	DeleteHospital  = "/delete-hospital/:hospital_id"
	AuditHospital   = "/hospital-audit"

	MappedHospitalService = "/map-hospital-service"
	//services
	AddService     = "/add-service"
	GetAllServices = "/get-all-service"
	GetServiceById = "/get-service-by-id/:service_id"
	UpdateService  = "/update-service"
	DeleteService  = "/delete-service/:service_id"
	AuditService   = "/service-audit"

	// Appointments
	ScheduleAppointment = "/schedule-appointment"
	GetAppointments     = "/get-appointments"

	AddRefRange       = "/add-range"
	UpdateRefRange    = "/update-range"
	DeleteRefRange    = "/delete-range/:test_reference_range_id"
	ViewRefRange      = "/view-range/:test_reference_range_id"
	ViewAllRefRange   = "/view-all-range"
	ViewAuditRefRange = "/view-audit-range"

	SubscriptionEnabledStatus = "/api/subscription/status"
	UpdateSubscriptionStatus  = "/subscription/update-status"

	SyncDigiLocker      = "/sync-digilocker"
	GetMedicalResource  = "/get-medical-resource"
	MedicalRecord       = "/medical-record-info"
	FetchUserTag        = "/get-report-tag"
	AddTag              = "/record/add-tag"
	LabReportName       = "/diagnostic-lab-report-name"
	UserMedicalRecord   = "/medical_records/:user_id"
	GetByRecordId       = "/medical_records/:id"
	UploadRecord        = "/medical_records"
	UpdateMedicalRecord = "/medical_records"
	DeleteMedicalRecord = "/medical_records/:id"

	AddOrder  = "/order"
	GetOrders = "/orders"

	MergeComponent          = "/merge-component"
	ReportDigitization      = "/report/digitization/:record_id"
	MoveRecord              = "/move-record"
	DigitizationStatus      = "/report/digitization-status/:record_id"
	HealthStats             = "/health-stats"
	ReportArchive           = "/diagnostic-report/archive"
	AddNote                 = "/add-report-note"
	SendSMS                 = "/send-sms"
	ShareReport             = "/share-report"
	RedirectURL             = "/r/:code"
	ValidateUserEmailMobile = "/checkUserMobileEmailExist"
	ResetPassword           = "/reset-password"
	SentLink                = "/reset-password-link"
	SentOTP                 = "/send-otp"
	VerifyOTP               = "/verify-otp"
	Postalcode              = "/address/postalcode"
	Messages                = "/messages"
	Reminder                = "/reminder"
	UpdateReminder          = "/update-reminder"
	Reminders               = "/reminders"
	Permission              = "/permission"
	ManageFamilyPermission  = "/family/manage-permission"
	Address                 = "/mapped-user/address"
	SOS                     = "/sos"
	ShareList               = "/share-list"
	FamilySubscription      = "/family-subscription"
	ActiveSubscriptionPlan  = "/active-subscription-plan"
	GetSubscriptionPlan     = "/get-subscription-plan"
	BIOCHATBOT              = "/ask-bio"
	RunningProcessStatus    = "/process-status"
	ActivityLog             = "/activity-log"
	RecipientDetails        = "/recipient-details"
	ProviderList            = "/provider-details"
	GmailReSync             = "/resync"
	AppSync                 = "/app-sync"
	OuthCallBack            = "/oauth2callback"
	WebSync                 = "/web-sync/:user_id"
	OutlookWebSync          = "/outlook/login/:user_id"
	OutlookCallBack         = "/outlook/callback"
	YahooWebSync            = "/yahoo/login/:user_id"
	YahooCallBack           = "/yahoo/callback"
	Transcribe              = "/transcribe"
)

const (
	Success             = "success"
	Failure             = "failure"
	Active              = "Active"
	InActive            = "Inactive"
	CREATE              = "Create"
	UPDATE              = "Update"
	DELETE              = "Delete"
	SUBSCRIPTIONENABLED = "subscription_enabled"
	Running             = "running"
)

const (
	KeyCloakErrorMessage = "User not found on keycloak server. please check!"
	AuditErrorMessage    = "Unable to show a history of this record. It has not been changed since it was created"
	AuditSuccessMessage  = "Audit records fetched successfully"
)

var ServiceError = "AI service unavailable"

type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusRetrying   JobStatus = "retrying"
	StatusSuccess    JobStatus = "success"
	StatusFailed     JobStatus = "failed"
)

type RecordCategory string

const (
	TESTREPORT       RecordCategory = "test_report"
	DUPLICATE        RecordCategory = "duplicate"
	INSURANCE        RecordCategory = "insurance"
	MEDICATION       RecordCategory = "medications"
	VACCINATION      RecordCategory = "vaccinations"
	Prescription     RecordCategory = "prescription"
	DISCHARGESUMMARY RecordCategory = "discharge_summary"
	OTHER            RecordCategory = "other"
	INVOICE          RecordCategory = "invoice"
	NONMEDICAL       RecordCategory = "non_medical"
	SCANS            RecordCategory = "scans"
)

type RecordSubCategory string

const (
	CTSCAN        RecordSubCategory = "ct-scan"
	MRI           RecordSubCategory = "mri"
	XRAY          RecordSubCategory = "x-ray"
	ANGIOGRAPHY   RecordSubCategory = "angiography"
	ENDOSCOPY     RecordSubCategory = "endoscopy"
	SONOGRAPHY    RecordSubCategory = "sonography"
	FRACTURE      RecordSubCategory = "fracture"
	OTHERS        RecordSubCategory = "other"
	SUPPORTINGDOC RecordSubCategory = "supporting_doc"
)

type SubscriptionStatus string

const (
	SUBSCRIPTIONACTIVE SubscriptionStatus = "ACTIVE"
	EXPIRED            SubscriptionStatus = "EXPIRED"
	NOTSTARTED         SubscriptionStatus = "NOTSTARTED"
	NOACTIVEPLAN       SubscriptionStatus = "NOACTIVEPLAN"
	PLANNOTFOUND       SubscriptionStatus = "PLANNOTFOUND"
	EXPIRINGSOON       SubscriptionStatus = "EXPIRINGSOON"
)

type MappingType string

const (
	MappingTypeA   MappingType = "A"
	MappingTypeS   MappingType = "S"
	MappingTypeC   MappingType = "C"
	MappingTypePCG MappingType = "PCG"
	MappingTypeHOF MappingType = "HOF"
	MappingTypeR   MappingType = "R"
	MappingTypeD   MappingType = "D"
	MappingTypeN   MappingType = "N"
	MappingTypeP   MappingType = "P"
)

var FallbackMappingTypes = []MappingType{
	MappingTypeA,
	MappingTypeS,
	MappingTypeC,
	MappingTypePCG,
	MappingTypeHOF,
	MappingTypeR,
	MappingTypeD,
	MappingTypeN,
	MappingTypeP,
}

type ProcessStep string

const (
	ProcessTokenExchange       ProcessStep = "token_exchange"
	ProcessFetchLabs           ProcessStep = "fetch_labs"
	FetchEmailsList            ProcessStep = "fetch_emails_list"
	CheckPasswordProtectedStep ProcessStep = "check_pdf_password_protected"
	ProcessSaveRecords         ProcessStep = "save_records"
	SingleRecordDigitization   ProcessStep = "single_record_digitization"
	ProcessGmailClient         ProcessStep = "gmail_client"
	ProcessDigitization        ProcessStep = "digitization"
	ProcessGmailSearch         ProcessStep = "gmail_search"
	DownloadAttachment         ProcessStep = "downloading_attachment"
	FindingEmailWithAttachment ProcessStep = "finding_email_attachment"
	ProcessVerifyCredentials   ProcessStep = "verify_credentials"
	CheckDocType               ProcessStep = "check_doc_type"
	RetryStep                  ProcessStep = "retrying"
	CallAIService              ProcessStep = "call_ai_service"
	MatchingReport             ProcessStep = "matching_report_name_with_self_or_relative_name"
	CheckReportDuplication     ProcessStep = "checking_report_duplication_by_collection_date_and_test_component"
)

type ProcessStepStatusMessage string

const (
	ProcessStarted                ProcessStepStatusMessage = "Process started"
	TokenExchangeSuccess          ProcessStepStatusMessage = "Token exchange success"
	TokenExchangeFailed           ProcessStepStatusMessage = "Token exchange failed"
	GmailClientCreateFailed       ProcessStepStatusMessage = "Failed to create Gmail client"
	GmailClientCreated            ProcessStepStatusMessage = "Gmail client created"
	FetchUserLab                  ProcessStepStatusMessage = "Fetching user diagnostic labs from his lab list"
	UserLabNotFound               ProcessStepStatusMessage = "Failed to fetch user labs or lab not found"
	FetchEmailAttachment          ProcessStepStatusMessage = "Fetching email attachments"
	CheckPasswordProtectedStepMsg ProcessStepStatusMessage = "Checking is pdf is password protected..."
	EmailAttachmentFetch          ProcessStepStatusMessage = "Email attachments fetch successfully"
	FailedFetchAttachment         ProcessStepStatusMessage = "Failed to fetch email attachments"
	InvalidCredentials            ProcessStepStatusMessage = "Invalid credentials"

	SaveRecord                        ProcessStepStatusMessage = "Storing medical records in the database"
	RecordSaveSuccess                 ProcessStepStatusMessage = "Medical records docs saved successfully"
	FailedSaveRecords                 ProcessStepStatusMessage = "Failed to save medical record"
	DigitizationTaskQueue             ProcessStepStatusMessage = "Creating digitization task queue in redis server to digitize medical records"
	UserProfileNotFound               ProcessStepStatusMessage = "Failed to load user profile to create digitization task"
	GmailSearchMessage                ProcessStepStatusMessage = "Searching for health records"
	DownloadAttachmentMessage         ProcessStepStatusMessage = "Downloading all attachment that are found in email."
	DownloadAttachmentComplete        ProcessStepStatusMessage = "All attachment downloaded successfully "
	CheckDocTypeMessage               ProcessStepStatusMessage = "Checking document type"
	CheckDocTypeFailedMessage         ProcessStepStatusMessage = "Document type checking failed"
	CheckDocTypeCompleted             ProcessStepStatusMessage = "Document type checking completed"
	FailedCheckDocType                ProcessStepStatusMessage = "Failed to check document type"
	DocsDigitizationMsg               ProcessStepStatusMessage = "Processing each record and digitizing"
	DigitizationFailed                ProcessStepStatusMessage = "Digitization failed"
	CallingAIServiceMsg               ProcessStepStatusMessage = "Calling Ai service to fetch report response in json format."
	CallingAIServiceSuccess           ProcessStepStatusMessage = "AI response success"
	CallingAIFailed                   ProcessStepStatusMessage = "AI service request failed"
	MatchingNameMsg                   ProcessStepStatusMessage = "Matching Report name with self and relative names"
	CheckReportDuplicationMsg         ProcessStepStatusMessage = "Checking for report duplication based on collection date and test component name to determine if it already exists in a previous report"
	ReportDuplicationSuccess          ProcessStepStatusMessage = "Report duplication check success"
	ManualRecordUploadDigitizationMsg ProcessStepStatusMessage = "Manual record upload  and digitization process start"
)

type ProcessType string

const (
	GmailSync          ProcessType = "gmail_sync"
	DocsDigitization   ProcessType = "docs_digitization"
	ManualRecordUpload ProcessType = "manual_record_upload"
)

type EntityType string

const (
	MedicalRecordEntity EntityType = "tbl_medical_record"
)

type UserRole string

const (
	Admin      UserRole = "admin"
	Doctor     UserRole = "doctor"
	Nurse      UserRole = "nurse"
	Caregiver  UserRole = "caregiver"
	Relative   UserRole = "relative"
	Patient    UserRole = "patient"
	Pharmacist UserRole = "pharmacist"
)

var ValidUserRoles = []UserRole{
	Admin,
	Doctor,
	Nurse,
	Caregiver,
	Relative,
	Patient,
	Pharmacist,
}

const (
	PermissionViewHealth           = "view_health"
	PermissionEditProfile          = "edit_profile"
	PermissionAddFamily            = "add_family"
	PermissionScheduleAppointments = "schedule_appointments"
	PermissionAddCaregiver         = "add_caregiver"
	PermissionUploadReport         = "upload_report"
)

type PermissionMessage string

const (
	PermissionViewProfile           PermissionMessage = "You don't have permission to view profile and health info"
	PermissionEditInfo              PermissionMessage = "You don't have permission to edit profile"
	PermissionViewMedicalRecord     PermissionMessage = "You don't have permission to view medical record"
	PermissionUploadMedicalRecord   PermissionMessage = "You don't have permission to upload medical record"
	PermissionScheduleAppointment   PermissionMessage = "You don't have permission to schedule appointment"
	PermissionRescheduleAppointment PermissionMessage = "You don't have permission to reschedule appointment"
	PermissionViewAppointment       PermissionMessage = "You don't have permission to view appointment"
	PermissionManage                PermissionMessage = "You don't have permission to manage permission"
	PermissionChangeOwner           PermissionMessage = "You don't have permission to change owner of record"
	PermissionHOFAssignUnassign     PermissionMessage = "You don't have permission to assign or unassign HOF"
)
