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
	yaml "gopkg.in/yaml.v2"
)

const apptitle = "RBLRRex v0.1"

var path2db = flag.String("db", `\sm-installs\rblr23\rex.db`, "Path to database")
var httpPort = flag.String("port", "8079", "Serve on this port")

const baseConfig = `

#######################       DO NOT USE TABS ANYWHERE IN THIS YAML DOCUMENT !!!

StartSlotIntervalMins: 10

StartSlots2Show: 3

PauseClockMins: 2

# If the stats screen is showing, it'll auto-refresh after this number of seconds
RefreshStatsIntervalSecs: 60

# If odo readings give more than this value, one of them is probably wrong
SaneMilesLimit: 1600 


# If you mess with the keys used here, you'll need to fix them elsewhere
# Code values 0, 1, 3 & 8 have specific meanings in ScoreMaster
StatusCodes:
  DNS:           {code: 0, title: 'Signed-up online, not seen at Squires'}
  Registered:    {code: 1, title: 'Registered at Squires'}
  CheckedOut:    {code: 2, title: 'Odo read, now out riding'}
  CheckedIn:     {code: 4, title: 'Odo read, now checking receipts'}
  Finisher:      {code: 8, title: 'Verified, within 24 hours'}
  Certificate:   {code: 9, title: 'Verified, > 24 hours'}
  DNF:           {code: 3, title: 'Ride abandoned, not returning'}


CheckinStatusCodes: [CheckedOut,CheckedIn]

CheckoutStatusCodes: [DNS,Registered,CheckedOut]


# Don't update the start/finish times for any of these statuses
DontRestartCodes: [CheckedOut,CheckedIn,Finisher,Certificate,DNF]

# Present statistics for these statuses, in the order listed her
RiderStatsList: [Registered,CheckedOut,CheckedIn,Finisher,Certificate,DNF,DNS]

`

type StatusDetails struct {
	Code  int    `yaml:"code"`
	Title string `yaml:"title"`
}

var CFG struct {
	StartSlotIntervalMins    int                      `yaml:"StartSlotIntervalMins"`
	StartSlots2Show          int                      `yaml:"StartSlots2Show"`
	PauseClockMins           int                      `yaml:"PauseClockMins"`
	CheckinStatusCodes       []string                 `yaml:"CheckinStatusCodes"`
	CheckoutStatusCodes      []string                 `yaml:"CheckoutStatusCodes"`
	DontRestartCodes         []string                 `yaml:"DontRestartCodes"`
	SaneMilesLimit           int                      `yaml:"SaneMilesLimit"`
	RiderStatsList           []string                 `yaml:"RiderStatsList"`
	StatusCodes              map[string]StatusDetails `yaml:"StatusCodes"`
	RefreshStatsIntervalSecs int                      `yaml:"RefreshStatsIntervalSecs"`
}
var DBH *sql.DB

func init() {

	file := strings.NewReader(baseConfig)
	D := yaml.NewDecoder(file)
	D.Decode(&CFG)

}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func DBExec(sqlx string) sql.Result {

	res, err := DBH.Exec(sqlx)
	if err != nil {
		fmt.Printf("DBExec = %v\n", sqlx)
		panic(err)
	}
	return res

}

func ajax_checkinRider(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	entrantid := r.FormValue("eid")
	finishodo := r.FormValue("fod")
	startodo := r.FormValue("sod")
	odounit := r.FormValue("omk")
	finishtime := r.FormValue("fti")
	if entrantid == "" || odounit == "" || finishtime == "" {
		fmt.Fprintf(w, `{"res":"Blank field"`)
		fmt.Fprintf(w, `,"entrantid":"%v"`, entrantid)
		fmt.Fprintf(w, `,"finishodo":"%v"`, finishodo)
		fmt.Fprintf(w, `,"odounit":"%v"`, odounit)
		fmt.Fprintf(w, `,"finishtime":"%v"`, finishtime)
		fmt.Fprint(w, `}`)
		return
	}
	odokms := "0"
	if odounit == "K" {
		odokms = "1"
	}
	var nrex int64 = 0
	var err error
	if finishodo != "" {
		sqlx := "UPDATE entrants SET EntrantStatus=" + strconv.Itoa(CFG.StatusCodes["CheckedIn"].Code)
		sqlx += ",OdoRallyStart=" + startodo
		sqlx += ",OdoRallyFinish=" + finishodo
		sqlx += ",OdoKms=" + odokms
		sqlx += ",FinishTime='" + finishtime + "'"
		sqlx += " WHERE EntrantID=" + entrantid
		if len(CFG.DontRestartCodes) > 0 {
			sqlx += " AND (EntrantStatus NOT IN ("
			x := ""
			for _, sc := range CFG.DontRestartCodes {
				if x != "" {
					x += ","
				}
				x += strconv.Itoa(CFG.StatusCodes[sc].Code)
			}
			sqlx += x + ")"
			sqlx += " OR FinishTime IS NULL OR FinishTime='')"
		}
		res := DBExec(sqlx)
		nrex, err = res.RowsAffected()
		checkerr(err)
	} else {
		nrex = 0
	}

	if nrex < 1 {
		sqlx := "UPDATE entrants SET OdoKms=" + odokms
		if finishodo != "" {
			sqlx += ",OdoRallyFinish=" + finishodo
			sqlx += ",OdoRallyStart=" + startodo
			sqlx += ",EntrantStatus=" + strconv.Itoa(CFG.StatusCodes["CheckedIn"].Code)
		}
		sqlx += " WHERE EntrantID=" + entrantid
		res := DBExec(sqlx)
		n, err := res.RowsAffected()
		checkerr(err)
		if n < 1 {
			fmt.Fprint(w, `{"res":"Database operation failed!"}`)
			return
		}
	}
	fmt.Fprint(w, `{"res":"ok"}`)
}

