package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {
	switch args["operation"] {
	case "add":
		return PerfomAdd(args, writer)
	case "list":
		return PerfomList(args, writer)
	case "findById":
		return PerfomFindById(args, writer)
	case "remove":
		return PerformRemove(args, writer)
	default:
		if err := validateArguments(args, false, false); err != nil {
			return err
		}
		return fmt.Errorf("Operation %v not allowed!", args["operation"])
	}
}

func PerfomAdd(args Arguments, writer io.Writer) error {
	if err := validateArguments(args, true, false); err != nil {
		return err
	}
	if err := AddItem(args["item"], args["fileName"]); err != nil {
		_, err2 := writer.Write([]byte(err.Error()))
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func PerfomList(args Arguments, writer io.Writer) error {
	if err := validateArguments(args, false, false); err != nil {
		return err
	}
	items, err := ListItems(args["fileName"])
	if err != nil {
		return err
	}
	if len(items) > 0 {
		data := dumpItemsToString(items)
		_, err = writer.Write([]byte(data))
		if err != nil {
			return err
		}
	}
	return nil
}

func PerfomFindById(args Arguments, writer io.Writer) error {
	if err := validateArguments(args, false, true); err != nil {
		return err
	}
	item, err := FindItemById(args["id"], args["fileName"])
	if err == nil {
		_, err := writer.Write([]byte(dumpItemToString(item)))
		if err != nil {
			return err
		}
	}
	return nil
}

func PerformRemove(args Arguments, writer io.Writer) error {
	if err := validateArguments(args, false, true); err != nil {
		return err
	}
	if err := RemoveItem(args["id"], args["fileName"]); err != nil {
		_, err2 := writer.Write([]byte(err.Error()))
		if err2 != nil {
			return err2
		}
	}
	return nil
}

// Parse command-line arguments with retrieving id value from item data map
func parseArgs() Arguments {
	var flagOperation = flag.String("operation", "", "operation to manipulate item's data")
	var flagFileName = flag.String("fileName", "", "file name to store item's data")
	var flagItem = flag.String("item", "", "map with item's data")
	var flagId = flag.String("id", "", "map with item's data")
	flag.Parse()
	return Arguments{
		"id":        *flagId,
		"operation": *flagOperation,
		"item":      *flagItem,
		"fileName":  *flagFileName,
	}
}

// Return errors of validation from command-line arguments
func validateArguments(args Arguments, isMandatoryItem bool, isMandatoryId bool) error {
	if len(args["fileName"]) == 0 {
		return errors.New("-fileName flag has to be specified")
	}
	_, err := os.Stat(args["fileName"])
	if os.IsNotExist(err) {
		file, err := os.Create(args["fileName"])
		if err != nil {
			return err
		}
		file.WriteString("[]")
		defer file.Close()
	} else {
		if err != nil {
			return err
		}
	}
	if len(args["operation"]) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	if isMandatoryItem && len(args["item"]) == 0 {
		return errors.New("-item flag has to be specified")
	}
	if isMandatoryId && len(args["id"]) == 0 {
		return errors.New("-id flag has to be specified")
	}
	return nil
}

// Add item to file if provided id is unique within file data
func AddItem(item string, fileName string) error {
	// Unmarshal item values
	var itemMap map[string]interface{}
	if err := json.Unmarshal([]byte(item), &itemMap); err != nil {
		return err
	}
	id := itemMap["id"].(string)
	// Check that there is no item with same id
	items, err := ListItems(fileName)
	if err != nil {
		return err
	}
	for _, m := range items {
		if m["id"].(string) == id {
			return fmt.Errorf("Item with id %v already exists", id)
		}
	}
	// Add item to items, then wtite items to file
	items = append(items, itemMap)
	data := dumpItemsToString(items)
	if err := os.WriteFile(fileName, []byte(data), 0644); err != nil {
		return err
	}
	return nil
}

// Return slice of items if there are valid data in file
func ListItems(fileName string) ([]map[string]interface{}, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, err
	}
	return items, nil
}

// Return item by id if provided id is exists within file
func FindItemById(id string, fileName string) (map[string]interface{}, error) {
	items, err := ListItems(fileName)
	if err != nil {
		return nil, err
	}
	for _, m := range items {
		if m["id"].(string) == id {
			return m, nil
		}
	}
	return nil, fmt.Errorf("Item with id %v not found", id)
}

// Remove item by id if provided id is exists within file
func RemoveItem(id string, fileName string) error {
	items, err := ListItems(fileName)
	if err != nil {
		return err
	}
	for i, m := range items {
		if m["id"].(string) == id {
			items = append(items[:i], items[i+1:]...)
			data := dumpItemsToString(items)
			if err := os.WriteFile(fileName, []byte(data), 0644); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("Item with id %v not found", id)
}

// Return string from slice of items presented in predefined order (id, email, age)
func dumpItemsToString(items []map[string]interface{}) string {
	var stringItems = make([]string, 0)
	for _, item := range items {
		stringItems = append(stringItems, dumpItemToString(item))
	}
	return "[" + strings.Join(stringItems, ",") + "]"
}

// Return string from items presented in predefined order (id, email, age)
func dumpItemToString(item map[string]interface{}) string {
	return fmt.Sprintf(
		"{\"id\":\"%v\",\"email\":\"%v\",\"age\":%v}",
		item["id"].(string), item["email"].(string), item["age"].(float64),
	)
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
