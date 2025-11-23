package codegen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/logger"
)

type GPT4VCompletionRequest struct {
	Model            string                   `json:"model"`
	Messages         []map[string]interface{} `json:"messages"`
	Functions        []interface{}            `json:"functions,omitempty"`
	FunctionCall     interface{}              `json:"function_call,omitempty"`
	Stream           bool                     `json:"stream,omitempty"`
	Temperature      float64                  `json:"temperature,omitempty"`
	TopP             float64                  `json:"top_p,omitempty"`
	MaxTokens        int                      `json:"max_tokens,omitempty"`
	N                int                      `json:"n,omitempty"`
	BestOf           int                      `json:"best_of,omitempty"`
	FrequencyPenalty float64                  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64                  `json:"presence_penalty,omitempty"`
	Seed             int                      `json:"seed,omitempty"`
	LogitBias        map[string]float64       `json:"logit_bias,omitempty"`
	Stop             interface{}              `json:"stop,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type GPT4VCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`            // Unix timestamp (not JSON time)
	Model   string   `json:"model"`              // Model name
	FP      string   `json:"system_fingerprint"` // System fingerprint string
	ID      string   `json:"id"`                 // Response ID
	Object  string   `json:"object"`             // Object type
}

type Choice struct {
	Message ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	url = "https://api.openai.com/v1/chat/completions"

	//For the table or list, use <ui-tabular> webcomponent to create it.
	s_OPEN_AI_SYSTEM_PROMPT = `You are an expert web developer who specializes in building working website prototypes from low-fidelity wireframes. Your job is to accept low-fidelity designs and turn them into interactive and responsive working prototypes. When sent new designs with the image data, you should reply with a high fidelity working prototype as a single HTML file.
	
	You final output html file should have the similar layout and elements in the image provided. 

	Use tailwind CSS for styling. If you must use other CSS, place it in a style tag.
	
	Put any JavaScript in a script tag. Use unpkg or skypack to import any required JavaScript dependencies. Use Google fonts to pull in any open source fonts you require. If you have any images, load them from Unsplash or use solid colored rectangles as placeholders. 
	
	The designs may include flow charts, diagrams, labels, arrows, sticky notes, screenshots of other applications, or even previous designs. Treat all of these as references for your prototype. Use your best judgement to determine what is an annotation and what should be included in the final result. Treat anything in the color red as an annotation rather than part of the design. Do NOT include any of those annotations in your final result.
	
	Your prototype should look and feel much more complete and advanced than the wireframes provided. Flesh it out, make it real! Try your best to figure out what the designer wants and make it happen. If there are any questions or underspecified features, use what you know about applications, user experience, and website design patterns to "fill in the blanks". If you're unsure of how the designs should work, take a guess—it's better for you to get it wrong than to leave things incomplete. 
	
	Remember: you love your designers and want them to be happy. The more complete and impressive your prototype, the happier they will be. Good luck, you've got this!`

	s_OPENAI_USER_PROMPT = "Here are the latest wireframes. There are also some previous outputs here. We have run their code through an 'HTML to screenshot' library, that attempts to generate a screenshot of the page. The generated screenshot may have some inaccuracies, so use your knowledge of HTML and web development to figure out what any annotations are referring to, which may be different to what is visible in the generated screenshot. Make a new website based on these wireframes and notes and send back just the HTML file contents."

	s_OPENAI_USER_PROMPT_WITH_PREVIOUS_DESIGN = "Here are the latest wireframes. There are also some previous outputs here. Could you make a new website based on these wireframes and notes and send back just the html file?"
)

var s_OPENAI_UI_SYSTEM_PROMPT = `You are a skilled front-end developer who builds interactive prototypes from wireframes, and is an expert at CSS Grid and Flex design.
Your role is to transform low-fidelity wireframes into working front-end HTML code.

YOU MUST FOLLOW FOLLOWING RULES:

- Use HTML, CSS, JavaScript to build a responsive, accessible, polished prototype
- Leverage Tailwind for styling and layout (import as script <script src="https://cdn.tailwindcss.com"></script>)
- Inline JavaScript when needed
- Fetch dependencies from CDNs when needed (using unpkg or skypack)
- Source images from Unsplash or create applicable placeholders
- Interpret annotations as intended vs literal UI
- Fill gaps using your expertise in UX and business logic
- generate primarily for desktop UI, but make it responsive.
- Use grid and flexbox wherever applicable.
- Convert the wireframe in its entirety, don't omit elements if possible.

If the wireframes, diagrams, or text is unclear or unreadable, refer to provided text for clarification.

Your goal is a production-ready prototype that brings the wireframes to life.

Please output JUST THE HTML file containing your best attempt at implementing the provided wireframes.`

