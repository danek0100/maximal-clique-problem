package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

const PATH = "./input_files"

var OPTIMAL_SOLUTION = map[string]int{
	"C125.9.clq":         34,
	"MANN_a27.clq":       126,
	"MANN_a9.clq":        16,
	"brock200_1.clq":     20,
	"brock200_2.clq":     10,
	"brock200_3.clq":     14,
	"brock200_4.clq":     16,
	"brock400_1.clq":     24,
	"brock400_2.clq":     25,
	"brock400_3.clq":     24,
	"brock400_4.clq":     24,
	"gen200_p0.9_44.clq": 40,
	"gen200_p0.9_55.clq": 48,
	"hamming8-4.clq":     16,
	"johnson16-2-4.clq":  8,
	"johnson8-2-4.clq":   4,
	"keller4.clq":        11,
	"p_hat1000-1.clq":    10,
	"p_hat1000-2.clq":    46,
	"p_hat1500-1.clq":    11,
	"p_hat300-3.clq":     34,
	"p_hat500-3.clq":     49,
	"san1000.clq":        10,
	"sanr200_0.9.clq":    41,
	"sanr400_0.7.clq":    21,
}

type VertexEdgePair struct {
	Vertex1 int
	Vertex2 int
}

func getTestingData(path string) (map[string][]VertexEdgePair, map[string][2]int, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}

	testingData := make(map[string][]VertexEdgePair)
	testingDataAmountInfo := make(map[string][2]int)

	for _, file := range files {
		if !file.IsDir() {
			content, err := ioutil.ReadFile(path + "/" + file.Name())
			if err != nil {
				return nil, nil, err
			}

			lines := strings.Split(string(content), "\n")
			var data []VertexEdgePair
			for _, line := range lines {
				if strings.HasPrefix(line, "p ") {
					parts := strings.Fields(line)
					amountVer, _ := strconv.Atoi(parts[2])
					amountEdge, _ := strconv.Atoi(parts[3])
					testingDataAmountInfo[file.Name()] = [2]int{amountVer, amountEdge}
				}
				if strings.HasPrefix(line, "e ") {
					parts := strings.Fields(line)
					ver1, _ := strconv.Atoi(parts[1])
					ver2, _ := strconv.Atoi(parts[2])
					data = append(data, VertexEdgePair{ver1, ver2})
				}
			}
			testingData[file.Name()] = data
		}
	}

	return testingData, testingDataAmountInfo, nil
}

func getAdjacencyMatrix(amountVer int, data []VertexEdgePair) [][]int {
	adjacencyMatrix := make([][]int, amountVer)
	for i := range adjacencyMatrix {
		adjacencyMatrix[i] = make([]int, amountVer)
	}

	for _, edge := range data {
		adjacencyMatrix[edge.Vertex1-1][edge.Vertex2-1] = 1
		adjacencyMatrix[edge.Vertex2-1][edge.Vertex1-1] = 1
	}

	return adjacencyMatrix
}

type VertexDegree struct {
	Vertex int
	Degree int
}

func getSortDegreeVert(adjacencyMatrix [][]int) map[int]int {
	degreeVertices := make(map[int]int)
	for i, row := range adjacencyMatrix {
		count := 0
		for _, value := range row {
			if value == 1 {
				count++
			}
		}
		degreeVertices[i] = count
	}

	sortedVertices := make([]VertexDegree, 0, len(degreeVertices))
	for vertex, degree := range degreeVertices {
		sortedVertices = append(sortedVertices, VertexDegree{vertex, degree})
	}
	sort.Slice(sortedVertices, func(i, j int) bool {
		return sortedVertices[i].Degree > sortedVertices[j].Degree
	})

	result := make(map[int]int)
	for _, vd := range sortedVertices {
		result[vd.Vertex] = vd.Degree
	}

	return result
}

func checkClique(adjacencyMatrix [][]int, clique []int) bool {
	for _, vert1 := range clique {
		for _, vert2 := range clique {
			if vert1 != vert2 {
				if adjacencyMatrix[vert1][vert2] == 0 {
					return false
				}
			}
		}
	}
	return true
}

func columnToName(col int) string {
	var columnName string
	for col > 0 {
		col--
		columnName = string('A'+col%26) + columnName
		col /= 26
	}
	return columnName
}