func ajax_checkoutRider(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	entrantid := r.FormValue("eid")
	startodo := r.FormValue("sod")
	odounit := r.FormValue("omk")
	starttime := r.FormValue("sti")
	if entrantid == "" || odounit == "" || starttime == "" {
		fmt.Fprintf(w, `{"res":"Blank field"`)
		fmt.Fprintf(w, `,"entrantid":"%v"`, entrantid)
		fmt.Fprintf(w, `,"startodo":"%v"`, startodo)
		fmt.Fprintf(w, `,"odounit":"%v"`, odounit)
		fmt.Fprintf(w, `,"starttime":"%v"`, starttime)
		fmt.Fprint(w, `}`)
		return
	}
	odokms := "0"
	if odounit == "K" {
		odokms = "1"
	}
	var nrex int64 = 0
	var err error
	if startodo != "" {
		sqlx := "UPDATE entrants SET EntrantStatus=" + strconv.Itoa(CFG.StatusCodes["Started"].Code)
		sqlx += ",OdoRallyStart=" + startodo
		sqlx += ",OdoKms=" + odokms
		sqlx += ",StartTime='" + starttime + "'"
		sqlx += " WHERE EntrantID=" + entrantid
		if len(CFG.DontRestartCodes) > 0 {
			sqlx += " AND (EntrantStatus NOT IN ("
			x := ""
			for _, sc := range CFG.DontRestartCodes {
				if x != "" {
					x += ","
				}
				x += strconv.Itoa(CFG.StatusCodes[sc].Code)
			}
			sqlx += x + ")"
			sqlx += " OR StartTime IS NULL OR StartTime='')"
		}
		res := DBExec(sqlx)
		nrex, err = res.RowsAffected()
		checkerr(err)
	} else {
		nrex = 0
	}

	if nrex < 1 {
		sqlx := "UPDATE entrants SET OdoKms=" + odokms
		if startodo != "" {
			sqlx += ",OdoRallyStart=" + startodo
			sqlx += ",EntrantStatus=" + strconv.Itoa(CFG.StatusCodes["Started"].Code)
		}
		sqlx += " WHERE EntrantID=" + entrantid
		res := DBExec(sqlx)
		n, err := res.RowsAffected()
		checkerr(err)
		if n < 1 {
			fmt.Fprint(w, `{"res":"Database operation failed!"}`)
			return
		}
	}
	fmt.Fprint(w, `{"res":"ok"}`)
}

func getIntegerFromDB(sqlx string, defx int) int {

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		var val int
		rows.Scan(&val)
		return val
	}
	return defx

}
func getStringFromDB(sqlx string, defx string) string {
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		var val string
		rows.Scan(&val)
		return val
	}
	return defx

}
func show_stats(w http.ResponseWriter, r *http.Request) {

	var statshdr = `
	<div class="topbar">
	<header><a href="/menu">&#9776;</a>
	<span class="functionlabel">CURRENT STATUS </span>
	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	<span id="timenow" class="smalltime" data-time="" data-refresh="1000"  data-paused="0" ></span>
	</header>
	</div>
	`
	var eventhdr = `
	<div class="RiderStats">
	<span class="header">%v - %v</span>
	</div>
	`

	start_html(w, r)
	fmt.Fprint(w, statshdr)
	event := getStringFromDB("SELECT RallyTitle FROM rallyparams", "RBLR1000")
	start := getStringFromDB("SELECT StartTime FROM rallyparams", "Today")
	fmt.Fprintf(w, eventhdr, event, start[0:4])
	fmt.Fprint(w, `<table class="RiderStats">`)

	for _, sc := range CFG.RiderStatsList {

		nRex := getIntegerFromDB("SELECT count(*) FROM entrants WHERE EntrantStatus="+strconv.Itoa(CFG.StatusCodes[sc].Code), 0)
		fmt.Fprintf(w, `<tr><td  title="%v" class="Status">%v</td><td class="Count">%v</td></tr>`, CFG.StatusCodes[sc].Title, sc, nRex)
	}
	fmt.Fprint(w, `</table>`)

	fmt.Fprint(w, `<script>refreshTime(); timertick = setInterval(refreshTime,1000);`)
	fmt.Fprintf(w, `function refreshPage(){let url='%v';window.location.href=url;}setInterval(refreshPage,%v);`, "/stats", CFG.RefreshStatsIntervalSecs*1000)
	fmt.Fprint(w, `</script>`)

}

