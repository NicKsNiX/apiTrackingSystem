package handlers

import (
	"apiTrackingSystem/internal/utils"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysSubMenu represents a row in sys_submenu
type SysSubMenu struct {
	ID                int64            `db:"ss_id" json:"ss_id"`
	MenuID            int64            `db:"sm_id" json:"sm_id"`
	Name              utils.NullString `db:"ss_name" json:"ss_name"`
	Link              utils.NullString `db:"ss_link" json:"ss_link"`
	Order             int64            `db:"ss_order" json:"ss_order"`
	Status            utils.NullString `db:"ss_status" json:"ss_status"`
	UpdatedAt         *time.Time       `db:"ss_updated_at" json:"ss_updated_at"`
	UpdatedBy         utils.NullString `db:"ss_updated_by" json:"ss_updated_by"`
	UpdateByFirstName utils.NullString `db:"ss_updated_by_firstname" json:"ss_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"ss_updated_by_lastname" json:"ss_updated_by_lastname"`
	MenuName          utils.NullString `db:"sm_name" json:"sm_name"`
	MenuIcon          utils.NullString `db:"sm_icon" json:"sm_icon"`
}

// ListSubMenus returns all submenu records
func ListSubMenus(c *fiber.Ctx, db *sqlx.DB) error {
	// require sm_id as query parameter
	smID := c.Query("sm_id")
	if smID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sm_id required"})
	}

	var subs []SysSubMenu
	query := `SELECT 
				ss.ss_id AS ss_id,
				ss.sm_id AS sm_id,
				ss.ss_name AS ss_name,
				ss.ss_link AS ss_link,
				ss.ss_order AS ss_order,
				ss.ss_status AS ss_status,
				ss.ss_updated_at AS ss_updated_at,
				ss.ss_updated_by AS ss_updated_by ,
				su.su_firstname AS ss_updated_by_firstname,
				su.su_lastname AS ss_updated_by_lastname,
				sm.sm_name AS sm_name,
				sm.sm_icon AS sm_icon		
			FROM sys_submenu ss 
			LEFT JOIN sys_user su ON ss.ss_updated_by = su.su_emp_code
			LEFT JOIN sys_menu sm ON ss.sm_id = sm.sm_id
			WHERE ss.sm_id = ?
			ORDER BY ss.ss_order ASC`
	if err := db.Select(&subs, query, smID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(subs)
}

// GetSelectSubMenu returns all active submenus for a given menu (sm_id)
func GetSelectSubMenu(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("sm_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var subs []SysSubMenu
	query := `SELECT 
				ss.ss_id AS ss_id,
				ss.ss_name AS ss_name,
				ss.ss_order AS ss_order,
				ss.ss_status AS ss_status
			FROM sys_submenu ss 
			WHERE ss.sm_id = ? AND ss.ss_status = 'active'
			ORDER BY ss.ss_order ASC
			`

	if err := db.Select(&subs, query, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	if len(subs) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "no active submenus found"})
	}
	return c.Status(200).JSON(subs)
}

func GetSubMenu(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("ss_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var sub SysSubMenu
	query := `SELECT 
					ss.ss_id AS ss_id,
	 				ss.sm_id AS sm_id,
	 				ss.ss_name AS ss_name,
	 				ss.ss_link AS ss_link,
	 				ss.ss_order AS ss_order,
	 				ss.ss_status AS ss_status
				FROM sys_submenu ss 
				WHERE ss.ss_id = ? LIMIT 1`
	if err := db.Get(&sub, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "submenu not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(sub)
}

func InsertMenuSub(c *fiber.Ctx, db *sqlx.DB) error {
	// parse a simple request body (plain types are easier to work with)
	var body struct {
		MenuID    *int64 `json:"sm_id"`
		Name      string `json:"ss_name"`
		Link      string `json:"ss_link"`
		CreatedBy string `json:"ss_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Link = strings.TrimSpace(body.Link)
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ss_name is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ss_created_by is required"})
	}

	// prepare nullable sm parameter once
	var sm interface{}
	if body.MenuID != nil {
		sm = *body.MenuID
	} else {
		sm = nil
	}

	// Use a transaction to make max-order + insert atomic
	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}
	defer tx.Rollback()

	// Duplicate check scoped to the same menu (including NULL)
	var count int
	dupQ := `SELECT COUNT(*) FROM sys_submenu WHERE ((? IS NULL AND sm_id IS NULL) OR sm_id = ?) AND (ss_name = ? OR ss_link = ?)`
	if err := tx.Get(&count, dupQ, sm, sm, body.Name, body.Link); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	// compute next order for this menu
	var maxOrder sql.NullInt64
	maxQ := `SELECT MAX(ss_order) FROM sys_submenu WHERE ((? IS NULL AND sm_id IS NULL) OR sm_id = ?)`
	if err := tx.Get(&maxOrder, maxQ, sm, sm); err != nil {
		return c.Status(500).JSON(5)
	}
	next := int64(1)
	if maxOrder.Valid {
		next = maxOrder.Int64 + 1
	}

	// insert (use computed next order)
	if _, err := tx.Exec(`INSERT INTO sys_submenu (sm_id, ss_name, ss_link, ss_order, ss_status, ss_created_at, ss_created_by, ss_updated_at, ss_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ? ,?)`,
		sm,
		body.Name,
		body.Link,
		next,
		"active",
		time.Now(),
		body.CreatedBy,
		time.Now(),
		body.CreatedBy,
	); err != nil {
		return c.Status(500).JSON(5)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5)
	}

	return c.Status(200).JSON(1)

}

func UpdateMenuSub(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"ss_id"`
		Name      string `json:"ss_name"`
		Link      string `json:"ss_link"`
		Order     *int64 `json:"ss_order"`
		UpdatedBy string `json:"ss_updated_by"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ss_id is required"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ss_updated_by is required"})
	}

	// Check duplicate ss_name (exclude current record)
	if body.Name != "" {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM sys_submenu WHERE ss_name = ? AND ss_link = ? AND ss_id <> ?`, body.Name, body.Link, body.ID); err != nil {
			return c.Status(500).JSON(5)
		}
		if count > 0 {
			return c.Status(200).JSON(2)
		}
	}

	var orderVal interface{}
	if body.Order != nil {
		orderVal = *body.Order
	} else {
		orderVal = nil
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE sys_submenu SET ss_name = ?, ss_link = ?, ss_order = ?, ss_updated_at = ?, ss_updated_by = ? WHERE ss_id = ?`,
		body.Name,
		body.Link,
		orderVal,
		now,
		body.UpdatedBy,
		body.ID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update error", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "submenu not found"})
	}

	return c.Status(200).JSON(1)
}

func UpdateMenuSubStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"ss_id"`
		Status    string `json:"ss_status"`
		UpdatedBy string `json:"ss_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ss_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "ss_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ss_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE sys_submenu SET ss_status = ?, ss_updated_at = ?, ss_updated_by = ? WHERE ss_id = ?`,
		body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update submenu status", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "submenu not found"})
	}
	return c.Status(200).JSON(1)
}