func saveResult(result map[string][3]interface{}, nameAlgorithm string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := []string{"MAX_CLIQUE", "TIME", "CLIQUE", "OPTIMAL_SOLUTION", "SOLVED"}
	for i, h := range headers {
		cell := columnToName(i+1) + fmt.Sprint(1)
		f.SetCellValue(sheet, cell, h)
	}

	row := 2
	for key, values := range result {
		cell := columnToName(1) + fmt.Sprint(row)
		f.SetCellValue(sheet, cell, key)

		maxClique := values[0].(int)
		time := values[1].(float64)
		clique := fmt.Sprintf("%v", values[2])
		optimalSolution := OPTIMAL_SOLUTION[key]
		solved := maxClique >= optimalSolution

		f.SetCellValue(sheet, "B"+fmt.Sprint(row), maxClique)
		f.SetCellValue(sheet, "C"+fmt.Sprint(row), time)
		f.SetCellValue(sheet, "D"+fmt.Sprint(row), clique)
		f.SetCellValue(sheet, "E"+fmt.Sprint(row), optimalSolution)
		f.SetCellValue(sheet, "F"+fmt.Sprint(row), solved)

		row++
	}

	err := f.SaveAs(nameAlgorithm + ".xlsx")
	return err
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getStartVert(degreeVertices map[int]int) int {
	maxDegree := -1
	for _, degree := range degreeVertices {
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	verticesWithSameDegree := []int{}
	for vertex, degree := range degreeVertices {
		if degree == maxDegree {
			verticesWithSameDegree = append(verticesWithSameDegree, vertex)
		}
	}

	return verticesWithSameDegree[rand.Intn(len(verticesWithSameDegree))]
}

func getNextVert(currentVert int, listOfVertices map[int]struct{}, adjacencyMatrix [][]int) int {
	listOfNeighbours := []int{}
	for i, edge := range adjacencyMatrix[currentVert] {
		if edge == 1 {
			if _, exists := listOfVertices[i]; exists {
				listOfNeighbours = append(listOfNeighbours, i)
			}
		}
	}

	if len(listOfNeighbours) == 0 {
		return -1
	}

	return listOfNeighbours[rand.Intn(len(listOfNeighbours))]
}

func getClique(degreeVertices map[int]int, clique []int, adjacencyMatrix [][]int) []int {
	listOfVertices := make(map[int]struct{})
	for k := range degreeVertices {
		listOfVertices[k] = struct{}{}
	}

	for len(listOfVertices) > 0 {
		nextVertex := getNextVert(clique[len(clique)-1], listOfVertices, adjacencyMatrix)

		if nextVertex == -1 {
			break
		}

		clique = append(clique, nextVertex)
		delete(listOfVertices, nextVertex)

		if adjacencyMatrix[clique[len(clique)-1]][clique[len(clique)-2]] == 1 {
			if !checkClique(adjacencyMatrix, clique) {
				clique = clique[:len(clique)-1]
			} else {
				delete(degreeVertices, nextVertex)
			}
		} else {
			clique = clique[:len(clique)-1]
		}
	}

	return clique
}

func findMaxClique(degreeVertices map[int]int, adjacencyMatrix [][]int) []int {
	currentCliqueBefore := []int{}
	currentCliqueAfter := []int{}

	startVertex := getStartVert(degreeVertices)
	currentCliqueBefore = append(currentCliqueBefore, startVertex)
	delete(degreeVertices, startVertex)

	currentCliqueBefore = getClique(degreeVertices, currentCliqueBefore, adjacencyMatrix)

	if len(currentCliqueBefore) > 1 {
		idx1 := rand.Intn(len(currentCliqueBefore))
		currentCliqueBefore = append(currentCliqueBefore[:idx1], currentCliqueBefore[idx1+1:]...)

		idx2 := rand.Intn(len(currentCliqueBefore))
		currentCliqueBefore = append(currentCliqueBefore[:idx2], currentCliqueBefore[idx2+1:]...)
	}

	currentCliqueAfter = getClique(degreeVertices, currentCliqueBefore, adjacencyMatrix)

	if len(currentCliqueBefore) >= len(currentCliqueAfter) {
		return currentCliqueBefore
	}
	return currentCliqueAfter
}

func main() {
	testingData, testingDataAmountInfo, err := getTestingData(PATH)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	//fmt.Println("=== Testing Data ===")
	//for file, data := range testingData {
	//	fmt.Printf("File: %s\n", file)
	//	for _, pair := range data {
	//		fmt.Printf("(%d, %d) ", pair.Vertex1, pair.Vertex2)
	//	}
	//	fmt.Println() // New line for better formatting
	//}

	fmt.Println("\n=== Testing Data Amount Info ===")
	for file, info := range testingDataAmountInfo {
		fmt.Printf("File: %s, Vertices: %d, Edges: %d\n", file, info[0], info[1])
	}

	fmt.Println()
	totalInfo := make(map[string][3]interface{})

	for file, data := range testingData {
		fmt.Println(file)
		dataInfo := testingDataAmountInfo[file]
		adjacencyMatrix := getAdjacencyMatrix(dataInfo[0], data)
		bestClique := []int{}

		for l := 0; l < 7000; l++ {
			timeStart := time.Now()

			degreeVertices := getSortDegreeVert(adjacencyMatrix)
			currentClique := findMaxClique(degreeVertices, adjacencyMatrix)

			timeEnd := time.Now()

			if len(currentClique) > len(bestClique) {
				bestClique = currentClique
				totalInfo[file] = [3]interface{}{len(bestClique), timeEnd.Sub(timeStart).Seconds(), bestClique}
				optimalSolution, exists := OPTIMAL_SOLUTION[file]
				if exists && len(bestClique) >= optimalSolution {
					fmt.Println("Optimal solution found")
					break
				}
			}
		}
	}

	for file, info := range totalInfo {
		fmt.Println("File:", file)
		fmt.Println("Max Clique Size:", info[0])
		fmt.Println("Time:", info[1], "seconds")
		fmt.Println("Clique:", info[2])
		fmt.Println()
	}

	err = saveResult(totalInfo, "results")
	if err != nil {
		fmt.Println("Error saving results:", err)
	} else {
		fmt.Println("Results saved successfully!")
	}
}
