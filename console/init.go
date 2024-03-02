package console

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/spf13/cobra"
	"github.com/widirahman62/elgoquent/console/stubs"
	"golang.org/x/mod/modfile"
)

type templateCaller struct{
    param []string
    call func(...string)string
}

type optionList struct{
    types string
    key *string
    totalItem *int
    dependOn *optionList
}

type teaModelCMD struct {
    textInput   map[int]*textinput.Model
    viewPort viewport.Model
    workdir,moduleName *string
    inputActiveAt int
    cmdPage int
    widthScreen int
	err      error
}

type teaModelBuilderCMD struct {
    teaModelCMD
    dirMap   *map[string]*string
    buildMap *map[*string]map[string]templateCaller
    optList []optionList
    optionNumber []string
    conflictIndex map[int]bool
    buildKeys []*string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Elgoquent in current project",
	Run: initApp,
}

func initApp(cmd *cobra.Command, args []string) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
        fmt.Println("error reading go.mod:", err)
    } else {
		insertCode(modfile.ModulePath(content))
	}
}

func insertCode(moduleName string) {
    workdir, err := os.Getwd()
    if err != nil {
        fmt.Println("error getting working directory:", err)
        return
    }
    generateToPath(&workdir,&moduleName)
}

func init() {
	rootCmd.AddCommand(initCmd) 
}

