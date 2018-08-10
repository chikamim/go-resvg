package resvg

import (
	"image/png"
	"io/ioutil"
	"os"
	"testing"
)

const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="-52 -53 100 100" stroke-width="2">
 <g fill="none">
  <ellipse stroke="#66899a" rx="6" ry="44"/>
  <ellipse stroke="#e1d85d" rx="6" ry="44" transform="rotate(-66)"/>
  <ellipse stroke="#80a3cf" rx="6" ry="44" transform="rotate(66)"/>
  <circle  stroke="#4b541f" r="44"/>
 </g>
 <g fill="#66899a" stroke="white">
  <circle fill="#80a3cf" r="13"/>
  <circle cy="-44" r="9"/>
  <circle cx="-40" cy="18" r="9"/>
  <circle cx="40" cy="18" r="9"/>
 </g>
</svg>`

func TestRenderPNGFromFile(t *testing.T) {
	ioutil.WriteFile("test.svg", []byte(svg), 0666)
	err := RenderPNGFromFile("test.svg", "test.png", &Options{BackgroundColor: "#eeddcc"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderPNGFromString(t *testing.T) {
	err := RenderPNGFromString(svg, "svg.png", &Options{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderImageFromString(t *testing.T) {
	img, err := RenderImageFromString(svg, &Options{Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	outfile, _ := os.Create("out.png")
	defer outfile.Close()
	png.Encode(outfile, img)
}

func TestHexColorToRGB(t *testing.T) {
	r, g, b, _ := hexColorToRGB("#01aaff")
	if r != 1 || g != 170 || b != 255 {
		t.Fatalf("fail %v %v %v", r, g, b)
	}
}
