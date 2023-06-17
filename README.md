# rblrrex RBLR1000 records maintenance

This software is designed exclusively to manage all administration, at Squires, of the 
RBLR1000 event run each June by IBAUK on behalf of the Royal British Legion Riders' Branch.

Initial signup uses the Wufoo interface used for other IBAUK events and those records are
transcribed into a ScoreMaster compatible database. Qualifying finishers are uploaded to
the IBAUK Rides database after the event.

## Records kept
### EntrantStatus
-  DNS:           {code: 0, title: 'Signed-up online, not seen at Squires'}
-  Registered:    {code: 1, title: 'Registered at Squires'}
-  CheckedOut:    {code: 2, title: 'Odo read, now out riding'}
-  CheckedIn:     {code: 4, title: 'Odo read, now checking receipts'}
-  Finisher:      {code: 8, title: 'Verified, within 24 hours'}
-  Certificate:   {code: 9, title: 'Verified, > 24 hours'}
-  DNF:           {code: 3, title: 'Ride abandoned, not returning'}


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
- If any details altered, certificate is pulled and marked for posting

### Check-out
- Riders gather in carpark before the off.
- Team pairs visit each bike to sign receipt logs and capture starting odos.
- EntrantStatus changed to CheckedOut.

### Check-in
- Riders return to Squires
- Team pairs visit each bike to sign receipt logs and capture final odos.
- EntrantStatus changed to CheckedIn.

### Verification
- Riders sort out their paperwork.
- Verifiers check receipts.
- EntrantStatus changed to Finisher or Certificate.
- Finishers receive unchanged certificate immediately.