package handlers

import (
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// MenuChild is a single submenu item in the response
type MenuChild struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Order  int    `json:"order"`
	Status string `json:"status"`
}

// MenuGroup is a top-level menu with children
type MenuGroup struct {
	Name     string      `json:"name"`
	Order    int         `json:"order"`
	Icon     string      `json:"icon"`
	Children []MenuChild `json:"children"`
}

func GetUserMenu(c *fiber.Ctx, db *sqlx.DB) error {

	uid := c.Params("id")
	if uid == "" {
		uid = c.Query("su_id")
	}
	if uid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "su_id is required (param or query)"})
	}
	suID, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid su_id"})
	}

	// query rows
	var rows []struct {
		SsID    int64  `db:"ss_id"`
		SpgID   int64  `db:"spg_id"` // unused but kept if needed
		SpdID   int64  `db:"spd_id"`
		SmName  string `db:"sm_name"`
		SmIcon  string `db:"sm_icon"`
		SmOrder int    `db:"sm_order"`
		SsName  string `db:"ss_name"`
		SsLink  string `db:"ss_link"`
		SsOrder int    `db:"ss_order"`
		SpdStat string `db:"spd_status"`
	}

	query := `SELECT
	  ss.ss_id,
	  spg.spg_id,
	  spd.spd_id,
			sm.sm_name,
			sm.sm_icon,
	  sm.sm_order,
	  ss.ss_name,
	  ss.ss_link,
	  ss.ss_order,
	  spd.spd_status
	FROM
	  sys_user su
	  LEFT JOIN sys_permission_group spg ON su.spg_id = spg.spg_id
	  LEFT JOIN sys_permission_detail spd ON spg.spg_id = spd.spg_id
	  LEFT JOIN sys_submenu ss ON ss.ss_id = spd.ss_id
	  LEFT JOIN sys_menu sm ON ss.sm_id = sm.sm_id
	WHERE
	  spd.spd_status = 'active' AND
	  ss.ss_status = 'active' AND
	  sm.sm_status = 'active' AND
	  su.su_status = 'active' AND
	  su.su_id = ?
	ORDER BY
	  sm.sm_order,
	  ss.ss_order`

	if err := db.Select(&rows, query, suID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}

	// group by menu (sm_name + order). Use sm_name as key but keep order using map and slice
	type groupKey struct {
		Name  string
		Order int
	}
	type groupVal struct {
		Children []MenuChild
		Icon     string
	}
	groups := map[groupKey]groupVal{}
	for _, r := range rows {
		key := groupKey{Name: r.SmName, Order: r.SmOrder}
		child := MenuChild{ID: r.SsID, Name: r.SsName, Link: r.SsLink, Order: r.SsOrder, Status: r.SpdStat}
		v := groups[key]
		v.Children = append(v.Children, child)
		// set icon from the row (overwrites but same value per group)
		v.Icon = r.SmIcon
		groups[key] = v
	}

	// build slice and sort by order
	out := make([]MenuGroup, 0, len(groups))
	for k, v := range groups {
		// sort children by order
		sort.Slice(v.Children, func(i, j int) bool { return v.Children[i].Order < v.Children[j].Order })
		out = append(out, MenuGroup{Name: k.Name, Order: k.Order, Icon: v.Icon, Children: v.Children})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Order < out[j].Order })

	return c.Status(200).JSON(out)
}
