package d2player

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

const white = 0xffffffff

const ( // for the dc6 frames
	questLogTopLeft = iota
	questLogTopRight
	questLogBottomLeft
	questLogBottomRight
)

const (
	questLogOffsetX, questLogOffsetY = 80, 64
)

const (
	iconOffsetY                = 88
	questOffsetX, questOffsetY = 4, 4
	q1SocketX, q1SocketY       = 100, 95
	q2SocketX, q2SocketY       = 200, 95
	q3SocketX, q3SocketY       = 300, 95
	q4SocketX, q4SocketY       = 100, 190
	q5SocketX, q5SocketY       = 200, 190
	q6SocketX, q6SocketY       = 300, 190
)

const (
	questLogCloseButtonX, questLogCloseButtonY = 358, 455
	questLogDescrButtonX, questLogDescrButtonY = 308, 457
	questNameLabelX, questNameLabelY           = 240, 297
	questDescrLabelX, questDescrLabelY         = 90, 317
)

// toset
const (
	questTabY  = 66
	questTab1X = 85
	questTab2X = 143
	questTab3X = 201
	questTab4X = 259
	questTab5X = 317
)

func (s *QuestLog) getPositionForSocket(number int) (x, y int) {
	pos := []struct {
		x int
		y int
	}{
		{q1SocketX, q1SocketY},
		{q2SocketX, q2SocketY},
		{q3SocketX, q3SocketY},
		{q4SocketX, q4SocketY},
		{q5SocketX, q5SocketY},
		{q6SocketX, q6SocketY},
	}

	return pos[number].x, pos[number].y
}

// NewQuestLog creates a new quest log
func NewQuestLog(asset *d2asset.AssetManager,
	ui *d2ui.UIManager,
	l d2util.LogLevel,
	act int) *QuestLog {
	originX := 0
	originY := 0

	//nolint:gomnd // this is only test
	qs := map[int]int{
		0:  -2,
		1:  -2,
		2:  -1,
		3:  0,
		4:  1,
		5:  2,
		6:  3,
		7:  0,
		8:  0,
		9:  0,
		10: 0,
		11: 0,
		12: 0,
		13: 0,
		14: 0,
		15: 0,
		16: 0,
		17: 0,
		18: 0,
		19: 0,
		20: 0,
		21: 0,
		22: 0,
		23: 0,
		24: 0,
		25: 0,
		26: 0,
	}

	var quests [d2enum.ActsNumber]*d2ui.WidgetGroup
	for i := 0; i < d2enum.ActsNumber; i++ {
		quests[i] = ui.NewWidgetGroup(d2ui.RenderPriorityQuestLog)
	}

	ql := &QuestLog{
		asset:     asset,
		uiManager: ui,
		originX:   originX,
		originY:   originY,
		act:       act,
		tab: [d2enum.ActsNumber]*questLogTab{
			{},
			{},
			{},
			{},
			{},
		},
		quests:      quests,
		questStatus: qs,
	}

	ql.Logger = d2util.NewLogger()
	ql.Logger.SetLevel(l)
	ql.Logger.SetPrefix(logPrefix)

	return ql
}

// QuestLog represents the quest log
type QuestLog struct {
	asset         *d2asset.AssetManager
	uiManager     *d2ui.UIManager
	panel         *d2ui.Sprite
	onCloseCb     func()
	panelGroup    *d2ui.WidgetGroup
	selectedTab   int
	selectedQuest int
	act           int
	tab           [d2enum.ActsNumber]*questLogTab

	questName   *d2ui.Label
	questDescr  *d2ui.Label
	quests      [d2enum.ActsNumber]*d2ui.WidgetGroup
	questStatus map[int]int

	originX int
	originY int
	isOpen  bool

	*d2util.Logger
}

/* questIconTab returns path to quest animation using its
act and number. From d2resource:
        QuestLogAQuestAnimation = "/data/global/ui/MENU/a%dq%d.dc6"
*/
func (s *QuestLog) questIconsTable(act, number int) string {
	return fmt.Sprintf(d2resource.QuestLogAQuestAnimation, act, number+1)
}

