package models

type PrintOptions struct {
	PageHeight  float64          `json:"page_height"`
	PageWidth   float64          `json:"page_width"`
	Orientation PrintOrientation `json:"orientation"`
}

type PrintOrientation string

const (
	Orientation_Landscape = "landscape"
	Orientation_Portrait  = "portrait"
)
