let deviceNames = [];

fetch("/config")
    .catch(e => console.error(e))
    .then(r => r.json())
    .then(c => c.devices.forEach(d => addDevice(d, false)));

function addDevice(name, doSave) {
    if (name.length > 0 && deviceNames.indexOf(name) === -1) {
        let devices = document.getElementById("devices");
        let device = render(
            `<div class="device" id="${name}">${name}<div onclick="removeDevice('${name}')">x</div></div>`
        );
        deviceNames.push(name);
        devices.appendChild(device);
        if (doSave === undefined || doSave) {
            save();
        }
    }
}

function removeDevice(name, doSave) {
    let devices = document.getElementById("devices");
    let device = document.getElementById(name);
    devices.removeChild(device);
    deviceNames = deviceNames.filter(d => d !== name);
    if (doSave === undefined || doSave) {
        save();
    }
}

function render(html) {
    let div = document.createElement("div");
    div.innerHTML = html;
    return div.firstChild;
}

function save() {
    let post = {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({devices: deviceNames})
    };

    fetch("/config", post)
        .catch(e => console.error(e))
        .then(() => console.log("saved"))
}