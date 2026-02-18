package handlers

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"apiTrackingSystem/internal/utils"
)

// SysPermissionGroup represents a row in sys_permission_group
type SysPermissionGroup struct {
	ID                 int64            `db:"spg_id" json:"spg_id"`
	Name               string           `db:"spg_name" json:"spg_name"`
	Status             string           `db:"spg_status" json:"spg_status"`
	CreatedAt          *time.Time       `db:"spg_created_at" json:"spg_created_at"`
	CreatedBy          string           `db:"spg_created_by" json:"spg_created_by"`
	UpdatedAt          *time.Time       `db:"spg_updated_at" json:"spg_updated_at"`
	UpdatedBy          utils.NullString `db:"spg_updated_by" json:"spg_updated_by"`
	CreatedByFirstName utils.NullString `db:"spg_created_by_firstname" json:"spg_created_by_firstname"`
	CreatedByLastName  utils.NullString `db:"spg_created_by_lastname" json:"spg_created_by_lastname"`
	UpdatedByFirstName utils.NullString `db:"spg_updated_by_firstname" json:"spg_updated_by_firstname"`
	UpdatedByLastName  utils.NullString `db:"spg_updated_by_lastname" json:"spg_updated_by_lastname"`
}

// SysPermissionDetail represents a row in sys_permission_detail
type SysPermissionDetail struct {
	ID                int64            `db:"spd_id" json:"spd_id"`
	SubmenuID         sql.NullInt64    `db:"ss_id" json:"ss_id"`
	SsLink            utils.NullString `db:"ss_link" json:"ss_link"`
	SsName            utils.NullString `db:"ss_name" json:"ss_name"`
	SmName            utils.NullString `db:"sm_name" json:"sm_name"`
	SpgID             sql.NullInt64    `db:"spg_id" json:"spg_id"`
	SmIcon            utils.NullString `db:"sm_icon" json:"sm_icon"`
	Order             sql.NullInt64    `db:"spd_order" json:"spd_order"`
	Status            string           `db:"spd_status" json:"spd_status"`
	CreatedAt         *time.Time       `db:"spd_created_at" json:"spd_created_at"`
	CreatedBy         utils.NullString `db:"spd_created_by" json:"spd_created_by"`
	UpdatedAt         *time.Time       `db:"spd_updated_at" json:"spd_updated_at"`
	UpdatedBy         utils.NullString `db:"spd_updated_by" json:"spd_updated_by"`
	CreateByFirstName utils.NullString `db:"spd_created_by_firstname" json:"spd_created_by_firstname"`
	CreateByLastName  utils.NullString `db:"spd_created_by_lastname" json:"spd_created_by_lastname"`
}

type MenuPermissionGroup struct {
	SpgID             int64            `db:"spg_id" json:"spg_id"`
	SpgName           utils.NullString `db:"spg_name" json:"spg_name"`
	SsID              int64            `db:"ss_id" json:"ss_id"`
	SsName            utils.NullString `db:"ss_name" json:"ss_name"`
	SsLink            utils.NullString `db:"ss_link" json:"ss_link"`
	SmID              int64            `db:"sm_id" json:"sm_id"`
	SmName            utils.NullString `db:"sm_name" json:"sm_name"`
	SpdStatus         string           `db:"spd_status" json:"spd_status"`
	CreateAt          *time.Time       `db:"spd_created_at" json:"spd_created_at"`
	CreatedBy         utils.NullString `db:"spd_created_by" json:"spd_created_by"`
	UpdatedBy         utils.NullString `db:"spd_updated_by" json:"spd_updated_by"`
	UpdatedAt         *time.Time       `db:"spd_updated_at" json:"spd_updated_at"`
	CreateByFirstName utils.NullString `db:"spd_created_by_firstname" json:"spd_created_by_firstname"`
	CreateByLastName  utils.NullString `db:"spd_created_by_lastname" json:"spd_created_by_lastname"`
}

