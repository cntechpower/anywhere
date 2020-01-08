const app = document.getElementById('root')

var table = document.createElement("table");
table.style.width = '50%';
table.setAttribute('border', '1');
table.setAttribute('cellspacing', '0');
table.setAttribute('cellpadding', '5');
table.setAttribute('class','fl-table');
var col = ["agent_id", "local_addr", "remote_addr", "is_whitelist_on", "whitelist_ips"];
// CREATE TABLE HEAD .
var tHead = document.createElement("thead");


// CREATE ROW FOR TABLE HEAD .
var hRow = document.createElement("tr");

// ADD COLUMN HEADER TO ROW OF TABLE HEAD.
for (var i = 0; i < col.length; i++) {
    var th = document.createElement("th");
    th.innerHTML = col[i];
    hRow.appendChild(th);
}
tHead.appendChild(hRow);
table.appendChild(tHead);

var tBody = document.createElement("tbody");


var request = new XMLHttpRequest()
request.open('GET', 'http://127.0.0.1:1112/api/v1/proxy/list', true)
request.onload = function () {
    // Begin accessing JSON data here
    var data = JSON.parse(this.response)
    if (request.status >= 200 && request.status < 400) {
        data.forEach(config => {
            var bRow = document.createElement("tr");
            var td = document.createElement("td");
            td.innerHTML = config.agent_id;
            bRow.appendChild(td);
            var td = document.createElement("td");
            td.innerHTML = config.remote_addr;
            bRow.appendChild(td);
            var td = document.createElement("td");
            td.innerHTML = config.local_addr;
            bRow.appendChild(td);
            var td = document.createElement("td");
            td.innerHTML = config.is_whitelist_on;
            bRow.appendChild(td);
            var td = document.createElement("td");
            td.innerHTML = config.whitelist_ips;
            bRow.appendChild(td);

            tBody.appendChild(bRow)
        })
        table.appendChild(tBody);
        app.appendChild(table)
    } else {
        const errorMessage = document.createElement('marquee')
        errorMessage.textContent = `Gah, it's not working!`
        app.appendChild(errorMessage)
    }
}

request.send()
