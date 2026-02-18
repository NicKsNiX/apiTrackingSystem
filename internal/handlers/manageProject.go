package handlers

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	// try several common layouts: date-only, space-separated datetime, RFC3339, RFC3339Nano
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	var lastErr error
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			d.Time = t
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("invalid date format: %s (%v)", s, lastErr)
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte("\"" + d.Time.Format("2006-01-02") + "\""), nil
}

func (d *Date) Scan(src interface{}) error {
	if src == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		d.Time = v
		return nil
	case []byte:
		s := string(v)
		if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
			d.Time = t
			return nil
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			d.Time = t
			return nil
		}
		return fmt.Errorf("cannot scan Date from bytes: %s", s)
	case string:
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			d.Time = t
			return nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			d.Time = t
			return nil
		}
		return fmt.Errorf("cannot scan Date from string: %s", v)
	default:
		return fmt.Errorf("unsupported scan source for Date: %T", src)
	}
}

func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	// return time.Time value; driver will format as DATE when column type is DATE
	return d.Time, nil
}

type DateTime struct {
	time.Time
}

type StringOrArray []string

func (s *StringOrArray) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || len(b) == 0 {
		*s = nil
		return nil
	}
	// try array first
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		*s = arr
		return nil
	}
	// try single string
	var single string
	if err := json.Unmarshal(b, &single); err == nil {
		*s = []string{single}
		return nil
	}
	return fmt.Errorf("invalid StringOrArray JSON: %s", string(b))
}

func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		dt.Time = time.Time{}
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339Nano,
	}
	var lastErr error
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			dt.Time = t
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("invalid datetime format: %s (%v)", s, lastErr)
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte("\"" + dt.Time.Format(time.RFC3339) + "\""), nil
}

func (dt *DateTime) Scan(src interface{}) error {
	if src == nil {
		dt.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		dt.Time = v
		return nil
	case []byte:
		s := string(v)
		// try RFC3339 then space-separated
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			dt.Time = t
			return nil
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
			dt.Time = t
			return nil
		}
		return fmt.Errorf("cannot scan DateTime from bytes: %s", s)
	case string:
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			dt.Time = t
			return nil
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local); err == nil {
			dt.Time = t
			return nil
		}
		return fmt.Errorf("cannot scan DateTime from string: %s", v)
	default:
		return fmt.Errorf("unsupported scan source for DateTime: %T", src)
	}
}

func (dt DateTime) Value() (driver.Value, error) {
	if dt.Time.IsZero() {
		return nil, nil
	}
	return dt.Time, nil
}

type SysInfoProject struct {
	MmmID        int64            `json:"mmm_id" db:"mmm_id"`
	MdtID        int64            `json:"mdt_id" db:"mdt_id"`
	IpID         int64            `json:"ip_id" db:"ip_id"`
	Model        utils.NullString `json:"ip_model" db:"ip_model"`
	PartNo       utils.NullString `json:"ip_part_no" db:"ip_part_no"`
	PartName     utils.NullString `json:"ip_part_name" db:"ip_part_name"`
	KickoffDate  *Date            `json:"ip_kickoff_date" db:"ip_kickoff_date"`
	SopDate      *Date            `json:"ip_sop_date" db:"ip_sop_date"`
	Code         utils.NullString `json:"ip_code" db:"ip_code"`
	Pos          utils.NullString `json:"ip_pos" db:"ip_pos"`
	DocumentNo   utils.NullString `json:"ip_document_no" db:"ip_document_no"`
	Revision     int              `json:"ip_revision" db:"ip_revision"`
	Status       utils.NullString `json:"ip_status" db:"ip_status"`
	CreatedBy    utils.NullString `json:"ip_created_by" db:"ip_created_by"`
	UpdatedBy    utils.NullString `json:"ip_updated_by" db:"ip_updated_by"`
	UpdatedAt    *DateTime        `json:"ip_updated_at" db:"ip_updated_at"`
	CreatedAt    *DateTime        `json:"ip_created_at" db:"ip_created_at"`
	CustomerName utils.NullString `json:"ip_customer_name" db:"ip_customer_name"`
	UpdatedByFN  utils.NullString `json:"ip_updated_by_firstname" db:"ip_updated_by_firstname"`
	UpdatedByLN  utils.NullString `json:"ip_updated_by_lastname" db:"ip_updated_by_lastname"`

	MtID   int64            `json:"mt_id" db:"mt_id"`
	MtName utils.NullString `json:"mt_name" db:"mt_name"`

	IceKKen1Date *Date `json:"ice_k_ken_1_date" db:"ice_k_ken_1_date"`
	IceKKen2Date *Date `json:"ice_k_ken_2_date" db:"ice_k_ken_2_date"`
	Ice1PPDate   *Date `json:"ice_1pp_date" db:"ice_1pp_date"`
	Ice2PPDate   *Date `json:"ice_2pp_date" db:"ice_2pp_date"`
	IcePPDate    *Date `json:"ice_pp_date" db:"ice_pp_date"`
	IceSopDate   *Date `json:"ice_sop_date" db:"ice_sop_date"`

	// NEW: end-date fields added to info_customer_event table
	IceKKen1EndDate *Date            `json:"ice_k_ken_1_end_date" db:"ice_k_ken_1_end_date"`
	IceKKen2EndDate *Date            `json:"ice_k_ken_2_end_date" db:"ice_k_ken_2_end_date"`
	Ice1PPEndDate   *Date            `json:"ice_1pp_end_date" db:"ice_1pp_end_date"`
	Ice2PPEndDate   *Date            `json:"ice_2pp_end_date" db:"ice_2pp_end_date"`
	IcePPEndDate    *Date            `json:"ice_pp_end_date" db:"ice_pp_end_date"`
	IceSopEndDate   *Date            `json:"ice_sop_end_date" db:"ice_sop_end_date"`
	IceUpdatedAt    *DateTime        `json:"ice_updated_at" db:"ice_updated_at"`
	IceUpdatedBy    utils.NullString `json:"ice_updated_by" db:"ice_updated_by"`

	IpfFileName       utils.NullString `json:"ipf_file_name" db:"ipf_file_name"`
	IpfFilePath       utils.NullString `json:"ipf_file_path" db:"ipf_file_path"`
	IpfUpdatedAt      *DateTime        `json:"ipf_updated_at" db:"ipf_updated_at"`
	IpfUpdatedBy      utils.NullString `json:"ipf_updated_by" db:"ipf_updated_by"`
	TrackingFileCount utils.NullString `json:"tracking_file_count" db:"tracking_file_count"`
}

func parseDatePtr(s string) *Date {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var d Date
	// try UnmarshalJSON style (wrap with quotes)
	if err := d.UnmarshalJSON([]byte(`"` + s + `"`)); err == nil {
		return &d
	}
	// fallback parse
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		d.Time = t
		return &d
	}
	return nil
}

