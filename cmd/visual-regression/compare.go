package main

import (
	"image"
	"image/color"
)

type Report struct {
	SameSize        bool    `json:"same_size"`
	WidthOld        int     `json:"width_old"`
	HeightOld       int     `json:"height_old"`
	WidthNew        int     `json:"width_new"`
	HeightNew       int     `json:"height_new"`
	ChangedPixels   int     `json:"changed_pixels"`
	TotalPixels     int     `json:"total_pixels"`
	DifferenceRatio float64 `json:"difference_ratio"`
}

func Compare(oldImage, newImage image.Image) (Report, *image.NRGBA) {
	oldBounds, newBounds := oldImage.Bounds(), newImage.Bounds()
	report := Report{SameSize: oldBounds.Dx() == newBounds.Dx() && oldBounds.Dy() == newBounds.Dy(), WidthOld: oldBounds.Dx(), HeightOld: oldBounds.Dy(), WidthNew: newBounds.Dx(), HeightNew: newBounds.Dy()}
	width, height := min(oldBounds.Dx(), newBounds.Dx()), min(oldBounds.Dy(), newBounds.Dy())
	report.TotalPixels = width * height
	diff := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			or, og, ob, oa := oldImage.At(oldBounds.Min.X+x, oldBounds.Min.Y+y).RGBA()
			nr, ng, nb, na := newImage.At(newBounds.Min.X+x, newBounds.Min.Y+y).RGBA()
			if or != nr || og != ng || ob != nb || oa != na {
				report.ChangedPixels++
				diff.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			}
		}
	}
	if report.TotalPixels > 0 {
		report.DifferenceRatio = float64(report.ChangedPixels) / float64(report.TotalPixels)
	}
	return report, diff
}
