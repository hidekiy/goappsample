package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

var (
	db  *sql.DB
	tpl = template.Must(template.ParseGlob("templates/*.html"))
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", http.DefaultServeMux))
	}()

	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/test?loc=Local&parseTime=true")
	if err != nil {
		log.Fatalln("Failed to connect to DB:", err)
	}
	defer db.Close()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/template", TestTemplate)
	router.GET("/api", TestAPI)
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		http.Error(w, fmt.Sprint(v), 500)
	}

	fmt.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := tpl.ExecuteTemplate(w, "index.html", nil)
	checkErr(err)
}

func TestTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{
			"My photos",
			"My blog",
		},
	}
	err := tpl.ExecuteTemplate(w, "test.html", data)
	checkErr(err)
}

func TestAPI(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rows, err := db.Query(`SELECT 1, "string123", TRUE, NOW()
		UNION SELECT 2, "string456", FALSE, STR_TO_DATE('2003-11-10 00:00:00', '%Y-%m-%d %H:%i:%s')`)
	if err != sql.ErrNoRows {
		checkErr(err)
	}
	defer rows.Close()

	data := make([][]interface{}, 0, 10)
	for rows.Next() {
		var a int
		var b string
		var c bool
		var d time.Time
		checkErr(rows.Scan(&a, &b, &c, &d))
		r := []interface{}{a, b, c, d}
		data = append(data, r)
	}

	w.Header().Set("Cache-Control", "private")
	json.NewEncoder(w).Encode(data)
}

func checkErr(err error) {
	if err != nil {
		log.Println("checkErr:", err)
		panic(err)
	}
}