func ListInfoProjects(c *fiber.Ctx, db *sqlx.DB) error {
	status := c.Query("mmm_id")
	query := `SELECT DISTINCT
					mmm.mmm_id AS mmm_id,
					info_project.ip_id AS ip_id,
					info_project.mdt_id AS mdt_id,
					info_project.ip_model AS ip_model,
					info_project.ip_part_no AS ip_part_no,
					info_project.ip_part_name AS ip_part_name,
					info_project.ip_kickoff_date AS ip_kickoff_date,
					info_project.ip_sop_date AS ip_sop_date,
					info_project.ip_code AS ip_code,
					info_project.ip_pos AS ip_pos,
					info_project.ip_document_no AS ip_document_no,
					info_project.ip_revision AS ip_revision,
					info_project.ip_status AS ip_status,
					info_project.ip_updated_at AS ip_updated_at,
					info_project.ip_updated_by AS ip_updated_by,
					info_project.ip_customer_name AS ip_customer_name,
					su.su_firstname AS ip_updated_by_firstname,
					su.su_lastname AS ip_updated_by_lastname,
					ice_k_ken_1_date AS ice_k_ken_1_date,
					ice_k_ken_2_date AS ice_k_ken_2_date,
					ice_1pp_date AS ice_1pp_date,
					ice_2pp_date AS ice_2pp_date,
					ice_pp_date AS ice_pp_date,
					ice_sop_date AS ice_sop_date,
					info_project.mt_id AS mt_id,
					ipf.ipf_file_name AS ipf_file_name,
					ipf.ipf_file_path AS ipf_file_path,
					mt.mt_name AS mt_name,
					ice.ice_k_ken_1_end_date AS ice_k_ken_1_end_date,
					ice.ice_k_ken_2_end_date AS ice_k_ken_2_end_date,
					ice.ice_1pp_end_date AS ice_1pp_end_date,
					ice.ice_2pp_end_date AS ice_2pp_end_date,
					ice.ice_pp_end_date AS ice_pp_end_date,
					ice.ice_sop_end_date AS ice_sop_end_date,
					IFNULL(itf.tracking_file_count, 0) AS tracking_file_count
				FROM info_project 
				LEFT JOIN sys_user su ON ip_updated_by = su_emp_code
				LEFT JOIN info_customer_event ice ON info_project.ip_id = ice.ip_id
				LEFT JOIN info_pos_file ipf ON info_project.ip_id = ipf.ip_id
				LEFT JOIN mst_template mt ON info_project.mt_id = mt.mt_id
				LEFT JOIN mst_model_master mmm ON info_project.ip_customer_name = mmm.mmm_customer_name
				LEFT JOIN (
                    SELECT 
                        ip_id,
                        COUNT(*) AS tracking_file_count
                    FROM info_tracking_file
					LEFT JOIN info_approval ia ON info_tracking_file.ipid_id = ia.ipid_id
					LEFT JOIN info_project_item_detail pid ON info_tracking_file.ipid_id = pid.ipid_id
					WHERE ia.ia_status = 'waiting' AND ia.ia_type = 'PJ' AND pid.ipid_status = 'waiting'
                    GROUP BY ip_id
				) itf ON info_project.ip_id = itf.ip_id
				WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		query += " AND mmm_id = ?"
		args = append(args, status)
	}
	query += " ORDER BY ip_id ASC"

	var list []SysInfoProject
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func IssueInfoProject(c *fiber.Ctx, db *sqlx.DB) error {
	var req SysInfoProject

	// -------------------------
	// 1) Parse request (multipart or JSON)
	// -------------------------
	if strings.Contains(strings.ToLower(c.Get("Content-Type")), "multipart/form-data") {
		// parse simple scalar form values into req
		if v := strings.TrimSpace(c.FormValue("mt_id")); v != "" {
			if id, err := strconv.ParseInt(v, 10, 64); err == nil {
				req.MtID = id
			}
		}

		// optional: mdt_id provided by front for document type
		if v := strings.TrimSpace(c.FormValue("mdt_id")); v != "" {
			if id, err := strconv.ParseInt(v, 10, 64); err == nil {
				req.MdtID = id
			}
		}

		req.Model = utils.NewNullString(c.FormValue("ip_model"))
		req.CustomerName = utils.NewNullString(c.FormValue("ip_customer_name"))
		req.PartNo = utils.NewNullString(c.FormValue("ip_part_no"))
		req.PartName = utils.NewNullString(c.FormValue("ip_part_name"))
		req.KickoffDate = parseDatePtr(c.FormValue("ip_kickoff_date"))
		req.SopDate = parseDatePtr(c.FormValue("ip_sop_date"))
		req.Code = utils.NewNullString(c.FormValue("ip_code"))
		req.Pos = utils.NewNullString(c.FormValue("ip_pos"))
		req.CreatedBy = utils.NewNullString(c.FormValue("ip_created_by"))

		req.IceKKen1Date = parseDatePtr(c.FormValue("ice_k_ken_1_date"))
		req.IceKKen2Date = parseDatePtr(c.FormValue("ice_k_ken_2_date"))
		req.Ice1PPDate = parseDatePtr(c.FormValue("ice_1pp_date"))
		req.Ice2PPDate = parseDatePtr(c.FormValue("ice_2pp_date"))
		req.IcePPDate = parseDatePtr(c.FormValue("ice_pp_date"))
		req.IceSopDate = parseDatePtr(c.FormValue("ice_sop_date"))

		// parse new end-date fields (if provided)
		req.IceKKen1EndDate = parseDatePtr(c.FormValue("ice_k_ken_1_end_date"))
		req.IceKKen2EndDate = parseDatePtr(c.FormValue("ice_k_ken_2_end_date"))
		req.Ice1PPEndDate = parseDatePtr(c.FormValue("ice_1pp_end_date"))
		req.Ice2PPEndDate = parseDatePtr(c.FormValue("ice_2pp_end_date"))
		req.IcePPEndDate = parseDatePtr(c.FormValue("ice_pp_end_date"))
		req.IceSopEndDate = parseDatePtr(c.FormValue("ice_sop_end_date"))

		// optional POS file metadata
		req.IpfFileName = utils.NewNullString(c.FormValue("ipf_file_name"))

		// IMPORTANT: ignore ipf_file_path from client
		// TH: ห้ามเชื่อ path จากฝั่ง client เพราะ backend เปิดจากเครื่อง client ไม่ได้
		req.IpfFilePath = utils.NewNullString("")
	} else {
		// fallback to JSON body
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
		}
	}

	// -------------------------
	// 2) Duplicate check by ip_code
	// -------------------------
	var exists int
	if err := db.Get(&exists, `SELECT 1 FROM info_project WHERE ip_code = ? LIMIT 1`, req.Code); err != nil && err != sql.ErrNoRows {
		return c.Status(500).JSON(5.1)
	}
	if exists == 1 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	// -------------------------
	// 3) Transaction
	// -------------------------
	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5.2)
	}
	defer func() { _ = tx.Rollback() }()

	// -------------------------
	// 4) Insert info_project
	// -------------------------
	projectParams := map[string]any{
		"mdt_id":           req.MdtID,
		"mt_id":            req.MtID,
		"ip_model":         req.Model,
		"ip_part_no":       req.PartNo,
		"ip_part_name":     req.PartName,
		"ip_kickoff_date":  req.KickoffDate,
		"ip_sop_date":      req.SopDate,
		"ip_code":          req.Code,
		"ip_pos":           req.Pos,
		"ip_document_no":   req.DocumentNo,
		"ip_created_at":    now,
		"ip_created_by":    req.CreatedBy,
		"ip_updated_at":    now,
		"ip_updated_by":    req.CreatedBy,
		"ip_customer_name": req.CustomerName,
	}

	res, err := tx.NamedExec(`
		INSERT INTO info_project
			(mdt_id, ip_model, ip_part_no, ip_part_name, ip_kickoff_date, ip_sop_date, ip_code, ip_pos,
			 ip_document_no, ip_revision, ip_status, ip_created_at, ip_created_by, ip_updated_at, ip_updated_by, ip_customer_name, mt_id)
		VALUES
			(:mdt_id, :ip_model, :ip_part_no, :ip_part_name, :ip_kickoff_date, :ip_sop_date, :ip_code, :ip_pos,
			 '-' , 0, 'added', :ip_created_at, :ip_created_by, :ip_updated_at, :ip_updated_by, :ip_customer_name, :mt_id)
	`, projectParams)
	if err != nil {
		return c.Status(500).JSON(5.3)
	}

	ipID, err := res.LastInsertId()
	if err != nil {
		return c.Status(500).JSON(5.4)
	}

	// If front provided an ip_code like PJC-26-001, split and insert into info_doc_run_no
	if req.Code.Valid && strings.TrimSpace(req.Code.String) != "" {
		parts := strings.SplitN(req.Code.String, "-", 3)
		if len(parts) == 3 {
			pos1 := strings.TrimSpace(parts[0])
			pos2 := strings.TrimSpace(parts[1])
			lastSeq := strings.TrimSpace(parts[2])
			// normalize parts to safe chars
			norm := regexp.MustCompile(`[^A-Za-z0-9_-]`)
			pos1 = norm.ReplaceAllString(pos1, "_")
			pos2 = norm.ReplaceAllString(pos2, "_")
			lastSeq = norm.ReplaceAllString(lastSeq, "_")

			_, err = tx.NamedExec(`
				INSERT INTO info_doc_run_no (mdt_id, idrn_pos_1, idrn_pos_2, idrn_last_seq, idrn_created_at, idrn_created_by, idrn_updated_at, idrn_updated_by)
				VALUES (:mdt_id, :idrn_pos_1, :idrn_pos_2, :idrn_last_seq, :idrn_created_at, :idrn_created_by, :idrn_updated_at, :idrn_updated_by)
			`, map[string]any{
				"mdt_id":          req.MdtID,
				"idrn_pos_1":      pos1,
				"idrn_pos_2":      pos2,
				"idrn_last_seq":   lastSeq,
				"idrn_created_at": now,
				"idrn_created_by": req.CreatedBy,
				"idrn_updated_at": now,
				"idrn_updated_by": req.CreatedBy,
			})
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "insert info_doc_run_no failed", "detail": err.Error()})
			}
		}
	}

	// -------------------------
	// 5) Insert info_customer_event
	// -------------------------
	eventParams := map[string]any{
		"ip_id":                ipID,
		"ice_status":           "active",
		"ice_created_at":       now,
		"ice_created_by":       req.CreatedBy,
		"ice_updated_at":       now,
		"ice_updated_by":       req.CreatedBy,
		"ice_k_ken_1_date":     req.IceKKen1Date,
		"ice_k_ken_2_date":     req.IceKKen2Date,
		"ice_1pp_date":         req.Ice1PPDate,
		"ice_2pp_date":         req.Ice2PPDate,
		"ice_pp_date":          req.IcePPDate,
		"ice_sop_date":         req.IceSopDate,
		"ice_k_ken_1_end_date": req.IceKKen1EndDate,
		"ice_k_ken_2_end_date": req.IceKKen2EndDate,
		"ice_1pp_end_date":     req.Ice1PPEndDate,
		"ice_2pp_end_date":     req.Ice2PPEndDate,
		"ice_pp_end_date":      req.IcePPEndDate,
		"ice_sop_end_date":     req.IceSopEndDate,
	}

	_, err = tx.NamedExec(`
		INSERT INTO info_customer_event
			(ip_id, ice_status, ice_created_at, ice_created_by, ice_updated_at, ice_updated_by,
			 ice_k_ken_1_date, ice_k_ken_2_date, ice_1pp_date, ice_2pp_date, ice_pp_date, ice_sop_date,
			 ice_k_ken_1_end_date, ice_k_ken_2_end_date, ice_1pp_end_date, ice_2pp_end_date, ice_pp_end_date, ice_sop_end_date)
		VALUES
			(:ip_id, :ice_status, :ice_created_at, :ice_created_by, :ice_created_at, :ice_created_by,
			 :ice_k_ken_1_date, :ice_k_ken_2_date, :ice_1pp_date, :ice_2pp_date, :ice_pp_date, :ice_sop_date,
			 :ice_k_ken_1_end_date, :ice_k_ken_2_end_date, :ice_1pp_end_date, :ice_2pp_end_date, :ice_pp_end_date, :ice_sop_end_date)
	`, eventParams)
	if err != nil {
		return c.Status(500).JSON(5.5)
	}

	uploadBase := os.Getenv("UPLOAD_BASE")
	if strings.TrimSpace(uploadBase) == "" {
		uploadBase = `C:\inetpub\wwwroot\apiTrackingSystemUat\uploads`
	}
	var dbFileName string // ex: NORAPHATsongkran
	var dbFilePath string // ex: uploads/<ip_code>/pos/NORAPHATsongkran.pdf

	// determine folder name from ip_code (fallback to ip_<id>) and ensure folder exists
	ipCodeFolder := ""
	if req.Code.Valid && strings.TrimSpace(req.Code.String) != "" {
		ipCodeFolder = req.Code.String
	} else {
		ipCodeFolder = fmt.Sprintf("ip_%d", ipID)
	}
	ipCodeFolder = regexp.MustCompile(`[^A-Za-z0-9_-]`).ReplaceAllString(ipCodeFolder, "_")
	if ipCodeFolder == "" {
		ipCodeFolder = fmt.Sprintf("ip_%d", ipID)
	}
	// always create uploads/<ip_code>/pos folder so UI can rely on it existing
	folderPath := filepath.Join(uploadBase, ipCodeFolder, "pos")
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "create upload dir failed", "detail": err.Error()})
	}

	formFile, errFile := c.FormFile("pos")
	if errFile != nil || formFile == nil {
		// accept alternative field name used by some clients
		formFile, errFile = c.FormFile("pos_file")
	}
	if errFile == nil && formFile != nil {
		uploadedName := filepath.Base(formFile.Filename)
		ext := strings.ToLower(filepath.Ext(uploadedName))

		noExt := strings.TrimSuffix(uploadedName, filepath.Ext(uploadedName))
		parts := strings.Split(noExt, "_")
		if len(parts) >= 2 {
			dbFileName = strings.TrimSpace(parts[1])
		} else {
			dbFileName = strings.TrimSpace(noExt)
		}

		dbFileName = regexp.MustCompile(`[^A-Za-z0-9_-]`).ReplaceAllString(dbFileName, "_")
		if dbFileName == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid filename (empty name after normalize)"})
		}

		// Save file on disk using normalized name inside existing folderPath
		savedDiskPath := filepath.Join(folderPath, dbFileName+ext)
		if err := c.SaveFile(formFile, savedDiskPath); err != nil {
			log.Printf("save pos file failed: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "save uploaded file failed", "detail": err.Error()})
		}

		// set DB path relative to uploads root (use forward slashes)
		dbFilePath = filepath.ToSlash(filepath.Join("uploads", ipCodeFolder, "pos", dbFileName+ext))

		// Set ipf_file_name in request (DB)
		req.IpfFileName = utils.NewNullString(dbFileName)
	}

	// -------------------------
	// 7) Insert info_pos_file (store normalized name/path)
	// -------------------------
	if dbFilePath != "" {
		fileParams := map[string]any{
			"ipf_file_name":  req.IpfFileName,
			"ipf_file_path":  dbFilePath,
			"ipf_status":     "active",
			"ipf_created_at": now,
			"ipf_created_by": req.CreatedBy,
			"ip_id":          ipID,
			"ipf_updated_at": now,
			"ipf_updated_by": req.CreatedBy,
		}

		_, err = tx.NamedExec(`
		INSERT INTO info_pos_file
			(ipf_file_name, ipf_file_path, ipf_status, ipf_created_at, ipf_created_by, ip_id, ipf_updated_at, ipf_updated_by)
		VALUES
			(:ipf_file_name, :ipf_file_path, :ipf_status, :ipf_created_at, :ipf_created_by, :ip_id, :ipf_updated_at, :ipf_updated_by)
	`, fileParams)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "insert info_pos_file failed", "detail": err.Error()})
		}
	}

	// -------------------------
	// 8) Commit
	// -------------------------
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5.7)
	}

	// (Optional) return web path for UI use
	// TH: จะส่ง path กลับให้ UI ก็ได้
	return c.Status(201).JSON(1)
}

func UpdateInfoProject(c *fiber.Ctx, db *sqlx.DB) error {
	var req SysInfoProject
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	if req.IpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id is required"})
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5.8)
	}
	defer func() { _ = tx.Rollback() }()

	// update info_project by ip_id
	projectParams := map[string]any{
		"ip_id":            req.IpID,
		"mt_id":            req.MtID,
		"ip_model":         req.Model,
		"ip_part_no":       req.PartNo,
		"ip_part_name":     req.PartName,
		"ip_kickoff_date":  req.KickoffDate,
		"ip_sop_date":      req.SopDate,
		"ip_code":          req.Code,
		"ip_pos":           req.Pos,
		"ip_document_no":   "-",
		"ip_updated_at":    now,
		"ip_updated_by":    req.UpdatedBy,
		"ip_customer_name": req.CustomerName,
	}

	_, err = tx.NamedExec(`
		UPDATE info_project SET
			ip_model = :ip_model,
			ip_part_no = :ip_part_no,
			ip_part_name = :ip_part_name,
			ip_kickoff_date = :ip_kickoff_date,
			ip_sop_date = :ip_sop_date,
			ip_code = :ip_code,
			ip_pos = :ip_pos,
			ip_document_no = :ip_document_no,
			ip_updated_at = :ip_updated_at,
			ip_updated_by = :ip_updated_by,
			ip_customer_name = :ip_customer_name,
			mt_id = :mt_id
		WHERE ip_id = :ip_id
	`, projectParams)
	if err != nil {
		return c.Status(500).JSON(5.9)
	}

	// upsert info_customer_event for this ip_id (includes new end-date fields)
	eventParams := map[string]any{
		"ip_id":                req.IpID,
		"ice_k_ken_1_date":     req.IceKKen1Date,
		"ice_k_ken_2_date":     req.IceKKen2Date,
		"ice_1pp_date":         req.Ice1PPDate,
		"ice_2pp_date":         req.Ice2PPDate,
		"ice_pp_date":          req.IcePPDate,
		"ice_sop_date":         req.IceSopDate,
		"ice_k_ken_1_end_date": req.IceKKen1EndDate,
		"ice_k_ken_2_end_date": req.IceKKen2EndDate,
		"ice_1pp_end_date":     req.Ice1PPEndDate,
		"ice_2pp_end_date":     req.Ice2PPEndDate,
		"ice_pp_end_date":      req.IcePPEndDate,
		"ice_sop_end_date":     req.IceSopEndDate,
		"ice_updated_at":       now,
		"ice_updated_by":       req.UpdatedBy,
		"ice_status":           "active",
	}
	// log.Printf("UpdateInfoProject - eventParams: %+v", eventParams)

	res, err := tx.NamedExec(`
		UPDATE info_customer_event SET
			ice_k_ken_1_date = :ice_k_ken_1_date,
			ice_k_ken_2_date = :ice_k_ken_2_date,
			ice_1pp_date = :ice_1pp_date,
			ice_2pp_date = :ice_2pp_date,
			ice_pp_date = :ice_pp_date,
			ice_sop_date = :ice_sop_date,
			ice_k_ken_1_end_date = :ice_k_ken_1_end_date,
			ice_k_ken_2_end_date = :ice_k_ken_2_end_date,
			ice_1pp_end_date = :ice_1pp_end_date,
			ice_2pp_end_date = :ice_2pp_end_date,
			ice_pp_end_date = :ice_pp_end_date,
			ice_sop_end_date = :ice_sop_end_date,
			ice_updated_at = :ice_updated_at,
			ice_updated_by = :ice_updated_by,
			ice_status = :ice_status
		WHERE ip_id = :ip_id
	`, eventParams)
	if err != nil {
		// log.Printf("UpdateInfoProject - update info_customer_event error: %v", err)
		return c.Status(500).JSON(5.10)
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		// insert if no existing event
		// reuse ice_updated_at/ice_updated_by as created values for simplicity
		_, err = tx.NamedExec(`
			INSERT INTO info_customer_event
				(ip_id, ice_status, ice_created_at, ice_created_by,
				 ice_k_ken_1_date, ice_k_ken_2_date, ice_1pp_date, ice_2pp_date, ice_pp_date, ice_sop_date,
				 ice_k_ken_1_end_date, ice_k_ken_2_end_date, ice_1pp_end_date, ice_2pp_end_date, ice_pp_end_date, ice_sop_end_date)
			VALUES
				(:ip_id, :ice_status, :ice_updated_at, :ice_updated_by,
				 :ice_k_ken_1_date, :ice_k_ken_2_date, :ice_1pp_date, :ice_2pp_date, :ice_pp_date, :ice_sop_date,
				 :ice_k_ken_1_end_date, :ice_k_ken_2_end_date, :ice_1pp_end_date, :ice_2pp_end_date, :ice_pp_end_date, :ice_sop_end_date)
		`, eventParams)
		if err != nil {
			log.Printf("UpdateInfoProject - insert info_customer_event error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "insert info_customer_event failed", "detail": err.Error()})
		}
	}

	// upsert info_pos_file if a file path is provided
	if req.IpfFilePath.Valid && req.IpfFilePath.String != "" {
		fileParams := map[string]any{
			"ip_id":          req.IpID,
			"ipf_file_name":  req.IpfFileName,
			"ipf_file_path":  req.IpfFilePath,
			"ipf_status":     "active",
			"ipf_updated_at": now,
			"ipf_updated_by": req.UpdatedBy,
			// include created values so the same map works for INSERT fallback
			"ipf_created_at": now,
			"ipf_created_by": req.CreatedBy,
		}

		res, err := tx.NamedExec(`
			UPDATE info_pos_file SET
				ipf_file_name = :ipf_file_name,
				ipf_file_path = :ipf_file_path,
				ipf_status = :ipf_status,
				ipf_updated_at = :ipf_updated_at,
				ipf_updated_by = :ipf_updated_by
			WHERE ip_id = :ip_id
		`, fileParams)
		if err != nil {
			log.Printf("UpdateInfoProject - update info_pos_file error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "update info_pos_file failed", "detail": err.Error()})
		}
		if ra, _ := res.RowsAffected(); ra == 0 {
			_, err = tx.NamedExec(`
				INSERT INTO info_pos_file
					(ipf_file_name, ipf_file_path, ipf_status, ipf_created_at, ipf_created_by, ip_id)
				VALUES
					(:ipf_file_name, :ipf_file_path, :ipf_status, :ipf_created_at, :ipf_created_by, :ip_id)
			`, fileParams)
			if err != nil {
				log.Printf("UpdateInfoProject - insert info_pos_file error: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": "insert info_pos_file failed", "detail": err.Error()})
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5.11)
	}

	return c.Status(200).JSON(1)
}

func SelectModel(c *fiber.Ctx, db *sqlx.DB) error {
	var models []struct {
		ID           int64            `db:"mmm_id" json:"mmm_id"`
		Model        string           `db:"mmm_model" json:"mmm_model"`
		CustomerName utils.NullString `db:"mmm_customer_name" json:"mmm_customer_name"`
	}
	if err := db.Select(&models, `SELECT mmm_id, mmm_model, mmm_customer_name FROM mst_model_master WHERE mmm_status = 'active' ORDER BY mmm_model ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(models)
}