const (
	completedFrame  = 24
	inProgresFrame  = 25
	notStartedFrame = 26
)

const questDescriptionLenght = 30

type questLogTab struct {
	button          *d2ui.Button
	invisibleButton *d2ui.Button
}

func (q *questLogTab) newTab(ui *d2ui.UIManager, tabType d2ui.ButtonType, x int) {
	q.button = ui.NewButton(tabType, "")
	q.invisibleButton = ui.NewButton(d2ui.ButtonTypeTabBlank, "")
	q.button.SetPosition(x, questTabY)
	q.invisibleButton.SetPosition(x, questTabY)
}

// Load the data for the hero status panel
func (s *QuestLog) Load() {
	var err error

	s.panelGroup = s.uiManager.NewWidgetGroup(d2ui.RenderPriorityQuestLog)

	frame := d2ui.NewUIFrame(s.asset, s.uiManager, d2ui.FrameLeft)
	s.panelGroup.AddWidget(frame)

	s.panel, err = s.uiManager.NewSprite(d2resource.QuestLogBg, d2resource.PaletteSky)
	if err != nil {
		s.Error(err.Error())
	}

	w, h := frame.GetSize()
	staticPanel := s.uiManager.NewCustomWidgetCached(s.renderStaticMenu, w, h)
	s.panelGroup.AddWidget(staticPanel)

	closeButton := s.uiManager.NewButton(d2ui.ButtonTypeSquareClose, "")
	closeButton.SetVisible(false)
	closeButton.SetPosition(questLogCloseButtonX, questLogCloseButtonY)
	closeButton.OnActivated(func() { s.Close() })
	s.panelGroup.AddWidget(closeButton)

	descrButton := s.uiManager.NewButton(d2ui.ButtonTypeQuestDescr, "")
	descrButton.SetVisible(false)
	descrButton.SetPosition(questLogDescrButtonX, questLogDescrButtonY)
	descrButton.OnActivated(s.onDescrClicked)
	s.panelGroup.AddWidget(descrButton)

	s.questName = s.uiManager.NewLabel(d2resource.Font16, d2resource.PaletteStatic)
	s.questName.Alignment = d2ui.HorizontalAlignCenter
	s.questName.Color[0] = rgbaColor(white)
	s.questName.SetPosition(questNameLabelX, questNameLabelY)
	s.panelGroup.AddWidget(s.questName)

	s.questDescr = s.uiManager.NewLabel(d2resource.Font16, d2resource.PaletteStatic)
	s.questDescr.Alignment = d2ui.HorizontalAlignLeft
	s.questDescr.Color[0] = rgbaColor(white)
	s.questDescr.SetPosition(questDescrLabelX, questDescrLabelY)
	s.panelGroup.AddWidget(s.questDescr)

	s.loadTabs()

	for i := 0; i < d2enum.ActsNumber; i++ {
		s.quests[i] = s.loadQuestIconsForAct(i + 1)
	}

	s.panelGroup.SetVisible(false)
}

func (s *QuestLog) loadTabs() {
	var buttonTypes = []struct {
		bt d2ui.ButtonType
		x  int
	}{
		{d2ui.ButtonTypeTab1, questTab1X},
		{d2ui.ButtonTypeTab2, questTab2X},
		{d2ui.ButtonTypeTab3, questTab3X},
		{d2ui.ButtonTypeTab4, questTab4X},
		{d2ui.ButtonTypeTab5, questTab5X},
	}

	for i := 0; i < d2enum.ActsNumber; i++ {
		currentValue := i
		s.tab[i].newTab(s.uiManager, buttonTypes[i].bt, buttonTypes[i].x)
		s.tab[i].invisibleButton.OnActivated(func() { s.setTab(currentValue) })
		s.panelGroup.AddWidget(s.tab[i].button)
		s.panelGroup.AddWidget(s.tab[i].invisibleButton)
	}

	s.setTab(s.act - 1)
}

