package main

import (
	"bufio"
	"code-generation/utils"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPrefixLength = 2
	maxPrefixLength = 6
	minCodeLength   = 4
	maxCodeLength   = 16
	maxNumCodes     = 100000000
	charset         = "0123456789"
	numWorkers      = 500
)

type AppConfig struct {
	Prefix        string
	Length        int
	NumCodes      int
	LineNumbers   bool
	Version       bool
	CodeTx        string
	FLineF        string
	HelpF         string
	RemainingTime time.Duration
	AppCommand    string
	OSSpecDir     string
	AppVer        string
}

type CodeResult struct {
	Code string
}

var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func main() {
	var config AppConfig
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&config.Prefix, "p", "", "Add prefix to the codes[2-6 characters]")
	flag.IntVar(&config.Length, "l", 6, "The length of the generated numbers[4-16 digits]")
	flag.IntVar(&config.NumCodes, "n", 1, "The number of generated codes[1-100 million]")
	flag.BoolVar(&config.LineNumbers, "a", false, "Add line numbers to the file")
	flag.BoolVar(&config.Version, "v", false, "About")
	flag.Parse()

	config.CodeTx = "CODES"
	config.FLineF = "codes were generated |"
	config.HelpF = "Each time you use a code, delete it. You can use [CTRL + X]"
	config.AppCommand = "codegen"
	config.OSSpecDir = "Documents"

	if runtime.GOOS == "linux" {
		config.OSSpecDir = ""
	}

	if flag.NFlag() == 0 {
		utils.Banner()
		fmt.Println()
		color.Green("%s %s %s %s %s", "Usage:\t", config.AppCommand, "[-p prefix] [-l length_number] [-n total_codes]\t\nexample:", config.AppCommand, "-p=FT -l=6 -n=100\t\n")
		color.Cyan("Options:\t\n\t-p\tAdd prefix to the codes (2-6 characters)\t\n\t-l\tThe length of the generated code number (4-16 digits)\t\n\t-n\tThe number of generated codes (1-100 million)\t\n\t-a\tAdd line numbers to the file\t\n\t-v\tAbout\t\n\t\n")
		color.Yellow("Tip: To stop code generation midway, simply press [CTRL + C]\n")
		return
	}

	if config.Version {
		utils.Banner()
	}

	if config.NumCodes == 1 {
		config.CodeTx = "CODE"
		config.FLineF = "code were generated at"
	}
	if !config.Version {
		if !validatePrefix(config.Prefix) {
			c := color.New(color.BgRed, color.FgBlack).Sprint("Error")
			fmt.Printf("%s Prefix length should be between %d and %d characters\n", c, minPrefixLength, maxPrefixLength)
			return
		}
		if !validateLength(config.Length) {
			c := color.New(color.BgRed, color.FgBlack).Sprint("Error")
			fmt.Printf("%s Code length should be between %d and %d digits\n", c, minCodeLength, maxCodeLength)
			return
		}
		if config.NumCodes <= 0 || config.NumCodes > maxNumCodes {
			c := color.New(color.BgRed, color.FgBlack).Sprint("Error")
			fmt.Printf("%s Number of codes should be between 1 and %d\n", c, maxNumCodes)
			return
		}
		psb := calculatePossibleOutcomes(config.Length)
		if psb == config.NumCodes {
			c := color.New(color.BgYellow, color.FgBlack).Sprint("Attention:")
			att := color.New(color.FgHiYellow).Add(color.Italic).Sprint("Code generation and saving may take a bit longer than usual based on the provided parameters!")
			fmt.Printf("%s %s\n",
				c, att)
		}
		if config.NumCodes > psb {
			c := color.New(color.BgYellow, color.FgBlack).Sprint("Warning")
			fmt.Printf("%s Maximum [%s] numbers can be created with a length of [%d]\n",
				c, humanize.Comma(int64(psb)), config.Length)
			return
		}

		hash, err := GenerateHash(generateCodeWithPool("X", 12))
		if err != nil {
			log.Fatalf("error %s", err)
		}

		homePath, err := os.UserHomeDir()
		if err != nil {
			color.Red("Error getting user home directory")
			return
		}

		path := filepath.Join(homePath, config.OSSpecDir, "Code Generator", "Files")
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			color.Red("Error creating folder: %v", errDir)
			return
		}

		filePath := filepath.Join(path, fmt.Sprintf("[%s]-%s_%v.txt", config.Prefix, config.CodeTx, sanitizeKey(hash)))

		file, err := os.Create(filePath)
		if err != nil {
			color.Red("Error creating file: %v", err)
			return
		}
		defer file.Close()

		loadingTicker := time.NewTicker(100 * time.Millisecond)
		loadingCounter := 0
		ready := false

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		percentage := 0.0

		done := make(chan struct{}) // Signal channel to notify workers to exit
		var wg sync.WaitGroup
		seenCodes := sync.Map{}

		resultChan := make(chan CodeResult, numWorkers) // Buffered channel for worker results

		// Create worker pool
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go worker(i, config.Prefix, config.Length, resultChan, done, &wg, &seenCodes)
		}

		// Write generated codes to file
		go func() {
			codesBuffer := bufio.NewWriter(file)

			// Generate the initial header and add it to the buffer
			_, _ = fmt.Fprintf(codesBuffer, "%s %s %v\n%s\n%s\n", humanize.Comma(int64(config.NumCodes)), config.FLineF, time.Now().Format("2006-01-02 15:04:05"), config.HelpF, strings.Repeat("-", 96))

			startTime := time.Now()
			for i := 1; i <= config.NumCodes; i++ {
				elapsedTime := time.Since(startTime)
				remainingIterations := config.NumCodes - i
				config.RemainingTime = utils.CalculateRemainingTime(elapsedTime, i, remainingIterations)

				codeResult := <-resultChan
				code := codeResult.Code

				if config.LineNumbers {
					_, _ = fmt.Fprintf(codesBuffer, "%d: %s\n", i, code)
				} else {
					_, _ = fmt.Fprintf(codesBuffer, "%s\n", code)
				}
				percentage = float64(i) / float64(config.NumCodes) * 100
			}

			close(done) // Signal workers to exit
			ready = true
			codesBuffer.Flush()

			if err != nil {
				c := color.New(color.BgRed, color.FgBlack).Sprint("Error!")
				fmt.Printf("Issue saving codes to file! %s%v\r\n", c, err)
				return
			}
		}()

		for {
			select {
			case <-loadingTicker.C:
				fmt.Print("\033[?25l")

				info := color.New(color.FgBlack, color.BgGreen).Sprintf("[ST: %.2f%% | RM: %s]", percentage, utils.FormatDuration(config.RemainingTime))
				fmt.Printf("\r%s Generating %s codes with prefix '%s' %s%10s", loadingAnimation(loadingCounter), humanize.Comma(int64(config.NumCodes)), config.Prefix, info, "")
				loadingCounter = (loadingCounter + 1) % 2

				if ready {
					clearLine()
					fmt.Print("\033[?25h")
					color.Yellow("Generated codes saved to %s\r\n", filePath)
					if runtime.GOOS == "windows" {
						c := color.New(color.BgGreen, color.FgBlack).Sprint("Press 'O' to open directory")
						fmt.Printf("%s or any key to exit\r\n", c)
						err := keyboard.Open()
						if err != nil {
							log.Fatal(err)
						}
						defer keyboard.Close()

						char, _, err := keyboard.GetKey()
						if err != nil {
							log.Fatal(err)
						}

						if char == 'o' || char == 'O' {
							openDirectoryInExplorer(filepath.Join(homePath, config.OSSpecDir, "Code Generator", "Files"))
							os.Exit(0)
						} else {
							os.Exit(0)
						}
					}
					return
				}
			case <-sigChan:
				fmt.Print("\033[?25h")
				if file != nil {
					file.Close()
				}
				files, _ := os.ReadDir(path)
				if len(files) == 1 {
					_ = os.RemoveAll(filepath.Join(homePath, config.OSSpecDir, "Code Generator"))
				} else {
					err := os.Remove(filePath)
					if err != nil {
						log.Println(err)
					}
				}
				clearLine()
				c := color.New(color.BgRed, color.FgBlack).Sprint("Interrupted!")
				fmt.Printf("%s Code generation interrupted. Cleaned up generated file.\r\n", c)
				os.Exit(0)
			}
		}
	}
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func generateCodeWithPool(prefix string, length int) string {
	builder := stringBuilderPool.Get().(*strings.Builder)
	defer stringBuilderPool.Put(builder)

	builder.Reset()
	builder.WriteString(prefix)
	builder.WriteString("-")

	for i := 0; i < length; i++ {
		builder.WriteByte(charset[rand.Intn(len(charset))])
	}

	return builder.String()
}

