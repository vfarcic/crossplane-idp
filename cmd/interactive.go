package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle   = focusedStyle.Copy()
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	// focusedButton       = focusedStyle.Copy().Render("[ Submit ]")
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	noStyle             = lipgloss.NewStyle()
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive walkthrough",
	Long:  `Interactive walkthrough`,
	Run: func(cmd *cobra.Command, args []string) {
		m := initialModel()
		if err := tea.NewProgram(m).Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

const (
	listHeight        = 14
	defaultWidth      = 20
	titleTypes        = "Which type would you like to use?"
	titleCompositions = "Which Composition would you like to use?"
)

type item string

type model struct {
	xrdApis                list.Model
	quitting               bool
	selectedXrdText        string
	selectedXrdKind        string
	selectedXrdApi         string
	selectedXrdVersion     string
	selectedXrdName        string
	selectedComposition    string
	selectedType           string
	types                  list.Model
	compositions           list.Model
	screen                 string
	yamlFields             []textinput.Model
	cursorMode             textinput.CursorMode
	focusIndex             int
	yamlManifest           string
	yamlManifestWithFields string
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

func initialModel() model {
	return model{
		xrdApis:      getXrdApiList(),
		types:        getTypeList(),
		compositions: getCompositionList("", ""),
		screen:       "xrds",
		yamlFields:   []textinput.Model{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.xrdApis.SetWidth(msg.Width)
		m.types.SetWidth(msg.Width)
		m.compositions.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.screen {
			case "xrds":
				i, ok := m.xrdApis.SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.selectedXrdText = string(i)
				m.selectedXrdKind = strings.Split(m.selectedXrdText, " ")[0]
				fullApi := strings.Split(strings.Split(m.selectedXrdText, "(")[1], ")")[0]
				m.selectedXrdName = strings.Split(fullApi, ".")[0]
				apiWithoutName := strings.ReplaceAll(fullApi, m.selectedXrdName+".", "")
				m.selectedXrdApi = strings.Split(apiWithoutName, "/")[0]
				m.selectedXrdVersion = strings.Split(apiWithoutName, "/")[1]
				m.compositions = getCompositionList(m.selectedXrdKind, m.selectedXrdApi+"/"+m.selectedXrdVersion)
				m.screen = "compositions"
				m.compositions.Title = "RXD: " + m.selectedXrdText + "\n\n" + titleCompositions
			case "compositions":
				i, ok := m.compositions.SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.selectedComposition = string(i)
				m.screen = "types"
				m.types.Title = "RXD: " + m.selectedXrdText + "\n\n" + titleTypes
			case "types":
				i, ok := m.types.SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.selectedType = string(i)
				crd := getCRD(m.selectedXrdName + "." + m.selectedXrdApi)
				xr := getXR(crd, m.selectedComposition)
				yamlManifest := getXRYaml(xr)
				yamlDataWithFields, yamlFields := m.getYamlDataWithFields(yamlManifest)
				m.yamlManifestWithFields = yamlDataWithFields
				m.yamlFields = yamlFields
				m.screen = "explain"
			case "explain":
				i, ok := m.types.SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.selectedType = string(i)
				m.screen = "end"
			default:
				m.screen = "end"
			}
			return m, cmd
		case "tab", "shift+tab", "up", "down":
			if m.screen == "explain" {
				if keypress == "up" || keypress == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.yamlFields) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.yamlFields)
				}
				cmds := make([]tea.Cmd, len(m.yamlFields))
				for i := 0; i <= len(m.yamlFields)-1; i++ {
					if i == m.focusIndex {
						cmds[i] = m.yamlFields[i].Focus()
						m.yamlFields[i].PromptStyle = focusedStyle
						m.yamlFields[i].TextStyle = focusedStyle
						continue
					}
					m.yamlFields[i].Blur()
					m.yamlFields[i].PromptStyle = noStyle
					m.yamlFields[i].TextStyle = noStyle
				}
				return m, tea.Batch(cmds...)
			}
		default:
			if m.screen == "explain" {
				crd := getCRD(m.selectedXrdName + "." + m.selectedXrdApi)
				xr := getXR(crd, m.selectedComposition)
				yamlManifest := getXRYaml(xr)
				yamlDataWithFields, _ := m.getYamlDataWithFields(yamlManifest)
				m.yamlManifestWithFields = yamlDataWithFields
			}
		}
	}
	switch m.screen {
	case "compositions":
		m.compositions, cmd = m.compositions.Update(msg)
	case "types":
		m.types, cmd = m.types.Update(msg)
	case "explain":
		cmd = m.updateInputs(msg)
	default:
		m.xrdApis, cmd = m.xrdApis.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("See ya!")
	}
	switch m.screen {
	case "compositions":
		return m.compositions.View()
	case "types":
		return m.types.View()
	case "explain":
		out := fmt.Sprintf(`XRD: %s
Composition: %s
Type: %s

Sample YAML:

---

%s

---
`,
			m.selectedXrdText,
			m.selectedComposition,
			m.selectedType,
			m.yamlManifestWithFields,
		)
		var b strings.Builder
		b.WriteString("\nFields:\n\n")
		for i := range m.yamlFields {
			b.WriteString(m.yamlFields[i].View())
			if i < len(m.yamlFields)-1 {
				b.WriteRune('\n')
			}
		}
		out += b.String()
		return out
	case "end":
		return `
TODO:

* Edit fields
* Contain it within a window
* Fix the option to choose claims
* Export to YAML
* kubectl apply
* Push to Git
* See all running claims, compositions, and managed resources
* Add to krew
* Sleep
		
The End`
	}
	return m.xrdApis.View()
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str))
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func getXrdApiList() list.Model {
	items := []list.Item{}
	xrds := getXrds()
	for _, xrd := range xrds.Items {
		items = append(items, item(xrd.Spec.Names.Kind+" ("+xrd.Metadata.Name+"/"+xrd.Spec.Versions[0].Name+")"))
	}
	return getListModel(items, "Which Crossplane Resource Definition (XRD) would you like to use?")
}

