package main

var vertexSource = `
void main() {
    gl_TexCoord[0] = gl_TextureMatrix[0] * gl_MultiTexCoord0;
    gl_Position = ftransform();
}
`

var fragmentSource = `
uniform sampler2D tex;
 
void main()
{
    // vec4 color = texture2D(tex,gl_TexCoord[0].st);
    // color.r = 1
    
    gl_FragColor = color;
}
`
