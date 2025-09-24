package glumby

/*
 * TextureCache
 */

import (
	"paleotronic.com/gl"
)

/*
TextureCache acts a manager for named textures, so that we can ensure that the same texture is only ever loaded once.
*/
type TextureCache struct {
	textures map[string]*Texture
}

// Create a new TextureCache instance
func NewTextureCache() *TextureCache {
	this := &TextureCache{}
	this.textures = make(map[string]*Texture)
	return this
}

// Test if file is already in the cache
func (this *TextureCache) Exists(file string) bool {
	_, ok := this.textures[file]
	return ok
}

// Free texture from the cache, and unbind it from opengl
func (this *TextureCache) Free(file string) {
	t, _ := this.textures[file]
	t.Unbind()
	handle := t.Handle()
	gl.DeleteTextures(1, &handle)
	// free
	delete(this.textures, file)
}

// Get a texture by name
func (this *TextureCache) GetTexture(file string) (*Texture, error) {

	t, ok := this.textures[file]

	if ok {
		return t, nil
	}

	// Here if the texture exists as a file, we will load it
	t, err := NewTexture(file)
	if err != nil {
		return nil, err
	}

	this.textures[file] = t
	return t, nil
}

func (this *TextureCache) SetTexture(file string, t *Texture) {
	if this.Exists(file) {
		this.Free(file)
	}
	this.textures[file] = t
}
