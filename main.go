package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	for {
		// Get current temperature(s)
		cmd := exec.Command("/usr/sbin/ipmitool", "-H", "192.168.1.99", "-U", "ADMIN", "-P", "ADMIN", "sdr", "type", "temperature")

		out, err := cmd.Output()
		if err != nil {
			panic(err)
		}

		// parsing regex
		re, err := regexp.Compile("[^|]*\\|[^|]*\\|[^|]*\\|[^|]*\\|[^0-9]*([0-9]+).*")
		if err != nil {
			panic(err)
		}

		temperatures := make([]int, 0, 3)

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			// parse out the temperature
			res := re.FindAllStringSubmatch(scanner.Text(), -1)
			if (len(res) > 0) && (len(res[0]) > 0) {
				i, err := strconv.Atoi(res[0][1])
				if err != nil {
					// bogus value? ignore
					continue
				}
				temperatures = append(temperatures, i)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}

		// work with the highest temperature of all we have
		curTemp := Max(temperatures)

		fmt.Printf("Current temperature: %v degrees celsius\r\n", curTemp)

		// Figure out our fan control based on that
		minTemp := 25 // no fans before this temperatuer
		maxTemp := 45 // max fans at this temperature

		rangeTemp := maxTemp - minTemp
		multi := 100 / float64(rangeTemp)

		t := Clamp((float64(curTemp-minTemp) * multi), 0, 100)
		fmt.Printf("Setting fan speed to %v%%\r\n", t)

		// scale this to a single byte (0-254)
		newSpeed := int(math.Ceil((255 / float64(100)) * t))
		fmt.Printf("Fan speed raw: %v (%v)\r\n", newSpeed, fmt.Sprintf("0x%x", newSpeed))

		// Send the final fan speed control command
		cmd = exec.Command("/usr/sbin/ipmitool", "-H", "192.168.1.99", "-U", "ADMIN", "-P", "ADMIN", "raw", "0x30", "0x70", "0x66", "0x01", "0x00", fmt.Sprintf("0x%x", newSpeed))
		err = cmd.Run()
		if err != nil {
			panic(err)
		}
		// Sleep until we next run (1 minute)
		time.Sleep(60 * time.Second)
	}
}

func Max(array []int) (max int) {
	max = array[0]
	for _, e := range array {
		if e > max {
			max = e
		}
	}
	return
}

func Clamp(val float64, min float64, max float64) float64 {
	//fmt.Println("Running with value: %v\r\n", val)
	if val > max {
		return max
	} else if val < min {
		return min
	}
	return val
}
