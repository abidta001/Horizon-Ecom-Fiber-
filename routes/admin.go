package routes

import (
	"horizon/controllers/admin"
	middleware "horizon/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(app *fiber.App) {

	//Authorization
	app.Post("/admin/login", admin.AdminLogin)

	//User Management
	app.Get("/admin/users", middleware.AdminJWT, admin.ViewUsers)
	app.Post("/admin/block-user", middleware.AdminJWT, admin.BlockUser)
	app.Post("/admin/unblock-user", middleware.AdminJWT, admin.UnblockUser)

	//Category Management
	app.Post("/admin/add-category", middleware.AdminJWT, admin.AddCategory)
	app.Put("/admin/edit-category", middleware.AdminJWT, admin.EditCategory)
	app.Delete("/admin/delete-category/:id", middleware.AdminJWT, admin.SoftDeleteCategory)
	app.Post("/admin/recover-category/:id", middleware.AdminJWT, admin.RecoverCategory)
	app.Get("/admin/view-categories", middleware.AdminJWT, admin.AdminViewCategories)

	//Product Management
	app.Post("/admin/add-products", middleware.AdminJWT, admin.AddProduct)
	app.Put("/admin/edit-product", middleware.AdminJWT, admin.EditProduct)
	app.Delete("/admin/delete-product/:id", middleware.AdminJWT, admin.SoftDeleteProduct)
	app.Post("/admin/recover-product/:id", middleware.AdminJWT, admin.RecoverProduct)
	app.Get("/admin/view-products", middleware.AdminJWT, admin.AdminViewProducts)
	app.Put("/admin/update-stock", middleware.AdminJWT, admin.UpdateProductStock)

	//Offer Management
	app.Post("/admin/add-offer", middleware.AdminJWT, admin.AddOffer)
	app.Delete("/admin/remove-offer/:product_id", middleware.AdminJWT, admin.RemoveOffer)
	app.Get("/admin/view-offers", middleware.AdminJWT, admin.ViewOffer)

	//Coupon Management
	app.Post("/admin/add-coupon", middleware.AdminJWT, admin.CreateCoupon)
	app.Get("/admin/view-coupon", middleware.AdminJWT, admin.ViewCouponsAdmin)
	app.Delete("/admin/remove-coupon/:id", middleware.AdminJWT, admin.RemoveCoupon)

	//Order Management
	app.Get("/admin/order-details", middleware.AdminJWT, admin.AdminListOrder)
	app.Patch("/admin/order-status/:order_id", middleware.AdminJWT, admin.AdminChangeOrderStatus)

	//Sales Report & DashBoard
	app.Post("/sales-report", middleware.AdminJWT, admin.GenerateSalesReport)
	app.Get("/dashboard/:period", middleware.AdminJWT, admin.GenerateDashboardReport)
	app.Get("/top-selling-report", middleware.AdminJWT, admin.GenerateTopSellingReport)
}
