package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"mime/multipart"
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

type ProjectItemDetail struct {
	IpID                 int64            `json:"ip_id" db:"ip_id"`
	IpCode               utils.NullString `json:"ip_code" db:"ip_code"`
	IpModel              utils.NullString `json:"ip_model" db:"ip_model"`
	IpPartNo             utils.NullString `json:"ip_part_no" db:"ip_part_no"`
	IpPartName           utils.NullString `json:"ip_part_name" db:"ip_part_name"`
	IpStatus             utils.NullString `json:"ip_status" db:"ip_status"`
	IaStatus             utils.NullString `json:"ia_status" db:"ia_status"`
	IpCreatedBy          utils.NullString `json:"ip_created_by" db:"ip_created_by"`
	IpCreatedAt          *DateTime        `json:"ip_created_at" db:"ip_created_at"`
	IpUpdatedByFirstname utils.NullString `json:"ip_updated_by_firstname" db:"ip_updated_by_firstname"`
	IpUpdatedByLastname  utils.NullString `json:"ip_updated_by_lastname" db:"ip_updated_by_lastname"`
	IpKickoffDate        *Date            `json:"ip_kickoff_date" db:"ip_kickoff_date"`
	IpSopDate            *Date            `json:"ip_sop_date" db:"ip_sop_date"`
	IpCustomerName       utils.NullString `json:"ip_customer_name" db:"ip_customer_name"`

	SuID        int64            `json:"su_id" db:"su_id"`
	SuEmpCode   utils.NullString `json:"su_emp_code" db:"su_emp_code"`
	SuFirstname utils.NullString `json:"su_firstname" db:"su_firstname"`
	SuLastname  utils.NullString `json:"su_lastname" db:"su_lastname"`
}

type ProjectTrackingItem struct {
	SuEmpCode   utils.NullString `json:"su_emp_code" db:"su_emp_code"`
	SuFirstname utils.NullString `json:"su_firstname" db:"su_firstname"`
	SuLastname  utils.NullString `json:"su_lastname" db:"su_lastname"`

	IpidID        int64            `db:"ipid_id" json:"ipid_id"`
	MppID         utils.NullInt64  `db:"mpp_id" json:"mpp_id"`
	ItemName      utils.NullString `db:"item_name" json:"item_name"`
	ItemType      utils.NullString `db:"item_type" json:"item_type"`
	Department    utils.NullString `db:"department" json:"department"`
	OwnerSuID     utils.NullInt64  `db:"owner_su_id" json:"owner_su_id"`
	StartDate     *Date            `db:"start_date" json:"start_date"`
	EndDate       *Date            `db:"end_date" json:"end_date"`
	StatusApprove utils.NullString `db:"status_approve" json:"status_approve"`
	IpidStatus    utils.NullString `db:"ipid_status" json:"ipid_status"`

	ItfFileName utils.NullString `db:"itf_file_name" json:"itf_file_name"`
	ItfFilePath utils.NullString `db:"itf_file_path" json:"itf_file_path"`

	IaNote utils.NullString `db:"ia_note" json:"ia_note"`
}

