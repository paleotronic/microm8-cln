package glumby

type MBOManager struct {
	blocks []*MeshBufferObject
	Count  int
}

func NewMBOManager() *MBOManager {
	m := &MBOManager{
		blocks: make([]*MeshBufferObject, 0),
		Count:  0,
	}
	return m
}

func (m *MBOManager) MBO(index int) *MeshBufferObject {
	for index > len(m.blocks)-1 {
		m.blocks = append(m.blocks, &MeshBufferObject{})
		m.Count++
	}
	return m.blocks[index]
}

func (m *MBOManager) EnsureCapacity(width, height, verticesPerItem int) {
	blocksNeeded := (width*height*verticesPerItem)/MAX_MESH_VERTICES + 1
	for len(m.blocks) < blocksNeeded {
		m.blocks = append(m.blocks, &MeshBufferObject{})
		m.Count++
	}
}

func (m *MBOManager) Free() {
	for i, _ := range m.blocks {
		m.blocks[i] = nil
	}
}
