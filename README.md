# rblrrex RBLR1000 records maintenance

This software is designed exclusively to manage all administration, at Squires, of the 
RBLR1000 event run each June by IBAUK on behalf of the Royal British Legion Riders' Branch.

Initial signup uses the Wufoo interface used for other IBAUK events and those records are
transcribed into a ScoreMaster compatible database. Qualifying finishers are uploaded to
the IBAUK Rides database after the event.

## Records kept
### EntrantStatus
- StatusCodes["DNS"] = 0        // Signed-up on web
- StatusCodes["Registered"] = 1 // Registered at Squires
- StatusCodes["Started"] = 2    // Checked-out by staff
- StatusCodes["Finished"] = 8   // Checked-in on time
- StatusCodes["Late"] = 9       // Checked-in > 24 hours
- StatusCodes["DNF"] = 3        // Ride abandoned


## At Squires
Before riders arrive, the rally team must set up tents and other facilities and prepare logs
ready for riders to register for the ride.

### Registration
- Rider arrives and is registered as being present. May need new entry as not already signed-up.
- All rider, pillion, NoK, bike and route details confirmed or updated.
- Confirm monies already donated.
- Record new donations.
- Collect signed disclaimer.
- EntrantStatus changed to Registered.

### Check-out
- Riders gather in carpark before the off.
- Team pairs visit each bike to sign receipt logs and capture starting odos.
- EntrantStatus changed to Started.