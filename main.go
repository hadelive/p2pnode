package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/hadelive/p2pnode/entity"
)

func main() {
	ctx := context.Background()
	mempool := NewMempool()

	// Bootstrap peers from env
	peers := []string{}
	if ps := os.Getenv("PEERS"); ps != "" {
			peers = strings.Split(ps, ",")
	}
	node, err := NewNode(ctx, mempool, peers)
	if err != nil {
			log.Fatal(err)
	}

	router := gin.Default()
	router.POST("/tx", func(c *gin.Context) {
			var tx Transaction
			if err := c.ShouldBindJSON(&tx); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TX"})
					return
			}
			mempool.Add(tx)
			if err := node.Broadcast(tx); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Broadcast fail"})
					return
			}
			c.JSON(http.StatusAccepted, gin.H{"status": "OK"})
	})
	router.GET("/mempool", func(c *gin.Context) {
			c.JSON(http.StatusOK, mempool.All())
	})
	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
