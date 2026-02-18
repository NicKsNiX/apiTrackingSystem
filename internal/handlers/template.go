package handlers

import (
	"database/sql"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysTemplate represents a row in mst_template
type SysTemplate struct {
	ID                int64            `db:"mt_id" json:"mt_id"`
	Name              string           `db:"mt_name" json:"mt_name"`
	Status            string           `db:"mt_status" json:"mt_status"`
	CreatedAt         *time.Time       `db:"mt_created_at" json:"mt_created_at"`
	CreatedBy         utils.NullString `db:"mt_created_by" json:"mt_created_by"`
	UpdatedAt         *time.Time       `db:"mt_updated_at" json:"mt_updated_at"`
	UpdatedBy         utils.NullString `db:"mt_updated_by" json:"mt_updated_by"`
	UpdateByFirstName utils.NullString `db:"mt_updated_by_firstname" json:"mt_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mt_updated_by_lastname" json:"mt_updated_by_lastname"`
}

// ListTemplates returns all templates
func ListTemplates(c *fiber.Ctx, db *sqlx.DB) error {
	var list []SysTemplate
	query := `SELECT mt_id AS mt_id,
					 mt_name AS mt_name,
					 mt_status AS mt_status,
					 mt_created_at AS mt_created_at,
					 mt_created_by AS mt_created_by,
					 mt_updated_at AS mt_updated_at,
					 mt_updated_by AS mt_updated_by ,
					 su.su_firstname AS mt_updated_by_firstname,
					 su.su_lastname AS mt_updated_by_lastname
			  FROM mst_template 
			  LEFT JOIN sys_user su ON mt_updated_by = su_emp_code
			  ORDER BY mt_id ASC`
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

// GetTemplate returns a template by id
func GetTemplate(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var t SysTemplate
	query := `SELECT mt_id AS mt_id, mt_name AS mt_name, mt_status AS mt_status, mt_created_at AS mt_created_at, mt_created_by AS mt_created_by, mt_updated_at AS mt_updated_at, mt_updated_by AS mt_updated_by FROM mst_template WHERE mt_id = ? LIMIT 1`
	if err := db.Get(&t, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "template not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(t)
}

func InsertTemplate(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Name      string `json:"mt_name"`
		CreatedBy string `json:"mt_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_name is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_created_by is required"})
	}

	// Check duplicate mt_name
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_template WHERE mt_name = ?`, body.Name); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO mst_template (mt_name, mt_status, mt_created_at, mt_created_by, mt_updated_at, mt_updated_by) VALUES (?, ?, ?, ?, ?, ?)`, body.Name, "active", now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}

	return c.Status(201).JSON(1)
}

func UpdateTemplate(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mt_id"`
		Name      string `json:"mt_name"`
		UpdatedBy string `json:"mt_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mt_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_name is required"})
	}

	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_updated_by is required"})
	}

	// Check duplicate mt_name (exclude current record)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_template WHERE mt_name = ? AND mt_id <> ?`, body.Name, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_template SET mt_name = ?, mt_updated_at = ?, mt_updated_by = ? WHERE mt_id = ?`, body.Name, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update template", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "template not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdateTemplateStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mt_id"`
		Status    string `json:"mt_status"`
		UpdatedBy string `json:"mt_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mt_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mt_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_template SET mt_status = ?, mt_updated_at = ?, mt_updated_by = ? WHERE mt_id = ?`,
		body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update template status", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()

	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "template not found"})
	}
	return c.Status(200).JSON(1)
}