func SelectPartNumber(c *fiber.Ctx, db *sqlx.DB) error {
	var parts []struct {
		ID     int64  `db:"ip_id" json:"ip_id"`
		PartNo string `db:"ip_part_no" json:"ip_part_no"`
	}
	if err := db.Select(&parts, `SELECT ip_part_no FROM info_project WHERE ip_status = 'active' ORDER BY ip_part_no ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(parts)
}

func GetMaxCode(c *fiber.Ctx, db *sqlx.DB) error {
	var res struct {
		MaxCode sql.NullString `db:"max_code"`
		MaxPos  sql.NullString `db:"max_pos"`
	}
	if err := db.Get(&res, `SELECT MAX(ip_code) as max_code , MAX(ip_pos) as max_pos FROM info_project`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"max_code": res.MaxCode.String, "max_pos": res.MaxPos.String})
}

func UpdateStatusProject(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"ip_id"`
		Status    string `json:"ip_status"`
		UpdatedBy string `json:"ip_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE info_project SET ip_status = ?, ip_updated_at = ?, ip_updated_by = ? WHERE ip_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}
	return c.Status(200).JSON(1)
}

func SelectCountAddedStatus(c *fiber.Ctx, db *sqlx.DB) error {
	// require emp_code from caller
	suEmpCode := c.Query("emp_code")
	if suEmpCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "emp_code required"})
	}

	// lookup spg_id in sys_user
	var spg sql.NullInt64
	if err := db.Get(&spg, `SELECT spg_id FROM sys_user WHERE su_emp_code = ? LIMIT 1`, suEmpCode); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	// only allow when spg_id == 1
	if !spg.Valid || spg.Int64 != 1 {
		return c.Status(200).JSON(5)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM info_project WHERE ip_status = 'added'`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"count": count})
}