func generateToPath(workdir,moduleName *string){
    dirPath := map[string]*string{
        "config": stringPointer("config"),
        "register": stringPointer("register/"+appname),
    }
    buildContent := map[*string]map[string]templateCaller{
        dirPath["config"] : {
            "database.go":templateCaller{
                param: []string{"config", packageName}, 
                call: new(stubs.BuilderStubs).ConfigDatabase,
            },
        },
        dirPath["register"]: {
            appname+".go":templateCaller{
                param: []string{appname, packageName, *moduleName+"/"+*dirPath["config"]},
                call: new(stubs.BuilderStubs).Register,
            },
        },
    }
     _,err := tea.NewProgram(initTeaModelBuilderCMD(workdir,moduleName,&dirPath,&buildContent),tea.WithAltScreen(),tea.WithMouseCellMotion()).Run()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func initTeaModelBuilderCMD(workdir,moduleName *string,dirMap *map[string]*string,buildMap *map[*string]map[string]templateCaller) *teaModelBuilderCMD {
    model := &teaModelBuilderCMD{
        dirMap: dirMap,
        buildMap: buildMap,
    }
    model.workdir, model.moduleName, model.cmdPage, model.inputActiveAt, model.buildKeys = workdir, moduleName, 1, builderInit, make([]*string, 0)
    sortBuildKeys(model.buildMap,&model.buildKeys)
    generateBuildPath(model.workdir,&model.optList,model.dirMap,&model.buildKeys,model.buildMap,&model.conflictIndex)
    model.optionNumber = generateOptionNumber(len(model.optList))
    inputs := make(map[int]*textinput.Model)
    inputs[builderInit] = new(textinput.Model)
    (*inputs[builderInit]) = textinput.New()
    inputs[builderInit].Prompt = ""
    inputs[builderInit].Validate = func(input string) error {
        if len(input) <= 0 {
            return fmt.Errorf("input required. please enter a choice")
        }
        switch input{
            case "Y","y","N","n","M","m":
                return nil
        }
        if containStringOR(input, model.optionNumber...) || (len(model.conflictIndex) > 0 && containStringOR(input,"S","s")){
            return nil
        }
        return fmt.Errorf("invalid input :'%v'. please try again", input)
    }
    inputs[builderChangeFilePath] = new(textinput.Model)
    (*inputs[builderChangeFilePath]) = textinput.New()
    inputs[builderChangeFilePath].Prompt = ""
    inputs[builderChangeFilePath].Validate = model.validateFilePath
    
    inputs[builderChangeDirPath] = new(textinput.Model)
    (*inputs[builderChangeDirPath]) = textinput.New()
    inputs[builderChangeDirPath].Prompt = ""
    inputs[builderChangeDirPath].Validate = model.validateDirPath

    inputs[builderMoveWorkdir] =  new(textinput.Model)
    (*inputs[builderMoveWorkdir]) = textinput.New()
    inputs[builderMoveWorkdir].Prompt = ""
    inputs[builderMoveWorkdir].Validate = model.validateDirPath
    model.textInput = inputs
    return model
}

func (m *teaModelBuilderCMD) Init() tea.Cmd {
	return textinput.Blink
}

func (m *teaModelBuilderCMD) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var output string
    switch m.cmdPage {
        case 0:
            m.viewPort.SetContent(wordwrap.String("All files and directories built successfully!\n\nPress any key to exit",m.widthScreen))
        case 1:
            output = m.promptGetInitLabel() + m.textInput[builderInit].View()
            if m.textInput[builderInit].Err != nil {
                output += "\n\n" + fmt.Sprint(m.textInput[builderInit].Err)
            }
            m.viewPort.SetContent(wordwrap.String(output,m.widthScreen))
        case 2:
            optIndex,_ := strconv.Atoi(m.textInput[builderInit].Value())
            output = "Current file path\t: "+ *m.workdir + "/" + *(*m.dirMap)[*m.optList[optIndex].dependOn.key] + "/" + *m.optList[optIndex].key + "\nInput new file name\t: "+*m.workdir + "/" + *(*m.dirMap)[*m.optList[optIndex].dependOn.key] +"/"+ m.textInput[builderChangeFilePath].View()
            if m.textInput[builderChangeFilePath].Err != nil {
                output += "\n\n" + fmt.Sprint(m.textInput[builderChangeFilePath].Err)
            }
            m.viewPort.SetContent(wordwrap.String(output,m.widthScreen))
        case 3:
            optIndex,_ := strconv.Atoi(m.textInput[builderInit].Value())
            output = "Current directory path\t: "+ *m.workdir + "/" + *(*m.dirMap)[*m.optList[optIndex].key] + "\nInput new location\t: "+*m.workdir+"/"+ m.textInput[builderChangeDirPath].View()
            if m.textInput[builderChangeDirPath].Err != nil {
                output += "\n\n" + fmt.Sprint(m.textInput[builderChangeDirPath].Err)
            }
            m.viewPort.SetContent(wordwrap.String(output,m.widthScreen))
        case 4:
            output = "Move all generated files and directories to another location inside the current project path.\n"
            output += "Input new location path : " + *m.workdir + "/" +  m.textInput[builderMoveWorkdir].View()
            if m.textInput[builderMoveWorkdir].Err != nil {
                output += "\n" + fmt.Sprint(m.textInput[builderMoveWorkdir].Err)
            }
            m.viewPort.SetContent(wordwrap.String(output,m.widthScreen))
    }
    if m.err != nil {
        m.viewPort.SetContent(wordwrap.String(fmt.Sprintf("prompt failed: %s\n\n(press ENTER to continue...)", m.err),m.widthScreen))
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
    if teaMsg, ok := msg.(teaErr); ok {
        m.err = teaMsg
        return m, nil
    }
    if teaMsg, ok := msg.(tea.KeyMsg); ok && teaMsg.Type == tea.KeyEnter && m.cmdPage == 1 {
        userInput := m.textInput[builderInit].Value()
            switch userInput {
            case "Y", "y":
                m.cmdPage = 0
                return m, m.buildFileAndDirectory()
            case "N", "n":
                return m, tea.Quit
            case "M", "m":
                m.cmdPage = 4
                m.textInput[builderMoveWorkdir].Focus()
                return m, nil
            }
            switch indexPath,_ := strconv.Atoi(userInput);{
            case len(m.conflictIndex) > 0 && containStringOR(userInput, "S", "s"):
                m.skipConflict()
            case containStringOR(userInput, m.optionNumber...) && m.optList[indexPath].types == "file":
                m.cmdPage = 2
                m.textInput[builderChangeFilePath].Focus()
            case containStringOR(userInput, m.optionNumber...) && m.optList[indexPath].types == "dir":
                m.cmdPage = 3
                m.textInput[builderChangeDirPath].Focus()
            }
            return m,nil
    }
    if teaMsg, ok := msg.(tea.KeyMsg); ok && teaMsg.Type == tea.KeyEnter && m.cmdPage != 1{
        switch m.cmdPage {
        case 2:
            return m, m.changePath(builderChangeFilePath)
        case 3:
            return m, m.changePath(builderChangeDirPath)
        case 4:
            return m, m.changePath(builderMoveWorkdir)
        }
        m.cmdPage = 1
        m.err = nil
        m.textInput[builderInit].Reset()
        m.textInput[builderInit].Focus()
        return m, nil
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

func (m *teaModelBuilderCMD) View() (output string) {
    if m.err != nil {
        m.inputActiveAt=-1
		return m.viewPort.View()
	}
    switch m.cmdPage {
        case 1:
            m.inputActiveAt=builderInit
        case 2:
            m.inputActiveAt=builderChangeFilePath
        case 3:
            m.inputActiveAt=builderChangeDirPath
        case 4:
            m.inputActiveAt = builderMoveWorkdir
    }
    return m.viewPort.View()
}

func (m *teaModelBuilderCMD) buildFileAndDirectory() tea.Cmd {
    openOrCreateFile, err := os.OpenFile("main.go", os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    defer openOrCreateFile.Close()
    var maindotgo []byte
    maindotgo, err = os.ReadFile(openOrCreateFile.Name())
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    lines := strings.Split(string(maindotgo), "\n")
    if !containStringOR(string(maindotgo), "package") { 
        lines = append(lines[:1], lines[0:]...)
        lines[0] = "package main"
        maindotgo = []byte(strings.Join(lines, "\n"))
    }
    for _ , list := range m.optList {
        switch {
        case list.types == "dir":
            if !containStringOR(string(maindotgo), `"`+ *m.moduleName + "/" + *(*m.dirMap)[*list.key] + `"`) && isAppNameFile((*m.buildMap)[(*m.dirMap)[*list.key]]) {
                lines = append(lines[:1], append([]string{"","import " + appname + ` "`+ *m.moduleName + "/" + *(*m.dirMap)[*list.key] + `"`,""}, lines[2:]...)...)
                maindotgo = []byte(strings.Join(lines, "\n"))
            }
            if er := os.MkdirAll(*m.workdir + "/" + *(*m.dirMap)[*list.key], 0666); er != nil {
                m.cmdPage = -1
                m.textInput[builderInit].Reset()
                return func() tea.Msg {return teaErr(er)}
            }
        case list.types == "file" && len((*m.buildMap)[(*m.dirMap)[*list.dependOn.key]][*list.key].param) != 0 && (*m.buildMap)[(*m.dirMap)[*list.dependOn.key]][*list.key].call != nil:
            fileItem:= (*m.buildMap)[(*m.dirMap)[*list.dependOn.key]][*list.key]
            readOrCreate, er :=  os.OpenFile(*m.workdir + "/" + *(*m.dirMap)[*list.dependOn.key]+ "/" + *list.key, os.O_RDWR|os.O_CREATE, 0666)
            defer readOrCreate.Close()
            if er != nil {
                m.cmdPage = -1
                m.textInput[builderInit].Reset()
                return func() tea.Msg {return teaErr(er)}
            }
            er = os.WriteFile(*m.workdir + "/" + *(*m.dirMap)[*list.dependOn.key]+ "/" + *list.key, []byte(fileItem.call(fileItem.param...)), 0666)
            if er != nil {
                m.cmdPage = -1
                m.textInput[builderInit].Reset()
                return func() tea.Msg {return teaErr(er)}
            }
        }
    }
    for lineIndex := range lines {
        switch {
        case containStringAND(lines[lineIndex], "func main()", "{") || containStringOR(lines[lineIndex], "func main(){"):
            if !containStringOR(string(maindotgo),"elgoquent.Register()") {
                lines = append(lines[:lineIndex+2], lines[lineIndex+1:]...)
                lines[lineIndex+1] = "\telgoquent.Register()"
            }
        case containStringOR(lines[lineIndex], "func main(){}", "func main()"):
            lines[lineIndex] = "func main(){"
            lines = append(lines[:lineIndex+1], lines[lineIndex+1:]...)
            lines = append(lines[:lineIndex+1], append([]string{"\telgoquent.Register()","}"}, lines[lineIndex+2:]...)...)
        case !(containStringAND(lines[lineIndex], "func main()", "{") || containStringOR(lines[lineIndex], "func main(){")) && !containStringOR(string(maindotgo),"elgoquent.Register()"):
            if lineIndex==len(lines)-1 {
                lines = append(lines[:lineIndex+1], lines[lineIndex:]...)
                lines = append(lines[:lineIndex+1], append([]string{"func main() {","\telgoquent.Register()","}"}, lines[lineIndex+2:]...)...)
            }
        }
    }
    maindotgo = []byte(strings.Join(lines, "\n"))
    err = os.WriteFile(openOrCreateFile.Name(), maindotgo, 0666)
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    openOrCreateFile,err = os.OpenFile(".env", os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    defer openOrCreateFile.Close()
    var envFile []byte
    envFile, err = os.ReadFile(openOrCreateFile.Name())
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    var envKey []string
    for k := range defaultEnv {
        envKey = append(envKey, k)
    }
    sort.Slice(envKey,func(i, j int) bool {
        return envKey[i] < envKey[j]
    })
    for _, key := range envKey { 
        if !strings.Contains(string(envFile), key+"=") && !containStringOR(string(envFile), key + " =") {
            envFileLine := strings.Split(string(envFile), "\n")
            envFileLine = append(envFileLine, key+"="+defaultEnv[key])
            envFile = []byte(strings.Join(envFileLine, "\n"))
        }
    }
    err = os.WriteFile(openOrCreateFile.Name(), envFile, 0666)
    if err != nil {
        m.cmdPage = -1
        m.textInput[builderInit].Reset()
        return func() tea.Msg {return teaErr(err)}
    }
    return nil
}

func (m *teaModelBuilderCMD) skipConflict(){
    for index := range m.optList {
        if m.conflictIndex[index] {
            *m.optList[index].dependOn.totalItem -= 1
        }
    }
    var currentIndex int
    for i,v := range m.optList {
        if (!m.conflictIndex[i] && v.totalItem == nil) || (v.totalItem != nil && *v.totalItem > 0) {
            m.optList[currentIndex] = v
            currentIndex++
        } else {
            switch v.types {
            case "file" :
                delete((*m.buildMap)[v.dependOn.key], *v.key)
            default :
                delete(*m.buildMap, (*m.dirMap)[*v.key])
                delete(*m.dirMap, *v.key)
            }
            delete(m.conflictIndex, i)
        }
    }
    m.optList = m.optList[:currentIndex]
    m.optionNumber = generateOptionNumber(len(m.optList))
    m.cmdPage = 1
    m.textInput[builderInit].Reset()
    m.textInput[builderInit].Focus()
}

func (m *teaModelBuilderCMD) changePath(builderIndex int) tea.Cmd {
    userInput := m.textInput[builderIndex].Value()
    if userInput == "" {
        goto End
    }
    switch builderIndex {
    case builderMoveWorkdir:
        *m.workdir += "/" + userInput
        goto End
    case builderChangeDirPath:
        builderInitIndex,_ := strconv.Atoi(m.textInput[builderInit].Value())
        *(*m.dirMap)[*m.optList[builderInitIndex].key] = userInput
    case builderChangeFilePath:
        builderInitIndex,_ := strconv.Atoi(m.textInput[builderInit].Value())
        if filepath.Ext(m.textInput[builderChangeFilePath].Value()) != filepath.Ext(*m.optList[builderInitIndex].key) {
            m.cmdPage = -1
            m.textInput[builderChangeFilePath].Reset()
            m.textInput[builderInit].Reset()
            return func() tea.Msg {return teaErr(fmt.Errorf("incorrect file type. please use a filename with the '%s' format for input",filepath.Ext(*m.optList[builderInitIndex].key)))}
        }
        if userInput != *m.optList[builderInitIndex].key {
            (*m.buildMap)[(*m.dirMap)[*m.optList[builderInitIndex].dependOn.key]][userInput] = (*m.buildMap)[(*m.dirMap)[*m.optList[builderInitIndex].dependOn.key]][*m.optList[builderInitIndex].key]
            delete((*m.buildMap)[(*m.dirMap)[*m.optList[builderInitIndex].dependOn.key]], *m.optList[builderInitIndex].key)      
        }
    }
    m.conflictIndex,m.buildKeys,m.optList = map[int]bool{},[]*string{},[]optionList{}
    sortBuildKeys(m.buildMap,&m.buildKeys)
    generateBuildPath(m.workdir,&m.optList,m.dirMap,&m.buildKeys,m.buildMap,&m.conflictIndex)
    m.optionNumber = generateOptionNumber(len(m.optList))
    End:
    m.textInput[builderIndex].Reset()
    m.cmdPage = 1
    m.textInput[builderInit].Reset()
    m.textInput[builderInit].Focus()
    return nil    
}

func stringPointer(v string) *string {
	return &v
}

func (m *teaModelBuilderCMD) validateDirPath(input string) error {
    if !regexp.MustCompile(`^((/[a-zA-Z0-9-/_]+)+|/)$`).MatchString(*m.workdir+"/"+input) || strings.Contains(*m.workdir+"/"+input, "//") {
        return fmt.Errorf("input must be in a valid directory path format without special character (except - or _ )")
    }
    return nil
}

func (m *teaModelBuilderCMD) validateFilePath(input string) error {
    if !regexp.MustCompile(`^((/[a-zA-Z0-9-/._]+)+|/)$`).MatchString(*m.workdir+"/"+input) || strings.Contains(*m.workdir+"/"+input, "//") {
        return fmt.Errorf("input must be in a valid file path format without special character (except - or _ )")
    }
    return nil
}

func isAppNameFile[V map[string]templateCaller | *templateCaller](val V) bool {
    switch value:=any(val).(type) {
    case map[string]templateCaller:
        for _, v := range value {
            if isAppNameFile(&v) {
                return true
            }
        }
    case *templateCaller:
        for _, item := range value.param {
            if item == appname {
                return true
            }
        }
    }
    return false
}

func containStringOR(text string, content ...string) bool {
	if len(content) == 0 {
		return false
	}
	for _, i := range content {
		if containStringItem(&text, &i) {
			return true
		}
	}
	return false
}

func containStringAND(text string, content ...string) bool {
	if len(content) == 0 {
		return false
	}
	for _, c := range content {
		if !containStringItem(&text, &c) {
			return false
		}
	}
	return true
}

func containStringItem(text, content *string) bool {
	textItems := strings.Fields(*text)
	contentItems := strings.Fields(*content)
	if len(textItems) < len(contentItems) {
		return false
	}
	for i := 0; i <= len(textItems)-len(contentItems); i++ {
		tmpTextItems := textItems[i : i+len(contentItems)]
		tmpText := strings.Join(tmpTextItems, " ")
		if tmpText == *content {
			return true
		}
	}
	return false
}

func (m *teaModelBuilderCMD) promptGetInitLabel() (label string) {
    label = "The following files will be created in current project path ("+ *m.workdir +").\n"
    for i, list := range m.optList {
        label += "["+ fmt.Sprint(i)+"] "
        switch list.types {
            case "file":
                label += "  "+ *list.key
            default:
                label += *(*m.dirMap)[*list.key]
        }
        if (m.conflictIndex)[i] {
            label += " (conflict)"
        }
        label += "\n"
    }
    if len(m.conflictIndex) > 0 {
            label += "---"+fmt.Sprint(len(m.conflictIndex))+" conflicting file(s) detected---\n"
            label += "Enter 'Y' to overwrite, 'N' to abort, 'S' to skip all conflicting items, '[number]' to change the path location, and 'M' to move all files that will be generated to another directory within the current project.\n"
            label += "Input : "
            return label
    } 
    label += "---All files are safe to generate---\n"
    label += "Enter 'Y' to create all files and directories, 'N' to abort, '[number]' to change the path location, and 'M' to move all files that will be generated to another directory within the current project.\n"
    label += "Input : "
    return label
}

func sortBuildKeys(buildMap *map[*string]map[string]templateCaller, buildKeys *[]*string) {
    for k := range *buildMap {
        *buildKeys = append(*buildKeys, k)
    }
    sort.Slice(*buildKeys, func(i, j int) bool {
		return *(*buildKeys)[i] < *(*buildKeys)[j] 
	})
}

func generateOptionNumber(length int)(optionNumber []string){
    for i := 0; i < length; i++ {
        optionNumber = append(optionNumber, strconv.Itoa(i)) 
    }
    return optionNumber
}

func generateBuildPath(workdir *string, optList *[]optionList,dirMap *map[string]*string, buildKeys *[]*string, buildMap *map[*string]map[string]templateCaller, conflictIndex *map[int]bool){
    if len(*optList) == 0 {
        for _, key  := range *buildKeys {
            generateBuildPathItem(optList,workdir,dirMap,key,(*buildMap)[key],conflictIndex)
        }
    }
}

func generateBuildPathItem(optList *[]optionList,workdir *string,dirMap *map[string]*string, keyBuild *string, buildItem map[string]templateCaller,conflictIndex *map[int]bool){
    for key, val := range *dirMap {
        if val == keyBuild {
            *optList = append(*optList,optionList{types: "dir",key: stringPointer(key)})
        } 
    }
    if *conflictIndex == nil {
		*conflictIndex = make(map[int]bool)
	}
    parentAtOptionListIndex := len(*optList)-1
    (*optList)[parentAtOptionListIndex].totalItem = new(int)
    for _,fileItem := range sortItemKeys(&buildItem) {
        *optList = append(*optList, optionList{types: "file",key: stringPointer(fileItem), dependOn: &(*optList)[parentAtOptionListIndex]})
        *(*optList)[parentAtOptionListIndex].totalItem += 1
        if _,err := os.Stat(*workdir+"/"+*keyBuild+"/"+fileItem); !os.IsNotExist(err){
            (*conflictIndex)[len(*optList)-1] = true
        }
    }
}

func sortItemKeys(item *map[string]templateCaller)(sorted []string){
    for i := range *item {
        sorted = append(sorted, i)
    }
    sort.Strings(sorted)
    return sorted
}
