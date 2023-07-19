package common

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	StartingLevel int
	UsingRedis    bool
)

func AskForContinueInt(msg string) int {

	fmt.Println(msg)
	fmt.Println("")
	var response string

	fmt.Scanln(&response)

	if response == "c" {
		os.Exit(2)
	}
	if intVar, err := strconv.Atoi(response); err == nil {
		return intVar
	}

	return 999
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func AskForLoopNum(msg string) int {

	fmt.Println(msg)
	fmt.Println("")
	var response string
	fmt.Println("How many loops to run? \ny(yes)-infinite\nn(no) cancel\n[num] number of loops ")
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	if intVar, err := strconv.Atoi(response); err == nil {

		return intVar
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return 10000
	case "n", "no":
		os.Exit(3)
		return 999
	default:
		return AskForLoopNum(msg)
	}
}
func Contains[K comparable](s []K, e K) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Check(msg string, e error) {

	if e != nil {
		fmt.Println(msg, e)
		os.Exit(1)
	}
}

func FileExists(absPath string) bool {
	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func ConvertStastoBTC(sats uint64) float64 {
	return float64(sats) / math.Pow10(8)
}

func round(f float64) int64 {
	if f < 0 {
		return int64(f - 0.5)
	}
	return int64(f + 0.5)
}
func ConvertBTCtoSats(btc float64) uint64 {
	sats := round(btc * 1e8)
	return uint64(sats)
}

func GetLastDump() string {
	files, err := ioutil.ReadDir("dumps/")
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	if len(files) == 0 {
		return ""
	}
	return strings.Replace(files[0].Name(), ".sql", "", 1)
}

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

func TruncateText(s string, max int) string {
	return s[:max]
}

func ArrayToString(a []uint, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

func WriteStrToFile(fn string, data string) {
	os.Truncate(fn, 0)

	file, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	Check("file didn't open", err)
	defer file.Close()

	_, err = file.WriteString(data)
	Check("write err", err)
	file.Sync()
}

func AppendStrtoFile(fn string, data string) {
	f, err := os.OpenFile(fn,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		log.Println(err)
	}
}

// ReadFileLines read each line of file in to string slice
func ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewReader(file)

	for {
		line, _, err := scanner.ReadLine()

		if err == io.EOF {
			break
		}

		lines = append(lines, string(line))
	}
	return lines, nil
}

// StartStoppingCh starts a go routine to check for incoming user input
func StartStoppingCh(ch chan<- bool) {
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) == "stop" {
			ch <- true
		} else {
			StartStoppingCh(ch)
		}
	}()

}

func LastTwoDigit(num int) string {
	s := strconv.Itoa(num)
	return s[len(s)-1:]
}

// SortandPrintMap prints outs a map[string]int oredered from largest to smallest
func SortandPrintMap(m map[string]int) {

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for _, kv := range ss {

		fmt.Printf("%s: %d\n", kv.Key, kv.Value)
	}

}
