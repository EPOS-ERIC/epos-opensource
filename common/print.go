package common

import (
	"fmt"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func PrintError(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[ERROR]\t%s%s\n", colorRed, message, colorReset)
}

func PrintWarn(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[WARNING]\t%s%s\n", colorYellow, message, colorReset)
}

func PrintInfo(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[INFO]\t%s%s\n", colorBlue, message, colorReset)
}

func PrintStep(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[STEP]\t%s%s\n", colorCyan, message, colorReset)
}

func PrintWait(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[WAITING]\t%s%s\n", colorPurple, message, colorReset)
}

func PrintDone(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[DONE]\t%s%s\n", colorGreen, message, colorReset)
}