func ListProjectItemDetails(c *fiber.Ctx, db *sqlx.DB) error {
	suParam := strings.TrimSpace(c.Query("su_id"))
	if suParam == "" {
		return c.Status(400).JSON(fiber.Map{"error": "su_id required"})
	}
	suID, err := strconv.ParseInt(suParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid su_id"})
	}

	modelParam := strings.TrimSpace(c.Query("mmm_id"))
	ipStatusParam := strings.TrimSpace(c.Query("ip_status"))

	query := `SELECT
				p.ip_id,
				p.ip_code,
				p.ip_model,
				p.ip_part_no,
				p.ip_status,
				su.su_emp_code AS ip_created_by,
				p.ip_created_at,
				su.su_firstname AS ip_updated_by_firstname,
				su.su_lastname  AS ip_updated_by_lastname
			FROM
			(
				SELECT
					d.ipid_id,
					CASE
						WHEN d.ipid_type = 'apqp' THEN ap.ip_id
						WHEN d.ipid_type = 'ppap' THEN pp.ip_id
						ELSE NULL
					END AS ip_id
				FROM info_project_item_detail d
				LEFT JOIN info_apqp_item ap
					ON d.ipid_type = 'apqp'
				AND d.ref_id    = ap.iai_id
				LEFT JOIN info_ppap_item pp
					ON d.ipid_type = 'ppap'
				AND d.ref_id    = pp.ipi_id
				WHERE d.su_id = ?
			) x
			JOIN info_project p
				ON p.ip_id = x.ip_id
			LEFT JOIN info_project_item_detail dd ON dd.ipid_id = x.ipid_id
			LEFT JOIN sys_user su ON su.su_id = dd.su_id
			LEFT JOIN mst_model_master mm ON mm.mmm_model = p.ip_model AND mm.mmm_customer_name = p.ip_customer_name
			`

	args := []interface{}{suID}

	if modelParam != "" || ipStatusParam != "" {
		query += "WHERE 1=1\n"
		if modelParam != "" {
			query += " AND mm.mmm_id = ?\n"
			args = append(args, modelParam)
		}
		if ipStatusParam != "" {
			query += " AND p.ip_status = ?\n"
			args = append(args, ipStatusParam)
		}
	}
	query += `WHERE p.ip_status = 'inprogress'`
	query += `GROUP BY
				p.ip_id, p.ip_code, p.ip_model, p.ip_part_no, p.ip_status,
				p.ip_created_by, p.ip_created_at, su.su_firstname, su.su_lastname
			`
	query += "ORDER BY\n                p.ip_id ASC;"

	var list []ProjectItemDetail
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func CountProjectTracking(c *fiber.Ctx, db *sqlx.DB) error {
	suID := c.Query("su_id")
	if suID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "su_id required"})
	}
	var res struct {
		TotalAll int `db:"total_all"`
	}

	query := `SELECT COUNT(DISTINCT t.ip_id) AS total_all
				FROM (
				-- PPAP -> ip_id
				SELECT ipi.ip_id
				FROM info_project_item_detail ipid
				JOIN info_ppap_item ipi
					ON ipi.ipi_id = ipid.ref_id
				JOIN info_project ip
					ON ip.ip_id = ipi.ip_id
				WHERE ipid.su_id = ?
					AND ip.ip_status = 'inprogress'
					AND ipid.ipid_type = 'ppap'

				UNION

				-- APQP -> ip_id
				SELECT iai.ip_id
				FROM info_project_item_detail ipid
				JOIN info_apqp_item iai
					ON iai.iai_id = ipid.ref_id
				JOIN info_project ip
					ON ip.ip_id = iai.ip_id
				WHERE ipid.su_id = ?
					AND ip.ip_status = 'inprogress'
					AND ipid.ipid_type = 'apqp'
				) t;

		;`

	if err := db.Get(&res, query, suID, suID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(res)
}

