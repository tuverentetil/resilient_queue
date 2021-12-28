package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type q struct {
	Timestamp int64  `json:"epoch"`
	Queue     string `json:"queue"`
	TTL       int64  `json:"TTL"`
	ID        uint   `json:"incident_id"`
}

// albums slice to seed record album data.
var queue = []q{}

func main() {
	router := gin.Default()
	router.GET("/queue", getQueue)
	//router.GET("/albums/:id", getAlbumByID)
	router.POST("/queue", postQueue)

	router.Run("localhost:8080")
}

func isAlive(rec *q) (int64, error) {
	now := time.Now().Unix()
	if (rec.Timestamp + rec.TTL) > now {
		fmt.Println("still valid")
		return (rec.Timestamp + rec.TTL) - now, errors.New("record still active")
	} else {
		return 0, nil
	}
}
func recordEx(rec *q) (bool, int64) {
	for i, e := range queue {
		if e.Queue == rec.Queue && e.ID == rec.ID {
			//send existing one for check if it's still valid
			t, err := isAlive(&e)
			if err != nil {
				fmt.Println(err)
				return false, t
			} else {
				queue[i] = queue[len(queue)-1]
				queue = queue[:len(queue)-1]
			}
		}
	}
	return true, 0
}

// getAlbums responds with the list of all albums as JSON.
func getQueue(c *gin.Context) {
	c.JSON(http.StatusOK, queue)
}

// postQueue adds an album from JSON received in the request body.
func postQueue(c *gin.Context) {
	// fmt.Println(&c)
	var newQueue q

	//set default epoch value to Now

	newQueue.Timestamp = time.Now().Unix()
	// Call BindJSON to bind the received JSON to
	// newQueue.
	if err := c.BindJSON(&newQueue); err != nil {
		return
	}
	ex, t := recordEx(&newQueue)
	if ex {
		// Add the new item to the queue.
		queue = append(queue, newQueue)
		c.JSON(http.StatusCreated, newQueue)
	} else {
		c.AbortWithStatusJSON(423, gin.H{"message": "record exist", "ttl": t})
	}

}
