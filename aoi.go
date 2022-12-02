// Package aoi 九宫格aoi
package aoi

import (
	"fmt"
	"math"
)

const (
	GridLength = 3                       // 变长3个格子
	GridNum    = GridLength * GridLength // 九宫格
)

// ObjID id 类型
// 可以是 int int32 int64 string任何可比较的类型
type ObjID interface {
	comparable
}

// Grid 格子
type Grid[T ObjID] struct {
	id                     int              // 格子id
	row, col               int              // 行列
	minX, minY, maxX, maxY int              // 格子范围
	objs                   map[T]struct{}   // obj
	surroundGrids          []*Grid[T]       // 九个格子(自己和周围的8个格子)
	surroundGridsMap       map[int]struct{} // map用作快速求交集并集
}

func (g *Grid[ObjID]) add(obj ObjID) {
	g.objs[obj] = struct{}{}
}
func (g *Grid[ObjID]) del(obj ObjID) {
	delete(g.objs, obj)
}
func (g *Grid[ObjID]) isSurround(gridID int) bool {
	_, ok := g.surroundGridsMap[gridID]
	return ok
}
func (g *Grid[ObjID]) addSurroundGrid(other *Grid[ObjID]) {
	if _, ok := g.surroundGridsMap[other.id]; ok {
		panic("duplicate grid")
	}
	g.surroundGridsMap[other.id] = struct{}{}
	g.surroundGrids = append(g.surroundGrids, other)
}

// Rectangle 矩形坐标
func (g *Grid[ObjID]) Rectangle() (int, int, int, int) {
	return g.minX, g.minY, g.maxX, g.maxY
}

// ObjIDs 当前格子的所有obj
func (g *Grid[ObjID]) ObjIDs() map[ObjID]struct{} {
	return g.objs
}

// Has 是否包含obj
func (g *Grid[ObjID]) Has(obj ObjID) bool {
	_, ok := g.objs[obj]
	return ok
}

// Foreach 遍历当前格子包含的obj
func (g *Grid[ObjID]) Foreach(f func(ObjID) bool) {
	for k := range g.objs {
		if !f(k) {
			return
		}
	}
}

// ID 格子id
func (g *Grid[ObjID]) ID() int {
	return g.id
}

func (g *Grid[ObjID]) String() string {
	return fmt.Sprintf("{%d:(%d,%d) %v}", g.id, g.row, g.col, g.objs)
}

// Grids 格子切片类型
type Grids[T ObjID] []*Grid[T]

// Foreach 遍历ObjID
func (gs Grids[T]) Foreach(f func(ObjID T) bool) {
	for _, g := range gs {
		for k := range g.objs {
			if !f(k) {
				return
			}
		}
	}
}

/*
3行row 4列col
	col0	col1 	col2 	col3
row2 |8		|9		|10		|11|
row1 |4 	|5		|6		|7|
row0 |0		|1		|2		|3|

格子7 在row1 col3
7=1*4=3
*/

type pos struct {
	x, y int
}

// AOIManager aoi管理器
type AOIManager[T ObjID] struct {
	minX, maxX, minY, maxY int        // 地图范围
	gridW, gridH           int        // 格子宽高
	row, col               int        // 行列
	grids                  []*Grid[T] // 所有格子
	pos                    map[T]*pos // 对象的坐标
}

