// Package aoi 九宫格aoi
package aoi

import (
	"fmt"
	"math"
)

/*
九宫格 aoi

负责创建和维护九宫格可见规则。
当外界触发进入，离开，移动事件时, 将触发的结果通知观察者

例如3行row 4列col

	   col0 col1 col2 col3
		+---+---+---+---+

row2| 8 | 9 | 10| 11|

	+---+---+---+---+

row1| 4 | 5 | 6 | 7 |

	+---+---+---+---+

row0| 0 | 1 | 2 | 3 |

	+---+---+---+---+

	 ^y
	 |
	 |
	 |
	 0------->x

格子7 在row1 col3
7=4*1+3
格子10 在row2 col2
10=4*2+2

相邻可见的格子:
0格子只能看到0,1,4,5
6格子能看到6,1,2,3,5,7,9,10,11

AOI事件规则:
!!!任何事件都不通知事件的trigger!!!

1. 进入事件: 通知九宫格内所有观察者进入事件
2. 离开事件: 通知九宫格内所有观察者离开事件
3. 更新事件:
a. 通知离开的的九宫格(之前所在的九宫格-新进入的九宫格)离开事件
b. 通知没变的九宫格(之前所在的九宫格和新进入的九宫格取交集)移动事件
c. 通知新进入的九宫格(新进入的九宫格-之前所在的九宫格)进入事件
*/
const (
	// GridLength 一边3个格子
	GridLength = 3
	// GridNum 九宫格
	GridNum = GridLength * GridLength
)

// EventType aoi 事件
type EventType int

const (
	// EnterView 进入事件
	// 事件制造者只是触发者，会回调所有观察者(例如npc进入，通知周围的player创建这个新进入的npc)
	// 事件制造者只是观察者，会回调所有人(例如隐身的gm进入，通知gm创建所有可见的人，不通知其他人gm进入)
	// 事件制造者既是触发者又是观察者，会回调所有人(例如player进入，通知player创建所有可见的人)
	EnterView EventType = iota

	// LeaveView 离开事件
	// 事件制造者只是触发者，会回调所有观察者(例如npc离开，通知周围的player删除这个离开的npc)
	// 事件制造者只是观察者，会回调所有人(例如gm离开，通知gm删除所有可见的人，不通知其他人gm离开)
	// 事件制造者既是触发者又是观察者，会回调所有人(例如player离开，通知player删除所有可见的人)
	LeaveView

	// UpdateView 更新事件
	// 事件制造者只是触发者，会回调所有观察者(例如npc移动，通知周围的player这个npc移动了)
	// 事件制造者只是观察者，不会回调任何人(例如gm移动，不会通知任何人)
	// 事件制造者既是触发者又是观察者，会回调所有观察者(例如player移动，通知周围的player这个player移动了)
	UpdateView
)

// ObjType 对象类型
type ObjType int

const (
	// Trigger 只触发事件，不观察事件(例如NPC)
	Trigger ObjType = 1 << iota

	// Observer 只观察事件，不触发事件(例如GM，隐身但能看到其他)
	Observer

	// TriggerAndObserver 既触发事件，又观察事件(例如Player)
	TriggerAndObserver = Trigger | Observer
)

func (o ObjType) IsTrigger() bool {
	return o&Trigger != 0
}

func (o ObjType) IsObserver() bool {
	return o&Observer != 0
}

// ObjID id
// 类型
type ObjID interface {
	comparable
}

// EventCallback 事件回调, event 事件类型, other 其他对象的id
// ***注意***
// 事件不会回调事件触发者
// Enter时的事件只会是EnterView
// Leave时的事件只会是LeaveView
type EventCallback[T ObjID] func(event EventType, other T)

// obj 对象
type obj struct {
	// 所在格子id
	gridID int
	// 坐标
	x, y int
	// 是否是观察者, 非观察者不接受事件通知
	ot ObjType
}

// AOIManager aoi管理器
type AOIManager[T ObjID] struct {
	minX, minY, maxX, maxY int        // 地图范围
	gridW, gridH           int        // 格子宽高
	row, col               int        // 总行数 总列数
	grids                  []*Grid[T] // 所有格子
	objs                   map[T]*obj // 对象的坐标
}

// NewAOIManager 构造
func NewAOIManager[T ObjID](width, height int, gridW, gridH int) (*AOIManager[T], error) {
	return NewAOIManagerFrom[T](0, 0, width, height, gridW, gridH)
}

