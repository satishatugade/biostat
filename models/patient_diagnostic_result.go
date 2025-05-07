package models

import "time"

type DiagnosticLab struct {
	DiagnosticLabId  uint64    `gorm:"column:diagnostic_lab_id;primaryKey;autoIncrement" json:"diagnostic_lab_id"`
	LabNo            string    `gorm:"column:lab_no" json:"lab_no"`
	LabName          string    `gorm:"column:lab_name" json:"lab_name"`
	LabAddress       string    `gorm:"column:lab_address" json:"lab_address"`
	City             string    `gorm:"column:city" json:"city"`
	State            string    `gorm:"column:state" json:"state"`
	PostalCode       string    `gorm:"column:postal_code" json:"postal_code"`
	LabContactNumber string    `gorm:"column:lab_contact_number" json:"lab_contact_number"`
	LabEmail         string    `gorm:"column:lab_email" json:"lab_email"`
	IsDeleted        int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy        string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy        string    `gorm:"column:updated_by" json:"updated_by"`

	PatientDiagnosticTests   []PatientDiagnosticTest   `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_diagnostic_test"`
	PatientReportAttachments []PatientReportAttachment `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_report_attachment"`
}

func (DiagnosticLab) TableName() string {
	return "tbl_diagnostic_lab"
}

type DiagnosticLabAudit struct {
	DiagnosticLabAuditId uint64    `gorm:"column:diagnostic_lab_audit_id;primaryKey;autoIncrement" json:"diagnostic_lab_audit_id"`
	DiagnosticLabId      uint64    `gorm:"column:diagnostic_lab_id" json:"diagnostic_lab_id"`
	LabNo                string    `gorm:"column:lab_no" json:"lab_no"`
	LabName              string    `gorm:"column:lab_name" json:"lab_name"`
	LabAddress           string    `gorm:"column:lab_address" json:"lab_address"`
	City                 string    `gorm:"column:city" json:"city"`
	State                string    `gorm:"column:state" json:"state"`
	PostalCode           string    `gorm:"column:postal_code" json:"postal_code"`
	LabContactNumber     string    `gorm:"column:lab_contact_number" json:"lab_contact_number"`
	LabEmail             string    `gorm:"column:lab_email" json:"lab_email"`
	IsDeleted            int       `gorm:"column:is_deleted" json:"is_deleted"`
	OperationType        string    `gorm:"column:operation_type" json:"operation_type"`
	CreatedAt            time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy            string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy            string    `gorm:"column:updated_by" json:"updated_by"`
}

func (DiagnosticLabAudit) TableName() string {
	return "tbl_diagnostic_lab_audit"
}

type PatientDiagnosticReport struct {
	PatientDiagnosticReportId uint64    `gorm:"column:patient_diagnostic_report_id;primaryKey;autoIncrement" json:"patient_diagnostic_report_id"`
	DiagnosticLabId           uint64    `gorm:"column:diagnostic_lab_id" json:"diagnostic_lab_id"`
	PatientId                 uint64    `gorm:"column:patient_id" json:"patient_id"`
	PaymentStatus             string    `gorm:"column:payment_status" json:"payment_status"`
	DoctorId                  uint64    `gorm:"column:doctor_id" json:"doctor_id"`
	CollectedDate             time.Time `gorm:"column:collected_date" json:"collected_date"`
	CollectedAt               string    `gorm:"column:collected_at" json:"collected_at"`
	ProcessedAt               string    `gorm:"column:processed_at" json:"processed_at"`
	ReportDate                time.Time `gorm:"column:report_date" json:"report_date"`
	ReportStatus              string    `gorm:"column:report_status" json:"report_status"`
	Observation               string    `gorm:"column:observation" json:"observation"`
	Comments                  string    `gorm:"column:comments" json:"comments"`
	ReviewBy                  string    `gorm:"column:review_by" json:"review_by"`
	ReviewDate                time.Time `gorm:"column:review_date" json:"review_date"`
	SharedFlag                string    `gorm:"column:shared_flag" json:"shared_flag"`
	SharedWith                string    `gorm:"column:shared_with" json:"shared_with"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`

	// PatientDiagnosticTests   []PatientDiagnosticTest   `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_diagnostic_test"`
	// PatientReportAttachments []PatientReportAttachment `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_report_attachment"`
	DiagnosticLab DiagnosticLab `gorm:"foreignKey:DiagnosticLabId;references:DiagnosticLabId"`
}

func (PatientDiagnosticReport) TableName() string {
	return "tbl_patient_diagnostic_report"
}

