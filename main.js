"use strict";

const K2M = 1.609;

const MyStackItem = 'odoStack';
var timertick;

function oi(obj) {
    obj.classList.add('oi');
}

function findAncestor (el,cls) {
    while ((el = el.parentNode) && el.className.indexOf(cls) < 0);
    return el;
}

function findChild(el,cls) {
//    console.log("findChild "+cls+" ?");
    let kids = el.children.length;
    if (kids < 1) return null;
    for (let i = 0; i < el.children.length; i++) {
//        console.log('cl=='+JSON.stringify(el.children[i].classList));
        if (el.children[i].classList?.contains(cls)) {
          return el.children[i];
        }
        let res = findChild(el.children[i],cls);
        if (res) {
            return res;
        }
    }
    return null;
}

function getHoursMins(stiso,ftiso) {

    const minsdiv = 60000; // Number of milliseconds in a minute

    let st = new Date(stiso);
    let ft = new Date(ftiso);
    let diff = ft.getTime() - st.getTime();
    let mins = Math.floor(diff / minsdiv);
    let hrs = Math.floor(mins / 60);
    mins = mins - (hrs * 60);
    return t2(hrs)+'h'+t2(mins);
}

function oc(obj) {

    let checkio = document.getElementById('checkio');
    if (!checkio) return;

    console.log('Checking '+JSON.stringify(checkio));
    // Now we're dealing with check-in & check-out

    let line = findAncestor(obj,'odoline');

    let odostart = findChild(line,'OdoRallyStart');
    let odofinish = findChild(line,'OdoRallyFinish');
    let odokms = findChild(line,'odokms');
    let odomiles = odofinish.value - odostart.value;
    if (odokms.value == 'K') {
        odomiles = Math.floor(odomiles / K2M);
    }
    if (odomiles > 0) {
        let omobj = findChild(line,'OdoMiles');
        omobj.innerText = ""+odomiles;
    }
    console.log('Getting starttime');
    let st = document.getElementById('starttime');
    let tn = document.getElementById('timenow').getAttribute('data-time');
    console.log('tn=='+JSON.stringify(tn));
    let osobj = findChild(line,'StartTime');
    if (checkio.value == 'I') {
        console.log('Setting st');
        st = osobj.getAttribute('data-time');
        let sml = document.getElementById('SaneMilesLimit');
        if (sml) {
            if (odomiles < 0 || odomiles > sml.value) {
                console.log('Duff miles');
                odostart.disabled = false;
            }
        }
    } else {
        console.log('Getting st.value == '+JSON.stringify(osobj));
        st = st.value;
    }
    console.log('checkio=='+checkio+' st=='+st+'  tn=='+tn);
    if (osobj && osobj.getAttribute('data-time') == '' && odostart.value != '') {
        osobj.setAttribute('data-time',st);
        osobj.innerText = st.value.substring(11);
    }
    let ofobj = findChild(line,'FinishTime');
    if (ofobj && ofobj.getAttribute('data-time') == '' && odofinish.value != '') {
        ofobj.setAttribute('data-time',tn);
        ofobj.innerText = tn.substring(11);
        let omhobj = findChild(line,'HoursMins');
        omhobj.innerText = getHoursMins(st,tn);
    }

    obj.classList.remove('oi');
    obj.classList.add('oc');

    // Now update the database
    let eid = findChild(line,'EntrantID');

    let url = '';
    if (checkio.value =='I') {
        url = "/acir?eid="+eid.value+"&fod="+odofinish.value+"&sod="+odostart.value+"&omk="+odokms.value+"&fti="+tn;
    } else {
        url = "/acor?eid="+eid.value+"&sod="+odostart.value+"&omk="+odokms.value+"&sti="+st;
    }
    let newTrans = {};
    newTrans.url = url;
    newTrans.obj = obj.id;
    newTrans.sent = false;
    const stackx = sessionStorage.getItem(MyStackItem);
    let stack = [];
    if (stackx != null) 
        stack = JSON.parse(stackx);
    stack.push(newTrans);
    sessionStorage.setItem(MyStackItem,JSON.stringify(stack));

}

