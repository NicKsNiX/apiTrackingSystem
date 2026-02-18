package handlers

import (
	"strings"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type SysProjectMasterPlan struct {
	IpmpID    int64            `db:"ipmp_id" json:"ipmp_id"`
	IpID      int64            `db:"ip_id" json:"ip_id"`
	Name      utils.NullString `db:"ipmp_name" json:"ipmp_name"`
	Date      *Date            `db:"ipmp_date" json:"ipmp_date"`
	StartDate *Date            `db:"ipmp_start_date" json:"ipmp_start_date"`
	EndDate   *Date            `db:"ipmp_end_date" json:"ipmp_end_date"`
	Type      utils.NullString `db:"ipmp_type" json:"ipmp_type"`
	Status    utils.NullString `db:"ipmp_status" json:"ipmp_status"`
	CreatedAt *DateTime        `db:"ipmp_created_at" json:"ipmp_created_at"`
	CreatedBy utils.NullString `db:"ipmp_created_by" json:"ipmp_created_by"`
}

// ListProjectMasterPlans returns master plans, optional filter by ip_id
func ListProjectMasterPlans(c *fiber.Ctx, db *sqlx.DB) error {
	ipID := c.Query("ip_id")
	query := `SELECT ipmp_id, ip_id, ipmp_name, ipmp_date, ipmp_start_date, ipmp_end_date, ipmp_type, ipmp_status, ipmp_created_at, ipmp_created_by FROM info_project_master_plan WHERE 1=1`
	args := []interface{}{}
	if ipID != "" {
		query += " AND ip_id = ?"
		args = append(args, ipID)
	}
	query += " ORDER BY ipmp_id ASC"

	var list []SysProjectMasterPlan
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

// GetProjectMasterPlan returns a single master plan by id
func GetProjectMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var item SysProjectMasterPlan
	if err := db.Get(&item, `SELECT ipmp_id, ip_id, ipmp_name, ipmp_date, ipmp_start_date, ipmp_end_date, ipmp_type, ipmp_status, ipmp_created_at, ipmp_created_by FROM info_project_master_plan WHERE ipmp_id = ?`, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(item)
}

// InsertProjectMasterPlan inserts a new master plan
func InsertProjectMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	// support both single object and array of objects (รองรับทั้ง object เดียวและ array)
	var reqs []SysProjectMasterPlan
	if err := c.BodyParser(&reqs); err != nil || len(reqs) == 0 {
		var single SysProjectMasterPlan
		if err2 := c.BodyParser(&single); err2 != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		reqs = append(reqs, single)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "begin tx failed", "detail": err.Error()})
	}
	// safety (กันพลาด rollback) - commit แล้ว rollback จะไม่ทำงาน
	defer func() { _ = tx.Rollback() }()

	insertQuery := `
		INSERT INTO info_project_master_plan
			(ip_id, ipmp_name, ipmp_date, ipmp_start_date, ipmp_end_date, ipmp_type, ipmp_status, ipmp_created_at, ipmp_created_by)
		VALUES
			(:ip_id, :ipmp_name, :ipmp_date, :ipmp_start_date, :ipmp_end_date, :ipmp_type, :ipmp_status, :ipmp_created_at, :ipmp_created_by)
	`

	ids := []int64{}

	// group incoming requests by ip_id (จัดกลุ่มตาม ip_id)
	groups := map[int64][]SysProjectMasterPlan{}
	for _, r := range reqs {
		groups[r.IpID] = append(groups[r.IpID], r)
	}

	// helper: compare dates (ฟังก์ชันเทียบวันที่)
	equalDates := func(a, b *Date) bool {
		if a == nil && b == nil {
			return true
		}
		if a == nil || b == nil {
			return false
		}
		return a.Time.Equal(b.Time)
	}

	for ipID, items := range groups {
		// fetch existing rows for this ip_id
		var existingRows []struct {
			ID        int64            `db:"ipmp_id"`
			Name      utils.NullString `db:"ipmp_name"`
			StartDate *Date            `db:"ipmp_start_date"`
			EndDate   *Date            `db:"ipmp_end_date"`
		}

		if err := tx.Select(&existingRows,
			`SELECT ipmp_id, ipmp_name, ipmp_start_date, ipmp_end_date FROM info_project_master_plan WHERE ip_id = ?`,
			ipID,
		); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query existing failed", "detail": err.Error()})
		}

		// map existing by name (ทำ map ของข้อมูลเดิมตาม name)
		existingMap := map[string]struct {
			ID    int64
			Start *Date
			End   *Date
		}{}
		for _, e := range existingRows {
			existingMap[e.Name.StringValue()] = struct {
				ID    int64
				Start *Date
				End   *Date
			}{ID: e.ID, Start: e.StartDate, End: e.EndDate}
		}

		// dedupe incoming names and preserve order (ตัดซ้ำตาม name และเก็บลำดับ)
		incomingSet := map[string]SysProjectMasterPlan{}
		incomingOrder := []string{}
		for _, it := range items {
			name := strings.TrimSpace(it.Name.StringValue())
			if name == "" {
				continue
			}
			// keep first occurrence (เก็บตัวแรก)
			if _, ok := incomingSet[name]; !ok {
				incomingSet[name] = it
				incomingOrder = append(incomingOrder, name)
			}
		}

		// check overlap (เช็คว่าชื่อที่ส่งมาซ้ำกับของเดิมกี่รายการ)
		overlapCount := 0
		for name := range incomingSet {
			if _, ok := existingMap[name]; ok {
				overlapCount++
			}
		}

		// if there are existing rows and no overlap -> replace (delete then insert)
		// ถ้ามีของเดิม และชื่อที่ส่งมา "ไม่ทับ" เลย => replace ทั้งชุด
		if len(existingRows) > 0 && overlapCount == 0 {
			if _, err := tx.Exec(`DELETE FROM info_project_master_plan WHERE ip_id = ?`, ipID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "delete failed", "detail": err.Error()})
			}

			for _, name := range incomingOrder {
				it := incomingSet[name]
				params := map[string]any{
					"ip_id":           it.IpID,
					"ipmp_name":       it.Name,
					"ipmp_date":       it.Date,
					"ipmp_start_date": it.StartDate,
					"ipmp_end_date":   it.EndDate,
					"ipmp_type":       "dateRange",
					"ipmp_status":     "inprogress",
					"ipmp_created_at": now,
					"ipmp_created_by": it.CreatedBy,
				}

				res, err := tx.NamedExec(insertQuery, params)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "insert failed", "detail": err.Error()})
				}
				id, _ := res.LastInsertId()
				ids = append(ids, id)
			}

			// IMPORTANT: if replace whole set, no need to "delete missing" again (กรณี replace ไม่ต้องลบซ้ำ)
			continue
		}

		// otherwise: upsert by name (ไม่ replace: insert เฉพาะใหม่ + update ถ้าวันเปลี่ยน)
		for _, name := range incomingOrder {
			it := incomingSet[name]

			if ex, ok := existingMap[name]; ok {
				// exists: update if dates changed (มีอยู่แล้ว: update ถ้าวันเปลี่ยน)
				if !equalDates(ex.Start, it.StartDate) || !equalDates(ex.End, it.EndDate) {
					params := map[string]any{
						"ipmp_id":         ex.ID,
						"ipmp_date":       it.Date,
						"ipmp_start_date": it.StartDate,
						"ipmp_end_date":   it.EndDate,
						"ipmp_type":       "dateRange",
						"ipmp_status":     "inprogress",
					}

					_, err := tx.NamedExec(`
						UPDATE info_project_master_plan
						SET ipmp_date = :ipmp_date,
						    ipmp_start_date = :ipmp_start_date,
						    ipmp_end_date = :ipmp_end_date,
						    ipmp_type = :ipmp_type,
						    ipmp_status = :ipmp_status
						WHERE ipmp_id = :ipmp_id
					`, params)
					if err != nil {
						return c.Status(500).JSON(5)
					}
					ids = append(ids, ex.ID)
				}
				continue
			}

			// insert new (เพิ่มใหม่)
			params := map[string]any{
				"ip_id":           it.IpID,
				"ipmp_name":       it.Name,
				"ipmp_date":       it.Date,
				"ipmp_start_date": it.StartDate,
				"ipmp_end_date":   it.EndDate,
				"ipmp_type":       "dateRange",
				"ipmp_status":     "inprogress",
				"ipmp_created_at": now,
				"ipmp_created_by": it.CreatedBy,
			}

			res, err := tx.NamedExec(insertQuery, params)
			if err != nil {
				return c.Status(500).JSON(5)
			}
			id, _ := res.LastInsertId()
			ids = append(ids, id)
		}

		// ✅ FIXED: delete existing rows that were not sent in the incoming list (per ip_id)
		// ✅ แก้แล้ว: ย้ายมาลบใน loop ของ ipID นี้เท่านั้น
		incomingNamesNorm := map[string]bool{}
		for n := range incomingSet {
			nn := strings.ToLower(strings.TrimSpace(n))
			if nn != "" {
				incomingNamesNorm[nn] = true
			}
		}

		for en, ex := range existingMap {
			enn := strings.ToLower(strings.TrimSpace(en))
			if enn == "" {
				continue
			}
			if !incomingNamesNorm[enn] {
				if _, err := tx.Exec(`DELETE FROM info_project_master_plan WHERE ipmp_id = ?`, ex.ID); err != nil {
					return c.Status(500).JSON(5)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5)
	}

	// you currently return 1 always (ตอนนี้คุณ return 1 ตลอด) - คง behavior เดิมไว้
	_ = ids // keep if you want to use later (กัน unused ในอนาคต)
	return c.Status(201).JSON(1)
}

// UpdateProjectMasterPlan updates an existing master plan
func UpdateProjectMasterPlan(c *fiber.Ctx, db *sqlx.DB) error {
	var req SysProjectMasterPlan
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if req.IpmpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ipmp_id is required"})
	}

	params := map[string]any{
		"ipmp_id":         req.IpmpID,
		"ipmp_name":       req.Name,
		"ipmp_date":       req.Date,
		"ipmp_start_date": req.StartDate,
		"ipmp_end_date":   req.EndDate,
		"ipmp_type":       req.Type,
		"ipmp_status":     "inprogress",
	}

	res, err := db.NamedExec(`
        UPDATE info_project_master_plan SET
            ipmp_name = :ipmp_name,
            ipmp_date = :ipmp_date,
            ipmp_start_date = :ipmp_start_date,
            ipmp_end_date = :ipmp_end_date,
            ipmp_type = :ipmp_type,
            ipmp_status = :ipmp_status
        WHERE ipmp_id = :ipmp_id
    `, params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed", "detail": err.Error()})
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.Status(200).JSON(1)
}
