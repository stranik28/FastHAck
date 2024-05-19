package tmp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ccuetoh/nsfw/nsfw-main"
	"github.com/ccuetoh/nsfw/nsfw-main/predicator"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func extractFrame(videoURL string, second int, framesDir string, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	cmd := exec.Command("ffmpeg", "-y", "-ss", fmt.Sprintf("%d", second), "-i", videoURL, "-frames:v", "1", "-q:v", "2", "-f", "image2pipe", "-vcodec", "png", "pipe:1")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		errChan <- fmt.Errorf("error running ffmpeg for second %d: %w, stderr: %s", second, err, stderrBuf.String())
		return
	}

	img, err := png.Decode(&stdoutBuf)
	if err != nil {
		errChan <- fmt.Errorf("error decoding image for second %d: %w, stderr: %s", second, err, stderrBuf.String())
		return
	}
	outFile, err := os.Create(fmt.Sprintf("%s_SECOND_%d.png", framesDir, second))
	if err != nil {
		errChan <- fmt.Errorf("error creating output file: %w", err)
		return
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		errChan <- fmt.Errorf("error encoding png: %w", err)
		return
	}
}

func getFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, fmt.Errorf("error getting file size: %w", err)
	}
	defer resp.Body.Close()

	size := resp.Header.Get("Content-Length")
	if size == "" {
		return 0, fmt.Errorf("content length not found")
	}

	return strconv.ParseInt(size, 10, 64)
}
func getFrames(videoURL string, framesDir string) { // Директория для сохранения кадров

	fileSize, _ := getFileSize(videoURL)

	var wg sync.WaitGroup
	optiTime := fileSize / 65_000_00
	fmt.Println(optiTime)
	if optiTime < 10 {
		optiTime = fileSize / 65_000_0
	}
	var param int64
	if optiTime < 100 {
		param = optiTime / 10
	} else {
		param = optiTime / 100
	}
	errChan := make(chan error, param)
	for i := 1; i <= int(optiTime); i += int(param) {
		wg.Add(1)
		go extractFrame(videoURL, i, framesDir, &wg, errChan)
	}

	// Закрываем каналы после завершения всех горутин
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Проверка на наличие ошибок
	for err := range errChan {
		if err != nil {
			log.Println(err)
		}
	}
}

func predictAndSaveResults(framesDir string, predictor *nsfw.Predictor, resultsFile string) (error, string) {
	results, err := os.Create(resultsFile)
	if err != nil {
		return err, ""
	}
	defer results.Close()

	writer := bufio.NewWriter(results)

	files, err := os.ReadDir(framesDir)
	if err != nil {
		return err, ""
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" {
			framePath := filepath.Join(framesDir, file.Name())
			image := predictor.NewImage(framePath, 3)
			prediction := predictor.Predict(image)
			if prediction.Porn > 0.5 || prediction.Hentai > 0.5 || prediction.Sexy > 0.5 {
				fmt.Println(prediction.Describe())
				if prediction.Porn > 0.5 {
					return nil, "Порно"
				} else if prediction.Hentai > 0.5 {
					return nil, "Хентай"
				} else if prediction.Sexy > 0.5 {
					return nil, "Секусуальный контен"
				}
				return nil, ""
			}
			writer.WriteString(fmt.Sprintf("%s: %s\n", file.Name(), prediction))
		}
	}

	return writer.Flush(), ""
}

func Run(videoURL string) (error, string) {
	logrus.SetLevel(logrus.InfoLevel)

	//videoURL := "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
	framesDir := "framess/"
	resultsFile := "results.txt"
	dir := filepath.Dir(framesDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	predict := predicator.Predictor()
	start := time.Now()
	getFrames(videoURL, framesDir)
	end := time.Now()
	err, res := predictAndSaveResults(framesDir, predict, resultsFile)
	if err != nil {
		//logrus.Fatal("unable to predict and save results", err)
		return err, ""
	}
	end = time.Now()
	fmt.Printf("Processing completed. Time elapsed: %d Milliseconds\n", end.Sub(start).Milliseconds())
	fmt.Printf("Processing completed. Time elapsed: %f Seconds\n", end.Sub(start).Seconds())
	if res == "" {
		fmt.Println("Access granted")
		return nil, ""
	} else {
		fmt.Println("Access red")
		return nil, res
	}

}
