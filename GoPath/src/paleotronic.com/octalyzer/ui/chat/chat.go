package chat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"paleotronic.com/api"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/server/chatapi"
	"paleotronic.com/utils"
)

type ChatCommand struct {
	Verb string
	Args []string
}

type ChatClient struct {
	InputBuffer                        HRec
	InsertPos                          int
	ViewPos                            int
	CurrentChannel                     string
	Joined                             map[string]bool
	Channels                           map[string]*chatapi.ChatDetails
	ChannelBuffer                      map[string][]*chatapi.MessageDetails
	JList, NJList                      []string
	SidebarWidth                       int
	ChatWidth                          int
	InputHeight                        int
	InputWidth                         int
	InputMax                           int
	InputPrompt                        string
	txt                                *types.TextBuffer
	Int                                interfaces.Interpretable
	running                            bool
	command                            chan ChatCommand
	message                            chan *chatapi.MessageDetails
	keyevent                           chan rune
	chatdetails                        chan *chatapi.BroadcastChatDetails
	chatjoin                           chan *chatapi.BroadcastChatJoin
	chatpart, chataway                 chan *chatapi.BroadcastChatLeave
	sidebarX, sidebarY                 int
	sidebarFG, sidebarBG, sidebarShade uint64
	inputX, inputY                     int
	inputFG, inputBG, inputShade       uint64
	chatlogX, chatlogY                 int
	chatlogFG, chatlogBG, chatlogShade uint64
	titleX, titleY                     int
	titleFG, titleBG, titleShade       uint64
	ScrollBack                         *ChatBuffer
	InScrollback                       bool
	keyFeederDisable                   bool
	InsColour                          int
	RemoteMode                         bool
	StartChat                          string
	lastHCMode                         bool
}

var menuFunc func(e interfaces.Interpretable)

func SetMenuHook(f func(e interfaces.Interpretable)) {
	menuFunc = f
}

func TestMenu(e interfaces.Interpretable) {
	if menuFunc != nil {
		menuFunc(e)
	}
}

func NewChatClient(e interfaces.Interpretable, startChat string) *ChatClient {
	width := 80
	sideWidth := 20
	chatWidth := width - sideWidth

	if startChat != "" {

	}

	c := &ChatClient{
		SidebarWidth:  sideWidth,
		ChatWidth:     chatWidth,
		InputHeight:   1,
		InputWidth:    width,
		InputPrompt:   "{{username}}: ",
		txt:           apple2helpers.GETHUD(e, "MONI").Control,
		Int:           e,
		command:       make(chan ChatCommand, 1000),
		message:       make(chan *chatapi.MessageDetails, 500),
		keyevent:      make(chan rune, 2000),
		chatdetails:   make(chan *chatapi.BroadcastChatDetails, 100),
		chatjoin:      make(chan *chatapi.BroadcastChatJoin, 100),
		chatpart:      make(chan *chatapi.BroadcastChatLeave, 100),
		chataway:      make(chan *chatapi.BroadcastChatLeave, 100),
		Joined:        make(map[string]bool),
		Channels:      make(map[string]*chatapi.ChatDetails),
		ChannelBuffer: make(map[string][]*chatapi.MessageDetails),
		InputMax:      256,
		InsColour:     15,
		StartChat:     startChat,
		RemoteMode:    (startChat != ""),
	}

	return c
}

func (c *ChatClient) CheckPalette() {
	if c.lastHCMode != settings.HighContrastUI {
		//log2.Printf("toggled")
		c.lastHCMode = settings.HighContrastUI
		c.SetupWindows()
		c.showSidebarState()
		c.showTitleState()
		c.ScrollBack.Draw(c)
		c.showInputState()
		c.txt.FullRefresh()
	}
}

/*
	CIM - chat incoming message
	COM - chat outgoing message
*/
func (c *ChatClient) Init() {

	c.txt.FGColor = 15
	c.txt.BGColor = 0
	c.txt.ClearScreen()

	s8webclient.CONN.GetDT().CustomHandler["KMB"] = c.HandleMessage
	s8webclient.CONN.GetDT().CustomHandler["KTB"] = c.HandleMessage
	s8webclient.CONN.GetDT().CustomHandler["KJB"] = c.HandleMessage
	s8webclient.CONN.GetDT().CustomHandler["KLB"] = c.HandleMessage
	s8webclient.CONN.GetDT().CustomHandler["KAB"] = c.HandleMessage
	if c.RemoteMode {
		remoteWidth := 24
		c.InputWidth = remoteWidth
		c.txt.AddNamedWindow(
			"title",
			0,
			0,
			remoteWidth,
			0,
		)
		// c.txt.AddNamedWindow(
		// 	"sidebar",
		// 	80-c.SidebarWidth,
		// 	1,
		// 	79,
		// 	46,
		// )
		c.txt.AddNamedWindow(
			"chatlog",
			0,
			1,
			remoteWidth,
			46,
		)
		c.txt.AddNamedWindow(
			"input",
			0,
			47,
			remoteWidth,
			47,
		)
		c.ScrollBack = NewChatBuffer(c.Int, remoteWidth, 2, 2)
	} else {

		c.txt.AddNamedWindow(
			"title",
			0,
			0,
			79,
			0,
		)
		c.txt.AddNamedWindow(
			"sidebar",
			80-c.SidebarWidth,
			1,
			79,
			46,
		)
		c.txt.AddNamedWindow(
			"chatlog",
			0,
			1,
			79-c.SidebarWidth,
			46,
		)
		c.txt.AddNamedWindow(
			"input",
			0,
			47,
			79,
			47,
		)
		c.ScrollBack = NewChatBuffer(c.Int, 60, 2, 2)
	}

	c.CheckPalette()
	c.SetupWindows()
}

