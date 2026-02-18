package handlers

import (
	"database/sql"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysMasterPlan represents a row in mst_master_plan
type SysMasterPlan struct {
	ID                int64            `db:"mmp_id" json:"mmp_id"`
	Name              string           `db:"mmp_name" json:"mmp_name"`
	Aname             utils.NullString `db:"aname" json:"aname"`
	Type              string           `db:"mmp_type" json:"mmp_type"`
	Status            string           `db:"mmp_status" json:"mmp_status"`
	CreatedAt         *time.Time       `db:"mmp_created_at" json:"mmp_created_at"`
	CreatedBy         utils.NullString `db:"mmp_created_by" json:"mmp_created_by"`
	UpdatedAt         *time.Time       `db:"mmp_updated_at" json:"mmp_updated_at"`
	UpdatedBy         utils.NullString `db:"mmp_updated_by" json:"mmp_updated_by"`
	UpdateByFirstName utils.NullString `db:"mmp_updated_by_firstname" json:"mmp_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mmp_updated_by_lastname" json:"mmp_updated_by_lastname"`

	MtID int64 `db:"mt_id" json:"mt_id"`
	IpID int64 `db:"ip_id" json:"ip_id"`
}

// ListMasterPlans returns all master plans
func ListMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var list []SysMasterPlan
	query := `  SELECT mmp.mmp_id AS mmp_id,
					   mmp.mmp_name AS mmp_name,
					   mmp.mmp_type AS mmp_type,
					   GROUP_CONCAT(ma.ma_name SEPARATOR ', ') AS aname,
					   mmp.mmp_status AS mmp_status,
					   mmp.mmp_updated_at AS mmp_updated_at,
					   mmp.mmp_updated_by AS mmp_updated_by,
					   su.su_firstname AS mmp_updated_by_firstname,
					   su.su_lastname AS mmp_updated_by_lastname
				FROM mst_master_plan mmp
				LEFT JOIN sys_user su ON mmp_updated_by = su_emp_code
				RIGHT JOIN mst_master_plan_detail mmpd ON mmpd.mmp_id = mmp.mmp_id
				LEFT JOIN mst_apqp ma ON mmpd.ma_id = ma.ma_id
				WHERE mmpd.mmpd_status = 'active'
				GROUP BY mmp.mmp_id
				ORDER BY mmp.mmp_id ASC`
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func GetMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var list []SysMasterPlan
	query := `  SELECT mmp.mmp_id AS mmp_id,
					   mmp.mmp_name AS mmp_name

				FROM mst_master_plan mmp

				WHERE mmp.mmp_status = 'active'
				GROUP BY mmp.mmp_id
				ORDER BY mmp.mmp_id ASC`
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func InsertMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Name      string  `json:"mmp_name"`
		ApqpIDs   []int64 `json:"ma_id"`
		CreatedBy string  `json:"mmp_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	// duplicate check: ensure mmp_name is unique
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_master_plan WHERE mmp_name = ?`, body.Name); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`INSERT INTO mst_master_plan (mmp_name, mmp_type, mmp_status, mmp_created_at, mmp_created_by, mmp_updated_at, mmp_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?)`, body.Name, "text", "active", now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	mmpID, err := res.LastInsertId()
	if err != nil {
		return c.Status(500).JSON(5)
	}

	inserted := 0
	if len(body.ApqpIDs) == 0 {
		if _, err := tx.Exec(`INSERT INTO mst_master_plan_detail (mmp_id, ma_id, sd_id, mmpd_created_at, mmpd_created_by, mmpd_updated_at, mmpd_updated_by, mmpd_status) VALUES (?, ?, ?, ?, ?, ?, ?, 'active')`, mmpID, nil, nil, now, body.CreatedBy, now, body.CreatedBy); err != nil {
			return c.Status(500).JSON(5)
		}
		inserted = 1
	} else {
		for _, ma := range body.ApqpIDs {
			if _, err := tx.Exec(`INSERT INTO mst_master_plan_detail (mmp_id, ma_id, sd_id, mmpd_created_at, mmpd_created_by, mmpd_updated_at, mmpd_updated_by, mmpd_status) VALUES (?, ?, ?, ?, ?, ?, ?, 'active')`, mmpID, ma, nil, now, body.CreatedBy, now, body.CreatedBy); err != nil {
				return c.Status(500).JSON(5)
			}
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdateMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64   `json:"mmp_id"`
		Name      string  `json:"mmp_name"`
		UpdatedBy string  `json:"mmp_updated_by"`
		ApqpIDs   []int64 `json:"ma_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	// duplicate check: ensure no other record has the same name
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_master_plan WHERE mmp_name = ? AND mmp_id <> ?`, body.Name, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`UPDATE mst_master_plan SET mmp_name = ?, mmp_type = ?, mmp_updated_at = ?, mmp_updated_by = ? WHERE mmp_id = ?`, body.Name, "text", now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "master plan not found"})
	}

	// Super-update details: fetch existing detail rows ordered by mmpd_id
	var existing []struct {
		MmpdID int64  `db:"mmpd_id"`
		MaID   int64  `db:"ma_id"`
		Status string `db:"mmpd_status"`
	}
	// fetch all existing detail rows for this mmp_id (we need full list to disable leftovers)
	if err := tx.Select(&existing, `SELECT mmpd_id, ma_id, mmpd_status FROM mst_master_plan_detail WHERE mmp_id = ? ORDER BY mmpd_id ASC`, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}

	// If incoming is empty -> disable all existing details (set ma_id = NULL)
	if len(body.ApqpIDs) == 0 {
		if len(existing) > 0 {
			for _, ex := range existing {
				if _, err := tx.Exec(`UPDATE mst_master_plan_detail SET mmpd_status = 'inactive', mmpd_updated_at = ?, mmpd_updated_by = ? WHERE mmpd_id = ?`, now, body.UpdatedBy, ex.MmpdID); err != nil {
					return c.Status(500).JSON(5)
				}
			}
		}
	} else {
		// update existing by position, insert extras, delete leftovers
		minLen := len(existing)
		if len(body.ApqpIDs) < minLen {
			minLen = len(body.ApqpIDs)
		}
		for i := 0; i < minLen; i++ {
			// if value changed or the row is currently inactive, update and ensure it's active
			if existing[i].MaID != body.ApqpIDs[i] || existing[i].Status != "active" {
				if _, err := tx.Exec(`UPDATE mst_master_plan_detail SET ma_id = ?, mmpd_status = 'active', mmpd_updated_at = ?, mmpd_updated_by = ? WHERE mmpd_id = ?`, body.ApqpIDs[i], now, body.UpdatedBy, existing[i].MmpdID); err != nil {
					return c.Status(500).JSON(5)
				}
			}
		}
		if len(body.ApqpIDs) > len(existing) {
			for i := len(existing); i < len(body.ApqpIDs); i++ {
				if _, err := tx.Exec(`INSERT INTO mst_master_plan_detail (mmp_id, ma_id, sd_id, mmpd_created_at, mmpd_created_by, mmpd_updated_at, mmpd_updated_by, mmpd_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, body.ID, body.ApqpIDs[i], nil, now, body.UpdatedBy, now, body.UpdatedBy, "active"); err != nil {
					return c.Status(500).JSON(5)
				}
			}
		}
		if len(existing) > len(body.ApqpIDs) {
			for i := len(body.ApqpIDs); i < len(existing); i++ {
				if _, err := tx.Exec(`UPDATE mst_master_plan_detail SET mmpd_status = 'inactive', mmpd_updated_at = ?, mmpd_updated_by = ? WHERE mmpd_id = ?`, now, body.UpdatedBy, existing[i].MmpdID); err != nil {
					return c.Status(500).JSON(5)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(200).JSON(1)
}

func UpdateMasterPlanStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mmp_id"`
		Status    string `json:"mmp_status"`
		UpdatedBy string `json:"mmp_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mmp_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mmp_status must be 'active' or 'inactive'"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_master_plan SET mmp_status = ? , mmp_updated_by = ?, mmp_updated_at = ? WHERE mmp_id = ?`, body.Status, body.UpdatedBy, now, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update status", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "master plan not found"})
	}
	return c.Status(200).JSON(1)
}

func SelectAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	var items []struct {
		ID    int64  `db:"ma_id" json:"ma_id"`
		MppID int64  `db:"mpp_id" json:"mpp_id"`
		Name  string `db:"ma_name" json:"ma_name"`
	}
	if err := db.Select(&items, `SELECT ma_id, mpp_id, ma_name FROM mst_apqp WHERE ma_status = 'active' ORDER BY mpp_id ASC`); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(items)
}

func GetMasterPlanStep2(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("mt_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var steps []struct {
		MtID      int64            `db:"mt_id" json:"mt_id"`
		IpID      int64            `db:"ip_id" json:"ip_id"`
		IpmpID    int64            `db:"ipmp_id" json:"ipmp_id"`
		Name      utils.NullString `db:"ipmp_name" json:"ipmp_name"`
		StartDate *Date            `db:"ipmp_start_date" json:"ipmp_start_date"`
		EndDate   *Date            `db:"ipmp_end_date" json:"ipmp_end_date"`
		Status    utils.NullString `db:"ipmp_status" json:"ipmp_status"`
	}
	query := `SELECT
				ip.mt_id,
				ipmp.ipmp_id,
				ipmp.ip_id,
				ipmp.ipmp_name,
				ipmp.ipmp_start_date,
				ipmp.ipmp_end_date,
				ipmp.ipmp_status
				FROM info_project_master_plan ipmp
				LEFT JOIN info_project ip ON ip.ip_id = ipmp.ip_id 
				WHERE ip.mt_id = ?`
	if err := db.Select(&steps, query, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	if len(steps) == 0 {
		var alt []struct {
			Name utils.NullString `db:"ipmp_name" json:"ipmp_name"`
		}
		altQuery := `SELECT mmp.mmp_name AS ipmp_name 
					FROM mst_template_detail mtd 
					LEFT JOIN mst_template mt ON mt.mt_id = mtd.mt_id 
					LEFT JOIN mst_master_plan mmp ON mmp.mmp_id = mtd.mmp_id 
					WHERE mtd.mt_id = ? AND mmp.mmp_status = 'active'`
		if err := db.Select(&alt, altQuery, id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "fallback query error", "detail": err.Error()})
		}
		return c.Status(200).JSON(alt)
	}

	return c.Status(200).JSON(steps)
}

