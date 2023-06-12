"use strict";

const K2M = 1.609;

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