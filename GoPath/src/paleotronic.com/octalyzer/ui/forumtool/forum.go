package forumtool

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"paleotronic.com/api"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/server/forumapi"
	"paleotronic.com/utils"
)

const (
	fcHeading   = 13
	fcPostTitle = 14
	fcPostText  = 7
	fcPostQuote = 6
	fcMenuText  = 12
	fcPrompt    = 15
)

type ForumApp struct {
	CurrentForumID  int32
	CurrentParentID int32
	Int             interfaces.Interpretable
	Forums          map[int32]*forumapi.ForumDetails //*forumapi.FetchForumsResponse
	Topics          *forumapi.FetchMessagesResponse
	Title           string
	Footer          string
	FullRefresh     bool
	Edit            *editor.CoreEdit
	running         bool
	txt             *types.TextBuffer
	UnreadMap       map[int32][]int32
	edit            *editor.CoreEdit
	save            bool
	threaded        bool
	desc            bool
}

func NewForumApp(e interfaces.Interpretable, forumId int32) *ForumApp {
	txt := apple2helpers.GETHUD(e, "TEXT")
	f := &ForumApp{
		CurrentForumID: forumId,
		Int:            e,
		txt:            txt.Control,
	}
	f.fetchForums()
	return f
}

func (f *ForumApp) fetchForums() error {
	m, err := s8webclient.CONN.FetchForums()
	if err == nil {
		f.Forums = make(map[int32]*forumapi.ForumDetails)
		log.Printf("Got forum details: %+v", m)
		for _, ff := range m.Forums {
			f.Forums[ff.ForumId] = ff
		}
	}
	return err
}

func (f *ForumApp) fetchTopics() error {
	if f.CurrentForumID == 0 {
		f.CurrentForumID = 1
	}
	m, err := s8webclient.CONN.FetchForumMessages(f.CurrentForumID, f.CurrentParentID)
	if err == nil {
		f.Topics = m
	}
	return err
}

func (f *ForumApp) fetchUnreadMessages() error {
	if f.Forums == nil {
		f.fetchForums()
		if f.Forums == nil {
			return errors.New("No forums")
		}
	}
	f.UnreadMap = make(map[int32][]int32)
	for _, ff := range f.Forums {
		id := ff.ForumId
		r, err := s8webclient.CONN.FetchForumUnread(id)
		if err == nil {
			f.UnreadMap[id] = r.MessageIds
		}
	}
	return nil
}

func elipsisTrimPad(text string, max int) string {
	if len(text) > max {
		return text[:max-3] + "..."
	}
	for len(text) < max {
		text += " "
	}
	return text
}

func (this *ForumApp) GetCRTLine(promptString string) string {

	command := ""
	collect := true
	display := this.Int

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	display.PutStr(promptString)

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	for collect {
		if cb != nil {
			cb(this.Int.GetMemIndex())
		}
		apple2helpers.TextShowCursor(this.Int)
		for this.Int.GetMemory(49152) < 128 {
			time.Sleep(5 * time.Millisecond)
			if cb != nil {
				cb(this.Int.GetMemIndex())
			}
			if this.Int.VM().IsDying() {
				return ""
			}
		}
		apple2helpers.TextHideCursor(this.Int)
		ch := rune(this.Int.GetMemory(49152) & 0xff7f)
		this.Int.SetMemory(49168, 0)
		switch ch {
		case 10:
			{
				display.SetSuppressFormat(true)
				display.PutStr("\r\n")
				display.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				display.SetSuppressFormat(true)
				display.PutStr("\r\n")
				display.SetSuppressFormat(false)
				return command
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					display.Backspace()
					display.SetSuppressFormat(true)
					display.PutStr(" ")
					display.SetSuppressFormat(false)
					display.Backspace()
					if cb != nil {
						cb(this.Int.GetMemIndex())
					}
				}
				break
			}
		default:
			{

				display.SetSuppressFormat(true)
				display.RealPut(rune(ch))
				display.SetSuppressFormat(false)

				if cb != nil {
					cb(this.Int.GetMemIndex())
				}

				command = command + string(ch)
				break
			}
		}
	}

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	return command

}

type MenuOption struct {
	Key         string
	Description string
	Hide        bool
}