func GetProjectStep2(c *fiber.Ctx, db *sqlx.DB) error {
	ipID := c.Query("mt_id")
	var project SysInfoProject
	if err := db.Get(&project, `SELECT info_project.ip_id AS ip_id,
					 info_project.mdt_id AS mdt_id,
					 info_project.ip_model AS ip_model,
					 info_project.ip_part_no AS ip_part_no,
					 info_project.ip_part_name AS ip_part_name,
					 info_project.ip_kickoff_date AS ip_kickoff_date,
					 info_project.ip_sop_date AS ip_sop_date,
					 info_project.ip_code AS ip_code,
					 info_project.ip_pos AS ip_pos,
					 info_project.ip_document_no AS ip_document_no,
					 info_project.ip_revision AS ip_revision,
					 info_project.ip_status AS ip_status,
					 info_project.ip_updated_at AS ip_updated_at,
					 info_project.ip_updated_by AS ip_updated_by,
					 info_project.ip_customer_name AS ip_customer_name
				FROM info_project 
				LEFT JOIN info_project_master_plan ON info_project.ip_id = info_project_master_plan.ip_id
				WHERE info_project.ip_id = ?`, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(project)
}

func InsertProjectStep3(c *fiber.Ctx, db *sqlx.DB) error {

	var body struct {
		IpID         int64         `json:"ip_id"`
		MaName       []string      `json:"ma_name"`
		MppID        []int64       `json:"mpp_id"`
		SuID         []string      `json:"su_id"`          // per row: "2,6"
		IpidLineCode []string      `json:"ipid_line_code"` // per row: "4,6"
		StartDate    StringOrArray `json:"ipid_start_date"`
		EndDate      StringOrArray `json:"ipid_end_date"`
		CreatedBy    string        `json:"created_by"`
	}

	raw := c.Body()
	if len(strings.TrimSpace(string(raw))) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "empty body"})
	}

	// Try BodyParser first (handles JSON and form/multipart).
	if err := c.BodyParser(&body); err != nil {
		// If BodyParser fails, attempt json.Unmarshal into struct.
		if err2 := json.Unmarshal(raw, &body); err2 != nil {
			// As a last-resort, try to extract at least ip_id from the raw JSON
			var m map[string]any
			if err3 := json.Unmarshal(raw, &m); err3 == nil {
				if v, ok := m["ip_id"]; ok {
					switch t := v.(type) {
					case float64:
						body.IpID = int64(t)
					case string:
						if id, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
							body.IpID = id
						}
					}

					var _m map[string]any
					if err := json.Unmarshal(raw, &_m); err == nil {
						m := _m
						// coerce ma_name if present
						if v, ok := m["ma_name"]; ok {
							switch t := v.(type) {
							case []any:
								out := make([]string, 0, len(t))
								for _, e := range t {
									if s, ok2 := e.(string); ok2 {
										out = append(out, s)
									}
								}
								if len(out) > 0 {
									body.MaName = out
								}
							case string:
								if len(body.MaName) == 0 {
									body.MaName = []string{t}
								}
							}
						}

						// coerce mpp_id if present
						if v, ok := m["mpp_id"]; ok {
							switch t := v.(type) {
							case []any:
								out := make([]int64, 0, len(t))
								for _, e := range t {
									switch x := e.(type) {
									case float64:
										out = append(out, int64(x))
									case string:
										if id, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64); err == nil {
											out = append(out, id)
										}
									}
								}
								if len(out) > 0 {
									body.MppID = out
								}
							case string:
								if id, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
									body.MppID = []int64{id}
								}
							}
						}

						// coerce su_id
						if v, ok := m["su_id"]; ok {
							switch t := v.(type) {
							case []any:
								out := make([]string, 0, len(t))
								for _, e := range t {
									switch x := e.(type) {
									case float64:
										out = append(out, strconv.FormatInt(int64(x), 10))
									case string:
										out = append(out, x)
									}
								}
								if len(out) > 0 {
									body.SuID = out
								}
							case string:
								body.SuID = []string{t}
							}
						}

						// coerce ipid_line_code
						if v, ok := m["ipid_line_code"]; ok {
							switch t := v.(type) {
							case []any:
								out := make([]string, 0, len(t))
								for _, e := range t {
									if s, ok2 := e.(string); ok2 {
										out = append(out, s)
									}
								}
								if len(out) > 0 {
									body.IpidLineCode = out
								}
							case string:
								body.IpidLineCode = []string{t}
							}
						}
					}
				}

				if v, ok := m["created_by"]; ok {
					if s, ok2 := v.(string); ok2 {
						body.CreatedBy = s
					}
				}
				if v, ok := m["ma_name"]; ok {
					// try to coerce to []string if possible
					if arr, ok2 := v.([]any); ok2 {
						out := make([]string, 0, len(arr))
						for _, e := range arr {
							if s, ok3 := e.(string); ok3 {
								out = append(out, s)
							}
						}
						body.MaName = out
					}
				}

				// coerce su_id (allow arrays of numbers/strings or single CSV string)
				if v, ok := m["su_id"]; ok {
					switch t := v.(type) {
					case []any:
						out := make([]string, 0, len(t))
						for _, e := range t {
							switch x := e.(type) {
							case nil:
								out = append(out, "")
							case float64:
								out = append(out, strconv.FormatInt(int64(x), 10))
							case string:
								out = append(out, x)
							default:
								out = append(out, fmt.Sprintf("%v", x))
							}
						}
						body.SuID = out
					case string:
						body.SuID = []string{t}
					}
				}

				// coerce ipid_line_code similar to sd_id
				if v, ok := m["ipid_line_code"]; ok {
					switch t := v.(type) {
					case []any:
						out := make([]string, 0, len(t))
						for _, e := range t {
							if e == nil {
								out = append(out, "")
								continue
							}
							if s, ok2 := e.(string); ok2 {
								out = append(out, s)
							} else {
								out = append(out, fmt.Sprintf("%v", e))
							}
						}
						body.IpidLineCode = out
					case string:
						body.IpidLineCode = []string{t}
					}
				}
			}
		}
	}

	// ------------------------
	// 2) Validate
	// ------------------------
	n := len(body.MaName)
	if body.IpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	if n == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ma_name required"})
	}
	body.CreatedBy = strings.TrimSpace(body.CreatedBy)
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "created_by required"})
	}

	mustEqual := func(arrLen int) bool { return arrLen == 0 || arrLen == n }
	// allow Start/End to be 0,1,n
	startOK := len(body.StartDate) == 0 || len(body.StartDate) == 1 || len(body.StartDate) == n
	endOK := len(body.EndDate) == 0 || len(body.EndDate) == 1 || len(body.EndDate) == n
	// allow Su/LineCode to be 0,1,n (single value reused for all rows)
	suOK := len(body.SuID) == 0 || len(body.SuID) == 1 || len(body.SuID) == n
	lineOK := len(body.IpidLineCode) == 0 || len(body.IpidLineCode) == 1 || len(body.IpidLineCode) == n

	if !mustEqual(len(body.MppID)) || !suOK || !lineOK || !startOK || !endOK {
		return c.Status(400).JSON(fiber.Map{
			"error": "all arrays must be equal length or omitted (dates can be 1 or n)",
		})
	}

	// ------------------------
	// Helpers
	// ------------------------
	parseCSVInts := func(s string) []int64 {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]int64, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if v, err := strconv.ParseInt(p, 10, 64); err == nil {
				out = append(out, v)
			}
		}
		return out
	}

	parseCSVStrings := func(s string) []string {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	// Convert StringOrArray date -> any (time.Time) or nil
	getDateAny := func(arr []string, row int) any {
		if len(arr) == 0 {
			return nil
		}
		idx := row
		if len(arr) == 1 {
			idx = 0
		}
		if idx < 0 || idx >= len(arr) {
			return nil
		}
		val := strings.TrimSpace(arr[idx])
		if val == "" || val == "null" {
			return nil
		}
		t, err := time.ParseInLocation("2006-01-02", val, time.Local)
		if err != nil {
			return nil
		}
		return t
	}

	// ------------------------
	// 3) Transaction
	// ------------------------
	now := time.Now()
	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "transaction begin failed", "detail": err.Error()})
	}
	defer func() { _ = tx.Rollback() }()

	// cleanup: delete any existing info_apqp_item for this ip_id whose name is NOT in incoming list
	incoming := map[string]struct{}{}
	for _, nm := range body.MaName {
		incoming[strings.TrimSpace(nm)] = struct{}{}
	}
	var existingItems []struct {
		IaiID   int64  `db:"iai_id"`
		IaiName string `db:"iai_name"`
	}
	if err := tx.Select(&existingItems, `SELECT iai_id, iai_name FROM info_apqp_item WHERE ip_id = ?`, body.IpID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query existing apqp items failed", "detail": err.Error()})
	}
	for _, it := range existingItems {
		nm := strings.TrimSpace(it.IaiName)
		if _, ok := incoming[nm]; !ok {
			// delete details (only type 'apqp') then delete the apqp item
			if _, err := tx.Exec(`DELETE FROM info_project_item_detail WHERE ref_id = ? AND ipid_type = 'apqp'`, it.IaiID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "delete apqp details failed", "detail": err.Error()})
			}
			if _, err := tx.Exec(`DELETE FROM info_apqp_item WHERE iai_id = ?`, it.IaiID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "delete apqp item failed", "detail": err.Error()})
			}
		}
	}

	// info_approval insertion intentionally removed

	// helper: upsert a single info_project_item_detail row (avoid duplicates)
	upsertDetail := func(refID interface{}, sd interface{}, su interface{}, line interface{}, start interface{}, end interface{}) (int64, error) {
		var existing struct {
			IpidID int64        `db:"ipid_id"`
			Start  sql.NullTime `db:"ipid_start_date"`
			End    sql.NullTime `db:"ipid_end_date"`
		}
		sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'apqp' LIMIT 1`
		if err := tx.Get(&existing, sel, refID, sd, sd, su, su, line, line); err == nil {
			needUpdate := false
			// start
			if start == nil && existing.Start.Valid {
				needUpdate = true
			}
			if startTime, ok := start.(time.Time); ok {
				if !existing.Start.Valid || !existing.Start.Time.Equal(startTime) {
					needUpdate = true
				}
			}
			// end
			if !needUpdate {
				if end == nil && existing.End.Valid {
					needUpdate = true
				}
				if endTime, ok := end.(time.Time); ok {
					if !existing.End.Valid || !existing.End.Time.Equal(endTime) {
						needUpdate = true
					}
				}
			}
			ipidID := existing.IpidID
			if needUpdate {
				if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, start, end, now, body.CreatedBy, ipidID); err != nil {
					return 0, err
				}
			}
			return ipidID, nil
		} else {
			if err != sql.ErrNoRows {
				return 0, err
			}

			var candidate struct {
				IpidID int64        `db:"ipid_id"`
				Start  sql.NullTime `db:"ipid_start_date"`
				End    sql.NullTime `db:"ipid_end_date"`
			}
			sel2 := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail
				WHERE ref_id = ? AND ipid_type = 'apqp'
				AND ((? IS NULL) OR sd_id = ? OR sd_id IS NULL)
				AND ((? IS NULL) OR su_id = ? OR su_id IS NULL)
				AND ((? IS NULL) OR ipid_line_code = ? OR ipid_line_code IS NULL)
				LIMIT 1`
			if err2 := tx.Get(&candidate, sel2, refID, sd, sd, su, su, line, line); err2 == nil {
				// Build UPDATE that sets only columns for which incoming value is non-nil
				parts := []string{"ipid_start_date = ?", "ipid_end_date = ?", "ipid_updated_at = ?", "ipid_updated_by = ?"}
				args := []interface{}{start, end, now, body.CreatedBy}
				if sd != nil {
					parts = append([]string{"sd_id = ?"}, parts...)
					args = append([]interface{}{sd}, args...)
				}
				if su != nil {
					parts = append([]string{"su_id = ?"}, parts...)
					args = append([]interface{}{su}, args...)
				}
				if line != nil {
					parts = append([]string{"ipid_line_code = ?"}, parts...)
					args = append([]interface{}{line}, args...)
				}
				// append WHERE id
				args = append(args, candidate.IpidID)
				upd := "UPDATE info_project_item_detail SET " + strings.Join(parts, ", ") + " WHERE ipid_id = ?"
				if _, err := tx.Exec(upd, args...); err != nil {
					return 0, err
				}
				return candidate.IpidID, nil
			}

			// no candidate -> insert new row
			res, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				refID, sd, su, line, "apqp", start, end, "inprogress", now, body.CreatedBy, now, body.CreatedBy,
			)
			if err != nil {
				return 0, err
			}
			id, _ := res.LastInsertId()
			return id, nil
		}
	}

	for i := 0; i < n; i++ {

		// ---- insert or reuse info_apqp_item ----
		var mppVal any = nil
		if len(body.MppID) == n && body.MppID[i] != 0 {
			mppVal = body.MppID[i]
		}
		var iaiID int64

		if mppVal != nil {
			if err := tx.Get(&iaiID, `SELECT iai_id FROM info_apqp_item WHERE ip_id = ? AND iai_name = ? AND mpp_id = ? LIMIT 1`,
				body.IpID, body.MaName[i], mppVal,
			); err != nil {
				if err == sql.ErrNoRows {
					res, err := tx.NamedExec(`
						INSERT INTO info_apqp_item
						(ipmp_id, mpp_id, iai_name, iai_created_at, iai_created_by, ip_id)
						VALUES
						(:ipmp_id, :mpp_id, :iai_name, :iai_created_at, :iai_created_by, :ip_id)
					`, map[string]any{
						"ipmp_id":        nil,
						"mpp_id":         mppVal,
						"iai_name":       body.MaName[i],
						"iai_created_at": now,
						"iai_created_by": body.CreatedBy,
						"ip_id":          body.IpID,
					})
					if err != nil {
						return c.Status(500).JSON(fiber.Map{"error": "insert info_apqp_item failed", "detail": err.Error()})
					}
					iaiID, _ = res.LastInsertId()
				} else {
					return c.Status(500).JSON(fiber.Map{"error": "query info_apqp_item failed", "detail": err.Error()})
				}
			}
		} else {
			if err := tx.Get(&iaiID, `SELECT iai_id FROM info_apqp_item WHERE ip_id = ? AND iai_name = ? LIMIT 1`,
				body.IpID, body.MaName[i],
			); err != nil {
				if err == sql.ErrNoRows {
					res, err := tx.NamedExec(`
						INSERT INTO info_apqp_item
						(ipmp_id, mpp_id, iai_name, iai_created_at, iai_created_by, ip_id)
						VALUES
						(:ipmp_id, :mpp_id, :iai_name, :iai_created_at, :iai_created_by, :ip_id)
					`, map[string]any{
						"ipmp_id":        nil,
						"mpp_id":         mppVal,
						"iai_name":       body.MaName[i],
						"iai_created_at": now,
						"iai_created_by": body.CreatedBy,
						"ip_id":          body.IpID,
					})
					if err != nil {
						return c.Status(500).JSON(fiber.Map{"error": "insert info_apqp_item failed", "detail": err.Error()})
					}
					iaiID, _ = res.LastInsertId()
				} else {
					return c.Status(500).JSON(fiber.Map{"error": "query info_apqp_item failed", "detail": err.Error()})
				}
			}
		}

		// ---- su list per row (split by comma) ----
		// Determine whether su was provided per-row (len==n) or as a single/global cell (len==1)
		suPerRow := len(body.SuID) == n
		globalSu := []int64{}
		if len(body.SuID) == 1 {
			globalSu = parseCSVInts(body.SuID[0])
		}
		suList := []int64{}
		if suPerRow {
			suList = parseCSVInts(body.SuID[i])
		}

		// ---- line_code list per row (split by comma) ----
		lineCodeList := []string{}
		if len(body.IpidLineCode) == n {
			lineCodeList = parseCSVStrings(body.IpidLineCode[i]) // "4,6" -> ["4","6"]
		} else if len(body.IpidLineCode) == 1 {
			lineCodeList = parseCSVStrings(body.IpidLineCode[0])
		}

		startDate := getDateAny([]string(body.StartDate), i)
		endDate := getDateAny([]string(body.EndDate), i)

		// Get line code value
		var lineVal any = nil
		if len(lineCodeList) == 1 {
			v := strings.TrimSpace(lineCodeList[0])
			if v != "" && !strings.EqualFold(v, "null") {
				lineVal = v
			}
		}

		// Helper: get sd_id from su_id by querying sys_user
		getSdIdFromSuId := func(suID int64) (interface{}, error) {
			var sdID sql.NullInt64
			if err := tx.Get(&sdID, `SELECT sd_id FROM sys_user WHERE su_id = ? LIMIT 1`, suID); err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			if sdID.Valid {
				return sdID.Int64, nil
			}
			return nil, nil
		}

		// Process su_id list and derive sd_id from each
		if suPerRow && len(suList) > 0 {
			// su provided per-row: insert one row per su with sd_id derived from su_id
			for _, su := range suList {
				sdID, err := getSdIdFromSuId(su)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "failed to get sd_id from su_id", "detail": err.Error()})
				}
				_, err = upsertDetail(iaiID, sdID, su, lineVal, startDate, endDate)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "insert info_project_item_detail failed", "detail": err.Error()})
				}
				// info_approval insertion removed; continue
			}
		} else if len(globalSu) > 0 {
			// su was provided as a single/global cell: insert one row per su with sd_id derived from su_id
			for _, su := range globalSu {
				sdID, err := getSdIdFromSuId(su)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "failed to get sd_id from su_id", "detail": err.Error()})
				}
				_, err = upsertDetail(iaiID, sdID, su, lineVal, startDate, endDate)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "insert info_project_item_detail failed", "detail": err.Error()})
				}
				// info_approval insertion removed; continue
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "commit failed", "detail": err.Error()})
	}

	return c.Status(201).JSON(1)
}

