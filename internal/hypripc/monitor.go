package hypripc

type Monitor struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	Serial          string  `json:"serial"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	RefreshRate     float64 `json:"refreshRate"`
	X               int     `json:"x"`
	Y               int     `json:"y"`
	ActiveWorkspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activeWorkspace"`
	SpecialWorkspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"specialWorkspace"`
	Reserved        []int   `json:"reserved"`
	Scale           float64 `json:"scale"`
	Transform       int     `json:"transform"`
	Focused         bool    `json:"focused"`
	DpmsStatus      bool    `json:"dpmsStatus"`
	Vrr             bool    `json:"vrr"`
	ActivelyTearing bool    `json:"activelyTearing"`
}