func (c *ChatClient) SetupWindows() {
	// Put some test crap in
	c.TitleDo(
		func(txt *types.TextBuffer) {
			if settings.HighContrastUI {
				txt.FGColor = 0
				txt.BGColor = 15
			} else {
				txt.FGColor = 2
				txt.BGColor = 15
			}
			txt.ClearScreenWindow()
			txt.PutStr("Title area")
		},
	)
	c.SidebarDo(
		func(txt *types.TextBuffer) {
			if settings.HighContrastUI {
				txt.FGColor = 15
				txt.BGColor = 0
			} else {
				txt.FGColor = 12
				txt.BGColor = 4
			}
			txt.ClearScreenWindow()
			txt.PutStr("Sidebar Area")
		},
	)
	c.ChatlogDo(
		func(txt *types.TextBuffer) {
			if settings.HighContrastUI {
				txt.BGColor = 15
				txt.FGColor = 0
			} else {
				txt.FGColor = 7
				txt.BGColor = 2
			}
			txt.ClearScreenWindow()
			txt.PutStr("Chatlog Area")
		},
	)
	c.InputDo(
		func(txt *types.TextBuffer) {
			if settings.HighContrastUI {
				txt.FGColor = 15
				txt.BGColor = 0
			} else {
				txt.FGColor = 15
				txt.BGColor = 1
			}
			txt.ClearScreenWindow()
			txt.PutStr("Input Area")
		},
	)

	c.PositionScreens()
}

func (c *ChatClient) setSlotAspect(index int, aspect float64) {
	mm := c.Int.GetMemoryMap()
	for cindex := 0; cindex < 9; cindex++ {
		control := types.NewOrbitController(mm, index, cindex-1)
		control.SetAspect(aspect)
	}
}

func (c *ChatClient) PositionScreens() {
	mm := c.Int.GetMemoryMap()
	if c.RemoteMode {
		c.Int.GetProducer().SetMasterLayerPos(0, 0.630, 0)
		c.Int.GetProducer().SetMasterLayerPos(1, -0.125, 0)
		c.setSlotAspect(0, 1.333333)
		c.setSlotAspect(1, 1.333333)
		mm.IntSetLayerState(0, 1)
		mm.IntSetLayerState(1, 1)
		mm.IntSetActiveState(0, 1)
		mm.IntSetActiveState(1, 1)
	} else {
		c.Int.GetProducer().SetMasterLayerPos(0, 0, 0)
		c.Int.GetProducer().SetMasterLayerPos(1, 0, 0)
		c.setSlotAspect(0, 1.45)
		c.setSlotAspect(1, 1.45)
		mm.IntSetLayerState(0, 1)
		mm.IntSetLayerState(1, 0)
		mm.IntSetActiveState(0, 1)
		mm.IntSetActiveState(1, 0)
	}
}

func (c *ChatClient) TitleDo(f func(txt *types.TextBuffer)) {
	// if c.RemoteMode {
	// 	return
	// }
	c.txt.HideCursor()
	c.txt.SetNamedWindow("title")
	c.txt.GotoXYWindow(c.titleX, c.titleY)
	c.txt.FGColor = c.titleFG
	c.txt.BGColor = c.titleBG
	c.txt.Shade = c.titleShade
	f(c.txt)
	c.titleShade = c.txt.Shade
	c.titleFG = c.txt.FGColor
	c.titleBG = c.txt.BGColor
	c.titleX, c.titleY = c.txt.GetXYWindow()
}

func (c *ChatClient) SidebarDo(f func(txt *types.TextBuffer)) {
	if c.RemoteMode {
		return
	}
	c.txt.HideCursor()
	c.txt.SetNamedWindow("sidebar")
	c.txt.GotoXYWindow(c.sidebarX, c.sidebarY)
	c.txt.FGColor = c.sidebarFG
	c.txt.BGColor = c.sidebarBG
	c.txt.Shade = c.sidebarShade
	f(c.txt)
	c.sidebarShade = c.txt.Shade
	c.sidebarFG = c.txt.FGColor
	c.sidebarBG = c.txt.BGColor
	c.sidebarX, c.sidebarY = c.txt.GetXYWindow()
}

func (c *ChatClient) ChatlogDo(f func(txt *types.TextBuffer)) {
	c.txt.HideCursor()
	c.txt.SetNamedWindow("chatlog")
	c.txt.GotoXYWindow(c.chatlogX, c.chatlogY)
	c.txt.FGColor = c.chatlogFG
	c.txt.BGColor = c.chatlogBG
	c.txt.Shade = c.chatlogShade
	f(c.txt)
	c.chatlogShade = c.txt.Shade
	c.chatlogFG = c.txt.FGColor
	c.chatlogBG = c.txt.BGColor
	c.chatlogX, c.chatlogY = c.txt.GetXYWindow()
}

func (c *ChatClient) InputDo(f func(txt *types.TextBuffer)) {
	c.txt.HideCursor()
	c.txt.SetNamedWindow("input")
	c.txt.GotoXYWindow(c.inputX, c.inputY)
	c.txt.FGColor = c.inputFG
	c.txt.BGColor = c.inputBG
	c.txt.Shade = c.inputShade
	f(c.txt)
	c.inputShade = c.txt.Shade
	c.inputFG = c.txt.FGColor
	c.inputBG = c.txt.BGColor
	c.inputX, c.inputY = c.txt.GetXYWindow()
	c.txt.ShowCursor()
}

func (c *ChatClient) Done() {
	delete(s8webclient.CONN.GetDT().CustomHandler, "KMB")
	delete(s8webclient.CONN.GetDT().CustomHandler, "KTB")
	delete(s8webclient.CONN.GetDT().CustomHandler, "KJB")
	delete(s8webclient.CONN.GetDT().CustomHandler, "KLB")
	delete(s8webclient.CONN.GetDT().CustomHandler, "KAB")
	if c.RemoteMode {
		e := c.Int.GetProducer().GetInterpreter(1)
		e.EndRemote()
	}
	c.RemoteMode = false
	c.PositionScreens()
}

func (c *ChatClient) GetPrompt() string {
	ps := c.InputPrompt
	ps = strings.Replace(ps, "{{username}}", s8webclient.CONN.Username, -1)
	return ps
}

