package server

import (
	activity_log "mangea-backend/internal/activity_log"
	"mangea-backend/internal/auth"
	"mangea-backend/internal/category"
	"mangea-backend/internal/dashboard"
	"mangea-backend/internal/order"
	"mangea-backend/internal/product"
	"mangea-backend/internal/report"
	"mangea-backend/internal/sync"
	"mangea-backend/internal/table"
	"mangea-backend/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type RouterDeps struct {
	DB        *sqlx.DB
	JWTSecret string
}

func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")

	// Public routes (no auth)
	registerAuthRoutes(v1, deps.DB, deps.JWTSecret)

	// Protected routes (auth required)
	protected := v1.Group("")
	protected.Use(auth.AuthMiddleware(deps.JWTSecret))
	{
		registerCategoryRoutes(protected, deps.DB)
		registerProductRoutes(protected, deps.DB)
		registerTableRoutes(protected, deps.DB)
		registerOrderRoutes(protected, deps.DB)
		registerSyncRoutes(protected, deps.DB)
		registerDashboardRoutes(protected, deps.DB)
		registerActivityLogRoutes(protected, deps.DB)
		registerStockRoutes(protected, deps.DB)
		registerReportRoutes(protected, deps.DB)
		registerUserManagementRoutes(protected, deps.DB)
	}

	return router
}

func registerAuthRoutes(rg *gin.RouterGroup, db *sqlx.DB, jwtSecret string) {
	repo := user.NewRepository(db)
	handler := user.NewHandler(repo, jwtSecret)

	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/login", handler.Login)
		authGroup.GET("/me", auth.AuthMiddleware(jwtSecret), handler.GetCurrentUser)
	}
}

func registerCategoryRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := category.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerProductRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := product.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerTableRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := table.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerOrderRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := order.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerSyncRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := sync.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerDashboardRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := dashboard.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerActivityLogRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := activity_log.NewHandler(db)
	handler.RegisterRoutes(rg)
}

func registerUserManagementRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	repo := user.NewRepository(db)
	handler := user.NewHandler(repo, "")

	users := rg.Group("/users")
	{
		// Read access: any authenticated user
		users.GET("", auth.RequireRole("admin", "owner"), handler.ListUsers)
		users.GET("/:id", handler.GetUserByID)

		// Management: admin/owner only
		users.POST("", auth.RequireRole("admin", "owner"), handler.CreateUser)
		users.PUT("/:id", auth.RequireRole("admin", "owner"), handler.UpdateUser)
		users.DELETE("/:id", auth.RequireRole("admin", "owner"), handler.DeleteUser)

		// Change password: self or admin/owner (enforced in handler)
		users.POST("/:id/change-password", handler.ChangePassword)
	}
}

func registerStockRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := product.NewHandler(db)

	stock := rg.Group("/stock")
	{
		stock.GET("/low", handler.GetLowStockProducts)
		stock.GET("/out", handler.GetOutOfStockProducts)
		stock.GET("/statistics", handler.GetStockStatistics)
		stock.PATCH("/products/:id", handler.UpdateStock)
		stock.POST("/products/:id/add", handler.AddStock)
		stock.POST("/products/:id/reduce", handler.ReduceStock)
	}
}

func registerReportRoutes(rg *gin.RouterGroup, db *sqlx.DB) {
	handler := report.NewHandler(db)
	handler.RegisterRoutes(rg)
}
