package area

// excelize docs https://www.bookstack.cn/read/excelize-2.2-zh/b2e09ee3ee36ed31.md

import (
	"bufio"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"io"
	"os"
	"phone-area/schema"
	"regexp"
	"strconv"
	"sync"
)

type Excel struct {
	File        string
	ProjectFile string
	OutputFile  string
}

func NewExcel(file, outputFile, projectFile string) *Excel {
	return &Excel{
		File:        file,
		ProjectFile: projectFile,
		OutputFile:  outputFile,
	}
}

func (a *Excel) Run() error {
	phonesInfo, err := a.getPhones()
	if err != nil {
		return err
	}

	projects, err := a.getProjects()
	if err != nil {
		return err
	}

	group := sync.WaitGroup{}
	for _, info := range phonesInfo {
		group.Add(1)
		go func(pi *schema.PhoneInfo) {
			res, err := getInfo(pi)
			if err != nil {
				return
			}

			pi.Province = res.Data.Province
			pi.City = res.Data.City
			pi.Area = pi.Province + pi.City
			pi.ServiceProvider = res.Data.ServiceProvider
			group.Done()
		}(info)
	}
	group.Wait()

	err = a.Save(phonesInfo, projects)
	if err != nil {
		return err
	}

	return nil
}

func (a *Excel) Save(phonesInfo schema.PhoneInfos, projects map[string]string) error {
	f := excelize.NewFile()
	sheetName := "实时数据抓取"
	f.NewSheet(sheetName)
	sheetIndex := f.GetSheetIndex(sheetName)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"9BBB59"},
		},
		Font: &excelize.Font{
			Bold:   true,
			Family: "宋体",
			Size:   11,
			Color:  "FFFFFF",
		},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
	})
	if err != nil {
		return err
	}

	textStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "宋体",
			Size:   11,
		},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
	})
	if err != nil {
		return err
	}

	urlStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "宋体",
			Size:   11,
		},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
	})
	if err != nil {
		return err
	}

	err = f.SetCellStyle(sheetName, "A1", "Z1", headerStyle)
	if err != nil {
		return err
	}
	_ = f.SetCellValue(sheetName, "A1", "访客手机号")
	_ = f.SetCellValue(sheetName, "B1", "项目")
	_ = f.SetCellValue(sheetName, "C1", "省份")
	_ = f.SetCellValue(sheetName, "D1", "市区")
	_ = f.SetCellValue(sheetName, "E1", "运营商")
	_ = f.SetCellValue(sheetName, "F1", "目标")
	_ = f.SetRowHeight(sheetName, 1, 20)
	_ = f.SetColWidth(sheetName, "A", "A", 20.0)
	_ = f.SetColWidth(sheetName, "B", "B", 20.0)
	_ = f.SetColWidth(sheetName, "C", "E", 10)
	_ = f.SetColWidth(sheetName, "F", "F", 90)

	var axis string
	offset := 2
	lastIndex := 1
	for index, info := range phonesInfo {
		lastIndex = index + offset
		axis = fmt.Sprintf("A%d", lastIndex)
		phone, err := strconv.ParseInt(info.Number, 10, 64)
		if err != nil {
			continue
		}
		_ = f.SetCellInt(sheetName, axis, int(phone))

		if o, ok := projects[info.WebSite]; ok {
			axis = fmt.Sprintf("B%d", lastIndex)
			_ = f.SetCellValue(sheetName, axis, o)
		}

		axis = fmt.Sprintf("C%d", lastIndex)
		_ = f.SetCellValue(sheetName, axis, info.Province)

		axis = fmt.Sprintf("D%d", lastIndex)
		_ = f.SetCellValue(sheetName, axis, info.City)

		axis = fmt.Sprintf("E%d", lastIndex)
		_ = f.SetCellValue(sheetName, axis, info.ServiceProvider)

		axis = fmt.Sprintf("F%d", lastIndex)
		_ = f.SetCellValue(sheetName, axis, info.WebSite)

		_ = f.SetRowHeight(sheetName, lastIndex, 20)
	}
	err = f.SetCellStyle(sheetName, "A2", fmt.Sprintf("E%d", lastIndex), textStyle)
	if err != nil {
		return err
	}
	err = f.SetCellStyle(sheetName, "F2", fmt.Sprintf("F%d", lastIndex), urlStyle)
	if err != nil {
		return err
	}

	// 过滤器
	err = f.AutoFilter(sheetName, "A1", "F1", "")
	if err != nil {
		return err
	}

	// 冻结窗口
	err = f.SetPanes(sheetName, `{"freeze":true,"split":false,"x_split":0,"y_split":1}`)
	if err != nil {
		return err
	}

	f.SetActiveSheet(sheetIndex)
	f.DeleteSheet("Sheet1")
	if err := f.SaveAs(a.OutputFile); err != nil {
		return err
	}

	return nil
}

func (a *Excel) getProjects() (map[string]string, error) {
	projects := make(map[string]string)
	if a.ProjectFile == "" {
		return projects, nil
	}

	fi, err := os.Open(a.ProjectFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		str := string(a)
		if str == "" {
			continue
		}

		compile := regexp.MustCompile("(\\S+)\\s+(\\S+)")
		subMatch := compile.FindStringSubmatch(str)
		if len(subMatch) > 0 {
			webSite := subMatch[1]
			project := subMatch[2]
			projects[webSite] = project
		}
	}

	return projects, nil
}

func (a *Excel) getPhones() (schema.PhoneInfos, error) {
	f, err := excelize.OpenFile(a.File)
	if err != nil {
		return nil, err
	}

	phoneInfos := schema.PhoneInfos{}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if len(row) < 2 || row[0] == "" || row[1] == "" {
			continue
		}
		number := row[0]
		website := row[1]

		phoneInfo := &schema.PhoneInfo{
		}
		compile := regexp.MustCompile("[0-9]{11}")
		subMatch := compile.FindStringSubmatch(number)
		if len(subMatch) > 0 {
			phoneInfo.Number = subMatch[0]
		}

		compile = regexp.MustCompile("https?://\\S*")
		subMatch = compile.FindStringSubmatch(website)
		if len(subMatch) > 0 {
			phoneInfo.WebSite = subMatch[0]
		}
		phoneInfos = append(phoneInfos, phoneInfo)
	}

	return phoneInfos, nil
}