func GetMasterPlanStep2T(c *fiber.Ctx, db *sqlx.DB) error {
	ipID := c.Query("ip_id")
	if ipID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id required"})
	}

	var rows []struct {
		MmpID   int64            `db:"mmp_id" json:"mmp_id"`
		MmpName utils.NullString `db:"mmp_name" json:"mmp_name"`
		Start   *Date            `db:"ipmp_start_date" json:"ipmp_start_date"`
		End     *Date            `db:"ipmp_end_date" json:"ipmp_end_date"`
		IStatus utils.NullString `db:"ipmp_status" json:"ipmp_status"`
		Status  int              `db:"status" json:"status"`
	}

	query := `SELECT DISTINCT
					m.mmp_id,
					m.mmp_name,
					i.ipmp_start_date,
					i.ipmp_end_date,
					i.ipmp_status,
				CASE
				   WHEN mtd.mtpd_status = 'inactive' THEN 0
					WHEN i.ipmp_id IS NOT NULL THEN 1
					WHEN i.ipmp_id IS NULL
						AND NOT EXISTS (
							SELECT 1
							FROM info_project_master_plan x
							WHERE x.ip_id = ?
						)
						AND mtd.mmp_id IS NOT NULL
					THEN 1
					ELSE 0
				END AS status
				FROM mst_master_plan m
				LEFT JOIN info_project_master_plan i
					ON m.mmp_name = i.ipmp_name
					AND i.ip_id = ?
				LEFT JOIN info_project ip 
					ON ip.ip_id = i.ip_id
				LEFT JOIN mst_template_detail mtd 
					ON m.mmp_id = mtd.mmp_id
					AND mtd.mt_id = (SELECT mt_id FROM info_project WHERE ip_id = ?)
					AND i.ipmp_id IS NULL
				WHERE m.mmp_status = 'active'
				ORDER BY m.mmp_id;`

	if err := db.Select(&rows, query, ipID, ipID, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(rows)
}

