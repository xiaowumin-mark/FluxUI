package widget

type scrollCommandKind uint8

const (
	scrollCmdToStart scrollCommandKind = iota + 1
	scrollCmdToEnd
	scrollCmdToOffset
	scrollCmdBy
)

type scrollCommand struct {
	kind   scrollCommandKind
	offset int
	delta  float32
}

// ScrollRef 是 ScrollView 的命令型引用，用于外部主动控制滚动位置。
// 命令在下一帧由 ScrollView 消费；未挂载时只保留最近的有界队列。
type ScrollRef struct {
	queue commandQueue[scrollCommand]
}

// NewScrollRef 创建一个可复用的 ScrollRef。
func NewScrollRef() *ScrollRef {
	return &ScrollRef{}
}

// ScrollToStart 滚动到起始位置。
func (r *ScrollRef) ScrollToStart() {
	r.enqueue(scrollCommand{kind: scrollCmdToStart})
}

// ScrollToTop 是 ScrollToStart 的语义别名（垂直滚动场景）。
func (r *ScrollRef) ScrollToTop() {
	r.ScrollToStart()
}

// ScrollToEnd 滚动到末尾位置。
func (r *ScrollRef) ScrollToEnd() {
	r.enqueue(scrollCommand{kind: scrollCmdToEnd})
}

// ScrollToBottom 是 ScrollToEnd 的语义别名（垂直滚动场景）。
func (r *ScrollRef) ScrollToBottom() {
	r.ScrollToEnd()
}

// ScrollToOffset 按主轴偏移量滚动到绝对位置（像素）。
func (r *ScrollRef) ScrollToOffset(offset int) {
	if offset < 0 {
		offset = 0
	}
	r.enqueue(scrollCommand{
		kind:   scrollCmdToOffset,
		offset: offset,
	})
}

// ScrollBy 按相对距离滚动（单位同 Gio List.ScrollBy）。
func (r *ScrollRef) ScrollBy(delta float32) {
	if delta == 0 {
		return
	}
	r.enqueue(scrollCommand{
		kind:  scrollCmdBy,
		delta: delta,
	})
}

func (r *ScrollRef) enqueue(cmd scrollCommand) {
	if r == nil {
		return
	}
	r.queue.enqueue(cmd)
}

func (r *ScrollRef) drainCommands() []scrollCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}