func getTypeList() list.Model {
	items := []list.Item{item("Composition"), item("Claim (TODO: don't use it, it does not work!")}
	return getListModel(items, titleTypes)
}

func getCompositionList(expectedKind, expectedApi string) list.Model {
	items := []list.Item{}
	compositions := getAllCompositions()
	for _, composition := range compositions.Items {
		if expectedKind == composition.Spec.CompositeTypeRef.Kind &&
			expectedApi == composition.Spec.CompositeTypeRef.ApiVersion {
			items = append(items, item(composition.Metadata.Name))
		}
	}
	return getListModel(items, titleCompositions)
}

func getListModel(items []list.Item, title string) list.Model {
	list := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	list.Title = title
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false)
	list.Styles.Title = titleStyle
	list.Styles.PaginationStyle = paginationStyle
	list.Styles.HelpStyle = helpStyle
	return list
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.yamlFields))
	for i := range m.yamlFields {
		m.yamlFields[i], cmds[i] = m.yamlFields[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *model) getYamlDataWithFields(yaml string) (string, []textinput.Model) {
	// # TODO: Add default OpenAPI values to fields
	yamlWithFields := yaml
	insertCount := strings.Count(yaml, "INSERT_HERE")
	if len(m.yamlFields) == 0 {
		m.yamlFields = make([]textinput.Model, insertCount)
		for i := range m.yamlFields {
			ti := textinput.New()
			ti.CursorStyle = cursorStyle
			ti.CharLimit = 156
			// TODO: Replace with field names
			ti.Placeholder = "Field " + strconv.Itoa(i)
			ti.PromptStyle = focusedStyle
			ti.TextStyle = focusedStyle
			ti.Width = 20
			m.yamlFields[i] = ti
		}
		m.yamlFields[0].Focus()
	}
	for i := range m.yamlFields {
		yamlWithFields = strings.Replace(yamlWithFields, "INSERT_HERE", m.yamlFields[i].Value(), 1)
	}
	return yamlWithFields, m.yamlFields
}
