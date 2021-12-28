package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type q struct {
	Timestamp int64  `json:"epoch"`
	Queue     string `json:"queue"`
	TTL       int64  `json:"TTL"`
	ID        int    `json:"incident_id"`
}

type TTLMap struct {
	m map[string]*q
	l sync.Mutex
}

func New() (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*q)}
	go func() {
		for range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				v.TTL--
				if v.TTL <= 0 {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k string, v q) bool {
	var s bool
	m.l.Lock()
	if _, ok := m.m[k]; !ok {
		m.m[k] = &v
		s = true
	} else {
		s = false
	}
	m.l.Unlock()
	return s
}

func (m *TTLMap) Get() []*q {
	m.l.Lock()
	var l []*q
	for _, v := range m.m {
		l = append(l, v)
	}
	m.l.Unlock()
	return l
}

var m *TTLMap

func main() {
	m = New()
	router := gin.Default()
	router.GET("/queue", getQueue)
	//router.GET("/albums/:id", getAlbumByID)
	router.POST("/queue", postQueue)
	router.Run(":8080")
}

func getQueue(c *gin.Context) {
	c.JSON(http.StatusOK, m.Get())
}

func postQueue(c *gin.Context) {
	// fmt.Println(&c)
	var nq q

	//set default epoch value to Now

	//newQueue.Timestamp = time.Now().UnixMilli()
	// Call BindJSON to bind the received JSON to
	// newQueue.
	if err := c.BindJSON(&nq); err != nil {
		return
	}
	s := m.Put(fmt.Sprint(nq.Queue+strconv.Itoa(nq.ID)), nq)
	if s {
		c.JSON(http.StatusCreated, nq)
	} else {
		c.AbortWithStatusJSON(423, gin.H{"message": "record exist"})
	}

}
