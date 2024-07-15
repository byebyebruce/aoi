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

type EventCallback[T ObjID] func(event AOIEvent, eventMaker T, eventWatcher T)

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

type point struct {
	x, y   int
	gridID int
}

// AOIManager aoi管理器
type AOIManager[T ObjID] struct {
	minX, maxX, minY, maxY int          // 地图范围
	gridW, gridH           int          // 格子宽高
	row, col               int          // 行列
	grids                  []*Grid[T]   // 所有格子
	pos                    map[T]*point // 对象的坐标
}

// NewAOIManager 构造
// minX, minY, maxX, maxY 可以是负数
func NewAOIManager[T ObjID](minX, minY, maxX, maxY int, gridW, gridH int) (*AOIManager[T], error) {
	if gridH <= 0 || gridW <= 0 {
		return nil, fmt.Errorf("gridH,gridW should not be 0")
	}
	if minX >= maxX || minY >= maxY {
		return nil, fmt.Errorf("min should be small than max")
	}
	// 列
	col := int(math.Ceil(float64(maxX-minX) / float64(gridW)))
	// 行
	row := int(math.Ceil(float64(maxY-minY) / float64(gridH)))
	m := &AOIManager[T]{
		minX:  minX,
		minY:  minY,
		maxX:  maxX,
		maxY:  maxY,
		gridH: gridH,
		gridW: gridW,
		col:   col,
		row:   row,
		grids: make([]*Grid[T], col*row),
		pos:   make(map[T]*point),
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
			gridMinX, gridMinY := col*m.gridW, row*m.gridH
			gridMaxX, gridMaxY := gridMinX+m.gridW, gridMinY+m.gridH
			if gridMaxX > m.maxX {
				gridMaxX = m.maxX
			}
			if gridMaxY > m.maxY {
				gridMaxY = m.maxY
			}
			m.grids[idx] = newGrid[ObjID](idx, gridMinX, gridMinY, gridMaxX, gridMaxY, row, col)
		}
	}

	// 预保存周围的格子
	for col := 0; col < m.col; col++ {
		for row := 0; row < m.row; row++ {
			// 当前格子的id
			gridIdx := m.gridIndex(row, col)
			g := m.grids[gridIdx]

			leftBottomRow, leftBottomCol := row-1, col-1
			for _row := leftBottomRow; _row < leftBottomRow+GridLength; _row++ {
				if _row < 0 || _row >= m.row {
					continue
				}
				for _col := leftBottomCol; _col < leftBottomCol+GridLength; _col++ {
					if _col < 0 || _col >= m.col {
						continue
					}
					// 周围格子的id(包括自己)
					surroundGridIdx := m.gridIndex(_row, _col)
					g.addSurroundGrid(m.grids[surroundGridIdx])
				}
			}
		}
	}
}

func (m *AOIManager[ObjID]) atGridIndex(posX, posY int) int {
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
func (m *AOIManager[ObjID]) Enter(id ObjID, posX, posY int, cb EventCallback[ObjID]) bool {
	if _, ok := m.pos[id]; ok {
		return false
	}
	g := m.AtGrid(posX, posY)
	g.add(id)
	m.pos[id] = &point{posX, posY, g.id}
	if cb != nil {
		for _, grid := range g.SurroundGrids() {
			grid.onEvent(Enter, id, cb)
		}
	}
	return true
}

// Leave 离开
func (m *AOIManager[ObjID]) Leave(id ObjID, cb EventCallback[ObjID]) bool {
	pos, ok := m.pos[id]
	if !ok {
		return false
	}
	g := m.grids[pos.gridID]
	g.del(id)
	delete(m.pos, id)

	if cb != nil {
		for _, grid := range g.SurroundGrids() {
			grid.onEvent(Leave, id, cb)
		}
	}

	return true
}

// Move 移动
/*
form 1 to 2
G1: 1所在个9九宫
G2: 2所在个9九宫
L: Leave
M: Move
E: Enter

公式
L = G1-G2
M = G2∩G1
E = G2-G1
+----+----+----+----+
|  L |  M |  M |  E |
+----+----+----+----+
|  L |  1 |  2 |  E |
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
	toGrid := m.AtGrid(toPosX, toPosY)

	// 更新坐标
	pos.x, pos.y, pos.gridID = toPosX, toPosY, toGrid.id

	// 格子没变
	if fromGrid.id == toGrid.id {
		for _, grid := range toGrid.SurroundGrids() {
			grid.onEvent(Move, id, cb)
		}
		return true
	}

	fromGrid.del(id)
	toGrid.add(id)

	// 做个优化如果格子跨度很大直接返回
	if abs(toGrid.row-fromGrid.row) >= GridLength || abs(toGrid.col-fromGrid.col) >= GridLength {
		for _, grid := range toGrid.SurroundGrids() {
			grid.onEvent(Enter, id, cb)
		}
		for _, grid := range fromGrid.SurroundGrids() {
			grid.onEvent(Leave, id, cb)
		}
		return true
	}

	// 离开的格子 = 原来所在的九宫格-到达的九宫格
	for _, grid := range fromGrid.SurroundGrids() {
		if !toGrid.isSurround(grid.id) {
			grid.onEvent(Leave, id, cb)
		}
	}

	// 没变的格子 = 原来所在的九宫格和到达的九宫格取交集
	for _, grid := range toGrid.SurroundGrids() {
		if fromGrid.isSurround(grid.id) {
			grid.onEvent(Move, id, cb)
		}
	}

	// 新进入的格子 = 到达的九宫格-原来所在的九宫格
	for _, grid := range toGrid.SurroundGrids() {
		if !fromGrid.isSurround(grid.id) {
			grid.onEvent(Enter, id, cb)
		}
	}

	return true
}

// AtGrid 坐标所在的格子
// 出地图边界给返回边界的格子
func (m *AOIManager[ObjID]) AtGrid(posX, posY int) *Grid[ObjID] {
	return m.grids[m.atGridIndex(posX, posY)]
}

// AllGrids 所有格子
func (m *AOIManager[ObjID]) AllGrids() []*Grid[ObjID] {
	return m.grids
}

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

func (m *AOIManager[ObjID]) Clear() {
	m.pos = make(map[ObjID]*point)
	for _, v := range m.grids {
		v.clear()
	}
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}
