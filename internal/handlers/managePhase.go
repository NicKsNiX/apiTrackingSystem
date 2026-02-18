package handlers

import (
	"database/sql"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysProjectPhase represents a row in mst_project_phase
type SysProjectPhase struct {
	ID                int64            `db:"mpp_id" json:"mpp_id"`
	Name              utils.NullString `db:"mpp_name" json:"mpp_name"`
	Order             int64            `db:"mpp_order" json:"mpp_order"`
	Status            utils.NullString `db:"mpp_status" json:"mpp_status"`
	CreatedAt         *time.Time       `db:"mpp_created_at" json:"mpp_created_at"`
	CreatedBy         utils.NullString `db:"mpp_created_by" json:"mpp_created_by"`
	UpdatedAt         *time.Time       `db:"mpp_updated_at" json:"mpp_updated_at"`
	UpdatedBy         utils.NullString `db:"mpp_updated_by" json:"mpp_updated_by"`
	UpdateByFirstName utils.NullString `db:"mpp_updated_by_firstname" json:"mpp_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mpp_updated_by_lastname" json:"mpp_updated_by_lastname"`
}

func ListProjectPhases(c *fiber.Ctx, db *sqlx.DB) error {
	var list []SysProjectPhase
	query := `SELECT
                    mpp_id AS mpp_id,
                    mpp_name AS mpp_name,
                    mpp_order AS mpp_order,
                    mpp_status AS mpp_status,
                    mpp_updated_at AS mpp_updated_at,
                    mpp_updated_by AS mpp_updated_by,
					su.su_firstname AS mpp_updated_by_firstname,
					su.su_lastname AS mpp_updated_by_lastname
                FROM mst_project_phase
				LEFT JOIN sys_user su ON mpp_updated_by = su_emp_code
                ORDER BY mpp_order ASC`
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func GetProjectPhase(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var p SysProjectPhase
	query := `SELECT mpp_id AS mpp_id, mpp_name AS mpp_name, mpp_order AS mpp_order, mpp_created_at AS mpp_created_at, mpp_created_by AS mpp_created_by, mpp_updated_at AS mpp_updated_at, mpp_updated_by AS mpp_updated_by FROM mst_project_phase WHERE mpp_id = ? LIMIT 1`
	if err := db.Get(&p, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "project phase not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(p)
}

func InsertProjectPhase(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Name      string `json:"mpp_name"`
		CreatedBy string `json:"mpp_created_by"`
		Order     int64  `json:"mpp_order"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_name is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_created_by is required"})
	}

	// duplicate check on name
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_project_phase WHERE mpp_name = ?`, body.Name); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	// compute order if not provided (zero => compute next)
	order := body.Order
	if order == 0 {
		var next int64
		if err := db.Get(&next, `SELECT IFNULL(MAX(mpp_order),0) + 1 FROM mst_project_phase`); err != nil {
			return c.Status(500).JSON(5)
		}
		order = next
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO mst_project_phase (mpp_name, mpp_order,mpp_status, mpp_created_at, mpp_created_by, mpp_updated_at, mpp_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?)`, body.Name, order, "active", now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdateProjectPhase(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mpp_id"`
		Name      string `json:"mpp_name"`
		Order     int64  `json:"mpp_order"`
		UpdatedBy string `json:"mpp_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_name is required"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_updated_by is required"})
	}

	// duplicate check excluding current id
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_project_phase WHERE mpp_name = ? AND mpp_id <> ?`, body.Name, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	// decide order: if not provided (zero) keep existing
	order := body.Order
	if order == 0 {
		var cur sql.NullInt64
		if err := db.Get(&cur, `SELECT mpp_order FROM mst_project_phase WHERE mpp_id = ?`, body.ID); err != nil {
			return c.Status(500).JSON(5)
		}
		order = cur.Int64
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_project_phase SET mpp_name = ?, mpp_order = ?, mpp_updated_at = ?, mpp_updated_by = ? WHERE mpp_id = ?`, body.Name, order, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "project phase not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdateProjectPhaseStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mpp_id"`
		Status    string `json:"mpp_status"`
		UpdatedBy string `json:"mpp_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpp_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_project_phase SET mpp_status = ?, mpp_updated_at = ?, mpp_updated_by = ? WHERE mpp_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "project phase not found"})
	}
	return c.Status(200).JSON(1)
}