function sendTransactions() {

    let stackx = sessionStorage.getItem(MyStackItem);
    if (stackx == null) return;

    let stack = JSON.parse(stackx);

    let N = stack.length;

    if (N < 1) return;

    for (let i = 0; i < N; i++) {
        
        if (stack[i].sent) continue;

        console.log(stack[i].url);
        fetch(stack[i].url,{method: "POST"})
        .then(res => res.json())
        .then(function (res) {
            console.log(res.res);
            if (res.res=="ok") {
                stack[i].sent = true;
                sessionStorage.setItem(MyStackItem,JSON.stringify(stack));
                document.getElementById(stack[i].obj).classList.replace('oc','ok');
        } else {
                //showerrormsg(res.res);
            }
        });

    }

}



function clickTime() {
    let timeDisplay = document.querySelector('#timenow');
    console.log('Clicking time');
    clearInterval(timertick);
    if (timeDisplay.getAttribute('data-paused') != 0) {
        timeDisplay.setAttribute('data-paused',0);
        timertick = setInterval(refreshTime,timeDisplay.getAttribute('data-refresh'));
        timeDisplay.classList.remove('held');
    } else {
        timeDisplay.setAttribute('data-paused',1);
        timertick = setInterval(clickTime,timeDisplay.getAttribute('data-pause'));
        timeDisplay.classList.add('held');
    }
    console.log('Time clicked');
}

function refreshTime() {
    let timeDisplay = document.querySelector('#timenow');
    let dt = new Date();
    timeDisplay.setAttribute('data-time', getRallyDateTime(dt));
    let dateString = dt.toLocaleString('en-GB',{hour:"2-digit",minute:"2-digit",second:"2-digit"});
    let formattedString = dateString.replace(", ", " - ");
    timeDisplay.innerHTML = formattedString;
    checkStartSlot();
}

function nextSlotTime(timeslot) {

    let slotinterval = document.getElementById('cfgStartSlotIntervalMins').value;

    let m = parseInt(timeslot.substring(14)) + 1;
    let h = parseInt(timeslot.substring(11,13));
    let ns = Math.floor(m/slotinterval);
    ns++;
    let ms = ns * slotinterval;
    if (ms >= 60) {
        h++;
        ms -= 60;
    }
    return timeslot.substring(0,11)+t2(h)+':'+t2(ms);
}

// This keeps the starting slot in line with the current time
function checkStartSlot() {

    let st = document.getElementById('starttime');
    if (!st) return;

    let MaxOptions = document.getElementById('cfgStartSlots2Show').value;   // Generate this number of options


    let dt = new Date();    // Get the current time
    let date = st.options[0].value.substring(0,10);
    let tn = date+'T'+dt.toLocaleString('en-GB',{hour:"2-digit",minute:"2-digit"});

    // Compare the first option, not the selected one. Need to remove old options even if unused.
    if (tn <= st.options[0].value) return;

    let nextslot = nextSlotTime(tn);

    while (st.options.length != 0)
        st.options.remove(st.options.length - 1);

    let i = 0;
    while (i++ < MaxOptions) {
        let opt = document.createElement('option');
        opt.value = nextslot;
        opt.innerHTML = nextslot.substring(11,16);
        st.appendChild(opt);
        nextslot = nextSlotTime(nextslot);
    }
    st.selectedIndex = 0;
    

}
function t2(n) {
    if (n < 10)
        return '0'+n;
    return n;
}

// This formats a date/time into the format used in a ScoreMaster database
function getRallyDateTime(D) {

    let yy = D.getFullYear();
    let mt = D.getMonth() + 1;
    let dy = D.getDate();
    let hh = D.getHours();
    let mm = D.getMinutes();
    return yy+'-'+t2(mt)+'-'+t2(dy)+'T'+t2(hh)+':'+t2(mm);
}

sessionStorage.removeItem(MyStackItem);
setInterval(sendTransactions,1000);
