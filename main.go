package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andlabs/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("ERROR:", err)
	}
}

// type inputFields struct {
// 	inputFolder  string
// 	outputFolder string
// 	inputBPM     float64
// 	outputBPM    float64
// }

func run() error {
	cliInputs := parseFlags()

	err := ui.Main(func() {
		window := ui.NewWindow("MeterSaber", 1200, 100, false)

		inputFolderEntry := ui.NewEntry()
		inputFolderEntry.SetText(cliInputs.inputFolder)
		inputFolderButton := ui.NewButton("Browse")
		inputFolderButton.OnClicked(func(btn *ui.Button) {
			inputFolderEntry.SetText(ui.OpenFile(window))
		})
		outputFolderEntry := ui.NewEntry()
		outputFolderEntry.SetText(cliInputs.outputFolder)
		outputFolderButton := ui.NewButton("Browse")
		outputFolderButton.OnClicked(func(btn *ui.Button) {
			outputFolderEntry.SetText(ui.OpenFile(window))
		})
		inputBpmEntry := ui.NewEntry()
		if cliInputs.inputBPM != 0 {
			inputBpmEntry.SetText(strconv.FormatFloat(cliInputs.inputBPM, 'f', -1, 64))
		}
		outputBpmEntry := ui.NewEntry()
		if cliInputs.outputBPM != 0 {
			outputBpmEntry.SetText(strconv.FormatFloat(cliInputs.outputBPM, 'f', -1, 64))
		}
		button := ui.NewButton("Convert")

		box := ui.NewVerticalBox()
		box.SetPadded(true)
		box.Append(ui.NewLabel("All fields are required"), true)

		inputFolderBox := ui.NewHorizontalBox()
		inputFolderBox.SetPadded(true)
		inputFolderBox.Append(inputFolderButton, false)
		inputFolderBox.Append(inputFolderEntry, true)
		inputFolderGroup := ui.NewGroup("input folder")
		inputFolderGroup.SetChild(inputFolderBox)
		box.Append(inputFolderGroup, false)

		outputFolderBox := ui.NewHorizontalBox()
		outputFolderBox.SetPadded(true)
		outputFolderBox.Append(outputFolderButton, false)
		outputFolderBox.Append(outputFolderEntry, true)
		outputFolderGroup := ui.NewGroup("output folder")
		outputFolderGroup.SetChild(outputFolderBox)
		box.Append(outputFolderGroup, false)

		bottomBox := ui.NewHorizontalBox()
		bottomBox.SetPadded(true)

		inputBpmGroup := ui.NewGroup("input bpm")
		inputBpmGroup.SetChild(inputBpmEntry)
		bottomBox.Append(inputBpmGroup, true)

		outputBpmGroup := ui.NewGroup("output bpm")
		outputBpmGroup.SetChild(outputBpmEntry)
		bottomBox.Append(outputBpmGroup, true)

		box.Append(bottomBox, false)

		buttonsBox := ui.NewHorizontalBox()
		buttonsBox.SetPadded(true)
		buttonsBox.Append(button, true)
		box.Append(buttonsBox, false)

		window.SetMargined(true)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {
			inputs, err := validateInputs(inputFolderEntry.Text(), outputFolderEntry.Text(), inputBpmEntry.Text(), outputBpmEntry.Text())
			if err != nil {
				ui.MsgBoxError(window, "invalid input", err.Error())
				return
			}
			if err := process(inputs); err != nil {
				ui.MsgBoxError(window, "processing error", err.Error())
				return
			}
			ui.MsgBox(window, "success", "new beatmaps are in "+inputs.outputFolder)
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err != nil {
		return err
	}
	return nil
	// return process(inputs)
}

func validateInputs(inputFolder, outputFolder, inputBPM, outputBPM string) (*inputFields, error) {
	in := &inputFields{}
	if err := ensureDir(inputFolder); err != nil {
		return nil, fmt.Errorf("input folder '%s': %s", inputFolder, err)
	}
	in.inputFolder = inputFolder

	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return nil, fmt.Errorf("couldn't create output folder '%s': %s", outputFolder, err)
	}
	if err := ensureDir(outputFolder); err != nil {
		return nil, fmt.Errorf("output folder '%s': %s", outputFolder, err)
	}
	in.outputFolder = outputFolder

	var err error
	in.inputBPM, err = strconv.ParseFloat(inputBPM, 64)
	if err != nil {
		return nil, fmt.Errorf("input bpm: %s", err)
	}
	if in.inputBPM <= 0 {
		return nil, errors.New("input bpm: must be > 0")
	}

	in.outputBPM, err = strconv.ParseFloat(outputBPM, 64)
	if err != nil {
		return nil, fmt.Errorf("output bpm: %s", err)
	}
	if in.outputBPM <= 0 {
		return nil, errors.New("output bpm: must be > 0")
	}
	return in, nil
}

func ensureDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return errors.New("no such file or directory")
	}
	if !fi.IsDir() {
		return errors.New("is a file, not a folder")
	}
	return nil
}

