package gamerunner

import (
	"encoding/json"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/lib/artemisia"
	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/encodingutil"
	"github.com/wmattei/go-snake/lib/internal"
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
	game     Game
	debugger *debugutil.Debugger
	signaler signaling_server.SignalingServer

	rawFrameCh     chan *encodingutil.Canvas
	encodedFrameCh chan *webrtcutil.Streamable
	closeSignal    chan bool

	dimensionsCh chan *encodingutil.Dimensions
	running      bool
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
	dimCh := make(chan *encodingutil.Dimensions)
	return &GameRunner{
		game:           game,
		debugger:       options.Debugger,
		signaler:       options.Signaler,
		encodedFrameCh: make(chan *webrtcutil.Streamable),
		rawFrameCh:     make(chan *encodingutil.Canvas),
		dimensionsCh:   dimCh,
	}
}

func (g *GameRunner) run(gameContext *GameContext) {
	if g.debugger != nil {
		// go g.debugger.StartDebugger()
	}

	gameRenderer := newGameRenderer(g.game, g.rawFrameCh)
	gameRenderer.debugger = g.debugger
	gameLoop := &gameLoop{
		closeSignal:  g.closeSignal,
		game:         g.game,
		gameContext:  gameContext,
		gameRenderer: gameRenderer,
	}
	go gameLoop.start()

	encoder := encodingutil.NewEncoder(&encodingutil.EncoderOptions{
		EncodedFrameChannel: g.encodedFrameCh,
		CanvasChannel:       g.rawFrameCh,
		CloseSignal:         g.closeSignal,
		Debugger:            g.debugger,
		DimensionsChannel:   internal.Debounce(g.dimensionsCh, time.Second),
	})
	go encoder.Start()

	track := g.signaler.GetVideoTrack()
	webrtcutil.StartStreaming(g.encodedFrameCh, track, g.debugger)

}

func (g *GameRunner) RunAfterResize() {
	g.signaler.Connect(func() {
		gameContext := NewGameContext()
		dataChannel := g.signaler.GetDataChannel()

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message Command
			json.Unmarshal(msg.Data, &message)
			gameContext.handleCommand(&message)

			if message.Type == Resize {
				if !g.running {
					g.run(gameContext)
					g.running = true
				}
				g.dimensionsCh <- &encodingutil.Dimensions{Width: gameContext.width, Height: gameContext.height}
			}
		})
	})
}

func (g *GameRunner) Run() {
	g.signaler.Connect(func() {
		gameContext := NewGameContext()
		dataChannel := g.signaler.GetDataChannel()

		g.run(gameContext)

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message Command
			json.Unmarshal(msg.Data, &message)
			gameContext.handleCommand(&message)

			if message.Type == Resize {
				g.dimensionsCh <- &encodingutil.Dimensions{Width: gameContext.width, Height: gameContext.height}
			}
		})
	})
}