// NewAOIManager 构造
func NewAOIManager[T ObjID](minX, minY, maxX, maxY int, gridW, gridH int) (*AOIManager[T], error) {
	if gridH == 0 || gridW == 0 {
		return nil, fmt.Errorf("gridH,gridW should not be 0")
	}
	if minX > maxX || minY > maxY {
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
		pos:   make(map[T]*pos),
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
			g := &Grid[ObjID]{
				id:   idx,
				minX: gridMinX, minY: gridMinY, maxX: gridMaxX, maxY: gridMaxY,
				objs:             make(map[ObjID]struct{}),
				col:              col,
				row:              row,
				surroundGrids:    make([]*Grid[ObjID], 0, GridNum),
				surroundGridsMap: make(map[int]struct{}, GridNum),
			}
			m.grids[idx] = g
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

func (m *AOIManager[ObjID]) atRowCol(posX, posY int) (row, col int) {
	col = int(float64(posX) / float64(m.gridW))
	if col < 0 {
		col = 0
	}
	if col >= m.col {
		col = m.col - 1
	}
	row = int(float64(posY) / float64(m.gridH))
	if row < 0 {
		row = 0
	}
	if row >= m.row {
		row = m.row - 1
	}
	return
}
func (m *AOIManager[ObjID]) atGridIndex(posX, posY int) int {
	row, col := m.atRowCol(posX, posY)
	return m.gridIndex(row, col)
}

// Enter 进入
func (m *AOIManager[ObjID]) Enter(obj ObjID, posX, posY int) Grids[ObjID] {
	m.pos[obj] = &pos{posX, posY}
	g := m.AtGrid(posX, posY)
	g.add(obj)
	return g.surroundGrids
}

// Leave 离开
func (m *AOIManager[ObjID]) Leave(obj ObjID, posX, posY int) Grids[ObjID] {
	g := m.AtGrid(posX, posY)
	g.del(obj)
	delete(m.pos, obj)
	return g.surroundGrids
}

// Move 移动
// currentGrids 没变的格子,直接广播坐标
// enterGrids 新进入的格子,广播创建obj
// leaveGrids 离开的格子,广播删除obj
func (m *AOIManager[ObjID]) Move(obj ObjID, toPosX, toPosY int) (currentGrids, enterGrids, leaveGrids Grids[ObjID]) {
	// 不存在按新进入
	pos, ok := m.pos[obj]
	if !ok {
		return nil, m.Enter(obj, toPosX, toPosY), nil
	}

	// 更新坐标
	oldPosX, oldPosY := pos.x, pos.y
	pos.x, pos.y = toPosX, toPosY

	oldRow, oldCol := m.atRowCol(oldPosX, oldPosY)
	toRow, toCol := m.atRowCol(toPosX, toPosY)
	toGrid := m.grids[m.gridIndex(toRow, toCol)]
	// 格子没变
	if oldCol == toCol && oldRow == toRow {
		return toGrid.surroundGrids, nil, nil
	}

	oldGrid := m.grids[m.gridIndex(oldRow, oldCol)]
	oldGrid.del(obj)
	toGrid.add(obj)

	// 做个优化如果格子跨度很大直接返回
	if abs(toRow-oldRow) >= GridLength || abs(toCol-oldCol) >= GridLength {
		return nil, toGrid.surroundGrids, oldGrid.surroundGrids
	}

	enterGrids = make([]*Grid[ObjID], 0, GridLength)
	leaveGrids = make([]*Grid[ObjID], 0, GridLength)
	// 离开的格子 = 旧的九宫格-到达的九宫格
	for _, grid := range oldGrid.surroundGrids {
		if !toGrid.isSurround(grid.id) {
			leaveGrids = append(leaveGrids, grid)
		}
	}
	// 新进入的格子 = 到达的九宫格-旧的九宫格
	for _, grid := range toGrid.surroundGrids {
		if !oldGrid.isSurround(grid.id) {
			enterGrids = append(enterGrids, grid)
		}
	}
	// 没变的格子 = 到达的九宫格-新进入的格子
	for _, grid := range toGrid.surroundGrids {
		has := false
		for _, v := range enterGrids {
			if v.id == grid.id {
				has = true
				break
			}
		}
		if !has {
			currentGrids = append(currentGrids, grid)
		}
	}
	return
}

// AtGrid 坐标所在的格子
// 出地图边界给返回边界的格子
func (m *AOIManager[ObjID]) AtGrid(posX, posY int) *Grid[ObjID] {
	return m.grids[m.atGridIndex(posX, posY)]
}

// SurroundGrids 坐标所在包围的九宫格
func (m *AOIManager[ObjID]) SurroundGrids(posX, posY int) Grids[ObjID] {
	return m.grids[m.atGridIndex(posX, posY)].surroundGrids
}

// AllGrids 所有格子
func (m *AOIManager[ObjID]) AllGrids() Grids[ObjID] {
	return m.grids
}

func (m *AOIManager[ObjID]) String() string {
	str := ""
	for _, v := range m.grids {
		str += v.String() + "\n"
	}
	return str
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}
