package routes

import (
	"horizon/controllers/users"
	middleware "horizon/middlewares"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {

	//Authorization
	app.Post("/user/signup", users.Signup)
	app.Post("/user/verify-otp", users.VerifyOTP)
	app.Post("/user/resend-otp", users.ResendOTP)
	app.Post("user/login", users.Login)
	app.Get("/auth/google/login", users.GoogleLogin)
	app.Get("/auth/google/callback", users.GoogleCallback)
	//View Products
	app.Get("/categories", users.ViewCategories)
	app.Get("/products", users.ViewProducts)
	app.Get("/product/filter", users.SearchProducts)

	userRoutes := app.Group("/user", middleware.AuthMiddleware)

	//Profile
	userRoutes.Get("/profile", users.ShowUserProfile)
	userRoutes.Post("/edit-profile", users.EditUserProfile)
	userRoutes.Get("/view-wallet", users.ViewWalletBalance)
	userRoutes.Get("/wallet-transaction", users.ViewWalletTransactions)

	//Address
	userRoutes.Post("/add-address", users.AddAddress)
	userRoutes.Get("/list-address", users.ViewAddresses)
	userRoutes.Put("/update-address", users.EditAddress)
	userRoutes.Delete("/addresses/:id", users.DeleteAddress)

	//wishlist
	userRoutes.Post("/add-wishlist", users.AddToWishlist)
	userRoutes.Delete("/wishlist/:product_id", users.RemoveFromWishlist)
	userRoutes.Delete("/clear-wishlist", users.ClearWishlist)
	userRoutes.Get("/wishlist", users.ViewWishlist)

	//Cart
	userRoutes.Post("/add-cart", users.AddToCart)
	userRoutes.Delete("/remove-cart/:product_id", users.RemoveFromCart)
	userRoutes.Get("/list-cart", users.ViewCart)
	userRoutes.Delete("/clear-cart", users.ClearCart)
	userRoutes.Get("/checkout", users.Checkout)
	userRoutes.Post("/wallet-purchase", users.UseWalletForPurchase)

	app.Get("paypal/success", users.PayPalSuccess)
	app.Get("paypal/cancel", users.PayPalCancel)

	//Coupon
	userRoutes.Post("/apply-coupon", users.ApplyCoupon)
	userRoutes.Get("/coupon", users.ViewCoupons)

	//Order
	userRoutes.Get("view-orders", users.ViewOrder)
	userRoutes.Post("cancel-order/:order_id", users.CancelOrder)

	//Invoice
	userRoutes.Get("/invoice/:orderID", users.GetInvoice)

}
