package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/slmcmahon/go-azdo"
	slmcommon "github.com/slmcmahon/go-common"
)

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

	pat, err := slmcommon.CheckEnvOrFlag(patFlag, "AZDO_PAT")
	if err != nil {
		log.Fatal(err)
	}
	org, err := slmcommon.CheckEnvOrFlag(orgFlag, "AZDO_ORG")
	if err != nil {
		log.Fatal(err)
	}
	project, err := slmcommon.CheckEnvOrFlag(projectFlag, "AZDO_PROJECT")
	if err != nil {
		log.Fatal(err)
	}

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

	azdo.CompareAndPrintDifference(group1.Name, variablesInGroup1, group2.Name, variablesInGroup2)
	azdo.CompareAndPrintDifference(group2.Name, variablesInGroup2, group1.Name, variablesInGroup1)
}
