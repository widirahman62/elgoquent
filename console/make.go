package console

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/spf13/cobra"
	"github.com/widirahman62/elgoquent/console/stubs"
	"golang.org/x/mod/modfile"
)


type teaModelMakeCMD struct{
	teaModelCMD
	args []string
}

var objectType = map[string]func(...string) string {
	"odm": new(stubs.ModelStubs).ODM,
} 

var modelCmd = &cobra.Command{
	Use:   "make:model [model name] [object type]",
	Example: "elgoquent make:model user",
	Short: "Create a new Elgoquent model file in current directory",
	DisableFlagsInUseLine: true,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !regexp.MustCompile(`^([a-z0-9]+|/)$`).MatchString(args[0]) || strings.Contains(args[0], "'") {
			fmt.Println("error: model name must be lowercase without spaces or special characters")
			os.Exit(1)
		}
		workdir, err := os.Getwd()
		if err != nil {
			fmt.Println("error getting working directory:", err)
			os.Exit(1)
		}
		content, err := os.ReadFile("go.mod")
		if err != nil {
        	fmt.Println("error reading go.mod:", err)
			os.Exit(1)
    	} 
		if _, ok := objectType[args[1]]; !ok {
			fmt.Println("object type not found")
			os.Exit(1)
		} 
		_,err = tea.NewProgram(initTeaModelMakeModelCMD(&workdir,modfile.ModulePath(content), makeModel,&args[0],&args[1]) ,tea.WithAltScreen(),tea.WithMouseCellMotion()).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Model created successfully")
	},
}

func initTeaModelMakeModelCMD(workdir *string,moduleName string, makeCMD int, modelName, objectType *string) *teaModelMakeCMD {
	model := new(teaModelMakeCMD)
	model.workdir, model.moduleName, model.inputActiveAt,model.args = workdir, &moduleName, makeModel, append([]string{*modelName},*objectType)
	callBuilderCMD := new(teaModelBuilderCMD)
	callBuilderCMD.workdir = workdir
	inputs := make(map[int]*textinput.Model)
    inputs[makeModel] = new(textinput.Model)
	(*inputs[makeModel]) = textinput.New()
	inputs[makeModel].Prompt = ""
	inputs[makeModel].Validate = callBuilderCMD.validateDirPath
	model.textInput = inputs
	return model
}

func (m *teaModelMakeCMD) Init() tea.Cmd {
	return textinput.Blink
}

func (m *teaModelMakeCMD) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if teaMsg, ok := msg.(teaErr); ok {
        m.err = teaMsg
        return m, nil
    }
	switch m.inputActiveAt {
	case makeModel:
		m.cmdPage = 1
	}
	var output string
	switch m.cmdPage {
        case 1:
			output = fmt.Sprintf("Your elgoquent model will be created in the following location\nmodel name :%s\npath : %s/%s",m.args[0],*m.workdir,m.textInput[makeModel].View())
			if m.textInput[makeModel].Err!= nil {
                output += "\n\n" + fmt.Sprint(m.textInput[makeModel].Err)
            }
			if m.err != nil {
				m.viewPort.SetContent(wordwrap.String(m.err.Error(),m.widthScreen))
			} else {
				m.viewPort.SetContent(wordwrap.String(output,m.widthScreen))
			}
	}

	if teaMsg, ok := msg.(tea.KeyMsg); (ok && teaMsg.Type == tea.KeyCtrlC )|| (ok && m.cmdPage == 0 ){ return m, tea.Quit} 
	if teaMsg, ok := msg.(tea.WindowSizeMsg); ok {
        m.widthScreen = teaMsg.Width
        m.viewPort = viewport.New(m.widthScreen,teaMsg.Height)
        m.viewPort.MouseWheelDelta = 1
    }
	if teaMsg, ok := msg.(tea.MouseMsg); ok {
        switch teaMsg.Button {
        case tea.MouseButtonWheelUp:
            m.viewPort.LineUp(1)
        case tea.MouseButtonWheelDown:
            m.viewPort.LineDown(1)
        }
    }
	if teaMsg, ok := msg.(tea.KeyMsg); ok {
        switch teaMsg.Type {
        case tea.KeyUp:
            m.viewPort.LineUp(1)
        case tea.KeyDown:
            m.viewPort.LineDown(1)
        }
    }
	
	if teaMsg, ok := msg.(tea.KeyMsg); ok && teaMsg.Type == tea.KeyEnter {
		switch {
		case m.err != nil:
			m.err = nil
			return m, nil
        case m.cmdPage == 1:
            return m, m.makeModelFile()
        }
	}
	indexCmd, cmds := 0,make([]tea.Cmd, len(m.textInput))
	for i, input := range m.textInput {
        input.Blur()
        if i == m.inputActiveAt {
            input.Focus()
        }
		(*m.textInput[i]), cmds[indexCmd] = (*m.textInput[i]).Update(msg)
        indexCmd++
	}
    var cmd tea.Cmd
    m.viewPort, cmd = m.viewPort.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *teaModelMakeCMD) View() string {
	switch m.cmdPage {
        case 1:
            m.inputActiveAt=makeModel
    }
    return m.viewPort.View()
}

func (m *teaModelMakeCMD) makeModelFile() tea.Cmd {
	if _, err := os.Stat(*m.workdir+"/"+m.textInput[makeModel].Value()); !os.IsNotExist(err) {
		 return func() tea.Msg {return teaErr(fmt.Errorf("directory already exists\n(press ENTER to try again...)"))}
	}
	if er := os.MkdirAll(path.Clean(*m.workdir+"/"+m.textInput[makeModel].Value()), 0666); er != nil {
		return func() tea.Msg {return teaErr(fmt.Errorf("directory failed to created\n(press ENTER to try again...)"))}
    }
	readOrCreate, er :=  os.OpenFile(path.Clean(*m.workdir+"/"+m.textInput[makeModel].Value())+"/"+m.args[0]+".go", os.O_RDWR|os.O_CREATE, 0666)
	if er != nil {
		return func() tea.Msg {return teaErr(fmt.Errorf("model file failed to created\n(press ENTER to try again...)"))}
	}
	defer readOrCreate.Close()
	er = os.WriteFile(path.Clean(*m.workdir+"/"+m.textInput[makeModel].Value())+"/"+m.args[0]+".go", []byte(objectType[m.args[1]](m.args[0],packageName)), 0666)
	if er != nil {
		return func() tea.Msg {return teaErr(fmt.Errorf("model file content failed to created\n(press ENTER to try again...)"))}
	}
	return tea.Quit
}

func init() {
	rootCmd.AddCommand(modelCmd)
}
