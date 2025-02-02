package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/api"
)

func main() {
	// // Load templates
	// pattern := filepath.Join("..", "templates", "*.html")
	// templates := template.Must(template.ParseGlob(pattern))

	// // Define static assets location
	// r := mux.NewRouter()
	// staticFilesDirectory := http.Dir("../assets/")
	// staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFilesDirectory))
	// r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	// // Define routes
	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	if err := templates.ExecuteTemplate(w, "index", nil); err != nil {
	// 		log.Fatalf("Failed to execute index template: %s", err)
	// 	}

	// }).Methods("GET")

	// log.Fatal(http.ListenAndServe(":8080", r))

	router := gin.Default()

	router.LoadHTMLGlob(filepath.Join("..", "templates", "*.html"))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"title": " Go PB - Home",
			},
		)
	})

	router.GET("/u/login", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"login.html",
			gin.H{
				"title": " Go PB - Login",
			},
		)
	})

	router.GET("/u/register", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"register.html",
			gin.H{
				"title": " Go PB - Register",
			},
		)
	})

	router.GET("/p/:id", func(c *gin.Context) {
		var p api.Paste
		id := c.Param("id")
		resp, err := http.Get("http://localhost:8080/paste/" + id)

		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}
		// Check API response status
		if resp.StatusCode != http.StatusOK {
			c.String(resp.StatusCode, "api: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			return
		}
		// Get the paste from the body
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}
		// Try to parse JSON into api.Paste
		if err := json.Unmarshal(b, &p); err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "failed to parse api response")
			return
		}
		// Send HTML
		c.HTML(
			http.StatusOK,
			"view.html",
			gin.H{
				"Title":    p.Title,
				"Body":     p.Body,
				"Language": p.Syntax,
			},
		)
	})

	router.POST("/p/", func(c *gin.Context) {
		var p api.Paste
		var data struct {
			api.Paste
			Url string
		}

		// Get the paste title and body from the form
		if b, ok := c.GetPostForm("body"); !ok || len(b) == 0 {
			c.String(http.StatusBadRequest, "body cannot be empty")
			return
		}
		p.Body = c.PostForm("body")
		p.Title = c.DefaultPostForm("title", "untitled")
		p.DeleteAfterRead, _ = strconv.ParseBool(c.PostForm("delete_after_read"))
		p.Syntax = c.DefaultPostForm("syntax", "none")

		// Try to create a new paste by calling the API
		paste, _ := json.Marshal(p)
		resp, err := http.Post("http://localhost:8080/paste", "application/json", bytes.NewBuffer(paste))

		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}

		// Check API response status
		if resp.StatusCode != http.StatusOK {
			c.String(resp.StatusCode, "api: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			return
		}

		// Get API response body
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected error parsing api response")
			return
		}

		if err := json.Unmarshal(b, &data); err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "failed to parse api response")
			return
		}

		c.Redirect(http.StatusFound, "/p/"+data.Url)
	})

	router.Static("/assets", "../assets")

	router.Run("127.0.0.1:8000")
}
