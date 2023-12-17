package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	org     = "slb-it"
	project = "es-competency-management-system"
)

type User struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
}

type Variable struct {
	Value string `json:"value"`
}

type VariableGroup struct {
	Name      string              `json:"name"`
	Variables map[string]Variable `json:"variables"`
}

type VariableGroupsResponse struct {
	Count int             `json:"count"`
	Value []VariableGroup `json:"value"`
}

// CompareAndPrintDifference compares two variable groups and prints variables that are in the first group but not in the second.
func CompareAndPrintDifference(group1Name string, group1Vars map[string]struct{}, group2Name string, group2Vars map[string]struct{}) {
	var hasDifference bool

	fmt.Printf("\nVariables in %s but not in %s:\n", group1Name, group2Name)
	for name := range group1Vars {
		if _, exists := group2Vars[name]; !exists {
			fmt.Println(" - " + name)
			hasDifference = true
		}
	}

	if !hasDifference {
		fmt.Println(" - No differences found")
	}
}

// checkEnvOrFlag checks if the command-line flag is set; if not, it checks for an environment variable.
// If neither is set, it logs a fatal error.
func checkEnvOrFlag(flagValue, envVarName string) string {
	if flagValue != "" {
		return flagValue
	}

	envValue, exists := os.LookupEnv(envVarName)
	if !exists {
		log.Fatalf("No value was provided for '%s'.\n\nEither provide it as a command-line argument or set an environment variable called '%s'.\n\nExiting.", envVarName, envVarName)
	}

	return envValue
}

func main() {
	var (
		lib1        int
		lib2        int
		patFlag     string
		orgFlag     string
		projectFlag string
	)

	flag.IntVar(&lib1, "lib1", 0, "Variable Library 1")
	flag.IntVar(&lib2, "lib2", 0, "Variable Library 2")
	flag.StringVar(&patFlag, "pat", "", "Personal Access Token")
	flag.StringVar(&orgFlag, "org", "", "Azure Devops Organization")
	flag.StringVar(&projectFlag, "project", "", "Azure DevOps Project")
	flag.Parse()

	pat := checkEnvOrFlag(patFlag, "AZDO_PAT")
	org := checkEnvOrFlag(orgFlag, "AZDO_ORG")
	project := checkEnvOrFlag(projectFlag, "AZDO_PROJECT")

	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?groupIds=%d,%d&api-version=6.0-preview.2", org, project, lib1, lib2)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth("", pat)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response VariableGroupsResponse
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		panic(err)
	}

	group1 := response.Value[0]
	group2 := response.Value[1]

	variablesInGroup1 := make(map[string]struct{})
	variablesInGroup2 := make(map[string]struct{})

	for name := range group1.Variables {
		variablesInGroup1[name] = struct{}{}
	}

	for name := range group2.Variables {
		variablesInGroup2[name] = struct{}{}
	}

	CompareAndPrintDifference(group1.Name, variablesInGroup1, group2.Name, variablesInGroup2)
	CompareAndPrintDifference(group2.Name, variablesInGroup2, group1.Name, variablesInGroup1)
}