func (f *ForumApp) optionMenu(
	title string,
	options []MenuOption,
	showOptions bool,
	promptString string,
	showKeys bool,
	defKey string,
	showHelp bool,
	dblSpace bool,
) string {

	if showHelp {
		options = append(options, MenuOption{"?", "Show help", true})
	}

	f.txt.PutStr("\r\n")

	if showOptions {
		f.txt.FGColor = fcHeading
		f.txt.PutStr(title + "\r\n")
	}

showHelp:

	f.txt.PutStr("\r\n")

	keys := make([]string, len(options))
	valid := make(map[string]MenuOption)

	for i, option := range options {
		if showOptions && !option.Hide {
			f.txt.FGColor = fcMenuText
			f.txt.PutStr(
				fmt.Sprintf(
					"[%2s] %s\r\n",
					option.Key,
					option.Description,
				),
			)
			if dblSpace {
				f.txt.PutStr("\r\n")
			}
		}
		keys[i] = option.Key
		valid[option.Key] = option
		valid[strings.ToLower(option.Key)] = option
	}

prompt:

	f.txt.PutStr("\r\n")

	f.txt.FGColor = fcPrompt
	rprompt := promptString
	if showKeys {
		rprompt += "(" + strings.Join(keys, ",") + ") "
		if defKey != "" {
			rprompt += "[" + defKey + "]"
		}
	}
	rprompt += ": "
	resp := f.GetCRTLine(rprompt)
	if resp == "" {
		resp = defKey
	}
	_, ok := valid[resp]
	if !ok {
		goto prompt
	}
	if resp == "?" && showHelp {
		goto showHelp
	}

	return resp

}

func (f *ForumApp) checkUnreadMessages() {
	err := f.fetchUnreadMessages()
	if err != nil {
		f.txt.FGColor = fcPrompt
		f.txt.PutStr("No unread messages.\r\n")
		return
	}
	for id, messages := range f.UnreadMap {
		f.txt.FGColor = fcPrompt
		f.txt.PutStr(fmt.Sprintf("Forum '%s' has %d unread messages\r\n", f.Forums[id].Name, len(messages)))
		f.fetchUnreadForForum(id)
	}
}

func (f *ForumApp) fetchUnreadForForum(forumId int32) ForumAction {
	//f.fetchUnreadMessages()

	ids, ok := f.UnreadMap[forumId]
	if !ok {
		return forumNextForum
	}

	for _, messageId := range ids {
		msg, err := s8webclient.CONN.FetchMessage(forumId, messageId)
		if err != nil {
			f.txt.FGColor = fcPrompt
			f.txt.PutStr(fmt.Sprintf("Unable to fetch messages (%s)\r\n", err.Error()))
			return forumNextForum
		}
		action := f.displayMessagePaged(msg.Message, f.pagingPromptScan, f.eomPromptScan)
		switch action {
		case forumNextForum:
			return action
		case forumContinue:
			// loop will continue
		}
	}

	return forumNextForum

}

func getTimeDifference(t time.Time) string {
	now := time.Now()
	desc := "from now"
	diff := now.Sub(t)
	if diff > 0 {
		desc = "ago"
	}
	var units string
	switch {
	case diff > 7*24*time.Hour:
		units = "week"
		diff = diff / (7 * 24 * time.Hour)
	case diff > 24*time.Hour:
		units = "day"
		diff = diff / (24 * time.Hour)
	case diff > time.Hour:
		units = "hour"
		diff = diff / time.Hour
	case diff > time.Minute:
		units = "minute"
		diff = diff / time.Minute
	default:
		units = "second"
		diff = diff / time.Second
	}

	if diff > 1 {
		units += "s"
	}

	return fmt.Sprintf("%d %s %s", diff, units, desc)
}

func wrap(text string, lineWidth int) (allWrapped string) {

	lines := strings.Split(text, "\n")

	allWrapped = ""
	for _, text = range lines {
		words := strings.Fields(text)
		if len(words) == 0 {
			continue
		}
		wrapped := words[0]
		spaceLeft := lineWidth - len(wrapped)
		for _, word := range words[1:] {
			if len(word)+1 > spaceLeft {
				wrapped += "\n" + word
				spaceLeft = lineWidth - len(word)
			} else {
				wrapped += " " + word
				spaceLeft -= 1 + len(word)
			}
		}
		allWrapped += "\n"
		allWrapped += wrapped
		spaceLeft = lineWidth
	}
	return
}