func (c *ChatClient) KeyFeeder() {
	for c.running {
		if c.Int.GetMemoryMap().KeyBufferPeek(c.Int.GetMemIndex()) != 0 && !c.keyFeederDisable {
			keycode := c.Int.GetMemoryMap().KeyBufferGet(c.Int.GetMemIndex())
			c.keyevent <- rune(keycode)
		} else {
			if c.Int.GetMemoryMap().IntGetSlotMenu(c.Int.GetMemIndex()) {
				if menuFunc != nil {
					menuFunc(c.Int)
				}
				c.Int.GetMemoryMap().IntSetSlotMenu(c.Int.GetMemIndex(), false)
			}
			time.Sleep(16 * time.Millisecond)
		}
	}
}

func (c *ChatClient) fetchChats() {
	chats, err := s8webclient.CONN.FetchChats()
	if err != nil {
		log.Printf("Fetching chats failed: %v", err)
		return
	}
	if chats == nil {
		log.Println("fetch chats returned nil :(")
		return
	}
	log.Printf("Fetch chats response: %+v", chats)
	c.JList = make([]string, 0)
	c.NJList = make([]string, 0)
	c.Channels = make(map[string]*chatapi.ChatDetails)
	c.Joined = make(map[string]bool)
	for _, cd := range chats.Joined {
		c.Joined[cd.Name] = true
		c.Channels[strings.ToLower(cd.Name)] = cd
		c.JList = append(c.JList, cd.Name)
	}
	for _, cd := range chats.NotJoined {
		c.Channels[strings.ToLower(cd.Name)] = cd
		c.NJList = append(c.NJList, cd.Name)
	}
	sort.Strings(c.JList)
	sort.Strings(c.NJList)
	log.Printf("channel details: %+v", c.Channels)
	log.Printf("joined: %+v", c.Joined)
	log.Printf("current: %s", c.CurrentChannel)
	c.showSidebarState()
	c.checkSubs() // make sure we are subscribed etc properly
	c.showInputState()
}

func (c *ChatClient) checkSubs() {
	for _, cd := range c.Channels {
		if c.Joined[cd.Name] {
			s8webclient.CONN.SubscribeChannel(fmt.Sprintf("microchat-%.8x", cd.ChatId))
		} else {
			s8webclient.CONN.UnsubscribeChannel(fmt.Sprintf("microchat-%.8x", cd.ChatId))
		}
	}
}

func (c *ChatClient) JoinChat(name string) (*chatapi.JoinChatResponse, error) {
	c.fetchChats()
	cd, ok := c.Channels[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("Join failed: not existing")
	}

	if strings.HasPrefix(name, "@") && !c.RemoteMode {
		return nil, fmt.Errorf("Join failed: chat unavailable")
	}

	// force early sub
	c.Joined[cd.Name] = true
	c.checkSubs()

	r, err := s8webclient.CONN.JoinChat(cd.ChatId)
	log.Printf("Join result = %v", err)
	if r != nil {
		c.chatdetails <- &chatapi.BroadcastChatDetails{
			ChatId:  r.ChatId,
			Details: r.Details,
		}
		c.ChannelBuffer[name] = r.Backlog
		c.CurrentChannel = name
		c.showSidebarState()
		c.showTitleState()
		c.LoadBacklog()
		if c.RemoteMode {
			c.Console("Channel topic is: %s\r\n", r.Details.Topic)
		}
	}
	log.Printf("channel details: %+v", c.Channels)
	log.Printf("joined: %+v", c.Joined)
	log.Printf("current: %s", c.CurrentChannel)
	return r, err
}

func (c *ChatClient) LoadBacklog() {
	c.ScrollBack.Empty()
	c.ScrollBack.SuppressDraw = true
	for i := len(c.ChannelBuffer[c.CurrentChannel]) - 1; i >= 0; i-- {
		//c.showChatMessage(c.ChannelBuffer[c.CurrentChannel][i])
		c.AddToBacklog(c.ChannelBuffer[c.CurrentChannel][i])
	}
	c.ScrollBack.SuppressDraw = false
	c.ScrollBack.Draw(c)
	c.showTitleState()
	c.showInputState()
}

func (c *ChatClient) LeaveChat(name string) error {
	cd, ok := c.Channels[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("Leave failed: not existing")
	}
	_, err := s8webclient.CONN.LeaveChat(cd.ChatId)
	if err == nil {
		c.fetchChats()
		if c.CurrentChannel == name {
			if len(c.JList) > 0 {
				c.CurrentChannel = c.JList[0]
				c.JoinChat(c.JList[0])
			} else {
				c.CurrentChannel = ""
				c.ScrollBack.Empty()
				c.ScrollBack.Printfr("You are not joined to a chat. Join with /join <name>.\r\n")
			}
			c.showSidebarState()
			c.showTitleState()
			c.showInputState()
			c.ScrollBack.Draw(c)
		}
	}
	return err
}

func (c *ChatClient) general() {

	// chats already fetched...
	if c.RemoteMode {
		cd := c.Channels[strings.ToLower(c.StartChat)]
		if cd != nil && cd.RemintHost != "" {
			c.JoinChat(c.StartChat)
			log.Printf("Attempting to connect to remote on %s:%d", cd.RemintHost, cd.RemintPort)

			c.Int.GetProducer().Activate(1)

			e := c.Int.GetProducer().GetInterpreter(1)
			e.Bootstrap("fp", true)

			//log.Printf("Trying: @share.connect{\"%s\", %d, %d}", cd.RemintHost, cd.RemintPort, e.GetMemIndex()+1)
			//e.Parse("fp")
			e.Parse(fmt.Sprintf("@share.connect{\"%s\", %d, %d}", cd.RemintHost, cd.RemintPort, e.GetMemIndex()+1))
			//go e.ConnectRemote(cd.RemintHost, utils.IntToStr(int(cd.RemintPort)), 0)
			c.PositionScreens()
			time.AfterFunc(3*time.Second, c.PositionScreens)
		}
		return
	}

	if len(c.JList) > 0 {
		c.JoinChat(c.JList[0])
	}

}

