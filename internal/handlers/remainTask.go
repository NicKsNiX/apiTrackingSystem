package handlers

import (
	"database/sql"
	"errors"
	"html"
	"strconv"
	"strings"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type RemainTask struct {
	ID                 int64            `db:"ip_id" json:"ip_id"`
	Ipid               int64            `db:"ipid_id" json:"ipid_id"`
	Name               utils.NullString `db:"ip_name" json:"ip_name"`
	Code               utils.NullString `db:"ip_code" json:"ip_code"`
	Model              utils.NullString `db:"ip_model" json:"ip_model"`
	PartName           utils.NullString `db:"ip_part_name" json:"ip_part_name"`
	PartNo             utils.NullString `db:"ip_part_no" json:"ip_part_no"`
	Description        utils.NullString `db:"ip_description" json:"ip_description"`
	Status             string           `db:"ip_status" json:"ip_status"`
	IppapName          utils.NullString `db:"ipap_name" json:"ipap_name"`
	CreatedAt          *time.Time       `db:"ip_created_at" json:"ip_created_at"`
	CreatedBy          utils.NullString `db:"ip_created_by" json:"ip_created_by"`
	UpdatedAt          *time.Time       `db:"ip_updated_at" json:"ip_updated_at"`
	UpdatedBy          utils.NullString `db:"ip_updated_by" json:"ip_updated_by"`
	UpdatedByFirstName utils.NullString `db:"ip_updated_by_firstname" json:"ip_updated_by_firstname"`
	UpdatedByLastName  utils.NullString `db:"ip_updated_by_lastname" json:"ip_updated_by_lastname"`
}

func ListRemainTasks(c *fiber.Ctx, db *sqlx.DB) error {
	// parse optional sd_id query parameter
	sdStr := strings.TrimSpace(c.Query("sd_id"))
	var sdParam interface{} = nil
	if sdStr != "" {
		if v, err := strconv.ParseInt(sdStr, 10, 64); err == nil {
			sdParam = v
		} else {
			return c.Status(400).JSON(fiber.Map{"error": "invalid sd_id"})
		}
	}

	// require su_id query parameter
	suStr := strings.TrimSpace(c.Query("su_id"))
	var suParam interface{} = nil
	if suStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "su_id required"})
	}
	if v, err := strconv.ParseInt(suStr, 10, 64); err == nil {
		suParam = v
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "invalid su_id"})
	}

	query := `SELECT DISTINCT
					x.ipid_id,
					x.ip_code,
					x.ip_part_no,
					x.item_name,
					x.ip_model,
					 CASE
                            WHEN a.ia_status IS NULL THEN x.ipid_status
                            WHEN x.ipid_status = 'waiting' AND a.ia_status = 'approve' THEN a.ia_status
                            ELSE a.ia_status
                        END AS status_approve,
					x.owner_su_id,
					su.su_emp_code,
					su.su_firstname,
					su.su_lastname,
					x.ipid_updated_at,
					x.itf_file_name,
					x.itf_file_path
				FROM
				(
						SELECT
							ai.mpp_id                                  AS mpp_id, 
							ip.ip_code                                 AS ip_code,
							ip.ip_part_no                              AS ip_part_no,
							ip.ip_model                                AS ip_model,
							ai.iai_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							sd.sd_dept_aname                           AS department,
							pid.su_id                                  AS owner_su_id,
							ai.iai_created_at                          AS start_date,
							pid.ipid_updated_at                          AS ipid_updated_at,
							pid.ipid_id                                AS ipid_id,
							pid.ipid_status                           AS ipid_status,
							tf.itf_file_name                           AS itf_file_name,
							tf.itf_file_path                           AS itf_file_path
						FROM info_project_item_detail pid
						JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id AND pid.ipid_type = 'apqp'
						JOIN sys_department sd ON sd.sd_id = pid.sd_id
						LEFT JOIN info_project ip ON ip.ip_id = ai.ip_id
						RIGHT JOIN info_tracking_file tf ON tf.ipid_id = pid.ipid_id
						WHERE pid.sd_id = ?

						UNION ALL

						SELECT
							NULL                                       AS mpp_id,                              
							ip.ip_code                                 AS ip_code,
							ip.ip_part_no                              AS ip_part_no,
							ip.ip_model                                AS ip_model,
							pi.ipi_name                                AS item_name,
							pid.ipid_type                              AS item_type,
							sd.sd_dept_aname                           AS department,
							pid.su_id                                  AS owner_su_id,
							pi.ipi_created_at                          AS start_date,
							pid.ipid_updated_at                          AS ipid_updated_at,
							pid.ipid_id                                AS ipid_id,
							pid.ipid_status                           AS ipid_status,
							tf.itf_file_name                           AS itf_file_name,
							tf.itf_file_path                           AS itf_file_path
						FROM info_project_item_detail pid
						JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id  AND pid.ipid_type = 'ppap'
						JOIN sys_department sd ON sd.sd_id = pid.sd_id
						LEFT JOIN info_project ip ON ip.ip_id = pi.ip_id
						RIGHT JOIN info_tracking_file tf ON tf.ipid_id = pid.ipid_id
						WHERE pid.sd_id = ?
							
				) x
				LEFT JOIN info_approval a
					ON a.ipid_id = x.ipid_id
				LEFT JOIN sys_user su
					ON su.su_id = x.owner_su_id
				LEFT JOIN sys_workflow sw 
				    ON x.owner_su_id = sw.su_id 
				WHERE a.ia_status = 'waiting' AND a.ia_is_action = 1 AND a.su_id = ? AND a.ia_type = 'Leader'
				ORDER BY
					x.ip_code ASC,
					x.item_type ASC,
					x.mpp_id ASC;`

	// placeholders: sd_id (apqp), sd_id (ppap), su_id
	args := []interface{}{sdParam, sdParam, suParam}

	var rows []struct {
		IpidID        int64            `db:"ipid_id" json:"ipid_id"`
		IpCode        utils.NullString `db:"ip_code" json:"ip_code"`
		IpPartNo      utils.NullString `db:"ip_part_no" json:"ip_part_no"`
		ItemName      utils.NullString `db:"item_name" json:"item_name"`
		IpModel       utils.NullString `db:"ip_model" json:"ip_model"`
		IpidUpdateAt  *time.Time       `db:"ipid_updated_at" json:"ipid_updated_at"`
		StatusApprove utils.NullString `db:"status_approve" json:"status_approve"`
		OwnerSuID     utils.NullInt64  `db:"owner_su_id" json:"owner_su_id"`
		SuEmpCode     utils.NullString `db:"su_emp_code" json:"su_emp_code"`
		SuFirstName   utils.NullString `db:"su_firstname" json:"su_firstname"`
		SuLastName    utils.NullString `db:"su_lastname" json:"su_lastname"`
		ItfFileName   utils.NullString `db:"itf_file_name" json:"itf_file_name"`
		ItfFilePath   utils.NullString `db:"itf_file_path" json:"itf_file_path"`
	}

	if err := db.Select(&rows, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	return c.Status(200).JSON(rows)
}

