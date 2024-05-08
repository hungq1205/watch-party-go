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

const msgBox = document.getElementById("message-box")
const sentBtn = document.getElementById("sent-message")
const inputText = document.getElementById("input-text")

media.src({
    src: 'https://1080.opstream4.com/20231201/49651_f34a17a8/3000k/hls/mixed.m3u8',
    type: 'application/x-mpegURL'
});

media.addEventListener("play", sendMovieData);
media.addEventListener("pause", sendMovieData);
media.addEventListener("timeupdate", sendMovieData);

var ws = new WebSocket("ws://localhost:3000/ws");

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

function sendMovieData() {
    ws.send(JSON.stringify(
        {
            datatype: 1,
            movie_url: media.currentSrc(),
            elapsed: media.currentTime(),
            is_pause: media.paused()
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
    }

};