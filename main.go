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
StartSlotIntervalMins: 10
StartSlots2Show: 3
PauseClockMins: 2

CheckoutStatusCodes: [DNS,Registered,Started,Finished]
DontRestartCodes: [Started,Finished,DNF]
`

// Entrant status codes, ScoreMaster plus extras
var StatusCodes map[string]int

var CFG struct {
	StartSlotIntervalMins int      `yaml:"StartSlotIntervalMins"`
	StartSlots2Show       int      `yaml:"StartSlots2Show"`
	PauseClockMins        int      `yaml:"PauseClockMins"`
	CheckoutStatusCodes   []string `yaml:"CheckoutStatusCodes"`
	DontRestartCodes      []string `yaml:"DontRestartCodes"`
}
var DBH *sql.DB

func init() {

	file := strings.NewReader(baseConfig)
	D := yaml.NewDecoder(file)
	D.Decode(&CFG)

	StatusCodes = make(map[string]int)

	StatusCodes["DNS"] = 0        // Signed-up on web
	StatusCodes["Registered"] = 1 // Registered at Squires
	StatusCodes["Started"] = 2    // Checked-out by staff
	StatusCodes["Finished"] = 8   // Checked-in on time
	StatusCodes["Late"] = 9       // Checked-in > 24 hours
	StatusCodes["DNF"] = 3        // Ride abandoned

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
		sqlx := "UPDATE entrants SET EntrantStatus=" + strconv.Itoa(StatusCodes["Started"])
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
				x += strconv.Itoa(StatusCodes[sc])
			}
			sqlx += x + ")"
			sqlx += " OR StartTime IS NULL)"
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
			sqlx += ",EntrantStatus=" + strconv.Itoa(StatusCodes["Started"])
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

func main() {

	var err error
	flag.Parse()

	DBH, err = sql.Open("sqlite3", *path2db)
	if err != nil {
		fmt.Printf("%v: Can't access database %v [%v] run aborted\n", apptitle, path2db, err)
		os.Exit(1)
	}
	defer DBH.Close()

	http.HandleFunc("/", show_checkout)
	http.HandleFunc("/acor", ajax_checkoutRider)
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

func show_checkout(w http.ResponseWriter, r *http.Request) {

	var show_codes []int
	for i := 0; i < len(CFG.CheckoutStatusCodes); i++ {
		show_codes = append(show_codes, StatusCodes[CFG.CheckoutStatusCodes[i]])
	}
	show_odos(w, r, true, show_codes)

}
func show_odos(w http.ResponseWriter, r *http.Request, check_out bool, show_status []int) {

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
		<span class="td timeonly">{{if .Finished}}{{.HoursMins}}{{end}}</span>
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

	var cohdr = `
	<div class="topbar">
	<header>
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
	}
	fmt.Fprint(w, `<div class="table">`)

	const K2M = 1.609

	var start, finish, odokms, EntrantStatus int
	for rows.Next() {
		var odoParams odoParamsVar
		odoParams.FinishReadOnly = check_out
		odoParams.StartReadOnly = !check_out
		rows.Scan(&odoParams.RiderFirst, &odoParams.RiderLast, &odoParams.EntrantID, &start, &finish, &odokms, &odoParams.StartTimeISO, &EntrantStatus, &odoParams.FinishTimeISO)
		ok := false
		for i := 0; i < len(show_status); i++ {
			if EntrantStatus == show_status[i] {
				ok = true
				break
			}
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
