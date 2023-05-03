package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const BufferSize = 20

var buffer []int
var mutex sync.Mutex            //互斥锁对象
var cond = sync.NewCond(&mutex) //条件变量对象
var show binding.String         //显示当前生产者消费者工作情况
var proNum int
var conNum int
var proNumStr binding.String
var conNumStr binding.String
var percentage binding.Float
var done = true
var doneMutex sync.Mutex
var wg sync.WaitGroup

func begin() {
	doneMutex.Lock()
	done = false
	doneMutex.Unlock()
	str, err := proNumStr.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	proNum, err = strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	str, err = conNumStr.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	conNum, err = strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	wg.Add(2)
	go producer(&wg, proNum)
	go consumer(&wg, conNum)
	wg.Wait()
	fmt.Printf("All Done\n")
}

func producer(wg *sync.WaitGroup, proNum int) {
	defer wg.Done()
	for i := 0; i < proNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for !done {
				mutex.Lock()
				for len(buffer) == BufferSize {
					cond.Wait()
				}
				time.Sleep(1 * time.Second)
				item := len(buffer) + 1
				buffer = append(buffer, item)
				show.Set(fmt.Sprintf("Producer(%d) produces item(%d)", i, item))
				percentage.Set(float64(item) / float64(BufferSize))
				cond.Signal()
				mutex.Unlock()
			}
		}(i)
	}
}

func consumer(wg *sync.WaitGroup, conNum int) {
	defer wg.Done()
	for i := 0; i < conNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for !done {
				mutex.Lock()
				for len(buffer) == 0 {
					cond.Wait()
				}
				time.Sleep(1 * time.Second)
				item := len(buffer)
				buffer = buffer[0 : item-1]
				show.Set(fmt.Sprintf("Consumer(%d) consumes item(%d)", i, item))
				percentage.Set(float64(item-1) / float64(BufferSize))
				cond.Signal()
				mutex.Unlock()
			}
		}(i)
	}
	wg.Done()
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("producer-consumer")
	proNumStr = binding.NewString()
	conNumStr = binding.NewString()
	show = binding.NewString()
	inputProLabel := widget.NewLabel("Enter the number of producers")
	inputConLabel := widget.NewLabel("Enter the number of consumers")
	inputLabelContainer := container.New(layout.NewHBoxLayout(), inputProLabel, layout.NewSpacer(), inputConLabel)
	inputProNum := widget.NewEntryWithData(proNumStr)
	inputConNum := widget.NewEntryWithData(conNumStr)
	inputBoxContainer := container.New(layout.NewHBoxLayout(), inputProNum, layout.NewSpacer(), inputConNum)
	showLabel := widget.NewLabelWithData(show)
	showLabelContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), showLabel, layout.NewSpacer())
	showBufferContainer := container.New(layout.NewHBoxLayout())
	grids := make([]*widget.Label, 20)
	for i := 0; i < BufferSize; i++ {
		grids[i] = widget.NewLabel(strconv.Itoa(i + 1))
	}
	for _, grid := range grids {
		showBufferContainer.Add(grid)
	}
	percentage = binding.NewFloat()
	percentage.Set(float64(len(buffer)) / float64(BufferSize))
	progressBar := widget.NewProgressBarWithData(percentage)
	progressBarContainer := container.New(layout.NewMaxLayout(), progressBar)
	startButton := widget.NewButton("Start", begin)
	buttonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), startButton, layout.NewSpacer())
	allContainer := container.New(layout.NewVBoxLayout(), inputLabelContainer, inputBoxContainer, showLabelContainer, showBufferContainer, progressBarContainer, buttonContainer)
	window.SetContent(allContainer)
	window.ShowAndRun()
}