func getCodeFromImage(system_prompt string, user_prompt string, user_prompt_withprevious string, image string, apiKey string, openaimodel string, text string, grid string, theme string, previouseobj []map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "CodeAi"}
	result := make(map[string]interface{})

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegeneration.GetHtmlCodeFromImage", elapsed)
	}()

	if apiKey == "" {
		iLog.Error("API key is required")
		return result, errors.New("API key is required")
	}

	if openaimodel == "" {
		iLog.Error("OpenAI model is required")
		return result, errors.New("OpenAI model is required")
	}

	if image == "" {
		iLog.Error("Image is required")
		return result, errors.New("Image is required")
	}

	userContent := make([]map[string]interface{}, 0)
	userContent = append(userContent, map[string]interface{}{
		"type": "text",
		"text": user_prompt,
	})

	image_url := map[string]interface{}{
		"url":    image,
		"detail": "high",
	}
	userContent = append(userContent, map[string]interface{}{
		"type":      "image_url",
		"image_url": image_url,
	})

	if text != "" {
		userContent = append(userContent, map[string]interface{}{
			"type": "text",
			"text": fmt.Sprintf("Here's a list of all the text that we found in the design. Use it as a reference if anything is hard to read in the screenshot(s):\n%s", text),
		})
	}

	if grid != "" {
		userContent = append(userContent, map[string]interface{}{
			"type": "text",
			"text": fmt.Sprintf("The designs have a %s grid overlaid on top. Each cell of the grid is 10x10px.", grid), // Assuming grid is a color, adjust size as needed
		})
	}

	if previouseobj != nil && len(previouseobj) > 0 {
		for _, obj := range previouseobj {
			userContent = append(userContent, map[string]interface{}{
				"type": "text",
				"text": `The designs also included one of your previous result. Here's the image that you used as its source:`,
			})

			preimage_url := map[string]interface{}{
				"url":    obj["source"],
				"detail": "high",
			}

			userContent = append(userContent, map[string]interface{}{
				"type":      "image_url",
				"image_url": preimage_url,
			})

			userContent = append(userContent, map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf(`And here's the code that you generated from that image:%v`, obj["content"]),
			})
		}
	}

	message0 := map[string]interface{}{"role": "system",
		"content": system_prompt,
	}

	message1 := map[string]interface{}{"role": "user",
		"content": userContent,
	}

	messages := make([]map[string]interface{}, 2)
	messages[0] = message0
	messages[1] = message1

	request := GPT4VCompletionRequest{
		Model:       openaimodel, //"gpt-4o",
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0,
		Seed:        42,
		N:           1,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshalling request: %v", err))
		return result, err
	}

	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second) // Set a timeout
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestJson))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating request: %s", err))
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in calling OpenAi API: %s", err))
		return result, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading response body: %s", err))
		return result, err
	}

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling response: %s", err))
		return result, err
	}
	//	data := result.Choices[0].Message.Content
	iLog.Debug(fmt.Sprintf("Response data: %v", result))
	return result, nil
}

func GetHtmlCodeFromImage(image string, apiKey string, openaimodel string, text string, grid string, theme string, previouseobj []map[string]interface{}) (map[string]interface{}, error) {
	// Get the HTML code from an image
	result, err := getCodeFromImage(s_OPEN_AI_SYSTEM_PROMPT, s_OPENAI_USER_PROMPT, s_OPENAI_USER_PROMPT_WITH_PREVIOUS_DESIGN, image, apiKey, openaimodel, text, grid, theme, previouseobj)
	return result, err
}