func NotifyRemainTasks(c *fiber.Ctx, db *sqlx.DB) error {
	var r RemainTask
	query := `SELECT COUNT(*) FROM info_project_master_plan WHERE ipmp_status = 'inprogress'`
	if err := db.Get(&r, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(r)
}

func SelectModelMaster(c *fiber.Ctx, db *sqlx.DB) error {
	var models []struct {
		ID    int64  `db:"mmm_id" json:"mmm_id"`
		Model string `db:"mmm_model" json:"mmm_model"`
	}
	if err := db.Select(&models, `SELECT mmm_id, mmm_model FROM mst_model_master WHERE mmm_status = 'active' ORDER BY mmm_model ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(models)
}

func SelectPartNumberRT(c *fiber.Ctx, db *sqlx.DB) error {
	var parts []struct {
		ID     int64  `db:"ip_id" json:"ip_id"`
		PartNo string `db:"ip_part_no" json:"ip_part_no"`
	}
	if err := db.Select(&parts, `SELECT ip_part_no FROM info_project WHERE ip_status = 'active' ORDER BY ip_part_no ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(parts)
}

func SelectKickOffDateRT(c *fiber.Ctx, db *sqlx.DB) error {
	var parts []struct {
		ID     int64  `db:"ip_id" json:"ip_id"`
		PartNo string `db:"ip_kickoff_date" json:"ip_kickoff_date"`
	}
	if err := db.Select(&parts, `SELECT ip_kickoff_date FROM info_project WHERE ip_status = 'active' ORDER BY ip_kickoff_date ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(parts)
}

func GetCountItem(c *fiber.Ctx, db *sqlx.DB) error {
	var r struct {
		CountItemUp int64 `db:"count_item" json:"count_item"`
	}
	if err := db.Get(&r, `SELECT COUNT(*) AS count_item
							FROM info_project_item_detail ipid
							RIGHT JOIN info_tracking_file itf
							        ON ipid.ipid_id = itf.ipid_id 
							LEFT JOIN info_approval ia
							        ON ipid.ipid_id = ia.ipid_id
							LEFT JOIN sys_workflow sw
							        ON ia.su_id = sw.su_id 

						 	WHERE ia.ia_status = 'waiting' AND sw.su_id = ? AND sw.sw_order <> 0 AND ia_type = 'Leader'`, c.Query("su_id")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(r)
}

func UpdateStatusFileProject(c *fiber.Ctx, db *sqlx.DB) error {
	// Try query params first
	ipidIDStr := c.Query("ipid_id")
	newStatus := c.Query("new_status")
	updatedBy := c.Query("updated_by")
	note := c.Query("note")

	// Fallback to form values
	if ipidIDStr == "" {
		ipidIDStr = c.FormValue("ipid_id")
	}
	if newStatus == "" {
		newStatus = c.FormValue("new_status")
	}
	if updatedBy == "" {
		updatedBy = c.FormValue("updated_by")
	}
	if note == "" {
		note = c.FormValue("note")
	}

	// Accept JSON body as fallback
	if ipidIDStr == "" || newStatus == "" || updatedBy == "" {
		var body struct {
			IpidID    int64  `json:"ipid_id"`
			NewStatus string `json:"new_status"`

			UpdatedBy string `json:"updated_by"`
			Note      string `json:"note"`
		}
		if err := c.BodyParser(&body); err == nil {
			if ipidIDStr == "" && body.IpidID != 0 {
				ipidIDStr = strconv.FormatInt(body.IpidID, 10)
			}
			if newStatus == "" && body.NewStatus != "" {
				newStatus = body.NewStatus
			}

			if updatedBy == "" && body.UpdatedBy != "" {
				updatedBy = body.UpdatedBy
			}

			// accept optional note for reject
			if note == "" && body.Note != "" {
				note = body.Note
			}
		}
	}

	// Validate required fields
	if ipidIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_id is required"})
	}

	ipidID, err := strconv.ParseInt(ipidIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_id format"})
	}

	if newStatus == "" {
		return c.Status(400).JSON(fiber.Map{"error": "new_status (or status) is required"})
	}
	if updatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "updated_by is required"})
	}

	// Validate status value
	validIPIDStatuses := map[string]bool{
		"done": true, "inprogress": true, "delay": true, "reject": true, "waiting": true,
	}
	if !validIPIDStatuses[newStatus] {
		return c.Status(400).JSON(fiber.Map{"error": "invalid new_status value"})
	}

	// Map to approval status
	approvalStatus := "waiting"
	switch newStatus {
	case "done":
		approvalStatus = "approve"
	case "reject":
		approvalStatus = "reject"
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "transaction error", "detail": err.Error()})
	}
	defer func() { _ = tx.Rollback() }()

	// Do not update ipid_status here per request. Verify the ipid exists instead.
	var exists int
	if err := tx.Get(&exists, `SELECT 1 FROM info_project_item_detail WHERE ipid_id = ? LIMIT 1`, ipidID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(404).JSON(fiber.Map{"error": "ipid_id not found", "ipid_id": ipidID})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to verify ipid existence", "detail": err.Error()})
	}

	// Special handling for reject status - update all waiting Leader approvals
	if newStatus == "reject" {
		_, err = tx.Exec(`
			UPDATE info_approval
			SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_note = ?
			WHERE ipid_id = ? AND ia_status = 'waiting' AND ia_type = 'Leader'
		`, approvalStatus, now, updatedBy, note, ipidID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update Leader approvals on reject", "detail": err.Error()})
		}
	}

	// Update info_approval (priority: action row first, then promote next Leader or insert PJ)
	// 1) load current action row (ia_is_action = 1)
	var curr struct {
		IAID        sql.NullInt64  `db:"ia_id"`
		SuID        sql.NullInt64  `db:"su_id"`
		IaLevel     sql.NullInt64  `db:"ia_level"`
		IaStatus    sql.NullString `db:"ia_status"`
		IaType      sql.NullString `db:"ia_type"`
		IaStatusFlg sql.NullString `db:"ia_status_flg"`
		IaRound     sql.NullInt64  `db:"ia_round"`
		IaCreatedBy sql.NullString `db:"ia_created_by"`
	}
	err = tx.Get(&curr, `SELECT ia_id, su_id, ia_level, ia_status, ia_type, ia_status_flg, ia_round, ia_created_by FROM info_approval WHERE ipid_id = ? AND ia_is_action = 1 LIMIT 1`, ipidID)
	if err != nil && err != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch current approval action", "detail": err.Error()})
	}

	// If we have a current action row, update it and then try to promote next Leader or insert PJ
	if err == nil {
		// update current action row: set status and clear ia_is_action
		if newStatus == "reject" {
			if _, err = tx.Exec(`UPDATE info_approval SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_note = ? WHERE ia_id = ?`, approvalStatus, now, updatedBy, note, curr.IAID.Int64); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to update current approval action", "detail": err.Error()})
			}
		} else {
			if _, err = tx.Exec(`UPDATE info_approval SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_is_action = 0 WHERE ia_id = ?`, approvalStatus, now, updatedBy, curr.IAID.Int64); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to update current approval action", "detail": err.Error()})
			}
		}

		// find next Leader waiting & active with minimal ia_level
		var nextID sql.NullInt64
		if err := tx.Get(&nextID, `SELECT ia_id FROM info_approval WHERE ipid_id = ? AND ia_status = 'waiting' AND ia_type = 'Leader' AND ia_status_flg = 'active' ORDER BY ia_level ASC LIMIT 1`, ipidID); err != nil {
			if err != sql.ErrNoRows {
				return c.Status(500).JSON(fiber.Map{"error": "failed to query next leader approval", "detail": err.Error()})
			}
		}

		if nextID.Valid {
			// promote that row to action
			if _, err := tx.Exec(`UPDATE info_approval SET ia_is_action = 1, ia_updated_at = ?, ia_updated_by = ? WHERE ia_id = ?`, now, updatedBy, nextID.Int64); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to promote next approval", "detail": err.Error()})
			}
		} else {
			// no leader waiting -> normally create new PJ approval row
			// but if this update was a reject, do NOT insert a new PJ row
			if newStatus == "reject" {
				// skip inserting PJ on reject
			} else {
				// ensure we have at least su_id from current; if not, fallback to inserting a generic row with created_by
				suIDVal := interface{}(nil)
				if curr.SuID.Valid {
					suIDVal = curr.SuID.Int64
				}
				iaLevel := int64(0)
				if curr.IaLevel.Valid {
					iaLevel = curr.IaLevel.Int64
				}
				iaStatusFlg := "active"
				if curr.IaStatusFlg.Valid && strings.TrimSpace(curr.IaStatusFlg.String) != "" {
					iaStatusFlg = curr.IaStatusFlg.String
				}

				if _, err := tx.Exec(`INSERT INTO info_approval (ipid_id, su_id, ia_level, ia_status, ia_is_action, ia_round, ia_created_at, ia_created_by, ia_updated_at, ia_updated_by, ia_status_flg, ia_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					ipidID, suIDVal, iaLevel, "waiting", 1, 0, now, curr.IaCreatedBy.String, now, curr.IaCreatedBy.String, iaStatusFlg, "PJ"); err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "failed to insert PJ approval row", "detail": err.Error()})
				}
			}
		}

	} else {
		// no current action row: fallback to previous behaviour - update all approvals for ipid_id
		var res3 sql.Result
		if newStatus == "reject" {
			res3, err = tx.Exec(`
				UPDATE info_approval
				SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?, ia_note = ?
				WHERE ipid_id = ?
			`, approvalStatus, now, updatedBy, note, ipidID)
		} else {
			res3, err = tx.Exec(`
				UPDATE info_approval
				SET ia_status = ?, ia_updated_at = ?, ia_updated_by = ?
				WHERE ipid_id = ?
			`, approvalStatus, now, updatedBy, ipidID)
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update approval status (fallback)", "detail": err.Error()})
		}

		ra3, _ := res3.RowsAffected()
		if ra3 == 0 {
			// ipid updated but no approval rows exist - still return success
			if err := tx.Commit(); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "commit error", "detail": err.Error()})
			}
			return c.Status(200).JSON(fiber.Map{"message": "status updated (no approval rows found)", "ipid_id": ipidID})
		}
	}

	// If approved (done), increment ia_round for approvals where the approver's sd_id matches the item's sd_id
	if newStatus == "done" {
		var sdID sql.NullInt64
		if err := tx.Get(&sdID, `SELECT sd_id FROM info_project_item_detail WHERE ipid_id = ? LIMIT 1`, ipidID); err == nil && sdID.Valid {
			// update ia_round for active approvers in the same department
			_, _ = tx.Exec(`
				UPDATE info_approval a
				JOIN sys_user su ON a.su_id = su.su_id
				SET a.ia_round = a.ia_round + 1
				WHERE a.ipid_id = ? AND a.ia_is_action = 1 AND su.sd_id = ?
			`, ipidID, sdID.Int64)
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "commit error", "detail": err.Error()})
	}

	// If there are still active Leader approvals waiting for this ipid, skip sending emails
	var pendingLeaderCount int
	if err := db.Get(&pendingLeaderCount, `SELECT COUNT(*) FROM info_approval WHERE ipid_id = ? AND ia_status = 'waiting' AND ia_type = 'Leader' AND ia_status_flg = 'active' AND ia_is_action = 1`, ipidID); err == nil {
		if pendingLeaderCount > 0 {
			return c.Status(200).JSON(fiber.Map{"message": "emails skipped - pending leader approvals remain", "pending_leader_count": pendingLeaderCount})
		}
	}

	// After successful commit, send notification emails for specific status changes
	// Build detail row from info_project_item_detail (join project and item name)
	var detail struct {
		ProjectCode sql.NullString `db:"ip_code"`
		PartNo      sql.NullString `db:"ip_part_no"`
		PartName    sql.NullString `db:"ip_part_name"`
		IpModel     sql.NullString `db:"ip_model"`
		ItemName    sql.NullString `db:"item_name"`
		ItemType    sql.NullString `db:"item_type"`
		StartDate   sql.NullTime   `db:"ipid_start_date"`
		EndDate     sql.NullTime   `db:"ipid_end_date"`
		Status      sql.NullString `db:"ipid_status"`
		OwnerSuID   sql.NullInt64  `db:"su_id"`
	}

	q := `SELECT
		ip.ip_code,
		ip.ip_part_no,
		ip.ip_part_name,
		ip.ip_model,
		COALESCE(ai.iai_name, pi.ipi_name) AS item_name,
		pid.ipid_type AS item_type,
		pid.ipid_start_date,
		pid.ipid_end_date,
		pid.ipid_status,
		pid.su_id
	FROM info_project_item_detail pid
	LEFT JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id AND pid.ipid_type = 'apqp'
	LEFT JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id AND pid.ipid_type = 'ppap'
	LEFT JOIN info_project ip ON ip.ip_id = COALESCE(ai.ip_id, pi.ip_id)
	WHERE pid.ipid_id = ? LIMIT 1`

	if err := db.Get(&detail, q, ipidID); err == nil {
		// HTML template (full-width plain table)
		var sb strings.Builder

		// handle notifications based on newStatus using a tagged switch
		switch newStatus {
		case "reject":
			// send to owner
			var ownerEmail sql.NullString
			var ownerFirstName sql.NullString
			if detail.OwnerSuID.Valid {
				_ = db.Get(&ownerEmail, `SELECT su_email FROM sys_user WHERE su_id = ? AND su_status = 'active' AND su_email IS NOT NULL AND su_email <> '' LIMIT 1`, detail.OwnerSuID.Int64)
				_ = db.Get(&ownerFirstName, `SELECT su_firstname FROM sys_user WHERE su_id = ? AND su_status = 'active' LIMIT 1`, detail.OwnerSuID.Int64)
			}
			// Get approver/rejector name
			var approverFirstName sql.NullString
			var approverLastName sql.NullString
			_ = db.Get(&approverFirstName, `SELECT su_firstname FROM sys_user WHERE su_emp_code = ? OR su_id = CAST(? AS UNSIGNED) LIMIT 1`, updatedBy, updatedBy)
			_ = db.Get(&approverLastName, `SELECT su_lastname FROM sys_user WHERE su_emp_code = ? OR su_id = CAST(? AS UNSIGNED) LIMIT 1`, updatedBy, updatedBy)
			sb.WriteString("<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>Dear, K." + html.EscapeString(getStringValue(ownerFirstName)) + "</h3>")

			sb.WriteString("<h4>This project item has been <b style='color: #dc2626;'>rejected</b> by " + html.EscapeString(getStringValue(approverFirstName)) + " " + html.EscapeString(getStringValue(approverLastName)) + "</h4>")
			sb.WriteString("<html><body style='font-family: Arial, sans-serif; background:#ffffff; padding:20px;'>")
			if strings.TrimSpace(note) != "" {
				sb.WriteString("<div style='margin-top:15px; padding:12px; background:#fee2e2; border-left:4px solid #dc2626; color:#7f1d1d; font-size:13px;'>")
				sb.WriteString("<b>Reason for Rejection:</b><br>")
				sb.WriteString(html.EscapeString(note))
				sb.WriteString("</div><br>")
			}

			sb.WriteString("<div style='margin:auto; background:#ffffff; border-radius:10px; border:1px solid #e0e6ed; padding:20px;'>")
			sb.WriteString("<div style='font-size:18px; font-weight:bold; color:#1f2d3d;'>Project Detail</div>")
			sb.WriteString("<div style='font-size:12px; color:#6b7280;'>Header information for project detail (ข้อมูลของโปรเจค)</div>")

			sb.WriteString("<hr style='border:none; border-top:1px dashed #d1d5db; margin:15px 0;'>")

			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse;'>")

			sb.WriteString("<tr>")
			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PROJECT CODE</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>#" + html.EscapeString(getStringValue(detail.ProjectCode)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>MODEL</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.IpModel)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>TEMPLATE</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>MITSUBISHI MOTOR</div>")
			sb.WriteString("</td>")
			sb.WriteString("</tr>")

			sb.WriteString("<tr>")
			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NO</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartNo)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td colspan='2' style='border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NAME</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartName)) + "</div>")
			sb.WriteString("</td>")
			sb.WriteString("</tr>")

			sb.WriteString("</table>")
			sb.WriteString("<div style='margin-top:20px; font-size:14px; color:#374151;'>")
			sb.WriteString("</div>")
			sb.WriteString("</div><br>")

			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse; border:1px solid #e5e7eb;'>")
			sb.WriteString("<thead><tr style='background:#f3f4f6; border-bottom:2px solid #d1d5db;'>")
			cols := []string{"Item Name", "Item Type", "Start Date", "End Date"}
			for _, c := range cols {
				sb.WriteString("<th>" + html.EscapeString(c) + "</th>")
			}
			sb.WriteString("</tr></thead><tbody><tr>")
			vals := []string{getStringValue(detail.ItemName), getStringValue(detail.ItemType), getDateValue(detail.StartDate), getDateValue(detail.EndDate)}
			for _, v := range vals {
				sb.WriteString("<td>" + html.EscapeString(v) + "</td>")
			}
			sb.WriteString("</tr>")
			sb.WriteString("</tbody></table>")

			sb.WriteString("<div style='margin-top:20px;'>")
			sb.WriteString("<a href='http://192.168.161.205:4005/login' style='display:inline-block; padding:10px 20px; background:#2563eb; color:#fff; text-decoration:none; border-radius:6px; font-weight:bold; font-size:14px;'>Open Project Management</a>")
			sb.WriteString("</div>")

			sb.WriteString("<div style='margin-top:30px; padding-top:20px; border-top:1px solid #e5e7eb; font-size:13px; color:#6b7280;'>")
			sb.WriteString("<p>Best Regards,<br><strong>System Service Department</strong></p>")
			sb.WriteString("</div>")

			sb.WriteString("</body></html>")

			if ownerEmail.Valid && strings.TrimSpace(ownerEmail.String) != "" {
				_ = SendMail([]string{ownerEmail.String}, "TBKK Project Control Notification : Your project has been rejected", sb.String(), "text/html; charset=utf-8")
			}

		case "done":
			// Get approver name
			var approverFirstName sql.NullString
			var approverLastName sql.NullString
			_ = db.Get(&approverFirstName, `SELECT su_firstname FROM sys_user WHERE su_emp_code = ? OR su_id = CAST(? AS UNSIGNED) LIMIT 1`, updatedBy, updatedBy)
			_ = db.Get(&approverLastName, `SELECT su_lastname FROM sys_user WHERE su_emp_code = ? OR su_id = CAST(? AS UNSIGNED) LIMIT 1`, updatedBy, updatedBy)

			// send to PROJECT CONTROL department users
			sb.WriteString("<html><body style='font-family: Arial, sans-serif; background:#ffffff; padding:20px;'>")
			sb.WriteString("<h4>This project item has been <b style='color: #10b981;'>Approved</b> by " + html.EscapeString(getStringValue(approverFirstName)) + " " + html.EscapeString(getStringValue(approverLastName)) + "</h4>")
			sb.WriteString("<div style='margin:auto; background:#ffffff; border-radius:10px; border:1px solid #e0e6ed; padding:20px;'>")
			sb.WriteString("<div style='font-size:18px; font-weight:bold; color:#1f2d3d;'>Project Detail</div>")
			sb.WriteString("<div style='font-size:12px; color:#6b7280;'>Header information for project detail (ข้อมูลของโปรเจค)</div>")

			sb.WriteString("<hr style='border:none; border-top:1px dashed #d1d5db; margin:15px 0;'>")

			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse;'>")

			sb.WriteString("<tr>")
			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PROJECT CODE</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>#" + html.EscapeString(getStringValue(detail.ProjectCode)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>MODEL</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.IpModel)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td style='width:33%; border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>TEMPLATE</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>MITSUBISHI MOTOR</div>")
			sb.WriteString("</td>")
			sb.WriteString("</tr>")

			sb.WriteString("<tr>")
			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NO</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartNo)) + "</div>")
			sb.WriteString("</td>")

			sb.WriteString("<td colspan='2' style='border:1px solid #e5e7eb; border-radius:8px;'>")
			sb.WriteString("<div style='font-size:12px; color:#374151;'>PART NAME</div>")
			sb.WriteString("<div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStringValue(detail.PartName)) + "</div>")
			sb.WriteString("</td>")
			sb.WriteString("</tr>")

			sb.WriteString("</table>")
			sb.WriteString("<div style='margin-top:20px; font-size:14px; color:#374151;'>")
			sb.WriteString("</div>")
			sb.WriteString("</div><br>")

			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse; border:1px solid #e5e7eb;'>")
			sb.WriteString("<thead><tr style='background:#f3f4f6; border-bottom:2px solid #d1d5db;'>")
			cols := []string{"Item Name", "Item Type", "Start Date", "End Date"}
			for _, c := range cols {
				sb.WriteString("<th>" + html.EscapeString(c) + "</th>")
			}
			sb.WriteString("</tr></thead><tbody><tr>")
			vals := []string{getStringValue(detail.ItemName), getStringValue(detail.ItemType), getDateValue(detail.StartDate), getDateValue(detail.EndDate)}
			for _, v := range vals {
				sb.WriteString("<td>" + html.EscapeString(v) + "</td>")
			}
			sb.WriteString("</tr></tbody></table>")

			sb.WriteString("<div style='margin-top:20px;'>")
			sb.WriteString("<a href='http://192.168.161.205:4005/login' style='display:inline-block; padding:10px 20px; background:#2563eb; color:#fff; text-decoration:none; border-radius:6px; font-weight:bold; font-size:14px;'>Open Project Management</a>")
			sb.WriteString("</div>")

			sb.WriteString("<div style='margin-top:30px; padding-top:20px; border-top:1px solid #e5e7eb; font-size:13px; color:#6b7280;'>")
			sb.WriteString("<p>Best Regards,<br><strong>System Service Department</strong></p>")
			sb.WriteString("</div>")

			sb.WriteString("</body></html>")

			// Send email to PJ (ipid_created_by)
			var pjData []struct {
				Email     string `db:"su_email"`
				FirstName string `db:"su_firstname"`
			}
			_ = db.Select(&pjData, `SELECT su.su_email, su.su_firstname FROM sys_user su LEFT JOIN info_project_item_detail pid ON su.su_emp_code = pid.ipid_created_by WHERE pid.ipid_id = ? AND su.su_status = 'active' AND su.su_email IS NOT NULL AND su.su_email <> ''`, ipidID)
			if len(pjData) > 0 {
				for _, pj := range pjData {
					pjEmailBody :=
						`<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>` +
							`Dear, K.` + html.EscapeString(pj.FirstName) + `<br><br>` +
							`</h3>` +
							sb.String()
					_ = SendMail([]string{pj.Email}, "TBKK Project Control Notification : waiting Approval", pjEmailBody, "text/html; charset=utf-8")
				}
			}

			// Send email to owner (su_id)
			if detail.OwnerSuID.Valid {
				var ownerEmail sql.NullString
				var ownerFirstName sql.NullString
				_ = db.Get(&ownerEmail, `SELECT su_email FROM sys_user WHERE su_id = ? AND su_status = 'active' AND su_email IS NOT NULL AND su_email <> '' LIMIT 1`, detail.OwnerSuID.Int64)
				_ = db.Get(&ownerFirstName, `SELECT su_firstname FROM sys_user WHERE su_id = ? AND su_status = 'active' LIMIT 1`, detail.OwnerSuID.Int64)
				if ownerEmail.Valid && strings.TrimSpace(ownerEmail.String) != "" {
					ownerEmailBody :=
						`<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>` +
							`Dear, K.` + html.EscapeString(getStringValue(ownerFirstName)) + `<br><br>` +
							`</h3>` +
							sb.String()
					_ = SendMail([]string{ownerEmail.String}, "TBKK Project Control Notification : File Approved", ownerEmailBody, "text/html; charset=utf-8")
				}
			}
		}
	}
	return c.Status(200).JSON(1)

}