func show_entrant_register(w http.ResponseWriter, r *http.Request) {

	var entrantform = `
	<div class="registerentrant">
	{{.EntrantID}} {{.Rider.Fullname}} {{.Bike}}
	</div>
	`
	r.ParseForm()
	entrantid := r.FormValue("eid")
	if entrantid == "" {
		return
	}
	en, _ := strconv.Atoi(entrantid)
	e := getEntrant(en)

	start_html(w, r)

	t, err := template.New("entrantform").Parse(entrantform)
	if err != nil {
		panic(err)
	}
	t.Execute(w, e)

}

func main() {

	var err error
	flag.Parse()

	DBH, err = sql.Open("sqlite3", *path2db)
	if err != nil {
		fmt.Printf("%v: Can't access database %v [%v] run aborted\n", apptitle, path2db, err)
		os.Exit(1)
	}
	defer DBH.Close()

	http.HandleFunc("/co", show_checkout)
	http.HandleFunc("/ci", show_checkin)
	http.HandleFunc("/acir", ajax_checkinRider)
	http.HandleFunc("/acor", ajax_checkoutRider)
	http.HandleFunc("/menu", show_menu)
	http.HandleFunc("/", show_menu)
	http.HandleFunc("/stats", show_stats)
	http.HandleFunc("/sent", show_entrant_register)
	err = http.ListenAndServe(":"+*httpPort, nil)
	if err != nil {
		panic(err)
	}

}

func next_slot(timeslot string) string {

	m, _ := strconv.Atoi(timeslot[14:])
	m++
	h, _ := strconv.Atoi(timeslot[11:13])
	ns := m / CFG.StartSlotIntervalMins
	ns++
	ms := ns * CFG.StartSlotIntervalMins
	if ms >= 60 {
		h++
		ms -= 60
	}
	return fmt.Sprintf("%v%02d:%02d", timeslot[0:11], h, ms)

}

