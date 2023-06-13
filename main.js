"use strict";

const K2M = 1.609;

var timertick;

function oi(obj) {
    obj.classList.add('oi');
}

function findAncestor (el,cls) {
    while ((el = el.parentNode) && el.className.indexOf(cls) < 0);
    return el;
}

function findChild(el,cls) {
    let kids = el.children.length;
    if (kids < 1) return null;
    for (let i = 0; i < el.children.length; i++) {
        if (el.children[i].classList.contains(cls)) {
          return el.children[i];
        }
        let res = findChild(el.children[i],cls);
        if (res) {
            return res;
        }
    }
    return null;
}

function oc(obj) {

    let line = findAncestor(obj,'odoline');
    let odostart = findChild(line,'OdoRallyStart');
    let odofinish = findChild(line,'OdoRallyFinish');
    let odokms = findChild(line,'odokms');
    let odomiles = odofinish.value - odostart.value;
    if (odokms.value == 'K') {
        odomiles = Math.floor(odomiles / K2M);
    }
    let omobj = findChild(line,'OdoMiles');
    omobj.innerText = ""+odomiles;
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

    const slotinterval = 10;

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

function checkStartSlot() {


    const MaxOptions = 5;

    let st = document.getElementById('starttime');
    if (!st) return;

    let dt = new Date();
    //let tn = dt.toISOString().substring(0,16);
    let tn = st.value.substring(0,10)+'T'+dt.toLocaleString('en-GB',{hour:"2-digit",minute:"2-digit"});
    console.log('tn is '+tn);
    let oldslot = new Date(st.value);
    console.log('old slot was '+oldslot.toISOString());
    if (tn <= st.value) return;
    console.log('Timenow is '+tn);
    let nextslot = nextSlotTime(tn);

    while (st.options.length != 0)
        st.options.remove(st.options.length - 1);

    let i = 0;
    while (i++ <= MaxOptions) {
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
function getRallyDateTime(D) {

    let yy = D.getFullYear();
    let mt = D.getMonth() + 1;
    let dy = D.getDate();
    let hh = D.getHours();
    let mm = D.getMinutes();
    return yy+'-'+t2(mt)+'-'+t2(dy)+'T'+t2(hh)+':'+t2(mm);
}