type PatientDiagnosticTest struct {
	PatientTestId             uint64    `gorm:"column:patient_test_id;primaryKey;autoIncrement" json:"patient_test_id"`
	PatientDiagnosticReportId uint64    `gorm:"column:patient_diagnostic_report_id" json:"patient_diagnostic_report_id"`
	DiagnosticTestId          uint64    `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	TestNote                  string    `gorm:"column:test_note" json:"test_note"`
	TestDate                  time.Time `gorm:"column:test_date" json:"test_date"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`

	DiagnosticTest DiagnosticTest `gorm:"foreignKey:DiagnosticTestId;references:DiagnosticTestId" json:"diagnostic_test"`
}

func (PatientDiagnosticTest) TableName() string {
	return "tbl_patient_diagnostic_test"
}

type PatientReportAttachment struct {
	AttachmentId              int       `gorm:"column:attachment_id;primaryKey;autoIncrement" json:"attachment_id"`
	PatientDiagnosticReportId int       `gorm:"column:patient_diagnostic_report_id" json:"patient_diagnostic_report_id"`
	FilePath                  string    `gorm:"column:file_path" json:"file_path"`
	FileType                  string    `gorm:"column:file_type" json:"file_type"`
	UploadedAt                time.Time `gorm:"column:uploaded_at" json:"uploaded_at"`
}

func (PatientReportAttachment) TableName() string {
	return "tbl_patient_report_attachment"
}

type PatientDiagnosticTestResultValue struct {
	TestResultValueId         int       `gorm:"column:test_result_value_id;primaryKey;autoIncrement" json:"test_result_value_id"`
	PatientDiagnosticReportId uint64    `gorm:"column:patient_diagnostic_report_id" json:"patient_diagnostic_report_id"`
	DiagnosticTestId          uint64    `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	PatientId                 uint64    `gorm:"column:patient_id" json:"patient_id"`
	DiagnosticTestComponentId uint64    `gorm:"column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
	ResultValue               float64   `gorm:"column:result_value" json:"result_value"`
	ResultStatus              string    `gorm:"column:result_status" json:"result_status"`
	ResultDate                time.Time `gorm:"column:result_date" json:"result_date"`
	ResultComment             string    `gorm:"column:result_comment" json:"result_comment"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`
	UDF1                      string    `gorm:"column:udf1" json:"udf1"`
	UDF2                      string    `gorm:"column:udf2" json:"udf2"`
	UDF3                      string    `gorm:"column:udf3" json:"udf3"`
	UDF4                      string    `gorm:"column:udf4" json:"udf4"`
}

func (PatientDiagnosticTestResultValue) TableName() string {
	return "tbl_patient_diagnostic_test_result_value"
}

type Diagnostic_Test_Component_ReferenceRange struct {
	TestName          string `json:"test_name"`
	TestComponentName string `json:"test_component_name"`
	DiagnosticTestReferenceRange
}

type DiagnosticTestReferenceRange struct {
	TestReferenceRangeId      uint64    `json:"test_reference_range_id" gorm:"primaryKey;autoIncrement"`
	DiagnosticTestId          uint64    `json:"diagnostic_test_id"`
	DiagnosticTestComponentId uint64    `json:"diagnostic_test_component_id"`
	Age                       int       `json:"age"`
	AgeGroup                  string    `json:"age_group"`
	Gender                    string    `json:"gender"`
	NormalMin                 float64   `json:"normal_min"`
	NormalMax                 float64   `json:"normal_max"`
	Units                     string    `json:"units"`
	IsDeleted                 int       `json:"is_deleted"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	CreatedBy                 string    `json:"created_by"`
}

func (DiagnosticTestReferenceRange) TableName() string {
	return "tbl_diagnostic_test_reference_range"
}

