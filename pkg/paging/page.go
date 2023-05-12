package paging

const PAGESIZE = 4096

type Page struct {
	data2        [PAGESIZE]byte
	currentIndex int
}

func NewPage() *Page {
	p := &Page{
		currentIndex: 0,
	}

	return p
}

func (p *Page) hasSufficientSpace(newData []byte) bool {
	// if (len(p.data) + len(newData)) > cap(p.data) {
	// 	return false
	// } else {
	// 	return true
	// }

	if (p.currentIndex + len(newData)) >= PAGESIZE {
		return false
	} else {
		return true
	}
}

func (p *Page) appendBytes(newData []byte) {
	copy(p.data2[p.currentIndex:p.currentIndex+len(newData)], newData)
	p.currentIndex += len(newData)
}

func (p *Page) getRelevantLen() uint32 {
	return uint32(p.currentIndex)
}