func CustomerEventGanttChart(c *fiber.Ctx, db *sqlx.DB) error {
	ipIDStr := c.Query("ip_id")
	if strings.TrimSpace(ipIDStr) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id", "detail": err.Error()})
	}

	var out []struct {
		Group  string `db:"group" json:"group"`
		Name   string `db:"name" json:"name"`
		Start  *Date  `db:"start" json:"start"`
		End    *Date  `db:"end" json:"end"`
		Status string `db:"status" json:"status"`
	}

	query := `SELECT
  'Customer EVT Event' AS ` + "`group`" + `,
  x.` + "`name`" + `,
  x.` + "`start`" + `,
  x.` + "`end`" + `,
  CASE
	WHEN x.` + "`start`" + ` IS NULL THEN ''
	WHEN x.` + "`start`" + ` <= CURDATE() THEN 'Done'
	ELSE 'In Process'
  END AS ` + "`status`" + `
	FROM (
	SELECT 'K-KEN#1' AS ` + "`name`" + `, ice.ice_k_ken_1_date AS ` + "`start`" + `, ice.ice_k_ken_1_end_date AS ` + "`end`" + `
	FROM info_customer_event ice WHERE ice.ip_id = ?

	UNION ALL
	SELECT 'K-KEN#2', ice.ice_k_ken_2_date, ice.ice_k_ken_2_end_date
	FROM info_customer_event ice WHERE ice.ip_id = ?

	UNION ALL
	SELECT '1PP', ice.ice_1pp_date, ice.ice_1pp_end_date
	FROM info_customer_event ice WHERE ice.ip_id = ?

	UNION ALL
	SELECT '2PP', ice.ice_2pp_date, ice.ice_2pp_end_date
	FROM info_customer_event ice WHERE ice.ip_id = ?

	UNION ALL
	SELECT 'PP', ice.ice_pp_date, ice.ice_pp_end_date
	FROM info_customer_event ice WHERE ice.ip_id = ?

	UNION ALL
	SELECT 'SOP', ice.ice_sop_date, ice.ice_sop_end_date
	FROM info_customer_event ice WHERE ice.ip_id = ?
	) x
	WHERE x.` + "`start`" + ` IS NOT NULL
	ORDER BY x.` + "`name`" + `;`

	if err := db.Select(&out, query, ipID, ipID, ipID, ipID, ipID, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	return c.Status(200).JSON(out)
}

