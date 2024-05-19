package predicator

import (
	"github.com/ccuetoh/nsfw/nsfw-main"
)

var predictor *nsfw.Predictor

func Predictor() *nsfw.Predictor {
	return predictor
}

func init() {
	predictor, _ = nsfw.NewLatestPredictor()
}
