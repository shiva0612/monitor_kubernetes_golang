package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/importConfig", importConfig)
	log.Println(r.Run(":8080"))
}
