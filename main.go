package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
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

		inputSongInfoEntry := ui.NewEntry()
		if cliInputs.inputFolder != "" {
			inputSongInfoEntry.SetText(filepath.Join(cliInputs.inputFolder, "info.json"))
		}

		inputSongInfoButton := ui.NewButton("Browse")

		outputFolderEntry := ui.NewEntry()
		outputFolderEntry.SetText(cliInputs.outputFolder)
		canOverrideOutputFolder := true
		outputFolderEntry.OnChanged(func(e *ui.Entry) {
			canOverrideOutputFolder = e.Text() == ""
		})
		updateOutputFolder := func() {
			if !canOverrideOutputFolder {
				return
			}
			if inputSongInfoEntry.Text() == "" {
				outputFolderEntry.SetText("")
				return
			}
			outputFolderEntry.SetText(filepath.Join(filepath.Dir(filepath.Dir(inputSongInfoEntry.Text())), "OUTPUT_FOLDER_NAME"))
		}
		inputSongInfoButton.OnClicked(func(btn *ui.Button) {
			inputSongInfoEntry.SetText(ui.OpenFile(window))
			updateOutputFolder()
		})
		inputSongInfoEntry.OnChanged(func(e *ui.Entry) {
			updateOutputFolder()
		})

		inputBpmEntry := ui.NewEntry()
		if cliInputs.inputBPM != 0 {
			inputBpmEntry.SetText(floatToString(cliInputs.inputBPM))
		}
		loadInputBpmButton := ui.NewButton("load from song info")
		loadInputBpmButton.OnClicked(func(btn *ui.Button) {
			if err := validateSongInfo(inputSongInfoEntry.Text()); err != nil {
				ui.MsgBoxError(window, "error", err.Error())
				return
			}

			bpm, err := loadBpmFromFolder(filepath.Dir(inputSongInfoEntry.Text()))
			if err != nil {
				ui.MsgBoxError(window, "error", "couldn't load bpm from song info '"+inputSongInfoEntry.Text()+"': "+err.Error())
				return
			}
			inputBpmEntry.SetText(floatToString(bpm))
		})

		multiplyButton := ui.NewButton("=")
		numerator := ui.NewSpinbox(1, math.MaxInt32)
		denominator := ui.NewSpinbox(1, math.MaxInt32)

		outputBpmEntry := ui.NewEntry()
		if cliInputs.outputBPM != 0 {
			outputBpmEntry.SetText(floatToString(cliInputs.outputBPM))
		}
		loadOutputBpmButton := ui.NewButton("load from output folder")
		loadOutputBpmButton.OnClicked(func(btn *ui.Button) {
			validateOutputFolder(outputFolderEntry.Text())
			bpm, err := loadBpmFromFolder(outputFolderEntry.Text())
			if err != nil {
				ui.MsgBoxError(window, "error", "couldn't load bpm from song info in directory '"+outputFolderEntry.Text()+"': "+err.Error())
				return
			}
			outputBpmEntry.SetText(floatToString(bpm))
		})

		multiplyButton.OnClicked(func(btn *ui.Button) {
			inputBPM, err := parsePositiveFloat(inputBpmEntry.Text())
			if err != nil {
				ui.MsgBoxError(window, "Error", "invalid input BPM '"+inputBpmEntry.Text()+"'")
				return
			}
			outputBpmEntry.SetText(floatToString(inputBPM * float64(numerator.Value()) / float64(denominator.Value())))
		})

		button := ui.NewButton("Convert")

		box := ui.NewVerticalBox()
		box.SetPadded(true)
		box.Append(ui.NewLabel("All fields are required"), false)

		inputSongInfoBox := ui.NewHorizontalBox()
		inputSongInfoBox.SetPadded(true)
		inputSongInfoBox.Append(inputSongInfoButton, false)
		inputSongInfoBox.Append(inputSongInfoEntry, true)
		inputSongInfoGroup := ui.NewGroup("input song info.json")
		inputSongInfoGroup.SetChild(inputSongInfoBox)
		box.Append(inputSongInfoGroup, false)

		outputFolderBox := ui.NewHorizontalBox()
		outputFolderBox.SetPadded(true)
		outputFolderBox.Append(outputFolderEntry, true)
		outputFolderGroup := ui.NewGroup("output folder")
		outputFolderGroup.SetChild(outputFolderBox)
		box.Append(outputFolderGroup, false)

		bpmBox := ui.NewHorizontalBox()
		bpmBox.SetPadded(true)

		inputBpmGroup := ui.NewGroup("input bpm")
		inputBpmBox := ui.NewVerticalBox()
		inputBpmBox.Append(loadInputBpmButton, true)
		inputBpmBox.Append(inputBpmEntry, false)
		inputBpmBox.Append(ui.NewLabel(""), true)

		inputBpmGroup.SetChild(inputBpmBox)
		bpmBox.Append(inputBpmGroup, true)

		calcBox := ui.NewHorizontalBox()
		calcBox.SetPadded(true)

		crossBox := ui.NewVerticalBox()
		crossBox.Append(ui.NewLabel(""), true)
		crossBox.Append(ui.NewLabel("╳"), false)
		crossBox.Append(ui.NewLabel(""), true)
		calcBox.Append(crossBox, false)

		multiplyBox := ui.NewVerticalBox()
		multiplyBox.Append(numerator, false)
		slashBox := ui.NewVerticalBox()
		slashBox.Append(ui.NewLabel(""), true)
		slashBox.Append(ui.NewLabel("━━━━━"), false)
		slashBox.Append(ui.NewLabel(""), true)
		middleBox := ui.NewHorizontalBox()
		middleBox.Append(slashBox, true)
		middleBox.Append(multiplyButton, false)
		multiplyBox.Append(middleBox, false)
		multiplyBox.Append(denominator, true)
		calcBox.Append(multiplyBox, false)
		calcGroup := ui.NewGroup("mini calculator")
		calcGroup.SetChild(calcBox)
		bpmBox.Append(calcGroup, false)

		outputBpmGroup := ui.NewGroup("output bpm")
		outputBpmBox := ui.NewVerticalBox()
		outputBpmBox.Append(loadOutputBpmButton, true)
		outputBpmBox.Append(outputBpmEntry, false)
		outputBpmBox.Append(ui.NewLabel(""), true)
		outputBpmGroup.SetChild(outputBpmBox)
		bpmBox.Append(outputBpmGroup, true)

		box.Append(bpmBox, false)

		buttonsBox := ui.NewHorizontalBox()
		buttonsBox.SetPadded(true)
		buttonsBox.Append(button, true)
		box.Append(buttonsBox, true)

		window.SetMargined(true)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {
			inputs, err := validateInputs(inputSongInfoEntry.Text(), outputFolderEntry.Text(), inputBpmEntry.Text(), outputBpmEntry.Text())
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

func loadBpmFromFolder(songFolderPath string) (float64, error) {
	if err := ensureDir(songFolderPath); err != nil {
		return 0, err
	}
	info, err := loadSongInfo(songFolderPath)
	if err != nil {
		return 0, err
	}
	return info.BeatsPerMinute, nil
}

func validateSongInfo(inputSongInfo string) error {
	if err := ensureFile(inputSongInfo); err != nil {
		return fmt.Errorf("input song info '%s': %s", inputSongInfo, err)
	}
	if filepath.Base(inputSongInfo) != "info.json" {
		return fmt.Errorf("input song info '%s': file must be named info.json", inputSongInfo)
	}
	return nil
}

func validateOutputFolder(outputFolder string) error {
	if err := ensureDir(outputFolder); err != nil {
		return fmt.Errorf("output folder '%s': %s", outputFolder, err)
	}
	return nil
}

func validateInputs(inputSongInfo, outputFolder, inputBPM, outputBPM string) (*inputFields, error) {
	in := &inputFields{}
	if err := validateSongInfo(inputSongInfo); err != nil {
		return nil, err
	}
	in.inputFolder = filepath.Dir(inputSongInfo)

	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return nil, fmt.Errorf("couldn't create output folder '%s': %s", outputFolder, err)
	}
	if err := validateOutputFolder(outputFolder); err != nil {
		return nil, err
	}
	in.outputFolder = outputFolder

	var err error
	in.inputBPM, err = parsePositiveFloat(inputBPM)
	if err != nil {
		return nil, fmt.Errorf("input bpm: %s", err)
	}

	in.outputBPM, err = parsePositiveFloat(outputBPM)
	if err != nil {
		return nil, fmt.Errorf("output bpm: %s", err)
	}
	return in, nil
}

func floatToString(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func parsePositiveFloat(input string) (float64, error) {
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, errors.New("input bpm: must be > 0")
	}
	return val, nil
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

func ensureFile(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return errors.New("no such file or directory")
	}
	if fi.IsDir() {
		return errors.New("is a folder, not a file")
	}
	return nil
}

func process(inputs *inputFields) error {
	songInfo, err := loadSongInfo(inputs.inputFolder)
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

func loadSongInfo(folderPath string) (*SongInfo, error) {
	filePath := filepath.Join(folderPath, "info.json")
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
