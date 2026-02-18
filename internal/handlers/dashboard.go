package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// DashboardProject represents the row returned by the dashboard query
type DashboardProject struct {
	IpCode         utils.NullString `db:"ip_code" json:"ip_code"`
	IpModel        utils.NullString `db:"ip_model" json:"ip_model"`
	IpSopYearMonth utils.NullString `db:"ip_sop_year_month" json:"ip_sop_year_month"`
	IpCustomerName utils.NullString `db:"ip_customer_name" json:"ip_customer_name"`
	IpmpName       utils.NullString `db:"ipmp_name" json:"ipmp_name"`
	IpmpStartDate  *time.Time       `db:"ipmp_start_date" json:"ipmp_start_date"`
	IpmpEndDate    *time.Time       `db:"ipmp_end_date" json:"ipmp_end_date"`
	IpmpStatus     utils.NullString `db:"ipmp_status" json:"ipmp_status"`
}

// MasterPlanSummary represents aggregated master plan statuses per project
type MasterPlanSummary struct {
	IpID                     int64            `db:"ip_id" json:"ip_id"`
	IpCustomerName           utils.NullString `db:"ip_customer_name" json:"ip_customer_name"`
	IpModel                  utils.NullString `db:"ip_model" json:"ip_model"`
	IpPartName               utils.NullString `db:"ip_part_name" json:"ip_part_name"`
	IpSopDate                *time.Time       `db:"ip_sop_date" json:"ip_sop_date"`
	KickOff                  utils.NullString `db:"Kick_Off" json:"kick_off"`
	SupplierKickOff          utils.NullString `db:"Supplier_Kick_Off" json:"supplier_kick_off"`
	MoldAndMCToolingReview   utils.NullString `db:"Mold_And_MC_Tooling_Review" json:"mold_and_mc_tooling_review"`
	MoldPO                   utils.NullString `db:"Mold_PO" json:"mold_po"`
	ToolingPO                utils.NullString `db:"Tooling_PO" json:"tooling_po"`
	OTSOffToolsSample        utils.NullString `db:"OTS_Off_Tools_Sample" json:"ots_off_tools_sample"`
	InitialPpk               utils.NullString `db:"Initial_Ppk" json:"initial_ppk"`
	OPSOffProcessSample      utils.NullString `db:"OPS_Off_Process_Sample" json:"ops_off_process_sample"`
	ResultPpkPass            utils.NullString `db:"Result_Ppk_Pass" json:"result_ppk_pass"`
	PreRAR                   utils.NullString `db:"Pre_R_A_R" json:"pre_ra_r"`
	RAR                      utils.NullString `db:"R_A_R" json:"r_a_r"`
	InternalAuditIATFSafety  utils.NullString `db:"Internal_Audit_IATF_Safety" json:"internal_audit_iatf_safety"`
	TBKKPPAPSubmitt          utils.NullString `db:"TBKK_PPAP_Submitt" json:"tbkk_ppap_submitt"`
	CustomerAuditPpap        utils.NullString `db:"Customer_Audit_ppap" json:"customer_audit_ppap"`
	CustomerPPAPApproved     utils.NullString `db:"Customer_PPAP_Approved" json:"customer_ppap_approved"`
	AssessmentProjectSignOff utils.NullString `db:"Assessment_Project_Sign_off" json:"assessment_project_sign_off"`
	PPPreProduct             utils.NullString `db:"PP_Pre_Product" json:"pp_pre_product"`
	TBKKSOPStartOfProduction utils.NullString `db:"TBKK_SOP_Start_Of_Production" json:"tbkk_sop_start_of_production"`
	InitialControl3Month     utils.NullString `db:"Initial_Control_3_Month" json:"initial_control_3_month"`
}

