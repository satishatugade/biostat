package constant

const (
	Patient             = "/registration"
	PatientInfo         = "/patient-info"
	PatientRegistration = "/patient-registration"
	PatientPrescription = "/patient-prescription"
	UpdatePrescription  = "/patient-prescription/:prescription_id"
	User                = "/getuser"
	Disease             = "/get-disease"
	Cause               = "/causes"
	AddCause            = "/add-cause"
	UpdateCause         = "/update-cause"
	DiseaseProfile      = "/disease-profile"
	DiagnosticTests     = "/diagnostic-tests"
	Medication          = "/get-medication"
	AddMedication       = "/add-medication"
	// UpdateMedication    = "/update-medication/:medication_id"
	UpdateMedication = "/update-medication"
)

const (
	Success = "success"
	Failure = "failure"
)