func GetListProjectTracking(c *fiber.Ctx, db *sqlx.DB) error {
	ipParam := strings.TrimSpace(c.Query("ip_id"))
	if ipParam == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}
	ipID, err := strconv.ParseInt(ipParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ip_id"})
	}

	suParam := strings.TrimSpace(c.Query("su_id"))
	var suID interface{}
	if suParam == "" {
		suID = nil
	} else {
		id, err := strconv.ParseInt(suParam, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid su_id"})
		}
		suID = id
	}

	lineParam := strings.TrimSpace(c.Query("ipid_line_code"))
	var lineArg interface{}
	if lineParam == "" {
		lineArg = nil
	} else {
		lineArg = lineParam
	}

	query := `SELECT
					DISTINCT x.ipid_id,
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
					ia.ia_note,
					CASE
						WHEN ia.ia_status IS NULL AND ia.ia_type IS NULL AND ipid_status = 'done' AND tf.itf_file_path IS NOT NULL THEN 8
						WHEN ia.ia_status IS NULL AND ia.ia_type IS NULL AND tf.itf_file_path IS NOT NULL THEN 7
                        WHEN ia.ia_status IS NULL AND ia.ia_type IS NULL THEN 5
                        WHEN ia.ia_status = 'waiting' AND x.ipid_status = 'waiting' AND ia.ia_type = 'Leader' THEN 1
                        WHEN ia.ia_status = 'Approve' AND x.ipid_status = 'waiting' AND ia.ia_type = 'Leader' THEN 2
                        WHEN ia.ia_status = 'reject' AND ia.ia_type = 'Leader' THEN 3
                        WHEN ia.ia_status = 'reject' AND ia.ia_type = 'PJ' THEN 4
                        WHEN ia.ia_status = 'Approve' AND x.ipid_status = 'done' THEN 5
						WHEN ia.ia_status = 'waiting' AND ia.ia_type = 'PJ' THEN 6
                        ELSE NULL
                    END AS status_approve,
					tf.itf_file_name,
					tf.itf_file_path
				
				FROM
				(
					SELECT
					ai.mpp_id AS mpp_id,
					pid.ref_id AS ref_id,
					pid.sd_id AS sd_id,
					ai.iai_name AS item_name,
					pid.ipid_type AS item_type,
					sd.sd_dept_aname AS department,
					pid.su_id AS owner_su_id,
					pid.ipid_start_date AS start_date,
					pid.ipid_end_date AS end_date,
					pid.ipid_id AS ipid_id,
					pid.ipid_status AS ipid_status,
					NULL AS itf_file_name,
					NULL AS itf_file_path  
					FROM
					info_project_item_detail pid
					JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id
					AND pid.ipid_type = 'apqp'
					JOIN sys_department sd ON sd.sd_id = pid.sd_id
					WHERE
					ai.ip_id = ?
					AND (
						(
						pid.ipid_line_code IS NULL
						AND (
							? IS NULL
							OR pid.su_id = ?
						)
						)
						OR (
						pid.ipid_line_code IS NOT NULL
						AND pid.ipid_line_code = ?
						)
					)
					
					UNION ALL

					
					SELECT
					NULL AS mpp_id,
					pid.ref_id AS ref_id,
					pid.sd_id AS sd_id,
					pi.ipi_name AS item_name,
					pid.ipid_type AS item_type,
					sd.sd_dept_aname AS department,
					pid.su_id AS owner_su_id,
					pid.ipid_start_date AS start_date,
					pid.ipid_end_date AS end_date,
					pid.ipid_id AS ipid_id,
					pid.ipid_status AS ipid_status,
					NULL AS itf_file_name,
					NULL AS itf_file_path  
					FROM
					info_project_item_detail pid
					JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id
					AND pid.ipid_type = 'ppap'
					JOIN sys_department sd ON sd.sd_id = pid.sd_id
					WHERE
					pi.ip_id = ?
					AND (
						(
						pid.ipid_line_code IS NULL
						AND (
							? IS NULL
							OR pid.su_id = ?
						)
						)
						OR (
						pid.ipid_line_code IS NOT NULL
						AND pid.ipid_line_code = ?
						)
					)
				) x

				
				LEFT JOIN info_tracking_file tf
					ON tf.ipid_id = (
						SELECT tf_sub.ipid_id
						FROM info_tracking_file tf_sub
						JOIN info_project_item_detail pid_sub 
							ON pid_sub.ipid_id = tf_sub.ipid_id
						WHERE pid_sub.ref_id = x.ref_id
						AND pid_sub.sd_id = x.sd_id
						AND pid_sub.ipid_type = x.item_type
						LIMIT 1
					)

					LEFT JOIN sys_user su ON su.su_id = x.owner_su_id
					LEFT JOIN info_approval ia ON ia.ipid_id = x.ipid_id AND ia.ia_is_action = 1
				ORDER BY
				x.item_type ASC,
				x.start_date ASC,
				x.item_name ASC;
				`
	args := []interface{}{ipID, suID, suID, lineArg, ipID, suID, suID, lineArg}

	var list []ProjectTrackingItem
	// log SQL query and arguments for debugging
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func InsertProjectTracking(c *fiber.Ctx, db *sqlx.DB) error {
	type item struct {
		FileName  string
		FilePath  string
		IpID      int64
		IpidID    int64
		ItfType   string
		CreatedBy string
	}

	var maps []map[string]interface{}
	var uploadedFiles []*multipart.FileHeader

	// helpers to extract values robustly
	getString := func(m map[string]interface{}, k string) string {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				return t
			case float64:
				return strconv.FormatInt(int64(t), 10)
			default:
				return ""
			}
		}
		return ""
	}
	getInt64 := func(m map[string]interface{}, k string) int64 {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t == "" {
					return 0
				}
				n, _ := strconv.ParseInt(t, 10, 64)
				return n
			case float64:
				return int64(t)
			default:
				return 0
			}
		}
		return 0
	}

	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid multipart form", "detail": err.Error()})
		}

		// 1) flatten uploaded files (รับทุก key เช่น file, files, upload ฯลฯ)
		for _, fhs := range form.File {
			for _, fh := range fhs {
				uploadedFiles = append(uploadedFiles, fh)
			}
		}

		// Debug: log what form fields we received
		var receivedFields []string
		for k := range form.Value {
			receivedFields = append(receivedFields, k)
		}

		// 2) prefer `items` JSON if provided, fallback to `payload`
		var itemsJSON string
		if vals, ok := form.Value["items"]; ok && len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
			itemsJSON = vals[0]
		} else if vals, ok := form.Value["payload"]; ok && len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
			itemsJSON = vals[0]
		}

		if itemsJSON != "" {
			var raw interface{}
			if err := json.Unmarshal([]byte(itemsJSON), &raw); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "invalid items json", "detail": err.Error()})
			}
			switch v := raw.(type) {
			case []interface{}:
				for _, e := range v {
					m, ok := e.(map[string]interface{})
					if !ok {
						return c.Status(400).JSON(fiber.Map{"error": "invalid item format in items"})
					}
					maps = append(maps, m)
				}
			case map[string]interface{}:
				maps = append(maps, v)
			default:
				return c.Status(400).JSON(fiber.Map{"error": "invalid items json"})
			}
		} else {
			// 3) fallback: build single item from flat fields (แบบ Postman ที่ส่ง itf_file_name, ip_id,...)
			m := map[string]interface{}{}
			for k, vals := range form.Value {
				if len(vals) > 0 {
					m[k] = vals[0]
				}
			}

			// ถ้าไม่มี field ใดเลย แต่มีไฟล์ ก็ยังให้ผ่านได้โดยสร้าง item จากไฟล์
			if len(m) == 0 && len(uploadedFiles) > 0 {
				m["itf_file_name"] = uploadedFiles[0].Filename
			}

			// ถ้า fallback path ถูกใช้ แต่ไม่มี required fields ให้ return error พร้อม debug info
			if len(m) == 0 || (m["ip_id"] == nil && m["itf_file_name"] == nil) {
				return c.Status(400).JSON(fiber.Map{
					"error":  "no items field provided, and fallback parsing found no valid fields",
					"detail": fmt.Sprintf("Received form fields: %v. Expected either 'items' field with JSON array, or individual fields: ip_id, ipid_id, itf_file_name, itf_type, itf_created_by", receivedFields),
				})
			}

			maps = append(maps, m)
		}
	} else {
		// JSON body
		body := c.Body()
		var raw interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
		}
		switch v := raw.(type) {
		case []interface{}:
			for _, e := range v {
				m, ok := e.(map[string]interface{})
				if !ok {
					return c.Status(400).JSON(fiber.Map{"error": "invalid item format"})
				}
				maps = append(maps, m)
			}
		case map[string]interface{}:
			maps = append(maps, v)
		default:
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
	}

	if len(maps) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no items to insert"})
	}

	// Build items
	var items []item
	for idx, m := range maps {
		it := item{
			FileName: getString(m, "itf_file_name"),
			// Always construct file path from uploaded files, ignore client-provided path
			FilePath:  "",
			IpID:      getInt64(m, "ip_id"),
			IpidID:    getInt64(m, "ipid_id"),
			ItfType:   getString(m, "itf_type"),
			CreatedBy: getString(m, "itf_created_by"),
		}
		// Validate critical fields
		if it.IpID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "ip_id is missing or invalid", "detail": fmt.Sprintf("item[%d]: ip_id=%v (type: %T, raw value: %v)", idx, it.IpID, m["ip_id"], m["ip_id"])})
		}
		if it.IpidID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "ipid_id is missing or invalid", "detail": fmt.Sprintf("item[%d]: ipid_id=%v (raw value: %v)", idx, it.IpidID, m["ipid_id"])})
		}
		if strings.TrimSpace(it.FileName) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "itf_file_name is missing", "detail": fmt.Sprintf("item[%d]: file name cannot be empty", idx)})
		}
		if strings.TrimSpace(it.CreatedBy) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "itf_created_by is missing", "detail": fmt.Sprintf("item[%d]: created by cannot be empty", idx)})
		}
		items = append(items, it)
	}

	// parse optional su_id for notification (can be provided as query/form)
	suIDStr := strings.TrimSpace(c.Query("su_id"))
	if suIDStr == "" {
		suIDStr = strings.TrimSpace(c.FormValue("su_id"))
	}

	// Check if files are provided - required for tracking
	if len(uploadedFiles) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no files uploaded", "detail": "at least one file is required for tracking"})
	}

	// Save files (multipart only)
	if len(uploadedFiles) > 0 {
		// สร้าง slice copy เพื่อ track ไฟล์ที่ใช้แล้ว
		remainingFiles := make([]*multipart.FileHeader, len(uploadedFiles))
		copy(remainingFiles, uploadedFiles)

		// prepare upload base (server absolute path) from ENV with fallback
		uploadBase := os.Getenv("UPLOAD_BASE")
		if strings.TrimSpace(uploadBase) == "" {
			uploadBase = `C:\inetpub\wwwroot\apiTrackingSystemUat\uploads`
		}
		// cache for ip_id -> folder name
		ipCodeCache := map[int64]string{}
		// regexp for sanitizing folder names
		reSan := regexp.MustCompile("[^A-Za-z0-9_-]")

		for i := range items {
			it := &items[i]

			// match file ด้วยชื่อ - sequential matching (เอา file แรกที่ชื่อตรง)
			var matched *multipart.FileHeader
			var matchedIdx int = -1
			for idx, fh := range remainingFiles {
				if fh.Filename == it.FileName {
					matched = fh
					matchedIdx = idx
					break
				}
			}

			if matched == nil {
				return c.Status(400).JSON(fiber.Map{"error": "uploaded file not matched with itf_file_name", "detail": fmt.Sprintf("file '%s' not found in uploaded files", it.FileName)})
			}

			// ลบ file ที่ใช้แล้วออกจาก remainingFiles เพื่อ avoid duplicate usage
			remainingFiles = append(remainingFiles[:matchedIdx], remainingFiles[matchedIdx+1:]...)

			// lookup ip_code (cached)
			ipFolder, ok := ipCodeCache[it.IpID]
			if !ok {
				var ipCode sql.NullString
				if err := db.Get(&ipCode, "SELECT ip_code FROM info_project WHERE ip_id = ? LIMIT 1", it.IpID); err != nil {
					if err == sql.ErrNoRows {
						ipFolder = fmt.Sprintf("ip_%d", it.IpID)
					} else {
						return c.Status(500).JSON(fiber.Map{"error": "query ip_code failed", "detail": err.Error()})
					}
				} else {
					if ipCode.Valid && strings.TrimSpace(ipCode.String) != "" {
						ipFolder = ipCode.String
					} else {
						ipFolder = fmt.Sprintf("ip_%d", it.IpID)
					}
				}
				ipFolder = reSan.ReplaceAllString(ipFolder, "_")
				if ipFolder == "" {
					ipFolder = fmt.Sprintf("ip_%d", it.IpID)
				}
				ipCodeCache[it.IpID] = ipFolder
			}

			// determine itf_type folder (fallback to 'other')
			itfType := strings.TrimSpace(it.ItfType)
			if itfType == "" {
				itfType = "other"
			}
			itfType = reSan.ReplaceAllString(itfType, "_")

			// create dest folder: uploadBase/<ipFolder>/<itfType>
			destDir := filepath.Join(uploadBase, ipFolder, itfType)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "could not create directories", "detail": err.Error()})
			}

			// save file
			savedDiskPath := filepath.Join(destDir, it.FileName)
			if err := c.SaveFile(matched, savedDiskPath); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "could not save uploaded file", "detail": err.Error()})
			}

			// construct DB relative path under 'uploads'
			it.FilePath = filepath.ToSlash(filepath.Join("uploads", ipFolder, itfType, it.FileName))
		}
	}

	// Validate all items have file paths before database operations
	for _, it := range items {
		if strings.TrimSpace(it.FilePath) == "" && strings.TrimSpace(it.FileName) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "file path validation failed", "detail": "item has no file associated"})
		}
	}

	// DB Transaction
	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not begin transaction", "detail": err.Error()})
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	insertStmt := `INSERT INTO info_tracking_file (itf_file_name, itf_file_path, ip_id, ipid_id, itf_type, itf_created_at, itf_created_by, itf_updated_at, itf_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	updateTrackingStmt := `UPDATE info_tracking_file SET itf_file_name = ?, itf_file_path = ?, itf_type = ?, itf_updated_at = ?, itf_updated_by = ? WHERE itf_id = ?`
	updateStmt := `UPDATE info_project_item_detail SET ipid_status = ? WHERE (ref_id, ipid_type) IN (SELECT ref_id, ipid_type FROM (SELECT ref_id, ipid_type FROM info_project_item_detail WHERE ipid_id = ?) AS sq)`

	// Track newly inserted items to send emails only for new items
	newItemIpidIDs := map[int64]bool{}

	for _, it := range items {

		var existingID int64
		errGet := tx.Get(&existingID, "SELECT itf_id FROM info_tracking_file WHERE ip_id = ? AND ipid_id = ? LIMIT 1", it.IpID, it.IpidID)
		switch errGet {
		case nil:
			if _, err := tx.Exec(updateTrackingStmt, it.FileName, it.FilePath, it.ItfType, now, it.CreatedBy, existingID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "update tracking error", "detail": err.Error()})
			}
		case sql.ErrNoRows:
			if _, err := tx.Exec(insertStmt, it.FileName, it.FilePath, it.IpID, it.IpidID, it.ItfType, now, it.CreatedBy, now, it.CreatedBy); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "insert error", "detail": err.Error()})
			}
			// Mark this item as newly inserted (for email sending later)
			newItemIpidIDs[it.IpidID] = true
		default:
			return c.Status(500).JSON(fiber.Map{"error": "select tracking error", "detail": errGet.Error()})
		}

		if _, err := tx.Exec(updateStmt, "waiting", it.IpidID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "update ipid error", "detail": err.Error()})
		}

		// Insert info_approval entries based on workflow for the responsible department
		// 1) find su_id for this ipid (from info_project_item_detail)
		var ownerSu sql.NullInt64
		if err := tx.Get(&ownerSu, `SELECT su_id FROM info_project_item_detail WHERE ipid_id = ? LIMIT 1`, it.IpidID); err != nil {
			// non-fatal if not found; continue
			if err != sql.ErrNoRows {
				return c.Status(500).JSON(fiber.Map{"error": "query su_id failed", "detail": err.Error()})
			}
		}
		if ownerSu.Valid {
			// 2) get sd_id from sys_user for that su_id
			var sdID sql.NullInt64
			if err := tx.Get(&sdID, `SELECT sd_id FROM sys_user WHERE su_id = ? LIMIT 1`, ownerSu.Int64); err != nil && err != sql.ErrNoRows {
				return c.Status(500).JSON(fiber.Map{"error": "query sd_id failed", "detail": err.Error()})
			}
			if sdID.Valid {
				// remove existing approvals for this ipid to avoid duplicates
				if _, err := tx.Exec(`UPDATE info_approval SET ia_status_flg = 'inactive', ia_is_action = 0, ia_updated_at = ?, ia_updated_by = ? WHERE ipid_id = ?`, now, it.CreatedBy, it.IpidID); err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "delete existing approvals failed", "detail": err.Error()})
				}

				// 3) query workflow rows for this department and insert approvals in order
				var wfRows []struct {
					SwOrder sql.NullInt64 `db:"sw_order"`
					SuID    sql.NullInt64 `db:"su_id"`
				}
				if err := tx.Select(&wfRows, `SELECT sw_order, su_id FROM sys_workflow WHERE sd_id = ? AND sw_status = 'active' ORDER BY sw_order ASC`, sdID.Int64); err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "query workflow failed", "detail": err.Error()})
				}
				for idx, r := range wfRows {
					if !r.SuID.Valid || !r.SwOrder.Valid {
						continue
					}
					iaIsAction := 0
					if idx == 0 {
						iaIsAction = 1
					}
					// If an approval row for this ipid_id+su_id already exists with final status (approve/reject),
					// mark the old row as inactive (archive it) before inserting a new approval row.
					if _, err := tx.Exec(`UPDATE info_approval SET ia_status_flg = 'inactive', ia_updated_at = ?, ia_updated_by = ? WHERE ipid_id = ? AND su_id = ? AND ia_status IN ('approve','reject')`, now, it.CreatedBy, it.IpidID, r.SuID.Int64); err != nil {
						return c.Status(500).JSON(fiber.Map{"error": "deactivate old approvals failed", "detail": err.Error()})
					}
					if _, err := tx.Exec(`INSERT INTO info_approval (ipid_id, su_id, ia_level, ia_status, ia_is_action, ia_round, ia_created_at, ia_created_by, ia_updated_at, ia_updated_by, ia_status_flg, ia_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
						it.IpidID, r.SuID.Int64, r.SwOrder.Int64, "waiting", iaIsAction, 0, now, it.CreatedBy, now, it.CreatedBy, "active", "Leader"); err != nil {
						return c.Status(500).JSON(fiber.Map{"error": "insert info_approval failed", "detail": err.Error()})
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "commit error", "detail": err.Error()})
	}

	// If su_id provided, send notification email (behaviour from SaveFileSendEmail)
	// Only send for newly inserted items, not for updates
	if len(items) > 0 && len(newItemIpidIDs) > 0 {
		ipID := items[0].IpID

		// Collect all ipid_ids that were sent in this request
		var sentIpidIDs []int64
		for ipidID := range newItemIpidIDs {
			sentIpidIDs = append(sentIpidIDs, ipidID)
		}

		// Send email to each approver with their assigned items
		type MailData struct {
			ProjectCode sql.NullString `db:"ip_code"`
			PartName    sql.NullString `db:"ip_part_name"`
			PartNo      sql.NullString `db:"ip_part_no"`
			IpModel     sql.NullString `db:"ip_model"`
			ItemName    sql.NullString `db:"item_name"`
			ItemType    sql.NullString `db:"item_type"`
			StartDate   sql.NullTime   `db:"start_date"`
			EndDate     sql.NullTime   `db:"end_date"`
		}

		// Group all items by approver to send consolidated emails
		approverItems := map[int64][]MailData{}

		// Get all unique approvers for the sent items
		var approverSuIDs []int64
		q, args, err := sqlx.In(`
			SELECT DISTINCT sw.su_id
			FROM sys_workflow sw
			JOIN (
				SELECT DISTINCT pid.sd_id
				FROM info_project_item_detail pid
				WHERE pid.ipid_id IN (?) AND pid.sd_id IS NOT NULL
			) item_sds ON item_sds.sd_id = sw.sd_id
			WHERE sw.sw_status = 'active'
			ORDER BY sw.sw_order ASC
		`, sentIpidIDs)
		if err == nil {
			q = db.Rebind(q)
			_ = db.Select(&approverSuIDs, q, args...)
		}

		// For each approver, collect ONLY the sent items
		for _, approverSuID := range approverSuIDs {
			// Get items assigned to this approver - FILTERED to ONLY sent ipid_ids
			q, args, err := sqlx.In(`SELECT DISTINCT
					ip.ip_code,
					ip.ip_part_name,
					ip.ip_part_no,
					ip.ip_model,
					x.item_name,
					x.item_type,
					x.start_date,
					x.end_date
				FROM
				(
						SELECT
							ai.ip_id                                   AS ip_id,
							ai.iai_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							pid.ipid_start_date                          AS start_date,
							pid.ipid_end_date                          AS end_date,
							pid.ipid_id                                AS ipid_id
						FROM info_project_item_detail pid
						JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id AND pid.ipid_type = 'apqp'
						WHERE ai.ip_id = ? AND pid.ipid_id IN (?)

						UNION ALL

						SELECT
							pi.ip_id                                   AS ip_id,
							pi.ipi_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							pid.ipid_start_date                       AS start_date,
							pid.ipid_end_date                          AS end_date,
							pid.ipid_id                                AS ipid_id
						FROM info_project_item_detail pid
						JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id AND pid.ipid_type = 'ppap'
						WHERE pi.ip_id = ? AND pid.ipid_id IN (?)
				) x
				LEFT JOIN info_project ip ON x.ip_id = ip.ip_id
				LEFT JOIN info_approval ia ON ia.ipid_id = x.ipid_id AND ia.su_id = ? AND ia_status = 'waiting'
				WHERE ia.su_id IS NOT NULL
				ORDER BY
					x.item_type ASC,
					x.start_date ASC,
					x.item_name ASC`, ipID, sentIpidIDs, ipID, sentIpidIDs, approverSuID)

			if err != nil {
				continue
			}
			q = db.Rebind(q)

			var list []MailData
			if err := db.Select(&list, q, args...); err != nil || len(list) == 0 {
				continue
			}

			// Set items to approver's list (replace, not append)
			approverItems[approverSuID] = list
		}

		// Now send one consolidated email per approver with all their items
		for approverSuID, allItems := range approverItems {
			// Get approver email
			var suEmail sql.NullString
			if err := db.Get(&suEmail, `SELECT su_email FROM sys_user WHERE su_id = ? AND su_status = 'active' AND su_email IS NOT NULL AND su_email <> '' LIMIT 1`, approverSuID); err != nil || !suEmail.Valid {
				continue
			}

			// Get approver's first and last name
			var approverFirstName, approverLastName sql.NullString
			db.QueryRow(`SELECT su_firstname, su_lastname FROM sys_user WHERE su_id = ? LIMIT 1`, approverSuID).Scan(&approverFirstName, &approverLastName)
			var approverName string
			if approverFirstName.Valid && approverLastName.Valid {
				approverName = approverFirstName.String + " " + approverLastName.String
			} else if approverFirstName.Valid {
				approverName = approverFirstName.String
			} else if approverLastName.Valid {
				approverName = approverLastName.String
			}
			if approverName == "" {
				approverName = fmt.Sprintf("%v", approverSuID)
			}

			if len(allItems) == 0 {
				continue
			}

			// Build HTML body with card layout showing first item's project details
			var sb strings.Builder
			var i int

			sb.WriteString("<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>Dear, K." + html.EscapeString(approverName) + "</h3>")
			sb.WriteString("<h4>You have item for <b style='color : #089633;'>Approval</b> in website Project Management</h4>")
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
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>#" + html.EscapeString(getStringValue(allItems[0].ProjectCode)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:50%; border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>MODEL</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(allItems[0].IpModel)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("</tr>")
			sb.WriteString("<tr>")

			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NO</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(allItems[0].PartNo)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td colspan='1' style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NAME</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(allItems[0].PartName)) + "</div>")
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

			for _, r := range allItems {
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
				i++
			}

			sb.WriteString("</tbody></table>")

			sb.WriteString("<div style='margin-top:20px;'>")
			sb.WriteString("<a href='http://192.168.161.205:4005/login' style='display:inline-block; padding:10px 20px; background:#2563eb; color:#fff; text-decoration:none; border-radius:6px; font-weight:bold; font-size:14px;'>Open Project Management</a>")
			sb.WriteString("</div>")

			sb.WriteString("<div style='margin-top:30px; padding-top:20px; border-top:1px solid #e5e7eb; font-size:13px; color:#6b7280;'>")
			sb.WriteString("<p>Best Regards,<br><strong>System Service Department</strong></p>")
			sb.WriteString("</div>")

			sb.WriteString("</body></html>")

			subject := "TBKK Project Control Notification : waiting Leader Approval"
			body := sb.String()
			_ = SendMail([]string{suEmail.String}, subject, body, "text/html; charset=utf-8")
		}
	}

	// done
	if len(items) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "ok", "inserted": 0})
	}
	return c.Status(201).JSON(fiber.Map{"status": "ok", "inserted": len(items)})
}