var ctable = [][]int{
	{8, 9, 10, 11, 12, 13, 14, 15},
	{15, 14, 13, 12, 11, 10, 9, 8},
	{8, 10, 12, 14, 15, 13, 11, 9},
	{9, 11, 13, 15, 14, 12, 10, 8},
	{9, 11, 13, 15, 14, 12, 10, 8},
	{8, 10, 12, 14, 15, 13, 11, 9},
	{15, 14, 13, 12, 11, 10, 9, 8},
	{8, 9, 10, 11, 12, 13, 14, 15},
}

func (c *ChatClient) nickColor(name string) uint64 {
	x := int(name[0]) & 7
	y := int(name[len(name)-1]) & 7
	return uint64(ctable[x][y])
}

func (c *ChatClient) showChatMessage(m *chatapi.MessageDetails) {

	if c.Channels != nil && c.Channels[c.CurrentChannel] != nil && c.Channels[c.CurrentChannel].ChatId != m.ChatId {
		return
	}

	if c.ChannelBuffer == nil {
		c.ChannelBuffer[c.CurrentChannel] = []*chatapi.MessageDetails{}
	}
	c.ChannelBuffer[c.CurrentChannel] = append(c.ChannelBuffer[c.CurrentChannel], m)

	c.AddToBacklog(m)
	c.showInputState()

}

func (c *ChatClient) Console(pattern string, args ...interface{}) {
	if c.ScrollBack.edit.Column > 0 {
		c.ScrollBack.Printf("\r\n")
	}
	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
	c.ScrollBack.Printfr(pattern, args...)
	c.ScrollBack.GotoBottom()
	c.ScrollBack.Draw(c)
}

func (c *ChatClient) AddToBacklog(m *chatapi.MessageDetails) {
	if c.ScrollBack.edit.Column > 0 {
		c.ScrollBack.Printf("\r\n")
	}
	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.nickColor(m.Creator))
	if m.IsAction {
		c.ScrollBack.Printf("* %s ", m.Creator)
	} else {
		c.ScrollBack.Printf("<%s> ", m.Creator)
	}
	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
	c.ScrollBack.Printfw("%s\r\n", m.Message)
	c.ScrollBack.GotoBottom()
	c.ScrollBack.Draw(c)
}

func (c *ChatClient) channelDetails() *chatapi.ChatDetails {
	details := c.Channels[strings.ToLower(c.CurrentChannel)]
	if details == nil {
		details = &chatapi.ChatDetails{
			Name:          "Not connected",
			ChatId:        0,
			ActiveUsers:   []string{},
			InactiveUsers: []string{},
			Topic:         "No chat joined. Try /join <chatname> to get started",
		}
	}
	return details
}

func (c *ChatClient) showInputState() {
	c.InputDo(
		func(txt *types.TextBuffer) {
			txt.ClearScreenWindow()
			txt.GotoXYWindow(0, 0)
			txt.FGColor = c.inputFG
			uname := c.GetPrompt()
			rw := c.InputWidth - len(uname) - 1
			txt.PutStr(uname)
			offs := c.ViewPos
			end := c.ViewPos + rw
			if end > len(c.InputBuffer.Data.Runes) {
				end = len(c.InputBuffer.Data.Runes)
			}
			chunkdata := c.InputBuffer.Data.Runes[offs:end]
			chunkcolor := c.InputBuffer.Colour.Runes[offs:end]

			for i, ch := range chunkdata {
				txt.FGColor = uint64(chunkcolor[i] & 15)
				txt.PutStr(string(ch))
			}

			txt.GotoXYWindow(len(uname)+c.InsertPos-offs, 0)
			txt.FGColor = c.inputFG
		},
	)
}

func (c *ChatClient) showTitleState() {
	c.TitleDo(
		func(txt *types.TextBuffer) {
			txt.FGColor = c.titleFG
			txt.BGColor = c.titleBG
			d := c.channelDetails()
			txt.ClearScreenWindow()
			txt.GotoXYWindow(0, 0)
			txt.FGColor = c.titleFG
			if !c.RemoteMode {
				txt.Printf("%s: %s", d.Name, utils.Unescape(d.Topic))
			} else {
				txt.Printf("%s", d.Name)
			}
			txt.FGColor = c.titleFG
			if c.InScrollback {
				txt.PutStr("\r\n(in scrollback, ESC to exit)")
			}
			txt.FGColor = c.titleFG
			txt.BGColor = c.titleBG
		},
	)
}

func (c *ChatClient) showSidebarState() {
	c.SidebarDo(
		func(txt *types.TextBuffer) {
			d := c.channelDetails()
			txt.ClearScreenWindow()
			txt.GotoXYWindow(0, 0)
			txt.FGColor = c.sidebarFG
			txt.Printf("Users\r\n")
			for i, n := range d.ActiveUsers {
				if i >= 20 {
					txt.Printf(" %s\r\n", "...")
					break
				}
				txt.FGColor = 15
				txt.Printf(" %s\r\n", n)
			}
			for i, n := range d.InactiveUsers {
				if i >= 20 {
					txt.Printf(" %s\r\n", "...")
					break
				}
				txt.FGColor = c.sidebarFG
				txt.Printf(" %s\r\n", n)
			}
			txt.FGColor = c.sidebarFG
			txt.Printf("\r\nChannels\r\n")
			for _, cn := range c.JList {
				if cd, ok := c.Channels[strings.ToLower(cn)]; ok && cd.RemintHost != "" && !c.RemoteMode {
					continue
				}
				txt.FGColor = 15
				txt.Printf(" %s\r\n", cn)
			}
			for _, cn := range c.NJList {
				if cd, ok := c.Channels[strings.ToLower(cn)]; ok && cd.RemintHost != "" && !c.RemoteMode {
					continue
				}
				txt.FGColor = c.sidebarFG
				txt.Printf(" %s\r\n", cn)
			}
			txt.FGColor = c.sidebarFG
		},
	)
}

type GridItem struct {
	FG    int
	BG    int
	ID    int
	Label string
}

