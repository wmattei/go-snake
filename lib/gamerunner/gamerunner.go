package gamerunner

import (
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/encodingutil"
	signaling_server "github.com/wmattei/go-snake/lib/signalingserver"
	"github.com/wmattei/go-snake/lib/webrtcutil"
)

type GameMetadata struct {
	GameName string
}

type Game interface {
	UpdateGameState(gameState *interface{}, command interface{}, dt int64)
	RenderFrame(gameState *interface{}, window *Window) []byte
	GetMetadata() *GameMetadata
}

type GameRunner struct {
	Game      Game
	Debugger  *debugutil.Debugger
	Signaler  signaling_server.SignalingServer
	commandCh chan interface{}

	playerConnectedCallback func(player *Player)
	player                  *Player
	rawFrameCh              chan *encodingutil.Canvas
	encodedFrameCh          chan *webrtcutil.Streamable
	closeSignal             chan bool
	gameStateCh             chan interface{}
}

type GameRunnerOptions struct {
	Debugger *debugutil.Debugger
	Signaler signaling_server.SignalingServer
}

type GameCommand struct {
	Type string
	Data interface{}
}

type Player struct {
	ID     string
	Window Window
}

func getGameRunnerOptions(opt *GameRunnerOptions) *GameRunnerOptions {
	if opt == nil {
		opt = &GameRunnerOptions{}
	}
	if opt.Debugger == nil {
		opt.Debugger = debugutil.NewDebugger(500)
	}
	if opt.Signaler == nil {
		opt.Signaler = &signaling_server.WebSocketSignalingServer{Port: "4000", MimeType: webrtc.MimeTypeH264}
	}
	return opt
}

func NewGameRunner(game Game, options *GameRunnerOptions) *GameRunner {
	options = getGameRunnerOptions(options)

	return &GameRunner{
		Game:           game,
		Debugger:       options.Debugger,
		Signaler:       options.Signaler,
		rawFrameCh:     make(chan *encodingutil.Canvas),
		encodedFrameCh: make(chan *webrtcutil.Streamable),
		closeSignal:    make(chan bool),
		gameStateCh:    make(chan interface{}),
		commandCh:      make(chan interface{}),
	}
}

func (g *GameRunner) OnPlayerJoined(callback func(player *Player)) {
	g.playerConnectedCallback = callback
}

func (g *GameRunner) StartEngine(initialGameState interface{}) {
	if constants.DEBUGGER {
		go g.Debugger.StartDebugger()
	}

	gameLoop := &gameLoop{
		gameState:      &initialGameState,
		closeSignal:    g.closeSignal,
		game:           g.Game,
		gameStateCh:    g.gameStateCh,
		commandChannel: g.commandCh,
	}
	go gameLoop.start()

	gameRenderer := &gameRenderer{
		gameStateCh: g.gameStateCh,
		rawFrameCh:  g.rawFrameCh,
		game:        g.Game,
		window:      &g.player.Window,
		debugger:    g.Debugger,
	}
	go gameRenderer.start()
}

func (g *GameRunner) OpenLobby() {
	g.Signaler.OnDataChannelEstablished(func(dataChannel *webrtc.DataChannel) {
		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message GameCommand
			err := json.Unmarshal(msg.Data, &message)
			if err != nil {
				fmt.Println("Error unmarshalling message:", err)
				return
			}

			if message.Type == "ping" {
				fmt.Println("Received ping")
				windowWidth := int(message.Data.(map[string]interface{})["width"].(float64))
				windowHeight := int(message.Data.(map[string]interface{})["height"].(float64))

				g.player = &Player{
					ID: "123",
					Window: Window{
						Width:  windowWidth,
						Height: windowHeight,
					},
				}

				if g.playerConnectedCallback != nil {
					g.playerConnectedCallback(g.player)
				}

				encoder := encodingutil.NewEncoder(&encodingutil.EncoderOptions{
					EncodedFrameChannel: g.encodedFrameCh,
					CanvasChannel:       g.rawFrameCh,
					CloseSignal:         g.closeSignal,
					Debugger:            g.Debugger,
					WindowHeight:        windowHeight,
					WindowWidth:         windowWidth,
				})
				encoder.Start()

				track := g.Signaler.GetVideoTrack()
				webrtcutil.StartStreaming(g.encodedFrameCh, track, g.Debugger)
			} else {
				g.commandCh <- message.Data
			}
		})
	})

	g.Signaler.Start()
}
