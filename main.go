package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const apptitle = "RBLRRex v0.1"

var path2db = flag.String("db", `\sm-installs\rblr23\rex.db`, "Path to database")
var httpPort = flag.String("port", "8079", "Serve on this port")

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
	err = http.ListenAndServe(":"+*httpPort, nil)
	if err != nil {
		panic(err)
	}

}

func show_odos(w http.ResponseWriter, r *http.Request) {

	type odoParamsVar struct {
		EntrantID      int
		RiderFirst     string
		RiderLast      string
		OdoRallyStart  string
		OdoRallyFinish string
		OdoMiles       int
		OdoKms         bool
		Started        bool
		StartTime      string
		StartTimeISO   string
		Finished       bool
		FinishTime     string
		FinishTimeISO  string
		HoursMins      string
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
		<span class="blspacer"> </span>
		<span class="td timeonly" data-time="{{.StartTimeISO}}">{{if .Started}}{{.StartTime}} - {{end}}</span>
		<span class="td timeonly" data-time="{{.FinishTimeISO}}">{{if .Finished}}{{.FinishTime}} = {{end}}</span>
		<span class="td timeonly">{{if .Finished}}{{.HoursMins}}{{end}}</span>
		<span class="td">{{if .Finished}}Odo miles: {{end}}<span class="OdoMiles">{{if .Finished}}{{.OdoMiles}}{{end}}</span>
		</div>
	</div>
	`
	var cohdr = `
	<div class="topbar">
	<span class="functionlabel">CHECK-OUT/START </span>
	<span class="clock">
	</span>
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
	fmt.Fprint(w, cohdr)
	fmt.Fprint(w, `<div class="table">`)

	const validEntrantStatus = 8
	const K2M = 1.609

	var start, finish, odokms, EntrantStatus int
	for rows.Next() {
		var odoParams odoParamsVar
		rows.Scan(&odoParams.RiderFirst, &odoParams.RiderLast, &odoParams.EntrantID, &start, &finish, &odokms, &odoParams.StartTimeISO, &EntrantStatus, &odoParams.FinishTimeISO)
		if EntrantStatus != validEntrantStatus {
			continue
		}
		odoParams.OdoKms = odokms == 1
		if start > 0 {
			odoParams.OdoRallyStart = strconv.Itoa(start)
			odoParams.Started = odoParams.StartTimeISO != ""
			if odoParams.Started {
				dt := strings.Split(odoParams.StartTimeISO, "T")
				if len(dt) > 1 {
					odoParams.StartTime = dt[1]
				}
			}
		}
		if finish > 0 {
			odoParams.OdoRallyFinish = strconv.Itoa(finish)
			odoParams.Finished = true // odoParams.FinishTime != ""
			odoParams.OdoMiles = finish - start
			if odoParams.OdoKms {
				odoParams.OdoMiles = int(float64(odoParams.OdoMiles) / K2M)
			}
			dt := strings.Split(odoParams.FinishTimeISO, "T")
			if len(dt) > 1 {
				odoParams.FinishTime = dt[1]
				dtStart, _ := time.Parse("2006-01-02T15:04", odoParams.StartTimeISO)
				dtFinish, _ := time.Parse("2006-01-02T15:04", odoParams.FinishTimeISO)
				diff := dtFinish.Sub(dtStart)
				minutes := int(diff.Minutes())
				hours := int(minutes / 60)
				minutes = minutes - (hours * 60)
				odoParams.HoursMins = fmt.Sprintf("%02dh%02d", hours, minutes)

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
	<body>`

	fmt.Fprint(w, shtml)
}
