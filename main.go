package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gookit/ini/v2"
)

type q struct {
	Queue string `json:"queue"`
	TTL   int64  `json:"TTL"`
	ID    int    `json:"incident_id"`
}

type TTLMap struct {
	m map[string]*q
	l sync.Mutex
}

type co struct {
	host string
	port int
	wl   []string
	lp   string
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
var conf *co

func main() {
	conf = getConfig()
	m = New()
	//coment it for debug mode
	gin.SetMode(gin.ReleaseMode)
	f, err := os.Create(conf.lp)
	if err != nil {
		fmt.Println("Log file " + conf.lp + " not writable!")
		//return
	}
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()
	router.GET("/queue", getQueue)
	router.POST("/queue", postQueue)
	//needs error handling if port is not INT
	router.Run(fmt.Sprint(conf.host + ":" + strconv.Itoa(conf.port)))
}

func getConfig() (conf *co) {
	cp, e := os.LookupEnv("QUEUE_CONF")
	if !e {
		ex, _ := os.Executable()
		cp = fmt.Sprint(filepath.Dir(ex) + "config.ini")
	}
	err := ini.LoadExists(cp)
	if err != nil {
		panic(err)
	}
	conf = &co{}
	conf.host = ini.String("host")
	conf.port = ini.Int("port")
	conf.lp = ini.String("log_path")
	conf.wl = ini.Strings("white_list")
	return conf
}

func getQueue(c *gin.Context) {
	c.JSON(http.StatusOK, m.Get())
}

func postQueue(c *gin.Context) {
	var nq q
	// Call BindJSON to bind the received JSON to
	if err := c.BindJSON(&nq); err != nil {
		return
	}

	if contains(conf.wl, c.ClientIP()) {
		s := m.Put(fmt.Sprint(nq.Queue+strconv.Itoa(nq.ID)), nq)
		if s {
			c.JSON(http.StatusCreated, nq)
		} else {
			c.AbortWithStatusJSON(423, gin.H{"message": "record exist"})
		}
	} else {
		c.AbortWithStatusJSON(403, gin.H{"message": "Host not allowed"})
	}
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}