func (s *QuestLog) loadQuestIconsForAct(act int) *d2ui.WidgetGroup {
	wg := s.uiManager.NewWidgetGroup(d2ui.RenderPriorityQuestLog)

	var questsInAct int
	if act == d2enum.Act4 {
		questsInAct = d2enum.HalfQuestsNumber
	} else {
		questsInAct = d2enum.NormalActQuestsNumber
	}

	var sockets []*d2ui.Sprite

	var buttons []*d2ui.Button

	var icon *d2ui.Sprite

	for n := 0; n < questsInAct; n++ {
		x, y := s.getPositionForSocket(n)

		socket, err := s.uiManager.NewSprite(d2resource.QuestLogSocket, d2resource.PaletteSky)
		if err != nil {
			s.Error(err.Error())
		}

		socket.SetPosition(x+questOffsetX, y+iconOffsetY+2*questOffsetY)
		sockets = append(sockets, socket)

		button := s.uiManager.NewButton(d2ui.ButtonTypeBlankQuestBtn, "")
		button.SetPosition(x+questOffsetX, y+questOffsetY)
		buttons = append(buttons, button)

		icon, err = s.makeQuestIconForAct(act, n)
		if err != nil {
			s.Error(err.Error())
		}

		icon.SetPosition(x+questOffsetX, y+questOffsetY+iconOffsetY)
		wg.AddWidget(icon)
	}

	for i := 0; i < questsInAct; i++ {
		currentQuest := i
		buttons[i].OnActivated(func() {
			var err error
			for j := 0; j < questsInAct; j++ {
				err = sockets[j].SetCurrentFrame(0)
				if err != nil {
					s.Error(err.Error())
				}
			}
			if act-1 == s.selectedTab {
				err = sockets[currentQuest].SetCurrentFrame(1)
				if err != nil {
					s.Error(err.Error())
				}
			}
			s.onQuestClicked(currentQuest + 1)
		})
	}

	for _, s := range sockets {
		wg.AddWidget(s)
	}

	for _, b := range buttons {
		wg.AddWidget(b)
	}

	wg.SetVisible(false)

	return wg
}

func (s *QuestLog) makeQuestIconForAct(act, n int) (*d2ui.Sprite, error) {
	icon, err := s.uiManager.NewSprite(s.questIconsTable(act, n), d2resource.PaletteSky)
	if err != nil {
		s.Error(err.Error())
	}

	switch s.questStatus[s.cordsToQuestID(act, n)] {
	case d2enum.QuestStatusCompleted:
		err = icon.SetCurrentFrame(completedFrame)
	case d2enum.QuestStatusCompleting:
		// that's not complet now
		err = icon.SetCurrentFrame(0)
		if err != nil {
			s.Error(err.Error())
		}

		icon.PlayForward()
		icon.SetPlayLoop(false)
		err = icon.SetCurrentFrame(completedFrame)
		s.questStatus[s.cordsToQuestID(act, n)] = d2enum.QuestStatusCompleted
	case d2enum.QuestStatusNotStarted:
		err = icon.SetCurrentFrame(notStartedFrame)
	default:
		err = icon.SetCurrentFrame(inProgresFrame)
	}

	return icon, err
}

func (s *QuestLog) setQuestLabel() {
	if s.selectedQuest == 0 {
		s.questName.SetText("")
		s.questDescr.SetText("")

		return
	}

	s.questName.SetText(s.asset.TranslateString(fmt.Sprintf("qstsa%dq%d", s.selectedTab+1, s.selectedQuest)))

	status := s.questStatus[s.cordsToQuestID(s.selectedTab+1, s.selectedQuest)]
	switch status {
	case d2enum.QuestStatusCompleted:
		s.questDescr.SetText(
			strings.Join(
				d2util.SplitIntoLinesWithMaxWidth(
					s.asset.TranslateString("qstsprevious"),
					questDescriptionLenght),
				"\n"),
		)
	case d2enum.QuestStatusNotStarted:
		s.questDescr.SetText("")
	default:
		s.questDescr.SetText(strings.Join(
			d2util.SplitIntoLinesWithMaxWidth(
				s.asset.TranslateString(
					fmt.Sprintf("qstsa%dq%d%d", s.selectedTab+1, s.selectedQuest, status),
				),
				questDescriptionLenght),
			"\n"),
		)
	}
}