// NewAOIManagerFrom 构造
// x, y 可以是负数
func NewAOIManagerFrom[T ObjID](x, y, width, height int, gridW, gridH int) (*AOIManager[T], error) {
	if gridH <= 0 || gridW <= 0 {
		return nil, fmt.Errorf("gridH,gridW should not be 0")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("width, height should not be 0")
	}
	maxX, maxY := x+width, y+height
	// 列
	col := int(math.Ceil(float64(width) / float64(gridW)))
	// 行
	row := int(math.Ceil(float64(height) / float64(gridH)))
	m := &AOIManager[T]{
		minX:  x,
		minY:  y,
		maxX:  maxX,
		maxY:  maxY,
		gridH: gridH,
		gridW: gridW,
		col:   col,
		row:   row,
		grids: make([]*Grid[T], 0, col*row),
		objs:  make(map[T]*obj),
	}
	m.init()
	return m, nil
}
func (m *AOIManager[ObjID]) init() {
	for row := 0; row < m.row; row++ {
		for col := 0; col < m.col; col++ {
			idx := m.gridIndex(row, col)
			if len(m.grids) != idx {
				panic(fmt.Sprintf("idx:%d len:%d", idx, len(m.grids)))
			}

			gridMinX, gridMinY := m.minX+col*m.gridW, m.minY+row*m.gridH
			gridMaxX, gridMaxY := gridMinX+m.gridW, gridMinY+m.gridH
			if gridMaxX > m.maxX {
				gridMaxX = m.maxX
			}
			if gridMaxY > m.maxY {
				gridMaxY = m.maxY
			}
			grid := newGrid[ObjID](idx, gridMinX, gridMinY, gridMaxX, gridMaxY, row, col)
			m.grids = append(m.grids, grid)
		}
	}

	// 预保存周围的格子
	for col := 0; col < m.col; col++ {
		for row := 0; row < m.row; row++ {
			// 当前格子的id
			gird := m.grids[m.gridIndex(row, col)]

			// 周围9个格子
			for i := 0; i < GridLength; i++ {
				_row := row - 1 + i
				if _row < 0 || _row >= m.row {
					continue
				}
				for j := 0; j < GridLength; j++ {
					_col := col - 1 + j
					if _col < 0 || _col >= m.col {
						continue
					}
					// 周围格子的id(包括自己)
					surroundGrid := m.grids[m.gridIndex(_row, _col)]
					gird.addSurroundGrid(surroundGrid)
				}
			}
		}
	}
}

// Enter 进入，cb是因
// eventType 只会是EnterView
// isObserver 是否是观察者. 只有观察者才会接受事件通知(一般player为观察者，npc为非观察者)
func (m *AOIManager[ObjID]) Enter(id ObjID, posX, posY int, ot ObjType, cb EventCallback[ObjID]) bool {
	if _, ok := m.objs[id]; ok {
		return false
	}
	var (
		g          = m.PosAtGrid(posX, posY)
		isObserver = ot.IsObserver()
	)
	g.add(id, isObserver)
	o := &obj{g.id, posX, posY, ot}
	m.objs[id] = o
	if cb == nil {
		return true
	}

	for _, sg := range g.SurroundGrids() {
		sg.invokeEvent(id, isObserver, EnterView, cb)
	}
	return true
}

// Leave 离开
// event 只会是LeaveView
func (m *AOIManager[ObjID]) Leave(id ObjID, cb EventCallback[ObjID]) bool {
	o, ok := m.objs[id]
	if !ok {
		return false
	}
	var (
		g          = m.grids[o.gridID]
		isObserver = o.ot.IsObserver()
	)
	g.del(id)
	delete(m.objs, id)

	if cb == nil {
		return true
	}
	for _, sg := range g.SurroundGrids() {
		sg.invokeEvent(id, isObserver, LeaveView, cb)
	}

	return true
}

