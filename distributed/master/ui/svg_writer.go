package ui

import (
	"bytes"
	"fmt"

	"github.com/ajstarks/svgo"
	"github.com/chrislusf/gleam/pb"
)

var (
	WidthStep       = 20 * m
	HightStep       = 3 * m
	HightStepHeader = 2 * m
	Margin          = 5 * m
	LineLength      = 5 * m
	SmallMargin     = 4
	VerticalGap     = 4 * m
)

type stepGroupPosition struct {
	input  point
	output point
}

func GenSvg(status *pb.FlowExecutionStatus) string {

	var svgWriter bytes.Buffer
	canvas := svg.New(&svgWriter)

	width := 100 * m

	height := doFlowExecutionStatus(canvas, status, width)
	height += Margin

	svgWriter.Truncate(0)

	canvas.Start(width, height)
	doFlowExecutionStatus(canvas, status, width)
	canvas.End()

	return svgWriter.String()
}

func doFlowExecutionStatus(canvas *svg.SVG, status *pb.FlowExecutionStatus, width int) (height int) {

	positions := make([]stepGroupPosition, len(status.GetStepGroups()))
	layerOfStepGroupIds := toStepGroupLayers(status)

	height = Margin - LineLength

	for layer := len(layerOfStepGroupIds) - 1; layer >= 0; layer-- {
		stepGroupIds := layerOfStepGroupIds[layer]

		// determine input points
		for idx, stepGroupId := range stepGroupIds {
			positions[stepGroupId].input = point{
				width/2 + (2*idx-(len(stepGroupIds)-1))*(WidthStep+VerticalGap)/2,
				height + LineLength,
			}
		}

		for _, stepGroupId := range stepGroupIds {
			stepGroup := status.StepGroups[stepGroupId]

			for _, parentId := range stepGroup.GetParentIds() {
				connect(canvas, positions[parentId].output, positions[stepGroupId].input,
					fmt.Sprintf("d%d", getLastStep(status, status.StepGroups[parentId]).OutputDatasetId))
			}

			positions[stepGroupId].output = doStepGroup(canvas, positions[stepGroupId].input, status, stepGroup)

			if positions[stepGroupId].output.y > height {
				height = positions[stepGroupId].output.y
			}

		}
	}

	return height
}

func doState(canvas *svg.SVG, input point, state string) (output point) {
	r := 3 * m
	canvas.Circle(input.x, input.y+r, r)
	canvas.Text(input.x, input.y+r+0.5*m, state, "text-anchor:middle;font-size:20px;fill:white")
	return point{input.x, input.y + 2*r}
}

func doStepGroup(canvas *svg.SVG, input point, status *pb.FlowExecutionStatus, stepGroup *pb.FlowExecutionStatus_StepGroup) (output point) {
	x, y := input.x-WidthStep/2, input.y
	w, h := WidthStep, len(stepGroup.GetStepIds())*(HightStep+SmallMargin)+SmallMargin

	stepOut := point{input.x, input.y}
	for _, stepId := range stepGroup.StepIds {
		step := status.GetStep(stepId)
		stepOut = doStep(canvas, stepOut, step)
	}

	rectstyle := fmt.Sprintf("stroke:%s;stroke-width:1;fill:%s", "black", "none")

	canvas.Rect(x, y, w, h, rectstyle)

	output.x = input.x
	output.y = input.y + h

	return
}

func doStep(canvas *svg.SVG, input point, step *pb.FlowExecutionStatus_Step) (output point) {
	output.x = input.x
	output.y = input.y + HightStep + SmallMargin

	x, y := input.x-WidthStep/2+SmallMargin, input.y+SmallMargin
	w, h := WidthStep-2*SmallMargin, HightStep
	fs := 14
	w2 := 3

	name := step.GetName()
	canvas.Rect(x, y, w, h, fmt.Sprintf("stroke:%s;stroke-width:1;fill:%s", "black", "gray"))
	canvas.Gstyle(fmt.Sprintf("font-size:%dpx", fs))
	canvas.Text(x+w2, y+HightStep-SmallMargin/2, name, "stroke:black;baseline-shift:50%")
	canvas.Gend()

	return
}

func getLastStep(status *pb.FlowExecutionStatus, stepGroup *pb.FlowExecutionStatus_StepGroup) (step *pb.FlowExecutionStatus_Step) {
	return status.Steps[stepGroup.StepIds[len(stepGroup.StepIds)-1]]
}
