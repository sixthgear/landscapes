package main

import "math"
import "noise"
import "github.com/go-gl/gl"

// import "log"
// import "fmt"
// import "math/rand"
// import "github.com/go-gl/glfw"

type Vertex struct {
	x, y, z float32
}

var colorMap [][3]float32 = [][3]float32{
	{0.0, 0.0, 0.5},
	{0.0, 0.0, 0.7},
	{0.1, 0.1, 0.9},
	{1.0, 1.0, 0.9},
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
	{1.0, 1.0, 1.0},
}

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

	diag := math.Hypot(float64(m.width/2), float64(m.depth/2))
	for z := 0; z < depth; z++ {
		for x := 0; x < width; x++ {
			fx := float64(x)
			fz := float64(z)
			d := math.Hypot(float64(m.width/2)-fx, float64(m.depth/2)-fz)
			d = 1.0 - d/diag
			h := noise.OctaveNoise2d(fx, fz, 4, 0.25, 1.0/120)
			h = (h + 1.0) * 0.5
			h = (h + d*0.25) * 768 * (d * d)
			m.heightMap = append(m.heightMap, float32(h))
			m.minHeight = math.Min(m.minHeight, h)
			m.maxHeight = math.Max(m.maxHeight, h)
		}
	}

	// gl.Enable(gl.TEXTURE_2D)
	// m.texture = gl.GenTexture()
	// m.texture.Bind(gl.TEXTURE_2D)

	// if !glfw.LoadTexture2D("rock.tga", glfw.BuildMipmapsBit) {
	// 	log.Fatal("Failed to load texture!")
	// }

	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	// m.texture.Unbind(gl.TEXTURE_2D)

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
			color[i] += float32(noise.OctaveNoise3d(float64(v.x), float64(v.y), float64(v.z), 4, 0.5, 2.0)) * 0.05
		}

	}
	return color
}

func (m *Map) BuildVertices() {

	s := m.gridSize

	for i := 0; i < (m.depth-1)*(m.width); i++ {

		var v0, v1 Vertex
		var n0, n1 Vertex
		x0 := 0
		z0 := i / (m.width)
		z1 := z0 + 1

		skip := false

		if z0%2 == 0 {
			// even rows go right
			x0 = i % (m.width)
			skip = x0 == 0 && z0 != 0
		} else {
			// odd rows go left                
			x0 = (m.width - 1) - (i % (m.width))
			skip = x0 == m.width-1
		}

		if !skip {
			v0 = Vertex{float32(x0 * s), m.heightMap[z0*m.width+x0], float32(z0 * s)}
			n0 = m.GetNormal(v0, x0, z0)
			c0 := m.getColorForVertex(v0)
			m.vertices = append(m.vertices, v0.x, v0.y, v0.z)
			m.normals = append(m.normals, n0.x, n0.y, n0.z)
			m.texcoords = append(m.texcoords, float32(x0)/8, -float32(z0)/8)
			m.colors = append(m.colors, c0[0], c0[1], c0[2])
		}

		v1 = Vertex{float32(x0 * s), m.heightMap[z1*m.width+x0], float32(z1 * s)}
		n1 = m.GetNormal(v1, x0, z1)
		c1 := m.getColorForVertex(v1)
		m.vertices = append(m.vertices, v1.x, v1.y, v1.z)
		m.normals = append(m.normals, n1.x, n1.y, n1.z)
		m.texcoords = append(m.texcoords, float32(x0)/8, -float32(z1)/8)
		m.colors = append(m.colors, c1[0], c1[1], c1[2])
	}

}

func (m *Map) GetNormal(v Vertex, x, z int) (normal Vertex) {

	s := m.gridSize

	// if normal_cache.has_key((x,z)):
	//     return normal_cache[(x,z)]

	var neighbors []Vertex
	var sum Vertex

	if x > 0 {
		neighbors = append(neighbors, Vertex{float32((x - 1) * s), m.heightMap[z*m.width+(x-1)], float32(z * s)})
	}
	if z > 0 {
		neighbors = append(neighbors, Vertex{float32(x * s), m.heightMap[(z-1)*m.width+x], float32((z - 1) * s)})
	}
	if x < m.width-1 {
		neighbors = append(neighbors, Vertex{float32((x + 1) * s), m.heightMap[z*m.width+(x+1)], float32(z * s)})
	}
	if z < m.depth-1 {
		neighbors = append(neighbors, Vertex{float32(x * s), m.heightMap[(z+1)*m.width+x], float32((z + 1) * s)})
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
	// normal_cache[(x,z)] = normal
	return normal
}

func (m *Map) Draw() {

	gl.PushMatrix()
	gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)

	// set wireframe details
	gl.LineWidth(0.4)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.VertexPointer(3, gl.FLOAT, 0, m.vertices)

	gl.EnableClientState(gl.NORMAL_ARRAY)
	gl.NormalPointer(gl.FLOAT, 0, m.normals)

	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.TexCoordPointer(2, gl.FLOAT, 0, m.texcoords)

	gl.EnableClientState(gl.COLOR_ARRAY)
	gl.ColorPointer(3, gl.FLOAT, 0, m.colors)

	//draw solids
	gl.Enable(gl.COLOR_MATERIAL)
	gl.Color4f(1, 1, 1, 1)
	// gl.Enable(gl.TEXTURE_2D)
	// m.texture.Bind(gl.TEXTURE_2D)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, len(m.vertices)/3)
	// m.texture.Unbind(gl.TEXTURE_2D)

	// draw wireframe
	// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	// gl.Color4f(0.1, 0.1, 0.1, 0.2)
	// gl.DrawArrays(gl.TRIANGLE_STRIP, 0, len(m.vertices)/3)

	gl.PopAttrib()
	gl.PopMatrix()

}
