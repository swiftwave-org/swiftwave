// Utility
function hideStatus() {
    document.getElementById('status_text').style.display = 'none';
    document.getElementById('loader_container').style.display = 'none';
}

function showStatus(text) {
    document.getElementById('status_text').getElementsByTagName("span")[0].innerText = text;
    document.getElementById('status_text').style.display = 'block';
    document.getElementById('loader_container').style.display = 'flex';
}


// Initiate terminal
const term = new Terminal({
    cursorBlink: true,
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
fitAddon.fit();

// Handle copy from terminal
term.attachCustomKeyEventHandler(function (e) {
    // Ctrl + Shift + C
    if (e.ctrlKey && e.shiftKey && (e.keyCode === 67)) {
        e.preventDefault()
        navigator.clipboard.writeText(term.getSelection()).catch(() => console.log(e.message));
        return false;
    }
});

// Terminal SSH initialization
async function init() {
    // check if not on localhost and not secure
    if ((location.hostname !== "localhost" && location.hostname !== "127.0.0.1" && location.hostname.endsWith(".local")) && location.protocol !== "https:") {
        showStatus("Please use a secure connection (https)");
        return;
    }
    const urlParams = new URLSearchParams(window.location.search);

    let data = {}
    if (urlParams.has('server')) {
        // fetch server id
        const serverId = urlParams.get('server');
        showStatus("Authenticating...");
        // generate a console request
        const response = await fetch(`/console/token/server/${serverId}`, {
            method: "POST", headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            showStatus("Error: " + response.statusText);
            return;
        }
        data = await response.json();
    } else if (urlParams.has('application')) {
        // fetch application id
        const applicationId = urlParams.get('application');
        showStatus("Fetching available servers...");
        // find servers
        const response = await fetch(`/console/application/${applicationId}/servers`, {
            method: "GET", headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            showStatus("Error: " + response.statusText);
            return;
        }
        const servers = await response.json();
        if (servers.length === 0) {
            showStatus("Error: No servers available");
            return;
        }
        showStatus("Select preferred server to connect for application console");
        const serverList = Object.keys(servers);
        let serverListString = "Available Servers:\n";
        for (let i = 0; i < serverList.length; i++) {
            serverListString += `[${i + 1}] ${serverList[i]}\n`
        }
        serverListString = serverListString + "[x] Cancel"
        serverListString = serverListString + "\n\nEnter serial no of server :";
        let serverSrNo = 0;
        while (serverSrNo < 1 || serverSrNo > serverList.length) {
            serverSrNo = prompt(serverListString);
            if (!serverSrNo) {
                showStatus("Error: No server selected");
            }
            if (serverSrNo === "x") {
                showStatus("Aborted ! Refresh to try again")
                return;
            }
        }
        // send request to server
        showStatus("Authenticating...");
        // /console/token/application/:id/:server_id
        const response2 = await fetch(`/console/token/application/${applicationId}/${servers[serverList[parseInt(serverSrNo) - 1]]}`, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response2.ok) {
            showStatus("Error: " + response2.statusText);
            return;
        }
        data = await response2.json();
    } else {
        showStatus("Console Info not found");
        return
    }

    console.log(data)

    // fetch request_id and token
    const requestId = data.request_id;
    const token = data.token;
    const target = "Will be added !" // TODO
    if (!requestId || !token) {
        showStatus("Error: Some error occurred");
        return;
    }
    document.title = `[Console] ${target}`;

    // connect to websocket using the request_id and token
    let protocol = "ws";
    if (location.protocol === "https:") {
        protocol = "wss";
    }
    showStatus("Connecting to server...");
    const ws = new WebSocket(`${protocol}://${location.host}/console/ws/${requestId}/${token}/${term.rows}/${term.cols}`);
    ws.binaryType = "arraybuffer";

    // error handler
    ws.onerror = function (e) {
        showStatus("Error: " + e.message);
    }
    ws.onopen = function () {
        hideStatus();
        // handle data received from server
        ws.onmessage = function (evt) {
            if (evt.data instanceof ArrayBuffer) {
                term.write(new Uint8Array(evt.data))
            } else {
                console.log("invalid data received")
            }
        }
        // handle data sent from terminal
        term.onData((e) => ws.send(new TextEncoder().encode(e)))
        // handle terminal resize
        window.addEventListener('resize', () => {
            fitAddon.fit();
            const payload = {
                cols: term.cols, rows: term.rows
            }
            let payloadStr = JSON.stringify(payload);
            payloadStr = "\x04" + payloadStr;
            ws.send(new TextEncoder().encode(payloadStr))
        });
    }
    // handle connection close
    ws.onclose = function () {
        showStatus("Connection lost. Refresh to reconnect.")
    }
}

// Initiate the connection
init()
    .then(() => {
        console.log("inited")
    })
    .catch((e) => {
        showStatus("Error: " + e.message);
    })