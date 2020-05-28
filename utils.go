package exec

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

// Cd is a remote helper function that runs a `cd` before a command
func (e *Exec) Cd(path string) {
	command := "cd " + e.Parse(path)
	color.Green("[%s] %s %s", e.ServerContext.Name, color.GreenString(">"), command)
	e.ServerContext.sshClient.env = command + "; "
}

// CommandExist checks if a remote command exists on server
func (e *Exec) CommandExist(command string) bool {
	return e.Remote("if hash %s 2>/dev/null; then echo 'true'; fi", command).Bool()
}

// Parse parses {{var}} with Get(var)
func (e *Exec) Parse(text string) string {
	re := regexp.MustCompile(`\{\{\s*([\w\.\/]+)\s*\}\}`)
	if !re.MatchString(text) {
		return text
	}
	return re.ReplaceAllStringFunc(text, func(str string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(str, "{{"), "}}")
		if e.Has(name) {
			return e.Parse(e.Get(name).String())
		}
		return str
	})
}

// RunIfNoBinary runs a remote command if a binary is not found
// command can be an array of string commands or one a string command
func (e *Exec) RunIfNoBinary(binary string, command interface{}) (o Output) {
	return e.Remote("if [ ! -e \"`which %s`\" ]; then %s; fi", binary, commandToString(command))
}

// RunIfNoBinaries runs multiple RunIfNoBinary
func (e *Exec) RunIfNoBinaries(config map[string]interface{}) {
	for binary, command := range config {
		e.RunIfNoBinary(binary, command)
	}
}

// RunIf runs a remote command if condition is true
// command can be an array of string commands or one a string command
func (e *Exec) RunIf(condition string, command interface{}) (o Output) {
	return e.Remote("if %s; then %s; fi", condition, commandToString(command))
}

// RunIfs runs multiple RunIf
func (e *Exec) RunIfs(config map[string]interface{}) {
	for condition, command := range config {
		e.RunIf(condition, command)
	}
}

// UploadFileSudo uploads a local file to a remote file with sudo
func (e *Exec) UploadFileSudo(source, destination string) {
	tempFile := "/tmp/" + uuid.NewV4().String()
	e.Upload(source, tempFile)
	e.Remote("sudo mv %s %s", tempFile, destination)
}

// UploadTemplateFileSudo parses a local template file with context, and uploads it to a remote file with sudo
func (e *Exec) UploadTemplateFileSudo(source, destination string, context interface{}) {
	tempFile := "/tmp/" + uuid.NewV4().String()

	t, err := template.New(path.Base(source)).ParseFiles(source)
	if err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, context); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}

	if err := ioutil.WriteFile(tempFile, tpl.Bytes(), os.FileMode(0644)); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		e.Upload(tempFile, tempFile)
		e.Local("rm %s", tempFile)
		e.Remote("sudo mv %s %s", tempFile, destination)
	}
}

// UploadTemplateStringSudo uploads a string content to a remote file with sudo
func (e *Exec) UploadTemplateStringSudo(content, destination string) {
	tempFile := "/tmp/" + uuid.NewV4().String()
	if err := ioutil.WriteFile(tempFile, []byte(content), os.FileMode(0644)); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		e.Upload(tempFile, tempFile)
		e.Local("rm %s", tempFile)
		e.Remote("sudo mv %s %s", tempFile, destination)
	}
}

// LocalTemplateFile parses a local template file with context, and moves it to a destination
func (e *Exec) LocalTemplateFile(source, destination string, context interface{}) {
	tempFile := "/tmp/" + uuid.NewV4().String()

	t, err := template.New(path.Base(source)).ParseFiles(source)
	if err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, context); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}

	if err := ioutil.WriteFile(tempFile, tpl.Bytes(), os.FileMode(0644)); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		e.Local("mv %s %s", tempFile, destination)
	}
}

// CompileLocalTemplateFile parses a local source file template with context and returns it
func (e *Exec) CompileLocalTemplateFile(source string, context interface{}) string {
	t, err := template.New(path.Base(source)).ParseFiles(source)
	if err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, context); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	return tpl.String()
}

