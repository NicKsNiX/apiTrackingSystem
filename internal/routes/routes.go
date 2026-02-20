package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"apiTrackingSystem/internal/handlers"
)

// Setup - registers all API routes (requires db for handlers needing DB)
func Setup(app *fiber.App, db *sqlx.DB) {
	// Health check (prefixed)
	app.Get("/apiTrackingSystem/health", func(c *fiber.Ctx) error { return c.SendString("Database connected and server is running!") })
	app.Post("/apiTrackingSystem/login", func(c *fiber.Ctx) error { return handlers.Login(c, db) })
	// User routes
	app.Get("/apiTrackingSystem/user/GetUserByDepartment", func(c *fiber.Ctx) error { return handlers.GetUserByDepartment(c, db) })
	app.Get("/apiTrackingSystem/user/:username", func(c *fiber.Ctx) error { return handlers.GetUser(c, db) })
	app.Get("/apiTrackingSystem/users", func(c *fiber.Ctx) error { return handlers.ListUsers(c, db) })
	app.Post("/apiTrackingSystem/user/updateUserStatus", func(c *fiber.Ctx) error { return handlers.UpdateUserStatus(c, db) })
	app.Post("/apiTrackingSystem/user/UpdatePermissionGroupUser", func(c *fiber.Ctx) error { return handlers.UpdatePermissionGroupUser(c, db) })
	// Permission group routes
	app.Get("/apiTrackingSystem/permission/ListPermissionGroups", func(c *fiber.Ctx) error { return handlers.ListPermissionGroups(c, db) })
	app.Get("/apiTrackingSystem/permission/GetSelectPermissionGroups", func(c *fiber.Ctx) error { return handlers.GetSelectPermissionGroups(c, db) })
	app.Get("/apiTrackingSystem/permission/GetPermissionGroup", func(c *fiber.Ctx) error { return handlers.GetPermissionGroup(c, db) })
	app.Post("/apiTrackingSystem/permission/InsertPermissionGroup", func(c *fiber.Ctx) error { return handlers.InsertPermissionGroup(c, db) })
	app.Post("/apiTrackingSystem/permission/UpdatePermissionGroup", func(c *fiber.Ctx) error { return handlers.UpdatePermissionGroup(c, db) })
	app.Post("/apiTrackingSystem/permission/UpdatePermissionGroupStatus", func(c *fiber.Ctx) error { return handlers.UpdatePermissionGroupStatus(c, db) })
	// Permission detail routes
	app.Get("/apiTrackingSystem/permission/ListPermissionDetail", func(c *fiber.Ctx) error { return handlers.ListPermissionDetail(c, db) })
	app.Post("/apiTrackingSystem/permission/InsertPermissionDetail", func(c *fiber.Ctx) error { return handlers.InsertPermissionDetail(c, db) })
	app.Post("/apiTrackingSystem/permission/UpdatePermissionDetailStatus", func(c *fiber.Ctx) error { return handlers.UpdatePermissionDetailStatus(c, db) })
	// Menu routes
	app.Get("/apiTrackingSystem/menuGroup/ListMenusGroup", func(c *fiber.Ctx) error { return handlers.ListMenusGroup(c, db) })
	app.Get("/apiTrackingSystem/menuGroup/GetSelectMenuGroup", func(c *fiber.Ctx) error { return handlers.GetSelectMenuGroup(c, db) })
	app.Get("/apiTrackingSystem/menuGroup/GetMenuGroup", func(c *fiber.Ctx) error { return handlers.GetMenuGroup(c, db) })
	app.Post("/apiTrackingSystem/menuGroup/InsertMenuGroup", func(c *fiber.Ctx) error { return handlers.InsertMenuGroup(c, db) })
	app.Post("/apiTrackingSystem/menuGroup/UpdateMenuGroupStatus", func(c *fiber.Ctx) error { return handlers.UpdateMenuGroupStatus(c, db) })
	app.Post("/apiTrackingSystem/menuGroup/UpdateMenuGroup", func(c *fiber.Ctx) error { return handlers.UpdateMenuGroup(c, db) })
	// Submenu routes
	app.Get("/apiTrackingSystem/submenu/ListSubMenus", func(c *fiber.Ctx) error { return handlers.ListSubMenus(c, db) })
	app.Get("/apiTrackingSystem/submenu/GetSelectSubMenu", func(c *fiber.Ctx) error { return handlers.GetSelectSubMenu(c, db) })
	app.Get("/apiTrackingSystem/submenu/GetSubMenu", func(c *fiber.Ctx) error { return handlers.GetSubMenu(c, db) })
	app.Post("/apiTrackingSystem/submenu/InsertMenuSub", func(c *fiber.Ctx) error { return handlers.InsertMenuSub(c, db) })
	app.Post("/apiTrackingSystem/submenu/UpdateMenuSubStatus", func(c *fiber.Ctx) error { return handlers.UpdateMenuSubStatus(c, db) })
	app.Post("/apiTrackingSystem/submenu/UpdateMenuSub", func(c *fiber.Ctx) error { return handlers.UpdateMenuSub(c, db) })
	// Department routes
	app.Get("/apiTrackingSystem/departments/ListDepartments", func(c *fiber.Ctx) error { return handlers.ListDepartments(c, db) })
	app.Get("/apiTrackingSystem/department/GetDepartment", func(c *fiber.Ctx) error { return handlers.GetDepartment(c, db) })
	app.Post("/apiTrackingSystem/department/UpdateDepartment", func(c *fiber.Ctx) error { return handlers.UpdateDepartment(c, db) })
	app.Post("/apiTrackingSystem/department/UpdateDepartmentStatus", func(c *fiber.Ctx) error { return handlers.UpdateDepartmentStatus(c, db) })
	// Template routes
	app.Get("/apiTrackingSystem/templates/ListTemplates", func(c *fiber.Ctx) error { return handlers.ListTemplates(c, db) })
	app.Get("/apiTrackingSystem/template/:id", func(c *fiber.Ctx) error { return handlers.GetTemplate(c, db) })
	app.Post("/apiTrackingSystem/template/InsertTemplate", func(c *fiber.Ctx) error { return handlers.InsertTemplate(c, db) })
	app.Post("/apiTrackingSystem/template/UpdateTemplate", func(c *fiber.Ctx) error { return handlers.UpdateTemplate(c, db) })
	app.Post("/apiTrackingSystem/template/UpdateTemplateStatus", func(c *fiber.Ctx) error { return handlers.UpdateTemplateStatus(c, db) })
	// Master plan routes
	app.Get("/apiTrackingSystem/masterPlan/ListMasterPlan", func(c *fiber.Ctx) error { return handlers.ListMasterPlan(c, db) })
	app.Get("/apiTrackingSystem/masterPlan/GetMasterPlan", func(c *fiber.Ctx) error { return handlers.GetMasterPlan(c, db) })
	app.Get("/apiTrackingSystem/masterPlan/SelectAPQP", func(c *fiber.Ctx) error { return handlers.SelectAPQP(c, db) })
	app.Post("/apiTrackingSystem/masterPlan/InsertMasterPlan", func(c *fiber.Ctx) error { return handlers.InsertMasterPlan(c, db) })
	app.Post("/apiTrackingSystem/masterPlan/UpdateMasterPlan", func(c *fiber.Ctx) error { return handlers.UpdateMasterPlan(c, db) })
	app.Post("/apiTrackingSystem/masterPlan/UpdateMasterPlanStatus", func(c *fiber.Ctx) error { return handlers.UpdateMasterPlanStatus(c, db) })
	app.Get("/apiTrackingSystem/masterPlan/GetMasterPlanStep2", func(c *fiber.Ctx) error { return handlers.GetMasterPlanStep2(c, db) })
	app.Get("/apiTrackingSystem/masterPlan/GetMasterPlanStep2T", func(c *fiber.Ctx) error { return handlers.GetMasterPlanStep2T(c, db) })
	app.Get("/apiTrackingSystem/masterPlan/GetMasterPlanStep3", func(c *fiber.Ctx) error { return handlers.GetMasterPlanStep3(c, db) })

	app.Get("/apiTrackingSystem/manageTemplate/ListTemplateDetails", func(c *fiber.Ctx) error { return handlers.ListTemplateDetails(c, db) })
	app.Get("/apiTrackingSystem/manageTemplate/SelectTemplate", func(c *fiber.Ctx) error { return handlers.SelectTemplate(c, db) })
	app.Get("/apiTrackingSystem/manageTemplate/SelectMasterPlan", func(c *fiber.Ctx) error { return handlers.SelectMasterPlan(c, db) })
	app.Post("/apiTrackingSystem/manageTemplate/InsertTemplateDetail", func(c *fiber.Ctx) error { return handlers.InsertTemplateDetail(c, db) })
	app.Post("/apiTrackingSystem/manageTemplate/UpdateTemplateDetail", func(c *fiber.Ctx) error { return handlers.UpdateTemplateDetail(c, db) })
	app.Post("/apiTrackingSystem/manageTemplate/UpdateTemplateDetailStatus", func(c *fiber.Ctx) error { return handlers.UpdateTemplateDetailStatus(c, db) })

	app.Get("/apiTrackingSystem/manageWorkflow/ListWorkflow", func(c *fiber.Ctx) error { return handlers.ListWorkflow(c, db) })
	app.Get("/apiTrackingSystem/manageWorkflow/SelectDepartmentMW", func(c *fiber.Ctx) error { return handlers.SelectDepartmentMW(c, db) })
	app.Post("/apiTrackingSystem/manageWorkflow/InsertWorkflow", func(c *fiber.Ctx) error { return handlers.InsertWorkflow(c, db) })
	app.Post("/apiTrackingSystem/manageWorkflow/UpdateWorkflow", func(c *fiber.Ctx) error { return handlers.UpdateWorkflow(c, db) })
	app.Post("/apiTrackingSystem/manageWorkflow/UpdateWorkflowStatus", func(c *fiber.Ctx) error { return handlers.UpdateWorkflowStatus(c, db) })
	app.Get("/apiTrackingSystem/manageWorkflow/SelectUserMW", func(c *fiber.Ctx) error { return handlers.SelectUserMW(c, db) })

	app.Get("/apiTrackingSystem/manageAPQP/ListAPQP", func(c *fiber.Ctx) error { return handlers.ListAPQP(c, db) })
	app.Post("/apiTrackingSystem/manageAPQP/InsertAPQP", func(c *fiber.Ctx) error { return handlers.InsertAPQP(c, db) })
	app.Post("/apiTrackingSystem/manageAPQP/UpdateAPQPStatus", func(c *fiber.Ctx) error { return handlers.UpdateAPQPStatus(c, db) })
	app.Post("/apiTrackingSystem/manageAPQP/UpdateAPQP", func(c *fiber.Ctx) error { return handlers.UpdateAPQP(c, db) })
	app.Get("/apiTrackingSystem/manageAPQP/SelectPhaseAPQP", func(c *fiber.Ctx) error { return handlers.SelectPhaseAPQP(c, db) })
	app.Get("/apiTrackingSystem/manageAPQP/SelectAPQPS", func(c *fiber.Ctx) error { return handlers.SelectAPQPS(c, db) })
	app.Get("/apiTrackingSystem/manageAPQP/GetListAPQPPhase", func(c *fiber.Ctx) error { return handlers.GetListAPQPPhase(c, db) })

	app.Get("/apiTrackingSystem/managePhase/ListProjectPhases", func(c *fiber.Ctx) error { return handlers.ListProjectPhases(c, db) })
	app.Post("/apiTrackingSystem/managePhase/InsertProjectPhase", func(c *fiber.Ctx) error { return handlers.InsertProjectPhase(c, db) })
	app.Post("/apiTrackingSystem/managePhase/UpdateProjectPhase", func(c *fiber.Ctx) error { return handlers.UpdateProjectPhase(c, db) })
	app.Post("/apiTrackingSystem/managePhase/UpdateProjectPhaseStatus", func(c *fiber.Ctx) error { return handlers.UpdateProjectPhaseStatus(c, db) })

	app.Get("/apiTrackingSystem/managePPAP/ListPPAPItems", func(c *fiber.Ctx) error { return handlers.ListPPAPItems(c, db) })
	app.Post("/apiTrackingSystem/managePPAP/InsertPPAPItem", func(c *fiber.Ctx) error { return handlers.InsertPPAPItem(c, db) })
	app.Post("/apiTrackingSystem/managePPAP/UpdatePPAPItem", func(c *fiber.Ctx) error { return handlers.UpdatePPAPItem(c, db) })
	app.Post("/apiTrackingSystem/managePPAP/UpdatePPAPItemStatus", func(c *fiber.Ctx) error { return handlers.UpdatePPAPItemStatus(c, db) })
	app.Post("/apiTrackingSystem/managePPAP/InsertPPAPItemStep4", func(c *fiber.Ctx) error { return handlers.InsertPPAPItemStep4(c, db) })
	app.Post("/apiTrackingSystem/managePPAP/InsertPPAPItemStep4Draft", func(c *fiber.Ctx) error { return handlers.InsertPPAPItemStep4Draft(c, db) })
	app.Get("/apiTrackingSystem/managePPAP/GetListPPAPItems", func(c *fiber.Ctx) error { return handlers.GetListPPAPItems(c, db) })
	app.Get("/apiTrackingSystem/managePPAP/GetListPPAPStep4", func(c *fiber.Ctx) error { return handlers.GetListPPAPStep4(c, db) })

	app.Get("/apiTrackingSystem/manageTemplatePPAP/ListPPAPDetails", func(c *fiber.Ctx) error { return handlers.ListPPAPDetails(c, db) })
	app.Post("/apiTrackingSystem/manageTemplatePPAP/InsertPPAPDetail", func(c *fiber.Ctx) error { return handlers.InsertPPAPDetail(c, db) })
	app.Post("/apiTrackingSystem/manageTemplatePPAP/UpdatePPAPDetail", func(c *fiber.Ctx) error { return handlers.UpdatePPAPDetail(c, db) })

	app.Post("/apiTrackingSystem/manageTemplatePPAP/ListInfoProjects", func(c *fiber.Ctx) error { return handlers.ListInfoProjects(c, db) })
	app.Post("/apiTrackingSystem/manageTemplatePPAP/UpdateInfoProject", func(c *fiber.Ctx) error { return handlers.UpdateInfoProject(c, db) })

	app.Get("/apiTrackingSystem/manageModel/ListModelMaster", func(c *fiber.Ctx) error { return handlers.ListModelMaster(c, db) })
	app.Post("/apiTrackingSystem/manageModel/InsertModelMaster", func(c *fiber.Ctx) error { return handlers.InsertModelMaster(c, db) })
	app.Post("/apiTrackingSystem/manageModel/UpdateModelMaster", func(c *fiber.Ctx) error { return handlers.UpdateModelMaster(c, db) })
	app.Post("/apiTrackingSystem/manageModel/UpdateModelMasterStatus", func(c *fiber.Ctx) error { return handlers.UpdateModelMasterStatus(c, db) })

	app.Get("/apiTrackingSystem/manageProject/ListInfoProjects", func(c *fiber.Ctx) error { return handlers.ListInfoProjects(c, db) })
	app.Get("/apiTrackingSystem/manageProject/SelectPartNumber", func(c *fiber.Ctx) error { return handlers.SelectPartNumber(c, db) })
	app.Get("/apiTrackingSystem/manageProject/SelectModel", func(c *fiber.Ctx) error { return handlers.SelectModel(c, db) })
	app.Post("/apiTrackingSystem/manageProject/IssueInfoProject", func(c *fiber.Ctx) error { return handlers.IssueInfoProject(c, db) })
	app.Post("/apiTrackingSystem/manageProject/UpdateInfoProject", func(c *fiber.Ctx) error { return handlers.UpdateInfoProject(c, db) })
	app.Get("/apiTrackingSystem/manageProject/SelectPPAPItem", func(c *fiber.Ctx) error { return handlers.SelectPPAPItem(c, db) })
	app.Get("/apiTrackingSystem/manageProject/GetMaxCode", func(c *fiber.Ctx) error { return handlers.GetMaxCode(c, db) })
	app.Get("/apiTrackingSystem/manageProject/SelectCountAddedStatus", func(c *fiber.Ctx) error { return handlers.SelectCountAddedStatus(c, db) })
	app.Get("/apiTrackingSystem/manageProject/CustomerEventGanttChart", func(c *fiber.Ctx) error { return handlers.CustomerEventGanttChart(c, db) })
	app.Get("/apiTrackingSystem/manageProject/InternalEventGanttChart", func(c *fiber.Ctx) error { return handlers.InternalEventGanttChart(c, db) })
	app.Get("/apiTrackingSystem/manageProject/GetinfoGanttchart", func(c *fiber.Ctx) error { return handlers.GetinfoGanttchart(c, db) })
	app.Get("/apiTrackingSystem/manageProject/GetListAPQPPPAPItem", func(c *fiber.Ctx) error { return handlers.GetListAPQPPPAPItem(c, db) })
	app.Post("/apiTrackingSystem/manageProject/UpdateStatusProject", func(c *fiber.Ctx) error { return handlers.UpdateStatusProject(c, db) })
	app.Post("/apiTrackingSystem/manageProject/InsertProjectStep3", func(c *fiber.Ctx) error { return handlers.InsertProjectStep3(c, db) })
	app.Post("/apiTrackingSystem/manageProject/UpdateStatusProjectItemDetail", func(c *fiber.Ctx) error { return handlers.UpdateStatusProjectItemDetail(c, db) })
	app.Post("/apiTrackingSystem/manageProject/UpdateStatusCompleteProject", func(c *fiber.Ctx) error { return handlers.UpdateStatusCompleteProject(c, db) })

	app.Get("/apiTrackingSystem/menu/GetUserMenu", func(c *fiber.Ctx) error { return handlers.GetUserMenu(c, db) })

	app.Get("/apiTrackingSystem/manageProjectTracking/ListProjectItemDetails", func(c *fiber.Ctx) error { return handlers.ListProjectItemDetails(c, db) })
	app.Get("/apiTrackingSystem/manageProjectTracking/CountProjectTracking", func(c *fiber.Ctx) error { return handlers.CountProjectTracking(c, db) })
	app.Get("/apiTrackingSystem/manageProjectTracking/GetListProjectTracking", func(c *fiber.Ctx) error { return handlers.GetListProjectTracking(c, db) })
	app.Post("/apiTrackingSystem/manageProjectTracking/InsertProjectTracking", func(c *fiber.Ctx) error { return handlers.InsertProjectTracking(c, db) })
	// app.Get("/apiTrackingSystem/manageProjectTracking/SaveFileSendEmail", func(c *fiber.Ctx) error { return handlers.SaveFileSendEmail(c, db) })

	app.Post("/apiTrackingSystem/projectMasterPlan/InsertProjectMasterPlan", func(c *fiber.Ctx) error { return handlers.InsertProjectMasterPlan(c, db) })

	app.Get("/apiTrackingSystem/remainTask/ListRemainTasks", func(c *fiber.Ctx) error { return handlers.ListRemainTasks(c, db) })
	app.Get("/apiTrackingSystem/remainTask/NotifyRemainTasks", func(c *fiber.Ctx) error { return handlers.NotifyRemainTasks(c, db) })
	app.Get("/apiTrackingSystem/remainTask/SelectModelMaster", func(c *fiber.Ctx) error { return handlers.SelectModelMaster(c, db) })
	app.Get("/apiTrackingSystem/remainTask/SelectPartNumberRT", func(c *fiber.Ctx) error { return handlers.SelectPartNumberRT(c, db) })
	app.Get("/apiTrackingSystem/remainTask/SelectKickOffDateRT", func(c *fiber.Ctx) error { return handlers.SelectKickOffDateRT(c, db) })
	app.Get("/apiTrackingSystem/remainTask/GetCountItem", func(c *fiber.Ctx) error { return handlers.GetCountItem(c, db) })
	app.Post("/apiTrackingSystem/remainTask/UpdateStatusFileProject", func(c *fiber.Ctx) error { return handlers.UpdateStatusFileProject(c, db) })

	app.Get("/apiTrackingSystem/sendMail/SendMailAuto", func(c *fiber.Ctx) error { return handlers.SendMailAuto(c, db) })

	app.Get("/apiTrackingSystem/dashboard/ListInprogressProjects", func(c *fiber.Ctx) error { return handlers.ListInprogressProjects(c, db) })
	app.Get("/apiTrackingSystem/dashboard/MasterPlanSummary", func(c *fiber.Ctx) error { return handlers.ListMasterPlanSummary(c, db) })

	app.Static("/uploads", `C:\inetpub\wwwroot\apiTrackingSystemUat\uploads`)
}

// http://192.168.161.219:9004/apiTrackingSystem/manageProject/UpdateStatusCompleteProject
// http://192.168.161.219:9004/apiTrackingSystem/manageModel/InsertModelMaster
// http://192.168.161.219:9004/apiTrackingSystem/manageModel/UpdateModelMaster
// http://192.168.161.219:9004/apiTrackingSystem/manageModel/UpdateModelMasterStatus
// docker compose down
// docker compose up -d --build
