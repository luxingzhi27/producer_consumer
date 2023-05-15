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
	"fyne.io/fyne/v2/theme"
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
	str, err := proNumStr.Get() //获得生产者数量
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	proNum, err = strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(proNum)
	str, err = conNumStr.Get() //获得消费者数量
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	conNum, err = strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(conNum)
	wg.Add(2)
	go producer(&wg, proNum) //启动生产者和消费者进程
	go consumer(&wg, conNum)
	wg.Wait()
	fmt.Printf("All Done\n")
}


func end(){
	doneMutex.Lock()
	done=true
	fmt.Println(done)
	doneMutex.Unlock()
	mutex.Lock()
	buffer=buffer[:0]
	percentage.Set(float64(len(buffer))/float64(BufferSize))
	mutex.Unlock()
}

func producer(wg *sync.WaitGroup, proNum int) {
	defer wg.Done()
	//启动生产者协程
	for i := 0; i < proNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for !done {
				mutex.Lock()                    //获取锁
				for len(buffer) == BufferSize { //缓冲区满时堵塞
					cond.Wait()
				}
				if done {
					mutex.Unlock()
					cond.Signal()
					fmt.Printf("Producer(%d) will be done\n",i)
					break
				}
				time.Sleep(1 * time.Second)
				item := len(buffer) + 1
				buffer = append(buffer, item)
				show.Set(fmt.Sprintf("Producer(%d) produces item(%d)", i, item)) //设置展示标签
				percentage.Set(float64(item) / float64(BufferSize))              //设置进度条百分比
				cond.Signal()
				mutex.Unlock() //释放锁
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
				if done{
					mutex.Unlock()
					fmt.Printf("Consumer(%d) will be done\n",i)
					break
				}
				for len(buffer) == 0 { //缓冲区空时堵塞
					cond.Wait()
				}
				if done {
					mutex.Unlock()
					cond.Signal()
					fmt.Printf("Producer(%d) will be done\n",i)
					break
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
}

func entry(){
	wg.Add(1)
	go begin()
}

func main() {
	myApp := app.New()
	lightTheme:=theme.LightTheme()
	darkTheme:=theme.DarkTheme()
	myApp.Settings().SetTheme(lightTheme)
	window := myApp.NewWindow("producer-consumer")
	proNumStr = binding.NewString()
	conNumStr = binding.NewString()
	show = binding.NewString()
	inputProLabel := widget.NewLabel("Enter the number of producers")
	inputConLabel := widget.NewLabel("Enter the number of consumers")
	inputLabelContainer := container.New(layout.NewHBoxLayout(), inputProLabel, layout.NewSpacer(), inputConLabel) //生产者消费者输入提示容器
	inputProNum := widget.NewEntryWithData(proNumStr)
	inputConNum := widget.NewEntryWithData(conNumStr)
	inputBoxContainer := container.New(layout.NewHBoxLayout(), inputProNum, layout.NewSpacer(), inputConNum) //生产者消费者输入框容器
	showLabel := widget.NewLabelWithData(show)
	showLabelContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), showLabel, layout.NewSpacer()) //展示标签容器
	showBufferContainer := container.New(layout.NewHBoxLayout())                                                   //缓冲区内容展示容器哦
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
	progressBarContainer := container.New(layout.NewMaxLayout(), progressBar) //进度条容器
	startButton := widget.NewButton("Start", entry)
	themeButton := widget.NewButton("Switch Theme", func(){
		if myApp.Settings().Theme()==lightTheme{
			myApp.Settings().SetTheme(darkTheme)
		}else{
			myApp.Settings().SetTheme(lightTheme)
		}
	})
	endButton:=widget.NewButton("End",end)
	buttonContainer := container.New(layout.NewHBoxLayout(), startButton, endButton,layout.NewSpacer(),themeButton) //按钮容器
	allContainer := container.New(layout.NewVBoxLayout(), inputLabelContainer, inputBoxContainer, showLabelContainer, showBufferContainer, progressBarContainer, buttonContainer)
	window.SetContent(allContainer)
	window.ShowAndRun()
}
