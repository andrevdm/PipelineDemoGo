package main

import (
	"fmt"
	"os"
	"bufio"
	"strconv"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// Demo of a pipeline of steps processing data
//
//  1) A ValueEvent is created and sent to the pipeline channel
//  2) The channel runs the value through its steps
//  3) For now the printStep is used to show the updated state
//
///////////////////////////////////////////////////////////////////////////////////////////////////


//Event containing the value to process
type ValueEvent struct {
	Value float64
	CustomData map[string]string
}

//The current state of a pipeline.
// Note this this is expected to be immutable
type PipelineState struct {
	ValueText string
	History []ValueEvent
	AnalyserData map[string]string
}


//Step to print the current state
func pringStep(evt ValueEvent, st PipelineState) PipelineState{
	fmt.Println( "pring" )
	fmt.Printf( "   %+v\n", st )
	return st
}


//Step to store history up to maxDepth items
// See mkAddHistoryStep
func addHistoryStep( maxDepth int, evt ValueEvent, st PipelineState) PipelineState{
	hst := append( st.History, evt)

	if len(hst) > maxDepth{
		hst = append( hst[:0], hst[1:]...)
	}
	
	st.History = hst
	return st
}

//A step function should only have to params. To pass the maxDepth to addHistoryStep this function
// is used to 'curry' the addHistoryStep function. 
func mkAddHistoryStep( maxDepth int ) func(ValueEvent, PipelineState)PipelineState{
	return func(evt ValueEvent, st PipelineState) PipelineState{
		return addHistoryStep( maxDepth, evt, st)
	}
}

//Step to calculate the average of the stored values
func avgOverHistoryStep(evt ValueEvent, st PipelineState) PipelineState{
	v := 0.0

	//Sum of values
	for _, h := range st.History{
		v += h.Value
	}

	//Average
	if( len(st.History) > 0 ){
		v = v / float64(len(st.History))
	}

	st.ValueText = fmt.Sprintf( "%.2f", v)
	return st
}

//Build a pipeline channel
func buildPipeline(name string, steps []func(ValueEvent, PipelineState)PipelineState) chan ValueEvent {
	chn := make(chan ValueEvent)
	state := PipelineState{AnalyserData: make(map[string]string)}
	
	go func(){
		for{
			evt := <- chn

			//Fold over the state
			for _, fn := range steps{
				state = fn( evt, state )
			}
			
		}
	}()

	return chn
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	var channels map[string]chan ValueEvent
	channels = make(map[string]chan ValueEvent)

	addHist := mkAddHistoryStep( 2 )
	channels["demo1"] = buildPipeline( "demo1", []func(ValueEvent, PipelineState)PipelineState{addHist, avgOverHistoryStep, pringStep} )

	//channels["demo2"] = buildPipeline( "demo2", []func(ValueEvent, PipelineState)PipelineState{addHist, runningAvgStep, pringStep} )

	for{
		text, _ := reader.ReadString('\n')
		v , err := strconv.ParseFloat( text[:len(text)-1], 64 )
		if err != nil {
			fmt.Printf( "Error getting float: %s", err )
		} else {
			channels["demo1"] <- ValueEvent{ Value: v, CustomData: map[string]string{"val":text} }
			//channels["demo2"] <- ValueEvent{ Value: v, CustomData: map[string]string{"val":text} }
		}
	}
}
