/** @type {RTCDataChannel} */
let dataChannel;

/** @type {RTCPeerConnection} */
let peerConnection;

let signalingChannel;

window.onload = async function () {
  const startBtn = document.getElementById("start_btn");
  startBtn.addEventListener("click", start);

  const stopBtn = document.getElementById("stop_btn");
  stopBtn.addEventListener("click", stop);
};

function stop() {
  dataChannel.close();
  peerConnection.close();
}

async function start() {
  signalingChannel = new WebSocket("ws://localhost:4000/ws");
  keyBindings();
  signalingChannel.addEventListener("open", async () => {
    peerConnection = new RTCPeerConnection();
    createDataChannel(peerConnection);

    peerConnection.onicecandidate = handleIceCandidateEvent;

    const offer = await peerConnection.createOffer();
    console.log("Setting local description...");
    await peerConnection.setLocalDescription(offer);

    console.log("Sending offer...");
    signalingChannel.send(JSON.stringify({ type: "offer", data: offer.sdp }));
  });

  signalingChannel.addEventListener("message", async (event) => {
    const message = JSON.parse(event.data);
    if (message.type === "ice") {
      console.log("Received ICE from server");
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
      console.log("Received answer from server, setting remote description...");
      const remoteDescription = new RTCSessionDescription({
        sdp: message.data,
        type: "answer",
      });
      await peerConnection.setRemoteDescription(remoteDescription);
    }
  });
}

const handleIceCandidateEvent = (event) => {
  if (event.candidate) {
    console.log("Sending ice candidate to server...");
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
    console.log("Data channel is closed");
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
