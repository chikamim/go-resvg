package resvg

/*
#cgo LDFLAGS: -L./lib -lresvg
#cgo pkg-config: cairo
#include <stdlib.h>
#include <cairo.h>
#define RESVG_CAIRO_BACKEND 1
#include "resvg.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"image"
	"unsafe"

	cairo "github.com/ungerik/go-cairo"
)

type Options struct {
	Width           int
	Height          int
	DPI             float64
	BackgroundColor string
}

func (o *Options) ResvgOption() *C.struct_resvg_options {
	opt := &C.struct_resvg_options{}
	C.resvg_init_options(opt)
	opt.dpi = C.double(o.DPI)

	if o.Width > 0 {
		opt.fit_to = C.struct_resvg_fit_to{C.RESVG_FIT_TO_WIDTH, C.float(o.Width)}
	}

	if o.Height > 0 {
		opt.fit_to = C.struct_resvg_fit_to{C.RESVG_FIT_TO_HEIGHT, C.float(o.Height)}
	}

	if len(o.BackgroundColor) > 0 {
		opt.draw_background = true
		opt.background = o.ResvgBackgroundColor()
	}

	return opt
}

func (o *Options) ResvgBackgroundColor() C.struct_resvg_color {
	r, g, b, _ := hexColorToRGB(o.BackgroundColor)
	return C.struct_resvg_color{C.uchar(r), C.uchar(g), C.uchar(b)}
}

func hexColorToRGB(hex string) (r, g, b uint8, err error) {
	format := "#%02x%02x%02x"
	n, err := fmt.Sscanf(hex, format, &r, &g, &b)
	if err != nil {
		return
	}
	if n != 3 {
		err = errors.New("invalid hex color")
		return
	}

	return r, g, b, nil
}

func RenderPNGFromFile(svgpath, pngpath string, option *Options) error {
	svgpathC := C.CString(svgpath)
	defer C.free(unsafe.Pointer(svgpathC))
	pngpathC := C.CString(pngpath)
	defer C.free(unsafe.Pointer(pngpathC))

	tree := &C.struct_resvg_render_tree{}
	opt := option.ResvgOption()

	res := C.resvg_parse_tree_from_file(svgpathC, opt, &tree)
	defer C.resvg_tree_destroy(tree)
	if res != 0 {
		return resvgError(res)
	}
	res = C.resvg_cairo_render_to_image(tree, opt, pngpathC)
	if res != 0 {
		return resvgError(res)
	}
	return nil
}

func RenderImageWithIDFromFile(svg, id string, option *Options) (img image.Image, err error) {
	svgC := C.CString(svg)
	defer C.free(unsafe.Pointer(svgC))
	idC := C.CString(id)
	defer C.free(unsafe.Pointer(idC))

	size := C.struct_resvg_size{}
	size.width = C.uint(option.Width)
	size.height = C.uint(option.Height)

	surface := C.cairo_image_surface_create(C.CAIRO_FORMAT_ARGB32, C.int(option.Width), C.int(option.Height))
	defer C.cairo_surface_destroy(surface)
	ctx := C.cairo_create(surface)
	defer C.cairo_destroy(ctx)

	tree := &C.struct_resvg_render_tree{}

	opt := option.ResvgOption()

	res := C.resvg_parse_tree_from_data(svgC, C.ulong(len(svg)), opt, &tree)
	defer C.resvg_tree_destroy(tree)
	if res != 0 {
		return img, resvgError(res)
	}
	C.resvg_cairo_render_to_canvas_by_id(tree, opt, size, idC, ctx)
	s := cairo.NewSurfaceFromC((cairo.Cairo_surface)(unsafe.Pointer(surface)), (cairo.Cairo_context)(unsafe.Pointer(ctx)))

	return s.GetImage(), nil
}

func RenderPNGFromString(svg, pngpath string, option *Options) error {
	svgC := C.CString(svg)
	defer C.free(unsafe.Pointer(svgC))
	pngpathC := C.CString(pngpath)
	defer C.free(unsafe.Pointer(pngpathC))

	tree := &C.struct_resvg_render_tree{}
	opt := option.ResvgOption()

	res := C.resvg_parse_tree_from_data(svgC, C.ulong(len(svg)), opt, &tree)
	defer C.resvg_tree_destroy(tree)
	if res != 0 {
		return resvgError(res)
	}
	res = C.resvg_cairo_render_to_image(tree, opt, pngpathC)
	if res != 0 {
		return resvgError(res)
	}
	return nil
}

func RenderImageFromString(svg string, option *Options) (img image.Image, err error) {
	svgC := C.CString(svg)
	defer C.free(unsafe.Pointer(svgC))

	size := C.struct_resvg_size{}
	size.width = C.uint(option.Width)
	size.height = C.uint(option.Height)

	surface := C.cairo_image_surface_create(C.CAIRO_FORMAT_ARGB32, C.int(option.Width), C.int(option.Height))
	defer C.cairo_surface_destroy(surface)
	ctx := C.cairo_create(surface)
	defer C.cairo_destroy(ctx)

	tree := &C.struct_resvg_render_tree{}

	opt := option.ResvgOption()

	res := C.resvg_parse_tree_from_data(svgC, C.ulong(len(svg)), opt, &tree)
	defer C.resvg_tree_destroy(tree)
	if res != 0 {
		return img, resvgError(res)
	}
	C.resvg_cairo_render_to_canvas(tree, opt, size, ctx)
	s := cairo.NewSurfaceFromC((cairo.Cairo_surface)(unsafe.Pointer(surface)), (cairo.Cairo_context)(unsafe.Pointer(ctx)))

	return s.GetImage(), nil
}

func resvgError(enum C.int) error {
	switch enum {
	case C.RESVG_ERROR_NOT_AN_UTF8_STR:
		return errors.New("only UTF-8 content are supported")
	case C.RESVG_ERROR_FILE_OPEN_FAILED:
		return errors.New("failed to open the provided file")
	case C.RESVG_ERROR_FILE_WRITE_FAILED:
		return errors.New("failed to write to the provided file")
	case C.RESVG_ERROR_INVALID_FILE_SUFFIX:
		return errors.New("only \\b svg and \\b svgz suffixes are supported")
	case C.RESVG_ERROR_MALFORMED_GZIP:
		return errors.New("compressed SVG must use the GZip algorithm")
	case C.RESVG_ERROR_PARSING_FAILED:
		return errors.New("failed to parse an SVG data")
	case C.RESVG_ERROR_NO_CANVAS:
		return errors.New("failed to allocate an image")
	default:
		return errors.New("unknown error")
	}
}
