package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// TerraformState provides enough from the state json file to get the type,
// and the attributes of the created type
type TerraformState struct {
	TerraformVersion string `json:"terraform_version"`
	Resources        []struct {
		Type      string `json:"type"`
		Instances []struct {
			Attributes map[string]interface{}
		}
	}
}

func main() {

	var tfstate, filename, planName string

	flag.StringVar(&tfstate, "tfstate", "", "Terraform state file location")
	flag.StringVar(&filename, "out", "tfdoc.html", "Path and name of the html file to create (e.g /tmp/myoutput.html)")
	flag.StringVar(&planName, "name", "Terraform Output", "Name of the report")
	flag.Parse()
	if len(tfstate) == 0 {
		fmt.Fprintf(os.Stderr, "You must specify the path to a terraform state file")
	}

	err := parseState(tfstate, filename, planName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was an error parsing the terraform state file\n%s", err)
	}

	fmt.Printf("Successfully parsed state file, html output is saved to the following location: %s", filename)
}

func parseState(statePath, outputPath, planName string) error {
	bytes, err := ioutil.ReadFile(statePath)
	if err != nil {
		return err
	}
	terraformState := &TerraformState{}

	if err := json.Unmarshal(bytes, terraformState); err != nil {
		fmt.Println(err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	file.WriteString(`<html>
	<head>
	<style>
	table, th, td {
	  border-collapse: collapse;
	  width: 100%;
	  border: 2px solid rgb(200, 200, 200);
      letter-spacing: 1px;
      font-family: sans-serif;
      font-size: .8rem;
	}
	th, td {
	  padding: 5px;
	  text-align: left;
	  background-color: rgb(235, 235, 235);
	  width: 50%;
	}
	caption {
		text-align: left;
		weight: bold;
	}
	tr:nth-child(even) td {
		background-color: rgb(250, 250, 250);

	}
	
	tr:nth-child(odd) td {
		background-color: rgb(240, 240, 240);
	}
	#tableHeader {
	  font-family: sans-serif;
	  font-size: 1.0rem;
	}
	h1 {
		font-family: sans-serif;
	}
	
	</style>
	</head>
	<body>
	`)

	file.WriteString("<h1>Terraform State Output - " + planName + "</h1>")
	file.WriteString("<br><div id=tableHeader>Built with Terraform: " + terraformState.TerraformVersion + "</div><br>")

	for _, resource := range terraformState.Resources {
		for _, instance := range resource.Instances {
			resourceWriter(resource.Type, instance.Attributes, file)
		}
	}
	file.WriteString("</body></html>")
	return err
}

func resourceWriter(resourceType string, attributes map[string]interface{}, file *os.File) {
	file.WriteString(
		"<br><br><b><div id=tableHeader>" + resourceType +
			`</div></b>
			 <table>
	  		    <tr>
			      <th>Attribute</th>
				  <th>Value</th>
	  			</tr>`)

	//TODO: fix this tasty spaghetti
	for key, value := range attributes {
		switch v := value.(type) {
		case []interface{}: //Sometime we get an arry of interfaces back from the state, with nested values
			for _, nestedValue := range v {
				switch v := nestedValue.(type) {
				case map[string]interface{}:
					file.WriteString("<td>" + key + "</td><td><table>")
					for nestedKey, nestedValue := range v {
						nestedVal := fmt.Sprintf("%v", nestedValue)
						file.WriteString("<tr><td>" +
							nestedKey + "</td><td>" +
							nestedVal + "</td></tr>")
					}
					file.WriteString("</table></td></tr>")
				}
			}

		case map[string]interface{}:
			file.WriteString("<td>" + key + "</td><td><table>")
			for nestedKey, nestedValue := range v {
				nestedVal := fmt.Sprintf("%v", nestedValue)
				file.WriteString("<tr><td>" + nestedKey + "</td><td>" + nestedVal + "</td></tr>")
			}
			file.WriteString("</table></td></tr>")

		default:
			val := fmt.Sprintf("%v", v)
			file.WriteString("<tr><td>" + key + "</td><td>" + val + "</td></tr>")
		}

	}
	file.WriteString("</table>")
}