func (this *ForumApp) GetCRTKey() rune {

	command := rune(0)

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}
	apple2helpers.TextShowCursor(this.Int)
	for this.Int.GetMemory(49152) < 128 {
		time.Sleep(5 * time.Millisecond)
		if cb != nil {
			cb(this.Int.GetMemIndex())
		}
		if this.Int.VM().IsDying() {
			return command
		}
	}
	apple2helpers.TextHideCursor(this.Int)
	command = rune(this.Int.GetMemory(49152) & 0xff7f)
	this.Int.SetMemory(49168, 0)

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	return command

}

func (f *ForumApp) eomPromptMsg() []ForumAction {
	f.txt.FGColor = fcPrompt

	var key string

rePrompt:
	f.txt.FGColor = fcPrompt

	prompt := "[RVLPTQ or ?]: "

	f.txt.PutStr("\r\n")
	//f.txt.PutStr(prompt)
	key = f.GetCRTLine(prompt)

	switch key {
	case "?":
		f.txt.FGColor = fcMenuText
		f.txt.PutStr("\r\n")
		f.txt.PutStr("[R]eply to message\r\n")
		f.txt.PutStr("[V]iew replies\r\n")
		f.txt.PutStr("[L]ike message\r\n")
		f.txt.PutStr("[P]arent of message\r\n")
		f.txt.PutStr("[T]hread\r\n")
		f.txt.PutStr("[Q]uit to post list\r\n")
		goto rePrompt
	case "Q", "q":
		return []ForumAction{forumQuitScan}
	case "R", "r":
		return []ForumAction{forumReplyMessage}
	case "V", "v":
		return []ForumAction{forumViewReplies}
	case "L", "l":
		return []ForumAction{forumLikeMessage}
	case "P", "p":
		return []ForumAction{forumViewParent}
	case "T", "t":
		return []ForumAction{forumContinue}
	default:
		return []ForumAction{forumContinue}
	}
}

type EOMPromptFunc func() ForumAction
type PAGEPromptFunc func() ForumAction

func (f *ForumApp) eomPromptScan() []ForumAction {
	var key string

rePrompt:
	f.txt.FGColor = fcPrompt

	prompt := "[NBRLPTSMAQ or ?]: "

	f.txt.PutStr("\r\n")
	//f.txt.PutStr(prompt)
	key = f.GetCRTLine(prompt)

	switch key {
	case "?":
		f.txt.FGColor = fcMenuText
		f.txt.PutStr("\r\n")
		f.txt.PutStr("[N]ext message\r\n")
		f.txt.PutStr("[B]ack to previous message\r\n")
		f.txt.PutStr("[R]eply to message\r\n")
		f.txt.PutStr("[V]iew replies\r\n")
		f.txt.PutStr("[L]ike message\r\n")
		f.txt.PutStr("[P]arent of message\r\n")
		f.txt.PutStr("[T]hread\r\n")
		f.txt.PutStr("[S]kip remaining messages in forum and go to next forum\r\n")
		f.txt.PutStr("[M]ark all messages in forum read and go to next forum\r\n")
		f.txt.PutStr("[A]ll messages in all forums marked read and quit\r\n")
		f.txt.PutStr("[Q]uit new message scan\r\n")
		goto rePrompt
	case "Q", "q":
		return []ForumAction{forumQuitScan}
	case "N", "n":
		return []ForumAction{forumNextMessage}
	case "B", "b":
		return []ForumAction{forumPrevMessage}
	case "R", "r":
		return []ForumAction{forumReplyMessage}
	case "V", "v":
		return []ForumAction{forumViewReplies}
	case "L", "l":
		return []ForumAction{forumLikeMessage}
	case "P", "p":
		return []ForumAction{forumShowParentMessage}
	case "T", "t":
		return []ForumAction{forumContinue}
	case "S", "s":
		return []ForumAction{forumNextForum}
	case "M", "m":
		return []ForumAction{forumMarkAllInForumRead, forumNextForum}
	default:
		return []ForumAction{forumContinue}
	}
}

func (f *ForumApp) pagingPromptScan() []ForumAction {

	var key string

rePrompt:
	f.txt.FGColor = fcPrompt

	prompt := "[CMSNQ or ?]: "

	f.txt.PutStr("\r\n")
	//f.txt.PutStr(prompt)
	key = f.GetCRTLine(prompt)

	switch key {
	case "?":
		f.txt.FGColor = fcMenuText
		f.txt.PutStr("\r\n")
		f.txt.PutStr("[C]ontinue reading (or Enter)\r\n")
		f.txt.PutStr("[M]ark read and go to next message\r\n")
		f.txt.PutStr("[S]top reading and go to prompt\r\n")
		f.txt.PutStr("[N]ext message without marking read\r\n")
		f.txt.PutStr("[Q]uit new message scan\r\n")
		goto rePrompt
	case "", "C", "c":
		return []ForumAction{forumContinue}
	case "M", "m":
		return []ForumAction{forumMarkRead, forumNextMessage}
	case "S", "s":
		return []ForumAction{forumQuitScan}
	case "N", "n":
		return []ForumAction{forumNextMessage}
	case "Q", "q":
		return []ForumAction{forumQuitScan}
	default:
		goto rePrompt
	}

}

