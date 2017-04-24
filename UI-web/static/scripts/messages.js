var frWeekDays = ["", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"];

var twoDigits = function (number) {
    var s = number.toString()
    return s.length === 1 ? "0" + s : s;
};

var weekdayDateStr = function (date) {
    var d = new Date(date);
    return frWeekDays[d.getDay()] + " à " + twoDigits(d.getHours()) + ":" + twoDigits(d.getMinutes()) + ":" + twoDigits(d.getSeconds());
};

var ws;

if ("WebSocket" in window) {

    // Let us open a web socket
    ws = new WebSocket("ws://localhost:8080/messages/ws");

    ws.onopen = function () {
        // Web Socket is connected
    };

    ws.onmessage = function (evt) {
        var received_msg = JSON.parse(evt.data);
        switch (received_msg.event) {
            case "message":
                var msg = JSON.parse(received_msg.payload);
                var divEvent = document.createElement("DIV");
                divEvent.className = "event";
                divEvent.innerHTML =
                    "<div class='content'>" +
                    "<div class='summary'>" +
                    "<div class='date'>" + weekdayDateStr(msg.date) + "</div>" +
                    "<a class='user'>&nbsp;" + msg.from + "&nbsp;:&nbsp;</a>" + msg.body + "</div></div>";
                f = document.getElementById("div-feed");
                f.insertBefore(divEvent, f.childNodes[0]);
                break;
            case "connected":
                console.log(received_msg);
                $(".ui.checkbox").checkbox("set checked");
                $(".ui.dimmer").dimmer("hide");
                break;
            case "disconnected":
                console.log(received_msg);
                $(".ui.checkbox").checkbox("set unchecked");
                break;
        }
    };

    ws.onclose = function () {
        // websocket is closed.
    };
}

else {
    // The browser doesn't support WebSocket
    alert("WebSocket NOT supported by your Browser!");
}

$(function () {
    $(".ui.checkbox").checkbox({
        beforeChecked: function () {
            ws.send('{"event":"connect"}');
            $(".ui.dimmer").dimmer("show");
            return false;
        },
        beforeUnchecked: function () {
            ws.send('{"event":"disconnect"}');
            return true;
        }
    });
    $("form").submit(function (event) {
        event.preventDefault();
        var evt = {
            event: "message",
            payload: $("textarea").val()
        };
        ws.send(JSON.stringify(evt));
        $("textarea").val("");
    });
});
