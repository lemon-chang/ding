package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"

	"github.com/tealeg/xlsx"
	"strings"
	"time"
)

func main() {
	var a int = 50
	a = 60
	fmt.Println(a)
	//https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.afb839b7W85NCP#/jsapi?api=biz.chat.chooseConversationByCorpId
	accessToken := "e2cdeaeeb4b13b8aaf9d1c08601b05d2"
	//openConversationId := "cidbGeQn/i+hZRXKpwdfb1nug==" //乐职院大群
	openConversationId := "cidXyY/L7HoOFKEbkFGabB4fg==" //乐知四期
	//coolAppCode := "COOLAPP-1-101DFE55065B213EF05B000F" //三月软件
	coolAppCode := "COOLAPP-1-102001DE38062133DFC5000V" //乐职院
	userIds, err := GetUserIds(accessToken, openConversationId, coolAppCode)
	if err != nil {
		fmt.Println(err)
	}

	var finalResults []FinalResult
	for _, userId := range userIds {
		result, err := PostGetUserDetail(accessToken, userId)
		if err != nil {
			zap.L().Error(fmt.Sprintf("通过UserId查询用户详细信息失败，userId = %s", userId), zap.Error(err))
		}
		var departmentNameList []string
		for i := 0; i < len(result.Result.DeptIdList); i++ {
			result1, _ := GetDepartmentName(accessToken, result.Result.DeptIdList[i])
			departmentNameList = append(departmentNameList, result1.Result.Name)
		}
		result.Result.DepartmentNameList = departmentNameList
		var finalResult FinalResult
		finalResult.Mobile = result.Result.Mobile
		finalResult.DepartmentNameList = result.Result.DepartmentNameList
		finalResult.Name = result.Result.Name
		fmt.Println(finalResult)
		finalResults = append(finalResults, finalResult)
	}
	Export(finalResults)

}

// HeaderColumn 表头字段定义
type HeaderColumn struct {
	Field string // 字段，数据映射到的数据字段名
	Title string // 标题，表格中的列名称
}

// SetHeader 写模式下，设置字段表头和字段顺序
// 参数 header 为表头和字段映射关系，如：HeaderColumn{Field:"title", Title: "标题", Requre: true}
// 参数 width  为表头每列的宽度，单位 CM：map[string]float64{"title": 0.8}
func SetHeader(sheet *xlsx.Sheet, header []*HeaderColumn, width map[string]float64) (*xlsx.Sheet, error) {
	if len(header) == 0 {
		return nil, errors.New("Excel.SetHeader 错误: 表头不能为空")
	}

	// 表头样式
	style := xlsx.NewStyle()

	font := xlsx.DefaultFont()
	font.Bold = true

	alignment := xlsx.DefaultAlignment()
	alignment.Vertical = "center"

	style.Font = *font
	style.Alignment = *alignment

	style.ApplyFont = true
	style.ApplyAlignment = true

	// 设置表头字段
	row := sheet.AddRow()
	row.SetHeightCM(1.0)
	row_w := make([]string, 0)
	for _, column := range header {
		row_w = append(row_w, column.Field)
		cell := row.AddCell()
		cell.Value = column.Title
		cell.SetStyle(style) //设置单元样式
	}

	// 表格列，宽度
	if len(row_w) > 0 {
		for k, v := range row_w {
			if width[v] > 0.0 {
				sheet.SetColWidth(k, k, width[v]*10)
			}
		}
	}

	return sheet, nil
}