type DiagnosticTestReferenceRangeAudit struct {
	TestReferenceRangeAuditId uint64    `json:"test_reference_range_audit_id" gorm:"primaryKey;autoIncrement"`
	TestReferenceRangeId      uint64    `json:"test_reference_range_id"`
	DiagnosticTestId          uint64    `json:"diagnostic_test_id"`
	DiagnosticTestComponentId uint64    `json:"diagnostic_test_component_id"`
	Age                       int       `json:"age"`
	AgeGroup                  string    `json:"age_group"`
	Gender                    string    `json:"gender"`
	NormalMin                 float64   `json:"normal_min"`
	NormalMax                 float64   `json:"normal_max"`
	Units                     string    `json:"units"`
	IsDeleted                 int       `json:"is_deleted"`
	OperationType             string    `json:"operation_type"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	CreatedBy                 string    `json:"created_by"`
	UpdatedBy                 string    `json:"updated_by"`
}

func (DiagnosticTestReferenceRangeAudit) TableName() string {
	return "tbl_diagnostic_test_reference_range_audit"
}

type DiagnosticResultRequest struct {
	PatientId                 uint64     `json:"patient_id,omitempty"`
	PatientDiagnosticReportId *uint64    `json:"patient_diagnostic_report_id,omitempty"`
	DiagnosticTestComponentId *uint64    `json:"diagnostic_test_component_id,omitempty"`
	ReportDateStart           *time.Time `json:"report_date_start,omitempty"`
	ReportDateEnd             *time.Time `json:"report_date_end,omitempty"`
	ResultDateStart           *time.Time `json:"result_date_start,omitempty"`
	ResultDateEnd             *time.Time `json:"result_date_end,omitempty"`
}

type ResultSummary struct {
	Summary string `json:"summary"`
}

type LabReport struct {
	ReportDetails struct {
		ReportDate       string `json:"report_date"`
		LabName          string `json:"lab_name"`
		LabEmail         string `json:"lab_email"`
		LabID            string `json:"lab_id"`
		LabLocation      string `json:"lab_location"`
		LabContactNumber string `json:"lab_contact_number"`
	} `json:"report_details"`
	Tests []struct {
		TestName       string `json:"test_name"`
		Interpretation string `json:"interpretation"`
		Components     []struct {
			TestComponentName string `json:"test_component_name"`
			ResultValue       string `json:"result_value"`
			Units             string `json:"units"`
			ReferenceRange    struct {
				Min string `json:"min"`
				Max string `json:"max"`
			} `json:"reference_range"`
		} `json:"components"`
	} `json:"tests"`
	RawText string `json:"raw_text"`
}

type PatientData struct {
	Patient PatientBasicInfo `json:"patient"`
}

type PatientBasicInfo struct {
	PatientName string    `json:"patient_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	BloodGroup  string    `json:"blood_group"`
	Reports     []Report  `json:"reports"`
}

type Report struct {
	DiseaseName       string              `json:"disease_name"`
	PaymentStatus     string              `json:"payment_status"`
	CollectedDate     time.Time           `json:"collected_date"`
	CollectedAt       string              `json:"collected_at"`
	ProcessedAt       string              `json:"processed_at"`
	ReportDate        time.Time           `json:"report_date"`
	ReportStatus      string              `json:"report_status"`
	Observation       string              `json:"observation"`
	Comments          string              `json:"comments"`
	DiagnosticLabInfo DiagnosticLabCenter `json:"diagnostic_lab"`
}

type DiagnosticLabCenter struct {
	LabNo                 string                       `json:"lab_no"`
	LabName               string                       `json:"lab_name"`
	LabAddress            string                       `json:"lab_address"`
	LabContactNumber      string                       `json:"lab_contact_number"`
	LabEmail              string                       `json:"lab_email"`
	PatientDiagnosticTest []PatientDiagnosticTestInput `json:"patient_diagnostic_test"`
}

type PatientDiagnosticTestInput struct {
	TestNote        string                `json:"test_note"`
	TestName        string                `json:"test_name"`
	TestDate        time.Time             `json:"test_date"`
	TestComponents  []TestComponent       `json:"test_components"`
	DiagnosticTests []DiagnosticTestInput `json:"diagnostic_test"`
}

type DiagnosticTestInput struct {
	TestLoincCode   string          `json:"test_loinc_code"`
	TestName        string          `json:"test_name"`
	TestType        string          `json:"test_type"`
	TestDescription string          `json:"test_description"`
	Category        string          `json:"category"`
	Units           string          `json:"units"`
	TestComponents  []TestComponent `json:"test_components"`
}

type TestComponent struct {
	TestComponentLoincCode string            `json:"test_component_loinc_code"`
	TestComponentName      string            `json:"test_component_name"`
	TestComponentType      string            `json:"test_component_type"`
	Description            string            `json:"description"`
	Units                  string            `json:"units"`
	TestComponentFrequency string            `json:"test_component_frequency"`
	TestResultValues       []TestResultValue `json:"test_result_value"`
}

type TestResultValue struct {
	ResultValue   float64   `json:"result_value"`
	ResultStatus  string    `json:"result_status"`
	ResultDate    time.Time `json:"result_date"`
	ResultComment string    `json:"result_comment"`
	Udf1          string    `json:"udf1"`
}