// CompileLocalTemplateString parses a local source string template with context and returns it
func (e *Exec) CompileLocalTemplateString(source string, context interface{}) string {
	t, err := template.New(uuid.NewV4().String()).Parse(source)
	if err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, context); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	}
	return tpl.String()
}

// ReplaceInRemoteFile replaces a search string with a replace string, in a remote file
func (e *Exec) ReplaceInRemoteFile(file, search, replace string) {
	tempFile := "/tmp/" + uuid.NewV4().String()
	e.Remote("sudo cp %s %s ; sudo chown %s %s", file, tempFile, e.ServerContext.GetUser(), tempFile)
	e.Download(tempFile, tempFile)
	if tempFileContent, err := ioutil.ReadFile(tempFile); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		tempFileContent := strings.Replace(string(tempFileContent), search, e.Parse(replace), -1)
		if err := ioutil.WriteFile(tempFile, []byte(tempFileContent), os.FileMode(0644)); err != nil {
			color.Red("[%s] %s %s", "local", "<", err)
		} else {
			e.UploadFileSudo(tempFile, file)
			e.Remote("sudo rm -rf %s", tempFile)
			e.Local("rm -rf %s", tempFile)
		}
	}
}

// AddInRemoteFile appends a text string to a remote file
func (e *Exec) AddInRemoteFile(text, file string) {
	tempFile := "/tmp/" + uuid.NewV4().String()
	e.Remote("sudo cp %s %s ; sudo chown %s %s", file, tempFile, e.ServerContext.GetUser(), tempFile)
	e.Download(tempFile, tempFile)
	if tempFileContent, err := ioutil.ReadFile(tempFile); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		tempFileContent := string(tempFileContent) + e.Parse(text)
		if err := ioutil.WriteFile(tempFile, []byte(tempFileContent), os.FileMode(0644)); err != nil {
			color.Red("[%s] %s %s", "local", "<", err)
		} else {
			e.UploadFileSudo(tempFile, file)
			e.Remote("sudo rm -rf %s", tempFile)
			e.Local("rm -rf %s", tempFile)
		}
	}
}

// RemoveFromRemoteFile cuts out a text string from remote file
func (e *Exec) RemoveFromRemoteFile(text, file string) {
	tempFile := "/tmp/" + uuid.NewV4().String()
	e.Remote("sudo cp %s %s ; sudo chown %s %s", file, tempFile, e.ServerContext.GetUser(), tempFile)
	e.Download(tempFile, tempFile)
	if tempFileContent, err := ioutil.ReadFile(tempFile); err != nil {
		color.Red("[%s] %s %s", "local", "<", err)
	} else {
		tempFileContent := strings.Replace(string(tempFileContent), e.Parse(text), "", -1)
		if err := ioutil.WriteFile(tempFile, []byte(tempFileContent), os.FileMode(0644)); err != nil {
			color.Red("[%s] %s %s", "local", "<", err)
		} else {
			e.UploadFileSudo(tempFile, file)
			e.Remote("sudo rm -rf %s", tempFile)
			e.Local("rm -rf %s", tempFile)
		}
	}
}

// IsInRemoteFile return true if text is found in a remote file
func (e *Exec) IsInRemoteFile(text, file string) bool {
	text = strings.Trim(text, " ")
	return e.Remote("if [ \"`sudo cat %s | grep '%s'`\" ]; then echo 'true'; fi", file, text).Bool()
}

// Ask asks a question and waits for an answer
// first item from attributes is set as default value, which is optional
func (e *Exec) Ask(question string, attributes ...string) string {
	scanner := bufio.NewScanner(os.Stdin)

	var defaultResponse string

	if len(attributes) == 1 {
		defaultResponse = attributes[0]
		question += fmt.Sprintf(" [default: %s]", defaultResponse)
	}

	color.Green("[%s] %s %s", "local", ">", color.WhiteString(question))
	scanner.Scan()
	response := strings.TrimSpace(scanner.Text())

	if response == "" {
		response = defaultResponse
	}

	return response
}

