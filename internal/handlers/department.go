package handlers

import (
	"apiTrackingSystem/internal/utils"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysDepartment represents a row in sys_department
type SysDepartment struct {
	ID                 int64            `db:"sd_id" json:"sd_id"`
	Aname              utils.NullString `db:"sd_dept_aname" json:"sd_dept_aname"`
	Name               utils.NullString `db:"sd_name" json:"sd_name"`
	Email              utils.NullString `db:"sd_email" json:"sd_email"`
	Status             utils.NullString `db:"sd_status" json:"sd_status"`
	Code               utils.NullString `db:"sd_code" json:"sd_code"`
	CreatedAt          *time.Time       `db:"sd_created_at" json:"sd_created_at"`
	CreatedBy          utils.NullString `db:"sd_created_by" json:"sd_created_by"`
	UpdatedAt          *time.Time       `db:"sd_updated_at" json:"sd_updated_at"`
	UpdatedBy          utils.NullString `db:"sd_updated_by" json:"sd_updated_by"`
	CreatedByFirstName utils.NullString `db:"sd_created_by_firstname" json:"sd_created_by_firstname"`
	CreatedByLastName  utils.NullString `db:"sd_created_by_lastname" json:"sd_created_by_lastname"`
	UpdatedByFirstName utils.NullString `db:"sd_updated_by_firstname" json:"sd_updated_by_firstname"`
	UpdatedByLastName  utils.NullString `db:"sd_updated_by_lastname" json:"sd_updated_by_lastname"`
}

// ListDepartments returns all departments
func ListDepartments(c *fiber.Ctx, db *sqlx.DB) error {
	var departments []SysDepartment
	query := `SELECT 
				 sd.sd_id AS sd_id,
				 sd.sd_dept_aname AS sd_dept_aname,
				 sd_name AS sd_name,
				 sd_email AS sd_email,
				 sd_status AS sd_status,
				 sd_code AS sd_code,
				 sd_created_at AS sd_created_at,
				 sd_created_by AS sd_created_by,
				 sd_updated_at AS sd_updated_at,
				 sd_updated_by AS sd_updated_by,
				 su.su_firstname AS sd_created_by_firstname,
				 su.su_lastname AS sd_created_by_lastname,
				 su.su_firstname AS sd_updated_by_firstname,
				 su.su_lastname AS sd_updated_by_lastname
			  FROM sys_department sd
			  LEFT JOIN sys_user su ON sd_updated_by = su_emp_code
			  ORDER BY sd_id ASC`
	if err := db.Select(&departments,
		query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(departments)
}

// GetDepartment returns a single department by id
func GetDepartment(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("sd_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var d SysDepartment
	query := `SELECT sd_id AS sd_id, sd_name AS sd_name, sd_email AS sd_email, sd_status AS sd_status, sd_code AS sd_code, sd_created_at AS sd_created_at, sd_created_by AS sd_created_by, sd_updated_at AS sd_updated_at, sd_updated_by AS sd_updated_by FROM sys_department WHERE sd_id = ? LIMIT 1`
	if err := db.Get(&d, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "department not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(d)
}

func UpdateDepartment(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"sd_id"`
		Email     string `json:"sd_email"`
		UpdatedBy string `json:"sd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	now := time.Now()
	result, err := db.Exec(`UPDATE sys_department SET sd_email = ?, sd_updated_at = ?, sd_updated_by = ? WHERE sd_id = ?`,
		body.Email,
		now,
		body.UpdatedBy,
		body.ID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update error", "detail": err.Error()})
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "department not found"})
	}
	return c.Status(200).JSON(1)
}
func UpdateDepartmentStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"sd_id"`
		Status    string `json:"sd_status"`
		UpdatedBy string `json:"sd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	now := time.Now()
	result, err := db.Exec(`UPDATE sys_department SET sd_status = ?, sd_updated_at = ?, sd_updated_by = ? WHERE sd_id = ?`,
		body.Status,
		now,
		body.UpdatedBy,
		body.ID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update error", "detail": err.Error()})
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "department not found"})
	}
	return c.Status(200).JSON(1)
}
