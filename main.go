package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/adrg/xdg"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var regionMap = map[string][]string{
	"Europe":        {"Albania", "Andorra", "Austria", "Belgium", "Bulgaria", "Croatia", "Czechia", "Cyprus", "Denmark", "England", "Estonia", "Faroe Island", "Finland", "France", "Germany", "Gibraltar", "Greece", "Greenland", "Hungary", "Iceland", "Ireland", "Italy", "Latvia", "Liechtenstein", "Lithuania", "Luxembourg", "Malta", "Montenegro", "Netherlands", "North Macedonia", "Norway", "Poland", "Portugal", "Romania", "Scotland", "Serbia", "Slovakia", "Slovenia", "Spain", "Sweden", "Switzerland", "Türkiye", "Ukraine", "Wales", "Åland"},
	"Asia":          {"Bangladesh", "Bhutan", "Cambodia", "Guam", "Hong Kong", "India", "India", "Indonesia", "Israel", "Japan", "Jordan", "Kazakhstan", "Kyrgyzstan", "Laos", "Macao", "Malaysia", "Mongolia", "Northern Mariana Islands", "Oman", "Philippines", "Qatar", "Russia", "Singapore", "South Korea", "Sri Lanka", "Taiwan", "Thailand", "United Arab Emirates", "West Bank"},
	"Africa":        {"Botswana", "Eswatini", "Ghana", "Kenya", "Lesotho", "Nigeria", "Rwanda", "Rèunion", "Senegal", "South Africa", "São Tomé and Príncipe", "Tunisia", "Uganda"},
	"North America": {"American Virgin Islands", "Bermuda", "British Virgin Islands", "Canada", "Dominican Republic", "Guatemala", "Hawaii", "Mexico", "Panama", "Puerto Rico", "United States of America"},
	"South America": {"Argentina", "Bolivia", "Brasil", "Chile", "Colombia", "Curaçao", "Ecuador", "Peru", "Uruguay"},
	"Oceania":       {"American Samoa", "Australia", "Christmas Islands", "New Zealand"},
}

var window fyne.Window

func main() {
	database, err := getOrCreateDatabase()

	geotracker := app.New()
	geotracker.Settings().SetTheme(NewTokyoNightStormTheme())

	window = geotracker.NewWindow("Geotracker")
	window.Resize(fyne.NewSize(600.0, 600.0))

	if err != nil {
		dialog.NewError(errors.New("Database couldn't be created"), window)
		return
	}

	entries, _ = fromJson(database)

	databaseFile, err := os.OpenFile(database, os.O_RDWR, 0644)
	if err != nil {
		dialog.NewError(errors.New("Database couldn't be opened with RDWR"), window)
		return
	}
	defer databaseFile.Close()
	defer writeDatabase(databaseFile)

	for k, _ := range regionMap {
		continents = append(continents, k)
	}
	sort.Strings(continents)

	tabs := container.NewAppTabs()
	tabs.Append(createNewEntryTab())
	tabs.Append(createNewResultsTab())
	tabs.Append(createNewInfoTab(database))

	tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Results" {
			ti.Content = createNewResultsTab().Content
			ti.Content.Refresh()
		}
	}

	window.SetContent(tabs)
	window.ShowAndRun()
}

func writeDatabase(databaseFile *os.File) {
	serialized, err := entries.toJson()
	if err != nil {
		return
	}
	databaseFile.WriteString(serialized)
}

type Entry struct {
	Region  string `json:"name"`
	Correct bool   `json:"correct"`
	Delta   uint   `json:"delta"`
}
type EntryList []Entry

func getOrCreateDatabase() (string, error) {
	dbDir := filepath.Join(xdg.ConfigHome, "geotracker")
	dbFile := filepath.Join(dbDir, "database.json")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return "", fmt.Errorf("failed to create database: %w", dbFile)
		}
		file.Close()
	}

	return dbFile, nil
}

func (el EntryList) toJson() (string, error) {
	bytes, err := json.Marshal(el)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func fromJson(path string) (el EntryList, _ error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&el); err != nil {
		return nil, fmt.Errorf("failed to decode database: %w", err)
	}

	return
}

var entries EntryList
var continents []string

type Results map[string]uint

func clamp(val int, min int, max int) uint {
	if min < 0 {
		min = 0
	}
	if val >= min && val <= max {
		return uint(val)
	} else if val < min {
		return uint(min)
	} else if val > max {
		return uint(max)
	}

	return uint(val)
}

func createResultData() Results {
	r := make(Results)
	for _, entry := range entries {
		region := entry.Region
		correct_bonus := 0
		if entry.Correct {
			correct_bonus = 100
		}
		value := clamp(5000-(5000-int(entry.Delta))+correct_bonus, 0, 5000)
		if r[region] != 0 {
			r[region] = (r[region] + value) / 2
		} else {
			r[region] = value
		}
	}

	return r
}

func createResultsTable() *container.Scroll {
	type Score struct {
		Region string
		Score  uint
	}
	var sorted []Score

	for region, score := range createResultData() {
		sorted = append(sorted, Score{Region: region, Score: score})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score < sorted[j].Score
	})

	var data [][]string
	for _, entry := range sorted {
		data = append(data, []string{entry.Region, strconv.Itoa(int(entry.Score))})
	}

	table := widget.NewTable(
		func() (int, int) {
			return len(data), 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(data[id.Row][id.Col])
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 200)
	dataTable := container.NewVScroll(table)
	dataTable.SetMinSize(fyne.NewSize(400, float32(50*10))) // be able to show at least 10 entries

	return dataTable

}