func InternalEventGanttChart(c *fiber.Ctx, db *sqlx.DB) error {
	ipIDStr := c.Query("ip_id")
	if strings.TrimSpace(ipIDStr) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id", "detail": err.Error()})
	}

	var out []struct {
		Group     string           `db:"group" json:"group"`
		Name      utils.NullString `db:"name" json:"name"`
		Start     *Date            `db:"start" json:"start"`
		End       *Date            `db:"end" json:"end"`
		Status    utils.NullString `db:"status" json:"status"`
		CreatedBy utils.NullString `db:"ip_created_by" json:"ip_created_by"`
		Firstname utils.NullString `db:"su_firstname" json:"su_firstname"`
		Lastname  utils.NullString `db:"su_lastname" json:"su_lastname"`
	}

	query := "SELECT " +
		"'TBKK Event' AS `group`," +
		"ipmp_name AS name," +
		"ipmp_start_date AS start," +
		"ipmp_end_date AS end," +
		"ipmp_status AS status," +
		"ip.ip_created_by AS ip_created_by," +
		"su.su_firstname AS su_firstname," +
		"su.su_lastname AS su_lastname " +
		"FROM " +
		"info_project_master_plan ipmp " +
		"LEFT JOIN info_project ip ON ipmp.ip_id = ip.ip_id " +
		"LEFT JOIN sys_user su ON ip.ip_created_by = su.su_emp_code " +
		"WHERE " +
		"ip.ip_id = ?"

	if err := db.Select(&out, query, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(out)
}

func GetinfoGanttchart(c *fiber.Ctx, db *sqlx.DB) error {
	ipIDStr := c.Query("ip_id")
	if strings.TrimSpace(ipIDStr) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id", "detail": err.Error()})
	}

	var out []struct {
		MppOrder      utils.NullInt64  `db:"mpp_order" json:"mpp_order"`
		MppName       utils.NullString `db:"mpp_name" json:"mpp_name"`
		IaiName       utils.NullString `db:"iai_name" json:"iai_name"`
		IpidStartDate *Date            `db:"ipid_start_date" json:"ipid_start_date"`
		IpidEndDate   *Date            `db:"ipid_end_date" json:"ipid_end_date"`
		IpidStatus    utils.NullString `db:"ipid_status" json:"ipid_status"`
		SdDeptAname   utils.NullString `db:"sd_dept_aname" json:"sd_dept_aname"`
		SuFirstname   utils.NullString `db:"su_firstname" json:"su_firstname"`
	}

	query := `SELECT
    COALESCE(mpp.mpp_order, 6) AS mpp_order,
    COALESCE(mpp.mpp_name, 'Customer PPAP Status') AS mpp_name,
    iai.iai_name,
    ipid.ipid_start_date,
    ipid.ipid_end_date,
    ipid.ipid_status,
    GROUP_CONCAT(
        COALESCE(sd.sd_dept_aname, '-')
        ORDER BY
        su.su_firstname SEPARATOR ' / '
    ) AS sd_dept_aname,
    GROUP_CONCAT(
        COALESCE(su.su_firstname, '-')
        ORDER BY
        su.su_firstname SEPARATOR ' / '
    ) AS su_firstname
FROM
    info_project AS ip
    LEFT JOIN info_apqp_item AS iai ON ip.ip_id = iai.ip_id
    LEFT JOIN mst_project_phase AS mpp ON iai.mpp_id = mpp.mpp_id
    LEFT JOIN info_project_item_detail AS ipid ON iai.iai_id = ipid.ref_id AND ipid.ipid_type = 'apqp'
    LEFT JOIN sys_department AS sd ON sd.sd_id = ipid.sd_id
    LEFT JOIN sys_user AS su ON su.su_id = ipid.su_id
WHERE
    ip.ip_id = ?
GROUP BY
    mpp.mpp_order,
    mpp.mpp_name,
    iai.iai_name,
    ipid.ipid_start_date,
    ipid.ipid_end_date,
    ipid.ipid_status

UNION ALL

SELECT
    6 AS mpp_order,
    'Customer PPAP Status' AS mpp_name,
    ipi.ipi_name AS iai_name,
    ipid.ipid_start_date,
    ipid.ipid_end_date,
    ipid.ipid_status,
    GROUP_CONCAT(
        COALESCE(sd.sd_dept_aname, '-')
        ORDER BY
        su.su_firstname SEPARATOR ' / '
    ) AS sd_dept_aname,
    GROUP_CONCAT(
        COALESCE(su.su_firstname, '-')
        ORDER BY
        su.su_firstname SEPARATOR ' / '
    ) AS su_firstname
FROM
    info_project AS ip
    LEFT JOIN info_ppap_item AS ipi ON ip.ip_id = ipi.ip_id
    LEFT JOIN info_project_item_detail AS ipid ON ipi.ipi_id = ipid.ref_id AND ipid.ipid_type = 'ppap'
    LEFT JOIN sys_department AS sd ON sd.sd_id = ipid.sd_id
    LEFT JOIN sys_user AS su ON su.su_id = ipid.su_id
WHERE
    ip.ip_id = ?
GROUP BY
    ipi.ipi_name,
    ipid.ipid_start_date,
    ipid.ipid_end_date,
    ipid.ipid_status;`

	if err := db.Select(&out, query, ipID, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(out)
}

func GetCountItemUpload(c *fiber.Ctx, db *sqlx.DB) error {
	ipIDStr := c.Query("ip_id")
	if strings.TrimSpace(ipIDStr) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id", "detail": err.Error()})
	}
	var count int64
	query := `SELECT COUNT(iai.iai_id) AS count
	FROM info_apqp_item AS iai
	WHERE iai.ip_id = ?`
	if err := db.Get(&count, query, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(count)
}

