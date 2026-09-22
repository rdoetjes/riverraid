package ui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DrawGlowCircle renders a soft radial glow around a center point.
func DrawGlowCircle(center rl.Vector2, radius float32, col rl.Color, layers int) {
	if layers <= 0 {
		layers = 4
	}
	baseAlpha := float32(col.A)
	for i := layers; i >= 1; i-- {
		factor := float32(i) / float32(layers)
		currentRadius := radius * factor
		alpha := uint8(baseAlpha * (1.0 - factor*0.7) * (1.0 / float32(layers)))
		rl.DrawCircleV(center, currentRadius, rl.Color{R: col.R, G: col.G, B: col.B, A: alpha})
	}
}

// DrawGlowLine renders an anti-aliased luminous vector beam with bloom.
func DrawGlowLine(start, end rl.Vector2, thick float32, col rl.Color) {
	// Outer glow
	outerCol := rl.Color{R: col.R, G: col.G, B: col.B, A: uint8(float32(col.A) * 0.25)}
	rl.DrawLineEx(start, end, thick*3.2, outerCol)

	// Mid glow
	midCol := rl.Color{R: col.R, G: col.G, B: col.B, A: uint8(float32(col.A) * 0.6)}
	rl.DrawLineEx(start, end, thick*1.8, midCol)

	// Core bright line
	rl.DrawLineEx(start, end, thick, col)
}

// DrawThickPolygonOutline draws connected lines along polygon vertices.
func DrawThickPolygonOutline(points []rl.Vector2, thick float32, col rl.Color) {
	n := len(points)
	if n < 2 {
		return
	}
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		rl.DrawLineEx(points[i], points[next], thick, col)
	}
}

// DrawConvexPolygonFilled fills a polygon defined by points using triangle fan.
func DrawConvexPolygonFilled(points []rl.Vector2, col rl.Color) {
	if len(points) < 3 {
		return
	}
	root := points[0]
	for i := 1; i < len(points)-1; i++ {
		rl.DrawTriangle(root, points[i], points[i+1], col)
	}
}

// DrawDropShadow draws a darkened translucent silhouette offset by altitude.
func DrawDropShadow(points []rl.Vector2, offset rl.Vector2, alpha uint8) {
	if len(points) < 3 {
		return
	}
	shadowPoints := make([]rl.Vector2, len(points))
	for i, p := range points {
		shadowPoints[i] = rl.Vector2Add(p, offset)
	}
	shadowColor := rl.Color{R: 5, G: 15, B: 25, A: alpha}
	DrawConvexPolygonFilled(shadowPoints, shadowColor)
}

// RotatePoint rotates a 2D point around an origin by radians.
func RotatePoint(pt, origin rl.Vector2, angleRad float32) rl.Vector2 {
	s := float32(math.Sin(float64(angleRad)))
	c := float32(math.Cos(float64(angleRad)))

	p := rl.Vector2Subtract(pt, origin)
	xNew := p.X*c - p.Y*s
	yNew := p.X*s + p.Y*c

	return rl.Vector2{X: xNew + origin.X, Y: yNew + origin.Y}
}

// TransformPoints rotates, scales, and translates a slice of local points.
func TransformPoints(localPoints []rl.Vector2, center rl.Vector2, scale float32, angleRad float32) []rl.Vector2 {
	res := make([]rl.Vector2, len(localPoints))
	s := float32(math.Sin(float64(angleRad)))
	c := float32(math.Cos(float64(angleRad)))

	for i, pt := range localPoints {
		scaledX := pt.X * scale
		scaledY := pt.Y * scale

		rotX := scaledX*c - scaledY*s
		rotY := scaledX*s + scaledY*c

		res[i] = rl.Vector2{X: rotX + center.X, Y: rotY + center.Y}
	}
	return res
}

// DrawBeveledRect renders a futuristic vector panel with angled corner chamfers.
func DrawBeveledRect(bounds rl.Rectangle, chamfer float32, fillColor, borderColor rl.Color, borderThick float32) {
	x := bounds.X
	y := bounds.Y
	w := bounds.Width
	h := bounds.Height
	c := chamfer
	if c > w/2 || c > h/2 {
		c = float32(math.Min(float64(w/2), float64(h/2)))
	}

	pts := []rl.Vector2{
		{X: x + c, Y: y},
		{X: x + w - c, Y: y},
		{X: x + w, Y: y + c},
		{X: x + w, Y: y + h - c},
		{X: x + w - c, Y: y + h},
		{X: x + c, Y: y + h},
		{X: x, Y: y + h - c},
		{X: x, Y: y + c},
	}

	// Fill polygon
	DrawConvexPolygonFilled(pts, fillColor)

	if borderThick > 0 {
		DrawThickPolygonOutline(pts, borderThick, borderColor)
	}
}

// DrawProgressBar renders a vector gauge bar with gradient and ticks.
func DrawProgressBar(bounds rl.Rectangle, progress float32, fillColor, emptyColor, borderColor rl.Color) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	// Background
	rl.DrawRectangleRec(bounds, emptyColor)

	// Fill
	fillRec := rl.Rectangle{
		X:      bounds.X + 2,
		Y:      bounds.Y + 2,
		Width:  (bounds.Width - 4) * progress,
		Height: bounds.Height - 4,
	}
	if fillRec.Width > 0 {
		rl.DrawRectangleRec(fillRec, fillColor)
		// Highlight glint line on top
		rl.DrawLineEx(
			rl.Vector2{X: fillRec.X, Y: fillRec.Y + 1},
			rl.Vector2{X: fillRec.X + fillRec.Width, Y: fillRec.Y + 1},
			1.5,
			rl.Color{R: 255, G: 255, B: 255, A: 140},
		)
	}

	// Border
	rl.DrawRectangleLinesEx(bounds, 1.5, borderColor)
}
