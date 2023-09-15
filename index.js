/** @type {RTCDataChannel} */
let dataChannel;

/** @type {RTCPeerConnection} */
let peerConnection;

let signalingChannel;

let videoElement;
let welcomePanelElement;

window.onload = async function () {
  const startBtn = document.getElementById("start_btn");
  startBtn.addEventListener("click", start);

  videoElement = document.getElementById("video");
  welcomePanelElement = document.getElementById("welcome_panel");

  keyBindings();
};

async function start() {
  signalingChannel = new WebSocket("ws://localhost:4000/ws");

  signalingChannel.addEventListener("open", async () => {
    peerConnection = new RTCPeerConnection();
    createDataChannel(peerConnection);

    peerConnection.onicecandidate = handleIceCandidateEvent;
    peerConnection.ontrack = handleTrackEvent;

    const offer = await peerConnection.createOffer({
      offerToReceiveVideo: true,
    });
    await peerConnection.setLocalDescription(offer);

    signalingChannel.send(JSON.stringify({ type: "offer", data: offer.sdp }));
  });

  signalingChannel.addEventListener("message", async (event) => {
    const message = JSON.parse(event.data);
    if (message.type === "ice") {
      const iceCandidate = new RTCIceCandidate({
        ...message.data,
        sdpMLineIndex: 0,
        sdpMid: "0",
      });
      try {
        await peerConnection.addIceCandidate(iceCandidate);
      } catch (error) {
        console.log(error);
      }
    }
    if (message.type === "answer") {
      const remoteDescription = new RTCSessionDescription({
        sdp: message.data,
        type: "answer",
      });
      await peerConnection.setRemoteDescription(remoteDescription);
    }
  });
}

const handleTrackEvent = (event) => {
  if (event.track.kind === "video") {
    videoElement.style.display = "block";
    welcomePanelElement.style.display = "none";

    videoElement.srcObject = event.streams[0];
  }
};

const handleIceCandidateEvent = (event) => {
  if (event.candidate) {
    signalingChannel.send(
      JSON.stringify({
        type: "ice",
        data: event.candidate.candidate,
      })
    );
  }
};

function sendCommand(command) {
  console.log("Sending command:", command);
  dataChannel.send(JSON.stringify({ type: "command", data: command }));
}

function createDataChannel(peerConnection) {
  dataChannel = peerConnection.createDataChannel("commandsChannel");

  dataChannel.onerror = (error) => {
    console.log("Error On Data channel:", error);
  };

  dataChannel.onclose = () => {
    setTimeout(() => {
      videoElement.style.display = "none";
      welcomePanelElement.style.display = "flex";
    }, 1000);
  };
}

function keyBindings() {
  document.addEventListener("keydown", (event) => {
    if (event.key === "ArrowUp") {
      sendCommand("UP");
    }
    if (event.key === "ArrowDown") {
      sendCommand("DOWN");
    }
    if (event.key === "ArrowLeft") {
      sendCommand("LEFT");
    }
    if (event.key === "ArrowRight") {
      sendCommand("RIGHT");
    }
  });
}
