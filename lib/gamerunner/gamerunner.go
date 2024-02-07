package gamerunner

import (
	"encoding/json"

	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/lib/artemisia"
	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/encodingutil"
	signaling_server "github.com/wmattei/go-snake/lib/signalingserver"
	"github.com/wmattei/go-snake/lib/webrtcutil"
)

type GameMetadata struct {
	GameName string
}

type Game interface {
	Update(ctx *GameContext)
	RenderFrame(frame *artemisia.Frame)
}

type GameRunner struct {
	Game     Game
	Debugger *debugutil.Debugger
	Signaler signaling_server.SignalingServer

	rawFrameCh     chan *encodingutil.Canvas
	encodedFrameCh chan *webrtcutil.Streamable
	closeSignal    chan bool
	gameStateCh    chan interface{}
}

type GameRunnerOptions struct {
	Debugger *debugutil.Debugger
	Signaler signaling_server.SignalingServer
}

func getGameRunnerOptions(opt *GameRunnerOptions) *GameRunnerOptions {
	if opt == nil {
		opt = &GameRunnerOptions{}
	}
	if opt.Debugger == nil {
		opt.Debugger = debugutil.NewDebugger(500)
	}
	if opt.Signaler == nil {
		opt.Signaler = signaling_server.NewWebSocketSignalingServer("4000", webrtc.MimeTypeH264)
	}
	return opt
}

func NewGameRunner(game Game) *GameRunner {
	return NewGameRunnerWithOptions(game, nil)
}

func NewGameRunnerWithOptions(game Game, options *GameRunnerOptions) *GameRunner {
	options = getGameRunnerOptions(options)
	return &GameRunner{
		Game:           game,
		Debugger:       options.Debugger,
		Signaler:       options.Signaler,
		encodedFrameCh: make(chan *webrtcutil.Streamable),
		rawFrameCh:     make(chan *encodingutil.Canvas),
	}
}

func (g *GameRunner) run(gameContext *GameContext) {
	if g.Debugger != nil {
		go g.Debugger.StartDebugger()
	}

	gameRenderer := newGameRenderer(g.Game, g.rawFrameCh)
	gameRenderer.debugger = g.Debugger
	gameLoop := &gameLoop{
		closeSignal:  g.closeSignal,
		game:         g.Game,
		gameContext:  gameContext,
		gameRenderer: gameRenderer,
	}
	go gameLoop.start()

	encoder := encodingutil.NewEncoder(&encodingutil.EncoderOptions{
		EncodedFrameChannel: g.encodedFrameCh,
		CanvasChannel:       g.rawFrameCh,
		CloseSignal:         g.closeSignal,
		Debugger:            g.Debugger,
		WindowHeight:        gameContext.height,
		WindowWidth:         gameContext.width,
	})
	go encoder.Start()

	track := g.Signaler.GetVideoTrack()
	webrtcutil.StartStreaming(g.encodedFrameCh, track, g.Debugger)

}

func (g *GameRunner) RunAfterResize() {
	g.Signaler.Connect(func() {
		gameContext := NewGameContext()
		dataChannel := g.Signaler.GetDataChannel()

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message Command
			json.Unmarshal(msg.Data, &message)
			gameContext.handleCommand(&message)
			if message.Type == Resize {
				g.run(gameContext)
			}
		})
	})
}

func (g *GameRunner) Run() {
	g.Signaler.Connect(func() {
		gameContext := NewGameContext()
		dataChannel := g.Signaler.GetDataChannel()

		g.run(gameContext)

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message Command
			json.Unmarshal(msg.Data, &message)
			gameContext.handleCommand(&message)
		})
	})
}

// func (g *GameRunner) StopEngine() {
// 	fmt.Println("Stopping engine")
// 	close(g.closeSignal)

// 	if constants.DEBUGGER {
// 		g.Debugger.StopDebugger()
// 	}
// }

// func (g *GameRunner) OpenLobby() {
// 	g.Signaler.OnDataChannelEstablished(func(dataChannel *webrtc.DataChannel, pc *webrtc.PeerConnection) {
// 		fmt.Println("Data channel established")

// 		if constants.DEBUGGER {
// 			go g.Debugger.StartDebugger()
// 		}

// 		g.rawFrameCh = make(chan *encodingutil.Canvas)
// 		g.encodedFrameCh = make(chan *webrtcutil.Streamable)
// 		g.closeSignal = make(chan bool)
// 		g.gameStateCh = make(chan interface{})
// 		g.commandCh = make(chan interface{})

// 		dataChannel.OnClose(func() {
// 			g.StopEngine()
// 			pc.Close()
// 		})

// 		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
// 			var message GameCommand
// 			err := json.Unmarshal(msg.Data, &message)
// 			if err != nil {
// 				fmt.Println("Error unmarshalling message:", err)
// 				return
// 			}

// 			if message.Type == "ping" {
// 				fmt.Println("Received ping")
// 				windowWidth := int(message.Data.(map[string]interface{})["width"].(float64))
// 				windowHeight := int(message.Data.(map[string]interface{})["height"].(float64))

// 				g.player = &Player{
// 					ID: "123",
// 					Window: Window{
// 						Width:  windowWidth,
// 						Height: windowHeight,
// 					},
// 				}

// 				if g.playerConnectedCallback != nil {
// 					g.playerConnectedCallback(g.player)
// 				}

// 				encoder := encodingutil.NewEncoder(&encodingutil.EncoderOptions{
// 					EncodedFrameChannel: g.encodedFrameCh,
// 					CanvasChannel:       g.rawFrameCh,
// 					CloseSignal:         g.closeSignal,
// 					Debugger:            g.Debugger,
// 					WindowHeight:        windowHeight,
// 					WindowWidth:         windowWidth,
// 				})
// 				encoder.Start()

// 				track := g.Signaler.GetVideoTrack()
// 				webrtcutil.StartStreaming(g.encodedFrameCh, track, g.Debugger)
// 			} else {
// 				g.commandCh <- message.Data
// 			}
// 		})
// 	})

// 	g.Signaler.Start()
// }
