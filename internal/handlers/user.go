package handlers

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysUser represents the user row we return to the frontend
type SysUser struct {
	ID                 int64         `db:"su_id" json:"id"`
	Aname              string        `db:"sd_dept_aname" json:"department_name"`
	Username           string        `db:"su_username" json:"username"`
	EmpCode            string        `db:"su_emp_code" json:"employeeID"`
	FirstName          string        `db:"su_firstname" json:"firstName"`
	LastName           string        `db:"su_lastname" json:"lastName"`
	Email              string        `db:"su_email" json:"email"`
	Status             string        `db:"su_status" json:"status"`
	SpgID              int           `db:"spg_id" json:"spg_id"`
	SpgName            string        `db:"spg_name" json:"spg_name"`
	SdID               sql.NullInt64 `db:"sd_id" json:"sd_id"`
	Department         string        `db:"sd_name" json:"department"`
	UpdatedByFirstName string        `db:"su_firstname_updated_by" json:"updated_by_firstname"`
	UpdatedByLastName  string        `db:"su_lastname_updated_by" json:"updated_by_lastname"`
	CreatedAt          *time.Time    `db:"su_created_at" json:"created_at"`
	UpdatedAt          *time.Time    `db:"su_updated_at" json:"updated_at"`
	UpdatedBy          string        `db:"su_updated_by" json:"updated_by"`
}

// GetUser fetches a single user by username and returns JSON to the frontend
func GetUser(c *fiber.Ctx, db *sqlx.DB) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username required"})
	}

	var u SysUser
	err := db.Get(&u, `SELECT su.su_id AS su_id, su.su_username AS su_username, su.su_emp_code AS su_emp_code, su.su_firstname AS su_firstname, su.su_lastname AS su_lastname, su.su_email AS su_email, su.su_status AS su_status, su.spg_id AS spg_id, su.sd_id AS sd_id, d.sd_name AS sd_name, su.su_created_at AS su_created_at, su.su_updated_at AS su_updated_at FROM sys_user su LEFT JOIN sys_department d ON su.sd_id = d.sd_id WHERE su.su_username = ? LIMIT 1`, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	return c.Status(200).JSON(u)
}

// ListUsers returns a paginated list of users (simple version)
func ListUsers(c *fiber.Ctx, db *sqlx.DB) error {

	var users []SysUser
	err := db.Select(&users, `SELECT
					su.su_id AS su_id,
					su.su_emp_code AS su_emp_code,
					su.su_firstname AS su_firstname,
					su.su_lastname AS su_lastname,
					d.sd_name AS sd_name,
					pg.spg_id AS spg_id,
					pg.spg_name AS spg_name,
					su.su_status AS su_status,
					su.su_updated_at AS su_updated_at,
					su.su_updated_by AS su_updated_by,
					updater.su_firstname AS su_firstname_updated_by,
					updater.su_lastname AS su_lastname_updated_by
					FROM
					sys_user su
					LEFT JOIN sys_permission_group pg ON su.spg_id = pg.spg_id
					LEFT JOIN sys_department d ON su.sd_id = d.sd_id
					LEFT JOIN sys_user updater ON updater.su_emp_code = su.su_updated_by
					ORDER BY
					su.su_id ASC
				`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(users)
}

// UpdateUserStatus updates a user's status by su_id
// Expects JSON: { "su_id": 123, "su_status": "active|inactive", "updated_by": "optional" }
func UpdateUserStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"su_id"`
		Status    string `json:"su_status"`
		UpdatedBy string `json:"updated_by"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}

	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "su_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "su_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		body.UpdatedBy = "system"
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE sys_user SET su_status = ?, su_updated_at = ?, su_updated_by = ? WHERE su_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(0)
	}

	return c.Status(200).JSON(1)
}

func UpdatePermissionGroupUser(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"su_id"`
		SpgID     int64  `json:"spg_id"`
		UpdatedBy string `json:"updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "su_id is required"})
	}
	if body.SpgID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "spg_id is required"})
	}
	if body.UpdatedBy == "" {
		body.UpdatedBy = "system"
	}
	now := time.Now()

	res, err := db.Exec(`UPDATE sys_user SET spg_id = ?, su_updated_at = ?, su_updated_by = ? WHERE su_id = ?`, body.SpgID, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user permission group", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(0)
	}
	return c.Status(200).JSON(1)
}

func GetUserByDepartment(c *fiber.Ctx, db *sqlx.DB) error {
	sdID := c.Query("sd_id")
	if sdID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sd_id required"})
	}
	var users []SysUser
	query := `SELECT
				su.su_id AS su_id,
				sd.sd_dept_aname AS sd_dept_aname,
				su.su_emp_code AS su_emp_code,
				su.su_firstname AS su_firstname,
				su.su_lastname AS su_lastname,
				su.su_email AS su_email,
				su.su_status AS su_status
			  FROM sys_user su
			  LEFT JOIN sys_department sd ON su.sd_id = sd.sd_id
			  WHERE su.sd_id = ?`
	if err := db.Select(&users, query, sdID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(users)
}
						
