package main

import "math"
import "github.com/sixthgear/noise"
import "github.com/go-gl/gl"

type Vertex struct {
	x, y, z float32
}

var colorMap [][3]float32 = [][3]float32{
	{0.0, 0.0, 0.7},
	{0.2, 0.5, 1.0},
	{0.5, 0.5, 0.4},
	{1.0, 1.0, 0.9},
	{0.1, 0.4, 0.1},
	{0.2, 0.2, 0.2},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{0.3, 0.3, 0.3},
	{1.0, 1.0, 1.0},
}

const (
	RIGHT = iota
	LEFT
)

type Map struct {
	width, depth int
	heightMap    []float32
	vertices     []float32
	normals      []float32
	colors       []float32
	texcoords    []float32
	texture      gl.Texture
	gridSize     int
	minHeight    float64
	maxHeight    float64
}

func Cross(v1 Vertex, v2 Vertex) Vertex {
	n := Vertex{}
	n.x = v1.y*v2.z - v1.z*v2.y
	n.y = v1.z*v2.x - v1.x*v2.z
	n.z = v1.x*v2.y - v1.y*v2.x
	return n
}

func GenerateMap(width, depth int, gridSize int) *Map {

	m := new(Map)

	m.width = width
	m.depth = depth
	m.gridSize = gridSize
	m.minHeight = 1000000
	m.maxHeight = 0

	diag := math.Hypot(float64(m.width/2), float64(m.depth/2))
	for z := 0; z < depth; z++ {
		for x := 0; x < width; x++ {

			fx := float64(x) + float64(z%2)*0.5
			fz := float64(z)

			d := math.Hypot(float64(m.width/2)-fx, float64(m.depth/2)-fz)
			d = 1.0 - d/diag
			h := noise.OctaveNoise2d(fx, fz, 4, 0.25, 1.0/28)
			h = (h + 1.0) * 0.5
			h = math.Sqrt(h) * 1024 * (math.Pow(d, 2))
			h = math.Max(h, 120)
			m.heightMap = append(m.heightMap, float32(h))
			m.minHeight = math.Min(m.minHeight, h)
			m.maxHeight = math.Max(m.maxHeight, h)
		}
	}

	return m
}

func (m *Map) getColorForVertex(v Vertex) [3]float32 {

	n := (float64(v.y) - m.minHeight) / (m.maxHeight - m.minHeight) * float64(len(colorMap)-1)
	i0 := int(math.Floor(n))
	i1 := int(math.Ceil(n))
	f := n - math.Floor(n)

	if f < 0.01 {
		i1 = i0
	}

	c0 := colorMap[i0]
	c1 := colorMap[i1]

	color := [3]float32{}
	for i := 0; i < 3; i++ {
		color[i] = c0[i] + (c1[i]-c0[i])*float32(f)
		if n > 2 {
			color[i] += float32(noise.OctaveNoise2d(float64(v.x), float64(v.z), 1, 0.25, 1/32.0)) * 0.04
		}

	}
	return color
}

func (m *Map) PlaceVertex(x, z int) Vertex {

	s := m.gridSize

	xOffset := float32(z % 2 * s / 2)
	vx := xOffset + float32(x*s)
	vy := m.heightMap[z*m.width+x]
	vz := float32(z * s)

	return Vertex{vx, vy, vz}
}

func (m *Map) BuildVertices() {

	for i := 0; i < (m.depth-1)*(m.width); i++ {

		x := i % (m.width)
		z := i / (m.width)
		num := 2
		dir := z % 2

		if dir == LEFT {
			x = (m.width - 1) - x
		}
		if dir == RIGHT && x == 0 && z > 0 {
			num = 1
		}
		if dir == LEFT && x == m.width-1 {
			num = 1
		}

		for j := 2 - num; j <= 1; j++ {
			v := m.PlaceVertex(x, z+j)
			n := m.GetNormal(v, x, z+j)
			c := m.getColorForVertex(v)
			m.vertices = append(m.vertices, v.x, v.y, v.z)
			m.normals = append(m.normals, n.x, n.y, n.z)
			m.colors = append(m.colors, c[0], c[1], c[2])
		}

	}

}

func (m *Map) GetNormal(v Vertex, x, z int) (normal Vertex) {

	var neighbors []Vertex
	var sum Vertex

	if x > 0 {
		neighbors = append(neighbors, m.PlaceVertex(x-1, z))
	}
	if z > 0 {
		neighbors = append(neighbors, m.PlaceVertex(x, z-1))
	}
	if x < m.width-1 {
		neighbors = append(neighbors, m.PlaceVertex(x+1, z))
	}
	if z < m.depth-1 {
		neighbors = append(neighbors, m.PlaceVertex(x, z+1))
	}

	for i, n1 := range neighbors {
		n2 := neighbors[(i+1)%len(neighbors)]
		d1 := Vertex{n1.x - v.x, n1.y - v.y, n1.z - v.z}
		d2 := Vertex{n2.x - v.x, n2.y - v.y, n2.z - v.z}
		cross := Cross(d1, d2)
		sum.x += cross.x
		sum.y += cross.y
		sum.z += cross.z
	}

	num := float32(len(neighbors))
	sum.x /= num
	sum.y /= num
	sum.z /= num

	length := float32(math.Sqrt(float64(sum.x*sum.x + sum.y*sum.y + sum.z*sum.z)))
	if length > 0 {
		normal.x = sum.x / length
		normal.y = sum.y / length
		normal.z = sum.z / length
	}

	if normal.y < 0 {
		normal.y *= -1
	}

	return normal
}

func (m *Map) Draw() {

	gl.PushMatrix()
	gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.VertexPointer(3, gl.FLOAT, 0, m.vertices)

	gl.EnableClientState(gl.NORMAL_ARRAY)
	gl.NormalPointer(gl.FLOAT, 0, m.normals)

	// gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	// gl.TexCoordPointer(2, gl.FLOAT, 0, m.texcoords)

	gl.EnableClientState(gl.COLOR_ARRAY)
	gl.ColorPointer(3, gl.FLOAT, 0, m.colors)

	// draw solids
	gl.Enable(gl.COLOR_MATERIAL)
	// gl.DrawArrays(gl.TRIANGLE_STRIP, 0, len(m.vertices)/3)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.LineWidth(1.0)
	gl.Color4f(1, 1, 1, 1)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, len(m.vertices)/3)

	gl.PopAttrib()
	gl.PopMatrix()

}