func (this *ChatClient) GridChooser(prompt string, items []GridItem, hmargin, vmargin int, itemwidth, itemheight int) int {

	this.keyFeederDisable = true
	defer func() {
		this.keyFeederDisable = false
	}()

	width := 80 - 2*hmargin
	height := 48 - 2*vmargin

	txt := this.txt

	txt.BGColor = 0
	txt.FGColor = 15
	txt.SetWindow(0, 0, 79, 47)

	apple2helpers.TextDrawBox(
		this.Int,
		hmargin,
		vmargin,
		width,
		height,
		prompt,
		true,
		false,
	)

	selected := 0 // index of selected item
	itemsperh := (width / itemwidth) - 1
	itemsperv := (height / itemheight) - 1

	soffset := 0
	maxitems := itemsperh * itemsperv

	var done bool

	fg := txt.FGColor
	bg := txt.BGColor

	for !done {
		// Draw grid
		for i := soffset; i < len(items) && i < soffset+maxitems; i++ {
			item := items[i]
			x := hmargin + (i%itemsperh)*itemwidth + 1
			y := vmargin + ((i-soffset)/itemsperh)*itemheight + 1
			//~ this.Int.SetMemory(36, uint64(x))
			//~ this.Int.SetMemory(37, uint64(y))

			txt.GotoXYWindow(x, y)

			txt.BGColor = bg
			txt.FGColor = fg
			if i == selected {
				txt.PutStr("[")
			} else {
				txt.PutStr(" ")
			}
			txt.BGColor = uint64(item.BG)
			txt.FGColor = uint64(item.FG)
			txt.PutStr(item.Label)
			txt.BGColor = bg
			txt.FGColor = fg
			if i == selected {
				txt.PutStr("]")
			} else {
				txt.PutStr(" ")
			}
		}

		// read a key
		for this.Int.GetMemoryMap().KeyBufferSize(this.Int.GetMemIndex()) == 0 {
			time.Sleep(50 * time.Millisecond)
		}

		ch := this.Int.GetMemoryMap().KeyBufferGetLatest(this.Int.GetMemIndex())

		switch ch {
		case vduconst.CSR_LEFT:
			if selected > 0 {
				selected--
			}
		case vduconst.CSR_RIGHT:
			if selected < len(items)-1 {
				selected++
			}
		case vduconst.CSR_DOWN:
			if selected < len(items)-itemsperh {
				selected += itemsperh
			}
		case vduconst.CSR_UP:
			if selected >= itemsperh {
				selected -= itemsperh
			}
		case 13:
			done = true
			break
		case 27:
			return -1
		}

		// if we are here, correct view
		for soffset > selected {
			soffset -= itemsperh
		}
		for selected >= soffset+maxitems {
			soffset += itemsperh
		}

	}

	return items[selected].ID
}

func (this *ChatClient) GetCharacter(prompt string) int {

	items := make([]GridItem, 0)

	for i := 1024; i < 1024+128; i++ {
		items = append(items,
			GridItem{
				ID:    i,
				FG:    15,
				BG:    0,
				Label: string(rune(i)),
			},
		)
	}

	gi := this.GridChooser(
		prompt,
		items,
		22,
		4,
		3,
		2,
	)

	return gi

}

func (this *ChatClient) GetColor(prompt string) int {

	items := []GridItem{
		GridItem{FG: 15, BG: 0, Label: "  ", ID: 0},
		GridItem{FG: 15, BG: 1, Label: "  ", ID: 1},
		GridItem{FG: 15, BG: 2, Label: "  ", ID: 2},
		GridItem{FG: 15, BG: 3, Label: "  ", ID: 3},
		GridItem{FG: 15, BG: 4, Label: "  ", ID: 4},
		GridItem{FG: 15, BG: 5, Label: "  ", ID: 5},
		GridItem{FG: 15, BG: 6, Label: "  ", ID: 6},
		GridItem{FG: 15, BG: 7, Label: "  ", ID: 7},
		GridItem{FG: 15, BG: 8, Label: "  ", ID: 8},
		GridItem{FG: 15, BG: 9, Label: "  ", ID: 9},
		GridItem{FG: 15, BG: 10, Label: "  ", ID: 10},
		GridItem{FG: 15, BG: 11, Label: "  ", ID: 11},
		GridItem{FG: 15, BG: 12, Label: "  ", ID: 12},
		GridItem{FG: 15, BG: 13, Label: "  ", ID: 13},
		GridItem{FG: 15, BG: 14, Label: "  ", ID: 14},
		GridItem{FG: 15, BG: 15, Label: "  ", ID: 15},
	}

	gi := this.GridChooser(
		prompt,
		items,
		22,
		9,
		4,
		2,
	)

	return gi

}

func (c *ChatClient) Help() {

	text := `

microChat help:
/me <text>         
  Perform action.
/join <channel>    
  Join named channel.
/part              
  Leave current channel.
/quit              
  Quit micro chat.
/topic <message>   
  Set channel topic.
/users
  List active/away users.
/help or /?        
  Show this help.
`

	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
	for _, txt := range strings.Split(text, "\n") {
		c.ScrollBack.Printfr("%s\r", txt)
	}

	c.ScrollBack.GotoBottom()
	c.ScrollBack.Draw(c)
}

func (c *ChatClient) HelpRemint() {

	text := `

microChat help:
/me <text>         
  Perform action.
/quit              
  Quit micro chat.
/topic <message>   
  Set channel topic.
/users
  List active/away users.
/help or /?        
  Show this help.
`

	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
	for _, txt := range strings.Split(text, "\n") {
		c.ScrollBack.Printfr("%s\r", txt)
	}

	c.ScrollBack.GotoBottom()
	c.ScrollBack.Draw(c)
}

func (c *ChatClient) InputClear() {
	c.InputBuffer.Data = runestring.Cast("")
	c.InputBuffer.Colour = runestring.Cast("")
	c.ViewPos = 0
	c.InsertPos = 0
	c.InsColour = 15
	c.showInputState()
}