func (f *ForumApp) displayMessagePaged(msg *forumapi.MessageDetails, pf func() []ForumAction, ef func() []ForumAction) ForumAction {

reDoAll:
	f.txt.PutStr("\r\n")
	f.txt.FGColor = fcPostTitle
	f.txt.PutStr(msg.Subject + "\r\n")
	f.txt.FGColor = fcPrompt
	f.txt.PutStr("Posted by " + msg.Creator + " " + getTimeDifference(time.Unix(0, msg.Created)) + " in " + f.Forums[msg.ForumId].Name + "\r\n")

	lpp := 40
	wdth := 72
	body := strings.Split(wrap(msg.Body, wdth), "\n")

	for i, v := range body {
		if i > 0 && i%lpp == 0 {

			actions := pf()

			for _, action := range actions {
				switch action {
				case forumMarkRead:
					f.markMessageRead(msg.ForumId, msg.MessageId)
				case forumQuitScan:
					return action
				case forumNextMessage:
					return action
				case forumPrevMessage:
					return action
				}
			}
		}
		f.txt.FGColor = fcPostText
		if strings.HasPrefix(v, "> ") {
			f.txt.FGColor = fcPostQuote
		}
		f.txt.PutStr(v + "\r\n")
	}

	// got to here -- message menu
	f.txt.FGColor = fcPrompt
	f.txt.PutStr("\r\nEnd of message\r\n\r\n")

	_ = f.markMessageRead(msg.ForumId, msg.MessageId)

showMenu:
	f.txt.FGColor = fcPrompt

	actions := ef()
	for _, action := range actions {
		switch action {
		case forumViewParent:
			log.Printf("View parent requested for message %d", msg.MessageId)
			if msg.ParentId < 1 {
				f.txt.FGColor = fcPrompt
				f.txt.PutStr("\r\nNo parent message.")
				return forumContinue
			}
			resp, err := s8webclient.CONN.FetchMessage(msg.ForumId, msg.ParentId)
			if err != nil {
				f.txt.FGColor = fcPrompt
				f.txt.PutStr("\r\nFailed to fetch parent message.")
				return forumContinue
			}
			msg = resp.Message
			goto reDoAll
		case forumViewReplies:
			posts, err := s8webclient.CONN.FetchForumMessages(msg.ForumId, msg.MessageId)
			log.Printf("View replies requested for message %d", msg.MessageId)
			if err != nil {
				f.txt.FGColor = fcPrompt
				f.txt.PutStr("\r\nFailed to fetch replies.")
				return forumContinue
			}
			if len(posts.Messages) == 0 {
				f.txt.FGColor = fcPrompt
				f.txt.PutStr("\r\nNo replies.")
				return forumContinue
			}
			f.displayPostList(msg.ForumId, msg.MessageId, posts.Messages)
		case forumReplyMessage:
			subject := msg.Subject
			if !strings.HasPrefix(subject, "Re: ") {
				subject = "Re: " + subject
			}

			f.txt.FGColor = fcPrompt
			qr := f.GetCRTLine("Quote message in reply [Y/n]? ")
			var err error
			var messageText string
			if strings.ToLower(qr) == "y" || qr == "" {
				messageText, err = f.editMessage(msg.ForumId, msg.MessageId, quote(msg.Body))
			} else {
				messageText, err = f.editMessage(msg.ForumId, msg.MessageId, "")
			}
			if err == nil {
				_, err := s8webclient.CONN.PostMessage(msg.ForumId, msg.MessageId, subject, messageText)
				if err != nil {
					f.txt.PutStr(fmt.Sprintf("Failed to post reply: %s\r\n", err.Error()))
					goto showMenu
				}
				f.txt.PutStr("Posted reply.\r\n")
			}
		default:
			//
		}
	}

	return forumContinue
}

func quote(text string) string {
	return strings.Join(strings.Split(wrap(text, 72), "\n"), "\n> ")
}