func (s *QuestLog) setTab(tab int) {
	s.selectedTab = tab
	s.selectedQuest = d2enum.QuestNone
	s.setQuestLabel()

	for i := 0; i < d2enum.ActsNumber; i++ {
		s.quests[i].SetVisible(tab == i)
	}

	for i := 0; i < d2enum.ActsNumber; i++ {
		s.tab[i].button.SetEnabled(i == tab)
	}
}

func (s *QuestLog) onQuestClicked(number int) {
	s.selectedQuest = number
	s.setQuestLabel()
	s.Infof("Quest number %d in tab %d clicked", number, s.selectedTab)
}

func (s *QuestLog) onDescrClicked() {
	//
}

// IsOpen returns true if the hero status panel is open
func (s *QuestLog) IsOpen() bool {
	return s.isOpen
}

// Toggle toggles the visibility of the hero status panel
func (s *QuestLog) Toggle() {
	if s.isOpen {
		s.Close()
	} else {
		s.Open()
	}
}

// Open opens the hero status panel
func (s *QuestLog) Open() {
	s.isOpen = true
	s.panelGroup.SetVisible(true)
	s.setTab(s.selectedTab)
}

// Close closed the hero status panel
func (s *QuestLog) Close() {
	s.isOpen = false
	s.panelGroup.SetVisible(false)

	for i := 0; i < d2enum.ActsNumber; i++ {
		s.quests[i].SetVisible(false)
	}

	s.onCloseCb()
}

// SetOnCloseCb the callback run on closing the HeroStatsPanel
func (s *QuestLog) SetOnCloseCb(cb func()) {
	s.onCloseCb = cb
}

// Advance updates labels on the panel
func (s *QuestLog) Advance(elapsed float64) {
	//
}

func (s *QuestLog) renderStaticMenu(target d2interface.Surface) {
	s.renderStaticPanelFrames(target)
}

// nolint:dupl // I think it is OK, to duplicate this function
func (s *QuestLog) renderStaticPanelFrames(target d2interface.Surface) {
	frames := []int{
		questLogTopLeft,
		questLogTopRight,
		questLogBottomRight,
		questLogBottomLeft,
	}

	currentX := s.originX + questLogOffsetX
	currentY := s.originY + questLogOffsetY

	for _, frameIndex := range frames {
		if err := s.panel.SetCurrentFrame(frameIndex); err != nil {
			s.Error(err.Error())
		}

		w, h := s.panel.GetCurrentFrameSize()

		switch frameIndex {
		case questLogTopLeft:
			s.panel.SetPosition(currentX, currentY+h)
			currentX += w
		case questLogTopRight:
			s.panel.SetPosition(currentX, currentY+h)
			currentY += h
		case questLogBottomRight:
			s.panel.SetPosition(currentX, currentY+h)
		case questLogBottomLeft:
			s.panel.SetPosition(currentX-w, currentY+h)
		}

		s.panel.Render(target)
	}
}

// copy from character select
func rgbaColor(rgba uint32) color.RGBA {
	result := color.RGBA{}
	a, b, g, r := 0, 1, 2, 3
	byteWidth := 8
	byteMask := 0xff

	for idx := 0; idx < 4; idx++ {
		shift := idx * byteWidth
		component := uint8(rgba>>shift) & uint8(byteMask)

		switch idx {
		case a:
			result.A = component
		case b:
			result.B = component
		case g:
			result.G = component
		case r:
			result.R = component
		}
	}

	return result
}

func (s *QuestLog) cordsToQuestID(act, number int) int {
	key := (act-1)*d2enum.NormalActQuestsNumber + number
	if act > d2enum.Act4 {
		key -= d2enum.HalfQuestsNumber
	}

	return key
}

//nolint:deadcode,unused // I think, it will be used, if not, we can just remove it
func (s *QuestLog) questIDToCords(id int) (act, number int) {
	act = 1

	for i := 0; i < d2enum.ActsNumber; i++ {
		if id < d2enum.NormalActQuestsNumber {
			break
		}

		act++

		id -= d2enum.NormalActQuestsNumber
	}

	number = id
	if act > d2enum.Act4 {
		number -= d2enum.HalfQuestsNumber
	}

	return act, number
}