func Export(finalResults []FinalResult) {
	file := xlsx.NewFile()                // NewWriter 创建一个Excel写操作实例
	sheet, err := file.AddSheet("number") //表实例
	if err != nil {
		fmt.Printf(err.Error())
	}

	headers := []*HeaderColumn{
		{Field: "Name", Title: "姓名"},
		{Field: "Mobile", Title: "电话"},
		{Field: "DepartmentNameList", Title: "部门列表"},
	}
	style := map[string]float64{
		"Name":               2.0,
		"Mobile":             2.0,
		"DepartmentNameList": 2.0,
	}
	sheet, _ = SetHeader(sheet, headers, style)

	for _, stu := range finalResults {
		data := make(map[string]string)
		data["Name"] = stu.Name
		data["Mobile"] = stu.Mobile
		ss := fmt.Sprintf(strings.Join(stu.DepartmentNameList, ","))
		data["DepartmentNameList"] = ss
		row := sheet.AddRow()
		row.SetHeightCM(0.8)
		for _, field := range headers {
			row.AddCell().Value = data[field.Field]
		}
	}
	outFile := "C:\\Users\\lenovo\\Desktop\\first.xlsx"
	err = file.Save(outFile)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Println("\n\nexport success")
}

func GetDepartmentName(access_token string, deptId int64) (response GetDepartmentDetailRequestResponse, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/department/get?access_token=" + access_token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	r := GetDepartmentDetailRequestBody{
		DeptId: deptId,
	}
	bodymarshal, err := json.Marshal(&r)
	if err != nil {
		return
	}
	reqBody := strings.NewReader(string(bodymarshal))
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	return
}
func PostGetUserDetail(access_token string, UserId string) (r Response, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + access_token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	b := Body{
		UserId: UserId,
	}
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	reqBody := strings.NewReader(string(bodymarshal))
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	//if r.Errcode == 33012 {
	//	return model.Tele{}, errors.New("无效的userId,请检查userId是否正确")
	//} else if r.Errcode == 400002 {
	//	return model.Tele{}, errors.New("无效的参数,请确认参数是否按要求输入")
	//} else if r.Errcode == -1 {
	//	return model.Tele{}, errors.New("系统繁忙")
	//}

	return
}

type Body struct {
	UserId   string `json:"userid"`
	Language string `json:"language"`
}
type GetDepartmentDetailRequestBody struct {
	DeptId int64 `json:"dept_id"`
}
type GetDepartmentDetailRequestResponse struct {
	Result GetDepartmentDetailRequestResult `json:"result"`
}
type Response struct {
	Result Result `json:"result"`
}
type FinalResult struct {
	Name   string `json:"name"`
	Mobile string `json:"mobile"`

	DepartmentNameList []string `json:"department_name_list"`
}
type GetDepartmentDetailRequestResult struct {
	ParentId int64  `json:"parent_id"`
	Name     string `json:"name"`
}

type Result struct {
	Mobile             string   `json:"mobile"`
	Name               string   `json:"name"`
	DeptIdList         []int64  `json:"dept_id_list"`
	DepartmentNameList []string `json:"department_name_list"`
}

//type FinalResult struct {
//	Mobile string `json:"mobile"`
//	Name   string `json:"name"`
//	DepartmentNameList []string `json:"department_name_list"`
//}

func GetUserIds(access_token, OpenConversationId string, coolAppcode string) (userIds []string, _err error) {
	olduserIds := []*string{}
	client, _err := createClient()
	if _err != nil {
		return
	}
	batchQueryGroupMemberHeaders := &dingtalkim_1_0.BatchQueryGroupMemberHeaders{}
	batchQueryGroupMemberHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	batchQueryGroupMemberRequest := &dingtalkim_1_0.BatchQueryGroupMemberRequest{
		OpenConversationId: tea.String(OpenConversationId),
		CoolAppCode:        tea.String(coolAppcode),
		MaxResults:         tea.Int64(300),
		NextToken:          tea.String("XXXXX"),
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.BatchQueryGroupMemberWithOptions(batchQueryGroupMemberRequest, batchQueryGroupMemberHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		olduserIds = result.Body.MemberUserIds
		return
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}

	}
	userIds = make([]string, len(olduserIds))
	for i, id := range olduserIds {
		userIds[i] = *id
	}
	return
}
func createClient() (_result *dingtalkim_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkim_1_0.Client{}
	_result, _err = dingtalkim_1_0.NewClient(config)
	return _result, _err
}