func (f *ForumApp) markMessageRead(forumId int32, messageId int32) error {
	log.Printf("Marking message read")
	_, err := s8webclient.CONN.MarkMessageRead(forumId, messageId)
	return err
}

func (f *ForumApp) displayPostList(forumId int32, parentId int32, list []*forumapi.MessageDetails) ForumAction {

	var m = make(map[string]*forumapi.MessageDetails)
	var err error

showForumMessages:

	r := NewRootNode()
	for _, msg := range list {
		_, err = r.PlaceMsg(msg)
		if err != nil {
			f.txt.FGColor = fcPrompt
			f.txt.PutStr(err.Error() + "\r\n")
			break
		}
	}
	r.SetDesc(f.desc)
	messages := r.GetAll()

	// we get threaded order by default, so we need to resort them
	if !f.threaded {
		if f.desc {
			sort.Sort(byTimeDesc(messages))
		} else {
			sort.Sort(byTimeAsc(messages))
		}
	}

	level := 0
	//lastId := int32(0)

	f.txt.PutStr("\r\n")

	// var lastParentId int32 = 0
	for i, p := range messages {
		// capture indentlevel for post
		if f.threaded {
			level = p.Generation() - 1
		}
		m[fmt.Sprintf("%d", i+1)] = p.Message
		f.txt.FGColor = fcPostTitle
		f.txt.PutStr(
			fmt.Sprintf(
				"%-8d %-16s %s%-30s\r\n",
				i+1,
				p.Message.Creator,
				spaces(level),
				elipsisTrimPad(p.Message.Subject, 30),
			),
		)

	}

	f.txt.PutStr("\r\n")

menu:
	f.txt.FGColor = fcPrompt
	option := f.forumMenu(forumId, m)
	switch option {
	case "A", "a":
		f.desc = false
		goto showForumMessages
	case "D", "d":
		f.desc = true
		goto showForumMessages
	case "T", "t":
		f.threaded = !f.threaded
		goto showForumMessages
	case "P", "p":
		subject := f.GetCRTLine("Enter Subject: ")
		messageText, err := f.editMessage(forumId, parentId, "")
		if err == nil {
			_, err := s8webclient.CONN.PostMessage(forumId, parentId, subject, messageText)
			if err != nil {
				f.txt.PutStr(fmt.Sprintf("Failed to post message: %s\r\n", err.Error()))
				goto menu
			}
			f.txt.FGColor = fcPrompt
			f.txt.PutStr("Posted message.\r\n")
			// refresh list
			posts, err := s8webclient.CONN.FetchForumMessages(forumId, parentId)
			log.Printf("Messages fetched: %v, %v", posts, err)
			if err == nil {
				list = posts.Messages
			}
			goto showForumMessages
		}
		goto menu
	case "Q", "q":
		return forumContinue
	case "L", "l":
		posts, err := s8webclient.CONN.FetchForumMessages(forumId, parentId)
		log.Printf("Messages fetched: %v, %v", posts, err)
		if err == nil {
			list = posts.Messages
		}
		goto showForumMessages
	case "N", "n":
		f.fetchUnreadForForum(forumId)
	case "S", "s":
		f.txt.FGColor = fcPrompt
		term := f.GetCRTLine("Search for: ")
		res, err := s8webclient.CONN.SearchForumMessages(forumId, term)
		if err != nil {
			f.txt.PutStr(fmt.Sprintf("Failed to search: %s\r\n", err.Error()))
			goto menu
		}
		list = res.Messages
		goto showForumMessages
	default:
		msg := m[option]
		if msg != nil {
			act := f.displayMessagePaged(msg, f.pagingPromptScan, f.eomPromptMsg)
			switch act {
			case forumViewReplies:
				posts, err := s8webclient.CONN.FetchForumMessages(forumId, msg.MessageId)
				log.Printf("Messages fetched: %v, %v", posts, err)
				if err == nil {
					f.displayPostList(forumId, msg.MessageId, posts.Messages)
				}
				goto showForumMessages
			case forumViewParent:
				pid := int32(0)
				if msg.ParentId > 0 {
					pid = msg.ParentId
				}
				posts, err := s8webclient.CONN.FetchForumMessages(forumId, pid)
				log.Printf("Messages fetched: %v, %v", posts, err)
				if err == nil {
					list = posts.Messages
				}
				goto showForumMessages
			}
			goto showForumMessages
		} else {
			posts, err := s8webclient.CONN.FetchForumMessages(forumId, parentId)
			log.Printf("Messages fetched: %v, %v", posts, err)
			if err == nil {
				list = posts.Messages
			}
			goto showForumMessages
		}
		goto menu
	}

	return forumContinue
}