func (c *ChatClient) InputLeft() {
	if c.InsertPos > 0 {
		c.InsertPos--
		if c.ViewPos > c.InsertPos {
			c.ViewPos = c.InsertPos
		}
	}
	c.showInputState()
}

func (c *ChatClient) InputRight() {
	if c.InsertPos < len(c.InputBuffer.Data.Runes) {
		rw := c.InputWidth - len(s8webclient.CONN.Username+":  ")
		c.InsertPos++
		if c.ViewPos+rw < c.InsertPos {
			c.ViewPos = c.InsertPos - rw
		}
	}
	c.showInputState()
}

func (c *ChatClient) isAway(user string) bool {
	d := c.Channels[strings.ToLower(user)]
	if d == nil {
		return false
	}
	for _, name := range d.InactiveUsers {
		if name == user {
			return true
		}
	}
	return false
}

type HRec struct {
	Data, Colour runestring.RuneString
}

func (this *ChatClient) ProcessHighlight(s runestring.RuneString) HRec {

	r := HRec{}
	r.Data = runestring.NewRuneString()
	r.Colour = runestring.NewRuneString()
	col := rune(this.chatlogFG | (this.chatlogBG << 4))
	bcol := rune(this.chatlogBG)
	shade := rune(0)
	inv := rune(0)
	nextCC := false
	for _, ch := range s.Runes {
		if nextCC {
			col = ch
			nextCC = false
			continue
		}
		if ch == 6 {
			nextCC = true
			continue
		}
		if ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15 {
			col = rune(ch - vduconst.COLOR0)
			continue
		}
		if ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15 {
			bcol = rune(ch - vduconst.BGCOLOR0)
			continue
		}
		if ch >= vduconst.SHADE0 && ch <= vduconst.SHADE7 {
			shade = rune(ch - vduconst.SHADE0)
			continue
		}
		if ch == vduconst.INVERSE_ON {
			if inv == 0 {
				inv = 256
			} else {
				inv = 0
			}
			continue
		}
		r.Data.Append(string(ch))
		r.Colour.Append(string(col | (bcol << 4) | inv | (shade << 16)))
	}
	return r
}

func (this *ChatClient) Recombine(hr HRec) runestring.RuneString {
	rs := runestring.NewRuneString()
	cc := 15
	bc := 0
	shade := 0
	inv := 0

	//////fmt.Printf("HR.Colour = %d, HR.Data = %d\n", len(hr.Colour.Runes), len(hr.Data.Runes))

	for i := 0; i < len(hr.Data.Runes); i++ {
		if cc&15 != int(hr.Colour.Runes[i]&15) {
			rs.AppendSlice([]rune{vduconst.COLOR0 + hr.Colour.Runes[i]&15})
			cc = int(hr.Colour.Runes[i] & 15)
		}
		// if bc&15 != int((hr.Colour.Runes[i]>>4)&15) {
		// 	rs.AppendSlice([]rune{vduconst.BGCOLOR0 + (hr.Colour.Runes[i]>>4)&15})
		// 	bc = int((hr.Colour.Runes[i] >> 4) & 15)
		// }
		if shade&15 != int((hr.Colour.Runes[i]>>16)&7) {
			rs.AppendSlice([]rune{vduconst.SHADE0 + (hr.Colour.Runes[i]>>16)&7})
			shade = int((hr.Colour.Runes[i] >> 16) & 7)
		}
		if inv != int(hr.Colour.Runes[i]&256) {
			rs.AppendSlice([]rune{vduconst.INVERSE_ON})
			inv = int(hr.Colour.Runes[i] & 256)
			////fmt.Printf("Inverse toggled at %d\n", i)
		}
		rs.AppendSlice([]rune{hr.Data.Runes[i]})
	}

	if inv != 0 {
		rs.AppendSlice([]rune{vduconst.INVERSE_ON})
	}

	if cc != 15 {
		rs.AppendSlice([]rune{vduconst.COLOR15})
	}

	if bc != 0 {
		rs.AppendSlice([]rune{vduconst.BGCOLOR0})
	}

	if shade != 0 {
		rs.AppendSlice([]rune{vduconst.SHADE0})
	}

	return rs
}

func (c *ChatClient) isJoined(chatId int32) bool {
	for _, cd := range c.Channels {
		if cd.ChatId == chatId {
			return len(cd.ActiveUsers) > 0
		}
	}
	return false
}

