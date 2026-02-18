package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"apiTrackingSystem/internal/utils"
)

// SysTemplate represents a row in mst_template
type templateDetail struct {
	ID                int64            `db:"mtpd_id" json:"mtpd_id"`
	MtID              int64            `db:"mt_id" json:"mt_id"`
	MmpID             int64            `db:"mmp_id" json:"mmp_id"`
	MtName            utils.NullString `db:"mt_name" json:"mt_name"`
	MmpName           utils.NullString `db:"mmp_name" json:"mmp_name"`
	Status            string           `db:"mtpd_status" json:"mtpd_status"`
	CreatedAt         *time.Time       `db:"mtpd_created_at" json:"mtpd_created_at"`
	CreatedBy         utils.NullString `db:"mtpd_created_by" json:"mtpd_created_by"`
	UpdatedAt         *time.Time       `db:"mtpd_updated_at" json:"mtpd_updated_at"`
	UpdatedBy         utils.NullString `db:"mtpd_updated_by" json:"mtpd_updated_by"`
	UpdateByFirstName utils.NullString `db:"mtpd_updated_by_firstname" json:"mtpd_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mtpd_updated_by_lastname" json:"mtpd_updated_by_lastname"`
}

// ListTemplateDetails lists template-detail rows for a given template (requires mt_id query param)
func ListTemplateDetails(c *fiber.Ctx, db *sqlx.DB) error {
	mtID := c.Query("mt_id")
	mmpID := c.Query("mmp_id")

	// build base query with explicit aliases to match templateDetail struct
	query := `SELECT
					mtpd.mtpd_id AS mtpd_id,
					mtpd.mt_id AS mt_id,
					mtpd.mmp_id AS mmp_id,
					mt.mt_name AS mt_name,
					mmp.mmp_name AS mmp_name,
					mtpd.mtpd_status AS mtpd_status,
					mtpd.mtpd_created_at AS mtpd_created_at,
					mtpd.mtpd_created_by AS mtpd_created_by,
					mtpd.mtpd_updated_at AS mtpd_updated_at,
					mtpd.mtpd_updated_by AS mtpd_updated_by,
					su.su_firstname AS mtpd_updated_by_firstname,
					su.su_lastname AS mtpd_updated_by_lastname
				FROM
					mst_template_detail mtpd
				LEFT JOIN mst_template mt ON mt.mt_id = mtpd.mt_id
				LEFT JOIN mst_master_plan mmp ON mmp.mmp_id = mtpd.mmp_id
				LEFT JOIN sys_user su ON mtpd.mtpd_updated_by = su.su_emp_code
				WHERE 1=1`

	args := []interface{}{}
	if mtID != "" {
		query += " AND mtpd.mt_id = ?"
		args = append(args, mtID)
	}
	if mmpID != "" {
		query += " AND mtpd.mmp_id = ?"
		args = append(args, mmpID)
	}

	var results []templateDetail
	if err := db.Select(&results, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(results)
}

func SelectTemplate(c *fiber.Ctx, db *sqlx.DB) error {
	var t []SysTemplate
	query := `SELECT mt_id AS mt_id,
				 mt_name AS mt_name,
				 mt_status AS mt_status,
				 mt_created_at AS mt_created_at,
				 mt_created_by AS mt_created_by,
				 mt_updated_at AS mt_updated_at,
				 mt_updated_by AS mt_updated_by
			FROM mst_template
			WHERE mt_status = 'active'`

	if err := db.Select(&t, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(t)
}

func SelectMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var m []SysMasterPlan
	query := `SELECT mmp_id AS mmp_id,
						 mmp_name AS mmp_name,
						 mmp_status AS mmp_status,
						 mmp_created_at AS mmp_created_at,
						 mmp_created_by AS mmp_created_by,
						 mmp_updated_at AS mmp_updated_at,
						 mmp_updated_by AS mmp_updated_by
					FROM mst_master_plan
					WHERE mmp_status = 'active'`

	if err := db.Select(&m, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(m)
}

func GetTemplateDetail(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	row := db.QueryRowx(`SELECT * FROM mst_template_detail WHERE mtpd_id = ? LIMIT 1`, id)
	m := make(map[string]interface{})
	if err := row.MapScan(m); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found", "detail": err.Error()})
	}
	return c.Status(200).JSON(m)
}

func InsertTemplateDetail(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		MtID      int64  `json:"mt_id"`
		MmpID     int64  `json:"mmp_id"`
		CreatedBy string `json:"mtpd_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.MtID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mt_id is required"})
	}
	if body.MmpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mmp_id is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_created_by is required"})
	}

	// duplicate check: same mt_id + mmp_id
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_template_detail WHERE mt_id = ? AND mmp_id = ?`, body.MtID, body.MmpID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO mst_template_detail (mt_id, mmp_id,mtpd_status, mtpd_created_at, mtpd_created_by, mtpd_updated_at, mtpd_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?)`, body.MtID, body.MmpID, "active", now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdateTemplateDetail(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mtpd_id"`
		MtID      int64  `json:"mt_id"`
		MmpID     int64  `json:"mmp_id"`
		UpdatedBy string `json:"mtpd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_id is required"})
	}
	if body.MtID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mt_id is required"})
	}
	if body.MmpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mmp_id is required"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_updated_by is required"})
	}

	// duplicate check
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_template_detail WHERE mt_id = ? AND mmp_id = ? AND mtpd_id <> ?`, body.MtID, body.MmpID, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_template_detail SET mmp_id = ?, mtpd_updated_at = ?, mtpd_updated_by = ? WHERE mtpd_id = ?`, body.MmpID, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "template detail not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdateTemplateDetailStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mtpd_id"`
		Status    string `json:"mtpd_status"`
		UpdatedBy string `json:"mtpd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mtpd_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_template_detail SET mtpd_status = ?, mtpd_updated_at = ?, mtpd_updated_by = ? WHERE mtpd_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "template detail not found"})
	}
	return c.Status(200).JSON(1)
}