func GetBPMLogicFromImage(image string, apiKey string, openaimodel string, text string, grid string, theme string, previouseobj []map[string]interface{}) (map[string]interface{}, error) {

	s_OPENAI_BPM_SYSTEM_PROMPT := `Here are the latest wireframes. There are also some previous outputs here. Could you make a new BPM Logic based on these wireframes and notes and send back just the BPM Logic? The generated BPM Logic object set in to the data node. 
		
		The BPM logic output should be the json object format. 
		
		The BPM Logic object includes the following fields: trancodename, version, status, inputs, firstfuncgroup, functiongroups, isdefault, uuid, description, system.createdby, system.createdon, system.updatedby, system.updatedon, and outputs.etc. 
		
		The firstfuncgroup is the start node of the BPM Logic. The functiongroups include the array of functiongroup. One functiongroup should include multiple functions. Please try to group as much as more functions into one functiongroup to simplify the BPM Logic. One functgroup can have maxium 8 functions.But at least 1 function. 
		
		All functiongroups are in one flow, please make sure the size and space between the functiongroups are enough, at east one block heigh and width must be have betwwen functiongroups. The functions in the same fucntiongroup must have the enough space between them.

		All id or uuid of each function, input, output and functiongroup and BPM Logic object should be unique in whole BPM logic object. The id or uuid should be generated by the system, which is uuid format.
		
		One functiongroup node includes the type which value is "FUNCGROUP", the name, 	functiongroupname, the x, y, height, width, description, routing, routerdef and functions. The funcgroup's functiongroupname and name can be same, but must be unique in one BPM object. The functions include the array of function. 
		
		When the functiongroup is a determination node which has multiple next functiongroups, the routing value is true, otherwise false. If there is not routing, the routerdef only requires defaultfuncgroup, other value can be empty. If routing is true. all node in  the routerdef must have value, and nextfuncgroups, values must be same size array. The defaultfuncgroup must have value, which is a validate's name in the object if it is not the last node. 
		
		The routerdef is used to define the routing between the functiongroups, which includes the defaultfuncgroup, nextfuncgroups, values, and variable. The values include the array of string. Which map to the nextfuncgroup in the nextfuncgroups. The variable is the session or function's output.
		
		The function is a json object which includes the type, content, functionName, inputs, functype, height, y, width, description, name, mapdata, x, id, and outputs. The function node type is "FUNCTION". mapdata, content, inputs and outputs are the json object. 
		
		The inputs and outputs include the array of input and output. The input and output is a json object which includes the datatype, defaultvalue, aliasname, id, list, name, description, value and initialvalue. Input has the additional node source and repeat, and output has additional outputdest. If it is a array of data, the list value is true. 
		
		The datatype is emun value. The enum value is the integer. The enum value is start from 0 to 4. The enum value is 0, 1, 2, 3, 4 which present the string, int, float, bool, datetime and object. 
		
		The outputdest is the output destination. The value can be 0:none, 1:session, 2:externaloutput. The source is the input source. The value can be 0:constant, 1:function, 2:session. if input is a array of data, and repeat value is true, which means the function will be executed multiple times according to literal.
		
		Normally, one block in the image presents a function or multiple functions in the BPM Logic.  And a couple of functions are combined in a fucntiongroup. Ideally, one fucntiongroup has 4 to 8 functions. The function's sequence in the functiongroup is the execution sequence. The functiongroup's sequence in the BPM Logic is the execution sequence. 
		
		According to the description or notes in or side of block in the image to descide the fucntion functype and add the inputs and outputs for each function. Per the functype, assign the content value as well. For example, the query statement for the query functype, and javascript code for the javascript functype. 

		There are following function functype: InputMap,GoExpr,Javascript,	Query,StoreProcedure,SubTranCode,TableInsert, TableUpdate, TableDelete, CollectionInsert,CollectionUpdate,CollectionDelete,ThrowError,SendMessage,SendEmail,ExplodeWorkFlow,StartWorkFlowTask,CompleteWorkFlowTask,SendMessagebyKafka,SendMessagebyMQTT,SendMessagebyAQMP,WebServiceCall
		The function functype value is the enum value. The enum value is the integer. The enum value is start from 1 to 22. 
		Here is a sample of BPM Logic object:
		{
			"trancodename": "Sample of BPM",
			"firstfuncgroup": "ValidateData",
			"system.updatedby": "system",
			"system.createdon": "2024-05-23T16:04:20.254Z",
			"version": "",
			"system.updatedon": {
			  "$date": "2024-05-23T19:20:25.866Z"
			},
			"status": null,
			"description": "undefined",
			"isdefault": false,
			"system.createdby": "system",
			"uuid": "af69275a_9567_4068_8108_7ef3d75988c0",
			"functiongroups": [
			  {
				"height": 100,
				"x": 120,
				"description": "",
				"name": "ValidateData",
				"type": "FUNCGROUP",
				"functiongroupname": "ValidateData",
				"elements": [],
				"width": 250,
				"functions": [
				  {
					"name": "Query Machine data from database",
					"inputs": [
					  {
						"defaultvalue": "",
						"description": "",
						"list": false,
						"source": 4,
						"value": "",
						"aliasname": "Machine",
						"datatype": null,
						"name": "Machine",
						"repeat": false,
						"id": "8416b6fd_ebb1_47e3_9759_bd7e409d6c69"
					  }
					],
					"height": 200,
					"outputs": [
					  {
						"datatype": 0,
						"defaultvalue": "",
						"description": "Name",
						"id": "840df059_9d65_4550_b7ff_38995e6f6a1d",
						"name": "Name",
						"outputdest": [],
						"aliasname": []
					  },
					  {
						"source": 0,
						"aliasname": [],
						"defaultvalue": "0",
						"id": "d37c4f0c_931c_483b_867d_077a249dbdc8",
						"outputdest": [],
						"name": "ColumnCount",
						"datatype": 1,
						"value": ""
					  },
					  {
						"datatype": 1,
						"value": "",
						"source": 0,
						"aliasname": [],
						"defaultvalue": "0",
						"id": "cd649bf4_9ccc_413b_af4b_d2f34b96951b",
						"outputdest": [],
						"name": "RowCount"
					  }
					],
					"width": 250,
					"functype": 3,
					"content": "Select Name \nFROM machines \nWhere Name = @Machine",
					"x": 420,
					"y": 120,
					"id": "043a3c6b_e665_40be_ac93_98bc8b135e6b",
					"type": "FUNCTION",
					"functionName": "Query Machine data from database",
					"description": "",
					"mapdata": {}
				  },
				  {
					"functionName": "CheckMachineData",
					"mapdata": {},
					"y": 370,
					"inputs": [
					  {
						"value": "",
						"description": "MachineName",
						"source": 1,
						"repeat": false,
						"list": false,
						"datatype": 0,
						"defaultvalue": "",
						"aliasname": "Query Machine data from database.Name",
						"id": "7dfa9b6b_eaff_4e4c_927a_172e49c9263f",
						"name": "MachineName"
					  }
					],
					"x": 950,
					"width": 250,
					"height": 200,
					"functype": 2,
					"id": "6f7f2bb8_2c18_464c_abb0_df7963e4a5d7",
					"content": "if(MachineName == \"\"){\n{\n\tRoute = \"Error\"    \n}\n  ",
					"name": "CheckMachineData",
					"outputs": [
					  {
						"name": "Route",
						"outputdest": [],
						"aliasname": [],
						"datatype": 0,
						"defaultvalue": "",
						"description": "Route",
						"id": "70161ebf_f751_4395_a2c5_711e6f8c05f2"
					  }
					],
					"description": "",
					"type": "FUNCTION"
				  }
				],
				"y": 207.5,
				"id": "53cfe1f6_fe39_4a6c_b6c4_a30f196bb260",
				"routing": true,
				"routerdef": {
				  "defaultfuncgroup": "",
				  "nextfuncgroups": [
					"ProcessData",
					"ShowError"
				  ],
				  "values": [
					"new value",
					"new value"
				  ],
				  "variable": "Javascript01.Route"
				}
			  },
			  {
				"elements": [],
				"height": 100,
				"description": "",
				"type": "FUNCGROUP",
				"id": "93b203e0_260f_4440_ace2_4d04a9c22cf7",
				"name": "ShowError",
				"width": 250,
				"functions": [
				  {
					"functype": 12,
					"outputs": [],
					"id": "3957b24f_9bab_470a_a4fb_c9a7a3e740ed",
					"inputs": [
					  {
						"datatype": 0,
						"defaultvalue": "",
						"value": "",
						"aliasname": "",
						"id": "9528c018_065b_4305_9079_ceacf4dd1c27",
						"name": "ErrorCode",
						"source": 0,
						"description": "ErrorCode",
						"list": false,
						"repeat": false
					  }
					],
					"x": 100,
					"y": 100,
					"type": "FUNCTION",
					"mapdata": {},
					"width": 250,
					"description": "ThrowError01",
					"content": "",
					"height": 200,
					"name": "ThrowError01"
				  }
				],
				"functiongroupname": "ShowError",
				"y": 357.5,
				"routerdef": {
				  "values": [],
				  "variable": "No Routing",
				  "defaultfuncgroup": "",
				  "nextfuncgroups": []
				},
				"x": 420,
				"routing": true
			  },
			  {
				"y": 357.5,
				"routerdef": {
				  "nextfuncgroups": [],
				  "values": [],
				  "variable": "No Routing",
				  "defaultfuncgroup": ""
				},
				"routing": true,
				"functions": [
				  {
					"inputs": [
					  {
						"repeat": false,
						"datatype": null,
						"aliasname": "Machine",
						"defaultvalue": "",
						"description": "",
						"id": "f7ab4b62_4785_4962_8f9d_1e2774caa8cc",
						"name": "MachineKey",
						"value": "",
						"list": false,
						"source": 4
					  },
					  {
						"id": "e48a3328_fe94_465a_82a8_ef7a1f68d881",
						"name": "TanbleName",
						"repeat": false,
						"datatype": null,
						"defaultvalue": "machine_states",
						"value": "machine_states",
						"aliasname": "",
						"list": false,
						"description": "",
						"source": null
					  },
					  {
						"value": "2",
						"datatype": 1,
						"defaultvalue": "",
						"description": "",
						"list": false,
						"aliasname": "",
						"name": "Status",
						"source": null,
						"id": "34f6c6c1_dd3d_4b49_87de_560d11e585c9",
						"repeat": false
					  },
					  {
						"datatype": 0,
						"value": "",
						"source": 0,
						"aliasname": "",
						"defaultvalue": "",
						"id": "4131fa68_2549_4807_8a0a_4c0852b839db",
						"name": "TableName"
					  },
					  {
						"aliasname": "",
						"defaultvalue": "true",
						"id": "0c844eb9_2be4_422e_9eb5_d47f6f154b80",
						"name": "Execution",
						"datatype": 3,
						"value": "true",
						"source": 0
					  },
					  {
						"defaultvalue": "CurrentUTCTime",
						"id": "2240d4bd_c32c_4d5d_9984_a8381e2e6877",
						"name": "UpdatedOn",
						"datatype": 4,
						"value": "",
						"source": 0,
						"aliasname": ""
					  },
					  {
						"value": "",
						"source": 0,
						"aliasname": "",
						"defaultvalue": "CurrentUser",
						"id": "006a5ff0_c85c_49d6_b6f1_ec8da57eef53",
						"name": "UpdatedBy",
						"datatype": 0
					  }
					],
					"functype": 7,
					"name": "Close the current machine state",
					"outputs": [
					  {
						"id": "e21e006a_98ca_42de_933b_d4b990378093",
						"outputdest": [],
						"name": "RowCount",
						"datatype": 1,
						"value": "",
						"source": 0,
						"aliasname": [],
						"defaultvalue": "0"
					  }
					],
					"description": "",
					"mapdata": {},
					"width": 250,
					"height": 200,
					"x": -270,
					"type": "FUNCTION",
					"y": -140,
					"functionName": "Close the current machine state",
					"id": "97ad151f_5674_4df0_a921_24edad30290c"
				  },
				  {
					"id": "b167d058_f31a_49a2_87a9_cf9791d7f040",
					"inputs": [
					  {
						"description": "",
						"id": "1a41827a_a9b2_4a91_92fa_f36e85aa9702",
						"list": false,
						"defaultvalue": "",
						"source": null,
						"datatype": null,
						"value": "machine_states",
						"aliasname": "",
						"name": "TableName",
						"repeat": false
					  },
					  {
						"datatype": null,
						"list": false,
						"aliasname": "State",
						"value": "",
						"defaultvalue": "",
						"description": "",
						"id": "061d7ab9_91a0_4cc0_83ff_6d32a12a09dc",
						"name": "State",
						"repeat": false,
						"source": 4
					  },
					  {
						"description": "",
						"name": "Status",
						"value": "1",
						"source": null,
						"repeat": false,
						"aliasname": "",
						"list": false,
						"datatype": null,
						"defaultvalue": "",
						"id": "f07929dc_5f65_4f1b_9742_3af525120f43"
					  },
					  {
						"source": 0,
						"value": "true",
						"aliasname": "",
						"datatype": 3,
						"defaultvalue": "",
						"id": "d54e28af_58d0_456a_bfb7_1cc1eabeab67",
						"name": "Execution"
					  },
					  {
						"defaultvalue": "CurrentUTCTime",
						"id": "4e3039a2_7229_4f93_b8f6_52f264c3c7cf",
						"name": "CreatedOn",
						"source": 2,
						"value": "",
						"aliasname": "",
						"datatype": 4
					  },
					  {
						"source": 2,
						"value": "",
						"aliasname": "",
						"datatype": 0,
						"defaultvalue": "CurrentUser",
						"id": "4401c808_62aa_444e_aecd_f6f06e5ba0d4",
						"name": "CreatedBy"
					  },
					  {
						"defaultvalue": "",
						"value": "",
						"id": "0696ff00_f704_4809_ad09_13f2e180cf75",
						"list": false,
						"repeat": false,
						"aliasname": "Machine",
						"datatype": null,
						"source": 4,
						"description": "",
						"name": "Machine"
					  }
					],
					"height": 200,
					"outputs": [
					  {
						"defaultvalue": "0",
						"id": "e7ad4102_3034_452c_84c7_79669b164040",
						"name": "Identify",
						"outputdest": [],
						"source": 0,
						"value": "",
						"aliasname": [],
						"datatype": 1
					  }
					],
					"width": 250,
					"description": "",
					"y": 180,
					"name": "Insert the new machine state",
					"type": "FUNCTION",
					"functionName": "Insert the new machine state",
					"x": -370,
					"functype": 6,
					"mapdata": {}
				  }
				],
				"name": "ProcessData",
				"id": "8e3cf0cb_46a8_48ea_afab_b419bada2478",
				"width": 250,
				"type": "FUNCGROUP",
				"description": "",
				"height": 100,
				"elements": [],
				"x": 120,
				"functiongroupname": "ProcessData"
			  }
			]
		  }
		Please per the above BPM Logic object definition to generate the BPM Logic object according to the wireframe image attached. And try to have multiple functions in one functiongroup. 
	`

	s_OPENAI_BPM_USER_PROMPT := "Here are the latest wireframes. There are also some previous outputs here. Could you make a new BPM Logic based on these wireframes and notes and send back just the BPM Logic?"

	result, err := getCodeFromImage(s_OPENAI_BPM_SYSTEM_PROMPT, s_OPENAI_BPM_USER_PROMPT, "", image, apiKey, openaimodel, text, grid, theme, previouseobj)
	return result, err
}