func process(inputs *inputFields) error {
	songInfo, err := loadSongInfo(filepath.Join(inputs.inputFolder, "info.json"))
	if err != nil {
		return err
	}

	for _, difficultyLevel := range songInfo.DifficultyLevels {
		beatMap, err := loadBeatmap(filepath.Join(inputs.inputFolder, difficultyLevel.JSONPath))
		if err != nil {
			return err
		}
		for i, note := range beatMap.Notes {
			beatMap.Notes[i].Time = convertTimeWithOffset(note.Time, inputs.inputBPM, inputs.outputBPM, difficultyLevel.Offset)
		}
		for i, obstacle := range beatMap.Obstacles {
			beatMap.Obstacles[i].Time = convertTimeWithOffset(obstacle.Time, inputs.inputBPM, inputs.outputBPM, difficultyLevel.Offset)
			beatMap.Obstacles[i].Duration = convertTime(obstacle.Duration, inputs.inputBPM, inputs.outputBPM)
		}
		beatMap.BeatsPerMinute = inputs.outputBPM
		if err := saveBeatmap(filepath.Join(inputs.outputFolder, difficultyLevel.JSONPath), beatMap); err != nil {
			return err
		}
	}
	return nil
}

func convertTimeWithOffset(oldTime, inputBPM, outputBPM float64, offset int) float64 {
	inputOffset := inputBPM * float64(offset) / 60000
	outputOffset := outputBPM * float64(offset) / 60000
	return convertTime(oldTime-inputOffset, inputBPM, outputBPM) + outputOffset
}

func convertTime(oldTime, inputBPM, outputBPM float64) float64 {
	return oldTime * outputBPM / inputBPM
}

func loadSongInfo(filePath string) (*SongInfo, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	songInfo := &SongInfo{}
	json.Unmarshal(raw, songInfo)
	return songInfo, nil
}

func loadBeatmap(filePath string) (*BeatMap, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	beatMap := &BeatMap{}
	json.Unmarshal(raw, beatMap)
	return beatMap, nil
}

func saveBeatmap(filePath string, beatMap *BeatMap) error {
	buffer, _ := json.Marshal(beatMap)
	return ioutil.WriteFile(filePath, buffer, 0644)
}

func parseFlags() *inputFields {
	in := inputFields{}
	flag.StringVar(&in.inputFolder, "inputFolder", "", "folder with existing BPM")
	flag.StringVar(&in.outputFolder, "outputFolder", "", "folder to save new BPM")
	flag.Float64Var(&in.inputBPM, "inputBPM", 0, "intended initial BPM")
	flag.Float64Var(&in.outputBPM, "outputBPM", 0, "intended new BPM")
	flag.Parse()
	return &in
}

type inputFields struct {
	inputFolder  string
	outputFolder string
	inputBPM     float64
	outputBPM    float64
}

type SongInfo struct {
	SongName         string            `json:"songName"`
	SongSubName      string            `json:"songSubName"`
	AuthorName       string            `json:"authorName"`
	BeatsPerMinute   float64           `json:"beatsPerMinute"`
	PreviewStartTime int               `json:"previewStartTime"`
	PreviewDuration  int               `json:"previewDuration"`
	CoverImagePath   string            `json:"coverImagePath"`
	EnvironmentName  string            `json:"environmentName"`
	DifficultyLevels []DifficultyLevel `json:"difficultyLevels"`
}

type DifficultyLevel struct {
	Difficulty     string `json:"difficulty"`
	DifficultyRank int    `json:"difficultyRank"`
	AudioPath      string `json:"audioPath"`
	JSONPath       string `json:"jsonPath"`
	Offset         int    `json:"offset"`
	OldOffset      int    `json:"oldOffset"`
}

type BeatMap struct {
	Version        string        `json:"_version"`
	BeatsPerMinute float64       `json:"_beatsPerMinute"`
	BeatsPerBar    int           `json:"_beatsPerBar"`
	NoteJumpSpeed  int           `json:"_noteJumpSpeed"`
	Shuffle        int           `json:"_shuffle"`
	ShufflePeriod  float64       `json:"_shufflePeriod"`
	Events         []interface{} `json:"_events"`
	Notes          []Note        `json:"_notes"`
	Obstacles      []Obstacle    `json:"_obstacles"`
}

type Note struct {
	Time         float64 `json:"_time"`
	LineIndex    int     `json:"_lineIndex"`
	LineLayer    int     `json:"_lineLayer"`
	Type         int     `json:"_type"`
	CutDirection int     `json:"_cutDirection"`
}

type Obstacle struct {
	Time      float64 `json:"_time"`
	LineIndex int     `json:"_lineIndex"`
	Type      int     `json:"_type"`
	Duration  float64 `json:"_duration"`
	Width     int     `json:"_width"`
}
