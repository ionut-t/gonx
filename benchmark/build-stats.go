package benchmark

import "github.com/ionut-t/gonx/utils"

type initialStats struct {
	Main      int64 `json:"main"`
	Runtime   int64 `json:"runtime"`
	Polyfills int64 `json:"polyfills"`
	Total     int64 `json:"total"`
}

type buildStats struct {
	Initial      initialStats `json:"initial"`
	Lazy         int64        `json:"lazy"`
	Assets       int64        `json:"assets"`
	Total        int64        `json:"total"`
	OverallTotal int64        `json:"overallTotal"` // includes assets
	Styles       int64        `json:"styles"`
}

func (stats *buildStats) String() string {
	return utils.PrettyJSON(stats)
}
