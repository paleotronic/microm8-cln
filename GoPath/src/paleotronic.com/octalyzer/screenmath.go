package main

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const FOV_TARGET_H = 90
const FOV_DEFAULT_V = 59.103

func screenPlaneDistance(dist float64, degrees float64) float64 {
	return (dist / 2) / math.Tan(mgl64.DegToRad(degrees)/2)
}

func fovCalc(spd float64, size float64) float64 {
	return mgl64.RadToDeg(math.Atan((size/2)/spd) * 2)
}

/*

#  tan(FOV_H/2) / screen_width = tan(FOV_V/2) / screen_height

> (tan(FOV_H/2) / screen_width) * screen_height = tan(FOV_V/2)

> atan( (tan(FOV_H/2) / screen_width) * screen_height ) = FOV_V / 2

> 2 * atan( (tan(FOV_H/2) / screen_width) * screen_height ) = FOV_V

*/

// Calculate the windows vertical FOV based on maintaining a strict horizontal
// FOV
func VFOVFromHFOV(fovH float64, screenWidth, screenHeight float64) float64 {

	return mgl64.RadToDeg(
		2 * math.Atan(
			(math.Tan(mgl64.DegToRad(fovH/2))/screenWidth)*screenHeight,
		),
	)

}
