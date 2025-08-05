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

	PatientDiagnosticTests   []PatientDiagnosticTest `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_diagnostic_test"`
	PatientReportAttachments PatientReportAttachment `gorm:"foreignKey:PatientDiagnosticReportId" json:"patient_report_attachment"`
}

func (DiagnosticLab) TableName() string {
	return "tbl_diagnostic_lab"
}

type AddLabRequest struct {
	LabNo            string   `json:"lab_no"`
	LabName          string   `json:"lab_name"`
	LabAddress       string   `json:"lab_address"`
	City             string   `json:"city"`
	State            string   `json:"state"`
	PostalCode       string   `json:"postal_code"`
	LabContactNumber string   `json:"lab_contact_number"`
	LabEmail         string   `json:"lab_email"`
	RelativeIds      []uint64 `json:"relatvies_ids"`
	IsSystemLab      bool     `json:"is_system_lab"`
	LabID            []uint64 `json:"lab_id"`
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
	PatientDiagnosticReportId uint64    `gorm:"column:patient_diagnostic_report_id;primaryKey" json:"patient_diagnostic_report_id"`
	DiagnosticLabId           uint64    `gorm:"column:diagnostic_lab_id" json:"diagnostic_lab_id"`
	PatientId                 uint64    `gorm:"column:patient_id" json:"patient_id"`
	PaymentStatus             string    `gorm:"column:payment_status" json:"payment_status"`
	ReportName                string    `gorm:"column:report_name" json:"report_name"`
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
	IsDeleted                 int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
	IsDigital                 bool      `gorm:"column:is_digital;default:false" json:"is_digital"`
	IsLabReport               bool      `gorm:"column:is_lab_report;default:false" json:"is_lab_report"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`

	PatientDiagnosticTests []PatientDiagnosticTest `gorm:"foreignKey:PatientDiagnosticReportId;references:PatientDiagnosticReportId" json:"patient_diagnostic_test"`
	DiagnosticLabs         DiagnosticLab           `gorm:"foreignKey:DiagnosticLabId;references:DiagnosticLabId" json:"diagnostic_lab"`
}

func (PatientDiagnosticReport) TableName() string {
	return "tbl_patient_diagnostic_report"
}

type DiagnosticReport struct {
	ReportName                string `json:"report_name"`
	PatientDiagnosticReportId string `json:"patient_diagnostic_report_id"`
}