func GetListAPQPPPAPItem(c *fiber.Ctx, db *sqlx.DB) error {
	ipIDStr := c.Query("ip_id")
	if strings.TrimSpace(ipIDStr) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id", "detail": err.Error()})
	}

	var out []struct {
		RefID         int64            `db:"ref_id" json:"ref_id"`
		IpID          int64            `db:"ip_id" json:"ip_id"`
		IpidID        int64            `db:"ipid_id" json:"ipid_id"`
		MppID         utils.NullInt64  `db:"mpp_id" json:"mpp_id"`
		ItemName      utils.NullString `db:"item_name" json:"item_name"`
		ItemType      utils.NullString `db:"item_type" json:"item_type"`
		Department    utils.NullString `db:"department" json:"department"`
		OwnerSuID     utils.NullInt64  `db:"owner_su_id" json:"owner_su_id"`
		SuEmpCode     utils.NullString `db:"su_emp_code" json:"su_emp_code"`
		SuFirstname   utils.NullString `db:"su_firstname" json:"su_firstname"`
		SuLastname    utils.NullString `db:"su_lastname" json:"su_lastname"`
		StartDate     *Date            `db:"start_date" json:"start_date"`
		EndDate       *Date            `db:"end_date" json:"end_date"`
		StatusApprove utils.NullString `db:"status_approve" json:"status_approve"`
		ItfFileName   utils.NullString `db:"itf_file_name" json:"itf_file_name"`
		ItfFilePath   utils.NullString `db:"itf_file_path" json:"itf_file_path"`
		ApproverSuID  utils.NullInt64  `db:"ia_su_id" json:"ia_su_id"`
		IaStatus      utils.NullString `db:"ia_status" json:"ia_status"`
		IaType        utils.NullString `db:"ia_type" json:"ia_type"`
		StatusProject utils.NullInt64  `db:"status_project" json:"status_project"`
	}

	query := `SELECT DISTINCT
                    x.ref_id,
                    x.ip_id,
					x.ipid_id,
					x.mpp_id,
					x.item_name,
					x.item_type,
					x.department,
					x.owner_su_id,
					su.su_emp_code,
					su.su_firstname,
					su.su_lastname,
					x.start_date,
					x.end_date,
					x.ipid_status AS status_approve,
					x.itf_file_name,
					x.itf_file_path,
					a.su_id AS ia_su_id,
					a.ia_status,
					a.ia_type,
					CASE
                        WHEN a.ia_status IS NULL AND a.ia_type IS NULL THEN 0
                        WHEN a.ia_status = 'waiting' AND a.ia_type = 'Leader' THEN 1
                        WHEN a.ia_status = 'waiting' AND a.ia_type = 'PJ' THEN 2
                        WHEN a.ia_status = 'Approve' AND a.ia_type = 'PJ' THEN 3
                        WHEN a.ia_status = 'reject' AND a.ia_type = 'PJ' THEN 4
						WHEN a.ia_status = 'reject' AND a.ia_type = 'Leader' THEN 5
                        ELSE NULL
                    END AS status_project
				FROM
				(
						SELECT
							pid.ref_id 									AS ref_id,
						    ai.ip_id                                   AS ip_id,
						    ipmp.ipmp_id AS ipmp_id,
							ai.mpp_id                                  AS mpp_id,
							ai.iai_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							sd.sd_dept_aname                           AS department,
							pid.su_id                                  AS owner_su_id,
							ai.iai_created_at                          AS start_date,
							ai.iai_created_at                          AS end_date,
							pid.ipid_id                                AS ipid_id,
							pid.ipid_status                           AS ipid_status,
							tf.itf_file_name                           AS itf_file_name,
							tf.itf_file_path                           AS itf_file_path
						FROM info_project_item_detail pid
						JOIN info_apqp_item ai
						ON ai.iai_id = pid.ref_id
						AND pid.ipid_type = 'apqp'
						JOIN sys_department sd
							ON sd.sd_id = pid.sd_id
						LEFT JOIN info_tracking_file tf
							ON tf.ipid_id = pid.ipid_id
						LEFT JOIN info_project_master_plan ipmp ON ipmp.ip_id = ai.ip_id
						WHERE ai.ip_id = ?

						UNION ALL

						SELECT
						    pid.ref_id 							AS ref_id,
							pi.ip_id                                   AS ip_id,
							NULL                                       AS ipmp_id,
							NULL                                       AS mpp_id,
							pi.ipi_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							sd.sd_dept_aname                           AS department,
							pid.su_id                                  AS owner_su_id,
							pi.ipi_created_at                          AS start_date,
							pi.ipi_created_at                          AS end_date,
							pid.ipid_id                                AS ipid_id,
							pid.ipid_status                            AS ipid_status,
							tf.itf_file_name                           AS itf_file_name,
							tf.itf_file_path                           AS itf_file_path
						FROM info_project_item_detail pid
						JOIN info_ppap_item pi
							ON pi.ipi_id = pid.ref_id  AND pid.ipid_type = 'ppap'
						JOIN sys_department sd
							ON sd.sd_id = pid.sd_id
						LEFT JOIN info_tracking_file tf
							ON tf.ipid_id = pid.ipid_id
						AND pid.ipid_type = 'ppap'
						
						WHERE pi.ip_id = ?
                        
				) x
				LEFT JOIN info_approval a
					ON a.ipid_id = x.ipid_id AND a.ia_is_action = 1
				LEFT JOIN sys_user su
					ON su.su_id = x.owner_su_id
				GROUP BY x.ref_id
				ORDER BY
					x.item_name ASC,
					x.item_type ASC,
					x.start_date ASC;`

	args := []interface{}{ipID, ipID}

	if err := db.Select(&out, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(out)
}