// ListInprogressProjects returns projects in progress with their master plan info
func ListInprogressProjects(c *fiber.Ctx, db *sqlx.DB) error {
	query := `SELECT
  ip.ip_code,
  ip.ip_model,
  DATE_FORMAT(ip.ip_sop_date, '%Y-%m') AS ip_sop_year_month,
  ip.ip_customer_name,
  ipmp.ipmp_name,
  ipmp.ipmp_start_date ,
  ipmp.ipmp_end_date ,
  ipmp.ipmp_status
FROM info_project AS ip
LEFT JOIN info_project_master_plan AS ipmp
  ON ip.ip_id = ipmp.ip_id
WHERE ip.ip_status = 'inprogress'
GROUP BY
  ip.ip_code,
  ip.ip_model,
  DATE_FORMAT(ip.ip_sop_date, '%Y-%m'),
  ip.ip_customer_name,
  ipmp.ipmp_name;`

	var rows []DashboardProject
	if err := db.Select(&rows, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(rows)
}

// ListMasterPlanSummary returns a pivoted view of master plan statuses per project
func ListMasterPlanSummary(c *fiber.Ctx, db *sqlx.DB) error {
	query := `SELECT
	ipmp.ip_id,
	ip.ip_customer_name,
	ip.ip_model,
	ip.ip_part_name,
	ip.ip_sop_date,
	MAX(CASE WHEN ipmp.ipmp_name = 'Kick Off' THEN ipmp.ipmp_status END) AS Kick_Off,

	MAX(CASE WHEN ipmp.ipmp_name = 'Supplier Kick Off' THEN ipmp.ipmp_status END) AS Supplier_Kick_Off,

	MAX(CASE WHEN ipmp.ipmp_name = 'Mold & M/C Tooling Review' THEN ipmp.ipmp_status END) AS Mold_And_MC_Tooling_Review,

	MAX(CASE WHEN ipmp.ipmp_name = 'Mold PO' THEN ipmp.ipmp_status END) AS Mold_PO,

	MAX(CASE WHEN ipmp.ipmp_name = 'Tooling PO' THEN ipmp.ipmp_status END) AS Tooling_PO,

	MAX(CASE WHEN ipmp.ipmp_name = 'OTS : Off Tools Sample' THEN ipmp.ipmp_status END) AS OTS_Off_Tools_Sample,

	MAX(CASE WHEN ipmp.ipmp_name = 'Initial Ppk' THEN ipmp.ipmp_status END) AS Initial_Ppk,

	MAX(CASE WHEN ipmp.ipmp_name = 'OPS : Off Process Sample' THEN ipmp.ipmp_status END) AS OPS_Off_Process_Sample,

	MAX(CASE WHEN ipmp.ipmp_name = 'Result Ppk Pass' THEN ipmp.ipmp_status END) AS Result_Ppk_Pass,

	MAX(CASE WHEN ipmp.ipmp_name = 'Pre-R@R' THEN ipmp.ipmp_status END) AS Pre_R_A_R,

	MAX(CASE WHEN ipmp.ipmp_name = 'R@R' THEN ipmp.ipmp_status END) AS R_A_R,

	MAX(CASE WHEN ipmp.ipmp_name = 'Internal Audit IATF & Safety' THEN ipmp.ipmp_status END) AS Internal_Audit_IATF_Safety,

	MAX(CASE WHEN ipmp.ipmp_name = 'TBKK PPAP Submitt' THEN ipmp.ipmp_status END) AS TBKK_PPAP_Submitt,

	MAX(CASE WHEN ipmp.ipmp_name = 'Customer Audit ppap' THEN ipmp.ipmp_status END) AS Customer_Audit_ppap,

	MAX(CASE WHEN ipmp.ipmp_name = 'Customer PPAP Approved' THEN ipmp.ipmp_status END) AS Customer_PPAP_Approved,

	MAX(CASE WHEN ipmp.ipmp_name = 'Assessment (Project  Sign-off)' THEN ipmp.ipmp_status END) AS Assessment_Project_Sign_off,

	MAX(CASE WHEN ipmp.ipmp_name = 'PP : Pre Product' THEN ipmp.ipmp_status END) AS PP_Pre_Product,

	MAX(CASE WHEN ipmp.ipmp_name = 'TBKK SOP : Start Of Production' THEN ipmp.ipmp_status END) AS TBKK_SOP_Start_Of_Production,

	MAX(CASE WHEN ipmp.ipmp_name = 'Initial Control 3 Month' THEN ipmp.ipmp_status END) AS Initial_Control_3_Month

FROM info_project_master_plan ipmp
LEFT JOIN info_project ip ON ip.ip_id = ipmp.ip_id
GROUP BY
	ipmp.ip_id,
	ip.ip_part_no;`

	var rows []MasterPlanSummary
	if err := db.Select(&rows, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(rows)
}
