var frWeekDays = ["", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"];

var twoDigits = function (number) {
  var s = number.toString()
    return s.length === 1 ? "0" + s : s;
};

var weekdayDateStr = function (date) {
    var d = new Date(date);
    return frWeekDays[d.getDay()] + " à " + twoDigits(d.getHours()) + ":" + twoDigits(d.getMinutes()) + ":" + twoDigits(d.getSeconds());
};

if ("WebSocket" in window) {

    // Let us open a web socket
    var ws = new WebSocket("ws://localhost:8080/messages/ws");

    ws.onopen = function () {
        // Web Socket is connected
    };

    ws.onmessage = function (evt) {
        var received_msg = JSON.parse(evt.data);
        var divEvent = document.createElement("DIV");
        divEvent.className = "event";
        divEvent.innerHTML =
            "<div class='content'>" +
            "<div class='summary'>" +
            "<div class='date'>" + weekdayDateStr(received_msg.date) + "</div>" +
            "<a class='user'>&nbsp;" + received_msg.from + "&nbsp;:&nbsp;</a>" +
            received_msg.body +
            "</div></div>";
        f = document.getElementById("div-feed");
        f.insertBefore(divEvent, f.childNodes[0]);
    };

    ws.onclose = function () {
        // websocket is closed.
    };
}

else {
    // The browser doesn't support WebSocket
    alert("WebSocket NOT supported by your Browser!");
}