type PatientDiagnosticTest struct {
	PatientTestId             uint64    `gorm:"column:patient_test_id;primaryKey;autoIncrement" json:"patient_test_id"`
	PatientDiagnosticReportId uint64    `gorm:"column:patient_diagnostic_report_id" json:"patient_diagnostic_report_id"`
	DiagnosticTestId          uint64    `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	TestNote                  string    `gorm:"column:test_note" json:"test_note"`
	TestDate                  time.Time `gorm:"column:test_date" json:"test_date"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`

	PatientDiagnosticReport PatientDiagnosticReport `gorm:"foreignKey:PatientDiagnosticReportId;references:PatientDiagnosticReportId" json:"-"`
	DiagnosticTest          DiagnosticTest          `gorm:"foreignKey:DiagnosticTestId;references:DiagnosticTestId" json:"diagnostic_test"`
}

func (PatientDiagnosticTest) TableName() string {
	return "tbl_patient_diagnostic_test"
}

type PatientReportAttachment struct {
	PatientDiagnosticReportId uint64           `gorm:"column:patient_diagnostic_report_id" json:"patient_diagnostic_report_id"`
	RecordId                  uint64           `gorm:"column:record_id" json:"record_id"`
	PatientId                 uint64           `gorm:"column:patient_id" json:"patient_id"`
	MedicalRecord             TblMedicalRecord `gorm:"foreignKey:RecordId;references:RecordId" json:"medical_report_attachment"`
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
	TestReferenceRangeId           uint64    `json:"test_reference_range_id" gorm:"primaryKey;autoIncrement"`
	DiagnosticTestId               uint64    `json:"diagnostic_test_id"`
	DiagnosticTestComponentId      uint64    `json:"diagnostic_test_component_id"`
	Age                            int       `json:"age"`
	AgeGroup                       string    `json:"age_group"`
	Gender                         string    `json:"gender"`
	NormalMin                      float64   `json:"normal_min"`
	NormalMax                      float64   `json:"normal_max"`
	BiologicalReferenceDescription *string   `json:"biological_reference_description"`
	Units                          string    `json:"units"`
	IsDeleted                      int       `json:"is_deleted"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
	CreatedBy                      string    `json:"created_by"`
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

type PatientTestReferenceRange struct {
	PatientId                 uint64     `gorm:"column:patient_id;not null"`
	DiagnosticTestID          uint64     `gorm:"column:diagnostic_test_id;not null"`
	DiagnosticTestComponentId uint64     `gorm:"column:diagnostic_test_component_id;not null"`
	NormalMin                 float64    `gorm:"column:normal_min"`
	NormalMax                 float64    `gorm:"column:normal_max"`
	BiologicalReferenceDesc   string     `gorm:"column:biological_reference_description"`
	Units                     string     `gorm:"column:units"`
	IsDeleted                 int        `gorm:"column:is_deleted;default:0"`
	CreatedAt                 time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt                 *time.Time `gorm:"column:updated_at"`
	CreatedBy                 string     `gorm:"column:created_by"`
	UpdatedBy                 string     `gorm:"column:updated_by"`
}

func (PatientTestReferenceRange) TableName() string {
	return "tbl_patient_test_reference_range"
}

type DiagnosticResultRequest struct {
	PatientId                 uint64     `json:"patient_id,omitempty"`
	PatientDiagnosticReportId *uint64    `json:"patient_diagnostic_report_id,omitempty"`
	DiagnosticTestComponentId *uint64    `json:"diagnostic_test_component_id,omitempty"`
	ReportDateStart           *time.Time `json:"report_date_start,omitempty"`
	ReportDateEnd             *time.Time `json:"report_date_end,omitempty"`
	ResultDateStart           *time.Time `json:"result_date_start,omitempty"`
	ResultDateEnd             *time.Time `json:"result_date_end,omitempty"`
	IsPinned                  *bool      `json:"is_pinned,omitempty"`
}

type ResultSummary struct {
	Summary string `json:"summary"`
}

type AudioNoteParsedData struct {
	Data struct {
		ParsedJSON    LabReport `json:"parsed_json"`
		Transcription string    `json:"transcription"`
		Type          string    `json:"type"`
	} `json:"data"`
	Message string `json:"message"`
}

type LabReport struct {
	ReportDetails struct {
		ReportName       string  `json:"report_name"`
		PatientName      string  `json:"patient_name"`
		ReportDate       string  `json:"report_date"`
		ReportTime       string  `json:"report_time"`
		CollectionDate   string  `json:"collection_date"`
		DiagnosticLabId  *uint64 `json:"diagnostic_lab_id"`
		SourceId         *uint64 `json:"source_id"`
		LabName          string  `json:"lab_name"`
		LabEmail         string  `json:"lab_email"`
		LabId            string  `json:"lab_id"`
		IsDigital        bool    `json:"is_digital"`
		IsLabReport      bool    `json:"is_lab_report"`
		IsUnknownRecord  bool    `json:"is_unknown_record"`
		IsDeleted        int     `json:"is_deleted"`
		LabLocation      string  `json:"lab_location"`
		LabContactNumber string  `json:"lab_contact_number"`
	} `json:"report_details"`
	Tests []struct {
		TestName       string `json:"test_name"`
		Interpretation string `json:"interpretation"`
		Components     []struct {
			TestComponentName              string  `json:"test_component_name"`
			ResultValue                    string  `json:"result_value"`
			Status                         string  `json:"status"`
			Units                          string  `json:"units"`
			Qualifier                      *string `json:"qualifier,omitempty"`
			BiologicalReferenceDescription *string `json:"biological_reference_description"`
			ReferenceRange                 struct {
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
	TestNote       string          `json:"test_note"`
	TestName       string          `json:"test_name"`
	TestDate       time.Time       `json:"test_date"`
	TestComponents []TestComponent `json:"test_components"`
	// DiagnosticTests []DiagnosticTestInput `json:"diagnostic_test"`
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

type TestResultAlert struct {
	ResultValue       float64 `json:"result_value"`
	NormalMin         float64 `json:"normal_min"`
	NormalMax         float64 `json:"normal_max"`
	ResultStatus      string  `json:"result_status"`
	TestName          string  `json:"test_name"`
	TestComponentName string  `json:"test_component_name"`
	ResultComment     string  `json:"result_comment"`
	ResultDate        string  `json:"result_date"`
}

type DiagnosticReportResponse struct {
	PatientDiagnosticReportID uint64  `json:"patient_diagnostic_report_id"`
	PatientID                 uint64  `json:"patient_id"`
	CollectedDate             string  `json:"collected_date"`
	ReportDate                string  `json:"report_date"`
	ReportName                string  `json:"report_name"`
	ReportStatus              string  `json:"report_status"`
	TestNote                  string  `json:"test_note"`
	TestDate                  string  `json:"test_date"`
	DiagnosticTestID          uint64  `json:"diagnostic_test_id"`
	DiagnosticTestComponentID uint64  `json:"diagnostic_test_component_id"`
	TestComponentName         string  `json:"test_component_name"`
	ResultValue               string  `json:"result_value"`
	NormalMin                 float64 `json:"normal_min"`
	NormalMax                 float64 `json:"normal_max"`
	Units                     string  `json:"units"`
	ResultStatus              string  `json:"result_status"`
	ResultDate                string  `json:"result_date"`
	ResultComment             string  `json:"result_comment"`
	DiagnosticLabID           uint64  `json:"diagnostic_lab_id"`
	LabName                   string  `json:"lab_name"`
	Qualifier                 string  `json:"qualifier"`
}

type DiagnosticReportFilter struct {
	ReportID          *string `json:"patient_diagnostic_report_id,omitempty"`
	ReportName        *string `json:"report_name,omitempty"`
	TestName          *string `json:"test_name,omitempty"`
	TestNote          *string `json:"test_note,omitempty"`
	Qualifier         *string `json:"qualifier,omitempty"`
	TestComponentName *string `json:"test_component_name,omitempty"`
	DiagnosticLabID   *uint64 `json:"diagnostic_lab_id,omitempty"`
	ReportStatus      *string `json:"report_status,omitempty"`
	ResultStatus      *string `json:"result_status,omitempty"`
	ResultDateFrom    *string `json:"from_date,omitempty"`
	ResultDateTo      *string `json:"to_date,omitempty"`
	ReportDate        *string `json:"report_date,omitempty"`
	OrderBy           *string `json:"order_by,omitempty"`
	OrderDir          *string `json:"order_dir,omitempty"`
}

type PatientTestComponentDisplayConfig struct {
	PatientId                 uint64    `gorm:"primaryKey;column:patient_id" json:"patient_id"`
	DiagnosticTestComponentId uint64    `gorm:"primaryKey;column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
	IsPinned                  *bool     `gorm:"column:is_pinned" json:"is_pinned"`
	DisplayPriority           *int      `gorm:"column:display_priority" json:"display_priority"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy                 string    `gorm:"column:created_by;" json:"created_by"`
	UpdatedBy                 string    `gorm:"column:updated_by;" json:"updated_by"`
}

func (PatientTestComponentDisplayConfig) TableName() string {
	return "tbl_patient_test_component_display_config"
}

type PatientDiagnosticLabMapping struct {
	PatientId       uint64    `gorm:"column:patient_id;not null" json:"patient_id"`
	DiagnosticLabId uint64    `gorm:"column:diagnostic_lab_id;not null" json:"diagnostic_lab_id"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy       string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy       string    `gorm:"column:updated_by" json:"updated_by"`
	IsDeleted       int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
}

func (PatientDiagnosticLabMapping) TableName() string {
	return "tbl_patient_diagnostic_lab_mapping"
}

type DiagnosticLabResponse struct {
	DiagnosticLabId  uint64    `json:"diagnostic_lab_id"`
	LabNo            string    `json:"lab_no"`
	LabName          string    `json:"lab_name"`
	LabAddress       string    `json:"lab_address"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	PostalCode       string    `json:"postal_code"`
	LabContactNumber string    `json:"lab_contact_number"`
	LabEmail         string    `json:"lab_email"`
	IsDeleted        int       `json:"is_deleted"`
	IsSystemLab      bool      `json:"is_system_lab"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedBy        string    `json:"created_by"`
	UpdatedBy        string    `json:"updated_by"`
}

type ReportRow struct {
	// Medical record fields
	RecordId       uint64
	RecordName     string
	RecordSize     int64
	FileType       string
	UploadSource   string
	SourceAccount  string
	RecordCategory string
	RecordURL      string
	DigitizeFlag   int
	Status         string

	PatientId        uint64
	CollectedAt      string
	CollectedDate    string
	ReportStatus     string
	ReportName       string
	IsPinned         bool
	Comments         string
	DiagnosticLabID  uint64
	LabNo            string
	Qualifier        string
	LabName          string
	LabAddress       string
	City             string
	State            string
	PostalCode       string
	LabContactNumber string
	LabEmail         string
	IsDeleted        int
	LabCreatedAt     string
	LabUpdatedAt     string
	LabCreatedBy     string
	LabUpdatedBy     string

	// Patient test fields
	PatientDiagnosticReportID uint64
	DiagnosticTestID          uint64
	TestNote                  string
	TestDate                  string
	TestCreatedAt             string
	TestUpdatedAt             string

	// Diagnostic test fields
	TestLoincCode   string
	TestName        string
	ReportDate      string
	TestType        string
	TestDescription string
	Category        string
	Units           string
	Property        string
	TimeAspect      string
	System          string
	Scale           string
	TestCreatedBy   string
	TestIsDeleted   int
	TestCreatedAt2  string
	TestUpdatedAt2  string

	// Test component fields
	DiagnosticTestComponentID uint64
	TestComponentLoincCode    string
	TestComponentName         string
	TestComponentType         string
	TestComponentDesc         string
	ComponentUnit             string
	ComponentProperty         string
	ComponentTimeAspect       string
	ComponentSystem           string
	ComponentScale            string
	TestComponentFrequency    string
	ComponentCreatedBy        string
	ComponentIsDeleted        int
	ComponentCreatedAt        string
	ComponentUpdatedAt        string

	// Test result value fields
	// ResultValue       interface{}
	ResultValue     string
	ResultStatus    string
	ResultDate      string
	ResultComment   string
	ResultCreatedAt string
	ResultUpdatedAt string
	Udf1            string
	Udf2            string
	Udf3            string
	Udf4            string

	// Reference range fields
	Age                            int
	AgeGroup                       string
	Gender                         string
	NormalMin                      string
	NormalMax                      string
	BiologicalReferenceDescription string
	// NormalMin            interface{}
	// NormalMax            interface{}
	RefUnits     string
	RefIsDeleted int
	RefCreatedAt string
	RefUpdatedAt string
	RefCreatedBy string
}

type ComponentKey struct {
	ComponentID uint64
	Name        string
	Units       string
	RefRange    string
	ReportName  string
	IsPinned    bool
}

type CellData struct {
	Value        string `json:"value"`
	ResultStatus string `json:"result_status"`
	ColourClass  string `json:"colour_class"`
	Colour       string `json:"colour"`
	Qualifier    string `json:"qualifier"`
	ReportId     string `json:"patient_diagnostic_report_id"`
	RecordId     string `json:"record_id"`
	ResultDate   string `json:"result_date"`
	ReportName   string `json:"report_name"`
}

type HealthVitalSource struct {
	SourceId     uint64    `json:"source_id" gorm:"primaryKey;column:source_id"`
	SourceName   string    `json:"source_name"`
	SourceTypeId uint64    `json:"-"`
	IsDeleted    int       `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

func (HealthVitalSource) TableName() string {
	return "tbl_health_vital_source"
}

type HealthVitalSourceType struct {
	SourceTypeId uint64              `json:"source_type_id" gorm:"primaryKey;column:source_type_id"`
	SourceType   string              `json:"source_type"`
	CreatedAt    time.Time           `json:"-"`
	Sources      []HealthVitalSource `json:"sources" gorm:"foreignKey:SourceTypeId;references:SourceTypeId"`
}

func (HealthVitalSourceType) TableName() string {
	return "tbl_health_vital_source_type"
}

type SampleComponentInfo struct {
	ReportID          uint64    `gorm:"column:patient_diagnostic_report_id"`
	CollectedDate     time.Time `gorm:"column:collected_date"`
	TestComponentName string    `gorm:"column:test_component_name"`
}
