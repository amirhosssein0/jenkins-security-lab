package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "up",
        })
    })

    r.GET("/ready", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
        "status": "ready",
        })
    })

    r.Run(":8080")
}