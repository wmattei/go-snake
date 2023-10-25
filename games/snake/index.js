let dataChannel;
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
  // Initialize signaling channel
  signalingChannel = new WebSocket("ws://localhost:4000/ws");

  // Event handler for signaling channel open
  signalingChannel.addEventListener("open", async () => {
    // Initialize peer connection and data channel
    peerConnection = new RTCPeerConnection();
    createDataChannel(peerConnection);

    peerConnection.onicecandidate = handleIceCandidateEvent;
    peerConnection.ontrack = handleTrackEvent;

    // Create and send offer
    const offer = await peerConnection.createOffer({
      offerToReceiveVideo: true,
    });
    await peerConnection.setLocalDescription(offer);
    signalingChannel.send(JSON.stringify({ type: "offer", data: offer.sdp }));
  });

  // Event handler for signaling channel messages
  signalingChannel.addEventListener("message", async (event) => {
    const message = JSON.parse(event.data);

    if (message.type === "ice") {
      handleIceMessage(message.data);
    }

    if (message.type === "answer") {
      handleAnswerMessage(message.data);
    }
  });
}

const handleTrackEvent = (event) => {
  if (event.track.kind === "video") {
    // Display video and hide welcome panel
    videoElement.style.display = "block";
    welcomePanelElement.style.display = "none";

    // Set video stream
    videoElement.srcObject = event.streams[0];
  }
};

const handleIceCandidateEvent = (event) => {
  if (event.candidate) {
    // Send ICE candidate
    signalingChannel.send(
      JSON.stringify({ type: "ice", data: event.candidate.candidate })
    );
  }
};

const handleIceMessage = async (iceData) => {
  const iceCandidate = new RTCIceCandidate({
    ...iceData,
    sdpMLineIndex: 0,
    sdpMid: "0",
  });
  try {
    await peerConnection.addIceCandidate(iceCandidate);
  } catch (error) {
    console.log("Error adding ICE candidate:", error);
  }
};

const handleAnswerMessage = async (answerData) => {
  const remoteDescription = new RTCSessionDescription({
    sdp: answerData,
    type: "answer",
  });
  await peerConnection.setRemoteDescription(remoteDescription);
};

function sendCommand(type, data) {
  dataChannel.send(JSON.stringify({ type, data }));
}

function createDataChannel(peerConnection) {
  // Create a data channel for commands
  dataChannel = peerConnection.createDataChannel("commandsChannel");

  dataChannel.onerror = (error) => {
    console.log("Error on data channel:", error);
  };

  dataChannel.onopen = () => {
    sendCommand("ping", {
      width: window.innerWidth,
      height: window.innerHeight,
    });
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
    // Send command based on key press
    if (event.key === "ArrowUp") {
      sendCommand("CHANGE_DIR", { dir: "UP" });
    } else if (event.key === "ArrowDown") {
      sendCommand("CHANGE_DIR", { dir: "DOWN" });
    } else if (event.key === "ArrowLeft") {
      sendCommand("CHANGE_DIR", { dir: "LEFT" });
    } else if (event.key === "ArrowRight") {
      sendCommand("CHANGE_DIR", { dir: "RIGHT" });
    }
  });
}
