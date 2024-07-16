// Package aoi 九宫格aoi
package aoi

import (
	"fmt"
	"math"
)

/*
 */
const (
	// GridLength 一边3个格子
	GridLength = 3
	// GridNum 九宫格
	GridNum = GridLength * GridLength
)

const (
	// Enter 进入
	Enter AOIEvent = iota
	// Leave 离开
	Leave
	// Move 移动
	Move
)

// ObjID id 类型
// 可以是 int int32 int64 string任何可比较的类型
type ObjID interface {
	comparable
}

type AOIEvent int

type EventCallback[T ObjID] func(event AOIEvent, other T)

/*
3行row 4列col

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
*/

type obj struct {
	x, y   int
	gridID int
}

// AOIManager aoi管理器
type AOIManager[T ObjID] struct {
	minX, minY, maxX, maxY int        // 地图范围
	gridW, gridH           int        // 格子宽高
	row, col               int        // 总行数 总列数
	grids                  []*Grid[T] // 所有格子
	pos                    map[T]*obj // 对象的坐标
}

// NewAOIManager 构造
func NewAOIManager[T ObjID](width, height int, gridW, gridH int) (*AOIManager[T], error) {
	return NewAOIManagerWithMinXY[T](0, 0, width, height, gridW, gridH)
}

// NewAOIManagerWithMinXY 构造
// minX, minY 可以是负数
func NewAOIManagerWithMinXY[T ObjID](minX, minY, width, height int, gridW, gridH int) (*AOIManager[T], error) {
	if gridH <= 0 || gridW <= 0 {
		return nil, fmt.Errorf("gridH,gridW should not be 0")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("width, height should not be 0")
	}
	maxX, maxY := minX+width, minY+height
	// 列
	col := int(math.Ceil(float64(width) / float64(gridW)))
	// 行
	row := int(math.Ceil(float64(height) / float64(gridH)))
	m := &AOIManager[T]{
		minX:  minX,
		minY:  minY,
		maxX:  maxX,
		maxY:  maxY,
		gridH: gridH,
		gridW: gridW,
		col:   col,
		row:   row,
		grids: make([]*Grid[T], 0, col*row),
		pos:   make(map[T]*obj),
	}
	m.init()
	return m, nil
}

func (m *AOIManager[ObjID]) gridIndex(row, col int) int {
	return row*m.col + col
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

// Enter 进入
func (m *AOIManager[ObjID]) Enter(id ObjID, posX, posY int, enterView func(other ObjID)) bool {
	if _, ok := m.pos[id]; ok {
		return false
	}
	g := m.PosAtGrid(posX, posY)
	g.add(id)
	m.pos[id] = &obj{posX, posY, g.id}
	if enterView != nil {
		g.ForeachInSurroundGrids(func(other ObjID) {
			if id == other {
				return
			}
			enterView(other)
		})
	}
	return true
}

// Leave 离开
func (m *AOIManager[ObjID]) Leave(id ObjID, leaveView func(other ObjID)) bool {
	pos, ok := m.pos[id]
	if !ok {
		return false
	}
	g := m.grids[pos.gridID]
	g.del(id)
	delete(m.pos, id)

	if leaveView != nil {
		g.ForeachInSurroundGrids(func(other ObjID) {
			leaveView(other)
		})
	}

	return true
}

// Move 移动
/*
fromGrid: 当前所在格子
toGrid: 移动到的格子

公式
Leave = fromGrid - toGrid
Move = fromGrid ∩ toGrid
Enter = toGrid - fromGrid

L: Leave
M: Move
E: Enter
+----+----+----+----+
|  L |  M |  M |  E |
+----+----+----+----+
|  L |from| to |  E |
+----+----+----+----+
|  L |  M |  M |  E |
+----+----+----+----+
*/
func (m *AOIManager[ObjID]) Move(id ObjID, toPosX, toPosY int, cb EventCallback[ObjID]) bool {
	pos, ok := m.pos[id]
	if !ok {
		return false
	}

	fromGrid := m.grids[pos.gridID]
	toGrid := m.PosAtGrid(toPosX, toPosY)

	// 更新坐标
	pos.x, pos.y, pos.gridID = toPosX, toPosY, toGrid.id
	if fromGrid.id != toGrid.id {
		fromGrid.del(id)
		toGrid.add(id)
	}

	if cb == nil {
		return true
	}

	// 情况1. 在同一个格子内移动
	if fromGrid.id == toGrid.id {
		toGrid.ForeachInSurroundGrids(func(other ObjID) {
			if id == other {
				return
			}
			cb(Move, other)
		})
		return true
	}

	// 情况2. 跨越3个格子
	if abs(toGrid.row-fromGrid.row) >= GridLength ||
		abs(toGrid.col-fromGrid.col) >= GridLength { // 直接跨越3个格子
		toGrid.ForeachInSurroundGrids(func(other ObjID) {
			if id == other {
				return
			}
			cb(Enter, other)
		})

		fromGrid.ForeachInSurroundGrids(func(other ObjID) {
			cb(Leave, other)
		})
		return true
	}

	// 离开的格子 = 原来所在的九宫格-到达的九宫格
	for _, grid := range fromGrid.SurroundGrids() {
		if !toGrid.isSurround(grid.id) {
			for other, _ := range grid.objs {
				cb(Leave, other)
			}
		}
	}

	// 没变的格子 = 原来所在的九宫格和到达的九宫格取交集
	for _, grid := range toGrid.SurroundGrids() {
		if fromGrid.isSurround(grid.id) {
			for other, _ := range grid.objs {
				if id == other {
					continue
				}
				cb(Move, other)
			}
		}
	}

	// 新进入的格子 = 到达的九宫格-原来所在的九宫格
	for _, grid := range toGrid.SurroundGrids() {
		if !fromGrid.isSurround(grid.id) {
			for other, _ := range grid.objs {
				if id == other {
					continue
				}
				cb(Enter, other)
			}
		}
	}

	return true
}

// ObjGrid 所在格子
func (m *AOIManager[ObjID]) ObjGrid(id ObjID) *Grid[ObjID] {
	p, ok := m.pos[id]
	if !ok {
		return nil
	}
	return m.grids[p.gridID]
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
	m.pos = make(map[ObjID]*obj)
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