func show_menu(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)
	fmt.Fprint(w, `<header><a href="/menu">&#9776;</a></header>`)
	fmt.Fprint(w, `<main>`)
	fmt.Fprint(w, `<ul>`)
	fmt.Fprint(w, `<li><a href="/co">Check-out</a></li>`)
	fmt.Fprint(w, `<li><a href="/ci">Check-in</a></li>`)
	fmt.Fprint(w, `<li><a href="/stats">Status</a></li>`)
	fmt.Fprint(w, `</ul>`)
	fmt.Fprint(w, `</main>`)
}
func show_checkin(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DEBUG: show_checkin")
	var show_codes []int
	for i := 0; i < len(CFG.CheckinStatusCodes); i++ {
		show_codes = append(show_codes, CFG.StatusCodes[CFG.CheckinStatusCodes[i]].Code)
	}
	show_odos(w, r, false, show_codes)
}
func show_checkout(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DEBUG: show_checkout")
	var show_codes []int
	for i := 0; i < len(CFG.CheckoutStatusCodes); i++ {
		show_codes = append(show_codes, CFG.StatusCodes[CFG.CheckoutStatusCodes[i]].Code)
	}
	show_odos(w, r, true, show_codes)

}
func show_odos(w http.ResponseWriter, r *http.Request, check_out bool, show_status []int) {

	fmt.Printf("DEBUG: show_odos out=%v, codes=%v\n", check_out, show_status)
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
		FinishReadOnly bool
		StartReadOnly  bool
		EntrantStatus  int
	}
	const odoline = `
	<div class="odoline">
		<div class="topline">
			<input type="hidden" class="td EntrantID hide" value="{{.EntrantID}}">
			<span class="td RiderName"><strong>{{.RiderLast}}</strong>, {{.RiderFirst}}</span>
			<span class="td odo">
				<input type="number" {{if .StartReadOnly}}disabled {{end}}class="bignumber OdoRallyStart" placeholder="start" value="{{.OdoRallyStart}}" oninput="oi(this);" onchange="oc(this);" id="sod{{.EntrantID}}">
			</span>
			<span class="td odo">
				<input type="number" {{if .FinishReadOnly}}disabled {{end}}class="bignumber OdoRallyFinish" placeholder="finish" value="{{.OdoRallyFinish}}" oninput="oi(this);" onchange="oc(this);" id="fod{{.EntrantID}}">
			</span>
			<span class="td mk">
				<select class="odokms" onchange="oc(this);" id="omk{{.EntrantID}}">
					<option value="M"{{if .OdoKms}}{{else}} selected{{end}}>M</option>
					<option value="K"{{if .OdoKms}} selected{{end}}>K</option>
				</select>
			</span>
		</div>
		
		<div class="bottomline">
		<span class="blspacer"> </span>
		<span class="td timeonly StartTime" data-time="{{.StartTimeISO}}">{{if .Started}}{{.StartTime}}{{end}}</span>
		<span class="td timeonly FinishTime" data-time="{{.FinishTimeISO}}">{{if .Finished}}{{.FinishTime}}{{end}}</span>
		<span class="td timeonly HoursMins">{{if .Finished}}{{.HoursMins}}{{end}}</span>
		<span class="td"><span class="OdoMiles">{{if .Finished}}{{.OdoMiles}}{{end}}</span>
		</div>
	</div>
	`

	msecs2pause := CFG.PauseClockMins * 60000

	rows, err := DBH.Query("SELECT StartTime FROM rallyparams")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	startingslot := time.Now().Format("2006-01-02T15:04")
	if rows.Next() {
		rows.Scan(&startingslot)
	}
	rows.Close()

	var cihdr = `
	<input type="hidden" id="checkio" value="I">
	<input type="hidden" id="SaneMilesLimit" value="` + strconv.Itoa(CFG.SaneMilesLimit) + `">
	<div class="topbar">
	<header><a href="/menu">&#9776;</a>
	<span class="functionlabel">CHECK-IN/FINISH </span>
	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	<span id="timenow" class="smalltime" data-time="" data-refresh="1000" data-pause="` + strconv.Itoa(msecs2pause) + `" data-paused="0" onclick="clickTime();">

	</span>
	</header>
	</div>
	`

	var cohdr = `
	<input type="hidden" id="checkio" value="O">
	<input type="hidden" id="SaneMilesLimit" value="` + strconv.Itoa(CFG.SaneMilesLimit) + `">
	<div class="topbar">
	<header><a href="/menu">&#9776;</a>
	<span class="functionlabel">CHECK-OUT/START </span>
	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	<span class="starttime">
		<select id="starttime" class="st">
			##OPTS##
		</select>
	</span>
	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	<span id="timenow" class="smalltime" data-time="" data-refresh="1000" data-pause="` + strconv.Itoa(msecs2pause) + `" data-paused="0" onclick="clickTime();">

	</span>
	</header>
	</div>
	`
	opts := ""
	sel := " selected"
	for i := 0; i < CFG.StartSlots2Show; i++ {
		opts += `<option value="` + startingslot + `"` + sel + `>` + startingslot[11:16] + `</option>`
		sel = ""
		startingslot = next_slot(startingslot)
	}
	cohdr = strings.ReplaceAll(cohdr, "##OPTS##", opts)

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
	rows, err = DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	start_html(w, r)
	if check_out {
		fmt.Fprint(w, cohdr)
	} else {
		fmt.Fprint(w, cihdr)
	}
	fmt.Fprint(w, `<div class="table">`)

	const K2M = 1.609

	var start, finish, odokms int
	for rows.Next() {
		var odoParams odoParamsVar
		odoParams.FinishReadOnly = check_out
		odoParams.StartReadOnly = !check_out
		rows.Scan(&odoParams.RiderFirst, &odoParams.RiderLast, &odoParams.EntrantID, &start, &finish, &odokms, &odoParams.StartTimeISO, &odoParams.EntrantStatus, &odoParams.FinishTimeISO)
		ok := false
		for i := 0; i < len(show_status); i++ {
			if odoParams.EntrantStatus == show_status[i] {
				ok = true
				break
			}
		}
		if odoParams.EntrantID == 127 {
			fmt.Printf("DEBUG: Entrant %v (%v %v) has status %v and ok=%v\n", odoParams.EntrantID, odoParams.RiderFirst, odoParams.RiderLast, odoParams.EntrantStatus, ok)
		}
		if !ok {
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
	fmt.Fprint(w, `<script>refreshTime(); timertick = setInterval(refreshTime,1000);</script>`)

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
	<input type="hidden" id="cfgStartSlotIntervalMins" value="` + strconv.Itoa(CFG.StartSlotIntervalMins) + `">
	<input type="hidden" id="cfgStartSlots2Show" value="` + strconv.Itoa(CFG.StartSlots2Show) + `">
	<input type="hidden" id="cfgPauseClockMins" value="` + strconv.Itoa(CFG.PauseClockMins) + `">
	`

	fmt.Fprint(w, shtml)
}