func validatePrefix(prefix string) bool {
	return len(prefix) >= minPrefixLength && len(prefix) <= maxPrefixLength
}

func validateLength(length int) bool {
	return length >= minCodeLength && length <= maxCodeLength
}

func GenerateHash(data string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(data), 14)
	return string(bytes), err
}

func sanitizeKey(input string) string {
	pattern := "[^a-zA-Z]+"
	re := regexp.MustCompile(pattern)
	sanitizedKey := re.ReplaceAllString(input, "")
	output := strings.ToLower(sanitizedKey)

	return output[4:30]
}

func calculatePossibleOutcomes(digits int) int {
	return int(math.Pow10(digits))
}

func worker(id int, prefix string, length int, resultChan chan<- CodeResult, done <-chan struct{}, wg *sync.WaitGroup, seenCodes *sync.Map) {
	defer wg.Done()

	for {
		select {
		case <-done:
			return // Exit the worker when done signal is received
		default:
			newCode := generateCodeWithPool(prefix, length)
			_, loaded := seenCodes.LoadOrStore(newCode, struct{}{})
			if !loaded {
				resultChan <- CodeResult{Code: newCode}
			}

		}
	}
}

func clearLine() {
	fmt.Print("\033[2K\r")
}

func openDirectoryInExplorer(directory string) {
	var cmd *exec.Cmd
	switch runtimeOS := runtime.GOOS; runtimeOS {
	case "windows":
		cmd = exec.Command("explorer", directory)
	default:
		fmt.Printf("Unsupported operating system: %s\r\n", runtimeOS)
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func loadingAnimation(counter int) string {
	animation := []string{"> ", " >"}
	return animation[counter]
}