func (c *ChatClient) Run() {

	apple2helpers.MonitorPanel(c.Int, true)
	apple2helpers.TEXTMAX(c.Int)
	c.Init()
	settings.DisableMetaMode[c.Int.GetMemIndex()] = true

	defer func() {
		settings.DisableMetaMode[c.Int.GetMemIndex()] = false
		apple2helpers.MonitorPanel(c.Int, false)
		c.Done()
	}()

	if s8webclient.CONN.Username == "system" {
		apple2helpers.MonitorPanel(c.Int, false)
		apple2helpers.PutStr(c.Int, "\r\nPlease sign in to microLink to use the \r\nchat.\r\n")
		return
	}

	c.running = true
	go c.KeyFeeder()

	c.showChatMessage(
		&chatapi.MessageDetails{
			Creator: "microchat",
			Message: "Connecting to chat...",
		},
	)

	c.fetchChats()
	c.general()
	c.showInputState()

	// join message
	c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
	c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
	c.Console("Welcome to microChat!\r\nType /? for help.\r\n")

	t := time.NewTicker(10 * time.Minute)
	cp := time.NewTicker(1 * time.Second)

	for c.running {
		select {
		case _ = <-cp.C:
			c.CheckPalette()
		case _ = <-t.C:
			// every 10 minutes, refresh the chat list
			c.fetchChats()
		case away := <-c.chataway:
			c.showChatMessage(
				&chatapi.MessageDetails{
					ChatId:   away.ChatId,
					Message:  "is away",
					Creator:  away.Username,
					IsAction: true,
				},
			)
			c.fetchChats()
			c.showInputState()
		case part := <-c.chatpart:
			c.showChatMessage(
				&chatapi.MessageDetails{
					ChatId:   part.ChatId,
					Message:  "has left",
					Creator:  part.Username,
					IsAction: true,
				},
			)
			c.fetchChats()
			c.showInputState()
		case join := <-c.chatjoin:
			c.showChatMessage(
				&chatapi.MessageDetails{
					ChatId:   join.ChatId,
					Message:  "is here",
					Creator:  join.Username,
					IsAction: true,
				},
			)
			c.fetchChats()
			c.showInputState()
		case details := <-c.chatdetails:

			if c.Channels[strings.ToLower(details.Details.Name)] != nil {
				if c.Channels[strings.ToLower(details.Details.Name)].Topic != details.Details.Topic {
					c.showChatMessage(
						&chatapi.MessageDetails{
							ChatId:   details.ChatId,
							Message:  "has changed the topic to: " + details.Details.Topic,
							Creator:  details.User,
							IsAction: true,
						},
					)
				}
			}

			c.Channels[strings.ToLower(details.Details.Name)] = details.Details
			c.showTitleState()
			c.showSidebarState()
			c.showInputState()
		case msg := <-c.message:
			// got a chat message
			//if c.isAway(msg.Creator) {
			c.fetchChats()
			//}
			log.Println(msg)
			c.showChatMessage(msg)
			c.showInputState()
		case command := <-c.command:
			log.Printf("Got a command: verb %s, args %+v", command.Verb, command.Args)
			// got a chat command
			switch strings.ToLower(command.Verb) {
			case "users":
				d := c.Channels[c.CurrentChannel]
				if d != nil {
					c.Console("Users:")
					for _, v := range d.ActiveUsers {
						c.Console("* %s", v)
					}
					for _, v := range d.InactiveUsers {
						c.Console("- %s", v)
					}
				}
			case "away":
				d := c.Channels[c.CurrentChannel]
				if d != nil {
					s8webclient.CONN.AwayChat(d.ChatId)
				}
			case "help", "?":
				c.Help()
			case "topic":
				d := c.channelDetails()
				if d.ChatId > 0 {
					msg, err := s8webclient.CONN.ChangeChatTopic(
						d.ChatId,
						strings.Join(command.Args, " "),
					)
					if err == nil && msg != nil {
						//c.showChatMessage(msg.Message)
						c.InputClear()
						c.showInputState()
					}
				}
			case "me":
				d := c.channelDetails()
				if d.ChatId > 0 {
					msg, err := s8webclient.CONN.PostChatMessage(
						d.ChatId,
						strings.Join(command.Args, " "),
						true,
					)
					if err == nil && msg != nil {
						//c.showChatMessage(msg.Message)
						c.InputClear()
						c.showInputState()
					}
				}
			case "join":
				if len(command.Args) > 0 && !c.RemoteMode {
					channel := command.Args[0]
					_, ok := c.Channels[strings.ToLower(channel)]
					if !ok {
						c.Console("Invalid channel: %s.\r\n", channel)
						c.showInputState()
					} else {
						c.JoinChat(channel)
					}
				}
			case "part":
				if !c.RemoteMode {
					c.LeaveChat(c.CurrentChannel)
				}
			case "quit":
				c.running = false
				break
			default:
				c.ScrollBack.edit.InsBGColor = int(c.chatlogBG)
				c.ScrollBack.edit.InsFGColor = int(c.chatlogFG)
				c.Console("\r\nUnknown command /%s\r\nType /? to see commands.\r\n", strings.ToLower(command.Verb))
				c.showInputState()
			}
			c.showInputState()
		case key := <-c.keyevent:
			log.Println(key)
			// got a key event
			switch key {
			case vduconst.SHIFT_CTRL_G:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				code := c.GetCharacter("Select symbol")
				c.ScrollBack.Draw(c)
				if code != -1 {
					if c.InsertPos >= len(c.InputBuffer.Data.Runes) {
						c.InputBuffer.Data.AppendSlice([]rune{rune(code)})
						c.InputBuffer.Colour.AppendSlice([]rune{rune(c.InsColour)})
						c.InputRight()
					} else {
						// insert midway
						bf := c.InputBuffer.Data.Runes[:c.InsertPos]
						af := c.InputBuffer.Data.Runes[c.InsertPos:]
						newrunes := []rune{}
						newrunes = append(newrunes, bf...)
						newrunes = append(newrunes, rune(code))
						newrunes = append(newrunes, af...)
						c.InputBuffer.Data.Runes = newrunes

						bfc := c.InputBuffer.Colour.Runes[:c.InsertPos]
						afc := c.InputBuffer.Colour.Runes[c.InsertPos:]
						newrunesc := []rune{}
						newrunesc = append(newrunesc, bfc...)
						newrunesc = append(newrunesc, rune(c.InsColour))
						newrunesc = append(newrunesc, afc...)
						c.InputBuffer.Colour.Runes = newrunesc

						c.InputRight()
					}
				}
			case vduconst.SHIFT_CTRL_F:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				code := c.GetColor("Select text color")
				c.ScrollBack.Draw(c)
				if code != -1 {
					// if c.InsertPos >= len(c.InputBuffer.Data.Runes) {
					c.InsColour = code | (2 << 4)
					// c.InputBuffer.AppendSlice([]rune{vduconst.COLOR0 + rune(code)})
					// c.InputRight()
					// } else {
					// 	// insert midway
					// 	bf := c.InputBuffer.Runes[:c.InsertPos]
					// 	af := c.InputBuffer.Runes[c.InsertPos:]
					// 	newrunes := []rune{}
					// 	newrunes = append(newrunes, bf...)
					// 	newrunes = append(newrunes, vduconst.COLOR0+rune(code))
					// 	newrunes = append(newrunes, af...)
					// 	c.InputBuffer.Runes = newrunes
					// 	c.InputRight()
					// }
				}
			case vduconst.CSR_LEFT:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				c.InputLeft()
			case vduconst.CSR_RIGHT:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				c.InputRight()
			case vduconst.CSR_DOWN:
				if c.InScrollback {
					c.ScrollBack.Down()
				}
				c.InScrollback = true
				c.ScrollBack.Draw(c)
				c.showTitleState()
				c.showInputState()
			case vduconst.CSR_UP:
				if c.InScrollback {
					c.ScrollBack.Up()
				}
				c.InScrollback = true
				c.ScrollBack.Draw(c)
				c.showTitleState()
				c.showInputState()
			case 27:
				c.InScrollback = false
				c.ScrollBack.GotoBottom()
				c.ScrollBack.Draw(c)
				c.showTitleState()
				c.showInputState()
				break
			case 13:
				// check and send message to chat
				if len(c.InputBuffer.Data.Runes) > 0 {
					if c.InScrollback {
						c.InScrollback = false
						c.ScrollBack.GotoBottom()
						c.ScrollBack.Draw(c)
						c.showTitleState()
						c.showInputState()
					}
					if c.InputBuffer.Data.Runes[0] == '/' {
						c.processChatCommand(utils.Escape(string(c.Recombine(c.InputBuffer).Runes)))
						c.InputClear()
						c.showInputState()
					} else {
						d := c.channelDetails()
						if d.ChatId > 0 {
							etext := utils.Escape(string(c.Recombine(c.InputBuffer).Runes))
							msg, err := s8webclient.CONN.PostChatMessage(
								d.ChatId,
								etext,
								false,
							)
							if err == nil && msg != nil {
								//c.showChatMessage(msg.Message)
							}
							c.InputClear()
							c.showInputState()
						}
					}
				}
			case 127:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				if len(c.InputBuffer.Data.Runes) > 0 {
					if c.InsertPos >= len(c.InputBuffer.Data.Runes) {
						c.InputBuffer.Data.Runes = c.InputBuffer.Data.Runes[:len(c.InputBuffer.Data.Runes)-1]
						c.InputBuffer.Colour.Runes = c.InputBuffer.Colour.Runes[:len(c.InputBuffer.Colour.Runes)-1]
						c.InputLeft()
					} else if c.InsertPos > 0 {
						bf := c.InputBuffer.Data.Runes[:c.InsertPos-1]
						af := c.InputBuffer.Data.Runes[c.InsertPos:]
						newrunes := []rune{}
						newrunes = append(newrunes, bf...)
						newrunes = append(newrunes, af...)
						c.InputBuffer.Data.Runes = newrunes

						bfc := c.InputBuffer.Colour.Runes[:c.InsertPos-1]
						afc := c.InputBuffer.Colour.Runes[c.InsertPos:]
						newrunesc := []rune{}
						newrunesc = append(newrunesc, bfc...)
						newrunesc = append(newrunesc, afc...)
						c.InputBuffer.Colour.Runes = newrunesc

						c.InputLeft()
					}
				}
			default:
				if c.InScrollback {
					c.InScrollback = false
					c.ScrollBack.GotoBottom()
					c.ScrollBack.Draw(c)
					c.showTitleState()
					c.showInputState()
				}
				if ((key >= 32 && key < 127) || (key >= 1024 && key <= 1024+127)) && len(c.InputBuffer.Data.Runes) < c.InputMax {
					// we need to insert at insert pos
					if c.InsertPos >= len(c.InputBuffer.Data.Runes) {
						c.InputBuffer.Data.AppendSlice([]rune{key})
						c.InputBuffer.Colour.AppendSlice([]rune{rune(c.InsColour)})
						c.InputRight()
					} else {
						// insert midway
						bf := c.InputBuffer.Data.Runes[:c.InsertPos]
						af := c.InputBuffer.Data.Runes[c.InsertPos:]
						newrunes := []rune{}
						newrunes = append(newrunes, bf...)
						newrunes = append(newrunes, key)
						newrunes = append(newrunes, af...)
						c.InputBuffer.Data.Runes = newrunes

						bfc := c.InputBuffer.Colour.Runes[:c.InsertPos]
						afc := c.InputBuffer.Colour.Runes[c.InsertPos:]
						newrunesc := []rune{}
						newrunesc = append(newrunesc, bfc...)
						newrunesc = append(newrunesc, rune(c.InsColour))
						newrunesc = append(newrunesc, afc...)
						c.InputBuffer.Colour.Runes = newrunesc

						c.InputRight()
					}
				}
			}
			c.showInputState()
		}
	}

}

