package cronlogger

type Application struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

type AppConfig struct {
	Applications []Application `json:"applications,omitempty"`
	DefaultColor string        `json:"defaultColor,omitempty"`
}
