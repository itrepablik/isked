![isked](https://user-images.githubusercontent.com/58651329/87672135-096ea980-c7a5-11ea-82cb-38267c9d34a6.png)
The **isked** package is the simplified and easy to use task schedulers for your Go projects.

# Installation
```go
go get -u github.com/itrepablik/isked
```

# Usage
These are some of the examples on how you can use isked package.
```go
package main

import (
	"fmt"

	"github.com/itrepablik/isked"
)

func myFunc1() {
	fmt.Println("calling from myFunc1(): ")
}

func myFunc2(s string) func() {
	return func() {
		fmt.Println("calling from myFunc4(): ", s)
	}
}

func main() {
	// Frequently methods:
	isked.TaskName("Task 1").Frequently().Seconds(7).ExecFunc(myFunc2("hello world")).AddTask()
	isked.TaskName("Task 2").Frequently().Minutes(1).ExecFunc(myFunc1).AddTask()
	isked.TaskName("Task 3").Frequently().Hours(2).ExecFunc(myFunc1).AddTask()

	// Daily methods:
	// The start date is tomorrow's date, not, today's date
	isked.TaskName("Task 4").Daily().At("14:18").ExecFunc(myFunc1).AddTask()

	// Weekly methods:
	isked.TaskName("Task 6").Weekly().Tuesday().At("17:30").ExecFunc(myFunc1).AddTask()

	// Monthly methods: 0 - means 'last day' of each month
	isked.TaskName("Task 7").Monthly().Every(0).At("09:30").ExecFunc(myFunc1).AddTask()
	isked.TaskName("Task 8").Monthly().Every(2).At("10:30").ExecFunc(myFunc1).AddTask()

	isked.Run()
}
```

# Subscribe to Maharlikans Code Youtube Channel:
Please consider subscribing to my Youtube Channel to recognize my work on any of my tutorial series. Thank you so much for your support!
https://www.youtube.com/c/MaharlikansCode?sub_confirmation=1

# License
Code is distributed under MIT license, feel free to use it in your proprietary projects as well.
