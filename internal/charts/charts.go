package charts

import (
	"bytes"
	"fitness-bot/internal/models"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

func GenerateProgressChart(exercises []*models.Exercise, exerciseName string) ([]byte, error) {
	if len(exercises) == 0 {
		return nil, nil
	}

	var xValues []time.Time
	var yValues []float64

	for i := len(exercises) - 1; i >= 0; i-- {
		xValues = append(xValues, exercises[i].CreatedAt)
		yValues = append(yValues, exercises[i].Weight)
	}

	graph := chart.Chart{
		Title: "Прогресс: " + exerciseName,
		TitleStyle: chart.Style{
			FontSize: 16,
		},
		Width:  800,
		Height: 400,
		XAxis: chart.XAxis{
			Name: "Дата",
			Style: chart.Style{
				FontSize: 10,
			},
			ValueFormatter: chart.TimeValueFormatterWithFormat("02.01"),
		},
		YAxis: chart.YAxis{
			Name: "Вес (кг)",
			Style: chart.Style{
				FontSize: 10,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
