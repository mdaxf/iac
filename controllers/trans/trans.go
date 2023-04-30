package trans

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
)

type TranCodeController struct {
}

func (e *TranCodeController) ExecuteTranCode(ctx *gin.Context) {
	/*	jsonString, err := json.Marshal(ctx.Request)
		if err != nil {
			fmt.Println("Error marshaling json:", err)
			return
		}
		log.Println(string(jsonString))  */
	var tcdata TranCodeData
	if err := ctx.BindJSON(&tcdata); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print(tcdata.TranCode)
	tcode, err := e.getTransCode(tcdata.TranCode)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := json.Marshal(tcode)
	if err != nil {
		fmt.Println("Error marshaling json:", err)
		return
	}
	log.Println(string(jsonString))

	tf := trancode.NewTranFlow(tcode, map[string]interface{}{}, map[string]interface{}{})
	outputs, err := tf.Execute()

	if err == nil {
		ctx.JSON(http.StatusOK, gin.H{"outputs": outputs})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"execution failed": err.Error()})
}

func (e *TranCodeController) getTransCode(name string) (types.TranCode, error) {
	log.Println(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	data, err := ioutil.ReadFile(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	if err != nil {
		log.Println(fmt.Errorf("failed to read configuration file: %v", err))
		return types.TranCode{}, fmt.Errorf("failed to read configuration file: %v", err)
	}
	log.Println(string(data))
	fmt.Println(string(data))
	return trancode.Bytetoobj(data)
}

type TranCodeData struct {
	TranCode string                 `json:"code"`
	inputs   map[string]interface{} `json:"Inputs"`
}