func UpdateStatusProjectItemDetail(c *fiber.Ctx, db *sqlx.DB) error {
	ipidID := c.Query("ipid_id")
	status := c.Query("status")
	note := c.Query("note")
	UpdateBy := c.Query("updateBy")

	if ipidID == "" {
		ipidID = c.FormValue("ipid_id")
	}
	if status == "" {
		status = c.FormValue("status")
	}
	if note == "" {
		note = c.FormValue("note")
	}
	if UpdateBy == "" {
		UpdateBy = c.FormValue("updateBy")
	}

	// accept JSON body as fallback
	if ipidID == "" || status == "" {
		var body struct {
			IpidID int64  `json:"ipid_id"`
			Status string `json:"status"`
			Note   string `json:"note"`
		}
		if err := c.BodyParser(&body); err == nil {
			if ipidID == "" && body.IpidID != 0 {
				ipidID = strconv.FormatInt(body.IpidID, 10)
			}
			if status == "" && body.Status != "" {
				status = body.Status
			}
			if note == "" && body.Note != "" {
				note = body.Note
			}
		}
	}

	if ipidID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_id required"})
	}
	id, err := strconv.ParseInt(ipidID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_id"})
	}

	// Get ref_id from the given ipid_id
	var refID int64
	var ipidType sql.NullString
	if err := db.Get(&refID, `SELECT ref_id FROM info_project_item_detail WHERE ipid_id = ?`, id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_id not found"})
	}

	// Get ipid_type from the given ipid_id
	if err := db.Get(&ipidType, `SELECT ipid_type FROM info_project_item_detail WHERE ipid_id = ?`, id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_type not found"})
	}

	// Get ia_type from info_approval for this ipid_id
	var iaType sql.NullString
	if err := db.Get(&iaType, `SELECT ia_type FROM info_approval WHERE ipid_id = ? AND ia_is_action = 1 LIMIT 1`, id); err != nil && err != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get ia_type", "detail": err.Error()})
	}

	// Get all ipid_ids with the same ref_id
	var ipidIDs []int64
	if err := db.Select(&ipidIDs, `SELECT ipid_id FROM info_project_item_detail WHERE ref_id = ?`, refID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get related ipid_ids", "detail": err.Error()})
	}

	// Update all ipid_ids with the same ref_id
	for _, ipidIDVal := range ipidIDs {
		query := `UPDATE info_project_item_detail
		SET ipid_status = ? , ipid_updated_at = ? , ipid_updated_by = ?
		WHERE ipid_id = ?`

		if _, err := db.Exec(query, status, time.Now(), UpdateBy, ipidIDVal); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
		}
	}

	newStatus := status
	if status == "done" {
		newStatus = "approve"
	}

	// If status is reject or approve, update info_approval table where ia_type matches
	if status == "reject" || newStatus == "approve" {
		// Build WHERE clause based on iaType
		var updateApprovalQuery string
		var args []interface{}

		if iaType.Valid && iaType.String != "" {
			// Update all approvals with same ref_id and ia_type
			updateApprovalQuery = `UPDATE info_approval 
			SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_note = ? 
			WHERE ipid_id IN (SELECT ipid_id FROM info_project_item_detail WHERE ref_id = ?) 
			AND ia_status = 'waiting' AND ia_is_action = 1 AND ia_type = ? AND ia_status_flg = 'active'`
			now := time.Now()
			args = []interface{}{newStatus, now, UpdateBy, note, refID, iaType.String}
		} else {
			// Fallback: update all approvals for these ipid_ids
			updateApprovalQuery = `UPDATE info_approval 
			SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_note = ? 
			WHERE ipid_id IN (SELECT ipid_id FROM info_project_item_detail WHERE ref_id = ?) 
			AND ia_status = 'waiting' AND ia_is_action = 1 AND ia_status_flg = 'active'`
			now := time.Now()
			args = []interface{}{newStatus, now, UpdateBy, note, refID}
		}

		if _, err := db.Exec(updateApprovalQuery, args...); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update approval status", "detail": err.Error()})
		}
	}

	// Send email to the owner with new HTML format (for reject, approve, and done)
	if status == "reject" || status == "done" || status == "approve" {
		var detail struct {
			ProjectCode sql.NullString `db:"ip_code"`
			PartNo      sql.NullString `db:"ip_part_no"`
			PartName    sql.NullString `db:"ip_part_name"`
			IpModel     sql.NullString `db:"ip_model"`
			OwnerSuID   sql.NullInt64  `db:"su_id"`
		}

		// Get project details (first item with same ref_id)
		q := `SELECT
		ip.ip_code,
		ip.ip_part_no,
		ip.ip_part_name,
		ip.ip_model,
		pid.su_id
	FROM info_project_item_detail pid
	LEFT JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id AND pid.ipid_type = 'apqp'
	LEFT JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id AND pid.ipid_type = 'ppap'
	LEFT JOIN info_project ip ON ip.ip_id = COALESCE(ai.ip_id, pi.ip_id)
	WHERE pid.ipid_id = ? LIMIT 1`

		if err := db.Get(&detail, q, id); err == nil {
			// Get owner email
			var ownerEmail sql.NullString
			if detail.OwnerSuID.Valid {
				_ = db.Get(&ownerEmail, `SELECT su_email FROM sys_user WHERE su_id = ? AND su_status = 'active' AND su_email IS NOT NULL AND su_email <> '' LIMIT 1`, detail.OwnerSuID.Int64)
			}

			// Get all items with the same ref_id
			type ItemDetail struct {
				ItemName  sql.NullString `db:"item_name"`
				ItemType  sql.NullString `db:"item_type"`
				StartDate sql.NullTime   `db:"ipid_start_date"`
				EndDate   sql.NullTime   `db:"ipid_end_date"`
			}

			var allItems []ItemDetail
			qItems := `SELECT
			COALESCE(ai.iai_name, pi.ipi_name) AS item_name,
			pid.ipid_type AS item_type,
			pid.ipid_start_date,
			pid.ipid_end_date
		FROM info_project_item_detail pid
		LEFT JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id AND pid.ipid_type = 'apqp'
		LEFT JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id AND pid.ipid_type = 'ppap'
		WHERE pid.ref_id = ? AND pid.ipid_type = ?
		ORDER BY pid.ipid_id`

			_ = db.Select(&allItems, qItems, refID, ipidType.String)

			var sb strings.Builder

			// Get owner firstname from sys_user using su_id from info_project_item_detail
			var ownerFirstname sql.NullString
			if detail.OwnerSuID.Valid {
				_ = db.Get(&ownerFirstname, `SELECT su_firstname FROM sys_user WHERE su_id = ? AND su_status = 'active' LIMIT 1`, detail.OwnerSuID.Int64)
			}

			ownerStr := "User"
			if ownerFirstname.Valid && strings.TrimSpace(ownerFirstname.String) != "" {
				ownerStr = ownerFirstname.String
			}

			// Get approver's firstname and lastname (person who approved/rejected)
			var approverFirstName, approverLastName sql.NullString
			_ = db.QueryRow(`SELECT su_firstname, su_lastname FROM sys_user WHERE su_emp_code = ? AND su_status = 'active' LIMIT 1`, UpdateBy).Scan(&approverFirstName, &approverLastName)
			var approverName string
			if approverFirstName.Valid && approverLastName.Valid {
				approverName = approverFirstName.String + " " + approverLastName.String
			} else if approverFirstName.Valid {
				approverName = approverFirstName.String
			} else if approverLastName.Valid {
				approverName = approverLastName.String
			} else {
				approverName = UpdateBy
			}

			sb.WriteString("<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>Dear, " + html.EscapeString(ownerStr) + "</h3>")

			// Determine status text based on status value
			var statusText string

			var reasonLabel string
			if status == "reject" {
				statusText = "<b style='color : #dc2626;'>rejected </b> by K." + html.EscapeString(approverName) + ""

				reasonLabel = "Reason for Rejection:"
			} else {
				statusText = "<b style='color : #16a34a;'>approved </b> by K." + html.EscapeString(approverName) + ""

				reasonLabel = "Approval Note:"
			}

			sb.WriteString("<h4>Your project item has been " + statusText + "</h4>")
			// Only show note box for reject status
			if status == "reject" && strings.TrimSpace(note) != "" {
				sb.WriteString("<div style='margin-top:15px; padding:12px; background:#fee2e2; border-left:4px solid #dc2626; color:#7f1d1d; font-size:13px;'>")
				sb.WriteString("<b>" + reasonLabel + "</b><br>")
				sb.WriteString(html.EscapeString(note))
				sb.WriteString("</div><br>")
			}
			sb.WriteString("<html><body style='font-family: Arial, sans-serif; background:#f6f8fb; padding:20px;'>")

			// Project Detail Card
			sb.WriteString("<div style='margin:auto; background:#ffffff; border-radius:10px; border:1px solid #e0e6ed; padding:20px;'>")
			sb.WriteString("<div style='font-size:18px; font-weight:bold; color:#1f2d3d;'>Project Detail</div>")
			sb.WriteString("<div style='font-size:12px; color:#6b7280;'>Item Information (ข้อมูลของรายการโปรเจค)</div>")

			sb.WriteString("<hr style='border:none; border-top:1px dashed #d1d5db; margin:15px 0;'>")

			// Project Information 2-Column Layout
			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse;'>")
			sb.WriteString("<tr>")

			sb.WriteString("<td style='width:50%; border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PROJECT CODE</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>#" + html.EscapeString(getStringValue(detail.ProjectCode)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:50%; border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>MODEL</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.IpModel)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("</tr>")
			sb.WriteString("<tr>")

			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NO</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartNo)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td colspan='1' style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NAME</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartName)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("</tr>")
			sb.WriteString("</table>")
			sb.WriteString("<div style='margin-top:20px; font-size:14px; color:#374151;'>")
			sb.WriteString("</div>")
			sb.WriteString("</div><br>")

			// Items Table - show ALL items consolidated
			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse; border:1px solid #e5e7eb;'>")
			sb.WriteString("<thead><tr style='background:#f3f4f6; border-bottom:2px solid #d1d5db;'>")
			cols := []string{"No.", "Item Name", "Item Type", "Start Date", "End Date"}
			for _, c := range cols {
				sb.WriteString("<th style='text-align:left; font-weight:bold; color:#374151; font-size:13px;'>" + html.EscapeString(c) + "</th>")
			}
			sb.WriteString("</tr></thead><tbody>")

			for i, r := range allItems {
				var startStr, endStr string
				if r.StartDate.Valid {
					startStr = r.StartDate.Time.Format("2006-01-02")
				}
				if r.EndDate.Valid {
					endStr = r.EndDate.Time.Format("2006-01-02")
				}

				rowBg := ""
				if i%2 == 0 {
					rowBg = " style='background:#fbfdff;'"
				}
				sb.WriteString("<tr" + rowBg + ">")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + strconv.Itoa(i+1) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(getStringValue(r.ItemName)) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(getStringValue(r.ItemType)) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(startStr) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(endStr) + "</td>")
				sb.WriteString("</tr>")
			}

			sb.WriteString("</tbody></table>")

			sb.WriteString("<div style='margin-top:20px;'>")
			sb.WriteString("<a href='http://192.168.161.205:4005/login' style='display:inline-block; padding:10px 20px; background:#2563eb; color:#fff; text-decoration:none; border-radius:6px; font-weight:bold; font-size:14px;'>Open Project Management</a>")
			sb.WriteString("</div>")

			sb.WriteString("<div style='margin-top:30px; padding-top:20px; border-top:1px solid #e5e7eb; font-size:13px; color:#6b7280;'>")
			sb.WriteString("<p>Best Regards,<br><strong>System Service Department</strong></p>")
			sb.WriteString("</div>")

			sb.WriteString("</body></html>")

			// Set subject based on status
			var subject string
			if status == "reject" {
				subject = "TBKK Project Control Notification : Item Rejected"
			} else {
				subject = "TBKK Project Control Notification : Item Approved"
			}

			if ownerEmail.Valid && strings.TrimSpace(ownerEmail.String) != "" {
				_ = SendMail([]string{ownerEmail.String}, subject, sb.String(), "text/html; charset=utf-8")
			}
		}
	}

	return c.Status(200).JSON(1)
}

func UpdateStatusCompleteProject(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		IpID      int64  `json:"ip_id"`
		UpdatedBy string `json:"updated_by"`
	}

	// Try to get from query first
	ipIDStr := c.Query("ip_id")
	if ipIDStr != "" {
		id, err := strconv.ParseInt(ipIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id"})
		}
		body.IpID = id
	}

	// Try body parser if not found
	if body.IpID == 0 {
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "body parse error"})
		}
	}

	if body.IpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}

	if body.UpdatedBy == "" {
		body.UpdatedBy = c.Query("updated_by")
	}

	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "updated_by required"})
	}

	// Start transaction
	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "transaction begin failed", "detail": err.Error()})
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()

	// Verify ip_id exists
	var exists int
	if err := tx.Get(&exists, `SELECT 1 FROM info_project WHERE ip_id = ? LIMIT 1`, body.IpID); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "ip_id not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	// 1. Update info_project to 'finished'
	if _, err := tx.Exec(`
		UPDATE info_project 
		SET ip_status = 'finished', ip_updated_at = ?, ip_updated_by = ?
		WHERE ip_id = ?
	`, now, body.UpdatedBy, body.IpID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update info_project failed", "detail": err.Error()})
	}

	// 2. Update info_project_master_plan to 'done'
	if _, err := tx.Exec(`
		UPDATE info_project_master_plan 
		SET ipmp_status = 'done'
		WHERE ip_id = ?
	`, body.IpID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update info_project_master_plan failed", "detail": err.Error()})
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "commit failed", "detail": err.Error()})
	}

	return c.Status(200).JSON(1)
}
