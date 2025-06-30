package main

import (
	"net/http"

	"github.com/azikazi/azikazi/database"
	"github.com/azikazi/azikazi/functions"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	database.ConnectDatabase()

	db := &database.Database{DB: database.DB} // db points to database.Database  which will store database.DB into its variable DB .... /:transactionid
	db.InitDatabase()

	router.POST("/createUser", functions.CreateUser)
	router.POST("/login", functions.LoginUser)
	// router.POST("/postask", functions.PostTask)
	router.POST("/postask", functions.AuthMiddleware(),  functions.PostTask)
	router.GET("/getask",  functions.GetAllTasks)

	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Your server is running well !!",
		})
	})
	err := router.Run(":1010")
	if err != nil {
		panic(err)
	}
}