func (c *ChatClient) processChatCommand(buffer string) {
	buffer = buffer[1:]
	parts := strings.Split(buffer, " ")
	c.command <- ChatCommand{
		Verb: parts[0],
		Args: parts[1:],
	}
}

func (cc *ChatClient) HandleMessage(c *client.DuckTapeClient, msg *ducktape.DuckTapeBundle) {
	log.Printf("Got message type: %s", msg.ID)
	switch msg.ID {
	case "KAB":
		resp := &chatapi.BroadcastChatLeave{}
		err := proto.Unmarshal(msg.Payload, resp)
		if err == nil && cc.isJoined(resp.ChatId) {
			cc.chataway <- resp
		}
	case "KLB":
		resp := &chatapi.BroadcastChatLeave{}
		err := proto.Unmarshal(msg.Payload, resp)
		if err == nil && cc.isJoined(resp.ChatId) && resp.Username != s8webclient.CONN.Username {
			cc.chatpart <- resp
		}
	case "KJB":
		resp := &chatapi.BroadcastChatJoin{}
		err := proto.Unmarshal(msg.Payload, resp)
		if err == nil && cc.isJoined(resp.ChatId) && resp.Username != s8webclient.CONN.Username {
			cc.chatjoin <- resp
		}
	case "KTB":
		resp := &chatapi.BroadcastChatDetails{}
		err := proto.Unmarshal(msg.Payload, resp)
		if err == nil && cc.isJoined(resp.ChatId) {
			cc.chatdetails <- resp
		}
	case "KMB":
		resp := &chatapi.BroadcastChatMessage{}
		err := proto.Unmarshal(msg.Payload, resp)
		if err == nil && cc.isJoined(resp.ChatId) {
			cc.message <- resp.Message
		}
	}
}