func createNewResultsTab() *container.TabItem {
	headerData := []string{"Region", "Approximate Score"}
	header := widget.NewTable(
		func() (int, int) {
			return 1, 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{
				Bold:      true,
				Italic:    false,
				Monospace: false,
				Symbol:    false,
				TabWidth:  0,
				Underline: true,
			})
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(headerData[id.Col])
		},
	)
	header.SetColumnWidth(0, 200)
	header.SetColumnWidth(1, 200)
	header.Resize(fyne.NewSize(400, 50))

	vsplit := container.New(layout.NewVBoxLayout(), header, createResultsTable())

	return container.NewTabItem("Results", vsplit)
}

func createNewEntryTab() *container.TabItem {
	selectedContinent := ""
	selectedRegion := ""

	pointsEntry := widget.NewEntry()
	pointsEntry.SetPlaceHolder("Enter points...")
	pointsEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		clean := strings.TrimSpace(s)
		num, err := strconv.ParseUint(clean, 10, 16)
		if err != nil {
			return errors.New("must be a number")
		}
		if num > 5000 {
			return errors.New("must be between 0-5000")
		}
		return nil
	}
	correctCheck := widget.NewCheck("Correct Region", nil)

	regionList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {},
	)
	regionList.OnSelected = func(id widget.ListItemID) {
		if selectedContinent != "" {
			selectedRegion = regionMap[selectedContinent][id]
		}
	}

	continentList := widget.NewList(
		func() int { return len(continents) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(continents[i])
		},
	)
	continentList.OnSelected = func(id widget.ListItemID) {
		selectedContinent = continents[id]
		updateRegionList(selectedContinent, regionList)
	}

	continentListContainer := container.NewStack(continentList)

	regionListContainer := container.NewVScroll(regionList)
	regionListContainer.SetMinSize(fyne.NewSize(200, 250))

	defaultSelectionPanel := container.NewGridWithColumns(2,
		continentListContainer,
		regionListContainer,
	)

	selectionContainer := container.NewStack(defaultSelectionPanel)

	searchField := widget.NewEntry()
	searchField.SetPlaceHolder("Search regions...")
	searchField.OnChanged = func(query string) {
		trimmed := strings.TrimSpace(query)
		if len(trimmed) < 2 {
			selectionContainer.Objects = []fyne.CanvasObject{defaultSelectionPanel}
			selectionContainer.Refresh()
			return
		} else {
			var matches []string
			for continent, regions := range regionMap {
				found := fuzzy.FindNormalizedFold(trimmed, regions)
				for _, match := range found {
					matches = append(matches, fmt.Sprintf("%s: %s", continent, match))
				}
			}
			searchList := widget.NewList(
				func() int { return len(matches) },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(id widget.ListItemID, obj fyne.CanvasObject) {
					obj.(*widget.Label).SetText(matches[id])
				},
			)
			searchList.OnSelected = func(id widget.ListItemID) {
				for continent, regions := range regionMap {
					for _, region := range regions {
						if strings.Contains(matches[id], region) {
							selectedContinent = continent
							selectedRegion = region
							return
						}
					}
				}
				dialog.ShowError(errors.New("Some unexpected result in the search happened. Please open a bug report."), window)
			}

			searchListScroll := container.NewVScroll(searchList)
			searchListScroll.SetMinSize(fyne.NewSize(400, float32(len(matches)-1)*50))

			selectionContainer.Objects = []fyne.CanvasObject{searchListScroll}
			selectionContainer.Refresh()
		}
	}

	addButton := widget.NewButton("Add Entry", func() {
		value, err := strconv.ParseUint(strings.TrimSpace(pointsEntry.Text), 10, 16)
		if err != nil {
			pointsEntry.SetText("Invalid input")
			return
		}
		if selectedContinent == "" || selectedRegion == "" {
			dialog.NewError(errors.New("Select a region"), window)
			return
		}
		entry := Entry{
			Region:  selectedRegion,
			Correct: correctCheck.Checked,
			Delta:   uint(value),
		}
		entries = append(entries, entry)
		pointsEntry.SetText("")
		correctCheck.SetChecked(false)
		searchField.SetText("")
		searchField.SetPlaceHolder("Search regions...")
		regionList.UnselectAll()
	})

	content := container.NewVBox(
		widget.NewLabel("New Game Entry"),
		pointsEntry,
		correctCheck,
		searchField,
		selectionContainer,
		addButton,
	)

	return container.NewTabItem("New Entry", content)
}

func createNewInfoTab(path string) *container.TabItem {
	pathLabel := widget.NewLabel("Database Path:")
	pathValue := widget.NewLabel(path)
	pathValue.Wrapping = fyne.TextWrapBreak

	openButton := widget.NewButton("Open", func() {
		err := openDir(path)
		if err != nil {
			dialog.NewError(err, window)
		}
	})

	copyright := widget.NewLabel("Copyright (c) 2025 SoraTenshi")

	content := container.NewVBox(
		pathLabel,
		pathValue,
		openButton,
		widget.NewSeparator(),
		copyright,
	)

	return container.NewTabItem("Info", content)
}

func openDir(path string) error {
	folderPath := filepath.Dir(path)

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", folderPath).Start()
	case "darwin":
		return exec.Command("open", folderPath).Start()
	case "linux":
		return exec.Command("xdg-open", folderPath).Start()
	default:
		return fmt.Errorf("unsupported os")
	}
}

func updateRegionList(continent string, regionList *widget.List) {
	if countries, ok := regionMap[continent]; ok {
		regionList.Length = func() int { return len(countries) }
		regionList.UpdateItem = func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(countries[i])
		}
		regionList.Refresh()
	}
}
