package main

import "strconv"

type PersonRecord struct {
	Fullname   string
	IBA_Number int
	RBL_Member int
	Novice     bool
	Address    string
	Postcode   string
	Phone      string
	Email      string
}
type EntrantRecord struct {
	EntrantID     int
	Bike          string
	BikeReg       string
	OdoKms        bool
	Rider         PersonRecord
	HasPillion    bool
	Pillion       PersonRecord
	EntrantStatus int
	NokName       string
	NokPhone      string
	NokRelation   string
	StartOdo      int
	FinishOdo     int
	OdoMiles      int
	StartTime     string
	FinishTime    string
	Route         int
}

func getEntrant(EntrantID int) *EntrantRecord {

	var e EntrantRecord

	sqlx := "SELECT EntrantID, Bike, BikeReg, OdoKms"
	sqlx += ",RiderName,RiderIBA,Phone,Email,ExtraData"
	sqlx += ",PillionName,PillionIBA"
	sqlx += ",EntrantStatus,NokName,NokPhone,NokRelation"
	sqlx += ",OdoRallyStart,OdoRallyFinish,StartTime,FinishTime,Class"
	sqlx += " FROM entrants WHERE EntrantID=" + strconv.Itoa(EntrantID)

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return &e
	}
	var odokms int
	var xdata string
	rows.Scan(&e.EntrantID, &e.Bike, &e.BikeReg, &odokms, &e.Rider.Fullname,
		&e.Rider.IBA_Number, &e.Rider.Phone, &e.Rider.Email, &xdata,
		&e.Pillion.Fullname, &e.Pillion.IBA_Number,
		&e.EntrantStatus, &e.NokName, &e.NokPhone, &e.NokRelation,
		&e.StartOdo, &e.FinishOdo, &e.StartTime, &e.FinishTime, &e.Route,
	)
	e.HasPillion = e.Pillion.Fullname != ""
	e.OdoKms = odokms == 1
	return &e
}
