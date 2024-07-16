package aoi

import "fmt"

// Grid 格子
type Grid[T ObjID] struct {
	id                     int              // 格子id
	row, col               int              // 行列
	minX, minY, maxX, maxY int              // 格子范围
	surroundGrids          []*Grid[T]       // 九个格子(自己和周围的8个格子)
	surroundGridsMap       map[int]struct{} // map用作快速求交集并集

	objs map[T]struct{} // obj
}

func newGrid[T ObjID](id int, gridMinX, gridMinY, gridMaxX, gridMaxY int, row, col int) *Grid[T] {
	return &Grid[T]{
		id:   id,
		minX: gridMinX, minY: gridMinY, maxX: gridMaxX, maxY: gridMaxY,
		objs:             make(map[T]struct{}),
		col:              col,
		row:              row,
		surroundGrids:    make([]*Grid[T], 0, GridNum),
		surroundGridsMap: make(map[int]struct{}, GridNum),
	}
}
func (g *Grid[ObjID]) add(obj ObjID) {
	g.objs[obj] = struct{}{}
}

func (g *Grid[ObjID]) del(obj ObjID) {
	delete(g.objs, obj)
}

func (g *Grid[ObjID]) clear() {
	g.objs = make(map[ObjID]struct{})
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
	if len(g.surroundGrids) > GridNum {
		panic("surround grid num is 9")
	}
}

// ID 格子id
func (g *Grid[ObjID]) ID() int {
	return g.id
}

// BoundingBox 范围
func (g *Grid[ObjID]) BoundingBox() (int, int, int, int) {
	return g.minX, g.minY, g.maxX, g.maxY
}

// RowCol 行列
func (g *Grid[ObjID]) RowCol() (int, int) {
	return g.row, g.col
}

// Contains 是否包含obj
func (g *Grid[ObjID]) Contains(obj ObjID) bool {
	_, ok := g.objs[obj]
	return ok
}

// ObjIDs 当前格子的所有obj
func (g *Grid[ObjID]) ObjIDs() map[ObjID]struct{} {
	return g.objs
}

// SurroundGrids 九宫格(包括自己)
func (g *Grid[ObjID]) SurroundGrids() []*Grid[ObjID] {
	return g.surroundGrids
}

// ForeachInSurroundGrids 遍历当前格子包含的obj
func (g *Grid[ObjID]) ForeachInSurroundGrids(f func(other ObjID)) {
	for _, v := range g.surroundGrids {
		for k := range v.objs {
			f(k)
		}
	}
}

func (g *Grid[ObjID]) String() string {
	return fmt.Sprintf("(%d:%d,%d)", g.id, g.row, g.col)
}