// ListPermissionGroups returns all permission groups
func ListPermissionGroups(c *fiber.Ctx, db *sqlx.DB) error {
	var groups []SysPermissionGroup
	query := `SELECT spg.spg_id AS spg_id,
	 spg.spg_name AS spg_name,
	 spg.spg_status AS spg_status,
	 spg.spg_created_at AS spg_created_at,
	 spg.spg_created_by AS spg_created_by,
	 spg.spg_updated_at AS spg_updated_at,
	 spg.spg_updated_by AS spg_updated_by,
	 su.su_firstname AS spg_created_by_firstname,
	 su.su_lastname AS spg_created_by_lastname,
	 su.su_firstname AS spg_updated_by_firstname,
	 su.su_lastname AS spg_updated_by_lastname

	 FROM sys_permission_group spg
	 LEFT JOIN sys_user su ON spg.spg_created_by = su.su_emp_code
	 ORDER BY spg.spg_id ASC`
	if err := db.Select(&groups, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(groups)
}

func ListPermissionDetail(c *fiber.Ctx, db *sqlx.DB) error {
	spg_id := c.Query("spg_id")
	// fmt.Println("spg_id =", spg_id)
	if spg_id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_id required"})
	}
	var groups []SysPermissionDetail
	query := `SELECT 
				sys_permission_detail.spd_id AS spd_id,
				sys_menu.sm_icon AS sm_icon,
				sys_menu.sm_name AS sm_name,
				sys_submenu.ss_name AS ss_name,
				sys_submenu.ss_link AS ss_link,
				sys_permission_detail.spd_status AS spd_status,
				sys_permission_detail.spd_updated_at AS spd_updated_at,
				sys_permission_detail.spd_updated_by AS spd_updated_by,
				su.su_firstname AS spd_created_by_firstname,
				su.su_lastname AS spd_created_by_lastname						  			  
	 		  FROM sys_permission_detail 
				LEFT JOIN sys_permission_group ON sys_permission_detail.spg_id = sys_permission_group.spg_id
				LEFT JOIN sys_submenu ON sys_permission_detail.ss_id = sys_submenu.ss_id
				LEFT JOIN sys_menu ON sys_submenu.sm_id = sys_menu.sm_id
				LEFT JOIN sys_user su ON sys_permission_detail.spd_updated_by = su.su_emp_code
			  WHERE sys_permission_group.spg_id = ?
			  ORDER BY sys_permission_group.spg_id ASC`
	if err := db.Select(&groups, query, spg_id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(groups)
}

func GetSelectPermissionGroups(c *fiber.Ctx, db *sqlx.DB) error {
	var groups []SysPermissionGroup
	query := `SELECT spg_id AS spg_id, spg_name AS spg_name FROM sys_permission_group WHERE spg_status = 'active' ORDER BY spg_id ASC`
	if err := db.Select(&groups, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(groups)
}

func GetPermissionGroup(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var g SysPermissionGroup
	query := `SELECT spg_id AS spg_id, spg_name AS spg_name, spg_status AS spg_status, spg_created_at AS spg_created_at, spg_created_by AS spg_created_by, spg_updated_at AS spg_updated_at, spg_updated_by AS spg_updated_by FROM sys_permission_group WHERE spg_id = ? LIMIT 1`
	if err := db.Get(&g, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "permission group not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(g)
}

func InsertPermissionGroup(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Name      string `json:"spg_name"`
		CreatedBy string `json:"spg_created_by"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}

	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_name is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_created_by is required"})
	}

	// Check if spg_name already exists
	var count int
	err := db.Get(&count, `SELECT COUNT(*) FROM sys_permission_group WHERE spg_name = ?`, body.Name)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	_, err = db.Exec(`INSERT INTO sys_permission_group (spg_name, spg_status, spg_created_at, spg_created_by, spg_updated_at, spg_updated_by) VALUES (?, 'active', ?, ?,?,?)`, body.Name, now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(5)
	}

	return c.Status(201).JSON(1)
}

func InsertPermissionDetail(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		SubmenuID int64  `json:"ss_id"`
		SpgID     int64  `json:"spg_id"`
		Order     *int   `json:"spd_order"`
		CreatedBy string `json:"spd_created_by"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}

	// Check duplicate: same submenu (ss_id) for the same permission group (spg_id)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sys_permission_detail WHERE ss_id = ? and spg_id = ?`, body.SubmenuID, body.SpgID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		// duplicate
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO sys_permission_detail (ss_id, spg_id, spd_order, spd_status, spd_created_at, spd_created_by, spd_updated_at, spd_updated_by) VALUES (?, ?, ?, 'active', ?, ?,?,?)`, body.SubmenuID, body.SpgID, body.Order, now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}

	// Success: return numeric code per project pattern
	return c.Status(201).JSON(1)
}

func UpdatePermissionGroup(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"spg_id"`
		Name      string `json:"spg_name"`
		UpdatedBy string `json:"spg_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "spg_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_name is required"})
	}

	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_updated_by is required"})
	}

	// Check for duplicate spg_name (exclude current record)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sys_permission_group WHERE spg_name = ? AND spg_id <> ?`, body.Name, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	_, err := db.Exec(`UPDATE sys_permission_group SET spg_name = ?, spg_updated_at = ?, spg_updated_by = ? WHERE spg_id = ?`, body.Name, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}

	return c.Status(200).JSON(1)
}

func UpdatePermissionGroupStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"spg_id"`
		Status    string `json:"spg_status"`
		UpdatedBy string `json:"spg_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "spg_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spg_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE sys_permission_group SET spg_status = ?, spg_updated_at = ?, spg_updated_by = ? WHERE spg_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update permission group", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "permission group not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdatePermissionDetailStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"spd_id"`
		Status    string `json:"spd_status"`
		UpdatedBy string `json:"spd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "spd_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "spd_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "spd_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE sys_permission_detail SET spd_status = ?, spd_updated_at = ?, spd_updated_by = ? WHERE spd_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update permission detail", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "permission detail not found"})
	}
	return c.Status(200).JSON(1)
}
