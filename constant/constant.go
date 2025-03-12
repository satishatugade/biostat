package constant

const (
	Patient                         = "/registration"
	PatientInfo                     = "/patient-info"
	PatientRegistration             = "/patient-registration"
	User                            = "/getuser"
	Disease                         = "/get-disease"
	DiseaseProfile                  = "/disease-profile"
	DiagnosticTests                 = "/diagnostic-tests"
	DiagnosticTest                  = "/diagnostic-test"
	SingleDiagnosticTest            = "/diagnostic-test/:diagnosticTestId"
	DiagnosticComponents            = "/diagnostic-components"
	DiagnosticComponent             = "/diagnostic-component"
	SingleDiagnosticComponent       = "/diagnostic-component/:diagnosticComponentId"
	DiagnosticTestComponentMappings = "/diagnostic-test-component-mappings"
	DiagnosticTestComponentMapping  = "/diagnostic-test-component-mapping"
)

const (
	Success = "success"
	Failure = "failure"
)
