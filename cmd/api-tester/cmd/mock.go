package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gaoyong06/api-tester/internal/mock"
	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/spf13/cobra"
)

var (
	// mock u547du4ee4u7684u6807u5fd7
	mockOutputFile string
	count         int
	schemaPath    string
	customRules   string
	generateAll   bool
	schemaType    string
	prettyPrint   bool
)

// mockCmd u8868u793a mock u5b50u547du4ee4
var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "u751fu6210u6a21u62dfu6570u636e",
	Long: `u751fu6210u6a21u62dfu6570u636eu547du4ee4u7528u4e8eu57fau4e8e OpenAPI/Swagger u89c4u8303u751fu6210u6a21u62dfu6570u636eu3002

u53efu4ee5u751fu6210u5355u4e2au6a21u5f0fu7684u6570u636euff0cu4e5fu53efu4ee5u751fu6210u6574u4e2a API u7684u6240u6709u6a21u5f0fu6570u636eu3002
u652fu6301u81eau5b9au4e49u89c4u5219u6765u63a7u5236u7279u5b9au5b57u6bb5u7684u751fu6210u65b9u5f0fu3002`,
	Run: func(cmd *cobra.Command, args []string) {
		// u68c0u67e5u5fc5u8981u53c2u6570
		if specFile == "" {
			cmd.Help()
			fmt.Println("\nu9519u8bef: u5fc5u987bu63d0u4f9b API u89c4u8303u6587u4ef6 (--spec)")
			os.Exit(1)
		}

		if !generateAll && schemaPath == "" {
			cmd.Help()
			fmt.Println("\nu9519u8bef: u5fc5u987bu63d0u4f9bu6a21u5f0fu8defu5f84 (--schema) u6216u8005u4f7fu7528 --all u751fu6210u6240u6709u6a21u5f0f")
			os.Exit(1)
		}

		// u89e3u6790 API u5b9au4e49
		apiDef, err := parser.ParseSwaggerFile(specFile)
		if err != nil {
			log.Fatalf("u65e0u6cd5u89e3u6790 API u5b9au4e49: %v", err)
		}

		// u521bu5efau6a21u62dfu6570u636eu751fu6210u5668
		generator := mock.NewMockDataGenerator()

		// u52a0u8f7du81eau5b9au4e49u89c4u5219uff08u5982u679cu6709uff09
		if customRules != "" {
			rules, err := loadCustomRules(customRules)
			if err != nil {
				log.Fatalf("u65e0u6cd5u52a0u8f7du81eau5b9au4e49u89c4u5219: %v", err)
			}
			generator.SetCustomRules(rules)
		}

		// u786eu4fddu8f93u51fau76eeu5f55u5b58u5728
		if mockOutputFile != "" {
			outputDir := filepath.Dir(mockOutputFile)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("u65e0u6cd5u521bu5efau8f93u51fau76eeu5f55: %v", err)
			}
		}

		// u751fu6210u6a21u62dfu6570u636e
		if generateAll {
			// u751fu6210u6240u6709u6a21u5f0fu7684u6570u636e
			result := make(map[string]interface{})

			// u751fu6210u6240u6709u5b9au4e49u7684u6a21u5f0fu6570u636e
			for name, schema := range apiDef.Schemas {
				fmt.Printf("u751fu6210u6a21u5f0f '%s' u7684u6a21u62dfu6570u636e...\n", name)
				
				// u751fu6210u591au4e2au793au4f8b
				examples := make([]interface{}, count)
				for i := 0; i < count; i++ {
					data, err := generator.GenerateFromSchema(schema)
					if err != nil {
						log.Printf("u8b66u544a: u65e0u6cd5u751fu6210u6a21u5f0f '%s' u7684u6570u636e: %v", name, err)
						continue
					}
					examples[i] = data
				}
				
				result[name] = examples
			}

			// u8f93u51fau6240u6709u6a21u5f0fu6570u636e
			outputMockData(result, mockOutputFile, prettyPrint)
		} else {
			// u751fu6210u6307u5b9au6a21u5f0fu7684u6570u636e
			var schema interface{}
			var schemaName string

			// u5148u5c1du8bd5u4eceu5b9au4e49u7684u6a21u5f0fu4e2du67e5u627e
			if s, ok := apiDef.Schemas[schemaPath]; ok {
				schema = s
				schemaName = schemaPath
			} else {
				// u5982u679cu4e0du662fu5b9au4e49u7684u6a21u5f0fu540du79f0uff0cu5c1du8bd5u89e3u6790u4e3au8defu5f84
				// u8fd9u91ccu53efu4ee5u5b9eu73b0u4eceu8defu5f84u4e2du63d0u53d6u6a21u5f0fu7684u903bu8f91
				// u4f8bu5982uff0cu4eceu8defu5f84 /pets/{petId} u7684 GET u64cdu4f5cu7684u54cdu5e94u4e2du63d0u53d6u6a21u5f0f
				// u8fd9u91ccu7b80u5316u5904u7406uff0cu76f4u63a5u62a5u9519
				log.Fatalf("u9519u8bef: u627eu4e0du5230u6307u5b9au7684u6a21u5f0f '%s'", schemaPath)
			}

			fmt.Printf("u751fu6210u6a21u5f0f '%s' u7684u6a21u62dfu6570u636e...\n", schemaName)

			// u751fu6210u591au4e2au793au4f8b
			examples := make([]interface{}, count)
			for i := 0; i < count; i++ {
				data, err := generator.GenerateFromSchema(schema)
				if err != nil {
					log.Fatalf("u65e0u6cd5u751fu6210u6a21u5f0fu6570u636e: %v", err)
				}
				examples[i] = data
			}

			// u8f93u51fau6307u5b9au6a21u5f0fu6570u636e
			var result interface{}
			if count == 1 {
				// u5982u679cu53eau751fu6210u4e00u4e2au793au4f8buff0cu76f4u63a5u8fd4u56deu5bf9u8c61u800cu4e0du662fu6570u7ec4
				result = examples[0]
			} else {
				result = examples
			}

			outputMockData(result, mockOutputFile, prettyPrint)
		}
	},
}