func GetWorkFlowFromImage(image string, apiKey string, openaimodel string, text string, grid string, theme string, previouseobj []map[string]interface{}) (map[string]interface{}, error) {

	s_OPENAI_WorkFlow_SYSTEM_PROMPT := `Here are the latest wireframes. There are also some previous outputs here. Could you make a new work flowc based on these wireframes and notes and send back just the work flow? The generated work flow object set in to the data node. 
	The work flow should be the same as the wireframes provided. the work flow output should be the json object format. 
	The work flow object includes the following fields: name, type, description, version, isdefault, uuid, system.createdby, system.createdon, system.updatedby, system.updatedon,nodes and links.
	The nodes include the array of node. The node is a json object which includes the type, description, name, height, y, width, x, id, roles, users, page, trancode and processdata. 
	the node type includes: start, end, task, gateway, ane notes. The roles and users are array of string which defines the role or user to execute the node. The page is the page to be opened when the node is clicked. The trancode is the trancode to be called when the node is executed. The processdata is the data to be passed to the trancode.
	The links include the array of link. The link is a json object which includes the source, target, type,name, label and id. The source and target are the node id. The type is the "link". The name is the link name. The label is the link label.
	The block in the image presents a node in the work flow. and create the link between the nodes according to the wireframes. 
	Please generate the work flow object according to the wireframe image attached.
	`

	s_OPENAI_WorkFlow_USER_PROMPT := "Here are the latest wireframes. There are also some previous outputs here. Could you make a new workflow based on these wireframes and notes and send back just the BPM Logic?"

	result, err := getCodeFromImage(s_OPENAI_WorkFlow_SYSTEM_PROMPT, s_OPENAI_WorkFlow_USER_PROMPT, "", image, apiKey, openaimodel, text, grid, theme, previouseobj)
	return result, err
}