// Move 移动
/*
fromGrid: 当前所在格子
toGrid: 移动到的格子

公式
EnterView = toGrid - fromGrid
UpdateView = fromGrid ∩ toGrid
LeaveView = fromGrid - toGrid

L: LeaveView
U: UpdateView
E: EnterView
+----+----+----+----+
|  L |  U |  U |  E |
+----+----+----+----+
|  L |from| to |  E |
+----+----+----+----+
|  L |  U |  U |  E |
+----+----+----+----+
*/
func (m *AOIManager[ObjID]) Move(id ObjID, toPosX, toPosY int, cb EventCallback[ObjID]) bool {
	o, ok := m.objs[id]
	if !ok {
		return false
	}

	var (
		fromGrid   = m.grids[o.gridID]
		toGrid     = m.PosAtGrid(toPosX, toPosY)
		isTrigger  = o.ot.IsTrigger()
		isObserver = o.ot.IsObserver()
	)

	// 更新坐标
	o.x, o.y, o.gridID = toPosX, toPosY, toGrid.id
	if fromGrid.id != toGrid.id {
		fromGrid.del(id)
		toGrid.add(id, isObserver)
	}

	if cb == nil {
		return true
	}

	// 情况1. 在同一个格子内移动
	if fromGrid.id == toGrid.id {
		if isTrigger {
			for _, sg := range toGrid.SurroundGrids() {
				sg.invokeEvent(id, false, UpdateView, cb)
			}
		}
		return true
	}

	// 情况2. 跨越3个格子
	if abs(toGrid.row-fromGrid.row) >= GridLength ||
		abs(toGrid.col-fromGrid.col) >= GridLength {
		for _, sg := range toGrid.SurroundGrids() {
			sg.invokeEvent(id, isObserver, EnterView, cb)
		}
		for _, sg := range fromGrid.SurroundGrids() {
			sg.invokeEvent(id, isObserver, LeaveView, cb)
		}
		return true
	}

	//情况3.
	// 1) 新进入的格子 = 到达的九宫格-原来所在的九宫格
	for _, grid := range toGrid.SurroundGrids() {
		if !fromGrid.isSurround(grid.id) {
			grid.invokeEvent(id, isObserver, EnterView, cb)
		}
	}

	// 2) 没变的格子 = 原来所在的九宫格和到达的九宫格取交集
	// 只有trigger才会通知
	if isTrigger {
		for _, grid := range fromGrid.SurroundGrids() {
			if toGrid.isSurround(grid.id) {
				grid.invokeEvent(id, false, UpdateView, cb)
			}
		}
	}

	// 3) 离开的格子 = 原来所在的九宫格-到达的九宫格
	for _, grid := range fromGrid.SurroundGrids() {
		if !toGrid.isSurround(grid.id) {
			grid.invokeEvent(id, isObserver, LeaveView, cb)
		}
	}

	return true
}

// ObjGrid obj所在的格子
func (m *AOIManager[ObjID]) ObjGrid(id ObjID) *Grid[ObjID] {
	o, ok := m.objs[id]
	if !ok {
		return nil
	}
	return m.grids[o.gridID]
}

// PosAtGrid 坐标所在的格子
// 出地图边界给返回边界的格子
func (m *AOIManager[ObjID]) PosAtGrid(posX, posY int) *Grid[ObjID] {
	return m.grids[m.posAtGridIndex(posX, posY)]
}

// AllGrids 所有格子
func (m *AOIManager[ObjID]) AllGrids() []*Grid[ObjID] {
	return m.grids
}

// Clear 清空
func (m *AOIManager[ObjID]) Clear() {
	m.objs = make(map[ObjID]*obj)
	for _, v := range m.grids {
		v.clear()
	}
}

// String 格式化输出
func (m *AOIManager[ObjID]) String() string {
	str := ""
	for row := 0; row < m.row; row++ {
		for col := 0; col < m.col; col++ {
			str += fmt.Sprintf("%10s ", m.grids[m.gridIndex(row, col)])
		}
		str += "\n"
	}
	return str
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func (m *AOIManager[ObjID]) gridIndex(row, col int) int {
	return row*m.col + col
}

func (m *AOIManager[ObjID]) posAtGridIndex(posX, posY int) int {
	var col, row int
	if posX <= m.minX {
		col = 0
	} else if posX >= m.maxX {
		col = m.col - 1
	} else {
		col = int(float64(posX-m.minX) / float64(m.gridW))
	}
	if posY <= m.minY {
		row = 0
	} else if posY >= m.maxY {
		row = m.row - 1
	} else {
		row = int(float64(posY-m.minY) / float64(m.gridH))
	}
	return m.gridIndex(row, col)
}
