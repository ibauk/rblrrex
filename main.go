package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const apptitle = "RBLRRex v0.1"

var path2db = flag.String("db", `\sm-installs\rblr23\rex.db`, "Path to database")

const baseConfig = `
StartSlotIntervalMins: 10
#`

var CFG struct {
	StartSlotIntervalMins int `yaml:"StartSlotIntervalMins"`
}
var DBH *sql.DB

func main() {

	var err error
	flag.Parse()
	DBH, err = sql.Open("sqlite3", *path2db)
	if err != nil {
		fmt.Printf("%v: Can't access database %v [%v] run aborted\n", apptitle, path2db, err)
		os.Exit(1)
	}
	defer DBH.Close()
	fmt.Println(baseConfig)

	http.HandleFunc("/", show_odos)
	err = http.ListenAndServe("127.0.0.1:8079", nil)
	if err != nil {
		panic(err)
	}

}

func show_odos(w http.ResponseWriter, r *http.Request) {

	var odoParams struct {
		EntrantID      int
		RiderFirst     string
		RiderLast      string
		OdoRallyStart  string
		OdoRallyFinish string
		OdoMiles       int
		OdoKms         bool
		Started        bool
		StartTime      string
		Finished       bool
		FinishTime     string
	}
	const odoline = `
	<div class="odoline">
		<div class="topline">
			<span class="td EntrantID">{{.EntrantID}}</span>
			<span class="td RiderName"><strong>{{.RiderLast}}</strong>, {{.RiderFirst}}</span>
			<span class="td odo">
				<input type="number" class="bignumber OdoRallyStart" placeholder="start" value="{{.OdoRallyStart}}" oninput="oi(this);" onchange="oc(this);">
			</span>
			<span class="td odo">
				<input type="number" class="bignumber OdoRallyFinish" placeholder="finish" value="{{.OdoRallyFinish}}" oninput="oi(this);" onchange="oc(this);">
			</span>
			<span class="td mk">
				<select class="odokms" onchange="oc(this);">
					<option value="M"{{if .OdoKms}}{{else}} selected{{end}}>M</option>
					<option value="K"{{if .OdoKms}} selected{{end}}>K</option>
				</select>
			</span>
		</div>
		
		<div class="bottomline">
		<span class="td">{{if .Finished}}Odo miles: {{end}}<span class="OdoMiles">{{if .Finished}}{{.OdoMiles}}{{end}}</span>
		<span class="td">{{if .Started}}Start: {{.StartTime}}{{end}}</span>
		<span class="td">{{if .Finished}}Finish: {{.FinishTime}}{{end}}</span>
		</div>
	</div>
	`
	sqlx := "SELECT trim(substr(RiderName,1,RiderPos-1)) AS RiderFirst"
	sqlx += ",trim(substr(RiderName,RiderPos+1)) AS RiderLast"
	sqlx += ",EntrantID,ifnull(OdoRallyStart,0),ifnull(OdoRallyFinish,0),OdoKms"
	sqlx += ",ifnull(StartTime,''),EntrantStatus,ifnull(FinishTime,'')"
	sqlx += " FROM (SELECT *,instr(RiderName,' ') AS RiderPos FROM entrants) "
	sqlx += " ORDER BY upper(RiderLast),upper(RiderFirst)"

	t, err := template.New("odoline").Parse(odoline)
	if err != nil {
		panic(err)
	}
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	start_html(w, r)
	fmt.Fprint(w, `<div class="table">`)

	const validEntrantStatus = 8
	const K2M = 1.609

	var start, finish, odokms, EntrantStatus int
	for rows.Next() {
		rows.Scan(&odoParams.RiderFirst, &odoParams.RiderLast, &odoParams.EntrantID, &start, &finish, &odokms, &odoParams.StartTime, &EntrantStatus, &odoParams.FinishTime)
		if EntrantStatus != validEntrantStatus {
			continue
		}
		odoParams.OdoKms = odokms == 1
		odoParams.OdoRallyStart = ""
		odoParams.OdoRallyFinish = ""
		odoParams.Started = false
		odoParams.Finished = false
		odoParams.OdoMiles = 0
		if start > 0 {
			odoParams.OdoRallyStart = strconv.Itoa(start)
			odoParams.Started = odoParams.StartTime != ""
		}
		if finish > 0 {
			odoParams.OdoRallyFinish = strconv.Itoa(finish)
			odoParams.Finished = true // odoParams.FinishTime != ""
			odoParams.OdoMiles = finish - start
			if odoParams.OdoKms {
				odoParams.OdoMiles = int(float64(odoParams.OdoMiles) / K2M)
			}
		}

		err := t.Execute(w, odoParams)
		if err != nil {
			panic(err)
		}

	}

	fmt.Fprint(w, `</div>`)

}

//go:embed reboot.css
var css_reboot string

//go:embed main.css
var css_main string

//go:embed main.js
var js_main string

func start_html(w http.ResponseWriter, r *http.Request) {

	var shtml = `
	<!DOCTYPE html>
	<html>
	<head>
	<title>rblrrex</title>
	<style>
	` + css_reboot + css_main + `
	</style>
	<script>` + js_main + `</script>
	</head>
	<body>
	`
	fmt.Fprint(w, shtml)
}