// loadCustomRules u4eceu6587u4ef6u52a0u8f7du81eau5b9au4e49u89c4u5219
func loadCustomRules(filePath string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("u65e0u6cd5u8bfbu53d6u89c4u5219u6587u4ef6: %v", err)
	}

	var rules map[string]interface{}
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("u65e0u6cd5u89e3u6790u89c4u5219u6587u4ef6: %v", err)
	}

	return rules, nil
}

// outputMockData u8f93u51fau6a21u62dfu6570u636e
func outputMockData(data interface{}, mockOutputFile string, prettyPrint bool) {
	// u5e8fu5217u5316u6570u636e
	var jsonData []byte
	var err error

	if prettyPrint {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		log.Fatalf("u65e0u6cd5u5e8fu5217u5316u6a21u62dfu6570u636e: %v", err)
	}

	// u8f93u51fau6570u636e
	if mockOutputFile == "" {
		// u8f93u51fau5230u6807u51c6u8f93u51fa
		fmt.Println(string(jsonData))
	} else {
		// u8f93u51fau5230u6587u4ef6
		if err := ioutil.WriteFile(mockOutputFile, jsonData, 0644); err != nil {
			log.Fatalf("u65e0u6cd5u5199u5165u8f93u51fau6587u4ef6: %v", err)
		}
		fmt.Printf("u6a21u62dfu6570u636eu5df2u4fddu5b58u5230: %s\n", mockOutputFile)
	}
}

func init() {
	rootCmd.AddCommand(mockCmd)

	// u672cu5730u6807u5fd7
	mockCmd.Flags().StringVar(&mockOutputFile, "output-file", "", "u8f93u51fau6587u4ef6u8defu5f84 (u9ed8u8ba4u8f93u51fau5230u6807u51c6u8f93u51fa)")
	mockCmd.Flags().IntVar(&count, "count", 1, "u751fu6210u7684u6570u636eu6570u91cf")
	mockCmd.Flags().StringVar(&schemaPath, "schema", "", "u8981u751fu6210u6570u636eu7684u6a21u5f0fu540du79f0u6216u8defu5f84")
	mockCmd.Flags().StringVar(&customRules, "rules", "", "u81eau5b9au4e49u751fu6210u89c4u5219u6587u4ef6 (JSON u683cu5f0f)")
	mockCmd.Flags().BoolVar(&generateAll, "all", false, "u751fu6210u6240u6709u6a21u5f0fu7684u6570u636e")
	mockCmd.Flags().StringVar(&schemaType, "type", "object", "u6a21u5f0fu7c7bu578b (object, array, string, number, boolean)")
	mockCmd.Flags().BoolVar(&prettyPrint, "pretty", true, "u7f8eu5316u8f93u51fa JSON")
}
