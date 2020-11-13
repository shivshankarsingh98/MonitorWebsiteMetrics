package  main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

var DimentionPriority map[string]int64

type TreeNode struct {
	DimentionName string
	Metrics       map[string]int64
	FistChild  *TreeNode
	NextSibling *TreeNode
}

var root = &TreeNode{}

type RequestData struct {
	Root           *TreeNode
	DimentionNames []string
	TotalDimentions int64
	Metrics         map[string]int64
}

func (reqData RequestData) QueryFromTree() map[string]int64 {
	if reqData.Root.FistChild == nil {
		return map[string]int64{}
	}

	var dimentionLevel int64 = 0
	firstChild := reqData.Root.FistChild
	temp := &TreeNode{}
	for dimentionLevel < reqData.TotalDimentions  {
		found :=  false

		for firstChild != nil {
			if firstChild.DimentionName == reqData.DimentionNames[dimentionLevel] {
				found = true
				break
			}else {
				firstChild = firstChild.NextSibling
			}
		}
		if found == false  {
			return map[string]int64{}
		}
		temp = firstChild
		if firstChild != nil  && found == true {
			firstChild = firstChild.FistChild
		}
		dimentionLevel += 1
	}
	return temp.Metrics
}

func (reqData RequestData) InsertIntoTree() {
	temp := reqData.Root
	if temp.Metrics == nil {
		temp.Metrics = map[string]int64{}
	}
	metric := map[string]int64{}
	for key, val := range temp.Metrics{
		metric[key] += val
	}
	for key, val := range reqData.Metrics{
		metric[key] += val
	}
	temp.Metrics = metric

	var dimentionLevel int64 = 0
	for dimentionLevel <  reqData.TotalDimentions {
		if temp.FistChild == nil {
			temp.FistChild = &TreeNode {reqData.DimentionNames[dimentionLevel],reqData.Metrics,nil,nil}
			temp = temp.FistChild
		}else {
			firstChild := temp.FistChild
			found := false
			for firstChild.NextSibling != nil {
				if firstChild.DimentionName == reqData.DimentionNames[dimentionLevel] {
					metric := map[string]int64{}

					for key, val := range firstChild.Metrics{
						metric[key] += val
					}
					for key, val := range reqData.Metrics{
						metric[key] += val
					}

					firstChild.Metrics = metric
					found = true
					temp = firstChild
					break
				}else {
					firstChild = firstChild.NextSibling
				}
			}

			if firstChild.DimentionName == reqData.DimentionNames[dimentionLevel] && firstChild.NextSibling == nil {
				metric := map[string]int64{}
				for key, val := range firstChild.Metrics{
					metric[key] += val
				}
				for key, val := range reqData.Metrics{
					metric[key] += val
				}
				firstChild.Metrics = metric
				temp = firstChild

			} else if found == false {
				firstChild.NextSibling = &TreeNode{reqData.DimentionNames[dimentionLevel], reqData.Metrics, nil, nil}
				temp = firstChild.NextSibling
			}
		}
		dimentionLevel += 1
	}
}


type Query struct {
	Dim     []Dimension `json:"dim"`
	Metrics []Metric    `json:"metrics"`
}

type Dimension struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type Metric struct {
	Key string `json:"key"`
	Val int64    `json:"val"`
}

type DimentionWithPriority struct {
	Name string
	Priority int64
}

func insert(w http.ResponseWriter, r *http.Request) {
	responseBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintln(w, "Invalid Request Body")
		return
	}

	var response Query
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "Invalid Request Body")
		return
	}

	dimentions := []DimentionWithPriority{}

	for _, dimDetail := range response.Dim {
		dimention := DimentionWithPriority{dimDetail.Val,DimentionPriority[dimDetail.Key]}
		dimentions = append(dimentions, dimention)
	}

	sort.SliceStable(dimentions, func(i, j int) bool {
		return dimentions[i].Priority < dimentions[j].Priority
	})

	// To check if the insert Query has proper priority or not
	dimentionIndex := 0
	for dimentionIndex < len(dimentions)-1 {
		if dimentions[dimentionIndex+1].Priority  != 	dimentions[dimentionIndex].Priority + 1	{
			w.WriteHeader(400)
			fmt.Fprintln(w, "Invalid Dimentions")
			return
		}
		dimentionIndex += 1
	}

	dim := []string{}
	for _, dimentionName := range  dimentions {
		dim = append(dim,dimentionName.Name)
	}

	metric := map[string]int64{}
	for _, val := range response.Metrics {
		metric[val.Key] = val.Val
	}
	dimLen := int64(len(dim))

	requestData := RequestData{}
	requestData.Root = root
	requestData.DimentionNames = dim
	requestData.TotalDimentions = dimLen
	requestData.Metrics = metric

	requestData.InsertIntoTree()
	w.WriteHeader(200)

}

func query(w http.ResponseWriter, r *http.Request) {

	responseBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintln(w, "Invalid Request Body")
		return
	}

	var response Query
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintln(w, "Invalid Request Body")
		return
	}

	dimentions := []DimentionWithPriority{}

	for _, dimDetails := range response.Dim {
		dimention := DimentionWithPriority{dimDetails.Val,DimentionPriority[dimDetails.Key]}
		dimentions = append(dimentions, dimention)
	}

	sort.SliceStable(dimentions, func(i, j int) bool {
		return dimentions[i].Priority < dimentions[j].Priority
	})

	dimentionIndex := 0
	for dimentionIndex < len(dimentions)-1 {
		if dimentions[dimentionIndex+1].Priority  != 	dimentions[dimentionIndex].Priority + 1	{
			w.WriteHeader(400)
			fmt.Fprintln(w, "Invalid Dimentions")
			return
		}
		dimentionIndex += 1
	}

	dim := []string{}
	for _, dimentionName := range  dimentions {
		dim = append(dim,dimentionName.Name)
	}

	dimLen := int64(len(dim))

	requestData := RequestData{}
	requestData.Root = root
	requestData.DimentionNames = dim
	requestData.TotalDimentions = dimLen

	result := requestData.QueryFromTree()

	var returnResponse Query
	returnResponse.Dim = response.Dim

	for  key, val := range result {
		a := Metric{Key: key, Val: val}
		returnResponse.Metrics = append(returnResponse.Metrics, a)
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(returnResponse)
}

func handleRequests() {
	http.HandleFunc("/v1/insert", insert)
	http.HandleFunc("/v1/query", query)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main(){
	setDimentionsPriority()
	handleRequests()
}

type PriorityData struct {
	Priority map[string]int64  `yaml:"Priority"`
}

func setDimentionsPriority() {
	jsonFile, err := ioutil.ReadFile("priority.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	prt := &PriorityData{}
	err = json.Unmarshal(jsonFile, prt)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	DimentionPriority = prt.Priority
}