package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/configs"
	"github.com/gottatouchsomegrass/url/pkg/routes"
	"github.com/gottatouchsomegrass/url/pkg/utils"
	"github.com/gottatouchsomegrass/url/platform/database"
	"github.com/joho/godotenv"
)

 func main()  {
	 if err:=godotenv.Load(); err!=nil {
		 log.Fatal("error loading .env file")
	 }

	 r := gin.Default()

	 dbq,err := database.OpenDBConnection()
	 if err!=nil {
		 log.Fatal(err)
	 }

	 urlController := controllers.NewURLController(
		 dbq.UrlQuery,
	 )

	 api := r.Group("/api/v1")
	 routes.PublicRoutes(api,urlController)

	 svr := configs.ConfigHTTPServer(r)
	 utils.StartSvrGracefulShutdown(svr)
 }