func GetMasterPlanStep3(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("ip_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var rows []struct {
		IpID   int64            `db:"ip_id" json:"ip_id"`
		IpmpID int64            `db:"ipmp_id" json:"ipmp_id"`
		Name   utils.NullString `db:"ipmp_name" json:"ipmp_name"`
		MmpID  sql.NullInt64    `db:"mmp_id" json:"mmp_id"`
		MaID   sql.NullInt64    `db:"ma_id" json:"ma_id"`
	}

	query := `SELECT
				ipmp.ip_id,
				ipmp.ipmp_id,
				ipmp.ipmp_name,
				mmp.mmp_id,
				mmpd.ma_id
				FROM info_project_master_plan ipmp
				LEFT JOIN mst_master_plan mmp ON mmp.mmp_name = ipmp.ipmp_name AND mmp.mmp_status = 'active'
				LEFT JOIN mst_master_plan_detail mmpd ON mmpd.mmp_id = mmp.mmp_id AND mmpd.mmpd_status = 'active'
				WHERE ipmp.ip_id = ?
				ORDER BY mmp.mmp_id, mmpd.mmpd_id`

	if err := db.Select(&rows, query, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	var out []struct {
		IpmpID int64  `json:"ipmp_id"`
		MmpID  *int64 `json:"mmp_id"`
		MaID   *int64 `json:"ma_id"`
	}

	for _, r := range rows {
		var mmpPtr *int64
		if r.MmpID.Valid {
			v := int64(r.MmpID.Int64)
			mmpPtr = &v
		}
		var maPtr *int64
		if r.MaID.Valid {
			v := int64(r.MaID.Int64)
			maPtr = &v
		}
		out = append(out, struct {
			IpmpID int64  `json:"ipmp_id"`
			MmpID  *int64 `json:"mmp_id"`
			MaID   *int64 `json:"ma_id"`
		}{

			IpmpID: r.IpmpID,
			MmpID:  mmpPtr,
			MaID:   maPtr,
		})
	}

	return c.Status(200).JSON(out)
}
