const media = videojs('media-player', {
    children: [
      'bigPlayButton',
      'controlBar'
    ],
    playbackRates: [1],
    preload: "auto",
    aspectRatio: "16:9",
    fluid: true,
    autoplay: true,
    controls: true
});

const msgBox = document.getElementById("message-box");
const sentBtn = document.getElementById("sent-message");
const inputText = document.getElementById("input-text");
const participantNum = document.getElementById("participant-value");
const webAddr = "117.6.56.99";

media.src({
    src: 'https://1080.opstream4.com/20231201/49651_f34a17a8/3000k/hls/mixed.m3u8',
    type: 'application/x-mpegURL'
});

var ws = new WebSocket('ws://' + webAddr + ':3000/ws');

ws.onopen = updateParticipantReq;

window.onload = function () {
    fetch('http://' + webAddr + ':3000/clientBoxData', {
        method: 'GET',
        credentials: 'include'
    }).then(function(res) {
        if (res.status == 200) 
            {
                res.json().then(function (data) {
                    if (data.is_owner)
                    {
                        ownerLayout(true);
                        setInterval(sendMovieData, 1000);
                        media.on("pause", sendMovieData);
                        media.on("timeupdate", sendMovieData);
                    }
                    else
                        ownerLayout(false);

                    document.getElementById("box-id").innerHTML = "#" + data.box_id;
                })
                .catch(err => alert(err));
            }
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
};

sentBtn.onclick = function () {
    let content = inputText.value;
    if (content == "")
        return;

    let msgNode = document.createElement("li");
    msgNode.classList.add("message");
    msgNode.classList.add("user-sent");
    msgNode.innerHTML = content;

    let senders = msgBox.querySelectorAll(".sender");
    if (senders.length == 0 || senders[senders.length - 1].innerHTML != "Me") {
        let senderNode = document.createElement("li");
        senderNode.classList.add("sender");
        senderNode.classList.add("user-sent");
        senderNode.innerHTML = "Me";
        msgBox.appendChild(senderNode);
    }

    msgBox.appendChild(msgNode);
    ws.send(JSON.stringify(
        {
            datatype: 0,
            content: content,
        }
    ));

    inputText.value = "";
};

document.getElementById("expand").onclick = function () {
    media.requestFullscreen();
}

document.getElementById("power-off").onclick = function () {
    ws.close();
    fetch('http://' + webAddr + ':3000/delete', {
        method: 'POST',
        credentials: 'include'
    }).then(function(res) {
        if (res.status == 200 || res.status == 308) 
            {
                window.location.replace('http://' + webAddr + ':3000/login');
            }
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
}

document.getElementById("leave").onclick = function () {
    ws.close();
    fetch('http://' + webAddr + ':3000/leave', {
        method: 'POST',
        credentials: 'include'
    }).then(function(res) {
        if (res.status == 200 || res.status == 308) 
            {
                window.location.replace('http://' + webAddr + ':3000/login');
            }
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
}

function sendMovieData() {
    ws.send(JSON.stringify(
        {
            datatype: 1,
            elapsed: media.currentTime(),
            is_pause: media.paused()
        }
    ));
}

function updateParticipantReq() {
    ws.send(JSON.stringify(
        {
            datatype: 2
        }
    ));
}

ws.onmessage = function(evt) {
    let data = JSON.parse(evt.data);

    if (data.datatype == 0) {
        let sender = data.username;
        let content = data.content;
    
        let msgNode = document.createElement("li");
        msgNode.classList.add("message");
        msgNode.innerHTML = content;
    
        let senders = msgBox.querySelectorAll(".sender");
        if (senders.length == 0 || senders[senders.length - 1].innerHTML != sender) {
            let senderNode = document.createElement("li");
            senderNode.classList.add("sender");
            senderNode.innerHTML = sender;
            msgBox.appendChild(senderNode);
        }
    
        msgBox.appendChild(msgNode);
    } else if (data.datatype == 1) {
        media.src(data.movie_url);
        media.currentTime(data.elapsed);
        if (data.is_pause)
            media.pause();
        else 
            media.play();
    } else if (data.datatype == 2) {
        participantNum.innerHTML = data.box_user_num;
    }
};

function ownerLayout(isActive) {
    if (isActive) {
        document.getElementById("request-pause").classList.add("d-none");
        document.getElementById("leave").classList.add("d-none");
        document.getElementById("power-off").classList.remove("d-none");
        media.controls(true);
    }
    else {
        document.getElementById("request-pause").classList.remove("d-none");
        document.getElementById("leave").classList.remove("d-none");
        document.getElementById("power-off").classList.add("d-none");
        media.controls(false);
    }
}