func (f *ForumApp) displayForumPosts(forumId int32, parentId int32) ForumAction {

	ff, ok := f.Forums[int32(forumId)]
	if !ok {
		return forumContinue
	}

	f.txt.FGColor = fcHeading
	f.txt.PutStr("\r\n" + strings.ToUpper(ff.Name) + " FORUM\r\n")

	f.CurrentForumID = forumId

	posts, err := s8webclient.CONN.FetchForumMessages(forumId, parentId)
	if err != nil {
		f.txt.FGColor = fcPrompt
		f.txt.PutStr("No messages\r\n")
		return forumContinue
	}

	return f.displayPostList(forumId, parentId, posts.Messages)
}

func spaces(size int) string {
	out := ""
	for i := 0; i < size; i++ {
		out += " "
	}
	return out
}

func (f *ForumApp) forumMenu(forumId int32, m map[string]*forumapi.MessageDetails) string {
	var key string

rePrompt:
	f.txt.FGColor = fcPrompt
	prompt := "[ADTLMNSP, # or  ? for Help]: "

	f.txt.PutStr("\r\n")
	key = f.GetCRTLine(prompt)

	switch {
	case key == "?":
		f.txt.FGColor = fcMenuText
		f.txt.PutStr("\r\n")
		f.txt.PutStr("[A]scending order by date/time\r\n")
		f.txt.PutStr("[D]escending order by date/time\r\n")
		f.txt.PutStr("[T]hreaded list view\r\n")
		f.txt.PutStr("[L]ist unread messages and choose by number\r\n")
		f.txt.PutStr("[M]ark all messages read\r\n")
		f.txt.PutStr("[N]ew messages in forum\r\n")
		f.txt.PutStr("[S]earch for messages\r\n")
		f.txt.PutStr("[P]ost a new message\r\n")
		f.txt.PutStr("[Q]uit forum and return to main menu\r\n")
		goto rePrompt
	case key == "A" || key == "a":
		return key
	case key == "D" || key == "d":
		return key
	case key == "T" || key == "t":
		return key
	case key == "L" || key == "l":
		return key
	case key == "M" || key == "m":
		return key
	case key == "N" || key == "n":
		return key
	case key == "S" || key == "s":
		return key
	case key == "P" || key == "p":
		return key
	case key == "Q" || key == "q":
		return key
	case m[key] != nil:
		return key
	default:
		return "L"
	}
}

func (f *ForumApp) mainMenu() {

	options := []MenuOption{
		MenuOption{"N", "NEW MESSAGES   Shows new messages in ALL forums", true},
	}

	if f.Forums != nil && len(f.Forums) > 0 {
		keys := make([]int, 0)
		for _, ff := range f.Forums {
			keys = append(keys, int(ff.ForumId))
		}
		sort.Ints(keys)

		i := 0
		for _, forumId := range keys {
			ff := f.Forums[int32(forumId)]
			options = append(
				options,
				MenuOption{
					fmt.Sprintf("%d", i+1),
					fmt.Sprintf("%-20s %-30s", strings.ToUpper(ff.GetName()), ff.GetDescription()),
					false,
				},
			)
			i++
			keys = append(keys, int(ff.ForumId))
		}
	}

	options = append(
		options,
		MenuOption{
			"Q",
			fmt.Sprintf("%-20s %-30s", "QUIT", "Exit the forums"),
			true,
		},
	)

	resp := f.optionMenu(
		"microM8 Forums",
		options,
		true,
		"Choose forum by # or type [N] for new messages",
		false,
		"",
		true,
		true,
	)

	log.Printf("Got resp: %s", resp)

	switch resp {
	case "N", "n":
		f.checkUnreadMessages()
	case "Q", "q":
		f.running = false
	default:
		ival := utils.StrToInt(resp)
		log.Printf("Viewing forum id %d", ival)
		f.displayForumPosts(int32(ival), -1)
	}

}

func (f *ForumApp) Run() {

	if s8webclient.CONN.Username == "system" {
		f.Int.PutStr("Please sign in to microLink to use the\r\nforums.\r\n")
		return
	}

	apple2helpers.TEXTMAX(f.Int)
	apple2helpers.Clearscreen(f.Int)
	f.running = true
	for f.running {
		f.mainMenu()
	}
}