// AskWithConfirmation asks a confirmation question and waits for an y/n answer
// first item from attributes is set as default value, which is optional
func (e *Exec) AskWithConfirmation(question string, attributes ...bool) bool {
	scanner := bufio.NewScanner(os.Stdin)

	var defaultResponse bool
	choices := map[string]bool{
		"y":   true,
		"yes": true,
		"n":   false,
		"no":  false,
	}

	if len(attributes) == 1 {
		defaultResponse = attributes[0]
		if defaultResponse {
			question += fmt.Sprintf(" [default: Y/n]")
		} else {
			question += fmt.Sprintf(" [default: y/N]")
		}
	}

	color.Green("[%s] %s %s", "local", ">", color.WhiteString(question))
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if choice, choiceValue := choices[response]; choiceValue {
		return choice
	}

	return defaultResponse
}

/*AskWithChoices asks a question with multiple choices and waits for an answer

First item from attributes must be a map with default and choices keys and string slice as values, example:
	```
	map[string]interface{}{
				"default": []string{
					"agent",
				},
				"choices": []string{
					"agent",
					"tty",
					"ssh",
				},
			}
    ```
*/
func (e *Exec) AskWithChoices(question string, attributes ...map[string]interface{}) (responses []string) {
	scanner := bufio.NewScanner(os.Stdin)

	var (
		attrs                map[string]interface{}
		defaultChoices       interface{}
		parsedDefaultChoices []string
		foundDefaultChoices  bool
		choices              interface{}
		parsedChoices        []string
		foundChoices         bool
		loop                 = true
	)

	if len(attributes) > 0 {
		attrs = attributes[0]

		defaultChoices, foundDefaultChoices = attrs["default"]
		choices, foundChoices = attrs["choices"]

		if foundDefaultChoices {
			parsedDefaultChoices = defaultChoices.([]string)
			foundDefaultChoices = len(parsedDefaultChoices) > 0
		}

		if foundChoices {
			parsedChoices = choices.([]string)
			foundChoices = len(parsedChoices) > 0
		}

		if foundDefaultChoices {
			question += fmt.Sprintf(" [default: %s]", strings.Join(parsedDefaultChoices, ", "))
		}

		if foundChoices {
			question += fmt.Sprintf("\nPlease pick one or more choices, one per line, from these: %s", strings.Join(parsedChoices, ", "))
		}
	}

	color.Green("[%s] %s %s", "local", ">", color.WhiteString(question))

	for loop {
		scanner.Scan()
		response := strings.TrimSpace(scanner.Text())

		if response == "" {
			loop = false
		}

		if foundChoices {
			if contains(parsedChoices, response) {
				responses = append(responses, response)
			}
		}
	}

	if len(responses) == 0 && foundDefaultChoices {
		return parsedDefaultChoices
	}

	return responses
}

func commandToString(run interface{}) string {
	var runS string
	rt := reflect.TypeOf(run)
	switch rt.Kind() {
	case reflect.Slice:
		runS = strings.Join(run.([]string), " ; ")
	case reflect.String:
		runS = run.(string)
	}
	return runS
}

func mergeArguments(removeMap map[string]string, maps ...map[string]*Argument) (output map[string]*Argument) {
	size := len(maps)
	if size == 0 {
		return output
	}
	if size == 1 {
		return maps[0]
	}
	output = make(map[string]*Argument)
	for _, m := range maps {
		for k, v := range m {
			if _, ok := removeMap[k]; !ok {
				output[k] = v
			}
		}
	}
	return output
}

func mergeOptions(removeMap map[string]string, maps ...map[string]*Option) (output map[string]*Option) {
	size := len(maps)
	if size == 0 {
		return output
	}
	if size == 1 {
		return maps[0]
	}
	output = make(map[string]*Option)
	for _, m := range maps {
		for k, v := range m {
			if _, ok := removeMap[k]; !ok {
				output[k] = v
			}
		}
	}
	return output
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
